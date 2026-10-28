# Proposal: Remove `scorecard-attestor`

## Why

`scorecard-attestor` has had no changes beyond dependency and toolchain
upkeep since 2024, and code search finds no users outside OpenSSF test
repositories. It is the only importer of `github.com/grafeas/kritis`,
pinned to a 2021 pseudo-version, and its image builds on every pull
request.

## What Changes

- **BREAKING**: remove the `scorecard-attestor` command and the
  `attestor/` Go packages. Current behavior: the repository builds and
  tests the attestor and documents it in `openspec/specs/attestor`.
  Target state: neither exists.
- Remove the attestor's Makefile targets, its `docker.yml` image build,
  its Dependabot path, and its threat model entries.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

None.

### Removed Capabilities

- `openspec/specs/attestor`: every requirement is removed, which retires
  the capability.

## Impact

- `attestor/` and `e2e/attestor_policy_test.go` are deleted.
- `go.mod` drops `github.com/grafeas/kritis` and the modules only it
  required.
- No change to the `scorecard` CLI, its checks, probes, scores, or
  output formats.
