# Scorecard v6 Implementation Plan

## Overview

**Phase 1 milestone:** Deliver OSPS Baseline Level 1 conformance evidence using existing infrastructure where possible.

v6 is a clean, backwards-compatible successor to v5. All v6 features land within the v5 module behind feature flags. When all flags graduate to default-on, the module path bumps from `github.com/ossf/scorecard/v5` to `github.com/ossf/scorecard/v6`.

This plan orders work by **dependency and risk**. Prove architectural abstractions with existing code before building new features. Each step declares what it requires and what it enables.

The vision and architectural rationale live in [`proposal.md`](proposal.md); this document is the execution plan.

---

## Phase 1: OSPS Baseline Level 1 conformance evidence

**Goal:** Scorecard produces complete OSPS Baseline Level 1 conformance evidence (PASS/FAIL/UNKNOWN/NOT_APPLICABLE per control) via CLI and GitHub Action, using extended JSON output.

**Success criteria:**
1. Complete L1 control coverage (all 9 gap controls closed + existing coverage validated)
2. Framework abstraction proven with existing checks before building OSPS Baseline
3. Conformance results in existing JSON output format (no new output formats required)
4. Evidence validated against OSPS Baseline v2026.02.19 controls
5. Existing checks, probes, and scores unchanged (v6 is additive)

**Not in Phase 1:**
- Cron infrastructure (deferred to Phase 2 - storage/serving cost evaluation needed)
- Additional output formats (in-toto, Gemara, OSCAL - deferred to Phase 2)
- Level 2 or Level 3 controls (release integrity, enforcement detection, multi-repo)
- Attestation mechanism (non-automatable controls)

---

### Dependency graph

```
Step 0: OpenFeature with existing env vars
  │
  └─► Step 1: Framework abstraction (proven with checks)
        │
        └─► Step 2: JSON output extension
              │
              └─► Step 3: OSPS Baseline as second framework
                    │
                    └─► Step 4: Complete L1 coverage (gap probes)
                          │
                          └─► Phase 1 complete: L1 conformance evidence
```

---

## Step 0: OpenFeature infrastructure

**Requires:** Nothing
**Enables:** Feature gating for all subsequent v6 work
**Estimated scope:** 1 PR (~200 lines)

### Problem

Scorecard has ad-hoc feature flags checked via `os.LookupEnv()`. A hard-coded delete list in `checks/all_checks.go` manually removes experimental checks. Flags don't have explicit lifecycle management. A `// UPGRADEv4: remove` comment at `options.go:251` shows cleanup doesn't happen organically.

### Solution

