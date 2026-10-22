# Proposal: Fail CLI invocations when scans or result output fail

## Why

Current behavior: the CLI can exit 0 after a repository cannot be parsed or
scanned, or its results cannot be written. A CI job can therefore report
success without producing the requested assessment.

The OpenSpec baseline records these defects in
`openspec/specs/cli/spec.md`. This proposal describes the intended
correction in `ossf/scorecard#5319` and references the tracking issue
`ossf/scorecard#5328`. It addresses that issue's skipped-repository defect
and only the exit-status part of its SARIF-without-policy defect.

## What Changes

- Record repository and result-output failures, finish the remaining
  repositories, and exit 1 if any execution failed.
- Preserve the original causes and repository context in the final error.
- Keep `Skipping` for repositories that cannot be scanned. Report output
  failures with `Failed to format results for <uri>`.
- Define immediate diagnostics separately from the final failure summary.
- Preserve exit 0 for successful execution with low or inconclusive scores.

This applies to both `--repos` and repositories selected through `--org`.
CI jobs that previously ignored individual scan failures will now fail.
Existing check runtime errors still cause a nonzero exit.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `openspec/specs/cli`: changes the Exit Status requirement and its
  scenarios for failed scans, result output, and multi-repository runs.

## Impact

- `cmd/root.go`: failure aggregation and stage-specific diagnostics.
- `cmd/root_test.go`: a compact batch table and a real output-failure test.
- No changes to flags, scoring, output schemas, or dependencies.
- SARIF without `--policy` still fails. This change makes that failure
  visible through the command's exit status.
- Output-file overwrites across repositories and logging deduplication
  remain outside this change.
