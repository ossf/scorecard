# Proposal: OSPS Baseline Conformance for OpenSSF Scorecard

## Summary

Add OSPS Baseline conformance evaluation to Scorecard, making it a credible tool for determining whether open source projects meet the security requirements defined by the Open Source Project Security (OSPS) Baseline specification. This is the central initiative for Scorecard's 2026 roadmap.

This is fundamentally a **product-level shift**: Scorecard today answers "how well does this repo follow best practices?" (graded 0-10 heuristics). OSPS conformance requires answering "does this project meet these MUST statements at this maturity level?" (PASS/FAIL/UNKNOWN/NOT_APPLICABLE per control, with evidence). The two models coexist — existing checks and scores are unchanged — but the conformance layer is a new product surface.

## Motivation

### Why now

1. **OSPS Baseline is the emerging standard.** The OSPS Baseline (v2025-10-10) defines 59 controls across 3 maturity levels. It is maintained within the ORBIT Working Group and is becoming the reference framework for open source project security posture.

2. **The ecosystem is moving.** The PVTR GitHub Repo Scanner already evaluates 39 of 52 control requirements and powers LFX Insights security results. The OSPS Baseline GitHub Action can upload SARIF. Best Practices Badge is staging Baseline-phase work. Scorecard's large install base is an advantage, but only if it ships a conformance surface.

3. **ORBIT WG alignment.** Scorecard sits within the OpenSSF alongside the ORBIT WG. The ORBIT charter's mission is "to develop and maintain interoperable resources related to the identification and presentation of security-relevant data." Scorecard producing interoperable conformance results is a natural fit.

4. **Regulatory pressure.** The EU Cyber Resilience Act (CRA) and similar regulatory frameworks increasingly expect evidence-based security posture documentation. OSPS Baseline conformance output positions Scorecard as a tool that produces CRA-relevant evidence artifacts.

### What Scorecard brings that others don't

- **Deep automated analysis.** 50+ probes with structured results provide granular evidence that PVTR's shallower checks cannot match (e.g., per-workflow token permission analysis, detailed branch protection rule inspection, CI/CD injection pattern detection).
- **Multi-platform support.** GitHub, GitLab, Azure DevOps, and local directory scanning.
- **Massive install base.** Scorecard Action, public API, and cron-based scanning infrastructure.
- **Existing policy machinery.** The `policy/` package and structured results were designed for exactly this kind of downstream consumption.

### What Scorecard must not do

- **Duplicate PVTR's role as a Privateer plugin.** PVTR is the Baseline evaluator in the ORBIT ecosystem diagram. Scorecard should complement it with deeper analysis and interoperable output, not fork the evaluation model.
- **Duplicate remediation engines.** Tools like Darn handle remediation. Scorecard exports stable, machine-readable findings for remediation tools to consume.
- **Turn OSPS controls into Scorecard checks.** OSPS conformance is a layer that consumes existing Scorecard signals, not 59 new checks.

## Current state

### Coverage snapshot (Scorecard signals vs. OSPS v2025-10-10)

| Level | Controls | Covered (✅) | Partial (⚠️) | Not covered (❌) |
|-------|----------|-------------|-------------|-----------------|
| 1     | 24       | 6           | 9           | 9               |
| 2     | 18       | 1           | 7           | 10              |
| 3     | 17       | 0           | 5           | 12              |

The full control-by-control mapping is in Appendix A of `docs/roadmap-ideas.md`.

### Existing Scorecard surfaces that matter

- **Checks** produce 0-10 scores — useful as signal but not conformance results
- **Probes** produce structured boolean findings — the right granularity for control mapping
- **Output formats** (JSON, SARIF, probe, in-toto) — OSPS output is a new format alongside these
- **Multi-repo scanning** (`--repos`, `--org`) — needed for OSPS-QA-04.02 (subproject conformance)
- **Serve mode** — HTTP surface for pipeline integration

## Open questions from maintainer review

The following questions were raised by Spencer (Steering Committee member) during review of the roadmap and need to be resolved before or during implementation:

### OQ-1: Attestation mechanism identity

