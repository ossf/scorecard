# check-evaluation Specification

## Purpose

Check evaluation turns the findings produced by a check's probes into a single
check result: an integer score from 0 to 10 (or -1 when inconclusive), a
human-readable reason, and a list of details. This spec covers the shared
scoring rules and result shapes, and the scoring policy of each check.

## Requirements

### Requirement: Check Score Range

The system SHALL score every conclusive check result as an integer between 0
(the minimum result score) and 10 (the maximum result score), inclusive.

#### Scenario: Conclusive score within range

- **WHEN** a check evaluation completes with enough evidence to score
- **THEN** the check result score is an integer from 0 to 10

#### Scenario: Out-of-range score becomes a runtime error

- **WHEN** a check evaluation attempts to assign a score below 0 or above 10
- **THEN** the check result is a runtime error result
- **AND** its reason asks the user to report the invalid score

### Requirement: Inconclusive Check Result

The system SHALL report a score of -1 with an explanatory reason and no error
when a check runs successfully but lacks enough evidence to assign a score.

#### Scenario: Not enough evidence

- **GIVEN** a check whose probes report that the evaluated subject is not
  applicable to the repository (for example, no releases or no workflows)
- **WHEN** the check is evaluated
- **THEN** the check result score is -1
- **AND** the result carries no error

### Requirement: Runtime Error Result

The system SHALL report a check that fails to run as a result with score -1,
the error attached, and a reason equal to the error message.

#### Scenario: Raw data collection fails

- **WHEN** collecting raw data or running probes for a check returns an error
- **THEN** the check result score is -1
- **AND** the result carries the error
- **AND** the result reason is the error message

### Requirement: Probe Set Validation

The system SHALL return a runtime error result with the reason "invalid probe
results" when the set of distinct probes in the findings passed to a check's
evaluation differs from the set of probes that check expects.

#### Scenario: Missing probe

- **GIVEN** the Security-Policy check expects findings from four probes
- **WHEN** the findings contain only three of those probes
- **THEN** the check result is a runtime error result
- **AND** the reason contains "invalid probe results"

#### Scenario: Unexpected probe

- **WHEN** the findings for a check contain a probe the check does not expect
- **THEN** the check result is a runtime error result

### Requirement: Proportional Scoring

The system SHALL compute a proportional score as 10 times the number of
successes divided by the total, rounded down to an integer and capped at 10,
and SHALL return 0 when the total is 0. When a check reports a proportional
score, the system SHALL append " -- score normalized to N" to the reason, where
N is the score.

#### Scenario: Partial success

- **WHEN** a proportional score is computed for 2 successes out of 3
- **THEN** the score is 6

#### Scenario: Zero total

- **WHEN** a proportional score is computed with a total of 0
- **THEN** the score is 0

#### Scenario: Normalized reason

- **WHEN** a check reports a proportional score of 6 with reason "R"
- **THEN** the result reason is "R -- score normalized to 6"

### Requirement: Weighted Proportional Scoring

The system SHALL compute a weighted proportional score over groups of
(successes, total, weight) as 10 times the sum of successes times weight
divided by the sum of totals times weight, rounded down and capped at 10. The
system SHALL ignore groups with a total of 0, SHALL return -1 when every group
has a total of 0, SHALL return 10 when every counted group has weight 0, and
SHALL return an error when any group has more successes than its total.

#### Scenario: Weighted groups

- **WHEN** a weighted score is computed for groups (1 of 2, weight 2) and
  (2 of 2, weight 8)
- **THEN** the score is 9

#### Scenario: No countable groups

- **WHEN** every group has a total of 0
- **THEN** the weighted score is -1

#### Scenario: Successes exceed total

- **WHEN** any group reports more successes than its total
- **THEN** the weighted score computation fails with an error

### Requirement: Detail Levels

The system SHALL attach to each check result an ordered list of details, each
tagged Info, Warn, or Debug. Which details are rendered is defined in
`result-formats`.

#### Scenario: Default detail levels for findings

- **GIVEN** a check whose policy logs every finding (Dependency-Update-Tool,
  Fuzzing, License, Packaging, or Security-Policy)
- **WHEN** the check is evaluated
- **THEN** findings with a true outcome are logged as Info
- **AND** findings with a false outcome are logged as Warn
- **AND** findings with any other outcome are logged as Debug

### Requirement: Retry On Unreachable Repository

