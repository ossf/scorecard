# Tasks: Remove `scorecard-attestor`

## 1. Code and build

- [x] 1.1 Delete `attestor/` and `e2e/attestor_policy_test.go`
- [x] 1.2 Remove the attestor Makefile targets, `docker.yml` matrix entry,
  Dependabot path, and `.gitignore` entry
- [x] 1.3 Run `go mod tidy`

## 2. Documentation

- [x] 2.1 Remove the attestor from `ARCHITECTURE.md`, the threat model,
  and `security-insights.yml`
- [x] 2.2 Remove the `attestor` row from `openspec/README.md` and the
  attestor from `openspec/config.yaml`
- [x] 2.3 Retire `openspec/specs/attestor` by archiving this change