> "The attestation/provenance layer. What is doing the attestation? Is this some OIDC? A personal token? A workflow (won't have the right tokens)?"
> — Spencer, on Section 5.1

This is a fundamental design question. Options include:
- **Repo-local metadata files** (e.g., Security Insights, `.osps-attestations.yml`): simplest, no cryptographic identity, maintainer self-declares by committing the file.
- **Signed attestations via Sigstore/OIDC**: strongest guarantees, but requires workflow identity and the right tokens — which Spencer correctly notes may not be available in all contexts.
- **Platform-native signals**: e.g., GitHub's private vulnerability reporting enabled status, which the platform attests implicitly.

**Recommendation to discuss**: Start with repo-local metadata files (unsigned) for the v1 attestation mechanism, with a defined extension point for signed attestations in a future iteration. This avoids blocking on the identity question while still making non-automatable controls reportable.

### OQ-2: Scorecard's role in enforcement detection vs. enforcement

> "I thought the other doc said Scorecard wasn't an enforcement tool?"
> — Spencer, on Q4 deliverables (enforcement detection)

This is a critical framing question. The roadmap proposes *detecting* whether enforcement exists (e.g., "are SAST results required to pass before merge?"), not *performing* enforcement. But the line between "detecting enforcement" and "being an enforcement tool" needs to be drawn clearly.

**Recommendation to discuss**: Scorecard detects and reports whether enforcement mechanisms are in place. It does not itself enforce. The `--fail-on=fail` CI gating is a reporting exit code, not an enforcement action — the CI system is the enforcer. This distinction should be documented explicitly.

### OQ-3: `scan_scope` field in output schema

> "Not sure I see the importance [of `scan_scope`]"
> — Spencer, on Section 9 (output schema)

The `scan_scope` field (repo|org|repos) in the proposed OSPS output schema may not carry meaningful information. If the output always describes a single repository's conformance, the scope is implicit.

**Recommendation to discuss**: Drop `scan_scope` from the schema unless multi-repo aggregation (OSPS-QA-04.02) produces a fundamentally different output shape. Revisit in Q4 when project-level aggregation is implemented.

### OQ-4: Evidence model — probes only, not checks

> "[Evidence] should be probe-based only, not check"
> — Spencer, on Section 9 (output schema)

Spencer's position is that OSPS evidence references should point to probe findings, not check-level results. This aligns with the architectural direction of Scorecard v5 (probes as the measurement unit, checks as scoring aggregations).

**Recommendation**: Adopt this. The `evidence` array in the OSPS output schema should reference probes and their findings only. Checks may be listed in a `derived_from` field for human context but are not evidence.

## Scope

### In scope

1. **OSPS conformance engine** — new package that maps controls to Scorecard probes, evaluates per-control status, handles applicability
2. **OSPS output format** — `--format=osps` producing a JSON conformance report
3. **Versioned mapping file** — data-driven YAML mapping OSPS control IDs to Scorecard probes, applicability rules, and evaluation logic
4. **Applicability engine** — detects preconditions (e.g., "has made a release") and outputs NOT_APPLICABLE
5. **Security Insights ingestion** — reads `security-insights.yml` to satisfy metadata-dependent controls, aligning with the ORBIT ecosystem data plane
6. **Attestation mechanism (v1)** — accepts repo-local metadata for non-automatable controls (pending OQ-1 resolution)
7. **New probes and probe enhancements** for gap controls:
   - Secrets detection (OSPS-BR-07.01)
   - Governance/docs presence (OSPS-GV-02.01, GV-03.01, DO-01.01, DO-02.01)
   - Dependency manifest presence (OSPS-QA-02.01)
   - Security policy deepening (OSPS-VM-02.01, VM-03.01, VM-01.01)
   - Release asset inspection (multiple L2/L3 controls)
   - Signed manifest support (OSPS-BR-06.01)
   - Enforcement detection (OSPS-VM-05.*, VM-06.* — pending OQ-2 resolution)
8. **CI gating** — `--fail-on=fail` exit code for pipeline integration
9. **Multi-repo project-level conformance** (OSPS-QA-04.02)
10. **Gemara Layer 4 compatibility** — output structurally compatible with ORBIT assessment result schemas