The system SHALL attempt a check up to three times when the check result
carries a repository-unreachable error, recording a Warn detail for each such
attempt, and SHALL report the last attempt's result.

#### Scenario: Transient unreachable repository

- **GIVEN** a check whose first attempt fails because the repository is
  unreachable
- **WHEN** the second attempt succeeds
- **THEN** the reported result is the second attempt's result
- **AND** the result details include a Warn detail describing the first
  failure

#### Scenario: Other error is not retried

- **GIVEN** a check whose first attempt returns a runtime error other than
  repository unreachable
- **WHEN** the check runs
- **THEN** the check is attempted only once
- **AND** the reported result carries that error

### Requirement: Finding Aggregation

The system SHALL attach to each check result the findings its probes produced
and SHALL also append those findings, in the order check results are
collected, to the run's overall list of findings. A runtime error result
SHALL contribute no findings.

#### Scenario: Check findings in the overall result

- **WHEN** the License check runs and its probes produce two findings
- **THEN** the License check result carries those two findings
- **AND** the run's overall findings include the same two findings

### Requirement: Binary-Artifacts Scoring

The system SHALL score Binary-Artifacts as 10 when the
hasUnverifiedBinaryArtifacts probe reports no binaries, and otherwise as 10
minus the number of binary findings, with a minimum of 0. Each detected binary
SHALL be logged as a Warn detail "binary detected" with its path and line.

#### Scenario: No binaries

- **WHEN** the probe reports no unverified binary artifacts
- **THEN** the score is 10 with reason "no binaries found in the repo"

#### Scenario: Some binaries

- **WHEN** the probe reports 3 unverified binary artifacts
- **THEN** the score is 7 with reason "binaries present in source code"

#### Scenario: Many binaries

- **WHEN** the probe reports 12 unverified binary artifacts
- **THEN** the score is 0

### Requirement: Branch-Protection Scoring

The system SHALL score Branch-Protection in five cumulative tiers worth 3, 3,
2, 1, and 1 points, aggregated across all evaluated branches, awarding each
tier's points in proportion to the settings satisfied and stopping at the first
tier that is not fully satisfied. The final score SHALL be rounded down. The
tiers SHALL be:

1. Blocking force pushes and blocking deletion.
2. Requiring at least one approving review, requiring pull requests to change
   code, requiring branches to be up to date, and requiring approval of the
   most recent push.
3. Requiring status checks to pass before merging.
4. Requiring at least two approving reviews and requiring code owner review.
5. Dismissing stale reviews and applying the rules to administrators.

A tier whose settings are all unavailable (for example, because the token lacks
administrator access) SHALL award its full points.

#### Scenario: No branches detected

- **WHEN** any branch-protection probe reports not applicable
- **THEN** the result is inconclusive with reason "unable to detect any
  development/release branches"

#### Scenario: Unprotected branch

- **WHEN** a branch is not protected
- **THEN** a Warn detail "branch protection not enabled for branch 'NAME'" is
  emitted
- **AND** no per-setting details are emitted for that branch

#### Scenario: Only basic protection

- **GIVEN** a single branch that blocks force pushes and deletion but satisfies
  none of the tier 2 settings
- **WHEN** the check is evaluated
- **THEN** the score is 3

#### Scenario: Fully protected

- **WHEN** every setting in all five tiers is satisfied on every branch
- **THEN** the score is 10 with reason "branch protection is fully enabled on
  development and all release branches"

#### Scenario: No protection

- **WHEN** no branch blocks force pushes or deletion
- **THEN** the score is 0 with reason "branch protection not enabled on
  development/release branches"

### Requirement: CI-Tests Scoring

The system SHALL score CI-Tests as the proportional score of merged pull
requests checked by a CI test over all merged pull requests considered, and
SHALL log each pull request's finding message as a Debug detail.

#### Scenario: No pull requests

- **WHEN** the testsRunInCI probe reports not applicable
- **THEN** the result is inconclusive with reason "no pull request found"

#### Scenario: Partially tested

- **WHEN** 15 of 20 merged pull requests were checked by a CI test
- **THEN** the score is 7
- **AND** the reason begins "15 out of 20 merged PRs checked by a CI test"

### Requirement: CII-Best-Practices Scoring

The system SHALL score CII-Best-Practices from the OpenSSF Best Practices badge
level: Gold 10, Silver 7, Passing 5, InProgress 2, and no badge 0.

#### Scenario: Passing badge

