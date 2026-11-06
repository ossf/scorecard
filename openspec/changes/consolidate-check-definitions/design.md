# Design: Define each check's wiring once

## Context

`ossf/scorecard#5425` added an unexported `definition` in `checks`, built
with a typed constructor `define[T]` so that a collector and the
`RawResults` field it fills must agree on type, and one `runDefinition`
body for collect, probe, and evaluate. Maintained (a `CheckRequest`
collector) and CI-Tests (a `RepoClient` collector, through an adapter) use
it. Their exported functions are one-line wrappers.

The other 18 checks follow the same body, with three exceptions:

- SBOM and Webhooks return a runtime error, and collect nothing, unless
  `SCORECARD_EXPERIMENTAL` is set. Probe mode collects their raw data
  without that variable.
- Packaging chooses collectors by client type, and probe mode chooses
  differently (see the proposal).
- Token-Permissions runs three probes that are not registered for direct
  use.

Probe mode calls the same collectors with the same arguments for the other
19 checks, and wraps errors the same way.

## Goals / Non-Goals

**Goals:**

- One definition per check, read by check execution, probe mode, and the
  registry.
- No observable change, shown by tests at each step.
- No new exported API.

**Non-Goals:**

- Changing Packaging's probe-mode behavior.
- Fixing the CI-Tests output order or consolidating the copies of each
  check name. The name constants in `checks`, and `CheckCITests` in
  `checks/evaluation`, are exported and stay; a test pins them to
  `internal/checknames`.
- Control catalog work. The definition stays execution wiring.

## Decisions

### D1: Definitions live in an internal package

The definition type, `define`, the run body, and the definitions move to
`internal/checkdef`. No package
under `checks/raw`, `checks/evaluation`, `probes`, `checker`, or
`internal` imports `checks`, so `checks` and `pkg/scorecard` can both
import it without a cycle. Check names come from `internal/checknames`.

Alternative: keep the definitions in `checks` and export an accessor for
probe mode. Rejected because the accessor would be permanent public API
for an implementation detail.

### D2: A definition carries what the registry needs

Beyond the pilot's fields, a definition records the check's supported
request types and whether it is experimental. `runDefinition` applies the
experimental gate, with the current warning and error text, before
collecting. Probe mode calls only the collector, so it stays ungated.

### D3: Probe mode uses a definition's probe-mode collector

A definition has one collector by default. Packaging also sets a probe-mode
collector that keeps the GitHub and GitLab only behavior and its error.
Probe mode keeps its `seen` set, its order, and its error wrapping;
`populateRawResults` looks the definition up by check name instead of
switching.

### D4: The registry is built from the definitions

`checks` registers one entry per definition, and its function is the
exported wrapper. `GetAll` and `GetAllWithExperimental` keep their
signatures and copy-on-return behavior. The experimental set comes from the
definitions instead of two hard-coded deletions.

### D5: Safety net

- Before probe mode changes, characterization tests pin probe-mode raw
  collection for all 20 checks, extending the pilot's Maintained and
  CI-Tests tests.
- Each migration batch keeps the replaced check bodies verbatim in test code
  and requires identical results, details, raw data, and repository client
  calls, with and without a shared `RawResults`.
- Before the registry changes, a test snapshots registered names, supported
  request types, and the experimental set.
- Consistency tests:
  - Every name in `internal/checknames` has exactly one definition, and
    every definition has a name from that list.
  - Every check named in a registered probe's required raw data has a
    definition.
  - The check-name constants in `checks`, and `CheckCITests` in
    `checks/evaluation`, equal `internal/checknames`.

## Risks / Trade-offs

- **Review load.** Batches of three to five checks and one reviewer per PR.
  As guidance, at most two simplification PRs are open at once; maintainers
  decide.
- **Collision with feature work in `checks/`.** Batches are small and
  independent, so a conflicting batch can wait.
- **The legacy copies in tests age.** They exist to prove one migration and
  are removed after the probe-mode switch is deleted.

## Migration Plan

1. Move the pilot's definition code to `internal/checkdef`; Maintained and
   CI-Tests keep passing their differential tests. This PR also carries
   this change.
2. Probe-mode characterization tests for the remaining 18 checks.
3. Migrate checks in batches of three to five. SBOM, Webhooks, and
   Packaging go last.
4. Point probe mode at the definitions and delete `assignRawData`.
5. Registry snapshot test, then build the registry from the definitions.
6. Consistency tests; remove the legacy copies.

Each step is one PR. Rollback is reverting that PR; earlier steps stand on
their own.

## Open Questions

None.
