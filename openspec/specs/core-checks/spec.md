# Core Checks System

## Purpose

The checks system provides high-level security assessments of open source repositories. Each check produces a score from 0-10 representing how well a project follows a particular security practice.

## Requirements

### Requirement: Check Registration
The system SHALL allow checks to self-register via `init()` functions using `registerCheck()`.

### Requirement: Parallel Execution
The system SHALL execute all enabled checks concurrently using goroutines.

### Requirement: Three-Tier Architecture
Each check SHALL follow the raw data collection -> probe execution -> evaluation pipeline.

### Requirement: Score Range
All checks SHALL produce a score between 0 (MinResultScore) and 10 (MaxResultScore), or an inconclusive/error result.

### Requirement: Automated Assessment
All checks SHALL be fully automatable and require no interaction from repository maintainers.

### Requirement: Actionable Results
All check results SHALL include actionable remediation guidance.

## Current Checks

Binary-Artifacts, Branch-Protection, CI-Tests, CII-Best-Practices, Code-Review, Contributors, Dangerous-Workflow, Dependency-Update-Tool, Fuzzing, License, Maintained, Packaging, Pinned-Dependencies, SAST, Security-Policy, Signed-Releases, Token-Permissions, Vulnerabilities, Webhooks (experimental), SBOM (experimental).
