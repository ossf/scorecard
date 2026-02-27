# Probes System

## Purpose

Probes are granular, individual heuristics that assess a specific behavior a project may or may not exhibit. They are the atomic units of analysis within Scorecard, composed into higher-level checks.

## Requirements

### Requirement: Boolean Naming
Probe names SHALL use camelCase and be phrased as boolean statements (e.g., `hasUnverifiedBinaryArtifacts`).

### Requirement: Three-File Structure
Each probe SHALL consist of exactly three files: `def.yml` (documentation), `impl.go` (implementation), `impl_test.go` (tests).

### Requirement: Finding Outcomes
Probes SHALL return one or more `finding.Finding` values with outcomes from: `OutcomeTrue`, `OutcomeFalse`, `OutcomeNotApplicable`, `OutcomeNotAvailable`.

### Requirement: Lifecycle Management
Each probe SHALL declare a lifecycle state in `def.yml`: `Experimental`, `Stable`, or `Deprecated`.

### Requirement: Registration
Probes SHALL register via `probes.MustRegister()` (check-associated) or `probes.MustRegisterIndependent()` (standalone), and be cataloged in `probes/entries.go`.

### Requirement: Remediation
Probes SHALL provide remediation guidance in their `def.yml` definition.

## Scenarios

### Scenario: Probe returns true finding
- GIVEN a repository that exhibits the behavior described by the probe
- WHEN the probe executes
- THEN it returns at least one finding with `OutcomeTrue`

### Scenario: Probe returns false finding
- GIVEN a repository that does not exhibit the behavior
- WHEN the probe executes
- THEN it returns at least one finding with `OutcomeFalse`

### Scenario: Probe lacks data
- GIVEN a repository where the relevant data is unavailable
- WHEN the probe executes
- THEN it returns a finding with `OutcomeNotAvailable`