Introduce [OpenFeature](https://openfeature.dev/) (`github.com/open-feature/go-sdk`) with `InMemoryProvider` that reads from existing env vars. **Prove the abstraction works with existing infrastructure before using it for v6.**

**What changes:**
- Add `internal/featureflags/` package with simple API:
  ```go
  featureflags.Init()  // reads SCORECARD_V6, SCORECARD_EXPERIMENTAL
  featureflags.Enabled(ctx, "experimental") bool
  ```
- `checker.Check` gains `FeatureGate string` field
- `checks/all_checks.go` replaces hard-coded delete list with OpenFeature check
- `GetAllWithExperimental()` → `GetAllRegistered()` (used by tooling)

**What doesn't change:**
- Existing env vars (`SCORECARD_V6`, `SCORECARD_EXPERIMENTAL`) still work
- User workflows unchanged
- CLI flags unchanged

**Backward compatibility:**
- `SCORECARD_V6=1` → all `v6.*` flags evaluate to `true`
- `SCORECARD_EXPERIMENTAL=1` → `experimental` flag evaluates to `true`
- Individual flags can be set: `SCORECARD_V6_CONFORMANCE=1` → just `v6.conformance-evaluation`

**Deliverable:** Single PR introducing `internal/featureflags/`, migrating existing flags, adding `FeatureGate` to `checker.Check`. Zero behavior change - pure refactor.

---

## Step 1: Framework abstraction

**Requires:** Step 0 (feature flags operational)
**Enables:** Steps 2-5 (output format, OSPS Baseline, probe catalog)
**Estimated scope:** 2-3 PRs (~500 lines)

### Problem

Checks are hard-coded compositions of probes into 0-10 scores. There's no abstraction for "a framework that maps probes to verdicts." Before building OSPS Baseline, we need to **prove the abstraction works** with existing code.

### Solution

**Model existing checks as a framework.** This validates the abstraction with known-good code before building anything new.

**Key insight from investigation:**
- Probes produce findings (this is reusable ✅)
- Check evaluation logic produces 0-10 scores (NOT reusable for conformance ❌)
- The *pattern* is reusable: "take findings, apply evaluation rules, produce verdict"
- Don't shoehorn - checks and conformance have different evaluation semantics

**Architecture:**

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

// CheckFramework wraps existing checks as a framework
type CheckFramework struct {
    checkFn checker.CheckFn
}

func (cf CheckFramework) Evaluate(findings []finding.Finding) (Result, error) {
    // Delegate to existing check evaluation logic
    checkResult := cf.checkFn(...)
    return ScoreResult{Score: checkResult.Score}, nil
}

// ConformanceFramework evaluates probe findings against control mappings
type ConformanceFramework struct {
    mappings map[string]ControlMapping  // control ID -> probe composition
}

func (cf ConformanceFramework) Evaluate(findings []finding.Finding) (Result, error) {
    // Apply control-specific evaluation logic
    // Returns PASS/FAIL/UNKNOWN/NOT_APPLICABLE
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

**Gated behind:** `v6.framework-abstraction` flag

---

## Step 2: JSON output extension

**Requires:** Step 1 (framework abstraction exists)
**Enables:** Step 3 (conformance engine testable), Step 4 (gap probes testable)
**Estimated scope:** 1 PR (~300 lines)

### Problem

No output format supports conformance results. Before building OSPS Baseline, we need a way to serialize and validate conformance verdicts.

### Solution

Extend existing JSON output with optional conformance results. **Use existing format, not a new one.**

**JSON schema extension:**

```json
{
  "date": "2026-04-01",
  "repo": {...},
  "scorecard": {...},
  "checks": [...],  // existing check scores (unchanged)
  "conformance": {  // NEW: optional, only present when v6 enabled
    "framework": "osps-baseline",
    "version": "2026.02.19",
    "controls": [
      {
        "id": "AC-01.01",
        "status": "PASS",
        "evidence": [
          {"probe": "branchesAreProtected", "outcome": "True"},
          {"probe": "requiresApproversForPullRequests", "outcome": "True"}
        ]
      },
      {
        "id": "AC-02.01",
        "status": "NOT_APPLICABLE",
        "reason": "Requires org-level permissions not available"
      }
    ]
  }
}
```

**Design decisions:**
- Conformance is top-level field (parallel to `checks`, not nested)
- Status values: `PASS`, `FAIL`, `UNKNOWN`, `NOT_APPLICABLE`
- Evidence references probe findings by probe ID
- Multiple frameworks can coexist (checks + conformance in same output)

**What this enables:**
- Test conformance engine outputs as we build OSPS Baseline (Step 3)
- Validate gap probes produce correct evidence (Step 4)
- Existing JSON consumers unaffected (new field is optional)

**Deliverable:** Single PR extending JSON output schema, adding conformance serialization.

**Gated behind:** `v6.conformance-evaluation` flag

---

## Step 3: OSPS Baseline as second framework

**Requires:** Steps 1 (framework abstraction) + 2 (JSON output)
**Enables:** Step 4 (gap probes can be tested against L1 controls)
**Estimated scope:** 3-4 PRs (~800 lines)

### Problem

Framework abstraction is proven with checks. Now build OSPS Baseline conformance using the proven architecture.

### Solution

Implement `ConformanceFramework` for OSPS Baseline v2026.02.19.

**Components:**

1. **Control-to-probe mappings** (versioned data file)
   ```yaml
   # osps-baseline-mappings-2026.02.19.yaml
   controls:
     - id: AC-01.01
       name: "MFA enforced for contributors"
       evaluation: all-of
       probes:
         - branchesAreProtected
         - requiresApproversForPullRequests
     - id: BR-01.01
       name: "Project has a license"
       evaluation: any-of
       probes:
         - hasLicenseFile
         - hasFSFOrOSIApprovedLicense
   ```

2. **Applicability engine**
   ```go
   // Detects preconditions and outputs NOT_APPLICABLE
   func isApplicable(control Control, repo RepoMetadata) (bool, string) {
       if control.RequiresPrecondition("has-releases") && !repo.HasReleases {
           return false, "Control requires releases but project has none"
       }
       return true, ""
   }
   ```

3. **Evaluation logic**
   ```go
   func evaluateControl(control Control, findings []finding.Finding) Status {
       if !isApplicable(control, repo) {
           return NOT_APPLICABLE
       }

       probeResults := matchFindingsToProbes(control.Probes, findings)

       switch control.Evaluation {
       case "all-of":
           return allProbesPass(probeResults) ? PASS : FAIL
       case "any-of":
           return anyProbePass(probeResults) ? PASS : FAIL
       }
   }
   ```

4. **Map existing probes to L1 controls**
   - Current coverage: 6 COVERED, 8 PARTIAL, 9 GAP, 2 NOT_OBSERVABLE
   - Step 3 focuses on mapping; Step 4 closes gaps

**Deliverable:**
- PR 1: Control mapping data structure + parser
- PR 2: Applicability engine
- PR 3: Conformance evaluation logic (PASS/FAIL/UNKNOWN/NOT_APPLICABLE)
- PR 4: Map existing probes to L1 controls where coverage exists

**Dependency:**
- Add `github.com/ossf/security-baseline` as data dependency for control definitions

**Gated behind:** `v6.conformance-evaluation` flag

---

## Step 4: Complete L1 coverage (gap probes)

**Requires:** Step 3 (conformance engine operational)
**Enables:** Phase 1 completion (all L1 controls evaluatable)
**Estimated scope:** 9-12 PRs (~1,200 lines)

### Problem

OSPS Baseline L1 has 9 GAP controls where Scorecard has no coverage. Phase 1 requires complete L1 coverage.

### Solution

Write new probes for gap controls. **Metadata ingestion already exists** via `checks/fileparser/` - no new infrastructure needed.

**Gap controls and probe work:**

1. **Governance/docs presence** (4 controls)
   - GV-02.01: Governance documentation
   - GV-03.01: Contribution guidelines
   - DO-01.01: Technical documentation
   - DO-02.01: User documentation
   - **Probes:** `hasGovernanceDocs`, `hasContributingGuide`, `hasTechnicalDocs`, `hasUserDocs`
   - **Reuses:** `fileparser.OnMatchingFileContentDo()` pattern from security policy check

2. **Dependency manifest presence** (1 control)
   - QA-02.01: Dependency manifest exists
   - **Probe:** `hasDependencyManifest`
   - **Detects:** `package.json`, `requirements.txt`, `go.mod`, `Cargo.toml`, `pom.xml`, etc.

3. **Security policy deepening** (3 controls)
   - VM-02.01: Security policy has vulnerability disclosure process
   - VM-03.01: Security policy has contact information
   - VM-01.01: Security policy exists and is accessible
   - **Enhancement:** Extend existing `securityPolicy*` probes with deeper content analysis
   - **Already exists:** `securityPolicyPresent`, `securityPolicyContainsVulnerabilityDisclosure`

4. **Secrets detection** (1 control)
   - BR-07.01: No secrets in repository
   - **Probe:** `hasNoSecretsInCode`
   - **Implementation:** Consume platform signals (GitHub secret scanning API where available)

5. **Security Insights metadata** (3 controls)
   - BR-03.01: Project has documented security contacts
   - BR-03.02: Project has documented security assessment
   - QA-04.01: Project has documented quality practices
   - **Probes:** `hasSecurityContacts`, `hasSecurityAssessment`, `hasQualityPractices`
   - **Reuses:** `fileparser.OnMatchingFileContentDo()` to find `.github/security-insights.yml`
   - **Parses:** YAML content with standard library
   - **No new infrastructure needed** - this is probe work using existing fileparser

**Deliverable:**
Individual PRs per probe or probe group. Each PR includes:
- Probe implementation (`probes/*/impl.go`, `probes/*/def.yml`)
- Control mapping in OSPS Baseline mapping file
- Tests validating probe produces correct findings

**Gated behind:** Individual probe feature flags (`v6.probe.<name>`)

---

## Phase 1 Complete

**At this milestone:**
1. ✅ Scorecard produces complete OSPS Baseline Level 1 conformance evidence
2. ✅ Available via CLI and GitHub Action (production-ready JSON output)
3. ✅ Framework abstraction proven with existing checks before OSPS Baseline
4. ✅ OpenFeature infrastructure operational with existing env vars
5. ✅ All 9 L1 gap controls closed
6. ✅ Existing checks, probes, scores unchanged (v6 is additive)

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

## Open questions

**For Phase 1:**
- Should OpenFeature introduction be proposed to maintainers as an RFC/issue before implementation, or submitted directly as a PR?
- What specific downstream consumer will validate Phase 1 JSON output? (AMPEL? Privateer? LFX Insights?)

**For Phase 2+ (deferred):**
- What provider should cron/server use long-term? (flagd, environment-based, custom?)
- How should flag keys be namespaced if Scorecard supports multiple framework evaluations beyond OSPS Baseline? (e.g., `v6.framework.osps` vs. `v6.framework.slsa`)
- What is the storage/serving cost model for conformance data across 1M+ repos in cron?
