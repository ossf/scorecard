# Proposal: Verify Azure CI build evidence

## Why

The Azure client converts policy evaluations into check runs without a tool
identity or evidence URL. CI-Tests therefore rejects approved build policies.
Adding an Azure Pipelines label to every approval can also turn reviewer or
work-item policies into false CI evidence.

The client currently uses the completed merge commit as the PR source head.
A PR validation build can test a separate merge revision.
These revisions need separate handling inside the Azure adapter.

## What changes

- Use the PR source head for `clients.PullRequest.HeadSHA`.
- Reuse hydrated PR details to correlate actual Azure Pipelines builds.
- Accept successful pre-merge source builds and verified PR merge builds.
- Return the `azure-pipelines` identity and the matched build's evidence URL.
- Follow build pagination and preserve explicit failures for unreadable or
  malformed evidence.
- Remove the policy-evaluation conversion and its unused client wiring.
- Document Azure detection and its code/build read-access requirements.

Successful CI builds retain the existing evidence standard.
Published test results and mandatory build-policy enforcement are not required.
Branch-Protection retains ownership of policy enforcement.

## Capabilities

### New capabilities

None.

### Modified capabilities

- `openspec/specs/repo-clients`: Azure source-head identity, verified build
  evidence, and collection errors.

Raw collection and scoring continue to follow their existing capabilities.

## Impact

- Azure client wiring, commit details, build reads, and focused client regressions.
- The CI-Tests entry in `docs/checks/internal/checks.yaml` and generated
  `docs/checks.md`.
- No new public fields, output schema, scoring formula, dependency, or CLI flags.
- No automatic commit selection, extension changes, or remote configuration writes.
- No general Azure SAST detection or cross-repository source-build discovery.

Changing Azure's `HeadSHA` also corrects the revision requested by existing
status and SAST consumers. Reviewer and merger identity handling stays unchanged.