### Out of scope

- Remediation automation (Darn's domain)
- Replacing PVTR as a Privateer plugin
- Changing existing check scores or behavior
- Platform enforcement (Minder's domain)
- OSPS Baseline specification changes (ORBIT WG's domain)

## Phased delivery

### Q1 2026: OSPS conformance alpha + ecosystem handshake

- OSPS output format (alpha) with `--format=osps`
- Versioned mapping file for v2025-10-10 (alpha)
- Applicability engine v1
- Design + document ORBIT interop commitments (Security Insights, Gemara compatibility, PVTR complementarity)
- Map existing probes to OSPS controls where coverage already exists

### Q2 2026: Level 1 coverage spike + declared metadata v1

- Secrets detection probe (OSPS-BR-07.01)
- Governance + docs presence probes (GV-02.01, GV-03.01, DO-01.01, DO-02.01)
- Dependency manifest presence probe (QA-02.01)
- Security policy deepening (VM-02.01, VM-03.01, VM-01.01)
- Security Insights ingestion v1 (BR-03.01, BR-03.02, QA-04.01)
- Attestation mechanism v1 for non-automatable controls
- CI gating: `--fail-on=fail` + coverage summary

### Q3 2026: Release integrity + Level 2 core

- Release asset inspection layer (detect compiled assets, SBOMs, licenses with releases)
- Signed manifest support (BR-06.01)
- Release notes/changelog detection (BR-04.01)
- Evidence bundle output v1 (OSPS result JSON + in-toto statement + SARIF for failures)

### Q4 2026: Enforcement detection + Level 3 + multi-repo

- SCA policy + enforcement detection (VM-05.*)
- SAST policy + enforcement detection (VM-06.*)
- Multi-repo project-level conformance aggregation (QA-04.02)
- Attestation integration GA

## Relationship to ORBIT ecosystem

```
┌─────────────────────────────────────────────────────┐
│                   ORBIT WG Ecosystem                │
│                                                     │
│  ┌──────────┐  ┌───────────┐  ┌──────────────────┐  │
│  │  OSPS    │  │  Gemara   │  │    Security       │  │
│  │ Baseline │  │  Schemas  │  │    Insights       │  │
│  │ (controls│  │  (L2/L4)  │  │    (metadata)     │  │
│  └────┬─────┘  └─────┬─────┘  └────────┬─────────┘  │
│       │              │                  │            │
│       ▼              ▼                  ▼            │
│  ┌─────────────────────────────────────────────┐     │
│  │         PVTR GitHub Repo Scanner            │     │
│  │  (Privateer plugin, LFX Insights driver)    │     │
│  └─────────────────────┬───────────────────────┘     │
│                        │                             │
│                  consumes ▲                           │
│                        │                             │
│  ┌─────────────────────┴───────────────────────┐     │
│  │         OpenSSF Scorecard                   │     │
│  │  (deep analysis, conformance output,        │     │
│  │   multi-platform, large install base)       │     │
│  └─────────────────────────────────────────────┘     │
│                                                     │
│  ┌──────────┐  ┌───────────┐                         │
│  │  Minder  │  │   Darn    │                         │
│  │ (enforce)│  │ (remediate│                         │
│  └──────────┘  └───────────┘                         │
└─────────────────────────────────────────────────────┘
```

**Scorecard's role**: Produce deep, probe-based conformance evidence that PVTR, Minder, and downstream consumers can use. Scorecard reads Security Insights (shared data plane), outputs Gemara L4-compatible results (shared schema), and fills analysis gaps where PVTR has `NotImplemented` steps.

**What Scorecard does NOT do**: Replace PVTR as the Privateer plugin, enforce policies (Minder's role), or perform remediation (Darn's role).

## Success criteria

1. `scorecard --format=osps --osps-level=1` produces a valid conformance report for any public GitHub repository
2. Level 1 auto-check coverage reaches ≥80% of controls (currently ~25%)
3. OSPS output is consumable by PVTR as supplementary evidence (validated with ORBIT WG)
4. All four open questions (OQ-1 through OQ-4) are resolved with documented decisions
5. No changes to existing check scores or behavior
