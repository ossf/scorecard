# cli Specification

## Purpose

The `scorecard` command-line tool runs Scorecard security checks or probes
against one or more source repositories and writes the results to standard
output or a file. This spec covers the command tree, input selection, flags,
feature-flag environment variables, validation errors, check selection,
progress messages, logging verbosity, and exit status.

## Requirements

### Requirement: Command Tree

The system SHALL provide a root `scorecard` command that runs a scan, a
`serve` subcommand that starts the HTTP server, and a `version` subcommand
that prints build information. The scan flags SHALL be accepted only by the
root command; subcommands SHALL reject them as unknown flags.

#### Scenario: Root command runs a scan

- **WHEN** a user runs `scorecard --repo=github.com/owner/repo`
- **THEN** the system scans that repository and prints the results

#### Scenario: Scan flags are rejected by subcommands

- **WHEN** a user runs `scorecard serve --verbosity=debug`
- **THEN** the system fails with `unknown flag: --verbosity`

### Requirement: Exactly One Input Source

The system MUST require exactly one of `--repo`, `--repos`, `--org`,
`--local`, `--npm`, `--pypi`, `--rubygems`, or `--nuget`. When zero or more
than one is set, the system MUST fail validation with the message
"exactly one of `repo`, `repos`, `org`, `npm`, `pypi`, `rubygems`, `nuget`
or `local` must be set".

#### Scenario: No input source

- **WHEN** a user runs `scorecard` with no input flag
- **THEN** the system reports the "exactly one of" validation error
- **AND** prints usage and exits with a non-zero status

#### Scenario: Two input sources

- **WHEN** a user runs `scorecard --repo=owner/repo --local=.`
- **THEN** the system reports the "exactly one of" validation error and exits
  with a non-zero status

### Requirement: Validation Error Reporting

The system SHALL validate all options before scanning and SHALL report every
failed validation together, prefixed with
"validating options: some options could not be validated". On a validation
failure the system SHALL print command usage. Once validation succeeds, the
system SHALL NOT print usage for later runtime errors.

#### Scenario: Multiple validation failures

- **WHEN** a user sets no input source and `--file-mode=svn`
- **THEN** the error lists both the input-source error and
  "unsupported file mode"

#### Scenario: Runtime error after validation

- **GIVEN** options that pass validation
- **WHEN** check selection fails because a check name is invalid
- **THEN** the system prints the error without printing usage

### Requirement: Remote Repository Input

The system SHALL accept a repository via `--repo` in the forms `owner/repo`,
`github.com/owner/repo`, or a full URL. The system SHALL try to interpret
the value as a GitHub repository first and as a GitLab repository otherwise.
When `SCORECARD_EXPERIMENTAL` is set, the system SHALL also try Azure DevOps
between GitHub and GitLab, and SHALL fail immediately for a value on an
Azure DevOps host that cannot be parsed.

#### Scenario: GitHub shorthand

- **WHEN** a user runs `scorecard --repo=ossf/scorecard`
- **THEN** the system scans `github.com/ossf/scorecard`

#### Scenario: Unparseable repository

- **WHEN** a user runs `scorecard --repo=::bad::`
- **THEN** the system prints `Skipping ::bad::: unable to parse as github,
  gitlab, or azuredevops: ...` to standard error

#### Scenario: Azure DevOps requires the experimental flag

- **GIVEN** `SCORECARD_EXPERIMENTAL` is set
- **WHEN** a user passes a `dev.azure.com` or legacy `visualstudio.com`
  repository URL
- **THEN** the system scans it as an Azure DevOps repository

### Requirement: Multiple Repository Input

The system SHALL accept a comma-separated list of repositories via
`--repos`, trimming whitespace and ignoring empty entries. The system SHALL
scan each repository in order and SHALL continue with the remaining
repositories when one cannot be parsed or scanned.

#### Scenario: Scan a list

