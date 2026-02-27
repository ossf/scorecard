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

---

## Maintainer review

### Stephen's notes

<!-- Stephen: Use this section to record your overall impressions, concerns,
     and positions on the proposal. Edit freely — this is your space. -->

**Overall assessment:**


**Key concerns or risks:**


**Things I agree with:**


**Things I disagree with or want to change:**

- "PVTR" is shorthand for "Privateer". Throughout this proposal it makes it appear as if https://github.com/ossf/pvtr-github-repo-scanner is separate from Privateer, when it is really THE Privateer plugin for GitHub repositories. Any references to PVTR should be corrected.
- This proposal does not contain an even consideration of the capabilities of [Darnit](https://github.com/kusari-oss/darnit) and [AMPEL](https://github.com/carabiner-dev/ampel). We should do that comparison to get a better idea of what should be in or out of scope for Scorecard.
- The timeline that is in this proposal is not accurate, as we're already about to enter Q2 2026. We should focus on phases and outcomes, and let maintainer bandwidth dictate delivery timing.
- Scorecard has an existing set of checks and probes, which is essentially a control catalog. We should make a plan to extract the Scorecard control catalog so that it can be used by other tools that can handle evaluation tasks.
- Use Mermaid when creating diagrams.
- We need to understand what level of coverage Scorecard currently has for OSPS Baseline and that analysis should be created in a separate file (in `docs/`). Assume that any existing findings are out-of-date.
- `docs/roadmap-ideas.md` will not be committed to the repo, as it is a rough draft which needs to be refined for public consumption. We should create `docs/ROADMAP.md` with a 2026 second-level heading with contains the publicly-consummable roadmap.

**Priority ordering — what matters most to ship first:**


### Clarifying questions

The following questions need your input before this proposal can move to design. Please fill in your response under each question.

#### CQ-1: Scorecard as a conformance tool — product identity

The proposal frames this as a "product-level shift" where Scorecard gains a second mode: conformance evaluation alongside its existing scoring. Does this framing match your vision, or do you see conformance as eventually *replacing* the scoring model? Should we be thinking about deprecating 0-10 scores long-term, or do both modes coexist indefinitely?

**Stephen's response:**

I believe the scoring model will continue to be useful to consumers and it should be maintained. For now, both modes should coexist. There is no need to make a decision about this for the current iteration of the proposal.

#### CQ-2: OSPS Baseline version targeting

The roadmap targets OSPS Baseline v2025-10-10. The PVTR scanner targets v2025-02-25. The Baseline is a living spec with periodic releases. How should Scorecard handle version drift? Options:
- Support only the latest version at any given time
- Support multiple versions concurrently via the versioned mapping file
- Pin to a version and update on a defined cadence (e.g., quarterly)

**Stephen's response:**

The current version of the OSPS Baseline is [v2026.02.19](https://baseline.openssf.org/versions/2026-02-19).

We should align with the latest version at first and have a process for aligning with new versions on a defined cadence. We should understand the [OSPS Baseline maintenance process](https://baseline.openssf.org/maintenance.html) and align with it.

The OSPS Baseline [FAQ](https://baseline.openssf.org/faq.html) and [Implementation Guidance for Maintainers](https://baseline.openssf.org/maintainers.html) may have guidance we should consider incorporating.

#### CQ-3: Security Insights as a hard dependency

Many OSPS controls depend on Security Insights data (official channels, distribution points, subproject inventory, core team). PVTR treats the Security Insights file as nearly required — most of its evaluation steps begin with `HasSecurityInsightsFile`.

Should Scorecard:
- Treat Security Insights the same way (controls that need it go UNKNOWN without it)?
- Provide a degraded but still useful evaluation without it?
- Accept alternative metadata sources (e.g., `.project`, custom config)?

This also raises a broader adoption question: most projects today don't have a `security-insights.yml`. How do we avoid making the OSPS output useless for the majority of repositories?

**Stephen's response:**

We should provide a degraded, but still-useful evaluation without a Security Insights file, especially since our probes today can already cover a lot of ground without it. It would be good for us to eventually support alternative metadata sources, but this should not be an immediate goal.

#### CQ-4: PVTR relationship — complement vs. converge

The proposal positions Scorecard as complementary to PVTR. But there's a deeper question: should this stay as two separate tools indefinitely, or is the long-term goal convergence (e.g., PVTR consuming Scorecard as a library, or Scorecard becoming a Privateer plugin itself)? Your position on this affects how tightly we couple the output formats and whether we invest in Gemara SDK integration.

**Stephen's response:**

Multiple tools should be able to consume Scorecard, so yes, we should invest in Gemara SDK integration.

#### CQ-5: Scope of new probes in 2026

The roadmap calls for significant new probe development (secrets detection, governance/docs presence, dependency manifests, release asset inspection, enforcement detection). That's a lot of new surface area. Should we:
- Build all of these within Scorecard?
- Prioritize a subset and defer the rest?
- Look for ways to consume signals from external tools (e.g., GitHub's secret scanning API, SBOM tools) rather than building detection from scratch?

If prioritizing, which new probes matter most to you?

**Stephen's response:**

We should prioritize OSPS Baseline Level 1 conformance work.
We should consider any signals that can be consumed from external sources.

#### CQ-6: Community and governance process

This is a major initiative touching Scorecard's product direction. What's the governance process for getting this approved?
- Does this need a formal proposal to the Scorecard maintainer group?
- Should this be presented at an ORBIT WG meeting?
- Do we need sign-off from the OpenSSF TAC?
- Who else beyond you and Spencer needs to weigh in?

**Stephen's response:**

We should have Stephen and Spencer sign off on this proposal as Steering Committee members. In addition, we should have reviews from:
- [blocking] At least 1 non-Steering Scorecard maintainer
- [non-blocking] Maintainers of tools in the WG ORBIT ecosystem

This does not require review from the TAC, but we should inform WG ORBIT members.

#### CQ-7: The "minimum viable conformance report"

If we had to ship the smallest useful thing in Q1, what would it be? The roadmap proposes the full OSPS output format + mapping file + applicability engine. But a simpler starting point might be:
- Just the mapping file (documentation-only, no runtime)
- A `--format=osps` that only reports on controls Scorecard already covers (no new probes, lots of UNKNOWN)
- Something else?

What would make Q1 a success in your eyes?

**Stephen's response:**

As previously mentioned, the quarterly targets are not currently accurate. One of our Q2 outcomes should be OSPS Baseline Level 1 conformance.

#### CQ-8: Existing Scorecard Action and API impact

Scorecard runs at scale via the Scorecard Action (GitHub Action) and the public API (api.scorecard.dev). Should OSPS conformance be available through these surfaces from day one, or should it start as a CLI-only feature? The API and Action have their own release and stability considerations.

**Stephen's response:**

We need to land these capabilities for as much surface area as possible.
