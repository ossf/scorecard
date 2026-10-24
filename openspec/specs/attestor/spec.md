# attestor Specification

## Purpose

`scorecard-attestor` runs a subset of Scorecard checks on the source
repository a container image was built from. It evaluates the raw results
against an attestation policy and, when the policy passes, signs a Google
Cloud Binary Authorization attestation for the image. It is designed to run
as a step in a Google Cloud Build pipeline.

## Requirements

### Requirement: Commands

The system SHALL provide a `scorecard-attestor` command with two
subcommands. `verify` SHALL evaluate a repository against a policy. `attest`
SHALL evaluate the policy and sign an attestation for an image only when the
policy passes.

#### Scenario: Verify only

- **WHEN** a user runs
  `scorecard-attestor verify --policy=policy.yaml --repo-url=github.com/owner/repo`
- **THEN** the system evaluates the policy and logs whether it passed
- **AND** creates no attestation

#### Scenario: Attest on pass

- **GIVEN** the repository satisfies the policy
- **WHEN** a user runs `scorecard-attestor attest` with a policy, repository,
  image, and signing key
- **THEN** the system signs an attestation for the image

### Requirement: Evaluation Inputs

The system MUST require the `--policy` and `--repo-url` flags on both
subcommands and SHALL reject the command when either flag is not given.
The system SHALL accept an optional `--commit`. When `--repo-url` is given
as an empty string, the system SHALL use the `REPO_NAME` environment
variable and MUST fail with `param repoURL is empty` if that is also empty.
When `--commit` is empty, the system SHALL use the `COMMIT_SHA` environment
variable, falling back to `HEAD`.

#### Scenario: Missing policy flag

- **WHEN** a user runs `scorecard-attestor verify --repo-url=owner/repo`
- **THEN** the command fails because the required flag `policy` is not set

#### Scenario: Commit from Cloud Build

- **GIVEN** `COMMIT_SHA=abc123` and no `--commit` flag
- **WHEN** the system runs
- **THEN** it scans the repository at commit `abc123`

#### Scenario: No commit available

- **GIVEN** neither `--commit` nor `COMMIT_SHA` is set
- **WHEN** the system runs
- **THEN** it logs `commit not specified, running on HEAD` and scans `HEAD`

### Requirement: Repository Host Resolution

The system SHALL interpret the repository URL as a GitLab repository when
possible and as a GitHub repository otherwise. When `SCORECARD_EXPERIMENTAL`
is set, the system SHALL first try Azure DevOps and MUST fail for an Azure
DevOps host URL that cannot be used.

#### Scenario: GitHub repository

- **WHEN** a user passes `--repo-url=github.com/owner/repo`
- **THEN** the system scans the GitHub repository

#### Scenario: Unusable repository

- **WHEN** the repository URL cannot be resolved to any supported host
- **THEN** the system fails with `couldn't set up clients: ...`

### Requirement: Attestation Policy Format

The system SHALL read the policy as a YAML file with these optional keys:
`preventBinaryArtifacts` (boolean), `allowedBinaryArtifacts` (list of glob
paths), `preventKnownVulnerabilities` (boolean),
`preventUnpinnedDependencies` (boolean), `allowedUnpinnedDependencies`
(list of entries with `filepath`, `packagename`, and `version`),
`ensureCodeReviewed` (boolean), and `codeReviewRequirements` (with
`requiredApprovers` list and `minReviewers` integer). Unknown keys SHALL be
ignored. A file that cannot be read or does not match these types MUST fail
with `fail to load scorecard attestation policy: ...`.

#### Scenario: Valid policy

- **GIVEN** a policy with `preventBinaryArtifacts: true` and no other keys
- **WHEN** the system loads it
- **THEN** only the binary artifact rule is enabled

#### Scenario: Wrong shape

- **GIVEN** a policy where `allowedUnpinnedDependencies` is a map instead of
  a list
- **WHEN** the system loads it
- **THEN** it fails with `fail to load scorecard attestation policy`

### Requirement: Checks Run Per Policy

The system SHALL run only the Scorecard checks needed by enabled rules:
`Binary-Artifacts` for `preventBinaryArtifacts`, `Vulnerabilities` for
`preventKnownVulnerabilities`, `Pinned-Dependencies` for
`preventUnpinnedDependencies`, and `Code-Review` for `ensureCodeReviewed`.
The scan SHALL use Scorecard's default commit window of 30 commits. A scan
failure MUST stop the command with `scorecard.Run: ...`.

#### Scenario: Two rules enabled

- **GIVEN** a policy enabling `preventBinaryArtifacts` and
  `ensureCodeReviewed`
- **WHEN** the system runs
- **THEN** it runs only the `Binary-Artifacts` and `Code-Review` checks

### Requirement: Rule Evaluation Order

The system SHALL evaluate enabled rules in this order: binary artifacts,
unpinned dependencies, known vulnerabilities, code review. The policy SHALL
fail at the first rule that fails and SHALL pass only when every enabled
rule passes. A policy with no enabled rules SHALL pass.

#### Scenario: Early failure

- **GIVEN** both binary artifact and code review rules are enabled
- **AND** the repository contains an unallowed binary
- **WHEN** the policy is evaluated
- **THEN** the policy fails and logs
  `image failed scorecard attestation policy check`

#### Scenario: All rules pass

- **WHEN** every enabled rule passes
- **THEN** the system logs `image passed scorecard attestation policy check`

### Requirement: Binary Artifact Rule

When `preventBinaryArtifacts` is true, the system SHALL fail the policy if
the repository contains any binary artifact whose path matches none of the
`allowedBinaryArtifacts` glob patterns.

#### Scenario: Allowed by glob