- **WHEN** the hasOpenSSFBadge probe reports a Passing badge
- **THEN** the score is 5 with reason "badge detected: Passing"

#### Scenario: No badge

- **WHEN** the probe reports no badge
- **THEN** the score is 0

#### Scenario: Unexpected badge data

- **WHEN** the probe returns more than one finding, omits the badge level, or
  reports an unknown level
- **THEN** the check result is a runtime error result

### Requirement: Code-Review Scoring

The system SHALL score Code-Review as 10 when all recent changesets were
approved, and otherwise as the proportional score of approved changesets over
total changesets.

#### Scenario: No changesets

- **WHEN** the codeApproved probe reports not applicable
- **THEN** the result is inconclusive with the probe's message as reason

#### Scenario: All reviewed

- **WHEN** the probe reports a true outcome
- **THEN** the score is 10 with reason "all changesets reviewed"

#### Scenario: Partially reviewed

- **WHEN** the probe reports 12 approved out of 30 changesets
- **THEN** the score is 4

#### Scenario: Probe error

- **WHEN** the probe reports an error outcome
- **THEN** the check result is a runtime error result

### Requirement: Contributors Scoring

The system SHALL score Contributors as 10 when more than 3 distinct
organizations or companies contributed, and otherwise as the proportional score
of that count over 3. When at least one is found, the system SHALL emit an Info
detail listing them in sorted order.

#### Scenario: Two organizations

- **WHEN** 2 contributing organizations are found
- **THEN** the score is 6
- **AND** an Info detail "found contributions from: A, B" is emitted

#### Scenario: Three or more organizations

- **WHEN** 3 or more contributing organizations are found
- **THEN** the score is 10

#### Scenario: None

- **WHEN** no contributing organizations are found
- **THEN** the score is 0 and no Info detail is emitted

### Requirement: Dangerous-Workflow Scoring

The system SHALL score Dangerous-Workflow as 0 when any workflow has an
untrusted checkout or a script injection pattern, and 10 otherwise, logging each
dangerous pattern as a Warn detail.

#### Scenario: No workflows

- **WHEN** a dangerous-workflow probe reports not applicable
- **THEN** the result is inconclusive with reason "no workflows found"

#### Scenario: Script injection found

- **WHEN** the hasDangerousWorkflowScriptInjection probe reports a true outcome
- **THEN** the score is 0 with reason "dangerous workflow patterns detected"

#### Scenario: Clean workflows

- **WHEN** neither probe reports a true outcome
- **THEN** the score is 10

### Requirement: Dependency-Update-Tool Scoring

The system SHALL score Dependency-Update-Tool as 10 when any configured
dependency update tool is detected, and 0 otherwise.

#### Scenario: Tool detected

- **WHEN** the dependencyUpdateToolConfigured probe reports a true outcome
- **THEN** the score is 10 with reason "update tool detected"

#### Scenario: No tool

- **WHEN** no finding has a true outcome
- **THEN** the score is 0 with reason "no update tool detected"

### Requirement: Fuzzing Scoring

The system SHALL score Fuzzing as 10 when any fuzzer is detected, and 0
otherwise.

#### Scenario: Fuzzer detected

- **WHEN** the fuzzed probe reports a true outcome
- **THEN** the score is 10 with reason "project is fuzzed"

#### Scenario: Not fuzzed

- **WHEN** no finding has a true outcome
- **THEN** the score is 0 with reason "project is not fuzzed"

### Requirement: License Scoring

The system SHALL score License as 9 points for a detected license file plus 1
point for an FSF- or OSI-approved license, and 0 when no license file is
detected.

#### Scenario: Approved license file

- **WHEN** both hasLicenseFile and hasFSFOrOSIApprovedLicense report true
- **THEN** the score is 10 with reason "license file detected"

#### Scenario: Unrecognized license

- **WHEN** only hasLicenseFile reports true
- **THEN** the score is 9

#### Scenario: No license file

- **WHEN** hasLicenseFile reports false
- **THEN** the score is 0 with reason "license file not detected"

### Requirement: Maintained Scoring

The system SHALL score Maintained as 0 when the repository is archived or was
created within the last 90 days, and otherwise as the proportional score of
recent commits plus recent issue activity by project members over a target of
12 (one per week over 90 days).

#### Scenario: Archived repository

- **WHEN** the archived probe reports true
- **THEN** the score is 0 with reason "project is archived"

#### Scenario: Recently created repository

