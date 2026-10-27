# Proposal: Map Azure DevOps branch protection

## Why

Current behavior: the Azure client returns only a name for the default
branch. A named branch receives a protection flag if any policy exists,
even when the policy is optional, disabled, or unrelated to PRs. Detailed
protection controls remain empty.

ADO-04 proposes a shared interpretation for default and named branches.
It adds required PR policy data and bounded permission evidence without
changing the shared scoring rules or claiming complete permission coverage.

## What Changes

- Read all applicable policy pages for the repository and canonical branch.
- Translate active required PR policies into conservative protection data.
- Account for self-votes, conditional paths, and build expiration.
- Reuse the default-branch snapshot for explicit access to that branch.
- Evaluate effective caller permissions only as evidence of a permitted
  exception, never as proof of a global prohibition.
- Keep permission controls unknown when their scope or availability
  cannot establish a value, and report authorization limitations.

The intended behavior does not treat a required-reviewer list as CODEOWNERS
enforcement or a bypass grant as administrator identity evidence.
Unknown deletion or force-push protection still receives no basic scoring
credit. A successful policy read does not guarantee a higher score.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `openspec/specs/repo-clients`: adds Azure policy translation and bounded
  permission-evidence requirements.

## Impact

- `clients/azuredevopsrepo`: branch reads, policy interpretation, SDK
  security wiring, and focused regression tests.
- `docs/checks/internal/checks.yaml` and generated `docs/checks.md`: Azure
  integration limits.
- No new dependencies, public schema changes, scoring-weight changes, or
  changes to other providers.
- No Azure configuration writes or organization-wide identity audit.
