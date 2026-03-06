# OpenSSF Scorecard Roadmap

## 2026

### Theme: Scorecard v6 — Open Source Security Evidence Engine

**Mission:** Scorecard produces trusted, structured security evidence for the
open source ecosystem.

Scorecard v6 evolves Scorecard from a scoring tool to an evidence engine. The
primary initiative for 2026 is adding
[OSPS Baseline](https://baseline.openssf.org/) conformance evaluation as the
first use case that proves this architecture. Scorecard accepts diverse inputs
about a project's security practices, normalizes them through probe-based
analysis, and packages the resulting evidence in interoperable formats for
downstream tools to act on.

Check scores (0-10) and conformance labels (PASS/FAIL/UNKNOWN) are parallel
evaluation layers over the same probe evidence, produced in a single run.
Existing checks, probes, and scores are unchanged — v6 is additive. The
conformance layer is a new product surface aligned with the
[ORBIT WG](https://github.com/ossf/wg-orbit) ecosystem.

**Target Baseline version:** [v2026.02.19](https://baseline.openssf.org/versions/2026-02-19)

**Current coverage:** See [docs/osps-baseline-coverage.md](osps-baseline-coverage.md)
for a control-by-control analysis.

### Phased delivery

Phases are ordered by outcome. Maintainer bandwidth dictates delivery timing.

#### Phase 1: Conformance foundation and Level 1 coverage

**Outcome:** Scorecard produces a useful OSPS Baseline Level 1 conformance
report for any public GitHub repository, available across CLI, Action, and
API surfaces.

Deliverables:

- Evidence model and output formats:
  - Enriched JSON (Scorecard-native)
  - In-toto predicates ([SVR](https://github.com/in-toto/attestation/blob/main/spec/predicates/svr.md);
    track [Baseline Predicate PR #502](https://github.com/in-toto/attestation/pull/502))
  - Gemara output (via [security-baseline](https://github.com/ossf/security-baseline)
    dependency)
  - OSCAL Assessment Results (via
    [go-oscal](https://github.com/defenseunicorns/go-oscal))
- Two-layer mapping model for OSPS Baseline v2026.02.19:
  - Check-level relations contributed upstream to security-baseline
  - Probe-level mappings maintained in Scorecard
- Applicability engine detecting preconditions (e.g., "has made a release")
- Map existing probes to OSPS controls where coverage exists today
- New probes for Level 1 gaps:
  - Governance and documentation presence
  - Dependency manifest presence
  - Security policy deepening
  - Secrets detection — consuming platform signals where available
- Metadata ingestion layer — Security Insights as first supported source;
  architecture supports additional metadata sources
- Scorecard control catalog extraction plan

#### Phase 2: Release integrity and Level 2 core

**Outcome:** Scorecard evaluates release-related OSPS controls, covering the
core of Level 2 and becoming useful for downstream due diligence workflows.

Deliverables:

- Release asset inspection layer
- Signed manifest support
- Release notes and changelog detection
- Attestation mechanism for non-automatable controls
- Evidence bundle output (conformance results + in-toto statement)
- Additional metadata sources for the ingestion layer

#### Phase 3: Enforcement detection, Level 3, and multi-repo

**Outcome:** Scorecard covers Level 3 controls including enforcement detection
and project-level aggregation.

Deliverables:

- SCA policy and enforcement detection
- SAST policy and enforcement detection
- Multi-repo project-level conformance aggregation
- Attestation integration GA

### Ecosystem alignment

Scorecard operates within the ORBIT WG ecosystem as an evidence engine. All
downstream tools consume Scorecard evidence on equal terms through published
output formats.

[Allstar](https://github.com/ossf/allstar), a Scorecard sub-project,
continuously monitors GitHub organizations and enforces Scorecard check
results as policies. OSPS conformance output could enable Allstar to enforce
Baseline conformance at the organization level.

Scorecard SHOULD NOT (per [RFC 2119](https://www.rfc-editor.org/rfc/rfc2119))
duplicate evaluation that downstream tools handle:

- **[Privateer](https://github.com/ossf/pvtr-github-repo-scanner)** — Baseline evaluation powered by Gemara and Security Insights
- **[Minder](https://github.com/mindersec/minder)** — Policy enforcement and remediation platform (OpenSSF Sandbox, ORBIT WG)
- **[AMPEL](https://github.com/carabiner-dev/ampel)** — Attestation-based policy enforcement; already consumes Scorecard probe results via [policy library](https://github.com/carabiner-dev/policies/tree/main/scorecard)
- **[Darnit](https://github.com/kusari-oss/darnit)** — Compliance audit and remediation

Scorecard's role is to produce deep, probe-based security evidence that these
tools and downstream consumers can use through interoperable output formats
(JSON, in-toto, Gemara, SARIF, OSCAL).

### Design principles

1. **Evidence is the product.** Scorecard's core output is structured,
   normalized probe findings. Check scores and conformance labels are parallel
   evaluation layers over the same evidence.
2. **Probes normalize diversity.** Each probe understands multiple ways a
   control outcome can be satisfied.
3. **UNKNOWN-first honesty.** If Scorecard cannot observe a control, the
   status is UNKNOWN with an explanation — never a false PASS or FAIL.
4. **All consumers are equal.** Downstream tools consume Scorecard evidence
   through published output formats.
5. **No metadata monopolies.** Probes may evaluate multiple sources for the
   same data. No single metadata file is required for meaningful results,
   though they may enrich results.
6. **Formats are presentation.** Output formats (JSON, in-toto, Gemara,
   SARIF, OSCAL) are views over the evidence model. No single format is
   privileged.

### Open questions

The following design questions are under active discussion among maintainers:

- **Attestation identity model** — How non-automatable controls are attested
  (repo-local metadata vs. signed attestations via Sigstore/OIDC). Decomposed
  into identity (who signs) and tooling (what generates, when) sub-questions.
- **Enforcement detection scope** — How Scorecard detects enforcement
  mechanisms without being an enforcement tool itself

### How to contribute

See the [proposal](../openspec/changes/osps-baseline-conformance/proposal.md)
for detailed requirements and open questions. Discussion and feedback are
welcome via GitHub issues and the Scorecard community meetings.
