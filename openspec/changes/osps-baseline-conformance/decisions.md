# OSPS Baseline Conformance — Feedback and Decisions

Companion document to [`proposal.md`](proposal.md). This document tracks
reviewer feedback, open questions, maintainer responses, and the decision
priority analysis.

For the proposal itself (motivation, scope, phased delivery, ecosystem
positioning), see [`proposal.md`](proposal.md).

For the control-by-control coverage analysis, see
[`docs/osps-baseline-coverage.md`](../../docs/osps-baseline-coverage.md).

---

## Open questions from maintainer review

The following questions were raised by Spencer (Steering Committee member)
during review of the roadmap and need to be resolved before or during
implementation.

### OQ-1: Attestation mechanism identity

> "The attestation/provenance layer. What is doing the attestation? Is this some OIDC? A personal token? A workflow (won't have the right tokens)?"
> — Spencer, on Section 5.1

**Stakeholders:** Spencer (raised this, flagged as blocking), Stephen, Steering Committee

This is a fundamental design question. Options include:
- **Repo-local metadata files** (e.g., Security Insights, `.osps-attestations.yml`): simplest, no cryptographic identity, maintainer self-declares by committing the file.
- **Signed attestations via Sigstore/OIDC**: strongest guarantees, but requires workflow identity and the right tokens — which Spencer correctly notes may not be available in all contexts.
- **Platform-native signals**: e.g., GitHub's private vulnerability reporting enabled status, which the platform attests implicitly.

**Recommendation to discuss**: Start with repo-local metadata files (unsigned) for the v1 attestation mechanism, with a defined extension point for signed attestations in a future iteration. This avoids blocking on the identity question while still making non-automatable controls reportable.

### OQ-2: Scorecard's role in enforcement detection vs. enforcement

> "I thought the other doc said Scorecard wasn't an enforcement tool?"
> — Spencer, on Q4 deliverables (enforcement detection)

**Stakeholders:** Spencer (raised this), Stephen, Steering Committee

This is a critical framing question. The roadmap proposes *detecting* whether enforcement exists (e.g., "are SAST results required to pass before merge?"), not *performing* enforcement. But the line between "detecting enforcement" and "being an enforcement tool" needs to be drawn clearly.

**Recommendation to discuss**: Scorecard detects and reports whether enforcement mechanisms are in place. It does not itself enforce. The `--fail-on=fail` CI gating is a reporting exit code, not an enforcement action — the CI system is the enforcer. This distinction should be documented explicitly.

### OQ-3: `scan_scope` field in output schema

> "Not sure I see the importance [of `scan_scope`]"
> — Spencer, on Section 9 (output schema)

**Stakeholders:** Stephen (can resolve alone)

The `scan_scope` field (repo|org|repos) in the proposed OSPS output schema may not carry meaningful information. If the output always describes a single repository's conformance, the scope is implicit.

**Recommendation to discuss**: Drop `scan_scope` from the schema unless multi-repo aggregation (OSPS-QA-04.02) produces a fundamentally different output shape. Revisit when project-level aggregation is implemented.

### OQ-4: Evidence model — probes only, not checks

> "[Evidence] should be probe-based only, not check"
> — Spencer, on Section 9 (output schema)

**Stakeholders:** Spencer (raised this), Stephen — effectively resolved (adopted)

Spencer's position is that OSPS evidence references should point to probe findings, not check-level results. This aligns with the architectural direction of Scorecard v5 (probes as the measurement unit, checks as scoring aggregations).

**Recommendation**: Adopt this. The `evidence` array in the OSPS output schema should reference probes and their findings only. Checks may be listed in a `derived_from` field for human context but are not evidence.

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

The following questions need input before this proposal can move to design.
Questions with Stephen's responses are answered; the rest are open.

#### CQ-1: Scorecard as a conformance tool — product identity

The proposal frames this as a "product-level shift" where Scorecard gains a second mode: conformance evaluation alongside its existing scoring. Does this framing match your vision, or do you see conformance as eventually *replacing* the scoring model? Should we be thinking about deprecating 0-10 scores long-term, or do both modes coexist indefinitely?

**Stephen's response:**

I believe the scoring model will continue to be useful to consumers and it should be maintained. For now, both modes should coexist. There is no need to make a decision about this for the current iteration of the proposal.

**Update:** Check scores and conformance labels are *parallel evaluation layers* over the same probe evidence, not two competing "modes." Both can appear in the same output. The three-tier architecture model (evidence layer → evaluation layers → output formats) replaces the original "two modes" framing. OSPS conformance is *one goal*, not *the* goal — Scorecard's broader identity is as an open source security evidence engine.

#### CQ-2: OSPS Baseline version targeting

The roadmap previously targeted OSPS Baseline v2025-10-10. The Privateer GitHub plugin targets v2025-02-25. The Baseline is a living spec with periodic releases. How should Scorecard handle version drift? Options:
- Support only the latest version at any given time
- Support multiple versions concurrently via the versioned mapping file
- Pin to a version and update on a defined cadence (e.g., quarterly)

