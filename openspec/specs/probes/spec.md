# probes Specification

## Purpose

Probes are the second tier of Scorecard. Each probe answers one yes-or-no
question about a repository by turning collected raw data into a list of
findings with outcomes, messages, locations, and remediation. This spec
covers the probe contract, the finding model, probe definition metadata,
how probes are registered and grouped under checks, how to run selected
probes directly, and the current probe catalog.

## Requirements

### Requirement: Probe Contract

Each probe SHALL take the raw results of a scan and return a list of
findings, the probe's name, and an error. A probe SHALL return at least one
finding whenever it succeeds. Every finding SHALL carry the probe's name.
A probe SHALL return an error when it receives no raw results or cannot
build a finding.

#### Scenario: Probe with nothing to report

- **GIVEN** raw data in which no binary files were found
- **WHEN** the `hasBinaryArtifacts` probe runs
- **THEN** it returns a single finding with outcome `False`

#### Scenario: Missing raw results

- **WHEN** a probe is called without raw results
- **THEN** it returns an error and no findings

### Requirement: Finding Outcomes

Every finding SHALL have exactly one of these outcomes: `True` (the answer
to the probe's question is yes), `False` (the answer is no),
`NotApplicable` (the finding should not count in evaluation),
`NotAvailable` (data needed to answer was unavailable, often because an API
returned no answer), `NotSupported` (the input is of a kind the probe does
not support), or `Error` (the probe could not determine a result). Parsing
any other outcome name from a definition file SHALL fail.

#### Scenario: Unknown branch protection setting

- **GIVEN** a branch whose force-push setting could not be read
- **WHEN** the `blocksForcePushOnBranches` probe runs
- **THEN** the finding for that branch has outcome `NotAvailable`

#### Scenario: Unsupported dependency entry

- **GIVEN** pinning raw data containing an entry with only a message
- **WHEN** the `pinsDependencies` probe runs
- **THEN** the finding for that entry has outcome `NotSupported`

### Requirement: Finding Fields

A finding SHALL contain the probe name, a human-readable message, and an
outcome. It MAY contain a location (path, file type, start line, end line,
snippet), a remediation (text, Markdown, effort, and optional patch), and a
map of string values. In JSON, these SHALL be the `probe`, `message`,
`outcome`, `location`, `remediation`, and `values` keys. Location and
remediation SHALL be omitted when absent, and the file type SHALL be
written as a number (0 none, 1 source, 2 binary, 3 text, 4 URL, 5 verified
binary).

#### Scenario: Finding with a location

- **GIVEN** a probe finding for a source file at line 12
- **WHEN** the finding is written as JSON
- **THEN** it contains `location.path`, `location.type` set to `1`, and
  `location.lineStart` set to `12`

### Requirement: Finding Values

Probes SHALL use the values map for machine-readable detail. The keys in use
are: `branchName` for branch protection probes;
`numberOfRequiredReviewers` for `requiresApproversForPullRequests`; `tool`
for `fuzzed`, `sastToolConfigured`, and `dependencyUpdateToolConfigured`;
`badgeLevel` for `hasOpenSSFBadge`; `dependencyType` for
`pinsDependencies`; `permissionLevel` and `tokenName` for token permission
probes; `releaseName` and `assetName` for release probes; `assetName` and
`assetURL` for `hasReleaseSBOM`; `approvedChangesets` and
`totalChangesets` for `codeApproved`; `totalPullRequestsMerged` and
`totalPullRequestsAnalyzed` for `sastToolRunsOnAllCommits`;
`commitsWithinThreshold`, `numberOfIssuesUpdatedWithinThreshold`, and
`lookBackDays` for maintenance probes.

#### Scenario: Pinned dependency type

- **GIVEN** an unpinned `FROM python:3.12` line in a Dockerfile
- **WHEN** the `pinsDependencies` probe runs
- **THEN** the `False` finding has value `dependencyType` set to
  `containerImage`

#### Scenario: Recent commits count

- **GIVEN** 12 commits in the last 90 days
- **WHEN** the `hasRecentCommits` probe runs
- **THEN** the `True` finding has `commitsWithinThreshold` set to `12` and
  `lookBackDays` set to `90`

### Requirement: Remediation Attachment

A new finding SHALL start with outcome `False` and the remediation from the
probe's definition file. When the outcome is set, the remediation SHALL be
kept only if the outcome equals the definition's `onOutcome` value and
SHALL be removed otherwise. A probe MAY replace the remediation with a more
specific one, as `pinsDependencies` does for unpinned GitHub Actions and
container images.

#### Scenario: Good outcome drops remediation

- **GIVEN** `securityPolicyPresent` has `onOutcome: False`
- **WHEN** it emits a finding with outcome `True`
- **THEN** the finding has no remediation

#### Scenario: Bad outcome keeps remediation

- **GIVEN** `hasDangerousWorkflowScriptInjection` has `onOutcome: True`
- **WHEN** it emits a finding with outcome `True`
- **THEN** the finding carries the remediation from its definition

### Requirement: Remediation Templating

The system SHALL replace `${{ finding.location.path }}` in remediation text
and Markdown with the finding's location path when a location is set, and
SHALL replace `${{ metadata.<key> }}` with values the probe supplies. Scan
metadata available for this SHALL include `repository.host`,
`repository.name`, `repository.uri`, `repository.sha1`,
`repository.defaultBranch`, and `localPath`. A probe MAY attach a
machine-applicable patch, as `hasDangerousWorkflowScriptInjection` does.

#### Scenario: OSV identifier in remediation

- **GIVEN** a vulnerability with OSV id `GHSA-xxxx-xxxx-xxxx`
- **WHEN** the `hasOSVVulnerabilities` probe emits its finding
- **THEN** the message links to `https://osv.dev/GHSA-xxxx-xxxx-xxxx`
- **AND** the remediation text has the id substituted for its placeholder

### Requirement: Probe Definition File

Each probe SHALL have a `def.yml` with these fields: `id`, `lifecycle`,
`short`, `motivation`, `implementation`, `outcome` (a list of sentences
describing possible outcomes), `remediation` (with `onOutcome`, `effort`,
and `text`, and optionally `markdown`), and optionally `ecosystem` (with
`languages` and `clients`). Loading a definition SHALL fail when `id` does
not match the probe's name, `onOutcome` is not a valid outcome, `effort`
is not `Low`, `Medium`, or `High`, `lifecycle` is not `experimental`,
`stable`, or `deprecated`, or an ecosystem language or client is not
supported. Because every finding is built from the definition, an invalid
definition SHALL make the probe fail when it runs.

#### Scenario: Mismatched identifier

- **GIVEN** a `def.yml` whose `id` differs from the probe's name
- **WHEN** the probe tries to create a finding
- **THEN** the probe returns an invalid-definition error

#### Scenario: Missing effort

- **GIVEN** a `def.yml` with no `remediation.effort`
- **WHEN** the definition is loaded
- **THEN** loading fails as invalid

### Requirement: Probe Lifecycle

The system SHALL mark each probe as `experimental` (semantics may change),
`stable` (semantics will not change apart from bug fixes), or `deprecated`
(no longer maintained). Twelve probes SHALL currently be experimental:
`codeApproved`, `codeReviewOneReviewers`, `contributorsFromOrgOrCompany`,
`hasNoGitHubWorkflowPermissionUnknown`, `hasReleaseSBOM`, `hasSBOM`,
`jobLevelPermissions`, `releasesHaveVerifiedProvenance`,
`sastToolRunsOnAllCommits`, `topLevelPermissions`, `unsafeblock`, and
`webhooksUseSecrets`. All other probes SHALL be stable.

#### Scenario: Experimental probe documented

- **WHEN** the probe documentation is generated
- **THEN** `unsafeblock` is listed with lifecycle `experimental`

### Requirement: Remediation Effort

Each probe SHALL declare a remediation effort of `Low`, `Medium`, or
`High`. In JSON output the effort SHALL be written as a number: 1 for low,
2 for medium, and 3 for high.

#### Scenario: Effort in output

- **GIVEN** `fuzzed` declares effort `High`
- **WHEN** a `False` finding from `fuzzed` is written as JSON
- **THEN** its `remediation.effort` is `3`

### Requirement: Ecosystem Metadata

A probe's `ecosystem.languages` SHALL use only supported language names:
`go`, `python`, `javascript`, `c++`, `c`, `typescript`, `java`, `c#`,
`ruby`, `php`, `starlark`, `scala`, `kotlin`, `swift`, `rust`, `haskell`,
`dockerfile`, `objectivec`, `f#`, or `all`. Its `ecosystem.clients` SHALL
use only `github`, `gitlab`, or `localdir`. The ecosystem block SHALL be
optional; nine probes currently omit it.

#### Scenario: Unsupported client name

- **GIVEN** a `def.yml` listing client `bitbucket`
- **WHEN** the definition is loaded
- **THEN** loading fails as invalid

### Requirement: Probe Registration

Each probe SHALL register itself by name when its package loads, either with
an implementation and the checks whose raw data it needs, or as an
independent probe with an implementation that takes the scan request.
Registration SHALL fail if the name is empty, no implementation is given,
or a raw-data probe lists no required checks. Looking up a name that is not
registered SHALL fail with a "probe not found" error.

#### Scenario: Probe declares its raw data

- **WHEN** the `hasPermissiveLicense` probe is registered
- **THEN** it declares that it needs the License raw data

#### Scenario: Unknown probe name

- **WHEN** a lookup is made for probe `noSuchProbe`
- **THEN** the system fails with `probe "noSuchProbe" not found`

### Requirement: Check Probe Groups

Each check SHALL collect its raw data, run a fixed list of probes over it in
order, and pass the combined findings to its evaluation. When a probe
returns an error, the system SHALL add an `Error` finding for that probe
with an internal-error message, continue with the remaining probes, and
then report a runtime error for the check.

#### Scenario: Check runs its probes

- **WHEN** the Security-Policy check runs
- **THEN** it runs `securityPolicyPresent`, `securityPolicyContainsLinks`,
  `securityPolicyContainsVulnerabilityDisclosure`, and
  `securityPolicyContainsText`, in that order

#### Scenario: Probe failure inside a check

- **GIVEN** one probe in a check's list returns an error
- **WHEN** the check runs
- **THEN** the check result is a runtime error

### Requirement: Independent Probes

The system SHALL support independent probes that collect their own data from
the scan request instead of using a check's raw data. `unsafeblock` SHALL be
the only independent probe. It SHALL inspect the repository languages and,
for Go, Java, and C#, report a `True` finding for each `unsafe` import in a
Go file, each reference to `sun.misc.Unsafe` or `jdk.internal.misc.Unsafe`
in Java, and each `.csproj` that enables `AllowUnsafeBlocks`. Files that
cannot be parsed SHALL produce `Error` findings. When no non-error finding
results, it SHALL add one `False` finding.

#### Scenario: Go unsafe import

- **GIVEN** a Go file that imports `unsafe`
- **WHEN** `unsafeblock` runs
- **THEN** it returns a `True` finding with the file and import line

#### Scenario: No unsafe code

- **GIVEN** a Go repository with no `unsafe` imports
- **WHEN** `unsafeblock` runs
- **THEN** it returns a single `False` finding

### Requirement: Running Selected Probes

When the user names probes with `--probes`, the system SHALL run only those
probes and no checks. It SHALL first collect the raw data each named probe
requires, once per check, then run the probes in the order given and return
their findings without check scores. An unknown probe name SHALL fail the
run. A probe error SHALL end the run with an internal error.

Probes that are not registered, including
`hasNoGitHubWorkflowPermissionUnknown`, `jobLevelPermissions`, and
`topLevelPermissions`, SHALL NOT be runnable this way. Raw data collected
this way SHALL not require `SCORECARD_EXPERIMENTAL`, and Packaging raw data
SHALL only be available for GitHub and GitLab repositories.

#### Scenario: Two probes sharing raw data

- **WHEN** a user runs `scorecard --repo=github.com/owner/repo
  --probes=hasLicenseFile,hasFSFOrOSIApprovedLicense`
- **THEN** License raw data is collected once
- **AND** the result contains findings from both probes and no check results

#### Scenario: Token permission probe requested

- **WHEN** a user runs `scorecard --repo=github.com/owner/repo
  --probes=topLevelPermissions`
- **THEN** the run fails with `probe "topLevelPermissions" not found`

#### Scenario: Packaging probe on a local directory

- **WHEN** a user runs `scorecard --local=.
  --probes=packagedWithAutomatedWorkflow`
- **THEN** the run fails because only GitHub and GitLab are supported

### Requirement: Probe Output Format

With `--format probe`, the system SHALL write one JSON object per
repository containing `date` (YYYY-MM-DD), `repo` with `name` and `commit`,
`scorecard` with `version` and `commit`, and `findings`, the list of all
findings from the run.

#### Scenario: Probe JSON output

- **WHEN** a user runs `scorecard --repo=github.com/owner/repo
  --probes=fuzzed --format probe`
- **THEN** the output has `date`, `repo`, `scorecard`, and `findings` keys
- **AND** every finding has `probe` set to `fuzzed`

### Requirement: Outcome Conventions

Probes SHALL follow these conventions. A probe that evaluates a set of items
(branches, releases, workflows, dependencies, files, tools, webhooks) SHALL
emit one finding per item. When the set is empty because the subject does
not exist (no branches, releases, pull requests, changesets, webhooks,
workflows, or dependencies), the probe SHALL emit a single `NotApplicable`
finding. When a probe looks for the presence of something and finds
nothing, it SHALL emit a single `False` finding, or `True` for probes whose
question names a risk. When an individual setting could not be read, the
finding for that item SHALL be `NotAvailable`.

#### Scenario: No releases

- **GIVEN** a repository with no releases
- **WHEN** the `releasesAreSigned` probe runs
- **THEN** it returns a single `NotApplicable` finding

#### Scenario: One finding per branch

- **GIVEN** raw data for branches `main` and `release-1.x`
- **WHEN** the `branchesAreProtected` probe runs
- **THEN** it returns one finding per branch, each with a `branchName`
  value

### Requirement: Probe Catalog by Check

The system SHALL provide 48 probes. The checks SHALL run these probes:

- Binary-Artifacts: `hasUnverifiedBinaryArtifacts`
- Branch-Protection: `blocksDeleteOnBranches`,
  `blocksForcePushOnBranches`, `branchesAreProtected`,
  `branchProtectionAppliesToAdmins`, `dismissesStaleReviews`,
  `requiresApproversForPullRequests`, `requiresCodeOwnersReview`,
  `requiresLastPushApproval`, `requiresUpToDateBranches`,
  `runsStatusChecksBeforeMerging`, `requiresPRsToChangeCode`
- CI-Tests: `testsRunInCI`
- CII-Best-Practices: `hasOpenSSFBadge`
- Code-Review: `codeApproved`
- Contributors: `contributorsFromOrgOrCompany`
- Dangerous-Workflow: `hasDangerousWorkflowScriptInjection`,
  `hasDangerousWorkflowUntrustedCheckout`
- Dependency-Update-Tool: `dependencyUpdateToolConfigured`
- Fuzzing: `fuzzed`
- License: `hasLicenseFile`, `hasFSFOrOSIApprovedLicense`
- Maintained: `archived`, `hasRecentCommits`,
  `issueActivityByProjectMember`, `createdRecently`
- Packaging: `packagedWithAutomatedWorkflow`
- Pinned-Dependencies: `pinsDependencies`
- SAST: `sastToolConfigured`, `sastToolRunsOnAllCommits`
- SBOM: `hasSBOM`, `hasReleaseSBOM`
- Security-Policy: `securityPolicyPresent`, `securityPolicyContainsLinks`,
  `securityPolicyContainsVulnerabilityDisclosure`,
  `securityPolicyContainsText`
- Signed-Releases: `releasesAreSigned`, `releasesHaveProvenance`
- Token-Permissions: `hasNoGitHubWorkflowPermissionUnknown`,
  `jobLevelPermissions`, `topLevelPermissions`
- Vulnerabilities: `hasOSVVulnerabilities`
- Webhooks: `webhooksUseSecrets`

#### Scenario: Maintained check probes

- **WHEN** the Maintained check runs
- **THEN** it runs `archived`, `hasRecentCommits`,
  `issueActivityByProjectMember`, and `createdRecently`

#### Scenario: Branch-Protection check probes

- **WHEN** the Branch-Protection check runs
- **THEN** it produces findings from the eleven branch protection probes

### Requirement: Probes Outside Checks

The system SHALL provide probes that no check runs but that can be run
with `--probes`: `hasPermissiveLicense` (License raw data),
`codeReviewOneReviewers` (Code-Review raw data), `hasBinaryArtifacts`
(Binary-Artifacts raw data, counting verified Gradle wrappers too),
`releasesHaveVerifiedProvenance` (Signed-Releases raw data, using package
provenance from deps.dev), and the independent probe `unsafeblock`.

#### Scenario: Stricter binary probe

- **GIVEN** a repository whose only binary is a verified
  `gradle-wrapper.jar`
- **WHEN** the user runs `--probes=hasBinaryArtifacts`
- **THEN** a `True` finding is reported for the jar
- **AND** `hasUnverifiedBinaryArtifacts` would report `False`

### Requirement: Selected Probe Semantics

The probes SHALL apply these thresholds and rules:

- `hasRecentCommits`, `issueActivityByProjectMember`, and `createdRecently`
  SHALL look back 90 days.
- `contributorsFromOrgOrCompany` SHALL count only contributors with at
  least 5 contributions and emit one `True` finding per organization or
  company.
- `codeApproved` SHALL skip approved changesets authored by bots, return
  `False` when any remaining changeset is unapproved, and return
  `NotApplicable` when no changesets exist or none were authored by humans.
- `releasesAreSigned` and `releasesHaveProvenance` SHALL examine those of
  the 5 most recent releases that have assets. A release SHALL be signed
  when an asset
  ends in `.asc`, `.minisig`, `.sig`, `.sign`, `.sigstore`, or
  `.sigstore.json`, and SHALL have provenance when an asset ends in
  `.intoto.jsonl`.
- `requiresCodeOwnersReview` SHALL return `False` for a branch that
  requires code owner review when the repository has no `CODEOWNERS` file.
- `hasOSVVulnerabilities` SHALL group vulnerabilities that share aliases and
  emit one `True` finding per group.

#### Scenario: Release with a signature

- **GIVEN** the latest release has an asset `app.tar.gz.sig`
- **WHEN** the `releasesAreSigned` probe runs
- **THEN** a `True` finding is returned with `releaseName` and `assetName`

#### Scenario: Code owner review without a CODEOWNERS file

- **GIVEN** `main` requires code owner review
- **AND** the repository has no `CODEOWNERS` file
- **WHEN** the `requiresCodeOwnersReview` probe runs
- **THEN** the finding for `main` has outcome `False`

### Requirement: Probe Documentation

The system SHALL generate `docs/probes.md` from every probe's `def.yml`,
listing for each probe its lifecycle, short description, motivation,
implementation, and outcomes. The system SHALL provide a `make` target that
scaffolds a new probe directory containing `def.yml`, `impl.go`, and
`impl_test.go`.

#### Scenario: Regenerating documentation

- **WHEN** a contributor changes a probe's `def.yml` and regenerates docs
- **THEN** `docs/probes.md` reflects the new definition text