- **WHEN** a user runs `scorecard --repos=ossf/scorecard,ossf/scorecard-action`
- **THEN** the system scans both repositories in that order

#### Scenario: One repository fails

- **GIVEN** the second of three repositories cannot be scanned
- **WHEN** the scan runs
- **THEN** the system prints `Skipping <uri>: <error>` to standard error for
  that repository
- **AND** still scans the third repository

### Requirement: Organization Input

The system SHALL accept a GitHub organization via `--org` as `owner`,
`github.com/owner`, `http://github.com/owner`, or `https://github.com/owner`.
The system SHALL scan every non-archived repository in that organization,
following API pagination. When the organization's repositories cannot be
listed, the system MUST fail with a non-zero exit status.

#### Scenario: Scan an organization

- **WHEN** a user runs `scorecard --org=github.com/ossf`
- **THEN** the system scans every non-archived repository owned by `ossf`
- **AND** skips archived repositories

#### Scenario: Organization listing fails

- **WHEN** listing repositories for the organization returns an error
- **THEN** the system fails with
  `listing repositories for org "<org>": ...` and a non-zero exit status

### Requirement: Local Directory Input

The system SHALL accept a local folder via `--local` and scan its files
without contacting a repository host for repository data.

#### Scenario: Scan a local folder

- **WHEN** a user runs `scorecard --local=./project --checks=License`
- **THEN** the system runs the License check against the files in
  `./project`

### Requirement: Package Manager Input

The system SHALL accept a package name via `--npm`, `--pypi`, `--rubygems`,
or `--nuget` and SHALL scan the source repository that the package registry
associates with that package. Failure to resolve the package MUST stop the
command with a non-zero exit status. Resolution rules are defined in the
package manager resolution spec.

#### Scenario: Scan an npm package

- **WHEN** a user runs `scorecard --npm=left-pad`
- **THEN** the system resolves the package's source repository and scans it

#### Scenario: Resolution fails

- **WHEN** the registry lookup for the package fails
- **THEN** the system fails with an error naming the package manager
  resolution step and exits with a non-zero status

### Requirement: Default Check Set

When neither `--checks` nor `--policy` selects checks, the system SHALL run
every registered check that supports the scan mode. The experimental checks
`Webhooks` and `SBOM` SHALL be excluded from the default set unless
`SCORECARD_EXPERIMENTAL` is set.

#### Scenario: Default remote scan

- **GIVEN** `SCORECARD_EXPERIMENTAL` is not set
- **WHEN** a user scans a remote repository without `--checks`
- **THEN** the system runs all checks except `Webhooks` and `SBOM`

#### Scenario: Experimental checks enabled

- **GIVEN** `SCORECARD_EXPERIMENTAL` is set
- **WHEN** a user scans a remote repository without `--checks`
- **THEN** the default set also includes `Webhooks` and `SBOM`

### Requirement: Explicit Check Selection

The system SHALL run only the checks named in the comma-separated
`--checks` flag. Check names SHALL match registered checks case-insensitively
for remote scans at `HEAD`. Experimental checks SHALL be selectable by name
without `SCORECARD_EXPERIMENTAL`. An unknown check name MUST fail the command
with `invalid check: <name>` and a non-zero exit status.

#### Scenario: Select two checks

- **WHEN** a user runs `scorecard --repo=owner/repo --checks=License,SAST`
- **THEN** the system runs only `License` and `SAST`

#### Scenario: Unknown check

- **WHEN** a user runs `scorecard --repo=owner/repo --checks=Nope`
- **THEN** the system fails with `invalid check: Nope` and a non-zero exit
  status

### Requirement: Local Mode Check Restrictions

