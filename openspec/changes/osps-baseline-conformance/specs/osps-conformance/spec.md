# OSPS Baseline Conformance

## Purpose

Enable Scorecard to evaluate repositories against the OSPS Baseline specification, producing per-control conformance results (PASS/FAIL/UNKNOWN/NOT_APPLICABLE/ATTESTED) with probe-based evidence, interoperable with the ORBIT WG ecosystem.

## Requirements

### Conformance Engine

#### Requirement: Conformance evaluation
Scorecard SHALL include a conformance engine that evaluates OSPS Baseline controls against Scorecard probe findings and outputs a per-control status.

#### Requirement: Status values
Each control evaluation SHALL produce one of: `PASS`, `FAIL`, `UNKNOWN`, `NOT_APPLICABLE`, or `ATTESTED`.

#### Requirement: UNKNOWN-first honesty
When Scorecard cannot observe a control due to insufficient permissions, missing platform support, or lack of data, the status SHALL be `UNKNOWN` with an explanation. The engine SHALL NOT produce `PASS` or `FAIL` for unobservable controls.

#### Requirement: Applicability detection
The conformance engine SHALL detect applicability preconditions (e.g., "project has made a release") and produce `NOT_APPLICABLE` when preconditions are not met.

#### Requirement: Backward compatibility
The conformance engine SHALL be additive. Existing checks, probes, scores, and output formats SHALL NOT change behavior.

### Mapping

#### Requirement: Versioned mapping file
The mapping between OSPS Baseline controls and Scorecard probes SHALL be maintained as a data-driven, versioned YAML file (e.g., `pkg/osps/mappings/v2026-02-19.yaml`), not hard-coded.

#### Requirement: Mapping contents
Each mapping entry SHALL specify:
- the OSPS control ID
- the maturity level (1, 2, or 3)
- the Scorecard probes that provide evidence
- applicability conditions (if any)
- the evaluation logic (how probe outcomes map to control status)

#### Requirement: Unmapped controls
Controls without mapped probes SHALL appear in output with status `UNKNOWN` and a note indicating no automated evaluation is available.

### Output

#### Requirement: OSPS output format
Scorecard SHALL support `--format=osps` producing a JSON conformance report containing:
- OSPS Baseline version
- target maturity level
- per-control status, evidence, limitations, and remediation
- summary counts
- tool metadata (Scorecard version, timestamp)

#### Requirement: Probe-based evidence
Evidence references in OSPS output SHALL reference probes and their findings. Check-level results SHALL NOT be used as evidence. Checks MAY be listed in a `derived_from` field for human context.

> **Open Question (OQ-4)**: Spencer's position — evidence should be probe-based only, not check-based. This spec adopts that position. Need to confirm this is the consensus view.

#### Requirement: Gemara Layer 4 compatibility
The OSPS output schema SHALL be structurally compatible with Gemara Layer 4 assessment results, enabling consumption by ORBIT ecosystem tools without transformation.

#### Requirement: CI gating
Scorecard SHALL support a `--fail-on=fail` flag (or equivalent) when using OSPS output. `UNKNOWN` statuses SHALL NOT cause failure by default; this SHALL be configurable.

### Metadata and Attestation

#### Requirement: Security Insights ingestion
Scorecard SHALL read Security Insights files (`security-insights.yml` or `.github/security-insights.yml`) to satisfy controls that depend on declared project metadata.

#### Requirement: Attestation for non-automatable controls
The conformance engine SHALL accept attestation evidence from a repo-local metadata file for controls that cannot be automated, producing status `ATTESTED` with evidence links.

> **Open Question (OQ-1)**: The identity and trust model for attestations is unresolved. What does the attestation? OIDC? A personal token? A workflow (which won't have the right tokens)? See proposal.md for options. Spencer flagged this as a blocking design question.

### Ecosystem Interoperability

#### Requirement: Complementarity with the Privateer plugin
Scorecard SHALL NOT duplicate the [Privateer plugin for GitHub repositories](https://github.com/ossf/pvtr-github-repo-scanner). Scorecard provides deep probe-based analysis; the Privateer plugin can consume Scorecard's OSPS output as supplementary evidence.

#### Requirement: No enforcement
Scorecard evaluates and reports conformance. It SHALL NOT enforce policies. The `--fail-on=fail` exit code is a reporting mechanism; the CI system is the enforcer.

> **Open Question (OQ-2)**: Spencer asked whether "enforcement detection" (Phase 3: detecting whether SCA/SAST gating exists) conflicts with Scorecard's stated non-enforcement role. Proposed distinction: Scorecard *detects* enforcement mechanisms, it does not *perform* enforcement. Needs maintainer consensus.

## Scenarios

### Scenario: Full Level 1 conformance report
- GIVEN a public GitHub repository with a Security Insights file
- WHEN `scorecard --repo=github.com/org/repo --format=osps --osps-level=1` is run
- THEN the output contains all Level 1 controls with status PASS, FAIL, UNKNOWN, or NOT_APPLICABLE
- AND each result includes probe-based evidence references

### Scenario: Permission-limited scan produces UNKNOWN
- GIVEN a scan token without admin access
- WHEN evaluating OSPS-AC-01.01 (MFA enforcement)
- THEN the status is `UNKNOWN`
- AND the limitations field explains "requires org admin visibility"

### Scenario: Release applicability triggers NOT_APPLICABLE
- GIVEN a repository that has never made a release
- WHEN evaluating OSPS-DO-01.01 (user guides for released software)
- THEN the status is `NOT_APPLICABLE`
- AND the applicability facts record `has_release=false`

### Scenario: Attestation for non-automatable control
- GIVEN a repo-local metadata file attesting a security assessment was performed
- WHEN evaluating OSPS-SA-03.01
- THEN the status is `ATTESTED`
- AND the evidence includes the attestation source and link

### Scenario: Missing Security Insights file
- GIVEN a repository without a Security Insights file
- WHEN evaluating controls dependent on Security Insights data
- THEN those controls evaluate to `UNKNOWN` with limitation "requires security-insights.yml"
- AND controls not dependent on Security Insights evaluate normally

### Scenario: Unmapped control
- GIVEN an OSPS control with no corresponding Scorecard probes in the mapping file
- WHEN the conformance engine evaluates it
- THEN the status is `UNKNOWN` with note "no automated evaluation available"

### Scenario: CI gating on conformance
- GIVEN `scorecard --format=osps --osps-level=1 --fail-on=fail`
- WHEN any Level 1 control evaluates to `FAIL`
- THEN the process exits with non-zero exit code
- AND `UNKNOWN` controls do not cause failure by default

### Scenario: Existing Scorecard behavior unchanged
- GIVEN any repository
- WHEN `scorecard --repo=github.com/org/repo --format=json` is run (without `--format=osps`)
- THEN the output is identical to Scorecard without OSPS conformance changes
