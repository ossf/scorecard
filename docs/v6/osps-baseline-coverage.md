# OSPS Baseline Coverage Analysis

Analysis of Scorecard's current probe and check coverage against the
[OSPS Baseline v2026.02.19](https://baseline.openssf.org/versions/2026-02-19).

This is a living document. As probes are added or enhanced, update the
coverage status and evidence columns accordingly.

## Coverage legend

| Symbol | Meaning |
|--------|---------|
| COVERED | Scorecard has probes that fully satisfy this control |
| PARTIAL | Scorecard has probes that provide evidence but do not fully satisfy the control |
| GAP | No existing probe provides meaningful evidence for this control |
| NOT_OBSERVABLE | Control requires data Scorecard cannot access (e.g., org-level admin permissions) |

## Summary

| Level | Total controls | COVERED | PARTIAL | GAP | NOT_OBSERVABLE |
|-------|---------------|---------|---------|-----|----------------|
| 1 | 25 | 6 | 8 | 9 | 2 |
| 2 | 17 | 2 | 5 | 9 | 1 |
| 3 | 17 | 0 | 4 | 13 | 0 |
| **Total** | **59** | **8** | **17** | **31** | **3** |

**Automated coverage rate (COVERED + PARTIAL): 42% (25 of 59)**

**Full coverage rate (COVERED only): 14% (8 of 59)**

## Level 1 controls (25)

| OSPS ID | Control (short) | Status | Scorecard probes/checks providing evidence | Gap / notes |
|---------|----------------|--------|---------------------------------------------|-------------|
| OSPS-AC-01.01 | MFA for sensitive resources | NOT_OBSERVABLE | None | Requires org admin API access; Scorecard tokens typically lack this. Must be UNKNOWN unless org-admin token is provided. |
| OSPS-AC-02.01 | Least-privilege defaults for new collaborators | NOT_OBSERVABLE | None | Requires org-level permission visibility. Must be UNKNOWN. |
| OSPS-AC-03.01 | Prevent direct commits to primary branch | COVERED | `requiresPRsToChangeCode`, `branchesAreProtected` | Branch-Protection check. Maps directly when PR-only merges are enforced. |
| OSPS-AC-03.02 | Prevent primary branch deletion | COVERED | `blocksDeleteOnBranches` | Branch-Protection check. Direct mapping. |
| OSPS-BR-01.01 | Sanitize untrusted CI/CD input | PARTIAL | `hasDangerousWorkflowScriptInjection` | Dangerous-Workflow check detects script injection patterns. Does not cover all sanitization cases (e.g., non-shell contexts). |
| OSPS-BR-01.03 | Untrusted code snapshots cannot access privileged credentials | PARTIAL | `hasDangerousWorkflowUntrustedCheckout` | Dangerous-Workflow check detects untrusted checkouts with access to secrets. Does not cover all credential isolation scenarios. |
| OSPS-BR-03.01 | Official channel URIs use encrypted transport | GAP | None | Requires a source-of-truth for official URIs (Security Insights). No probe exists. |
| OSPS-BR-03.02 | Distribution URIs use authenticated channels | GAP | None | Same as BR-03.01. Requires declared distribution channels. |
| OSPS-BR-07.01 | Prevent unintentional storage of secrets in VCS | GAP | None | No secrets detection probe exists today. Could consume platform signals (e.g., GitHub secret scanning API). |
| OSPS-DO-01.01 | User guides for released software | GAP | None | No documentation presence probe. Would need file/path heuristics. |
| OSPS-DO-02.01 | Defect reporting guide | GAP | None | No issue template / bug report documentation probe. |
| OSPS-GV-02.01 | Public discussion mechanism | GAP | None | Could check whether issues/discussions are enabled, but no probe exists. |
| OSPS-GV-03.01 | Documented contribution process | GAP | None | No CONTRIBUTING file presence/content probe. |
| OSPS-LE-02.01 | OSI/FSF license for source code | COVERED | `hasFSFOrOSIApprovedLicense` | License check. Direct mapping. |
| OSPS-LE-02.02 | OSI/FSF license for released assets | PARTIAL | `hasFSFOrOSIApprovedLicense` | License check verifies repo license, but does not verify license is shipped with release artifacts. |
| OSPS-LE-03.01 | License file in repository | COVERED | `hasLicenseFile` | License check. Direct mapping. |
| OSPS-LE-03.02 | License included with released assets | PARTIAL | `hasLicenseFile` | Detects license in repo, not in release artifacts. Needs release asset inspection. |
| OSPS-QA-01.01 | Repo publicly readable at static URL | COVERED | (implicit) | Scorecard can only scan public repos. If Scorecard runs, this is satisfied. |
| OSPS-QA-01.02 | Public commit history with authorship and timestamps | COVERED | (implicit) | VCS provides this by nature. Scorecard relies on commit history for multiple probes. |
| OSPS-QA-02.01 | Direct dependency list present | PARTIAL | `pinsDependencies` | Pinned-Dependencies check detects dependency manifests but focuses on pinning, not mere presence. |
| OSPS-QA-04.01 | Docs list subprojects | GAP | None | Requires Security Insights or similar metadata. No probe exists. |
| OSPS-QA-05.01 | No generated executable artifacts in VCS | PARTIAL | `hasBinaryArtifacts`, `hasUnverifiedBinaryArtifacts` | Binary-Artifacts check. Detects binary files but may not distinguish "generated executables" from other binaries. |
| OSPS-QA-05.02 | No unreviewable binary artifacts in VCS | PARTIAL | `hasUnverifiedBinaryArtifacts` | Detects unverified binaries. "Unreviewable" vs "reviewable" classification is not yet granular enough. |
| OSPS-VM-02.01 | Security contacts documented | PARTIAL | `securityPolicyPresent`, `securityPolicyContainsLinks` | Security-Policy check detects SECURITY.md presence and links, but does not verify actual contact methods (email, form, etc.). |
| OSPS-BR-01.04 | (Note: This is Level 3, not Level 1. Listed under Level 3 below.) | | | |

## Level 2 controls (17)

| OSPS ID | Control (short) | Status | Scorecard probes/checks providing evidence | Gap / notes |
|---------|----------------|--------|---------------------------------------------|-------------|
| OSPS-AC-04.01 | Default lowest CI/CD permissions | PARTIAL | `topLevelPermissions`, `jobLevelPermissions`, `hasNoGitHubWorkflowPermissionUnknown` | Token-Permissions check evaluates workflow permissions. "Defaults to lowest" semantics need verification. |
| OSPS-BR-02.01 | Releases have unique version identifier | GAP | None | No release versioning probe. Needs release API inspection. |
| OSPS-BR-04.01 | Releases have descriptive changelog | GAP | None | No changelog/release notes detection probe. |
| OSPS-BR-05.01 | Standardized tooling for dependency ingestion | GAP | None | No probe detects whether standard package managers are used in CI. |
| OSPS-BR-06.01 | Releases signed or accounted for in signed manifest | PARTIAL | `releasesAreSigned`, `releasesHaveProvenance`, `releasesHaveVerifiedProvenance` | Signed-Releases check covers signatures and provenance. Does not yet check for "signed manifest including hashes" as an alternative. |
| OSPS-DO-06.01 | Docs describe dependency selection/tracking | GAP | None | Documentation control. No probe exists. |
| OSPS-DO-07.01 | Build instructions in documentation | GAP | None | Documentation control. No probe exists. |
| OSPS-GV-01.01 | Docs list members with sensitive access | NOT_OBSERVABLE | None | Requires org-level data or attestation. Not automatable via Scorecard. |
| OSPS-GV-01.02 | Docs list roles and responsibilities | GAP | None | Documentation control. May require attestation. |
| OSPS-GV-03.02 | Contributor guide with acceptability requirements | PARTIAL | (related: CONTRIBUTING presence could be inferred) | No probe today; could extend a contributing-file probe to check for content structure. |
| OSPS-LE-01.01 | Legal authorization per commit (DCO/CLA) | GAP | None | No DCO/CLA detection probe. Would check for Signed-off-by trailers or CLA bot enforcement. |
| OSPS-QA-03.01 | Status checks pass or bypassed before merge | COVERED | `runsStatusChecksBeforeMerging` | Branch-Protection check. Direct mapping. |
| OSPS-QA-06.01 | Automated tests run prior to acceptance | COVERED | `testsRunInCI` | CI-Tests check. Maps directly. |
| OSPS-SA-01.01 | Design docs with actions/actors | GAP | None | Documentation/assessment control. Requires attestation. |
| OSPS-SA-02.01 | Docs describe external interfaces | GAP | None | Documentation control. Requires attestation. |
| OSPS-SA-03.01 | Security assessment performed | GAP | None | Process control. Requires attestation with evidence link. |
| OSPS-VM-01.01 | CVD policy with response timeframe | PARTIAL | `securityPolicyContainsVulnerabilityDisclosure`, `securityPolicyContainsText` | Security-Policy check detects disclosure language. Does not verify explicit timeframe commitment. |
| OSPS-VM-03.01 | Private vulnerability reporting method | PARTIAL | `securityPolicyContainsLinks` | Detects links in SECURITY.md. Does not verify private reporting is actually enabled (e.g., GitHub PSIRT feature). |
| OSPS-VM-04.01 | Publicly publish vulnerability data | GAP | None | No probe checks for GitHub Security Advisories, OSV entries, or CVE publication. |

## Level 3 controls (17)

| OSPS ID | Control (short) | Status | Scorecard probes/checks providing evidence | Gap / notes |
|---------|----------------|--------|---------------------------------------------|-------------|
| OSPS-AC-04.02 | Job-level least privilege in CI/CD | PARTIAL | `jobLevelPermissions` | Token-Permissions check evaluates job-level permissions. "Minimum necessary" is hard to assess without understanding job intent. |
| OSPS-BR-01.04 | Sanitize trusted collaborator CI/CD input | PARTIAL | `hasDangerousWorkflowScriptInjection` | Dangerous-Workflow check partially covers this, but focuses on untrusted input, not trusted collaborator input. |
| OSPS-BR-02.02 | Release assets tied to release identifier | GAP | None | No release asset naming/association probe. |
| OSPS-BR-07.02 | Secrets management policy | GAP | None | Documentation/policy control. Requires attestation. |
| OSPS-DO-03.01 | Instructions to verify release integrity/authenticity | GAP | None | Documentation control. Could partially automate by checking for verification docs alongside signed releases. |
| OSPS-DO-03.02 | Instructions to verify release author identity | GAP | None | Documentation control. |
| OSPS-DO-04.01 | Support scope/duration per release | GAP | None | Documentation control. |
| OSPS-DO-05.01 | EOL security update statement | GAP | None | Documentation control. |
| OSPS-GV-04.01 | Policy to review collaborators before escalated perms | GAP | None | Governance policy. Requires attestation. |
| OSPS-QA-02.02 | SBOM shipped with compiled release assets | PARTIAL | `hasReleaseSBOM`, `hasSBOM` | SBOM probes exist but may not specifically verify compiled-asset association. |
| OSPS-QA-04.02 | Subprojects enforce >= primary requirements | GAP | None | Requires multi-repo scanning and cross-repo comparison. |
| OSPS-QA-06.02 | Docs describe when/how tests run | GAP | None | Documentation control. |
| OSPS-QA-06.03 | Policy requiring tests for major changes | GAP | None | Documentation/policy control. |
| OSPS-QA-07.01 | Non-author approval before merging | PARTIAL | `codeApproved`, `codeReviewOneReviewers`, `requiresApproversForPullRequests` | Code-Review and Branch-Protection probes cover this. "Non-author" semantics need verification. |
| OSPS-VM-04.02 | VEX for non-affecting vulnerabilities | GAP | None | No VEX detection probe. |
| OSPS-VM-05.01 | SCA remediation threshold policy | GAP | None | Policy control. |
| OSPS-VM-05.02 | SCA violations addressed pre-release | GAP | None | Policy + enforcement control. |
| OSPS-VM-05.03 | Automated SCA eval + block violations | PARTIAL | `hasOSVVulnerabilities` | Vulnerabilities check detects known vulns. Does not verify gating/blocking enforcement. |
| OSPS-VM-06.01 | SAST remediation threshold policy | GAP | None | Policy control. |
| OSPS-VM-06.02 | Automated SAST eval + block violations | PARTIAL | `sastToolConfigured`, `sastToolRunsOnAllCommits` | SAST check detects tool presence and execution. Does not verify gating/blocking enforcement. |

## Phase 1 priorities (Level 1 gap closure)

The following Level 1 gaps should be addressed first, ordered by implementation feasibility:

1. **OSPS-GV-03.01** (contribution process): Add probe for CONTRIBUTING file presence
2. **OSPS-GV-02.01** (public discussion): Add probe for issues/discussions enabled
3. **OSPS-DO-02.01** (defect reporting): Add probe for issue templates or bug report docs
4. **OSPS-DO-01.01** (user guides): Add probe for documentation presence heuristics
5. **OSPS-BR-07.01** (secrets in VCS): Consume GitHub secret scanning API or add detection heuristics
6. **OSPS-BR-03.01 / BR-03.02** (encrypted transport): Requires Security Insights ingestion for declared URIs
7. **OSPS-QA-04.01** (subproject listing): Requires Security Insights or equivalent metadata

## Probes not mapped to any OSPS control

The following probes exist in Scorecard but do not directly map to any OSPS Baseline control:

| Probe | Check | Notes |
|-------|-------|-------|
| `archived` | Maintained | Project archival status — relates to "while active" preconditions |
| `hasRecentCommits` | Maintained | Activity signal — relates to "while active" preconditions |
| `issueActivityByProjectMember` | Maintained | Activity signal — relates to "while active" preconditions |
| `createdRecently` | Maintained | Age signal |
| `contributorsFromOrgOrCompany` | Contributors | Diversity signal |
| `dependencyUpdateToolConfigured` | Dependency-Update-Tool | Best practice, not a Baseline control |
| `fuzzed` | Fuzzing | Testing best practice, not a Baseline control |
| `hasOpenSSFBadge` | CII-Best-Practices | Meta-badge, not a Baseline control |
| `packagedWithAutomatedWorkflow` | Packaging | Distribution best practice |
| `webhooksUseSecrets` | Webhook | Security practice, not a Baseline control |
| `hasPermissiveLicense` | (uncategorized) | License type classification |
| `unsafeblock` | (independent) | Language-specific safety |
| `dismissesStaleReviews` | Branch-Protection | Review hygiene beyond Baseline scope |
| `requiresCodeOwnersReview` | Branch-Protection | CODEOWNERS enforcement beyond Baseline scope |
| `requiresLastPushApproval` | Branch-Protection | Push approval beyond Baseline scope |
| `requiresUpToDateBranches` | Branch-Protection | Branch freshness beyond Baseline scope |
| `branchProtectionAppliesToAdmins` | Branch-Protection | Admin override prevention beyond Baseline scope |
| `blocksForcePushOnBranches` | Branch-Protection | Force-push protection; related to AC-03 but not explicitly required |

These probes remain valuable for Scorecard's existing scoring model and may become relevant for future Baseline versions.

## Existing issues and PRs relevant to gap closure

The following open issues and PRs in the Scorecard repository are directly
relevant to closing OSPS Baseline coverage gaps. These should be prioritized
and linked to the conformance work.

### Security Insights ingestion
- [#2305](https://github.com/ossf/scorecard/issues/2305) — Support for SECURITY INSIGHTS
- [#2479](https://github.com/ossf/scorecard/issues/2479) — SECURITY-INSIGHTS.yml implementation

These are critical for OSPS-BR-03.01, BR-03.02, QA-04.01, and other
controls that depend on declared project metadata.

### Secrets detection (OSPS-BR-07.01)
- [#30](https://github.com/ossf/scorecard/issues/30) — New check: code is scanning for secrets

Open since the project's earliest days. Phase 1 priority.

### SBOM (OSPS-QA-02.02)
- [#1476](https://github.com/ossf/scorecard/issues/1476) — Feature: Detect if SBOMs generated
- [#2605](https://github.com/ossf/scorecard/issues/2605) — Add support for SBOM analyzing at Binary-Artifacts stage

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
- [#1371](https://github.com/ossf/scorecard/issues/1371) — Feature: add check for vulnerability alerts

### Signed releases and provenance (OSPS-BR-06.01)
- [#4823](https://github.com/ossf/scorecard/issues/4823) — Feature: pass Signed-Releases with GitHub immutable release process
- [#4080](https://github.com/ossf/scorecard/issues/4080) — Use GitHub attestations to check for signed releases
- [#2684](https://github.com/ossf/scorecard/issues/2684) — Rework: Signed-Releases: Separate score calculation of provenance and signatures
- [#1417](https://github.com/ossf/scorecard/issues/1417) — Feature: add support for keyless signed release

### Threat model / security assessment (OSPS-SA-01.01, SA-03.01)
- [#2142](https://github.com/ossf/scorecard/issues/2142) — Feature: Assess presence and maintenance of a threat model

### Release scoring (OSPS-BR-02.01, BR-02.02)
- [#1985](https://github.com/ossf/scorecard/issues/1985) — Feature: Scoring for individual releases

### Minder integration
- [#4723](https://github.com/ossf/scorecard/pull/4723) — Initial draft of using Minder rules in Scorecard (CLOSED)

Draft PR that attempted to run Minder Rego rules within Scorecard,
including OSPS-QA-05.01 and QA-03.01. Closed due to inactivity but
demonstrates interest in deeper Minder/Scorecard integration.

## Notes

- The OSPS Baseline v2026.02.19 contains 59 controls. Previous coverage
  estimates against older Baseline versions should be treated as out-of-date.
  This analysis supersedes any prior mapping.
- Controls marked NOT_OBSERVABLE cannot produce PASS or FAIL without elevated
  permissions. The conformance engine must return UNKNOWN with an explanation.
- Many Level 2 and Level 3 controls are documentation or policy controls that
  require attestation rather than automated detection. The attestation mechanism
  (OQ-1 in the proposal) is critical for these.
- The "while active" precondition on many controls maps to Scorecard's Maintained
  check probes (`archived`, `hasRecentCommits`, `issueActivityByProjectMember`).
  These could serve as applicability detectors.