In `--local` mode the system SHALL run only checks that can operate on files
alone: `Binary-Artifacts`, `Dangerous-Workflow`, `Dependency-Update-Tool`,
`Fuzzing`, `License`, `Packaging`, `Pinned-Dependencies`, `SAST`,
`Security-Policy`, `Token-Permissions`, and `Vulnerabilities`. The default
set SHALL silently omit other checks. Explicitly naming any other check with
`--checks` MUST fail with `Unsupported RequestType ... by check: <name>`. In
this mode check names given to `--checks` MUST match the registered name's
exact case.

#### Scenario: Default local scan

- **WHEN** a user runs `scorecard --local=.`
- **THEN** the system runs only the file-based checks listed above

#### Scenario: Unsupported check in local mode

- **WHEN** a user runs `scorecard --local=. --checks=Code-Review`
- **THEN** the system fails with
  `Unsupported RequestType [0] by check: Code-Review` and a non-zero exit
  status

#### Scenario: Wrong case in local mode

- **WHEN** a user runs `scorecard --local=. --checks=license`
- **THEN** the system fails with an `Unsupported RequestType` error for
  `license`

### Requirement: Commit Selection

The system SHALL analyze the commit given by `--commit`, defaulting to
`HEAD`. An empty `--commit` value MUST fail validation with
"commit should be non-empty". When the commit is not `HEAD`
(case-insensitive), the system SHALL run only checks that support
commit-based analysis: `Binary-Artifacts`, `CI-Tests`, `Code-Review`,
`Dangerous-Workflow`, `License`, `Pinned-Dependencies`, `Security-Policy`,
`Token-Permissions`, and `Vulnerabilities`; explicitly naming any other
check MUST fail with an `Unsupported RequestType` error.

#### Scenario: Scan a specific commit

- **WHEN** a user runs `scorecard --repo=owner/repo --commit=<sha>`
- **THEN** the system runs only commit-capable checks at `<sha>`

#### Scenario: Empty commit

- **WHEN** a user runs `scorecard --repo=owner/repo --commit=`
- **THEN** validation fails with "commit should be non-empty"

### Requirement: Commit Depth

The system SHALL limit commit-history analysis to the number of commits
given by `--commit-depth`, counting back from the analyzed commit. A value of
zero or less SHALL use the default depth of 30 commits.

#### Scenario: Default depth

- **WHEN** a user omits `--commit-depth`
- **THEN** commit-history checks examine up to the 30 most recent commits

#### Scenario: Custom depth

- **WHEN** a user runs with `--commit-depth=100`
- **THEN** commit-history checks examine up to 100 commits

### Requirement: Probe Execution

When `--probes` names one or more probes, the system SHALL run those probes
instead of checks, collecting only the raw data the probes require, and
SHALL report their findings. An unknown probe name SHALL cause that
repository to be skipped with `getting probe "<name>": ... not found`.

#### Scenario: Run a probe

- **WHEN** a user runs
  `scorecard --repo=owner/repo --probes=hasLicenseFile --format=probe`
- **THEN** the system outputs the findings for `hasLicenseFile` and runs no
  checks

#### Scenario: Unknown probe

- **WHEN** a user runs `scorecard --local=. --probes=nope`
- **THEN** the system prints `Skipping <path>: run: getting probe "nope": ...`

### Requirement: Output Format Selection

The system SHALL accept `--format` values `default`, `json`, `probe`, and
`intoto`, with `default` as the default. The value `sarif` SHALL be accepted
only when `ENABLE_SARIF` is set and the value `raw` only when `SCORECARD_V6`
is set. Any other value MUST fail validation with "unsupported format";
`sarif` without `ENABLE_SARIF` MUST fail with
"SARIF format is not supported yet", and `raw` without `SCORECARD_V6` MUST
fail with "raw option is not supported yet".

#### Scenario: JSON output

- **WHEN** a user runs `scorecard --repo=owner/repo --format=json`
- **THEN** the system writes the results as JSON

#### Scenario: SARIF without feature flag

- **GIVEN** `ENABLE_SARIF` is not set
- **WHEN** a user runs with `--format=sarif`
- **THEN** validation fails with "SARIF format is not supported yet"

