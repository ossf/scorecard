# OSPS Baseline Integration

## Purpose

Enable Scorecard probes to be mapped to OSPS Baseline controls, allowing Scorecard to report repository compliance against the Open Source Project Security Baseline specification.

## Requirements

### Requirement: Probe Annotation
Probe definition files (`def.yml`) SHALL support an optional `osps_baseline` field that maps the probe to one or more OSPS Baseline control IDs.

### Requirement: Mapping Types
Each probe-to-control mapping SHALL specify a mapping type: `direct`, `partial`, or `informational`.

### Requirement: Maturity Level Tracking
Each mapping SHALL include the OSPS Baseline maturity level (1, 2, or 3) of the associated control.

### Requirement: Baseline Output Format
Scorecard SHALL support an `osps-baseline` output format that groups probe results by OSPS Baseline control and reports per-control compliance status.

### Requirement: Maturity Level Calculation
The baseline output SHALL calculate the highest maturity level achieved (i.e., the highest level where all controls at that level pass).

### Requirement: Backward Compatibility
The `osps_baseline` field SHALL be optional. Probes without this field SHALL continue to function as before with no behavior change.

### Requirement: Generated Mapping Documentation
The build system SHALL generate a mapping document (`docs/osps-baseline-mapping.yaml`) from probe annotations.

## Scenarios

### Scenario: Probe maps to a baseline control
- GIVEN a probe with `osps_baseline` annotation mapping to control OSPS-BR-01
- WHEN Scorecard runs with `--format osps-baseline`
- THEN the output includes OSPS-BR-01 with the probe's finding outcome

### Scenario: Control has no mapped probes
- GIVEN an OSPS Baseline control with no corresponding Scorecard probes
- WHEN Scorecard runs with `--format osps-baseline`
- THEN the control is listed with status `not_assessed`

### Scenario: Multiple probes map to one control
- GIVEN two probes both annotated with control OSPS-AC-01, one `direct` and one `partial`
- WHEN both probes return `OutcomeTrue`
- THEN the control status is `pass`

### Scenario: Partial coverage
- GIVEN two probes mapped to a control, one returning `OutcomeTrue` and one `OutcomeFalse`
- WHEN the failing probe has mapping type `direct`
- THEN the control status is `fail`

### Scenario: Probe without baseline annotation
- GIVEN a probe with no `osps_baseline` field in its `def.yml`
- WHEN Scorecard runs with `--format osps-baseline`
- THEN the probe's results are excluded from the baseline output
- AND the probe continues to function normally in other output formats