- **WHEN** the createdRecently probe reports true and the repository is not
  archived
- **THEN** the score is 0

#### Scenario: Moderate activity

- **WHEN** 4 commits and 2 issue activities are found in the last 90 days
- **THEN** the score is 5

#### Scenario: High activity

- **WHEN** 12 or more combined commits and issue activities are found
- **THEN** the score is 10

### Requirement: Packaging Scoring

The system SHALL score Packaging as 10 when a packaging workflow is detected,
and report an inconclusive result otherwise.

#### Scenario: Packaging workflow found

- **WHEN** the packagedWithAutomatedWorkflow probe reports true
- **THEN** the score is 10 with reason "packaging workflow detected"

#### Scenario: No packaging workflow

- **WHEN** no finding has a true outcome
- **THEN** the result is inconclusive with reason "packaging workflow not
  detected"

### Requirement: Pinned-Dependencies Scoring

The system SHALL score Pinned-Dependencies as a weighted proportional score of
pinned dependencies over all dependencies, grouped by dependency type, with
each type weighted equally. Within GitHub Actions, when both GitHub-owned and
third-party actions are present, third-party actions SHALL carry weight 8 and
GitHub-owned actions weight 2 of the type's 10. The system SHALL emit a Warn
detail for each unpinned dependency, an Info summary per dependency type, and
SHALL exclude dependencies whose pinning is unsupported or errored from the
score.

#### Scenario: No dependencies

- **WHEN** the pinsDependencies probe reports not applicable
- **THEN** the result is inconclusive with reason "no dependencies found"

#### Scenario: All pinned

- **WHEN** every counted dependency is pinned
- **THEN** the score is 10 with reason "all dependencies are pinned"

#### Scenario: Mixed GitHub Actions

- **GIVEN** 1 of 2 GitHub-owned actions pinned and 2 of 2 third-party actions
  pinned, and no other dependencies
- **WHEN** the check is evaluated
- **THEN** the score is 9

#### Scenario: Unpinned dependency details

- **WHEN** a dependency is not pinned
- **THEN** a Warn detail with its path, line range, snippet, and any
  remediation is emitted

#### Scenario: Unsupported dependency

- **WHEN** a finding reports pinning is not supported for a dependency
- **THEN** a Debug detail is emitted and the dependency is not scored

### Requirement: SAST Scoring

The system SHALL score SAST as follows: 10 when a SAST tool other than CodeQL
is configured; 10 when SAST runs on all analyzed pull requests; when CodeQL is
configured but not all pull requests are analyzed, the rounded-down weighted
average of 3 times the pull-request coverage score plus 7 times 10, divided by
10; when CodeQL is not configured, the pull-request coverage score. When there
are no pull requests to analyze, the system SHALL score 10 if CodeQL is
configured and 0 otherwise.

#### Scenario: Non-CodeQL tool configured

- **WHEN** the sastToolConfigured probe reports a tool such as Sonar or Snyk
- **THEN** the score is 10 with reason "SAST tool detected"

#### Scenario: CodeQL with partial coverage

- **GIVEN** CodeQL is configured and 0 of 3 pull requests were analyzed
- **WHEN** the check is evaluated
- **THEN** the score is 7

#### Scenario: No tool, partial coverage

- **GIVEN** no SAST tool is configured and 1 of 3 pull requests were analyzed
- **WHEN** the check is evaluated
- **THEN** the score is 3

#### Scenario: No pull requests and no tool

- **GIVEN** no SAST tool is configured and there are no pull requests
- **WHEN** the check is evaluated
- **THEN** the score is 0 with reason "no SAST tool detected"

### Requirement: SBOM Scoring

The system SHALL score SBOM as 0 when no SBOM file is found in the repository,
5 when an SBOM file is found in the repository but not in release artifacts,
and 10 when an SBOM file is also found in release artifacts.

#### Scenario: SBOM in repository only

- **WHEN** hasSBOM reports true and hasReleaseSBOM reports false
- **THEN** the score is 5 with reason "SBOM file found in project"

#### Scenario: SBOM in releases

- **WHEN** both hasSBOM and hasReleaseSBOM report true
- **THEN** the score is 10

#### Scenario: No SBOM file

- **WHEN** hasSBOM reports false
- **THEN** the score is 0 with reason "SBOM file not detected"

### Requirement: Security-Policy Scoring

