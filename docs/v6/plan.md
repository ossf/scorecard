# Scorecard v6 Implementation Plan

## Overview

**Phase 1 milestone:** Deliver OSPS Baseline Level 1 conformance evidence using existing infrastructure where possible.

v6 is a clean, backwards-compatible successor to v5. All v6 features land within the v5 module behind feature flags. When all flags graduate to default-on, the module path bumps from `github.com/ossf/scorecard/v5` to `github.com/ossf/scorecard/v6`.

This plan orders work by **dependency and risk**. Prove architectural abstractions with existing code before building new features. Each step declares what it requires and what it enables.

The vision and architectural rationale live in [`proposal.md`](proposal.md); this document is the execution plan.

---

## Phase 1: OSPS Baseline Level 1 conformance evidence

**Goal:** Scorecard produces complete OSPS Baseline Level 1 conformance evidence (PASS/FAIL/UNKNOWN/NOT_APPLICABLE per control) via CLI and GitHub Action, using production-ready JSON output.

**Success criteria:**
1. Complete L1 control coverage (all gap controls closed + existing coverage validated)
2. Framework abstraction proven with existing checks before building OSPS Baseline
3. Production-ready conformance results in JSON output
4. Evidence validated against OSPS Baseline v2026.02.19 controls
5. Existing checks, probes, and scores unchanged (v6 is additive)
6. Existing flagged features (`SCORECARD_V6`, `SCORECARD_EXPERIMENTAL`) promoted or migrated

**Forge support in Phase 1:**
- **GitHub:** Primary target (full L1 coverage)
- **GitLab:** Deferred to a future phase
- **Azure DevOps:** Deferred to a future phase
- **Local directory:** Conformance results for file-based probes only

**Not in Phase 1:**
- Cron infrastructure (deferred to Phase 2 — storage/serving cost evaluation needed)
- Additional output formats (in-toto, Gemara, OSCAL — deferred to Phase 2)
- Level 2 or Level 3 controls (release integrity, enforcement detection, multi-repo)
- Attestation mechanism (non-automatable controls)

---

### Dependency graph

```
Step 0: OpenFeature with existing env vars
  │
  └─► Step 1: Framework abstraction (proven with checks)
        │
        └─► Step 2: JSON output
              │
              └─► Step 3: OSPS Baseline as second framework
                    │
                    └─► Step 4: Human review of L1 coverage analysis
                          │
                          └─► Step 5: Complete L1 coverage (gap probes)
                                │
                                └─► Phase 1 complete: L1 conformance evidence
```

---

## Step 0: OpenFeature infrastructure

**Requires:** Nothing
**Enables:** Feature gating for all subsequent v6 work; promotion of existing flagged features
**Estimated scope:** 1 PR (~200 lines)

### Problem

Scorecard has ad-hoc feature flags checked via `os.LookupEnv()`. A hard-coded delete list in `checks/all_checks.go` manually removes experimental checks. Flags don't have explicit lifecycle management. A `// UPGRADEv4: remove` comment at `options.go:251` shows cleanup doesn't happen organically.

The existing `SCORECARD_V6` env var predates the conformance work described in this plan. Several features under flag have an implicit contract to be promoted in v6.

### Feature promotion

The following features must be addressed as part of the OpenFeature migration:

| Feature | Current Gate | TODO/Comment | Recommendation |
|---------|-------------|-------------|----------------|
| Webhooks check | `SCORECARD_EXPERIMENTAL` | "remove this check when v6 is released" | Promote to always-on in Phase 1 |
| SBOM check | `SCORECARD_EXPERIMENTAL` | "remove this check when v6 is released" | Promote to always-on in Phase 1 |
| Raw output format | `SCORECARD_V6` | none | Promote to always-on in Phase 1 |
| Azure DevOps support | `SCORECARD_EXPERIMENTAL` | none | Keep behind `scorecard.experimental` |
| SARIF output format | `ENABLE_SARIF` | "UPGRADEv4: remove" | Remove gate entirely (always-on) |
| Raw results in CheckRequest | none (commented) | "UPGRADEv6: return raw results instead of scores" | Implement in Phase 1 |

### Solution

