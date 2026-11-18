# OSPS Baseline Coverage Analysis

Analysis of Scorecard's current probe and check coverage against the
[OSPS Baseline v2026.08.28](https://baseline.openssf.org/versions/2026-08-28),
the release Scorecard's [conformance mode](conformance.md) uses.

This is a living document. As probes are added or enhanced, update the
coverage status and evidence columns accordingly.

## Coverage legend

| Symbol | Meaning |
|--------|---------|
| COVERED | Scorecard has probes that fully satisfy this requirement |
| PARTIAL | Scorecard has probes that provide evidence but do not fully satisfy the requirement |
| GAP | No existing probe provides meaningful evidence for this requirement |
| NOT_OBSERVABLE | Requirement needs data Scorecard cannot access (e.g., org-level or repository admin settings) |

## Conformance mode

Coverage here means a probe collects evidence for the requirement. It does
not mean Scorecard reports a result for it. The experimental conformance
mode (`--catalog=osps-baseline`, see [conformance.md](conformance.md))
evaluates two requirements today:

- OSPS-AC-03.01 from `branchesAreProtected`, `requiresPRsToChangeCode`, and
  `branchProtectionAppliesToAdmins`
- OSPS-AC-03.02 from `blocksDeleteOnBranches`

Every other requirement is reported as not evaluated. The mappings that
conformance mode evaluates are defined in
`internal/conformance/ospsbaseline`; where this document and that package
differ, the package describes what Scorecard evaluates.

## Summary

Each requirement is listed at the lowest maturity level it applies to.

| Level | Requirements | COVERED | PARTIAL | GAP | NOT_OBSERVABLE |
|-------|--------------|---------|---------|-----|----------------|
| 1 | 24 | 4 | 7 | 11 | 2 |
| 2 | 19 | 2 | 3 | 12 | 2 |
| 3 | 21 | 0 | 6 | 15 | 0 |
| **Total** | **64** | **6** | **16** | **38** | **4** |

**Automated coverage rate (COVERED + PARTIAL): 34% (22 of 64)**

**Full coverage rate (COVERED only): 9% (6 of 64)**

## Level 1 requirements (24)

| OSPS ID | Requirement (short) | Status | Scorecard probes/checks providing evidence | Gap / notes |
|---------|---------------------|--------|---------------------------------------------|-------------|
| OSPS-AC-01.01 | MFA for sensitive resources | NOT_OBSERVABLE | None | Requires org admin API access; Scorecard tokens typically lack this. Must be UNKNOWN unless org-admin token is provided. |
| OSPS-AC-02.01 | Least-privilege defaults for new collaborators | NOT_OBSERVABLE | None | Requires org-level permission visibility. Must be UNKNOWN. |
| OSPS-AC-03.01 | Prevent direct commits to primary branch | COVERED | `branchesAreProtected`, `requiresPRsToChangeCode`, `branchProtectionAppliesToAdmins` | Branch-Protection check. Evaluated by conformance mode. Whether classic protection applies to admins is visible only with an admin token, and classic bypass lists aren't collected yet ([#5451](https://github.com/ossf/scorecard/issues/5451)). |
| OSPS-AC-03.02 | Prevent primary branch deletion | COVERED | `blocksDeleteOnBranches` | Branch-Protection check. Evaluated by conformance mode: a rule that blocks deletion is PASS; a rule that allows it is UNKNOWN, because Scorecard can't see whether the platform asks for confirmation. |
| OSPS-BR-01.01 | Sanitize untrusted CI/CD input | PARTIAL | `hasDangerousWorkflowScriptInjection` | Dangerous-Workflow check detects script injection patterns. Does not cover all sanitization cases (e.g., non-shell contexts). |
| OSPS-BR-01.03 | Untrusted code snapshots cannot access privileged credentials | PARTIAL | `hasDangerousWorkflowUntrustedCheckout` | Dangerous-Workflow check detects untrusted checkouts with access to secrets. Does not cover all credential isolation scenarios. |
| OSPS-BR-03.01 | Official channel URIs use encrypted transport | GAP | None | Requires a source of truth for official URIs, such as Security Insights. No probe exists. |
| OSPS-BR-03.02 | Distribution URIs use authenticated channels | GAP | None | Same as BR-03.01. Requires declared distribution channels. |
| OSPS-BR-07.01 | Prevent unintentional storage of secrets in VCS | GAP | None | No secrets detection probe exists today. Could consume platform signals (e.g., GitHub secret scanning API). Applies at maturity-1 only in v2026.08.28. |
| OSPS-DO-01.01 | User guides for released software | GAP | None | No documentation presence probe. Would need file/path heuristics. |
| OSPS-DO-02.01 | Defect reporting guide | GAP | None | No issue template / bug report documentation probe. |
| OSPS-GV-02.01 | Public discussion mechanism | GAP | None | No probe. Whether issues or discussions are enabled could be read from the repository GraphQL query Scorecard already makes. |
| OSPS-GV-03.01 | Documented contribution process | GAP | None | No CONTRIBUTING file presence/content probe. v2026.08.28 also accepts a clear statement that public contributions are not accepted. |
| OSPS-LE-02.01 | OSI/FSF license for source code | COVERED | `hasFSFOrOSIApprovedLicense` | License check. Direct mapping. |
| OSPS-LE-02.02 | OSI/FSF license for released assets | PARTIAL | `hasFSFOrOSIApprovedLicense` | License check verifies repo license, but does not verify license is shipped with release artifacts. |
| OSPS-LE-03.01 | License file in repository | COVERED | `hasLicenseFile` | License check. Direct mapping, including the `LICENSES/` directory that v2026.08.28 added. |
| OSPS-LE-03.02 | License included with released assets | PARTIAL | `hasLicenseFile` | Detects license in repo, not in release artifacts. Needs release asset inspection. |
| OSPS-QA-01.01 | Repo publicly readable at static URL | GAP | None | Not implicit: Scorecard can scan private repositories with a token, so visibility has to be read. The repository GraphQL query could add the `visibility` field. |
| OSPS-QA-01.02 | Public commit history with authorship and timestamps | GAP | None | The VCS records authorship and timestamps; whether that record is public depends on visibility, as for QA-01.01. |
| OSPS-QA-02.01 | Direct dependency list present | GAP | None | No probe. `pinsDependencies` reads workflows, Dockerfiles, and scripts, not language manifests. A probe could use the file list Scorecard already fetches. |
| OSPS-QA-04.01 | Docs list subprojects | GAP | None | Requires Security Insights or similar metadata. No probe exists. |
| OSPS-QA-05.01 | No generated executable artifacts in VCS | PARTIAL | `hasBinaryArtifacts`, `hasUnverifiedBinaryArtifacts` | Binary-Artifacts check. Detects binary files but may not distinguish "generated executables" from other binaries. |
| OSPS-QA-05.02 | No unreviewable binary artifacts in VCS | PARTIAL | `hasUnverifiedBinaryArtifacts` | Detects unverified binaries. "Unreviewable" vs "reviewable" classification is not yet granular enough. |
| OSPS-VM-02.01 | Security contacts documented | PARTIAL | `securityPolicyPresent`, `securityPolicyContainsLinks` | Security-Policy check detects SECURITY.md presence and links, but does not verify actual contact methods (email, form, etc.). Applies at maturity-1 only in v2026.08.28. |

## Level 2 requirements (19)

| OSPS ID | Requirement (short) | Status | Scorecard probes/checks providing evidence | Gap / notes |
|---------|---------------------|--------|---------------------------------------------|-------------|
| OSPS-AC-04.01 | Default lowest CI/CD permissions | NOT_OBSERVABLE | None | The requirement is about the repository's default token permissions, which only a repository admin can read through the API. Token-Permissions probes read the permissions workflows declare, which is a different question. |
| OSPS-BR-02.01 | Releases have unique version identifier | GAP | None | No release versioning probe. Release tags are already fetched for Signed-Releases. |
| OSPS-BR-04.01 | Releases have descriptive changelog | GAP | None | No changelog/release notes detection probe. |
| OSPS-BR-05.01 | Standardized tooling for dependency ingestion | GAP | None | No probe detects whether standard package managers are used in CI. |
| OSPS-BR-06.01 | Releases signed or accounted for in signed manifest | PARTIAL | `releasesAreSigned`, `releasesHaveProvenance`, `releasesHaveVerifiedProvenance`, `releasesAreImmutable` | Signed-Releases check covers signatures and provenance, and counts GitHub immutable releases as signed. Does not yet check for "signed manifest including hashes" as an alternative. |
| OSPS-DO-06.01 | Docs describe dependency selection/tracking | GAP | None | Documentation requirement. No probe exists. |
| OSPS-DO-07.01 | Build instructions in documentation | GAP | None | Documentation requirement. No probe exists. |
| OSPS-GV-01.01 | Docs list members with sensitive access | NOT_OBSERVABLE | None | Requires org-level data or attestation. Not automatable via Scorecard. |
| OSPS-GV-01.02 | Docs list roles and responsibilities | GAP | None | Documentation requirement. May require attestation. |
| OSPS-GV-03.02 | Contributor guide with acceptability requirements | GAP | None | No probe today; a contributing-file probe for GV-03.01 could be extended to check content structure. |
| OSPS-LE-01.01 | Legal authorization per commit (DCO/CLA) | GAP | None | No DCO/CLA detection probe. Would check for Signed-off-by trailers or CLA bot enforcement. |
| OSPS-QA-03.01 | Status checks pass or bypassed before merge | COVERED | `runsStatusChecksBeforeMerging` | Branch-Protection check. Direct mapping. |
| OSPS-QA-06.01 | Automated tests run prior to acceptance | COVERED | `testsRunInCI` | CI-Tests check. Maps directly. |
| OSPS-SA-01.01 | Design docs with actions/actors | GAP | None | Documentation/assessment requirement. Requires attestation. |
| OSPS-SA-02.01 | Docs describe external interfaces | GAP | None | Documentation requirement. Requires attestation. |
| OSPS-SA-03.01 | Security assessment performed | GAP | None | Process requirement. Requires attestation with evidence link. |
| OSPS-VM-01.01 | CVD policy with response timeframe | PARTIAL | `securityPolicyContainsVulnerabilityDisclosure`, `securityPolicyContainsText` | Security-Policy check detects disclosure language. Does not verify explicit timeframe commitment. |
| OSPS-VM-03.01 | Private vulnerability reporting method | PARTIAL | `securityPolicyContainsLinks` | Detects links in SECURITY.md. Does not verify private reporting is actually enabled (e.g., GitHub private vulnerability reporting). |
| OSPS-VM-04.01 | Publicly publish vulnerability data | GAP | None | No probe checks for GitHub Security Advisories, OSV entries, or CVE publication. |

## Level 3 requirements (21)

| OSPS ID | Requirement (short) | Status | Scorecard probes/checks providing evidence | Gap / notes |
|---------|---------------------|--------|---------------------------------------------|-------------|
| OSPS-AC-04.02 | Job-level least privilege in CI/CD | PARTIAL | `jobLevelPermissions`, `topLevelPermissions`, `hasNoGitHubWorkflowPermissionUnknown` | Token-Permissions check evaluates workflow permissions. "Minimum necessary" is hard to assess without understanding job intent. |
| OSPS-BR-01.04 | Sanitize trusted collaborator CI/CD input | PARTIAL | `hasDangerousWorkflowScriptInjection` | Dangerous-Workflow check partially covers this, but focuses on untrusted input, not trusted collaborator input. |
| OSPS-BR-02.02 | Release assets tied to release identifier | GAP | None | No release asset naming/association probe. Asset names are already fetched for Signed-Releases. |
| OSPS-BR-07.02 | Secrets management policy | GAP | None | Documentation/policy requirement. Requires attestation. |
| OSPS-DO-03.01 | Instructions to verify release integrity/authenticity | GAP | None | Documentation requirement. Could partially automate by checking for verification docs alongside signed releases. |
| OSPS-DO-03.02 | Instructions to verify release author identity | GAP | None | Documentation requirement. |
| OSPS-DO-04.01 | Support scope/duration per release | GAP | None | Documentation requirement. |
| OSPS-DO-05.01 | EOL security update statement | GAP | None | Documentation requirement. |
| OSPS-GV-04.01 | Policy to review collaborators before escalated perms | GAP | None | Governance policy. Requires attestation. |
| OSPS-QA-02.02 | SBOM shipped with compiled release assets | PARTIAL | `hasReleaseSBOM`, `hasSBOM` | SBOM probes exist but may not specifically verify compiled-asset association. |
| OSPS-QA-04.02 | Subprojects enforce >= primary requirements | GAP | None | Requires multi-repo scanning and cross-repo comparison. |
| OSPS-QA-06.02 | Docs describe when/how tests run | GAP | None | Documentation requirement. |
| OSPS-QA-06.03 | Policy requiring tests for major changes | GAP | None | Documentation/policy requirement. |
| OSPS-QA-07.01 | Non-author approval before merging | PARTIAL | `requiresApproversForPullRequests`, `codeApproved`, `codeReviewOneReviewers` | Branch-Protection and Code-Review probes cover this. GitHub doesn't count an author's own approval, but an app's approval can count, so a "human" approval isn't shown. |
| OSPS-SA-03.02 | Threat model and attack surface analysis performed | GAP | None | Process requirement. Requires attestation with evidence link. |
| OSPS-VM-04.02 | VEX for non-affecting vulnerabilities | GAP | None | No VEX detection probe. |
| OSPS-VM-05.01 | SCA remediation threshold policy | GAP | None | Policy requirement. |
| OSPS-VM-05.02 | SCA violations addressed pre-release | GAP | None | Policy + enforcement requirement. |
| OSPS-VM-05.03 | Automated SCA eval + block violations | PARTIAL | `hasOSVVulnerabilities` | Vulnerabilities check detects known vulns. Does not verify gating/blocking enforcement. |
| OSPS-VM-06.01 | SAST remediation threshold policy | GAP | None | Policy requirement. |
| OSPS-VM-06.02 | Automated SAST eval + block violations | PARTIAL | `sastToolConfigured`, `sastToolRunsOnAllCommits` | SAST check detects tool presence and execution. Does not verify gating/blocking enforcement. |

## Phase 1 priorities (Level 1 gap closure)

The following Level 1 gaps should be addressed first, ordered by implementation feasibility:

1. **OSPS-GV-03.01** (contribution process): Add probe for CONTRIBUTING file presence, or a statement that contributions are not accepted
2. **OSPS-GV-02.01** (public discussion): Add probe for issues/discussions enabled
3. **OSPS-QA-01.01 / QA-01.02** (public repository): Read repository visibility from the same GraphQL query as GV-02.01
4. **OSPS-QA-02.01** (dependency list): Add probe for language dependency manifests in the file list Scorecard already fetches
5. **OSPS-DO-02.01** (defect reporting): Add probe for issue templates or bug report docs
6. **OSPS-DO-01.01** (user guides): Add probe for documentation presence heuristics
7. **OSPS-BR-07.01** (secrets in VCS): Consume GitHub secret scanning API or add detection heuristics
8. **OSPS-BR-03.01 / BR-03.02** (encrypted transport): Requires declared URIs, such as Security Insights
9. **OSPS-QA-04.01** (subproject listing): Requires Security Insights or equivalent metadata

## Probes not mapped to any OSPS requirement

The following probes exist in Scorecard but do not directly map to any OSPS Baseline requirement:

| Probe | Check | Notes |
|-------|-------|-------|
| `archived` | Maintained | Project archival status |
| `hasRecentCommits` | Maintained | Activity signal |
| `issueActivityByProjectMember` | Maintained | Activity signal |
| `createdRecently` | Maintained | Age signal |
| `contributorsFromOrgOrCompany` | Contributors | Diversity signal |
| `dependencyUpdateToolConfigured` | Dependency-Update-Tool | Best practice, not a Baseline requirement |
| `fuzzed` | Fuzzing | Testing best practice, not a Baseline requirement |
| `hasOpenSSFBadge` | CII-Best-Practices | Meta-badge, not a Baseline requirement |
| `packagedWithAutomatedWorkflow` | Packaging | Distribution best practice |
| `webhooksUseSecrets` | Webhooks | Security practice, not a Baseline requirement |
| `hasPermissiveLicense` | (uncategorized) | License type classification |
| `unsafeblock` | (independent) | Language-specific safety |
| `dismissesStaleReviews` | Branch-Protection | Review hygiene beyond Baseline scope |
| `requiresCodeOwnersReview` | Branch-Protection | CODEOWNERS enforcement beyond Baseline scope |
| `requiresLastPushApproval` | Branch-Protection | Push approval beyond Baseline scope |
| `requiresUpToDateBranches` | Branch-Protection | Branch freshness beyond Baseline scope |
| `blocksForcePushOnBranches` | Branch-Protection | Force-push protection; related to AC-03 but not explicitly required |

These probes remain valuable for Scorecard's existing scoring model and may become relevant for future Baseline versions.

## Existing issues and PRs relevant to gap closure

The following issues and PRs in the Scorecard repository relate to
closing OSPS Baseline coverage gaps. Several were closed as not planned in
November 2026; they stay listed for their discussion, and the gaps they
describe remain.

### Conformance mode work

- [#5449](https://github.com/ossf/scorecard/pull/5449) — Add experimental OSPS Baseline conformance mode (merged)
- [#5452](https://github.com/ossf/scorecard/issues/5452) — Proposal: Graduate or remove experimental OSPS Baseline conformance mode
- [#5451](https://github.com/ossf/scorecard/issues/5451) — Proposal: Collect branch protection bypass allowances for required pull requests
- [#5450](https://github.com/ossf/scorecard/issues/5450) — Issue: branchesAreProtected reports False when protection is unknown

### Security Insights ingestion

- [#2305](https://github.com/ossf/scorecard/issues/2305) — Support for SECURITY INSIGHTS (closed as not planned)
- [#2479](https://github.com/ossf/scorecard/issues/2479) — SECURITY-INSIGHTS.yml implementation (closed as not planned)

Declared project metadata is needed for OSPS-BR-03.01, BR-03.02, QA-04.01,
and other requirements.

### Secrets detection (OSPS-BR-07.01)

- [#30](https://github.com/ossf/scorecard/issues/30) — New check: code is scanning for secrets (closed as not planned)

### SBOM (OSPS-QA-02.02)

- [#1476](https://github.com/ossf/scorecard/issues/1476) — Feature: Detect if SBOMs generated (closed as not planned)
- [#2605](https://github.com/ossf/scorecard/issues/2605) — Add support for SBOM analyzing at Binary-Artifacts stage (closed as not planned)

The SBOM check and probes (`hasSBOM`, `hasReleaseSBOM`) already exist but
may need enhancement for compiled release asset association.

### Changelog / release notes (OSPS-BR-04.01)

- [#4824](https://github.com/ossf/scorecard/issues/4824) — Feature: New Check: Check if the project has and maintains a CHANGELOG

Direct match for Phase 2 deliverable.

### Private vulnerability reporting (OSPS-VM-03.01)

- [#2465](https://github.com/ossf/scorecard/issues/2465) — Factor whether or not private vulnerability reporting is enabled into the scorecard

Direct match. GitHub's private vulnerability reporting API could provide
platform-level evidence.

### Vulnerability disclosure improvements (OSPS-VM-01.01, VM-04.01)

- [#4192](https://github.com/ossf/scorecard/issues/4192) — Test for security policy in other places than SECURITY.md
- [#4789](https://github.com/ossf/scorecard/issues/4789) — Rethinking vulnerability check scoring logic
- [#1371](https://github.com/ossf/scorecard/issues/1371) — Feature: add check for vulnerability alerts (closed as not planned)

### Signed releases and provenance (OSPS-BR-06.01)

- [#4823](https://github.com/ossf/scorecard/issues/4823) — Feature: pass Signed-Releases with GitHub immutable release process (completed)
- [#4080](https://github.com/ossf/scorecard/issues/4080) — Use GitHub attestations to check for signed releases
- [#2684](https://github.com/ossf/scorecard/issues/2684) — Rework: Signed-Releases: Separate score calculation of provenance and signatures
- [#1417](https://github.com/ossf/scorecard/issues/1417) — Feature: add support for keyless signed release (closed as not planned)

### Threat model / security assessment (OSPS-SA-01.01, SA-03.01, SA-03.02)

- [#2142](https://github.com/ossf/scorecard/issues/2142) — Feature: Assess presence and maintenance of a threat model (closed as not planned)

### Release scoring (OSPS-BR-02.01, BR-02.02)

- [#1985](https://github.com/ossf/scorecard/issues/1985) — Feature: Scoring for individual releases

### Minder integration

- [#4723](https://github.com/ossf/scorecard/pull/4723) — Initial draft of using Minder rules in Scorecard (CLOSED)

Draft PR that attempted to run Minder Rego rules within Scorecard,
including OSPS-QA-05.01 and QA-03.01. Closed due to inactivity but
demonstrates interest in deeper Minder/Scorecard integration.

## Notes

- OSPS Baseline v2026.08.28 has 65 requirement IDs: 64 active, plus
  OSPS-BR-01.02, retired in
  [ossf/security-baseline#443](https://github.com/ossf/security-baseline/pull/443).
  This analysis covers the 64 active requirements and supersedes the earlier
  analysis against v2026.02.19.
- From v2026.02.19 to v2026.08.28, no requirement was added, removed, or
  moved to another level. The "While active" precondition was dropped from
  every requirement that had it. OSPS-GV-03.01 now also accepts a statement
  that public contributions are not accepted, and OSPS-LE-03.01 accepts a
  `LICENSES/` directory.
- Unreleased changes on the security-baseline `main` branch apply
  OSPS-VM-02.01 at all three levels and reword OSPS-AC-03.02. Update this
  document when they ship in a release.
- Requirements marked NOT_OBSERVABLE cannot produce PASS or FAIL without
  elevated permissions. The conformance engine must return UNKNOWN with an
  explanation.
- Many Level 2 and Level 3 requirements are documentation or policy
  requirements that need attestation rather than automated detection. The
  attestation mechanism (OQ-1 in the proposal) is critical for these.