The system SHALL score Security-Policy as 0 when no security policy file is
present, and otherwise award 6 points when the policy contains links (URLs or
email addresses), 3 points when it contains sufficient free text, and 1 point
when it mentions vulnerability disclosure.

#### Scenario: Complete policy

- **WHEN** a policy file is present and all three content probes report true
- **THEN** the score is 10 with reason "security policy file detected"

#### Scenario: Policy with links only

- **WHEN** a policy file is present and only securityPolicyContainsLinks
  reports true
- **THEN** the score is 6

#### Scenario: No policy

- **WHEN** securityPolicyPresent reports false
- **THEN** the score is 0 with reason "security policy file not detected"

### Requirement: Signed-Releases Scoring

The system SHALL score Signed-Releases per release as 10 for a release with
provenance, 8 for a signed release without provenance, and 0 otherwise, and
report the rounded-down average over the evaluated releases. Each release name
SHALL be logged once as a Debug detail, and a missing-signature warning SHALL
be suppressed for releases that have provenance.

#### Scenario: No releases

- **WHEN** a release probe reports not applicable
- **THEN** the result is inconclusive with reason "no releases found"

#### Scenario: Mixed releases

- **GIVEN** 3 releases where one has provenance, one is signed only, and one
  has neither
- **WHEN** the check is evaluated
- **THEN** the score is 6

#### Scenario: Nothing signed

- **WHEN** no release is signed or has provenance
- **THEN** the score is 0

#### Scenario: Too many releases

- **WHEN** findings cover more than 5 releases
- **THEN** the check result is a runtime error result

### Requirement: Token-Permissions Scoring

The system SHALL score Token-Permissions starting from 10 and deducting for
GitHub workflow token permissions, per workflow file:

- Top-level write to `contents`, `packages`, or `actions` deducts 10;
  `deployments` or `security-events` deducts 1; `checks` or `statuses` deducts
  0.5.
- Top-level `write-all` deducts 0.5, or yields 0 if the same workflow also has
  job-level write or undeclared job-level permissions.
- Undeclared top-level permissions deduct 0.5, or yield 0 if the same workflow
  has undeclared job-level permissions.
- Job-level write yields 0 if the same workflow has top-level write, or deducts
  0.5 if top-level permissions are undeclared.

The score SHALL be floored at 0 and truncated to an integer. Read and none
permissions SHALL be logged as Info details and write or undeclared top-level
permissions as Warn details.

#### Scenario: Least privilege

- **WHEN** all workflows declare read-only or no permissions
- **THEN** the score is 10 with reason "GitHub workflow tokens follow principle
  of least privilege"

#### Scenario: Top-level contents write

- **WHEN** a workflow grants top-level `contents: write`
- **THEN** the score is 0

#### Scenario: Top-level checks and statuses write

- **WHEN** a workflow grants top-level `checks: write` and `statuses: write`
- **THEN** the score is 9

#### Scenario: Undeclared permissions

- **WHEN** a workflow declares no permissions at either the top level or job
  level
- **THEN** the score is 0

#### Scenario: No tokens found

- **WHEN** the permission probes report not applicable
- **THEN** the result is inconclusive with reason "No tokens found"

### Requirement: Vulnerabilities Scoring

The system SHALL score Vulnerabilities as 10 minus the number of known
vulnerabilities found by the hasOSVVulnerabilities probe, with a minimum of 0,
logging each vulnerability as a Warn detail.

#### Scenario: No vulnerabilities

- **WHEN** no vulnerabilities are found
- **THEN** the score is 10 with reason "0 existing vulnerabilities detected"

#### Scenario: Several vulnerabilities

- **WHEN** 4 vulnerabilities are found
- **THEN** the score is 6

#### Scenario: Many vulnerabilities

- **WHEN** 15 vulnerabilities are found
- **THEN** the score is 0

### Requirement: Webhooks Scoring

The system SHALL score Webhooks as 10 when the repository has no webhooks or all
webhooks have a secret, 0 when no webhook has a secret, and otherwise the
proportional score of webhooks with a secret over all webhooks.

#### Scenario: No webhooks

- **WHEN** the webhooksUseSecrets probe reports not applicable
- **THEN** the score is 10 with reason "project does not have webhook"

#### Scenario: Some webhooks without secrets

- **WHEN** 1 of 3 webhooks lacks a secret
- **THEN** the score is 6

#### Scenario: No webhook has a secret

- **WHEN** every webhook lacks a secret
- **THEN** the score is 0 with reason "no hook(s) have a secret configured"
