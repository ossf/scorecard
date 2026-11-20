# Tasks: Verify Azure CI build evidence

## 1. Source identity and builds

- [x] 1.1 Reproduce the source-head mismatch through `Client.ListCommits`.
- [x] 1.2 Use the source head and reuse hydrated PR details without losing
  reviewer/merger identities.
- [x] 1.3 Reproduce missing identity and URL through `Client.ListCheckRunsForRef`.
- [x] 1.4 Replace policy conversion with actual build reads and verified
  source/merge correlation.
- [x] 1.5 Follow build pages and preserve explicit API and malformed-data errors.
- [x] 1.6 Replace obsolete policy tests with minimal client regressions.

## 2. Documentation and validation

- [x] 2.1 Replay actual SDK requests through CI collection, findings, and scoring.
- [x] 2.2 Document Azure support, revision handling, permissions, and limitations.
- [x] 2.3 Regenerate and validate the checks documentation.
- [x] 2.4 Run affected suites, race checks, provider compatibility, lint, and build.
- [x] 2.5 Validate this OpenSpec change strictly.
- [x] 2.6 Confirm that root command, extension, dependencies, and unrelated
  evaluation reports remain unchanged.
