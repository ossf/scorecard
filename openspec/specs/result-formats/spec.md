# result-formats Specification

## Purpose

Scorecard renders the result of a scan in one of several output formats
selected with `--format`: a human-readable text table, JSON, SARIF, a flat
probe-findings JSON, raw collected data, and an in-toto attestation
statement. This spec defines which formats exist, how they are gated, what
each one contains, and where the output is written.

## Requirements

### Requirement: Supported Format Names

The system SHALL accept exactly the `--format` values `default`, `json`,
`sarif`, `probe`, `raw`, and `intoto`, SHALL use `default` when `--format`
is not given, and MUST reject any other value with an "unsupported format"
validation error.

#### Scenario: Format omitted

- **WHEN** a user runs a scan without `--format`
- **THEN** the results are rendered in the `default` text format

#### Scenario: Unknown format

- **WHEN** a user passes `--format yaml`
- **THEN** option validation fails with an "unsupported format" error
- **AND** no scan is performed

### Requirement: SARIF Format Gating

The system SHALL reject `--format sarif` with a "SARIF format is not
supported yet" validation error unless the `ENABLE_SARIF` environment
variable is set (to any value). The `--format` help text MUST list `sarif`
as a possible value only when `ENABLE_SARIF` is set.

#### Scenario: SARIF without the feature flag

- **GIVEN** `ENABLE_SARIF` is not set
- **WHEN** a user passes `--format sarif`
- **THEN** option validation fails with "SARIF format is not supported yet"

#### Scenario: SARIF with the feature flag

- **GIVEN** `ENABLE_SARIF=1` is set
- **WHEN** a user passes `--format sarif`
- **THEN** the format passes validation

### Requirement: Raw Format Gating

The system SHALL reject `--format raw` with a "raw option is not supported
yet" validation error unless the `SCORECARD_V6` environment variable is set
(to any value). The `--format` help text MUST NOT list `raw` as a possible
value.

#### Scenario: Raw without the feature flag

- **GIVEN** `SCORECARD_V6` is not set
- **WHEN** a user passes `--format raw`
- **THEN** option validation fails with "raw option is not supported yet"

#### Scenario: Raw with the feature flag

- **GIVEN** `SCORECARD_V6` is set
- **WHEN** a user passes `--format raw`
- **THEN** the format passes validation and raw results are written

### Requirement: Output Destination

The system SHALL write formatted results to standard output by default.
When `--output` (shorthand `-o`) names a file, the system SHALL create or
truncate that file, write the results to it, and print
`Writing results to <file>` on standard error. If the file cannot be
created, formatting MUST fail with an "unable to create output file" error.

#### Scenario: Results written to a file

- **WHEN** a user passes `--output results.json --format json`
- **THEN** the JSON result is written to `results.json`
- **AND** `Writing results to results.json` is printed on standard error
- **AND** nothing is written to standard output for the result

#### Scenario: Unwritable output path

- **WHEN** `--output` points into a directory that does not exist
- **THEN** formatting fails with "unable to create output file"

### Requirement: Formatting Failures Are Reported Per Repository

The system SHALL report a failure to format the results of a repository on
standard error as `Failed to format results for <uri>: <error>` and MUST
continue processing any remaining repositories.

#### Scenario: Formatting error does not stop a multi-repo scan

- **GIVEN** a scan of two repositories with `--repos`
- **WHEN** formatting the first repository's results fails
- **THEN** the error is printed on standard error
- **AND** the second repository is still scanned and formatted

### Requirement: Progress Banners in Default Format

When the format is `default`, the system SHALL print progress banners on
standard error: `Starting (<repo>) [<check>]` per enabled check before the
scan, `Finished (<repo>) [<check>]` per check after it, followed by a
`RESULTS` heading. When probes are selected with `--probes`, the banners
MUST name probes (`Starting (<repo>) probe [<probe>]`,
`Finished (<repo>) probe <probe>`) and the `RESULTS` heading is omitted.
Other formats MUST NOT print these banners.

#### Scenario: Banners for checks

- **WHEN** a user scans a repository with the default format
- **THEN** standard error contains one `Starting` and one `Finished` line
  per enabled check, followed by `RESULTS`

#### Scenario: No banners for machine formats

- **WHEN** a user scans a repository with `--format json`
- **THEN** no `Starting` or `Finished` banners are printed

### Requirement: Check Ordering

The system SHALL sort check results by check name before rendering them in
any check-based format.

#### Scenario: Stable check order

- **WHEN** checks finish in arbitrary order
- **THEN** the rendered checks appear in ascending order of check name

### Requirement: Aggregate Score

