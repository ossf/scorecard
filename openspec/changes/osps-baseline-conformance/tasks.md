# Tasks: OSPS Baseline conformance, first deliverable

## 1. Tracking

- [ ] 1.1 Open a tracking issue for experimental conformance mode with the
  target release, the graduation criteria, and the 12-month
  graduate-or-remove deadline.
- [ ] 1.2 Ask maintainers whether probe names should be a versioned
  interface, since conformance mappings depend on them.

## 2. Evaluation

- [x] 2.1 Add `internal/conformance/` with the result type, requirement
  results, evidence, and coverage.
- [x] 2.2 Add the `osps-baseline` v2026.08.28 release package: the
  requirement IDs and applicability groups for the profile, and the
  mappings for OSPS-AC-03.01 and OSPS-AC-03.02.
- [x] 2.3 Evaluate OSPS-AC-03.01 and OSPS-AC-03.02 from default-branch
  findings, getting the default branch name from the repository client.
- [x] 2.4 Test PASS, FAIL, UNKNOWN, and NEEDS_REVIEW for the two
  requirements, and NOT_APPLICABLE with a test-only requirement limited
  to `maturity-3`.
- [x] 2.5 Test that every probe a mapping names is registered, so a
  renamed probe fails the build.

## 3. Command line and output

- [x] 3.1 Add `--catalog` and `--profile`, their validation, and the
  conflicts with `--checks`, `--probes`, `--policy`, `--repos`, and
  `--org`.
- [x] 3.2 Run the mapped probes in conformance mode without computing
  check scores.
- [x] 3.3 Write conformance JSON.
- [ ] 3.4 Write the Gemara Evaluation Log and validate a generated log
  with `cue vet` against Gemara v1.6.0 in CI.
- [x] 3.5 Mark the options experimental in `--help`, warn on standard
  error, and set `experimental` in the JSON output.
- [x] 3.6 Confirm that native output for the same repository is
  unchanged.

## 4. Documentation and comparison

- [x] 4.1 Document conformance mode, its inputs, and its results for
  users.
- [ ] 4.2 Run conformance mode on ossf/scorecard and compare the two
  results with Privateer's published evaluation; record any
  differences and why.

## 5. Follow-up changes

- [ ] 5.1 Propose OSPS-QA-07.01, OSPS-VM-02.01, and OSPS-BR-06.01, whose
  probe evidence is weaker and needs its own review.
- [ ] 5.2 Propose Scorecard Action support.
- [ ] 5.3 Agree a publishing route to grc.store with the ORBIT WG.
- [ ] 5.4 Propose signing once the attestation identity question is
  resolved.
