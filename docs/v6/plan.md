# Scorecard v6 Implementation Plan

## Overview

v6 is a clean, backwards-compatible successor to v5. v5 does not need parallel
maintenance. All v6 features land within the v5 module behind feature flags.
When all flags graduate to default-on, the module path bumps from
`github.com/ossf/scorecard/v5` to `github.com/ossf/scorecard/v6`.

This plan orders work by **dependency** — each step declares what it requires
and what it enables. No step should begin until its prerequisites are met.
The vision and architectural rationale live in
[`proposal.md`](proposal.md); this document is the execution plan.

---

## Dependency graph

```
Step 0: Feature flag infrastructure (OpenFeature)
  │
  ├─► Step 1: Evidence model + framework abstraction
  │     │
  │     ├─► Step 2: Conformance engine + applicability
  │     │     │
  │     │     ├─► Step 3: Output formats (staggered)
  │     │     │     3a: Enriched JSON (no external dep)
  │     │     │     3b: In-toto evidence predicate
  │     │     │     3c: Gemara (via security-baseline)
  │     │     │     3d: OSCAL (via go-oscal)
  │     │     │
  │     │     └─► Step 4: L1 probe coverage + metadata ingestion
  │     │           (parallel with Step 3)
  │     │
  │     └─► Step 5: Probe catalog extraction
  │
  ╰── Phase 1 complete ──────────────────────────────
  │
  ├─► Step 6: Release integrity probes (L2 coverage)
  │     │
  │     └─► Step 8: Evidence bundle v1
  │
  ├─► Step 7: Attestation mechanism (needs design review)
  │
  ╰── Phase 2 complete ──────────────────────────────
  │
  ├─► Step 9: Enforcement detection (L3 coverage)
  ├─► Step 10: Multi-repo aggregation
  ├─► Step 11: Attestation GA
  │
  ╰── Phase 3 complete → module path bump to v6 ────
```

---

## Step 0: Feature flag infrastructure (OpenFeature)

**Requires:** Nothing
**Enables:** All subsequent v6 work

### Problem

Scorecard has three ad-hoc feature flags, all simple boolean env vars:

| Flag | Env Var | Spread |
|------|---------|--------|
| `EnableSarif` | `ENABLE_SARIF` | 2 files (options only) |
| `EnableScorecardV6` | `SCORECARD_V6` | 2 files (options only) |
| `EnableScorecardExperimental` | `SCORECARD_EXPERIMENTAL` | **22 files** |

All are checked via `os.LookupEnv()` — no granularity, no targeting, no
observability. A `// UPGRADEv4: remove` comment at `options.go:251` shows
flag cleanup doesn't happen organically. A hard-coded delete list in
`checks/all_checks.go` manually removes `CheckWebHooks` and `CheckSBOM`
when `SCORECARD_EXPERIMENTAL` is unset.

### Solution

