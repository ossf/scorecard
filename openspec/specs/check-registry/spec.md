# check-registry Specification

## Purpose

The check registry defines which Scorecard checks exist, how each is
documented, which repositories and request modes each supports, and how the
set of checks to run is selected from requested check names and executed. It
covers experimental-check gating, check-name matching, concurrent execution,
and validation of the check documentation. Flag definitions are in `cli`, and
policy-driven selection is in `policy`.

## Requirements

### Requirement: Registered Checks

The system SHALL register exactly these checks by name: Binary-Artifacts,
Branch-Protection, CI-Tests, CII-Best-Practices, Code-Review, Contributors,
Dangerous-Workflow, Dependency-Update-Tool, Fuzzing, License, Maintained,
Packaging, Pinned-Dependencies, SAST, SBOM, Security-Policy, Signed-Releases,
Token-Permissions, Vulnerabilities, and Webhooks. Each check result SHALL carry
its check's registered name.

#### Scenario: Full check list

- **GIVEN** the `SCORECARD_EXPERIMENTAL` environment variable is set
- **WHEN** the full list of checks is requested
- **THEN** it contains the 20 checks named above

#### Scenario: Result naming

- **WHEN** the Token-Permissions check runs
- **THEN** its result is named "Token-Permissions"

### Requirement: Check Risk Levels

The system SHALL assign each check one of the risk levels Critical, High,
Medium, or Low, as follows:

- Critical: Dangerous-Workflow, Webhooks.
- High: Binary-Artifacts, Branch-Protection, Code-Review,
  Dependency-Update-Tool, Maintained, Signed-Releases, Token-Permissions,
  Vulnerabilities.
- Medium: Fuzzing, Packaging, Pinned-Dependencies, SAST, SBOM,
  Security-Policy.
- Low: CI-Tests, CII-Best-Practices, Contributors, License.

#### Scenario: Risk lookup

- **WHEN** the documentation for Dangerous-Workflow is read
- **THEN** its risk is Critical

#### Scenario: Low-risk check

- **WHEN** the documentation for License is read
- **THEN** its risk is Low

### Requirement: Experimental Checks

The system SHALL treat SBOM and Webhooks as experimental. They SHALL be
excluded from the default set of checks unless the `SCORECARD_EXPERIMENTAL`
environment variable is set (to any value), and SHALL remain selectable by
name without it. When either is run without that variable set, the check
SHALL collect no raw data and SHALL return a runtime error result stating that
`SCORECARD_EXPERIMENTAL` is not set.

#### Scenario: Default run without the variable

- **GIVEN** `SCORECARD_EXPERIMENTAL` is not set
- **WHEN** Scorecard runs with no explicit check selection
- **THEN** SBOM and Webhooks are not run

#### Scenario: Default run with the variable

- **GIVEN** `SCORECARD_EXPERIMENTAL` is set
- **WHEN** Scorecard runs with no explicit check selection
- **THEN** SBOM and Webhooks are run along with the other supported checks

#### Scenario: Explicitly selected without the variable

- **GIVEN** `SCORECARD_EXPERIMENTAL` is not set
- **WHEN** the user runs `--checks=SBOM`
- **THEN** the SBOM result is a runtime error result with score -1
- **AND** a Warn detail says the variable is not set
- **AND** no SBOM raw data is collected

### Requirement: Request Type Support

The system SHALL declare, for each check, which request modes it supports:
file-based (analysis of file contents only, required for local directories)
and commit-based (analysis at a commit other than HEAD). A run SHALL require
file-based support when it scans a local directory, and commit-based support
when the requested commit is not `HEAD` (compared case-insensitively). Checks
SHALL support the modes as follows:

- Both modes: Binary-Artifacts, Dangerous-Workflow, License,
  Pinned-Dependencies, Security-Policy, Token-Permissions, Vulnerabilities.
- File-based only: Dependency-Update-Tool, Fuzzing, Packaging, SAST.
- Commit-based only: CI-Tests, Code-Review.
- Neither: Branch-Protection, CII-Best-Practices, Contributors, Maintained,
  SBOM, Signed-Releases, Webhooks.

When the checks are not chosen explicitly, the system SHALL silently skip
checks that do not support every required mode. When a named check does not
support every required mode, the run MUST fail with
`check is not supported for this request: "<name>" does not support <modes>
data, which this scan requires`, where the modes are rendered as `file-based`
and `commit-based` rather than as their internal numeric values.

#### Scenario: Local directory

- **WHEN** Scorecard runs against a local directory with default checks
- **THEN** only Binary-Artifacts, Dangerous-Workflow, Dependency-Update-Tool,
  Fuzzing, License, Packaging, Pinned-Dependencies, SAST, Security-Policy,
  Token-Permissions, and Vulnerabilities are run

#### Scenario: Non-HEAD commit

- **WHEN** Scorecard runs with `--commit` set to a SHA other than HEAD
- **THEN** checks lacking commit-based support, such as Branch-Protection and
  SAST, are not run

#### Scenario: Explicit selection of an unsupported check

- **WHEN** the user runs `--local=.` with `--checks=Code-Review`
- **THEN** Scorecard fails with
  `check is not supported for this request: "Code-Review" does not support file-based data, which this scan requires`

#### Scenario: Runner sanity check

