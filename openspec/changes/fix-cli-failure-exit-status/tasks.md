# Tasks: Fail CLI invocations when scans or result output fail

## 1. CLI behavior and regressions

- [x] 1.1 Record repository and result-output failures while continuing the
  remaining scans.
- [x] 1.2 Restore the formatting-error diagnostic without reporting a
  completed scan as skipped.
- [x] 1.3 Add a compact batch-processing table and a real formatter-failure
  regression.

## 2. Validation

- [x] 2.1 Run focused and adjacent tests, command race checks, and lint.
- [x] 2.2 Confirm coverage of the changed command lines.
- [x] 2.3 Validate the change and record baseline-spec findings with OpenSpec.
- [x] 2.4 Confirm executable failure and success exit codes with offline
  smoke cases.

The change passes strict validation with OpenSpec 1.14.3. All 15 baseline
specs pass normal validation. Strict baseline validation fails on existing
requirement-length warnings in 14 specs. The baseline files are unchanged.
