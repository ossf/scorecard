# result-formats Specification

## Purpose

Scorecard renders the result of a scan in one of several output formats
selected with `--format`: a human-readable text table, JSON, SARIF, a flat
probe-findings JSON, raw collected data, and an in-toto attestation
statement. This spec defines which formats exist, how they are gated, what
each one contains, how scores, details, remediation, and annotations are
rendered, and where the output is written.

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
The file SHALL be created and truncated once per scanned repository, so
with several repositories it holds only the last repository's result. On
standard output, each repository's result SHALL be written as a separate,
independent document with no surrounding array or separator.

#### Scenario: Results written to a file

- **WHEN** a user passes `--output results.json --format json`
- **THEN** the JSON result is written to `results.json`
- **AND** `Writing results to results.json` is printed on standard error
- **AND** nothing is written to standard output for the result

#### Scenario: Unwritable output path

- **WHEN** `--output` points into a directory that does not exist
- **THEN** formatting fails with "unable to create output file"

#### Scenario: Several repositories written to one file

- **WHEN** a user passes `--repos=a/one,a/two --output results.json
  --format json`
- **THEN** `Writing results to results.json` is printed once per repository
- **AND** after the run `results.json` contains only the result for
  `a/two`

#### Scenario: Several repositories on standard output

- **WHEN** a user passes `--repos=a/one,a/two --format json` without
  `--output`
- **THEN** standard output contains two JSON objects, one per repository,
  one after the other

### Requirement: Rendering Order

The formatters SHALL render check results in the order they appear in the
result and SHALL NOT sort them. The `probe` format SHALL list findings in
the order they were collected. Sorting of check results by the CLI is
defined in `cli`, and by the HTTP server in `http-server`.

#### Scenario: Library caller passes unsorted results

- **GIVEN** a result whose checks are ordered `SAST`, `License`
- **WHEN** a Go caller renders it with the `json` format
- **THEN** `checks` lists `SAST` before `License`

### Requirement: Default Text Format

The `default` format SHALL print `Aggregate score: <score> / 10`, with the
aggregate score rounded to one decimal place, or `Aggregate score: ?` when
the aggregate score is -1, then `Check scores:` and a table with columns
`Score`, `Name`, `Reason`, and `Documentation/Remediation`. A `Details`
column SHALL be added before `Documentation/Remediation` when details are
shown, and an `Annotation` column SHALL be added last when annotations are
shown. Each row's score MUST be shown as `<n> / 10`, or `?` when the check
is inconclusive, and the `Documentation/Remediation` column MUST contain the
check's documentation URL as defined in `check-registry`. The aggregate
score computation is defined in `aggregate-scoring`.

#### Scenario: Text output

- **GIVEN** one check `Check-Name` scoring 5 with reason "half score
  reason"
- **WHEN** the result is rendered in the default format
- **THEN** the output starts with `Aggregate score: 5.0 / 10`
- **AND** the table row shows `5 / 10`, `Check-Name`, the reason, and the
  documentation URL

#### Scenario: Fractional aggregate is rounded

- **GIVEN** an aggregate score of 6.666...
- **WHEN** results are rendered in the default format
- **THEN** the output begins with `Aggregate score: 6.7 / 10`

#### Scenario: Inconclusive aggregate

- **GIVEN** an aggregate score of -1
- **WHEN** results are rendered in the default format
- **THEN** the output begins with `Aggregate score: ?`

#### Scenario: Inconclusive check

- **WHEN** a check scored -1
- **THEN** its Score cell shows `?`

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

#### Scenario: Pinned dependency detail with remediation

- **GIVEN** an unpinned GitHub Action with a pinning suggestion
- **WHEN** results are rendered with `--show-details`
- **THEN** the Warn detail ends with `: update your workflow using`
  followed by the StepSecurity `secureworkflow` URL

### Requirement: JSON Format