The system SHALL compute the aggregate score as the risk-weighted mean of
the check scores, using weights Critical=10, High=7.5, Medium=5, and
Low=2.5 taken from each check's documented risk. Checks with an
inconclusive score (-1) MUST be excluded from both the numerator and the
weights. When no check has a conclusive score, the aggregate score MUST be
-1 (inconclusive). The aggregate depends on which checks were run.

#### Scenario: Weighted mean

- **GIVEN** a High-risk check scoring 10 and a Low-risk check scoring 0
- **WHEN** the aggregate score is computed
- **THEN** the aggregate is (7.5 x 10 + 2.5 x 0) / 10 = 7.5

#### Scenario: All checks inconclusive

- **GIVEN** every check scored -1
- **WHEN** the aggregate score is computed
- **THEN** the aggregate is -1

#### Scenario: Unknown risk level

- **GIVEN** a check whose documented risk is not one of the four levels
- **WHEN** the aggregate score is computed
- **THEN** formatting fails with an internal error naming the check

### Requirement: Default Text Format

The `default` format SHALL print `Aggregate score: <score> / 10` (score with
one decimal place) or `Aggregate score: ?` when inconclusive, then
`Check scores:` and a table with columns `Score`, `Name`, `Reason`, and
`Documentation/Remediation`. Each row's score MUST be shown as `<n> / 10`,
or `?` when the check is inconclusive, and the documentation column MUST
contain the check's documentation URL.

#### Scenario: Text output

- **GIVEN** one check `Check-Name` scoring 5 with reason "half score
  reason"
- **WHEN** the result is rendered in the default format
- **THEN** the output starts with `Aggregate score: 5.0 / 10`
- **AND** the table row shows `5 / 10`, `Check-Name`, the reason, and the
  documentation URL

#### Scenario: Inconclusive check

- **WHEN** a check scored -1
- **THEN** its Score cell shows `?`

### Requirement: Check Documentation URL

The system SHALL build each check's documentation URL as
`https://github.com/ossf/scorecard/blob/<ref>/docs/checks.md#<check-name>`,
where `<check-name>` is the lowercased check name and `<ref>` is the commit
of the Scorecard build, or `main` when that commit is empty or `unknown`.

#### Scenario: Development build

- **GIVEN** a Scorecard build whose commit is `unknown`
- **WHEN** the URL for `Check-Name` is rendered
- **THEN** it is
  `https://github.com/ossf/scorecard/blob/main/docs/checks.md#check-name`

### Requirement: Show Details

When `--show-details` is set, the `default`, `json`, and `intoto` formats
SHALL include each check's detail messages; otherwise details MUST be
omitted. In the text table, details MUST appear in an extra `Details`
column. Debug-level details MUST be included only when `--verbosity debug`
is set; Info and Warn details are included at the other levels.

#### Scenario: Details hidden by default

- **WHEN** a user runs with `--format json` and no `--show-details`
- **THEN** each check's `details` field is null

#### Scenario: Debug details

- **GIVEN** a check produced a Debug detail
- **WHEN** a user runs with `--show-details` and the default verbosity
- **THEN** the Debug detail is not shown
- **AND** it is shown when `--verbosity debug` is also set

### Requirement: Detail Line Rendering

The system SHALL render each detail as `<Type>: <message>` where `<Type>`
is `Info`, `Warn`, or `Debug`. When the detail has a file path, the system
MUST append `: <path>`, then `:<start-line>` when a start line is known,
then `-<end-line>` when the end line is greater than the start line. For
details not backed by a probe finding that carry remediation, the system
MUST append `: <remediation text>`.

#### Scenario: Detail with location

- **GIVEN** a Warn detail "warn message" at `src/file1.cpp` line 5
- **WHEN** details are shown
- **THEN** the detail renders as `Warn: warn message: src/file1.cpp:5`

#### Scenario: Detail with a line range

- **GIVEN** a detail at `a.yml` lines 3 to 7
- **THEN** the location renders as `: a.yml:3-7`

### Requirement: JSON Format