#### Scenario: Unknown format

- **WHEN** a user runs with `--format=xml`
- **THEN** validation fails with "unsupported format"

### Requirement: Details and Annotations

The system SHALL include per-check detail lines in `default`, `json`,
`intoto`, and `sarif` output when `--show-details` is set, filtered by the
log level from `--verbosity`. The system SHALL include maintainer
annotations from the repository's Scorecard configuration file in
`default`, `json`, and `intoto` output when `--show-annotations` is set.

#### Scenario: Show details

- **WHEN** a user runs with `--show-details`
- **THEN** each check result includes its detail lines

#### Scenario: Details hidden by default

- **WHEN** a user omits `--show-details`
- **THEN** the output contains scores and reasons without detail lines

### Requirement: Output Destination

The system SHALL write results to standard output unless `--output` (short
form `-o`) names a file. When a file is named, the system SHALL create or
truncate it, write the results there, and print
`Writing results to <file>` to standard error. Failure to create or write
the file SHALL be reported as
`Failed to format results for <uri>: ...` on standard error.

#### Scenario: Write to a file

- **WHEN** a user runs with `-o results.json --format=json`
- **THEN** the JSON results are written to `results.json`
- **AND** `Writing results to results.json` is printed to standard error

### Requirement: Policy File

When `ENABLE_SARIF` is set, the system SHALL register a `--policy` flag that
reads a YAML policy file with `version: 1` and a `policies` map of check
name to `mode` (`enforced` or `disabled`) and `score` (0 to 10). The policy
SHALL select the checks to run when `--checks` is not given, skipping checks
the scan mode cannot run. Every check to run MUST have a policy entry. A
policy file that cannot be read or fails validation MUST stop the command
with a non-zero exit status. Without `ENABLE_SARIF`, `--policy` SHALL be an
unknown flag.

#### Scenario: Policy selects checks

- **GIVEN** `ENABLE_SARIF` is set
- **WHEN** a user runs with `--policy=policy.yml --format=sarif` and no
  `--checks`
- **THEN** the system runs only the checks listed in the policy

#### Scenario: Invalid policy

- **GIVEN** a policy file with `version: 2`, an unknown check name, a mode
  other than `enforced` or `disabled`, a score outside 0 to 10, or a check
  listed twice
- **WHEN** the command runs
- **THEN** it fails with a `readPolicy` error and a non-zero exit status

#### Scenario: Selected check lacks a policy

- **GIVEN** `ENABLE_SARIF` is set and the policy omits `SAST`
- **WHEN** a user runs with `--policy=policy.yml --checks=SAST`
- **THEN** the command fails with "checks don't have policies"

### Requirement: File Fetch Mode

The system SHALL accept `--file-mode` values `archive` (default) and `git`,
case-insensitively. In `git` mode the system SHALL fetch repository files
with git instead of the host's export archive. Any other value MUST fail
validation with "unsupported file mode".

#### Scenario: Git file mode

- **WHEN** a user runs with `--file-mode=git`
- **THEN** repository files are fetched with git

#### Scenario: Invalid file mode

- **WHEN** a user runs with `--file-mode=svn`
- **THEN** validation fails with "unsupported file mode"

### Requirement: Result Metadata

The system SHALL append each comma-separated value of `--metadata` to the
metadata of every result it produces.

#### Scenario: Attach metadata

- **WHEN** a user runs with `--metadata=team=infra,tier=1`
- **THEN** each result's metadata includes `team=infra` and `tier=1`

### Requirement: Progress Messages

In `default` format only, the system SHALL print
`Starting (<uri>) [<check>]` for each enabled check before a scan and
`Finished (<uri>) [<check>]` after it, followed by a `RESULTS` banner, all
to standard error. When probes are selected, the messages SHALL read
`Starting (<uri>) probe [<probe>]` and `Finished (<uri>) probe <probe>`
without the banner. Results SHALL be sorted by check name.