- **GIVEN** `allowedBinaryArtifacts: ["tools/*.jar"]`
- **AND** the only binary is `tools/build.jar`
- **WHEN** the rule is evaluated
- **THEN** the rule passes and logs that the artifact was ignored

#### Scenario: Unallowed binary

- **GIVEN** an empty allow list and a binary at `bin/app`
- **WHEN** the rule is evaluated
- **THEN** the rule fails and logs `binary detected path:bin/app ...`

### Requirement: Unpinned Dependency Rule

When `preventUnpinnedDependencies` is true, the system SHALL fail the policy
if any dependency is not pinned by hash or commit and is not allowed. A
dependency SHALL be allowed when its name equals an entry's `packagename` or
its file path starts with an entry's `filepath`. The `version` field SHALL
NOT affect evaluation.

#### Scenario: Allowed by file path

- **GIVEN** `allowedUnpinnedDependencies: [{filepath: Dockerfile}]`
- **AND** the only unpinned dependency is declared in `Dockerfile`
- **WHEN** the rule is evaluated
- **THEN** the rule passes

#### Scenario: Unpinned dependency

- **GIVEN** no allow list and an unpinned dependency
- **WHEN** the rule is evaluated
- **THEN** the rule fails and logs `found unpinned dependency ...`

### Requirement: Known Vulnerability Rule

When `preventKnownVulnerabilities` is true, the system SHALL fail the policy
if the Vulnerabilities check reports one or more known vulnerabilities from
OSV, and SHALL log the number found.

#### Scenario: Vulnerability present

- **GIVEN** the scan reports one vulnerability
- **WHEN** the rule is evaluated
- **THEN** the rule fails and logs `found 1 vulnerabilities in package`

### Requirement: Code Review Rule

When `ensureCodeReviewed` is true, the system SHALL check every
default-branch changeset in the commit window. A changeset SHALL fail when
its count of `APPROVED` reviews is below `minReviewers`, or when
`requiredApprovers` is non-empty and none of those users approved it. When
neither `minReviewers` nor `requiredApprovers` is set, `minReviewers` SHALL
default to 1.

#### Scenario: Default requirement

- **GIVEN** `ensureCodeReviewed: true` without `codeReviewRequirements`
- **AND** a changeset with no approvals
- **WHEN** the rule is evaluated
- **THEN** the rule fails and logs `not enough approvals for ...`

#### Scenario: Required approver missing

- **GIVEN** `requiredApprovers: [alice]` and a changeset approved only by
  `bob`
- **WHEN** the rule is evaluated
- **THEN** the rule fails

#### Scenario: Required approver present

- **GIVEN** `requiredApprovers: [alice]` and every changeset approved by
  `alice`
- **WHEN** the rule is evaluated
- **THEN** the rule passes

### Requirement: Signing Key Selection

The `attest` subcommand MUST require `--image` and one signing key. The
system SHALL use a Cloud KMS key from `--kms-key-name` first, then a PGP
key file from `--pgp-private-key` with optional `--pgp-passphrase`, then a
PKIX key file from `--pkix-private-key` with `--pkix-alg`. A KMS key MUST be
accompanied by `--kms-digest-alg` set to `SHA256`, `SHA384`, or `SHA512`,
matched case-insensitively. A PKIX key MUST be accompanied by a known
signature algorithm.

#### Scenario: No signing key

- **GIVEN** the policy passes
- **WHEN** `attest` runs without any key flag
- **THEN** it fails with
  `neither kms_key_name, pgp_private_key, or pkix_private_key is specified`

#### Scenario: KMS without digest

- **WHEN** `attest` runs with `--kms-key-name` and no `--kms-digest-alg`
- **THEN** it fails with `kms_digest_alg is unspecified, ...`

#### Scenario: KMS signing

- **WHEN** `attest` runs with `--kms-key-name` and `--kms-digest-alg=sha256`
- **THEN** the system signs with the KMS key using `SHA256`

### Requirement: Attestation Creation

The system SHALL store the attestation in the project given by
`--attestation-project`, defaulting to the project parsed from the image
URL, under the note `projects/<project>/notes/ossf-scorecard-attestation`.
The system SHALL replace an existing attestation only when `--overwrite` is
set. An invalid note name or signing failure MUST fail the command.

#### Scenario: Default attestation project

- **GIVEN** `--image=gcr.io/my-proj/app@sha256:abcd` and no
  `--attestation-project`
- **WHEN** the policy passes and `attest` signs
- **THEN** the attestation is stored under
  `projects/my-proj/notes/ossf-scorecard-attestation`

### Requirement: Exit Status

The system SHALL exit with status 0 when the policy is evaluated, whether
it passes or fails; a failing policy SHALL only suppress signing. The
system MUST print the error to standard error and exit with status 1 when
an input is missing, the policy cannot be loaded, the scan fails, or
signing fails.

#### Scenario: Policy fails

- **WHEN** `attest` evaluates a policy that fails
- **THEN** the system creates no attestation and exits with status 0

#### Scenario: Internal error

- **WHEN** the policy file does not exist
- **THEN** the system prints the error and exits with status 1

### Requirement: Container Image Distribution

The system SHALL be distributable as a container image whose entrypoint is
`/scorecard-attestor`. The project's Cloud Build configuration SHALL build
that image and publish it as `gcr.io/openssf/scorecard-attestor` tagged
`latest` and with the source commit SHA.

#### Scenario: Cloud Build publish

- **WHEN** the attestor Cloud Build configuration runs for commit `<sha>`
- **THEN** it publishes `gcr.io/openssf/scorecard-attestor:latest` and
  `gcr.io/openssf/scorecard-attestor:<sha>`