Introduce [OpenFeature](https://openfeature.dev/)
(`github.com/open-feature/go-sdk`) — a CNCF graduated, vendor-agnostic
feature flagging API. The Go SDK is lightweight (mostly stdlib), Apache-2.0
licensed.

**What OpenFeature replaces (internal feature gating):**
- `EnableScorecardV6` / `SCORECARD_V6`
- `EnableScorecardExperimental` / `SCORECARD_EXPERIMENTAL`
- `EnableSarif` / `ENABLE_SARIF` (eventually)

**What it does NOT replace (user-facing CLI flags):**
- `--format`, `--repo`, `--checks-to-run`, `--probes-to-run`
- These stay as Cobra flags — they're user intent, not feature gates

### Design decisions

**DD-0a: Wrapper location — `internal/featureflags/`**

```go
package featureflags

// Init configures the OpenFeature provider. Call once at startup.
// By default, uses an InMemoryProvider that reads from environment variables.
func Init(opts ...Option) { ... }

// Enabled returns true if the named feature flag is enabled.
func Enabled(ctx context.Context, key string) bool { ... }

// Shutdown cleans up the OpenFeature provider.
func Shutdown() { ... }
```

Lives at `internal/featureflags/` — not public API. Downstream tools control
flags via environment variables or the OpenFeature global API. Can be promoted
to `pkg/featureflags/` later if programmatic control demand arises.

**DD-0b: Check registration — `FeatureGate` field replaces hard-coded
delete list**

All checks register unconditionally. The `checker.Check` struct gains a
`FeatureGate string` field. `getAll()` consults OpenFeature to filter:

```go
type Check struct {
    Fn                    CheckFn
    SupportedRequestTypes []RequestType
    FeatureGate           string // OpenFeature flag key; empty = always enabled
}

func getAll(ctx context.Context) checker.CheckNameToFnMap {
    possibleChecks := checker.CheckNameToFnMap{}
    for k, v := range allChecks {
        if v.FeatureGate != "" && !featureflags.Enabled(ctx, v.FeatureGate) {
            continue
        }
        possibleChecks[k] = v
    }
    return possibleChecks
}
```

`GetAllWithExperimental()` becomes `GetAllRegistered()` (used by
`validate/main.go` and policy tooling).

**DD-0c: Backward compatibility**

An `InMemoryProvider` maps existing env vars to OpenFeature flag keys:
- `SCORECARD_V6=1` → all `v6.*` flags evaluate to `true`
- `SCORECARD_EXPERIMENTAL=1` → all experimental flags evaluate to `true`

Existing user workflows (env vars, CI configurations) continue unchanged.

### Proposed v6 feature keys

```
v6.conformance-evaluation   — Conformance engine (Step 2)
v6.evidence-predicate        — scorecard.dev/evidence/v0.1 output (Step 3b)
v6.oscal-output              — OSCAL Assessment Results (Step 3d)
v6.gemara-output             — Gemara output (Step 3c)
v6.framework-abstraction     — Pluggable framework mappings (Step 1)
v6.applicability-engine      — NOT_APPLICABLE detection (Step 2)
v6.metadata-ingestion        — Metadata sources (Step 4)
v6.probe-catalog             — Probe definitions artifact (Step 5)
```

### Deliverable

Single PR: introduce `internal/featureflags/`, migrate existing flags, add
`FeatureGate` to `checker.Check`, replace `GetAllWithExperimental()` →
`GetAllRegistered()`. Zero behavior change — pure refactor.

---

## Step 1: Evidence model + framework abstraction

**Requires:** Step 0 (feature flags operational)
**Enables:** Steps 2, 3, 4, 5

### Problem

Scorecard has no type definitions for conformance results. There's no
abstraction for "a framework that maps controls to probe compositions" — checks
are the only evaluation surface, and their composition logic is baked into
`probes/entries.go` without a generalizable interface.

### Solution

Define the core types that all subsequent steps build on:

1. **Conformance result types** — PASS / FAIL / UNKNOWN / NOT_APPLICABLE per
   control, with probe-level evidence references
2. **Probe-to-control mapping format** — versioned data file defining which
   probes compose into which control outcomes, with evaluation logic
   (all-of, any-of, weighted)
3. **Framework abstraction interface** — checks and OSPS Baseline are both
   "frameworks" that compose probe findings into verdicts. Define the shared
   interface so both use the same composition engine.
4. **`security-baseline` dependency** — `github.com/ossf/security-baseline`
   as data dependency for control definitions, Gemara types, OSCAL catalog
   models

### Deliverable

New packages: `conformance/types/` (result types, mapping schema),
`conformance/framework/` (framework abstraction interface). Gated behind
`v6.framework-abstraction`.

---

## Step 2: Conformance engine + applicability

**Requires:** Step 1 (types and framework abstraction exist)
**Enables:** Steps 3, 4 (output formats and probes can test against it)

### Problem

No code evaluates probe findings against control mappings. No code detects
when a control is not applicable to a given repository.

### Solution

1. **Conformance evaluator** — takes probe findings + mapping definitions,
   produces per-control conformance results. Core evaluation logic:
   probe composition, status determination, evidence attachment.
2. **Applicability engine** — detects preconditions (e.g., "has made a
   release," "uses GitHub Actions") and outputs NOT_APPLICABLE for controls
   that don't apply.
3. **Map existing probes** to OSPS controls where coverage exists today —
   the first real test of the framework abstraction.

### Deliverable

`conformance/engine/` package. Gated behind `v6.conformance-evaluation`.
Integration test: run conformance evaluation against a known repo and
verify expected PASS/FAIL/UNKNOWN/NOT_APPLICABLE results.

---

## Step 3: Output formats (staggered)

**Requires:** Step 2 (conformance engine produces results to serialize)
**Enables:** Downstream consumer validation (success criteria #2 and #3)

Each format is an independent deliverable. Ship in order of dependency cost:

### Step 3a: Enriched JSON

**Requires:** Step 2
No external dependency. Scorecard-native schema. First format to validate
the evidence model end-to-end.

### Step 3b: In-toto evidence predicate

**Requires:** Step 3a (JSON schema stable)
New `scorecard.dev/evidence/v0.1` predicate type. Framework-agnostic,
probe-level evidence. Existing `scorecard.dev/result/v0.1` preserved
unchanged. Gated behind `v6.evidence-predicate`.

### Step 3c: Gemara output

**Requires:** Step 1 (security-baseline dependency)
Transitive via security-baseline. Gated behind `v6.gemara-output`.

### Step 3d: OSCAL Assessment Results

**Requires:** Step 1 (security-baseline dependency for catalog models)
Via [go-oscal](https://github.com/defenseunicorns/go-oscal). Gated behind
`v6.oscal-output`.

### Deliverable

One PR per format, or group 3a+3b and 3c+3d. Each format validated with at
least one downstream consumer (success criteria #2).

---

## Step 4: L1 probe coverage + metadata ingestion

**Requires:** Step 1 (framework abstraction for mapping probes to controls)
**Can proceed in parallel with:** Steps 2 and 3
**Enables:** Useful Level 1 conformance reports (success criteria #1)

### Problem

Current probe coverage has gaps for OSPS Baseline Level 1 controls.
Metadata-dependent controls (BR-03.01, BR-03.02, QA-04.01) have no
ingestion path.

### Solution

1. **New probes** for Level 1 gaps:
   - Governance/docs presence (GV-02.01, GV-03.01, DO-01.01, DO-02.01)
   - Dependency manifest presence (QA-02.01)
   - Security policy deepening (VM-02.01, VM-03.01, VM-01.01)
   - Secrets detection (BR-07.01) — consume platform signals where possible
2. **Metadata ingestion layer v1** — Security Insights as first supported
   source; architecture supports additional sources (SBOMs, VEX, platform
   APIs). Gated behind `v6.metadata-ingestion`.

### Deliverable

Individual probe PRs (can merge independently). Metadata ingestion as
separate PR. Each probe includes mapping to OSPS control(s) via the
framework abstraction from Step 1.

---

## Step 5: Probe catalog extraction

**Requires:** Step 1 (framework abstraction stable)
**Enables:** Downstream tool integration (AMPEL, Privateer can discover
what Scorecard evaluates)

### Problem

Scorecard's probe definitions (`probes/*/def.yml`) are internal. External
tools can't discover what Scorecard evaluates or compose their own mappings.

### Solution

Extract Scorecard checks into an in-project control framework representation
using the same unified framework abstraction as OSPS Baseline. Package probe
definitions as a consumable artifact. Gated behind `v6.probe-catalog`.

### Deliverable

Probe catalog build step + published artifact. Validated by at least one
external consumer.

---

**Phase 1 complete.** At this point, Scorecard produces useful OSPS Baseline
Level 1 conformance reports via CLI and GitHub Action. The conformance engine,
framework abstraction, evidence model, and at least enriched JSON output are
operational.

**Gate for Phase 2:** Phase 1 must demonstrate value before Phase 2 work
begins. Specifically: conformance engine operational, at least one output
format validated with a downstream consumer, L1 coverage is useful.

---

## Step 6: Release integrity probes

**Requires:** Phase 1 complete (conformance engine proven)
**Enables:** Level 2 coverage

1. Release asset inspection layer (detect compiled assets, SBOMs, licenses)
2. Signed manifest support (BR-06.01)
3. Release notes/changelog detection (BR-04.01)

### Deliverable

Individual probe PRs with OSPS control mappings.

---

## Step 7: Attestation mechanism

**Requires:** Step 2 (conformance engine — needs to know what's attestable)
**Enables:** Non-automatable control coverage
**Needs:** Own design review before implementation

### Problem

Some OSPS Baseline controls cannot be automatically verified (e.g., "has a
security response plan"). Phase 1 defers these as UNKNOWN. Phase 2 needs an
inbound mechanism for maintainers to attest.

### Open design questions (must be resolved before implementation)

- Identity model: OIDC vs. repo-local metadata vs. platform-native signals
- Trust boundary: who can attest, how are attestations verified
- Outbound signing: signing Scorecard's own output (separate from inbound)

### Deliverable

Design review document (separate proposal or issue), then implementation.

---

## Step 8: Evidence bundle v1

**Requires:** Steps 3 (output formats) + 6 (release integrity probes)
**Enables:** Complete Level 2 output package

Conformance results + in-toto statement + SARIF for failures, packaged as a
single evidence bundle. Additional metadata sources for the ingestion layer.

### Deliverable

Bundle format definition + serialization. One PR.

---

**Phase 2 complete.** Scorecard evaluates release-related OSPS controls,
covering the core of Level 2. Useful for downstream due diligence workflows.

**Gate for Phase 3:** Phase 2 must demonstrate L2 value. Attestation design
must be reviewed and approved before Phase 3 attestation GA work begins.

---

## Step 9: Enforcement detection

**Requires:** Phase 2 proven
**Enables:** Level 3 coverage

- SCA policy + enforcement detection (VM-05.*)
- SAST policy + enforcement detection (VM-06.*)

Scorecard detects signals of enforcement (e.g., "SCA tool is configured,"
"SAST results required before merge") but does not itself enforce policies.

### Deliverable

New probes with OSPS control mappings.

---

## Step 10: Multi-repo aggregation

**Requires:** Step 2 (conformance engine works for single repos)
**Enables:** Project-level conformance (QA-04.02)

Multi-repo project-level conformance aggregation — evaluate a project that
spans multiple repositories.

### Deliverable

`--repos` / `--org` flag support for conformance output. Design may need its
own review depending on aggregation semantics.

---

## Step 11: Attestation GA

**Requires:** Step 7 (attestation v1 proven)
**Enables:** Full Level 3 coverage with non-automatable controls

Production-ready attestation integration with resolved identity model.

### Deliverable

Graduation of attestation from experimental to default-on.

---

**Phase 3 complete.** Scorecard covers Level 3 controls including enforcement
detection and project-level aggregation.

**Module path bump:** When all v6 feature flags are default-on and stable,
bump module path from `github.com/ossf/scorecard/v5` to
`github.com/ossf/scorecard/v6`. Remove feature gates. Clean successor,
no parallel maintenance.

---

## Open questions

- What provider should cron/server use long-term? (flagd, environment-based,
  custom?) Decide when cron integration is needed (Phase 2+).
- Should OpenFeature introduction be proposed to maintainers as an RFC/issue
  before implementation, or submitted directly as a PR?
- How should flag keys be namespaced if Scorecard supports multiple framework
  evaluations beyond OSPS Baseline? (e.g., `v6.framework.osps` vs.
  `v6.framework.slsa`)