#### Scenario: Default format banners

- **WHEN** a user runs `scorecard --local=. --checks=License`
- **THEN** standard error shows `Starting (.) [License]`,
  `Finished (.) [License]`, and `RESULTS`

#### Scenario: Machine-readable format has no banners

- **WHEN** a user runs with `--format=json`
- **THEN** no `Starting` or `Finished` messages are printed

### Requirement: Exit Status

The system MUST exit with status 1 and print
`error during command execution: <error>` when validation fails, the policy
file is invalid, check selection fails, the organization cannot be listed,
or package resolution fails. When a repository cannot be parsed or scanned,
the system SHALL print `Skipping <uri>: <error>` to standard error and
continue without changing the exit status. When any check reports a runtime
error, the system SHALL print `Check <name> failed for <uri>: <error>` to
standard error and finish all repositories. It SHALL then exit with status 1
and "one or more checks failed during execution". Check scores and policy
thresholds SHALL NOT affect the exit status.

#### Scenario: Check runtime error

- **GIVEN** a check returns a runtime error for one repository
- **WHEN** the scan of all repositories completes
- **THEN** the system exits with status 1 and
  "one or more checks failed during execution"

#### Scenario: Skipped repository

- **WHEN** a user runs `scorecard --repo=::bad::`
- **THEN** the system prints a `Skipping` message and exits with status 0

#### Scenario: Low score

- **WHEN** every check succeeds but scores 0
- **THEN** the system exits with status 0

### Requirement: Logging Verbosity

The system SHALL set the log level from `--verbosity`, defaulting to
`info`. Accepted values SHALL be `trace`, `debug`, `info`, `warn`, `error`,
`panic`, and `fatal`, case-insensitively; any other value SHALL fall back to
`info` without an error.

#### Scenario: Debug logging

- **WHEN** a user runs with `--verbosity=debug`
- **THEN** debug-level log messages and details are emitted

#### Scenario: Unknown level

- **WHEN** a user runs with `--verbosity=loud`
- **THEN** the system logs at `info` level and does not fail

### Requirement: Feature Flag Environment Variables

The system SHALL treat each feature-flag variable as enabled when it is
present in the environment, regardless of its value. `ENABLE_SARIF` enables
SARIF output and the `--policy` flag. `SCORECARD_V6` enables `raw` output.
`SCORECARD_EXPERIMENTAL` enables experimental checks and Azure DevOps
repositories. Scan flags SHALL have no
environment-variable equivalents.

#### Scenario: Presence enables the flag

- **GIVEN** `ENABLE_SARIF=false` is set
- **WHEN** a user runs with `--format=sarif`
- **THEN** SARIF output is accepted

### Requirement: GitHub Integration Check Override

When `CI` is `true`, `SCORECARD_INTERNAL_GITHUB_INTEGRATION` is `1`, and
`GITHUB_EVENT_NAME` is `dynamic`, the system SHALL replace the `--checks`
list with the comma-separated, whitespace-trimmed value of
`SCORECARD_INTERNAL_GITHUB_CHECKS`.

#### Scenario: Integration overrides checks

- **GIVEN** the three integration variables are set as described and
  `SCORECARD_INTERNAL_GITHUB_CHECKS=License, SAST`
- **WHEN** the command runs with `--checks=Fuzzing`
- **THEN** the system runs `License` and `SAST`

### Requirement: Version Command

When the user runs `scorecard version`, the system SHALL print the tool
name, description, git version, git commit, tree state, build date, Go
version, compiler, and platform. With `--json`, the system SHALL print the
same data as JSON.

#### Scenario: Text version

- **WHEN** a user runs `scorecard version`
- **THEN** the output includes `GitVersion:`, `GitCommit:`, and `Platform:`

#### Scenario: JSON version

- **WHEN** a user runs `scorecard version --json`
- **THEN** the output is a JSON object with the version fields