The `json` format SHALL write a single JSON object with the fields `date`
(RFC 3339 timestamp), `repo` (`name`, `commit`), `scorecard` (`version`,
`commit`), `score` (aggregate score with one decimal place, -1.0 when
inconclusive), `checks`, and `metadata`. Each element of `checks` MUST
contain `details`, `score` (integer), `reason`, `name`, and
`documentation` with `url` and `short` (the check's short description).

#### Scenario: JSON result

- **WHEN** a repository is scanned with `--format json`
- **THEN** the output is one JSON object containing `date`, `repo`,
  `scorecard`, `score`, `checks`, and `metadata`
- **AND** an aggregate score of 6 is serialized as `6.0`

### Requirement: JSON Schema File

The project SHALL ship a JSON Schema for the `json` format at
`pkg/scorecard/json.v2.schema` that requires the top-level fields `date`,
`repo`, `scorecard`, `score`, `checks`, and `metadata`, and the per-check
fields `details`, `score`, `reason`, `name`, and `documentation`. The JSON
output MUST validate against this schema.

#### Scenario: Output validates

- **WHEN** a JSON result is produced with details shown
- **THEN** it validates against `json.v2.schema`

### Requirement: Metadata

The system SHALL copy the values passed with `--metadata` (comma-separated,
repeatable) into the `metadata` array of the `json`, `intoto`, and `raw`
formats.

#### Scenario: Metadata passthrough

- **WHEN** a user passes `--metadata team=a,tier=1 --format json`
- **THEN** `metadata` is `["team=a", "tier=1"]`

### Requirement: Maintainer Annotations

The system SHALL read maintainer annotations from the first file found
among `scorecard.yml`, `scorecard.yaml`, `.scorecard.yml`,
`.scorecard.yaml`, `.github/scorecard.yml`, and `.github/scorecard.yaml`
in the scanned repository. When `--show-annotations` is set, the `default`
format MUST add an `Annotation` column and the `json` and `intoto` formats
MUST add an `annotations` array to each annotated check, containing the
human-readable explanation of each annotation reason. Annotations MUST NOT
be shown for a check with the maximum score of 10, and the `annotations`
field MUST be omitted when empty.

#### Scenario: Annotated check in JSON

- **GIVEN** the repository's `.scorecard.yml` annotates `Check-Name` with
  reason `test-data`
- **AND** `Check-Name` scores 5
- **WHEN** a user runs with `--format json --show-annotations`
- **THEN** the check's `annotations` contains "The files or code snippets
  are only used for test or example purposes."

#### Scenario: Annotation on a perfect score

- **GIVEN** an annotated check scores 10
- **WHEN** annotations are shown
- **THEN** no annotation is listed for that check

### Requirement: SARIF Document Structure

The `sarif` format SHALL write a SARIF 2.1.0 document, indented with three
spaces, whose `$schema` is the OASIS SARIF 2.1.0 schema URL. Checks MUST be
grouped into runs by category, and runs MUST be ordered by category name.
The category MUST be `branch-protection` for the Branch-Protection check;
otherwise `local` if the check supports local repositories, else
`online-scm` if it supports GitHub or GitLab. Each run's
`automationDetails.id` MUST be
`supply-chain/<category>/<scorecard-commit>-<date in RFC 822 numeric-zone
form>`, and its tool driver MUST be named `Scorecard` with
`informationUri` `https://github.com/ossf/scorecard` and
`semanticVersion` set to the Scorecard version.

#### Scenario: Local check run

- **GIVEN** a check that supports local repositories
- **WHEN** results are rendered as SARIF
- **THEN** the check appears in a run whose `automationDetails.id` begins
  with `supply-chain/local/`

### Requirement: SARIF Rules

The system SHALL add one rule per check run to its run's tool driver,
whether or not the check produced results. Each rule MUST have `id`
`<check name without dashes>ID`, `name` equal to the check name,
`helpUri` equal to the documentation URL, `shortDescription` equal to the
check name, `fullDescription` and `help.text` equal to the check's short
description, and `help.markdown` containing the check's documented
remediation bullets, severity, and long description. The rule's
`defaultConfiguration.level` MUST be `error`, `precision` MUST be `high`,
and `problem.severity` and `security-severity` MUST map from the check's
risk as Critical to error/9.0, High to error/7.0, Medium to warning/4.0,
and Low to recommendation/1.0. Rule `tags` MUST come from the check's
documented tags.

#### Scenario: Rule for a passing check

- **GIVEN** a check that satisfies the policy
- **WHEN** results are rendered as SARIF
- **THEN** a rule for the check is still present in its run
- **AND** the run has no result for that check

### Requirement: SARIF Results Depend on Policy

The system SHALL emit SARIF results for a check only when the policy
enforces that check and the check's score is below the policy's minimum
score. Checks whose policy mode is disabled, checks with an inconclusive
score, and checks that have maintainer annotations MUST NOT produce
results. If a rendered check has no entry in the policy, SARIF formatting
MUST fail with a "missing policy for check" error.

#### Scenario: Score below threshold

- **GIVEN** a policy enforcing `Token-Permissions` with score 8
- **AND** `Token-Permissions` scored 5
- **WHEN** results are rendered as SARIF
- **THEN** at least one result with `ruleId` `TokenPermissionsID` is
  emitted

#### Scenario: Disabled check

- **GIVEN** a policy that marks a check `disabled`
- **WHEN** results are rendered as SARIF
- **THEN** no result is emitted for that check

#### Scenario: No policy supplied

- **GIVEN** `--format sarif` without `--policy`
- **WHEN** results are formatted
- **THEN** formatting fails with "missing policy for check"

### Requirement: SARIF Result Locations

When `--show-details` is set, the system SHALL create one SARIF result per
Warn detail of a failing check that has a file path and is not a URL. Each
location MUST use the detail's path as `artifactLocation.uri` with
`uriBaseId` `%SRCROOT%`, and its region MUST be derived from the file type:
start and end line for source files, start line for URLs, character
offset for text files, and byte offset plus start line for binaries. Line
numbers MUST be at least 1. When no such location exists, or details are
not shown, the system MUST emit a single result whose location URI is
`no file associated with this alert` at line 1.

#### Scenario: Location from a Warn detail

- **GIVEN** a failing check with a Warn detail at `src/file1.cpp` line 5
- **WHEN** results are rendered as SARIF with `--show-details`
- **THEN** a result points at `src/file1.cpp` with `startLine` 5

#### Scenario: Default location

- **GIVEN** a failing check rendered without `--show-details`
- **THEN** exactly one result is emitted for the check
- **AND** its location URI is `no file associated with this alert`

### Requirement: SARIF Result Messages

Each SARIF result message SHALL begin with `score is <n>:` and a space.
For located results the detail message follows; for the default-location
result the check reason and its non-Info detail lines follow. The message
MUST end with `Click Remediation section below to solve this issue`.
When the detail carries remediation, the location message MUST include
`Remediation tip: <remediation markdown>`, and the result message MUST
instead end with
`Click Remediation section below for further remediation help`.

#### Scenario: Detail with remediation

- **GIVEN** a Warn detail "warn message" with remediation markdown
  "this is the custom markdown help"
- **WHEN** results are rendered as SARIF with `--show-details`
- **THEN** the result message is "score is 5: warn message\nRemediation
  tip: this is the custom markdown help\nClick Remediation section below
  for further remediation help"

### Requirement: Probe Format

The `probe` format SHALL write a single compact JSON object with `date`
(`YYYY-MM-DD`), `repo` (`name`, `commit`), `scorecard` (`version`,
`commit`), and `findings`. Each finding MUST contain `probe`, `message`,
and `outcome`, and MAY contain `location` (`path`, `type`, `lineStart`,
`lineEnd`, `snippet`), `remediation` (`text`, `markdown`, `effort`,
`patch`), and `values`. When probes are selected with `--probes`, the
findings MUST be those probes' findings; otherwise they MUST be the
findings produced by the checks that ran.

#### Scenario: Probe output

- **WHEN** a user runs with `--format probe`
- **THEN** the output contains `findings` with entries such as
  `{"probe": "...", "message": "...", "outcome": "True"}`
- **AND** the `date` has no time component

### Requirement: Raw Format

The `raw` format SHALL write a single JSON object with `date`
(`YYYY-MM-DD`), `repo`, `scorecard`, `metadata`, and `results`, where
`results` holds the raw data collected for the checks or probes that ran
(for example `workflows`, `binaries`, `branchProtections`,
`dependencyPinning`, `securityPolicies`). The project SHALL ship a JSON
Schema for this format at `pkg/scorecard/json.raw.schema`.

#### Scenario: Raw output

- **GIVEN** `SCORECARD_V6` is set
- **WHEN** a user runs with `--format raw --checks Binary-Artifacts`
- **THEN** the output's `results.binaries` lists the binary files found

### Requirement: In-toto Statement Format

The `intoto` format SHALL write a single-line, deterministic in-toto v1
statement with `_type` `https://in-toto.io/Statement/v1`, one subject
whose `name` is the repository name, whose `uri` is
`git+https://<repo>@<commit>`, and whose `digest.gitCommit` is the scanned
commit, and `predicateType` `https://scorecard.dev/result/v0.1`. The
`predicate` MUST be the `json` format result with the `repo` field
removed, honoring `--show-details` and `--show-annotations`. If the
statement fails in-toto validation, formatting MUST fail.

#### Scenario: Statement contents

- **WHEN** a user runs with `--format intoto`
- **THEN** the output has `_type`, `subject`, `predicateType`, and
  `predicate` fields
- **AND** `predicate` contains `scorecard`, `score`, `checks`, and
  `metadata` but no `repo`

#### Scenario: Deterministic output

- **WHEN** the same result is rendered twice
- **THEN** both outputs are byte-for-byte identical