**Stephen's response:**

The current version of the OSPS Baseline is [v2026.02.19](https://baseline.openssf.org/versions/2026-02-19).

We should align with the latest version at first and have a process for aligning with new versions on a defined cadence. We should understand the [OSPS Baseline maintenance process](https://baseline.openssf.org/maintenance.html) and align with it.

The OSPS Baseline [FAQ](https://baseline.openssf.org/faq.html) and [Implementation Guidance for Maintainers](https://baseline.openssf.org/maintainers.html) may have guidance we should consider incorporating.

#### CQ-3: Security Insights as a hard dependency

Many OSPS controls depend on Security Insights data (official channels, distribution points, subproject inventory, core team). The Privateer GitHub plugin treats the Security Insights file as nearly required — most of its evaluation steps begin with `HasSecurityInsightsFile`.

Should Scorecard:
- Treat Security Insights the same way (controls that need it go UNKNOWN without it)?
- Provide a degraded but still useful evaluation without it?
- Accept alternative metadata sources (e.g., `.project`, custom config)?

This also raises a broader adoption question: most projects today don't have a `security-insights.yml`. How do we avoid making the OSPS output useless for the majority of repositories?

**Stephen's response:**

We should provide a degraded, but still-useful evaluation without a Security Insights file, especially since our probes today can already cover a lot of ground without it. It would be good for us to eventually support alternative metadata sources, but this should not be an immediate goal.

**Update:** Reframed as a "metadata ingestion layer" that supports Security Insights as one source among several. SI is not privileged. Architecture invites contributions for alternative sources (SBOMs, VEX, platform APIs). No single metadata file is required for meaningful results, though metadata files may enrich results.

#### CQ-4: PVTR relationship — complement vs. converge

The proposal positions Scorecard as complementary to the Privateer plugin. But there's a deeper question: should this stay as two separate tools indefinitely, or is the long-term goal convergence (e.g., the Privateer plugin consuming Scorecard as a library, or Scorecard becoming a Privateer plugin itself)? Your position on this affects how tightly we couple the output formats and whether we invest in Gemara SDK integration.

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

What would make Q2 a success in your eyes?

**Stephen's response:**

As previously mentioned, the quarterly targets are not currently accurate. One of our Q2 outcomes should be OSPS Baseline Level 1 conformance.

#### CQ-8: Existing Scorecard Action and API impact

Scorecard runs at scale via the Scorecard Action (GitHub Action) and the public API (api.scorecard.dev). Should OSPS conformance be available through these surfaces from day one, or should it start as a CLI-only feature? The API and Action have their own release and stability considerations.

**Stephen's response:**

We need to land these capabilities for as much surface area as possible.

#### CQ-9: Coverage analysis and Phase 1 scope validation

**Stakeholders:** Stephen (can answer alone)

The coverage analysis (`docs/osps-baseline-coverage.md`) identifies 25 Level 1 controls. Of those, 6 are COVERED, 8 are PARTIAL, 9 are GAP, and 2 are NOT_OBSERVABLE. The Phase 1 plan targets closing the 9 GAP controls. Given that 2 controls (AC-01.01, AC-02.01) are NOT_OBSERVABLE without org-admin tokens, should Phase 1 explicitly include work on improving observability (e.g., documenting what tokens are needed, or providing guidance for org admins), or should those controls remain UNKNOWN until a later phase?

**Stephen's response:**


#### CQ-10: Mapping file ownership and contribution model

**Stakeholders:** Stephen, Eddie Knight, Baseline maintainers — partially superseded by CQ-17

The versioned mapping file (e.g., `pkg/osps/mappings/v2026-02-19.yaml`) is a critical artifact that defines which probes satisfy which OSPS controls. Who should own this file? Options:
- Scorecard maintainers only (changes require maintainer review)
- Community-contributed with maintainer approval (like checks/probes today)
- Co-maintained with ORBIT WG members who understand the Baseline controls

This also affects how we handle disagreements about whether a probe truly satisfies a control.

**Stephen's response:**


#### CQ-11: Backwards compatibility of OSPS output format

**Stakeholders:** Stephen, Spencer, Eddie Knight — depends on CQ-18 (output format decision)

The spec requires `--format=osps` as a new output format. Since this is a new surface, we have freedom to iterate on the schema. However, once shipped, consumers will depend on it. What stability guarantees should we offer?
- No guarantees during Phase 1 (alpha schema, may break between releases)
- Semver-like schema versioning from day one (breaking changes increment major version)
- Follow the Gemara L4 schema if one exists, inheriting its stability model

**Stephen's response:**


#### CQ-12: Probe gap prioritization for Phase 1

**Stakeholders:** Stephen (can answer alone)

The coverage analysis identifies 7 Level 1 GAP controls that need new probes (excluding the 2 that depend on Security Insights). Ranked by implementation feasibility:

1. OSPS-GV-03.01 — CONTRIBUTING file presence
2. OSPS-GV-02.01 — Issues/discussions enabled
3. OSPS-DO-02.01 — Issue templates or bug report docs
4. OSPS-DO-01.01 — Documentation presence heuristics
5. OSPS-BR-07.01 — Secrets detection (platform signal consumption)
6. OSPS-BR-03.01 / BR-03.02 — Encrypted transport (requires Security Insights)
7. OSPS-QA-04.01 — Subproject listing (requires Security Insights)

Do you agree with this priority ordering? Are there any controls you would move up or down, or any you would defer to Phase 2?

**Stephen's response:**


#### CQ-13: Minder and AMPEL integration surfaces

**Stakeholders:** Stephen, Minder maintainers, Adolfo García Veytia (AMPEL), Steering Committee

Two tools already consume Scorecard data for policy enforcement:

**[Minder](https://github.com/mindersec/minder)** (OpenSSF Sandbox, ORBIT WG) consumes Scorecard findings to enforce security policies and auto-remediate across repositories. Uses Rego-based rules and can enforce OSPS Baseline controls via policy profiles. A draft Scorecard PR (#4723, now stale) attempted deeper integration.

**[AMPEL](https://github.com/carabiner-dev/ampel)** (production v1.0.0) validates Scorecard attestations against policies in CI/CD pipelines. Already maintains [5 Scorecard-consuming policies](https://github.com/carabiner-dev/policies/tree/main/scorecard) and [36 OSPS Baseline policy mappings](https://github.com/carabiner-dev/policies/tree/main/groups/osps-baseline). Uses CEL expressions and in-toto attestations.

Questions:
- Should the OSPS conformance output be designed with Minder and AMPEL as explicit consumers (e.g., ensuring the output works as Minder policy input and as AMPEL attestation input)?
- Should we coordinate with both Minder maintainers and Adolfo during Phase 1 to validate the integration surface?
- Is there a risk of duplicating Baseline evaluation work that Minder or AMPEL already do via their own rules, and if so, how should we delineate?

**Stephen's response:**

All downstream tools — Privateer, AMPEL, Minder, Darnit, and others — are equal consumers of Scorecard's output formats. The output formats should serve different tool types equally (policy, remediation, dashboarding).

Scorecard SHOULD NOT (per [RFC 2119](https://www.rfc-editor.org/rfc/rfc2119)) duplicate evaluation that downstream tools handle, but this is not a MUST NOT. There could be scenarios where overlapping evaluation makes sense (e.g., Scorecard brings deeper analysis or different evidence sources).

Coordinate with downstream tool maintainers during Phase 1 to validate that output formats are consumable.


#### CQ-14: Darnit vs. Minder delineation

**Stakeholders:** Stephen (can answer alone)

The proposal lists both [Darnit](https://github.com/kusari-oss/darnit) and [Minder](https://github.com/mindersec/minder) as tools that handle remediation and enforcement. Their capabilities overlap in some areas (both can enforce Baseline controls, both can remediate). For Scorecard's purposes, the distinction matters primarily for the "What Scorecard must not do" boundary.

Is the current framing correct — that Scorecard is the measurement layer and both Minder and Darnit are downstream consumers? Or should we position Scorecard differently relative to one versus the other, given that Minder is an OpenSSF project in the same working group while Darnit is not?

**Stephen's response:**


#### CQ-15: Existing issues as Phase 1 work items

**Stakeholders:** Stephen (can answer alone)

The coverage analysis (`docs/osps-baseline-coverage.md`) now includes a section mapping existing Scorecard issues to OSPS Baseline gaps. Several long-standing issues align directly with Phase 1 priorities:

- [#30](https://github.com/ossf/scorecard/issues/30) — Secrets scanning (OSPS-BR-07.01), open since the project's earliest days
- [#2305](https://github.com/ossf/scorecard/issues/2305) / [#2479](https://github.com/ossf/scorecard/issues/2479) — Security Insights ingestion
- [#2465](https://github.com/ossf/scorecard/issues/2465) — Private vulnerability reporting (OSPS-VM-03.01)
- [#4824](https://github.com/ossf/scorecard/issues/4824) — Changelog check (OSPS-BR-04.01)
- [#4723](https://github.com/ossf/scorecard/pull/4723) — Minder/Rego integration draft (closed)

Should we adopt these existing issues as the starting work items for Phase 1, or create new issues that reference them? Some of these issues have significant discussion history that may contain design decisions worth preserving.

**Stephen's response:**


---

## ORBIT WG feedback

### Eddie Knight's feedback (ORBIT WG TSC Chair)

The following feedback was provided by Eddie Knight (ORBIT WG Technical Steering Committee Chair, maintainer of Gemara, Privateer, and OSPS Baseline) on [PR #4952](https://github.com/ossf/scorecard/pull/4952).

#### EK-1: Mapping file location

> "Regarding mappings between Baseline catalog<->Scorecard checks, it is possible to easily put that into a new file with Scorecard maintainers as codeowners, pending approval from OSPS Baseline maintainers for the change."

Eddie is offering to host the Baseline-to-Scorecard mapping in the OSPS Baseline repository (or a shared location) with Scorecard maintainers as CODEOWNERS. The current proposal places the mapping in the Scorecard repo (`pkg/osps/mappings/v2026-02-19.yaml`).

Mappings currently exist within the Baseline Catalog and are proposed for addition to the Scorecard repository as well. The mappings could be maintained in one or both of the projects. This affects ownership, versioning cadence, and who can update the mapping when controls or probes change.

The trade-offs:

- **In Scorecard repo**: Scorecard maintainers fully own the mapping. Mapping updates are coupled to Scorecard releases. Other tools cannot easily consume the mapping.
- **In Baseline repo (or shared)**: Mapping is co-owned. Versioned alongside the Baseline spec. End users and other tools (Privateer, Darnit, Minder) can consume the same mapping. Scorecard maintainers retain CODEOWNERS authority.

#### EK-2: Output format — no "OSPS output format"

> "There is not an 'OSPS output format,' and even the relevant Gemara schemas (which are quite opinionated) are still designed to support output in multiple output formats within the SDK, such as SARIF. I would expect that you'd keep your current output logic, and then _maybe_ add basic Gemara json/yaml as another option."

The current proposal defines `--format=osps` as a new output format. Eddie clarifies that the ORBIT ecosystem does not define a special "OSPS output format" — instead, the Gemara SDK supports multiple output formats (including SARIF). The suggestion is to keep Scorecard's existing output logic and optionally add Gemara JSON/YAML as another format option.

This is a significant clarification that affects the output requirements, the Phase 1 deliverables, and how we frame the conformance layer.

#### EK-3: Technical relationship with Privateer plugin

> "There is a stated goal of not duplicating the code from the plugin ossf/pvtr-github-repo-scanner, but the implementation plan as it's currently written does require duplication. In the current proposal, there would not be a technical relationship between the two codebases."

Eddie identifies a contradiction: the proposal says "do not duplicate Privateer" but proposes building a parallel conformance engine with no code-level relationship to the Privateer plugin. The current plan would result in two separate codebases evaluating the same OSPS controls independently.

#### EK-4: Catalog extraction needs an implementation plan

> "There is cursory mention of a scorecard _catalog extraction_, which I'm hugely in favor of, but I don't see an implementation plan for that."

The proposal mentions "Scorecard control catalog extraction plan" as a Phase 1 deliverable but does not specify what this means concretely or how it would be achieved.

#### EK-5: Alternative architecture — shared plugin model

> "An alternative plan would be for us to spend a week consolidating checks/probes into the pvtr plugin (with relevant CODEOWNERS), then update Scorecard to selectively execute the plugin under the covers."

Eddie proposes a fundamentally different architecture:

1. Consolidate Scorecard checks/probes into the [Privateer plugin](https://github.com/ossf/pvtr-github-repo-scanner) as shared evaluation logic
2. Scorecard executes the plugin under the covers for Baseline evaluation and then Scorecard handles follow-up logic such as scoring and storing the results
3. Privateer and LFX Insights can optionally run Scorecard checks via the same plugin

**Claimed benefits:**
- Extract the Scorecard control catalog for independent versioning and cross-catalog mapping to Baseline
- Instantly integrate Gemara into Scorecard
- Allow bidirectional check execution (Scorecard runs Privateer checks; Privateer runs Scorecard checks)
- Simplify contribution overhead for individual checks
- Improve both codebases through shared logic

**This is the central architectural decision for the proposal.** The Steering Committee needs to evaluate this against the current plan (Scorecard builds its own conformance engine).

### Adolfo García Veytia's feedback (AMPEL maintainer)

The following feedback was provided by Adolfo García Veytia (@puerco, maintainer of [AMPEL](https://github.com/carabiner-dev/ampel)) on [PR #4952](https://github.com/ossf/scorecard/pull/4952).

#### AP-1: Mapping file registry — single source preferred

> "It's great that you also see the need for machine-readable data. This would help projects like AMPEL write policies that enforce the baseline controls based on the results from Scorecard and other analysis tools."
>
> "Initially, we were trying to build the mappings into baseline itself. I still think it's the way to go as it would be better to have a single registry and data format of those mappings (in this case baseline's). Unfortunately, the way baseline considers its mappings [was demoted](https://github.com/ossf/security-baseline/pull/476) so we don't have that registry anymore."

Adolfo strongly supports machine-readable mapping data and prefers a single registry in the Baseline itself, though the Baseline's own mapping support was recently demoted (PR #476 in security-baseline). This aligns with Eddie's offer (EK-1) to host mappings in the Baseline repo, but adds the context that there is no longer an official registry for tool-to-control mappings.

AMPEL already maintains its own [Scorecard-to-Baseline policy mappings](https://github.com/carabiner-dev/policies/tree/main/groups/osps-baseline) (36 OSPS control policies, 5 of which directly consume Scorecard probe results). An official upstream mapping from Scorecard would benefit the entire ecosystem.

#### AP-2: Output format — use in-toto predicates, not a custom format

> "As others have mentioned, there is no _OSPS output format_ but there are two formal/in process of formalizing in-toto predicate types that are useful for this:
>
> **[Simple Verification Results](https://github.com/in-toto/attestation/blob/main/spec/predicates/svr.md)** — a simple predicate that communicates just the verified control labels along with the tool that performed the evaluation. It is a generalization of the VSA for non-SLSA controls.
>
> **[The "Baseline" Predicate](https://github.com/in-toto/attestation/pull/502)** — Still not merged, this predicate type was proposed by some of the baseline maintainers to capture an interoperability format more in line with the requirements in this spec, including manual assessments (what is named in this PR as 'ATTESTED')."

Adolfo identifies two concrete in-toto predicate types that Scorecard should consider for output instead of inventing a custom format:

1. **Simple Verification Results (SVR)**: Already merged in the in-toto attestation spec. Communicates verified control labels and the evaluating tool. Generalizes SLSA VSA to non-SLSA controls.
2. **Baseline Predicate**: Proposed by Baseline maintainers (PR #502, not yet merged). Designed for interoperability and includes support for manual assessments (ATTESTED status).

This is the most concrete guidance on output format so far and directly informs CQ-18.

#### AP-3: Attestation question conflates identity and tooling

> "The question here is conflating two domains. One question is _who_ signs the attestation, and how can those identities be trusted (identity). The other is _what_ (tool) generates the attestations, and more importantly, from scorecard's perspective, when. This hints at a policy implementation and the answers will most likely differ for projects and controls. Happy to chat about this one day."

Adolfo clarifies that OQ-1 (attestation mechanism identity) is actually two separate questions:
1. **Identity**: Who signs the attestation, and how are those identities trusted? (Sigstore/OIDC, personal keys, etc.)
2. **Tooling**: What tool generates the attestation, and when? (Scorecard during scan, CI pipeline, manual process)

The answers will differ per project and per control. This decomposition should inform how OQ-1 is resolved.

#### AP-4: AMPEL already consumes Scorecard data for Baseline enforcement

> "I agree with this role statement. Just as minder, ampel also can enforce Scorecard's data ([see an example here](https://github.com/carabiner-dev/policies/blob/ab1eb42ef179c7a0016d6b7ed72991774a48f151/scorecard/sast.json#L4)) and we also [maintain a mapping of some of scorecard's probes vs baseline controls](https://github.com/carabiner-dev/policies/blob/ab1eb42ef179c7a0016d6b7ed72991774a48f151/groups/osps-baseline/osps-vm-06.hjson#L5) that would greatly benefit from an official/upstream map.
>
> The probes can enrich the baseline ecosystem substantially and having the data accessible from other tools encourages other projects in the ecosystem to help maintain and improve them."

AMPEL is an active consumer of Scorecard data today:
- 5 production policies directly evaluate Scorecard probe results (SAST, binary artifacts, code review, dangerous workflows, token permissions)
- 36 OSPS Baseline policy mappings, several of which reference Scorecard checks
- An official upstream Scorecard-to-Baseline mapping would directly benefit AMPEL's policy library

This validates the proposal's direction of making Scorecard's probe results and control mappings available to the broader ecosystem.

### Mike Lieberman's feedback

The following feedback was provided by Mike Lieberman (@mlieberman85) on [PR #4952](https://github.com/ossf/scorecard/pull/4952).

#### ML-1: No "OSPS output format" exists

> "What is OSPS output format?"
> — on ROADMAP.md, Phase 1 deliverable

Mike echoes Eddie's (EK-2) and Adolfo's (AP-2) point: there is no defined "OSPS output format." This is the third reviewer to flag this, confirming it needs to be reframed. The output format question (CQ-18) now has concrete alternatives: Gemara SDK formats (Eddie), in-toto SVR/Baseline predicates (Adolfo), or extending existing Scorecard formats.

---

## Clarifying questions from ORBIT WG feedback

The following clarifying questions require Steering Committee decisions
informed by Eddie's, Adolfo's, and Mike's feedback.

#### CQ-16: Allstar's role in OSPS conformance enforcement

**Stakeholders:** Stephen (can answer alone)

[Allstar](https://github.com/ossf/allstar) is a Scorecard sub-project that continuously monitors GitHub organizations and enforces Scorecard check results as policies (branch protection, binary artifacts, security policy, dangerous workflows). It already enforces a subset of controls aligned with OSPS Baseline.

With OSPS conformance output, Allstar could potentially enforce Baseline conformance at the organization level — e.g., opening issues or auto-remediating when a repository falls below Level 1 conformance. Should the proposal explicitly include Allstar as a Phase 1 consumer of OSPS output, or should that be deferred? And more broadly, should Allstar be considered part of the "enforcement" boundary that Scorecard itself does not cross, even though it is a Scorecard sub-project?

**Stephen's response:**


#### CQ-17: Mapping file location — Scorecard repo or shared?

**Stakeholders:** Stephen, Eddie Knight, OSPS Baseline maintainers

Eddie offers to host the Baseline-to-Scorecard mapping in the Baseline repository with Scorecard maintainers as CODEOWNERS (EK-1). The current proposal places it in the Scorecard repo.

Options:
1. **Scorecard repo** (`pkg/osps/mappings/`): Scorecard owns the mapping entirely. Mapping is coupled to Scorecard releases and probe changes.
2. **Baseline repo** (or shared location): Co-owned with ORBIT WG. Other tools can consume the same mapping. Scorecard maintainers retain CODEOWNERS authority over their portion.
3. **Both**: Scorecard maintains a local mapping for runtime use; a shared mapping in the Baseline repo serves as the cross-tool reference. Keep them in sync.

Which approach do you prefer?

_Note that this question is negated if consolidating check logic within `pvtr-github-repo-scanner`, because then the mappings are managed within the control catalog in Gemara format._

**Stephen's response:**

**Decision: Option 3 (both) — two-layer mapping model.**

- *Upstream* ([security-baseline](https://github.com/ossf/security-baseline) repo): Check-level relations — "OSPS-AC-03 relates to Scorecard's Branch-Protection check." Scorecard maintainers contribute via PR. The Baseline repo already has `guideline-mappings` referencing Scorecard in 9 controls (mapping to 7 checks). Scorecard can PR the missing ones.
- *Internal* (Scorecard repo): Probe-level mappings — "OSPS-AC-03.01 is evaluated by probes X + Y with logic Z." These depend on probe implementation details and must live in Scorecard.

**Language nuance** (per [security-baseline PR #476](https://github.com/ossf/security-baseline/pull/476)): Mappings were renamed to "relations" to guard against legal issues. Use "informs" / "provides evidence toward" rather than "satisfies" / "demonstrates compliance with."

Taking a dependency on `github.com/ossf/security-baseline` is acceptable — it is a shared OpenSSF project with useful connectors.

**Go module concern:** go.mod lives in cmd/ but module path is repo root. Import from cmd/pkg/ is unusual. Called out as potential concern, not blocking.

#### CQ-18: Output format — `--format=osps` vs. ecosystem formats

**Stakeholders:** Stephen, Spencer (OQ-4 constrains this), Eddie Knight, Adolfo García Veytia

Three reviewers (Eddie, Adolfo, Mike) independently flagged that no "OSPS output format" exists. Eddie suggests Gemara SDK formats (EK-2). Adolfo identifies two concrete in-toto predicate types (AP-2): the [Simple Verification Results (SVR)](https://github.com/in-toto/attestation/blob/main/spec/predicates/svr.md) predicate (merged) and the [Baseline Predicate](https://github.com/in-toto/attestation/pull/502) (proposed, not yet merged).

Options:
1. **Keep `--format=osps`**: Define a Scorecard-specific conformance output format. Risk: inventing a format that three reviewers have said doesn't belong.
2. **Use `--format=gemara`** (or similar): Integrate the Gemara SDK and output Gemara assessment results in JSON/YAML. Aligns with ORBIT ecosystem, creates a Gemara SDK dependency.
3. **Use in-toto predicates**: Output conformance results as in-toto attestations using SVR or the Baseline predicate. Aligns with in-toto ecosystem and Adolfo's guidance. The Baseline predicate is not yet merged.
4. **Extend existing formats**: Add conformance data to `--format=json` and `--format=sarif` outputs. No new format flag needed.
5. **Combination**: Use Gemara SDK for structured output + in-toto predicates for attestation output. These are not mutually exclusive.

Which approach do you prefer?

**Stephen's response:**

**Decision: Option 5 (combination) — the evidence model is the core deliverable; output formats are presentation layers.**

Phase 1 ships:
- **Enriched JSON** (Scorecard-native, no external dependency)
- **In-toto predicates** — SVR first; track [Baseline Predicate PR #502](https://github.com/in-toto/attestation/pull/502). Multiple predicate types supported simultaneously. Existing Scorecard predicate type (`scorecard.dev/result/v0.1`) preserved for backwards compatibility.
- **Gemara output** — dependency already transitive via `github.com/ossf/security-baseline` (gemara v0.7.0). The existing formatter pattern (`As<Format>()` methods) makes adding this straightforward.
- **OSCAL Assessment Results** — using [go-oscal](https://github.com/defenseunicorns/go-oscal). The security-baseline repo already exports OSCAL Catalog format (control definitions) via go-oscal v0.6.3. Scorecard would produce OSCAL Assessment Results (findings per control for a given repo) — a complementary OSCAL model. AMPEL has native OSCAL support.

There is no "OSPS output format" (confirming Eddie's, Adolfo's, and Mike's feedback). The `--format=osps` flag is replaced by the specific format flags above.

#### CQ-19: Architectural direction — build vs. integrate

**Stakeholders:** Stephen, Spencer, Eddie Knight, Steering Committee, at least 1 non-Steering maintainer — this is the gating decision; most other open questions depend on its outcome

This is the central decision. Eddie proposes consolidating Scorecard checks/probes into the Privateer plugin and having Scorecard execute the plugin (EK-5). The current proposal has Scorecard building its own conformance engine.

**Option A: Scorecard builds its own conformance engine** (current proposal)
- Scorecard adds a mapping file, conformance evaluation logic, and output format
- No code-level dependency on Privateer
- Scorecard controls its own release cadence and architecture
- Risk: duplicates evaluation logic, no technical relationship with Privateer (EK-3)

**Option B: Shared plugin model** (Eddie's alternative)
- Scorecard checks/probes are consolidated into the Privateer plugin
- Scorecard executes the plugin under the covers
- Bidirectional: Privateer users can also run Scorecard checks e.g., LFX Insights
- Gemara integration comes for free via the plugin
- Risk: Scorecard releases are coupled to plugin's release cadence; CODEOWNERS in the second repo must be meticulously managed to avoid surprises; multi-platform support (GitLab, Azure DevOps, local) will require maintenance of independent plugins with isolated data collection for each platform

**Option C: Hybrid**
- Scorecard maintains its own probe execution (its core competency)
- Scorecard exports its probe results in a format the Privateer plugin can consume (Gemara L5)
- The Privateer plugin consumes Scorecard output as supplementary evidence
- Control catalog is extracted and shared, but evaluation logic stays separate
- Users will choose between the Privateer plugin and Scorecard for Baseline evaluations
- No code-level coupling, but interoperable output

Which option do you prefer? What are your concerns about taking a dependency on the Privateer plugin codebase?

**Stephen's response:**

**Decision: Option C (hybrid), designed so that scaling back to Option A remains straightforward if needed.**

The architecture must ensure:
1. Scorecard owns all probe execution (non-negotiable core competency)
2. Scorecard owns its own conformance evaluation logic (mapping, PASS/FAIL, applicability engine all live in Scorecard)
3. Interoperability is purely at the output layer — Gemara, in-toto, SARIF, OSCAL are presentation formats, not architectural dependencies
4. Evaluation logic is self-contained — Scorecard can produce conformance results using its own probes and mappings, independent of external evaluation engines

**Dependency guidance:** Only adopt reasonably stable dependencies when needed. `github.com/ossf/security-baseline` is an acceptable data dependency for control definitions.

**Flexibility:** Under this structure, scaling back to a fully independent model (Option A) remains straightforward — deprioritize or drop specific output formatters without affecting the evaluation layer.


#### CQ-20: Catalog extraction — what does it mean concretely?

**Stakeholders:** Stephen, Eddie Knight, Steering Committee

Eddie is "hugely in favor" of extracting the Scorecard control catalog (EK-4) but the proposal lacks an implementation plan. Concretely, this could mean:

1. **Machine-readable probe definitions**: Export `probes/*/def.yml` as a versioned catalog (already exists in the repo, but not packaged for external consumption)
2. **Gemara L2 control definitions**: Map Scorecard probes to Gemara Layer 2 schema entries, making them available in the Gemara catalog
3. **Shared evaluation steps**: Extract Scorecard's probe logic into a reusable Go library or Privateer plugin steps that other tools can execute
4. **API-level catalog**: Expose probe definitions via the Scorecard API so tools can discover what Scorecard can evaluate

What level of extraction do you envision? Is option 2 (Gemara L2 integration) the right target, or should we start simpler?

**Stephen's response:**


#### CQ-21: Privateer code duplication — is it acceptable?

**Stakeholders:** Stephen, Spencer, Eddie Knight, Steering Committee — flows from CQ-19

Eddie points out that the current proposal would result in two codebases evaluating the same OSPS controls independently (EK-3). Even if the proposal says "don't duplicate Privateer," building a separate conformance engine effectively does that.

Is some duplication acceptable if it means Scorecard retains architectural independence? Or is avoiding duplication a hard constraint that should drive us toward the shared plugin model (CQ-19 Option B)?

**Stephen's response:**

Resolved by CQ-19 decision. Option C (hybrid) accepts that some evaluation overlap may occur. Scorecard SHOULD NOT duplicate evaluation that downstream tools handle (RFC 2119 SHOULD NOT, not MUST NOT). Scorecard retains architectural independence — interoperability is at the output layer, not the evaluation layer.


#### CQ-22: Attestation decomposition — identity vs. tooling

**Stakeholders:** Stephen, Spencer, Adolfo García Veytia, Eddie Knight

Adolfo points out that OQ-1 (attestation mechanism identity) conflates two questions (AP-3):

1. **Identity**: Who signs the attestation, and how are those identities trusted? (Sigstore/OIDC, personal keys, platform-native)
2. **Tooling**: What tool generates the attestation, and when? (Scorecard during scan, CI pipeline post-scan, manual maintainer process)

The answers will differ per project and per control. Should OQ-1 be decomposed into these two sub-questions, and should the design allow different identity/tooling combinations per control?

Adolfo has offered to discuss this in depth.

**Stephen's response:**

Acknowledged. OQ-1 should be decomposed into identity and tooling sub-questions as Adolfo suggests. The design should allow different identity/tooling combinations per control. Detailed resolution deferred to discussion with Adolfo and Spencer.


#### CQ-23: Mapping registry — where should the canonical mapping live?

**Stakeholders:** Stephen, Eddie Knight, Adolfo García Veytia, Baseline maintainers

Three perspectives have emerged on where Scorecard-to-Baseline mappings should live:

- **Eddie (EK-1)**: Host in the Baseline repo with Scorecard maintainers as CODEOWNERS
- **Adolfo (AP-1)**: Prefers a single registry in the Baseline itself, but notes the Baseline's mapping support was [demoted](https://github.com/ossf/security-baseline/pull/476)
- **Current proposal**: Host in Scorecard repo (`pkg/osps/mappings/`)

Additionally, AMPEL already maintains [independent Scorecard-to-Baseline mappings](https://github.com/carabiner-dev/policies/tree/main/groups/osps-baseline) in its policy library. An official upstream mapping would benefit both AMPEL and the wider ecosystem.

This extends CQ-17 with Adolfo's context about the demoted Baseline registry. Should the Scorecard mapping effort also advocate for restoring a shared registry in the Baseline spec?

**Stephen's response:**

Resolved by the two-layer mapping model (see CQ-17). Check-level relations are contributed upstream to `ossf/security-baseline` via PR, using the existing `guideline-mappings` structure. Probe-level mappings live in Scorecard. This approach works with the current state of the security-baseline repo without requiring restoration of the demoted mapping registry.


---

## Decision priority analysis

The open questions have dependencies between them. Answering them in the
wrong order will result in rework. The recommended sequence follows.

### Tier 1 — Gating decisions

| Question | Status | Resolution |
|----------|--------|------------|
| **CQ-19** | **RESOLVED** | Option C (hybrid), designed so that scaling back to Option A remains straightforward. Scorecard owns probe execution and evaluation; interoperability at output layer only. |
| **OQ-1** | **OPEN** | Attestation identity model. Spencer flagged as blocking. CQ-22 decomposes into identity vs. tooling. |

### Tier 2 — Downstream of CQ-19

| Question | Status | Resolution |
|----------|--------|------------|
| **CQ-18** | **RESOLVED** | Enriched JSON + in-toto predicates + Gemara + OSCAL Assessment Results. No "OSPS output format." |
| **CQ-17/CQ-23** | **RESOLVED** | Two-layer mapping model: check-level relations in security-baseline, probe-level mappings in Scorecard. |
| **CQ-22** | **PARTIALLY RESOLVED** | OQ-1 decomposed into identity vs. tooling sub-questions (per Adolfo). Detailed resolution deferred to discussion with Adolfo and Spencer. |
| **OQ-2** | **OPEN** | Enforcement detection scope. Affects Phase 3 scope. Needs Spencer + Stephen + Steering Committee. |

### Tier 3 — Important but non-blocking for Phase 1 start

| Question | Status | Notes |
|----------|--------|-------|
| **CQ-20** | **OPEN** | Catalog extraction scope. Flows from CQ-19 (now resolved). |
| **CQ-21** | **RESOLVED** | Some duplication acceptable. RFC 2119 SHOULD NOT, not MUST NOT. |
| **CQ-13** | **RESOLVED** | All consumers equal. RFC 2119 SHOULD NOT duplicate evaluation. |
| **CQ-11** | **OPEN** | Output stability guarantees. CQ-18 now resolved; this can proceed. |

### Tier 4 — Stephen can answer alone (any time)

| Question | Notes |
|----------|-------|
| **CQ-9** | NOT_OBSERVABLE controls — implementation detail, UNKNOWN-first principle already agreed. |
| **CQ-12** | Probe gap priority ordering — coverage doc already proposes an order. |
| **CQ-14** | Darnit vs. Minder delineation — ecosystem positioning Stephen can articulate. |
| **CQ-15** | Existing issues as Phase 1 work items — backlog triage. |
| **CQ-16** | Allstar's role — Scorecard sub-project under same Steering Committee. |

### Effectively resolved

| Question | Resolution |
|----------|-----------|
| **OQ-3** | Drop `scan_scope` from the schema (Spencer's feedback). |
| **OQ-4** | Evidence is probe-based only, not check-based (adopted). |
| **CQ-10** | Superseded by CQ-17 (two-layer mapping model). |

### Recommended next steps

1. **Resolve OQ-1/CQ-22** with Spencer, Adolfo, and the Steering Committee. Spencer flagged OQ-1 as blocking. Adolfo's decomposition (identity vs. tooling) clarifies what needs to be decided. Adolfo has offered to discuss.
2. **Resolve CQ-20** (catalog extraction scope) — now unblocked by CQ-19 resolution.
3. **Resolve CQ-11** (output stability guarantees) — now unblocked by CQ-18 resolution.
4. **Answer the Tier 4 questions** at any time — they are independent and don't block others.
5. **Begin Phase 1 implementation** — the gating architectural decisions (CQ-19, CQ-18, CQ-17) are resolved.
