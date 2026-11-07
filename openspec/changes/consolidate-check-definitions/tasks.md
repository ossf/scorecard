# Tasks: Define each check's wiring once

## 1. Move the definition code

- [x] 1.1 Move the definition type, `define`, the run body, and the
  Maintained and CI-Tests definitions to `internal/checkdef`.
- [x] 1.2 Keep the exported `checks.Maintained` and `checks.CITests` as
  wrappers; differential tests still pass.

## 2. Characterize probe mode

- [x] 2.1 Add probe-mode raw-collection tests for the 17 remaining checks
  other than Packaging: collectors run, client calls, stored raw data, and
  error text. No registered probe needs Token-Permissions raw data, so its
  test uses a probe registered only in the test.
- [x] 2.2 Cover Packaging on GitHub, a local directory, and an unsupported
  client. GitLab stays unpinned: its client cannot take a test transport.

## 3. Migrate checks in batches

- [x] 3.1 Batch 1: Binary-Artifacts, Branch-Protection, CII-Best-Practices,
  Contributors, Dangerous-Workflow.
- [x] 3.2 Batch 2: Fuzzing, License, Pinned-Dependencies, SAST,
  Security-Policy.
- [x] 3.3 Batch 3: Code-Review, Dependency-Update-Tool, Signed-Releases,
  Token-Permissions, Vulnerabilities.
- [x] 3.4 Batch 4: SBOM and Webhooks (experimental gate), Packaging (two
  collectors).
- [x] 3.5 Each batch adds differential tests against the replaced bodies.

## 4. Point probe mode at the definitions

- [x] 4.1 Look up definitions by check name in `populateRawResults`.
- [x] 4.2 Delete `assignRawData` and its lint suppression.

## 5. Build the registry from the definitions

- [x] 5.1 Snapshot registered names, request types, and the experimental
  set.
- [x] 5.2 Register checks from the definitions; replace per-check `init()`
  registration.

## 6. Consistency and cleanup

- [ ] 6.1 Add the consistency tests in design D5.
- [ ] 6.2 Remove the legacy check bodies from test code.
- [ ] 6.3 Rewrite `checks/write.md` for definitions. It predates the
  raw/evaluation split, request types, and probes.

## 7. Validation

- [ ] 7.1 `SKIP_GINKGO=1 go test -race ./...` and `make lint` for each step.
- [ ] 7.2 `openspec validate consolidate-check-definitions --strict`.