- **WHEN** a check is executed with a request requiring a mode the check does
  not support
- **THEN** the check result is a runtime error result marked as an unsupported
  check

### Requirement: Repository Type Support

The system SHALL document, for each check, the repository types it supports
among GitHub, GitLab, Azure DevOps, and local, and SHALL silently skip a check
whose documented repository types (compared case-insensitively) exclude the
type of the repository being scanned. A check with no documentation entry
SHALL be treated as supporting every type. The documented support SHALL be:

- GitHub only: Signed-Releases, Webhooks.
- GitHub and GitLab: Maintained, SBOM.
- GitHub and Azure DevOps: Contributors.
- GitHub and local: Dangerous-Workflow, Packaging, Token-Permissions.
- GitHub, Azure DevOps, and local: SAST.
- GitHub, GitLab, and Azure DevOps: Branch-Protection, CI-Tests,
  CII-Best-Practices, Code-Review.
- All four: Binary-Artifacts, Dependency-Update-Tool, Fuzzing, License,
  Pinned-Dependencies, Security-Policy, Vulnerabilities.

#### Scenario: GitLab repository

- **WHEN** Scorecard runs against a GitLab repository with default checks
- **THEN** Signed-Releases, Token-Permissions, and Dangerous-Workflow are not
  run

#### Scenario: Explicit selection of an unsupported repository type

- **WHEN** the user selects `--checks=Signed-Releases` for a GitLab repository
- **THEN** Signed-Releases is skipped without an error

### Requirement: Check Selection

The system SHALL run every default check that supports the run's request modes
and repository type when no check names are requested, and SHALL run only the
named checks when check names are requested (for example, the comma-separated
`--checks` list). A requested name SHALL match a registered check
case-insensitively, except that the request-mode lookup uses the exact name:
when a request mode is required (a local directory or a non-HEAD commit), a
name whose case differs from the registered name MUST fail with
`invalid check name: "<name>" must be written "<registered>" when a request
type is required`. A name that matches no registered check in any casing MUST
fail the run with `invalid check name: "<name>" is not a known check`, and the
message SHALL point the user at `scorecard --help`. Neither failure is reported
as an internal error. Policy-driven selection is defined in `policy`.

#### Scenario: Selecting checks

- **WHEN** the user runs `--checks=License,code-review` against a remote
  repository at HEAD
- **THEN** only License and Code-Review are run

#### Scenario: Unknown check

- **WHEN** the user runs `--checks=Not-A-Check` against a remote repository at
  HEAD
- **THEN** the run fails with an error containing
  `"Not-A-Check" is not a known check`

#### Scenario: Wrong case in local mode

- **WHEN** the user runs `--local=.` with `--checks=license`
- **THEN** the run fails with
  `invalid check name: "license" must be written "License" when a request type is required`

#### Scenario: Wrong case at a non-HEAD commit

- **WHEN** the user runs `--commit=<sha>` with `--checks=license` against a
  remote repository, where `<sha>` is not HEAD
- **THEN** the run fails with
  `invalid check name: "license" must be written "License" when a request type is required`

#### Scenario: Flag help

- **WHEN** the user views the `--checks` flag help
- **THEN** it lists the names of the default checks

### Requirement: Concurrent Check Execution

The system SHALL run every selected check concurrently against the same
repository snapshot and collect one result per selected check before
returning. Result ordering is defined by the callers (`cli` and
`http-server`).

#### Scenario: Results collected

- **WHEN** five checks are selected
- **THEN** the result contains five check results

### Requirement: Check Documentation

The system SHALL maintain, for every check, documentation with a risk level, a
short description, a long description, a list of remediation steps, a
comma-separated list of tags, and a list of supported repository types. Each
check SHALL have a documentation URL of the form
`https://github.com/ossf/scorecard/blob/<commitish>/docs/checks.md#<name>`,
where `<name>` is the lowercased check name and `<commitish>` is the Scorecard
commit, or `main` when that is empty or `unknown`.

#### Scenario: Documentation URL

- **WHEN** the documentation URL for Code-Review is requested with an unknown
  Scorecard commit
- **THEN** the URL is
  `https://github.com/ossf/scorecard/blob/main/docs/checks.md#code-review`

#### Scenario: Short description in JSON output

- **WHEN** results are rendered as JSON
- **THEN** each check includes its documentation URL and short description

#### Scenario: Generated reference

- **WHEN** `make generate-docs` runs
- **THEN** `docs/checks.md` is regenerated with one section per check, sorted
  by name, containing its long description and remediation steps

### Requirement: Check Documentation Validation

The system SHALL provide a documentation validator, run by
`make validate-docs`, that fails when any registered check (including
experimental checks) lacks a documentation entry with a matching name, a
non-empty short description, long description, remediation, and tags, or a
risk of Critical, High, Medium, or Low; and fails when the documentation
contains a check that is not registered.

#### Scenario: Missing remediation

- **WHEN** a registered check's documentation has empty remediation
- **THEN** validation fails naming that check

#### Scenario: Invalid risk

- **WHEN** a check's documented risk is "Severe"
- **THEN** validation fails reporting the invalid risk

#### Scenario: Documented check not registered

- **WHEN** the documentation contains a check that is not registered
- **THEN** validation fails naming that check
