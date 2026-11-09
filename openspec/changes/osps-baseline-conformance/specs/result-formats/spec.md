## MODIFIED Requirements

### Requirement: Supported Format Names

The system SHALL accept exactly the `--format` values `default`, `json`,
`sarif`, `probe`, `raw`, `intoto`, and `gemara`, SHALL use `default` when
`--format` is not given, and MUST reject any other value with an
"unsupported format" validation error. `gemara` SHALL be accepted only in
conformance mode. In conformance mode, `json` SHALL select the conformance
JSON format, and the other values SHALL be rejected as defined in
`conformance-evaluation`.

#### Scenario: Format omitted

- **WHEN** a user runs a scan without `--format`
- **THEN** the results are rendered in the `default` text format

#### Scenario: Unknown format

- **WHEN** a user passes `--format yaml`
- **THEN** option validation fails with an "unsupported format" error
- **AND** no scan is performed

#### Scenario: Gemara outside conformance mode

- **WHEN** a user passes `--format gemara` without `--catalog`
- **THEN** option validation fails with "`--format=gemara` requires
  `--catalog`"

## ADDED Requirements

### Requirement: Conformance JSON Format

In conformance mode, the `json` format SHALL write one JSON object with
`experimental` set to `true`, `date` (RFC 3339), `repo` (`name`,
`commit`), `scorecard` (`version`, `commit`), `catalog` (`id`,
`version`), `profile`, `results`, and `notEvaluated`. `notEvaluated`
MUST list the requirement IDs in the profile that have no mapping.

#### Scenario: Conformance JSON output

- **WHEN** a user runs with `--catalog=osps-baseline
  --profile=maturity-1 --format json`
- **THEN** the output has `experimental`, `catalog`, `profile`,
  `results`, and `notEvaluated` keys

### Requirement: Conformance JSON Results

Each entry in the conformance JSON `results` MUST contain `requirement`,
`result`, `explanation`, and `evidence`. `evidence` SHALL list the probe
findings used, in the `probe` format's finding shape.

#### Scenario: Result entry

- **WHEN** a conformance run writes `json`
- **THEN** each result's `result` is one of `PASS`, `FAIL`, `UNKNOWN`,
  `NOT_APPLICABLE`, or `NEEDS_REVIEW`
- **AND** each `evidence` entry has `probe`, `message`, and `outcome`

### Requirement: Gemara Evaluation Log Format

The `gemara` format SHALL write one Gemara Layer 5 Evaluation Log as
JSON that validates against the `#EvaluationLog` definition of the
Gemara schema release the system targets, starting with v1.6.0.

#### Scenario: Schema validation

- **GIVEN** a conformance run written with `--format gemara`
- **WHEN** the output is checked with
  `cue vet -c -d '#EvaluationLog'` against Gemara v1.6.0
- **THEN** validation succeeds

### Requirement: Gemara Evaluation Log Contents

The Gemara Evaluation Log SHALL record:

- `metadata.author` as Scorecard, with type `Software` and the Scorecard
  version;
- `target` as the repository, with the commit assessed;
- one `metadata.mapping-references` entry whose `version` is the catalog
  release version;
- one Control Evaluation per evaluated control, with one Assessment Log
  per evaluated requirement, whose `applicability` is the selected
  profile;
- each finding used as an `evidence` entry.

#### Scenario: Catalog version recorded

- **WHEN** a user runs with `--catalog=osps-baseline@v2026.08.28
  --profile=maturity-1 --format gemara`
- **THEN** `metadata.mapping-references` has an entry with `version`
  `v2026.08.28`

### Requirement: Gemara Result Mapping

Results SHALL map to Gemara values as `PASS` to `Passed`, `FAIL` to
`Failed`, `UNKNOWN` to `Unknown`, `NOT_APPLICABLE` to `Not Applicable`,
and `NEEDS_REVIEW` to `Needs Review`. Requirements that are not evaluated
SHALL be left out of the log.

#### Scenario: Unknown result

- **GIVEN** OSPS-AC-03.01 is `UNKNOWN`
- **WHEN** the run is written with `--format gemara`
- **THEN** its Assessment Log `result` is `Unknown`
