# aggregate-scoring Specification

## Purpose

Aggregate scoring combines the individual check scores of a Scorecard run into
one overall score from 0 to 10. Each check is weighted by its documented risk
level, inconclusive checks are left out, and the result is presented with one
decimal place in text and JSON output.

## Requirements

### Requirement: Risk-Weighted Average

The system SHALL compute the aggregate score as the sum of each conclusive
check's score multiplied by its risk weight, divided by the sum of those risk
weights. The risk weights SHALL be 10 for Critical, 7.5 for High, 5 for Medium,
and 2.5 for Low.

#### Scenario: Mixed risk levels

- **GIVEN** results Dangerous-Workflow (Critical) 10, Code-Review (High) 4, and
  License (Low) 10
- **WHEN** the aggregate score is computed
- **THEN** the aggregate score is (100 + 30 + 25) / 20 = 7.75

#### Scenario: Equal risk levels

- **GIVEN** results Fuzzing (Medium) 0 and SAST (Medium) 10
- **WHEN** the aggregate score is computed
- **THEN** the aggregate score is 5

### Requirement: Exclusion Of Inconclusive Checks

The system SHALL leave out of both the numerator and the denominator every
check whose score is below 0, which covers inconclusive results and runtime
error results.

#### Scenario: One inconclusive check

- **GIVEN** results Binary-Artifacts (High) 10 and Signed-Releases (High) -1
- **WHEN** the aggregate score is computed
- **THEN** the aggregate score is 10

#### Scenario: Runtime error check

- **GIVEN** results License (Low) 0 and Vulnerabilities (High) with a runtime
  error
- **WHEN** the aggregate score is computed
- **THEN** the aggregate score is 0

### Requirement: All Checks Inconclusive

The system SHALL report an aggregate score of -1 when no check contributes a
conclusive score, including when there are no check results.

#### Scenario: Every check inconclusive

- **GIVEN** every check result has score -1
- **WHEN** the aggregate score is computed
- **THEN** the aggregate score is -1

#### Scenario: Empty repository

- **GIVEN** the scanned repository has no commits, so no checks ran
- **WHEN** the aggregate score is computed
- **THEN** the aggregate score is -1

### Requirement: Dependence On Selected Checks

The system SHALL compute the aggregate score only from the check results
present in the run, so the aggregate score changes with the set of checks
selected.

#### Scenario: Subset of checks

- **GIVEN** a run limited by `--checks=License` where License scores 9
- **WHEN** the aggregate score is computed
- **THEN** the aggregate score is 9

### Requirement: Unknown Check Or Risk

The system SHALL fail the aggregate score computation with an internal error,
and fail output rendering, when a check result has no documentation entry or
its documented risk is not one of Critical, High, Medium, or Low.

#### Scenario: Undocumented check

- **GIVEN** a result for a check named "Unknown-Check" that has no
  documentation
- **WHEN** the results are rendered in the default or JSON format
- **THEN** rendering fails with an error naming the check

### Requirement: Full Precision Value

The system SHALL keep the aggregate score as an unrounded decimal value and
round it only when presenting it.

#### Scenario: Fractional aggregate

- **GIVEN** results whose weighted average is 6.666...
- **WHEN** the aggregate score is read programmatically
- **THEN** the value is 6.666... rather than a rounded number

### Requirement: Text Presentation

The system SHALL print the aggregate score in the default output format as
"Aggregate score: X / 10", with X rounded to one decimal place, or as
"Aggregate score: ?" when the aggregate score is -1. It SHALL follow this with
a table of check scores, each shown as "N / 10" or as "?" when the check score
is -1.

#### Scenario: Conclusive aggregate

- **GIVEN** an aggregate score of 6.666...
- **WHEN** results are rendered in the default format
- **THEN** the output begins with "Aggregate score: 6.7 / 10"

#### Scenario: Inconclusive aggregate

- **GIVEN** an aggregate score of -1
- **WHEN** results are rendered in the default format
- **THEN** the output begins with "Aggregate score: ?"

#### Scenario: Inconclusive check row

- **WHEN** a check with score -1 is rendered in the default format
- **THEN** its score cell shows "?"

### Requirement: JSON Presentation

The system SHALL emit the aggregate score in JSON output as the top-level
`score` number with exactly one decimal place, including `-1.0` when the
aggregate score is inconclusive, and SHALL emit each check's score as an
integer.

#### Scenario: Whole-number aggregate

- **GIVEN** an aggregate score of 8
- **WHEN** results are rendered as JSON
- **THEN** the top-level `score` is `8.0`

#### Scenario: Inconclusive aggregate in JSON

- **GIVEN** an aggregate score of -1
- **WHEN** results are rendered as JSON
- **THEN** the top-level `score` is `-1.0`
