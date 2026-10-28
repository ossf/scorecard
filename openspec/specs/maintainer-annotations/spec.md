# maintainer-annotations Specification

## Purpose

Maintainers can annotate checks in a Scorecard configuration file in their
repository to explain why a check scores low, for example because flagged
files are only test data. This spec covers how that file is discovered and
parsed, which annotations apply to a check result, and when annotations are
not loaded. How annotations are rendered is defined in `result-formats`, and
the `--show-annotations` flag in `cli`.

## Requirements

### Requirement: Annotation File Search Order

When running checks, the system SHALL look for an annotation file in the
scanned repository on every run, whether or not annotations are displayed,
trying these paths in order: `scorecard.yml`, `scorecard.yaml`,
`.scorecard.yml`, `.scorecard.yaml`, `.github/scorecard.yml`, and
`.github/scorecard.yaml`. When a file is found, the system SHALL log
`using maintainer annotations: <path>` at Info level.

#### Scenario: Multiple candidate files exist

- **GIVEN** a repository containing both `.scorecard.yml` and
  `.github/scorecard.yml`
- **WHEN** Scorecard runs its checks
- **THEN** `.scorecard.yml` is used as the annotation file

#### Scenario: No annotation file

- **GIVEN** a repository containing none of the six paths
- **WHEN** Scorecard runs its checks
- **THEN** the result has no annotations configured

### Requirement: First-Readable Selection

The system SHALL use the first path in the search order that the repository
client can open, regardless of whether its content is valid, and SHALL NOT
fall through to a later path when that file fails to parse.

#### Scenario: First readable file is invalid

- **GIVEN** a repository whose `scorecard.yml` is not valid YAML
- **AND** whose `.github/scorecard.yml` is valid
- **WHEN** Scorecard runs its checks
- **THEN** the parse failure of `scorecard.yml` is logged
- **AND** `.github/scorecard.yml` is not read

### Requirement: Non-Fatal Parse Failure

The system SHALL treat an annotation file that cannot be read, cannot be
parsed as YAML, or fails validation as having no annotations. It SHALL log
`couldn't parse maintainer annotations: <error>` at Info level and complete
check execution with unchanged check results. Validation SHALL fail for the
whole file when any annotation names a check outside the 20 registered check
names (matched case-insensitively). It SHALL also fail when any reason is not
`test-data`, `remediated`, `not-applicable`, `not-supported`, or
`not-detected`. One invalid entry discards every other annotation in the file.

#### Scenario: Invalid YAML

- **GIVEN** the selected annotation file cannot be parsed as YAML
- **WHEN** Scorecard runs its checks
- **THEN** every check result is produced as usual
- **AND** no annotations are applied

#### Scenario: Unknown check name invalidates the file

- **GIVEN** an annotation file with one valid annotation for `Fuzzing`
- **AND** a second annotation naming the check `Not-A-Check`
- **WHEN** the file is parsed
- **THEN** parsing fails with
  `configuration file is not valid: check is not valid: Not-A-Check`
- **AND** the `Fuzzing` annotation is not applied

#### Scenario: Unknown reason invalidates the file

- **GIVEN** an annotation file whose only annotation gives reason `wontfix`
- **WHEN** the file is parsed
- **THEN** parsing fails with
  `configuration file is not valid: reason is not valid: wontfix`
- **AND** no annotations are applied

### Requirement: Annotation Applicability

The system SHALL apply to a check result the human-readable explanation of
every reason of every annotation that names the check, matched
case-insensitively, in file order. The system SHALL apply no annotations to a
check whose score is 10. The explanations SHALL be:

- `test-data`: "The files or code snippets are only used for test or example
  purposes."
- `remediated`: "The dangerous files or code snippets are necessary but
  remediations were already applied."
- `not-applicable`: "The check or probe is not applicable in this case."
- `not-supported`: "The check or probe is fulfilled but in a way that is not
  supported by Scorecard."
- `not-detected`: "The check or probe is fulfilled but in a way that is
  supported by Scorecard but it was not detected."

#### Scenario: Annotation on a non-perfect check

- **GIVEN** an annotation file annotating `binary-artifacts` with reason
  `test-data`
- **WHEN** Binary-Artifacts scores 9
- **THEN** the annotations for Binary-Artifacts are "The files or code
  snippets are only used for test or example purposes."

#### Scenario: Annotation on a perfect check

- **GIVEN** an annotation file annotating Binary-Artifacts
- **WHEN** Binary-Artifacts scores 10
- **THEN** no annotations apply to Binary-Artifacts

#### Scenario: Check named in two annotations

- **GIVEN** one annotation for SAST with reason `not-supported`
- **AND** a later annotation for SAST and Fuzzing with reason `not-detected`
- **WHEN** SAST scores 0
- **THEN** both explanations apply to SAST, `not-supported` first

### Requirement: Probe-Mode Exclusion

The system SHALL NOT look for, load, or apply an annotation file when probes
are run directly instead of checks, or when the repository has no commits.

#### Scenario: Direct probes are requested

- **GIVEN** a repository with a valid `.scorecard.yml`
- **WHEN** a user runs Scorecard with `--probes=fuzzed`
- **THEN** no annotation file is read
- **AND** the result has no annotations configured

#### Scenario: Empty repository

- **WHEN** Scorecard scans a repository with no commits
- **THEN** no annotation file is read