Introduce [OpenFeature](https://openfeature.dev/) (`github.com/open-feature/go-sdk`) with `InMemoryProvider` that reads from existing env vars. **Prove the abstraction works with existing infrastructure before using it for v6.**

**What changes:**
- Add `internal/featureflags/` package with simple API:
  ```go
  featureflags.Init()  // reads SCORECARD_V6, SCORECARD_EXPERIMENTAL
  featureflags.Enabled(ctx, "scorecard.v6") bool
  ```
- `checker.Check` gains `FeatureGate string` field
- `checks/all_checks.go` replaces hard-coded delete list with OpenFeature check
- `GetAllWithExperimental()` → `GetAllRegistered()` (used by tooling)
- Webhooks, SBOM, raw output format, and SARIF gates removed (promoted to always-on)

**What doesn't change:**
- Existing env vars (`SCORECARD_V6`, `SCORECARD_EXPERIMENTAL`) still work as aliases
- User workflows unchanged
- CLI flags unchanged

### Flag structure (recommendation — pending approval)

Two flags for Phase 1. The OpenFeature spec does not prescribe naming conventions; this is a project-level decision.

```
scorecard.experimental    — staging area for unproven features (Azure DevOps)
scorecard.v6              — all v6 conformance features (single gate)
```

Per-feature granular flags (e.g., `v6.conformance-evaluation`, `v6.probe.<name>`) are deferred until actual need arises (e.g., cron rollout requiring partial enablement). Two flags is sufficient for Phase 1.

### Testing strategy (recommendation — pending approval)

E2E tests currently do not set feature flag env vars, meaning experimental and v6 features have zero e2e coverage. The OpenFeature migration should include:
- E2E test suite runs twice: once with default flags, once with `SCORECARD_V6=1`
- Test helper that initializes OpenFeature `TestProvider` with specific flag states
- Each new v6 feature gets at least one e2e test with the flag enabled

**Deliverable:** Single PR introducing `internal/featureflags/`, migrating existing flags, promoting graduated features, adding `FeatureGate` to `checker.Check`. Behavioral changes limited to promoting features that were already committed to ship in v6.

---

## Step 1: Framework abstraction

**Requires:** Step 0 (feature flags operational)
**Enables:** Steps 2-5 (output format, OSPS Baseline, gap probes)
**Estimated scope:** 2-3 PRs (~500 lines)

### Problem

Checks are hard-coded compositions of probes into 0-10 scores. There's no abstraction for "a framework that maps probes to verdicts." Before building OSPS Baseline, we need to **prove the abstraction works** with existing code.

### Solution

**Model existing checks as a framework.** This validates the abstraction with known-good code before building anything new.

**Key insight from investigation:**
- Probes produce findings (reusable)
- Check evaluation logic produces 0-10 scores (NOT reusable for conformance)
- The *pattern* is reusable: "take findings, apply evaluation rules, produce verdict"
- Don't shoehorn — checks and conformance have different evaluation semantics

### Baseline levels are one framework, not three (recommendation — pending approval)

OSPS Baseline has three maturity levels (L1, L2, L3). These are **not** separate frameworks. Levels are additive — L1 controls are a subset of L2, which are a subset of L3. The conformance evaluation takes a level parameter:

```go
framework.Evaluate(findings, Level1)  // evaluates only L1 controls
framework.Evaluate(findings, Level2)  // evaluates L1 + L2 controls
```

Phase 1 delivers L1. Phase 2 adds L2 controls to the same framework. Phase 3 adds L3. One framework, expanding control set.

### Architecture

```go
// Framework represents an evaluation surface over probe findings
type Framework interface {
    // Name returns the framework identifier (e.g., "scorecard-checks", "osps-baseline")
    Name() string

    // Evaluate takes probe findings and produces framework-specific results
    Evaluate(findings []finding.Finding) (Result, error)
}

// Result represents a framework's evaluation outcome
type Result interface {
    // Type returns the result type (score, conformance label, etc.)
    Type() string

    // Value returns the type-specific value
    Value() any
}
```

**What this proves:**
1. Existing checks work through the abstraction (no behavior change)
2. The interface supports different evaluation semantics (scores vs. labels)
3. Probe findings are framework-agnostic (same input, different outputs)

**Deliverable:**
- PR 1: Define `Framework` interface and `Result` types
- PR 2: Implement `CheckFramework` wrapping existing checks
- PR 3: Validate all existing checks produce identical scores through abstraction

**Gated behind:** `scorecard.v6`

---

## Step 2: JSON output

**Requires:** Step 1 (framework abstraction exists)
**Enables:** Step 3 (conformance engine testable), Step 5 (gap probes testable)
**Estimated scope:** 1 PR (~300 lines)

### Problem

No output format supports conformance results. Before building OSPS Baseline, we need a way to serialize and validate conformance verdicts.

### Solution

Extend existing JSON output with conformance results, gated behind `scorecard.v6`.

### Schema design (recommendation — pending approval)

**Observation:** `checks` and `conformance` as parallel top-level keys is structurally odd if existing checks will eventually be modeled as a control framework. A more natural design treats all evaluation surfaces uniformly.

**Option A: Extend existing schema (backward-compatible)**

Add a `conformance` field alongside existing `checks`:
```json
{
  "checks": [...],  // existing, unchanged
  "conformance": {  // new, optional
    "framework": "osps-baseline",
    "version": "2026.02.19",
    "controls": [...]
  }
}
```

Pros: No breaking changes. Cons: Structurally inconsistent when checks become a framework.

**Option B: Design a new schema (recommendation)**

All evaluation surfaces live under a unified `evaluations` key:
```json
{
  "date": "2026-04-01",
  "repo": {...},
  "scorecard": {...},
  "findings": [...],           // flat probe findings (the raw evidence)
  "evaluations": {             // framework results
    "scorecard-checks": {      // existing check scores, same data, new location
      "version": "5.0.0",
      "results": [{"name": "Branch-Protection", "score": 8, ...}]
    },
    "osps-baseline": {         // conformance results
      "version": "2026.02.19",
      "level": 1,
      "controls": [{"id": "AC-01.01", "status": "PASS", ...}]
    }
  }
}
```

Old schema (`"checks": [...]`) stays available as the default for backward compatibility. New schema available via `--format=json-v6` or similar. When v6 becomes default, new schema becomes the default JSON output.

Pros: Naturally supports additional frameworks without schema changes; structurally consistent. Cons: More work upfront; migration path needed.

### What this enables

- Test conformance engine outputs as we build OSPS Baseline (Step 3)
- Validate gap probes produce correct evidence (Step 5)
- Existing JSON consumers unaffected (backward-compatible path preserved)

**Deliverable:** Single PR adding conformance output serialization.

**Gated behind:** `scorecard.v6`

---

## Step 3: OSPS Baseline as second framework

**Requires:** Steps 1 (framework abstraction) + 2 (JSON output)
**Enables:** Step 4 (coverage review), Step 5 (gap probes testable against L1 controls)
**Estimated scope:** 3-4 PRs (~800 lines)

### Problem

Framework abstraction is proven with checks. Now build OSPS Baseline conformance using the proven architecture.

### Control catalog: import security-baseline package (recommendation — pending approval)

Import the Go package from [`github.com/ossf/security-baseline`](https://github.com/ossf/security-baseline) rather than maintaining a separate versioned data file:

- **Control catalog** (what controls exist, their names, levels, requirements) comes from the upstream dependency
- **Probe-to-control mappings** (which Scorecard probes satisfy which controls) live in Scorecard
- **OSCAL and Gemara type definitions** come from security-baseline (useful for Phase 2 output formats)

This avoids Scorecard maintaining its own copy of the Baseline spec, which would drift from upstream.

```go
import "github.com/ossf/security-baseline/baseline"

controls := baseline.LoadCatalog("2026.02.19")
for _, control := range controls.Level(1) {
    mapping := scorecard.GetMapping(control.ID)  // Scorecard's probe mapping
    result := evaluate(mapping, findings)
}
```

### Solution

Implement `ConformanceFramework` for OSPS Baseline v2026.02.19.

**Components:**

1. **Control catalog** — imported from `github.com/ossf/security-baseline`
2. **Probe-to-control mappings** — maintained in Scorecard, defining which probes compose into which control outcomes and with what evaluation logic (all-of, any-of)
3. **Applicability engine** — detects preconditions (e.g., "has made a release") and outputs NOT_APPLICABLE
4. **Evaluation logic** — PASS/FAIL/UNKNOWN/NOT_APPLICABLE per control, with probe-level evidence
5. **Map existing probes to L1 controls** where coverage exists today

### Forge-specific behavior

Phase 1 targets GitHub only. GitLab and Azure DevOps conformance support is deferred to a future phase.

**Deliverable:**
- PR 1: Add `github.com/ossf/security-baseline` dependency; control mapping data structure
- PR 2: Applicability engine
- PR 3: Conformance evaluation logic (PASS/FAIL/UNKNOWN/NOT_APPLICABLE)
- PR 4: Map existing probes to L1 controls where coverage exists

**Gated behind:** `scorecard.v6`

---

## Step 4: Human review of L1 coverage analysis

**Requires:** Step 3 (conformance engine operational with existing probe mappings)
**Enables:** Step 5 (validated understanding of what probes need to be written)

### Problem

The existing [`osps-baseline-coverage.md`](osps-baseline-coverage.md) identifies gap controls, but this analysis must be validated by humans before writing new probes. Writing probes based on stale or incorrect gap analysis wastes effort.

### Solution

Before writing gap probes, review the coverage analysis against the current Baseline spec. This review confirms:
- Which controls truly have gaps vs. which are already covered by existing probes
- Whether existing probe-to-control mappings are correct
- Whether the gap count (currently estimated at 9) is accurate

**Deliverable:** Updated `osps-baseline-coverage.md` with human-validated mappings. This is a review task, not a code task.

---

## Step 5: Complete L1 coverage (gap probes)

**Requires:** Step 4 (human-validated coverage analysis)
**Enables:** Phase 1 completion (all L1 controls evaluatable)
**Estimated scope:** Depends on validated gap count (currently estimated 9-12 PRs, ~1,200 lines)

### Problem

OSPS Baseline L1 has gap controls where Scorecard has no coverage. Phase 1 requires complete L1 coverage.

### Solution

Write new probes for validated gap controls. **Metadata ingestion already exists** via `checks/fileparser/` — no new infrastructure needed.

**Estimated gap controls and probe work** (subject to Step 4 review):

1. **Governance/docs presence**
   - GV-02.01: Governance documentation
   - GV-03.01: Contribution guidelines
   - DO-01.01: Technical documentation
   - DO-02.01: User documentation
   - **Reuses:** `fileparser.OnMatchingFileContentDo()` pattern from security policy check

2. **Dependency manifest presence**
   - QA-02.01: Dependency manifest exists
   - **Detects:** `package.json`, `requirements.txt`, `go.mod`, `Cargo.toml`, `pom.xml`, etc.

3. **Security policy deepening**
   - VM-02.01: Security policy has vulnerability disclosure process
   - VM-03.01: Security policy has contact information
   - VM-01.01: Security policy exists and is accessible
   - **Enhancement:** Extend existing `securityPolicy*` probes with deeper content analysis
   - **Already exists:** `securityPolicyPresent`, `securityPolicyContainsVulnerabilityDisclosure`

4. **Secrets detection**
   - BR-07.01: No secrets in repository
   - **Implementation:** Consume platform signals (GitHub secret scanning API where available)

5. **Security Insights metadata**
   - BR-03.01: Project has documented security contacts
   - BR-03.02: Project has documented security assessment
   - QA-04.01: Project has documented quality practices
   - **Reuses:** `fileparser.OnMatchingFileContentDo()` to find `.github/security-insights.yml`
   - **Parses:** YAML content with standard library
   - **No new infrastructure needed** — this is probe work using existing fileparser

**Deliverable:**
Individual PRs per probe or probe group. Each PR includes:
- Probe implementation (`probes/*/impl.go`, `probes/*/def.yml`)
- Control mapping in OSPS Baseline mapping file
- Tests validating probe produces correct findings

**Gated behind:** `scorecard.v6`

---

## Phase 1 Complete

**At this milestone:**
1. Scorecard produces complete OSPS Baseline Level 1 conformance evidence
2. Available via CLI and GitHub Action (production-ready JSON output)
3. Framework abstraction proven with existing checks before OSPS Baseline
4. OpenFeature infrastructure operational with existing env vars
5. All L1 gap controls closed
6. Existing checks, probes, scores unchanged (v6 is additive)
7. Previously flagged features (Webhooks, SBOM, raw format, SARIF) promoted
8. GitHub supported; GitLab and Azure DevOps deferred

**Deferred to Phase 2:**
- Probe catalog extraction (publish probe definitions as consumable artifact)
- Additional output formats (in-toto, Gemara, OSCAL)
- Cron infrastructure (needs storage/serving cost evaluation)
- Level 2 controls (release integrity)
- Attestation mechanism (non-automatable controls)

**Gate for Phase 2:**
Phase 1 must demonstrate value before Phase 2 begins. Success criteria:
- Conformance engine operational with complete L1 coverage
- Framework abstraction proven stable (no major refactors needed)
- OpenFeature flag management working smoothly

---

## Phase 2 (future design)

**Not part of this plan.** When Phase 1 proves value, Phase 2 design will cover:
- Probe catalog extraction (framework abstraction stable; catalog documents reality)
- Additional output formats (in-toto, Gemara, OSCAL)
- Release integrity probes (Level 2 core)
- Attestation mechanism design review
- Evidence bundle packaging
- Cron infrastructure (with storage/serving cost model)

**Each Phase 2 deliverable will be separately scoped and approved.**

---

## Phase 3 (future design)

**Not part of this plan.** When Phase 2 proves value, Phase 3 design will cover:
- Enforcement detection (Level 3)
- Multi-repo project-level conformance
- Attestation GA
- Module path bump to `v6`

---

## Codebase reuse map

v6 should extend existing infrastructure rather than duplicate it. This section
documents what already exists and where each step plugs in.

### Execution pipeline

The main execution flow in `pkg/scorecard/scorecard.go` is:

```
Run() → runScorecard() → populateRawResults() → runEnabledChecks() / runEnabledProbes() → FormatResults()
```

The conformance evaluator fits into this flow without creating a parallel
pipeline:
1. Probes run as normal (unchanged)
2. After probe execution, conformance evaluator consumes `Result.Findings`
3. Conformance results added to `Result` struct
4. `FormatResults()` serializes both checks and conformance

### Reusable components

| Component | Location | How v6 uses it |
|-----------|----------|----------------|
| Probe runner | `probes/zrunner/runner.go` | Runs probes; conformance consumes findings output |
| Probe registration | `internal/probes/probes.go` | New probes use `probes.MustRegister()` |
| Probe grouping | `probes/entries.go` | Add OSPS Baseline probe group |
| Finding types | `finding/finding.go` | Reuse `Outcome` type — `OutcomeNotApplicable` already exists |
| Format dispatcher | `pkg/scorecard/scorecard_result.go:142-182` | Add v6 format case to switch statement |
| Raw data population | `pkg/scorecard/scorecard_result.go:278-410` | Add OSPS Baseline to `assignRawData()` |
| Evaluation pattern | `checks/evaluation/*.go` (19 files) | Template for conformance evaluation functions |
| File parsing | `checks/fileparser/listing.go` | Reuse for Security Insights, governance docs, etc. |
| Score constructors | `checker/check_result.go:107-245` | Reuse pattern for conformance result constructors |
| RepoClient interface | `clients/repo_client.go` | Already injected via `CheckRequest`; no changes |
| Config system | `config/config.go` | Extensible for conformance annotations |
| Policy system | `policy/policy.go` | Extensible for conformance enforcement rules |
| Result struct | `pkg/scorecard/scorecard_result.go` | Add `Conformance` field alongside existing `Checks` |

### Duplication risks

The following plan elements risk duplicating existing infrastructure and should
be validated during implementation:

1. **Framework `Result` interface** — the plan proposes a new `Result`
   interface, but conformance evaluation could follow the simpler existing
   pattern: a function taking `(findings []finding.Finding) → ConformanceResult`,
   matching how `checks/evaluation/*.go` works. Validate whether the interface
   abstraction adds value beyond the function pattern.

2. **Conformance status types** — `finding.Outcome` already defines
   `OutcomeTrue`, `OutcomeFalse`, `OutcomeNotAvailable`, `OutcomeNotApplicable`.
   Per-control conformance status could map directly to these existing types
   rather than defining new PASS/FAIL/UNKNOWN/NOT_APPLICABLE constants.
   Evaluate whether the existing types are sufficient or whether control-level
   status is semantically distinct from probe-level outcome.

3. **Applicability engine** — `OutcomeNotApplicable` already exists as a
   finding outcome. Applicability could be implemented as probes that produce
   `NotApplicable` findings for controls whose preconditions aren't met,
   rather than a separate engine. Evaluate whether the probe-based approach
   or a dedicated engine is cleaner.

4. **Gap probe overlap** — several "gap" controls (VM-01.01 security policy
   exists, GV-03.01 contribution guidelines) may already be partially covered
   by existing probes (`securityPolicyPresent`, etc.). Step 4 (human review)
   will catch this, but implementation should verify before writing new probes.

---

## Resolved decisions

- **OpenFeature introduction:** Submit directly as a PR (no separate RFC/issue)
- **Downstream consumer validation:** Solicit volunteers at a future date
- **Flag naming convention:** OpenFeature spec does not prescribe naming; `scorecard.experimental` and `scorecard.v6` as project convention for Phase 1

## Open questions

**For Phase 2+ (deferred):**
- What provider should cron/server use long-term? (flagd, environment-based, custom?)
- What is the storage/serving cost model for conformance data across 1M+ repos in cron?
