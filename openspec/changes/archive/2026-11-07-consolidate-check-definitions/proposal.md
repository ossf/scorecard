# Proposal: Define each check's wiring once

## Why

Current behavior: each check's wiring is written twice. The check's own
function collects raw data, stores it, runs the check's probes, and
evaluates the findings. Probe mode (`--probes`, `WithProbes`) collects the
same raw data through a separate 20-case switch in `pkg/scorecard`, which
carries a complexity lint suppression. Adding a check means writing both, and
nothing checks that they agree.

`ossf/scorecard#5425` introduced a check definition: one value that pairs a
check's collector with the raw-data field it fills, its probes, and its
evaluation. Maintained and CI-Tests run from definitions today, with
differential tests showing identical results, details, raw data, and
repository client calls. This change finishes that work for the other 18
checks and makes probe mode read the same definitions.

## What Changes

Target state, not current behavior:

- Every check runs from a definition. The exported check functions,
  check-name constants, `GetAll`, and `GetAllWithExperimental` keep their
  signatures and results.
- The definitions move to an internal package, so both check execution and
  probe mode can read them without new exported API.
- Probe mode collects raw data through the definitions. The 20-case switch
  and its lint suppression are deleted.
- The check registry (names, supported request types, experimental gating)
  is built from the definitions instead of per-check registration.
- Tests guarantee that every check has exactly one definition and that every
  check a registered probe needs has a collector.

Behavior does not change. Scores, findings, details, raw data, error text,
output formats, flags, and environment variables stay as they are.

Packaging is the one check whose raw data differs between the two modes.
The check merges the GitHub and GitLab collectors on a local directory and
returns empty data on other hosts. Probe mode supports only GitHub and
GitLab and fails elsewhere. This change keeps both behaviors: Packaging's
definition carries a separate probe-mode collector. Unifying them changes
`--probes` output and belongs in its own change.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `openspec/specs/raw-data-collection`: adds two requirements. A check's raw
  data is the same whether the check runs or its probes are named, and
  Packaging is the recorded exception. Both are current behavior, now stated
  and tested.
- `openspec/specs/probes`: Probe Registration gains the guarantee that every
  check a registered probe needs is a registered check. This is current
  behavior, now stated and tested.

The implementation behind `openspec/specs/check-registry` (Registered
Checks, Experimental Checks, Request Type Support) is rewritten, and its
requirements are unchanged. `openspec/specs/public-go-api` is unchanged.

## Impact

- `checks/`: each check file's body becomes a definition. Per-check
  `init()` registration is replaced by registration from the definitions.
- New package `internal/checkdef` holding the definition type, the run
  body, and the 20 definitions.
- `pkg/scorecard/scorecard_result.go`: probe mode reads the definitions;
  `assignRawData` is deleted.
- Tests: probe-mode characterization for all 20 checks before the switch is
  deleted; old-vs-new differential tests per batch; consistency tests.
- No changes to dependencies, scoring, output schemas, or exported Go API.
- Out of scope: unifying Packaging's two collectors, the CI-Tests ordering
  noted in `ossf/scorecard#5425`, and linking checks to a control catalog.