The `json` format SHALL write a single JSON object with the fields `date`
(RFC 3339 timestamp), `repo` (`name`, `commit`), `scorecard` (`version`,
`commit`), `score` (aggregate score with exactly one decimal place, -1.0
when inconclusive), `checks`, and `metadata`. Each element of `checks` MUST
contain `details`, `score` (integer), `reason`, `name`, and
`documentation` with `url` and `short` (the check's short description).
The `date` SHALL be the time recorded when the run starts, after the
repository client is initialized and before checks run, not the time the
run completes.

#### Scenario: JSON result

- **WHEN** a repository is scanned with `--format json`
- **THEN** the output is one JSON object containing `date`, `repo`,
  `scorecard`, `score`, `checks`, and `metadata`
- **AND** an aggregate score of 6 is serialized as `6.0`

#### Scenario: Inconclusive aggregate in JSON

- **GIVEN** an aggregate score of -1
- **WHEN** results are rendered as JSON
- **THEN** the top-level `score` is `-1.0`

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

When `--show-annotations` is set, the `default` format SHALL add an
`Annotation` column and the `json` and `intoto` formats SHALL add an
`annotations` array to each annotated check, containing the human-readable
explanation of each annotation reason. The `annotations` field MUST be
omitted when empty. Which annotations apply to a check is defined in
`maintainer-annotations`.

#### Scenario: Annotated check in JSON

- **GIVEN** the repository's `.scorecard.yml` annotates `Check-Name` with
  reason `test-data`
- **AND** `Check-Name` scores 5
- **WHEN** a user runs with `--format json --show-annotations`
- **THEN** the check's `annotations` contains "The files or code snippets
  are only used for test or example purposes."

#### Scenario: Annotations not requested

- **GIVEN** an annotated check scoring 5
- **WHEN** a user runs with `--format json` and no `--show-annotations`
- **THEN** the check has no `annotations` field

### Requirement: SARIF Document Structure

The `sarif` format SHALL write a SARIF 2.1.0 document, indented with three
spaces, whose `$schema` is the OASIS SARIF 2.1.0 schema URL. Checks MUST be
grouped into runs by category, and runs MUST be ordered by category name.
The category MUST be `branch-protection` for the Branch-Protection check;
otherwise `local` if the check supports local repositories, else
`online-scm` if it supports GitHub or GitLab. Each run's
`automationDetails.id` MUST be
`supply-chain/<category>/<scorecard-commit>-<date in RFC 822 numeric-zone
form>`, and its tool driver MUST have `informationUri`
`https://github.com/ossf/scorecard` and `semanticVersion` set to the
Scorecard version. The driver name SHALL be `Scorecard`; when the GitHub
integration is enabled (`CI` is `true`,
`SCORECARD_INTERNAL_GITHUB_INTEGRATION` is `1`, and `GITHUB_EVENT_NAME` is
`dynamic`), it SHALL instead be the whitespace-trimmed value of
`SCORECARD_INTERNAL_GITHUB_SARIF_TOOL_NAME` converted to title case.

#### Scenario: Local check run

- **GIVEN** a check that supports local repositories
- **WHEN** results are rendered as SARIF
- **THEN** the check appears in a run whose `automationDetails.id` begins
  with `supply-chain/local/`

#### Scenario: Driver name under the GitHub integration

- **GIVEN** the GitHub integration is enabled and
  `SCORECARD_INTERNAL_GITHUB_SARIF_TOOL_NAME` is `acme scanner`
- **WHEN** results are rendered as SARIF
- **THEN** each run's tool driver is named `Acme Scanner`

### Requirement: SARIF Rules

The system SHALL add one rule per check run to its run's tool driver,
whether or not the check produced results. Each rule MUST have `id`
`<check name without dashes>ID`, `name` equal to the check name,
`helpUri` equal to the documentation URL, `shortDescription` equal to the
check name, `fullDescription` and `help.text` equal to the check's short
description, and `help.markdown` containing, in order, a
`**Remediation (click "Show more" below)**:` section with one Markdown list
item per documented remediation bullet, the severity, and the long
description. The rule's
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

#### Scenario: Rule help lists remediation

- **GIVEN** a check documented with remediation items "a" and "b"
- **WHEN** results are rendered as SARIF
- **THEN** the rule's `help.markdown` begins with the remediation heading
  followed by list items for "a" and "b"

### Requirement: SARIF Results Depend on Policy

The system SHALL emit SARIF results for a check only when the policy
enforces that check and the check's score is below the policy's minimum
score. Checks whose policy mode is disabled, checks with an inconclusive
score, and checks that have maintainer annotations MUST NOT produce
results; annotated checks are omitted whether or not `--show-annotations`
is set. If a rendered check has no entry in the policy, SARIF formatting
MUST fail with a "missing policy for check" error.

#### Scenario: Score below threshold

- **GIVEN** a policy enforcing `Token-Permissions` with score 8
- **AND** `Token-Permissions` scored 5
- **WHEN** results are rendered as SARIF
- **THEN** at least one result with `ruleId` `TokenPermissionsID` is
  emitted

#### Scenario: Score meets the threshold

- **GIVEN** a policy enforcing `Branch-Protection` with score 5
- **AND** `Branch-Protection` scored 5
- **WHEN** results are rendered as SARIF
- **THEN** the run contains the `Branch-Protection` rule but no result for
  it

#### Scenario: Annotated check without show-annotations

- **GIVEN** a policy enforcing a check with score 8
- **AND** the check scored 5 and has a maintainer annotation
- **WHEN** results are rendered as SARIF without `--show-annotations`
- **THEN** no result is emitted for that check

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

The `probe` format SHALL write one compact JSON object per repository with
`date`
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

#### Scenario: Selected probe output

- **WHEN** a user runs `scorecard --repo=github.com/owner/repo
  --probes=fuzzed --format probe`
- **THEN** the output has `date`, `repo`, `scorecard`, and `findings` keys
- **AND** every finding has `probe` set to `fuzzed`

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
