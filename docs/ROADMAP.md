# OpenSSF Scorecard Roadmap

## 2026

### Theme: OSPS Baseline Conformance

Scorecard's primary initiative for 2026 is adding
[OSPS Baseline](https://baseline.openssf.org/) conformance evaluation,
enabling Scorecard to answer the question: _does this project meet the
security requirements defined by the OSPS Baseline at a given maturity level?_

This is a new product surface alongside Scorecard's existing 0-10 scoring
model. Existing checks, probes, and scores are unchanged. The conformance
layer consumes existing Scorecard signals and adds a per-control
PASS/FAIL/UNKNOWN/NOT_APPLICABLE/ATTESTED output aligned with the
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

- OSPS output format (`--format=osps`)
- Versioned mapping file (YAML) mapping OSPS controls to Scorecard probes
- Applicability engine detecting preconditions (e.g., "has made a release")
- New probes for Level 1 gaps:
  - Governance and documentation presence (OSPS-GV-02.01, GV-03.01,
    DO-01.01, DO-02.01)
  - Dependency manifest presence (OSPS-QA-02.01)
  - Security policy deepening (OSPS-VM-02.01, VM-03.01)
  - Secrets detection (OSPS-BR-07.01) — consuming platform signals where
    available
- Security Insights ingestion (OSPS-BR-03.01, BR-03.02, QA-04.01)
- CI gating via `--fail-on=fail`
- Scorecard control catalog extraction plan

#### Phase 2: Release integrity and Level 2 core

**Outcome:** Scorecard evaluates release-related OSPS controls, covering the
core of Level 2 and becoming useful for downstream due diligence workflows.

Deliverables:

- Release asset inspection layer
- Signed manifest support (OSPS-BR-06.01)
- Release notes and changelog detection (OSPS-BR-04.01)
- Attestation mechanism for non-automatable controls
- Evidence bundle output (OSPS result JSON + in-toto statement)
- Gemara SDK integration for interoperable output

#### Phase 3: Enforcement detection, Level 3, and multi-repo

**Outcome:** Scorecard covers Level 3 controls including enforcement detection
and project-level aggregation.

Deliverables:

- SCA policy and enforcement detection (OSPS-VM-05.*)
- SAST policy and enforcement detection (OSPS-VM-06.*)
- Multi-repo project-level conformance aggregation (OSPS-QA-04.02)
- Attestation integration GA

### Ecosystem alignment

Scorecard operates within the ORBIT WG ecosystem as a measurement and
evidence tool. [Allstar](https://github.com/ossf/allstar), a Scorecard
sub-project, continuously monitors GitHub organizations and enforces
Scorecard check results as policies. OSPS conformance output could enable
Allstar to enforce Baseline conformance at the organization level.

Scorecard does not duplicate:

- **[Minder](https://github.com/mindersec/minder)** — Policy enforcement and remediation platform (OpenSSF Sandbox, ORBIT WG)
- **[Privateer plugin for GitHub repositories](https://github.com/ossf/pvtr-github-repo-scanner)** — Baseline evaluation powered by Gemara and Security Insights
- **[Darnit](https://github.com/kusari-oss/darnit)** — Compliance audit and remediation
- **[AMPEL](https://github.com/carabiner-dev/ampel)** — Attestation-based policy enforcement

Scorecard's role is to produce deep, probe-based conformance evidence that
these tools and downstream consumers can use. Minder already consumes
Scorecard findings to enforce security policies across repositories.

### Design principles

1. **UNKNOWN-first honesty.** If Scorecard cannot observe a control, the
   status is UNKNOWN with an explanation — never a false PASS or FAIL.
2. **Probes are the evidence unit.** OSPS evidence references probes and
   their findings, not check-level scores.
3. **Additive, not breaking.** Existing checks, probes, scores, and output
   formats do not change behavior.
4. **Data-driven mapping.** The mapping between OSPS controls and Scorecard
   probes is a versioned YAML file, not hard-coded logic.
5. **Degraded-but-useful without Security Insights.** Projects without a
   `security-insights.yml` still get a meaningful (if incomplete) report.

### Open questions

The following design questions are under active discussion among maintainers:

- **Attestation identity model** — How non-automatable controls are attested
  (repo-local metadata vs. signed attestations via Sigstore/OIDC)
- **Enforcement detection scope** — How Scorecard detects enforcement
  mechanisms without being an enforcement tool itself
- **Evidence format** — Ensuring output compatibility with Gemara Layer 4
  assessment schemas

### How to contribute

See the [proposal](../openspec/changes/osps-baseline-conformance/proposal.md)
and [spec](../openspec/changes/osps-baseline-conformance/specs/osps-conformance/spec.md)
for detailed requirements. Discussion and feedback are welcome via GitHub
issues and the Scorecard community meetings.
