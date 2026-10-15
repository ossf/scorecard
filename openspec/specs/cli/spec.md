# cli Specification

## Purpose

The `scorecard` command-line tool runs Scorecard security checks or probes
against one or more source repositories and writes the results to standard
output or a file. This spec covers the command tree, input selection, the
scan flag list, feature-flag environment variables, validation errors,
progress messages, result ordering, logging verbosity, and exit status. The
behavior activated by check, probe, output, and policy flags is defined in
the capability each flag names.

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

### Requirement: Scan Flags

The root command SHALL register the scan flags below. Each flag's behavior
SHALL be the one defined by the capability named for it:

- `--checks` (comma-separated, default empty): check selection, defined in
  `check-registry`.
- `--probes` (comma-separated, default empty): probe execution, defined in
  `probes`.
- `--format` (default `default`): format names and gating, defined in
  `result-formats`.
- `--output` (shorthand `-o`, default empty): output destination, defined
  in `result-formats`.
- `--show-details` (boolean, default false): detail rendering, defined in
  `result-formats`.
- `--show-annotations` (boolean, default false): annotation rendering,
  defined in `result-formats`; annotation discovery is defined in
  `maintainer-annotations`.
- `--policy` (default empty): policy file and policy-driven selection,
  defined in `policy`.

The system SHALL register `--policy` only when `ENABLE_SARIF` is set;
without it, `--policy` MUST be rejected as an unknown flag.

#### Scenario: Output flag shorthand

- **WHEN** a user runs `scorecard --repo=owner/repo -o results.json`
- **THEN** the system treats `-o results.json` as `--output=results.json`

#### Scenario: Policy flag without SARIF enabled

- **GIVEN** `ENABLE_SARIF` is not set
- **WHEN** a user runs `scorecard --repo=owner/repo --policy=policy.yml`
- **THEN** the command fails with `unknown flag: --policy`

### Requirement: Commit Selection

The system SHALL analyze the commit given by `--commit`, defaulting to
`HEAD`. An empty `--commit` value MUST fail validation with
"commit should be non-empty". Which checks can run at a commit other than
`HEAD` is defined in `check-registry`.

#### Scenario: Scan a specific commit

- **WHEN** a user runs `scorecard --repo=owner/repo --commit=<sha>`
- **THEN** the system analyzes the repository at `<sha>`

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
without the banner. The check messages SHALL name the checks enabled before
the repository-type filter is applied, so they can name a check that the
scan does not run, and SHALL appear in no defined order.

#### Scenario: Default format banners

- **WHEN** a user runs `scorecard --local=. --checks=License`
- **THEN** standard error shows `Starting (.) [License]`,
  `Finished (.) [License]`, and `RESULTS`

#### Scenario: Probe banners

- **WHEN** a user runs `scorecard --local=. --probes=hasLicenseFile`
- **THEN** standard error shows `Starting (.) probe [hasLicenseFile]` and
  `Finished (.) probe hasLicenseFile`
- **AND** no `RESULTS` banner is printed

#### Scenario: Machine-readable format has no banners

- **WHEN** a user runs with `--format=json`
- **THEN** no `Starting` or `Finished` messages are printed

### Requirement: Result Ordering

The CLI SHALL sort each repository's check results by check name, in
ascending byte order, before formatting them. Formatters do not sort; see
`result-formats`.

#### Scenario: Checks complete out of order

- **GIVEN** concurrent checks finish in a non-alphabetic order
- **WHEN** the CLI formats the repository's results
- **THEN** the check results appear in ascending order of check name

#### Scenario: Uppercase sorts before lowercase

- **GIVEN** results for checks named `SAST` and `Fuzzing`
- **WHEN** the CLI formats the results
- **THEN** `Fuzzing` appears before `SAST`

### Requirement: Exit Status

The system MUST exit with status 1 and print
`error during command execution: <error>` when validation fails, the policy
file is invalid, check selection fails, the organization cannot be listed,
or package resolution fails. When a repository cannot be parsed or scanned,
the system SHALL print `Skipping <uri>: <error>` to standard error and
continue without changing the exit status. When formatting or writing a
repository's results fails, the system SHALL print
`Failed to format results for <uri>: <error>` to standard error, continue
with the remaining repositories, and leave the exit status unchanged. When
any check reports a runtime error, the system SHALL print
`Check <name> failed for <uri>: <error>` to standard error and finish all
repositories. It SHALL then exit with status 1 and
"one or more checks failed during execution". Check scores and policy
thresholds SHALL NOT affect the exit status.

#### Scenario: Check runtime error

- **GIVEN** a check returns a runtime error for one repository
- **WHEN** the scan of all repositories completes
- **THEN** the system exits with status 1 and
  "one or more checks failed during execution"

#### Scenario: Skipped repository

- **WHEN** a user runs `scorecard --repo=::bad::`
- **THEN** the system prints a `Skipping` message and exits with status 0

#### Scenario: Unknown probe skips the repository

- **WHEN** a user runs `scorecard --local=. --probes=nope`
- **THEN** the system prints `Skipping .: run: getting probe "nope": ...`
  to standard error and exits with status 0

#### Scenario: Formatting error does not stop a multi-repo scan

- **GIVEN** a scan of two repositories with `--repos`
- **WHEN** formatting the first repository's results fails
- **THEN** `Failed to format results for <uri>: <error>` is printed on
  standard error
- **AND** the second repository is still scanned and formatted
- **AND** the exit status is not changed by the formatting failure

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
`GITHUB_EVENT_NAME` is `dynamic`, the system SHALL replace the selected
checks with the comma-separated, whitespace-trimmed value of
`SCORECARD_INTERNAL_GITHUB_CHECKS`. The override SHALL apply to the `--checks`
list of a CLI scan and to the `checks` list of a `serve` request.

#### Scenario: Integration overrides checks

- **GIVEN** the three integration variables are set as described and
  `SCORECARD_INTERNAL_GITHUB_CHECKS=License, SAST`
- **WHEN** the command runs with `--checks=Fuzzing`
- **THEN** the system runs `License` and `SAST`

#### Scenario: Integration overrides serve requests

- **GIVEN** the three integration variables are set as described for a
  running `scorecard serve`
- **WHEN** a request names checks in its `checks` list
- **THEN** the server runs the checks from
  `SCORECARD_INTERNAL_GITHUB_CHECKS` instead

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
