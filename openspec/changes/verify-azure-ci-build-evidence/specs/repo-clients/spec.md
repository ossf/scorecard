## ADDED Requirements

### Requirement: Azure DevOps Pull Request Revision Identity

For a completed PR, the Azure DevOps client SHALL use
`LastMergeSourceCommit.CommitId` as `clients.PullRequest.HeadSHA`.
It SHALL retain source, target, and merge revisions privately for build
correlation. Missing source metadata and ambiguous PR identities MUST cause
an error. Existing author, reviewer, and merger identity handling MUST remain
unchanged.

#### Scenario: Source differs from completed merge

- **GIVEN** a completed PR has source commit S and merge commit F
- **WHEN** the client lists its commits
- **THEN** the associated PR has `HeadSHA` S
- **AND** existing CI status reads use S through the contract in
  `openspec/specs/raw-data-collection/spec.md`

#### Scenario: Duplicate PR records

- **GIVEN** multiple commits in the window refer to the same PR
- **WHEN** the client resolves its CI context
- **THEN** repeated records do not create an ambiguous PR identity

#### Scenario: Ambiguous source head

- **GIVEN** different PRs in the window share the requested source head
- **WHEN** the client cannot select a unique PR context
- **THEN** it reports an error instead of choosing a PR arbitrarily

### Requirement: Azure DevOps CI Build Evidence

The Azure client SHALL derive successful CI check runs from actual completed,
succeeded builds that finish no later than PR closure. Each run MUST match
the repository and source or verified PR merge revision, expose the
`azure-pipelines` identity, and include its build evidence URL. Policy
approvals alone MUST NOT produce CI evidence. The client SHALL follow build
continuation tokens.

#### Scenario: Exact source build

- **GIVEN** a successful build matches the repository, source branch, and
  final source commit of a merged PR
- **AND** it finishes no later than PR closure
- **WHEN** the client lists check runs for that source commit
- **THEN** it returns completed successful Azure Pipelines evidence
- **AND** no mandatory build-validation policy or published test result is required

#### Scenario: Separate tested merge

- **GIVEN** a successful build identifies the target repository, PR reason,
  and `refs/pull/<id>/merge` branch
- **WHEN** its tested SHA matches the known PR merge commit or its Git parents
  match the exact source/target pair
- **THEN** the client can return it as CI evidence for the source head
- **AND** the tested SHA remains distinct from the source and completed merge SHAs

#### Scenario: Unrelated or stale evidence

- **GIVEN** a build has another repository, branch, PR, source iteration, or
  source/target pair
- **WHEN** the client correlates builds
- **THEN** that build contributes no CI evidence
- **AND** a build that finishes after PR closure also contributes no evidence

#### Scenario: Build without success

- **GIVEN** a build is failed, canceled, partial, queued, or running
- **WHEN** the client lists CI check runs
- **THEN** it does not report that build as successful

#### Scenario: Approved non-build policy

- **GIVEN** reviewer, work-item, or other policies are approved
- **AND** no matching successful build exists
- **WHEN** the client lists CI check runs
- **THEN** those policy approvals produce no check runs

#### Scenario: Build after page one

- **GIVEN** a build response includes a continuation token
- **WHEN** the first matching build is on a later page
- **THEN** the client reads that page and returns its evidence

#### Scenario: Finding location

- **WHEN** a matched build produces a check run
- **THEN** its URL is the build's web link when available, otherwise its REST URL
- **AND** the unchanged CI probe can use that URL as the finding location
- **AND** scoring follows `openspec/specs/check-evaluation/spec.md`

### Requirement: Azure DevOps CI Evidence Failures

The client SHALL return no CI evidence for a valid query without a matching
build. Required read failures, malformed build or merge metadata, and
repeated continuation tokens MUST return errors. A readable tested commit
with different parents SHALL be a non-match. The client MUST NOT infer build
identity from internal policy context or substitute another revision.

#### Scenario: Build API unavailable

- **WHEN** an authentication, authorization, or transport failure prevents
  a required build read
- **THEN** the error propagates through existing CI collection
- **AND** the check does not disguise it as missing or successful CI

#### Scenario: Tested merge cannot be verified

- **WHEN** a required tested commit cannot be read or lacks its identity or parents
- **THEN** the client returns an error instead of assuming a merge relationship

#### Scenario: Malformed build

- **WHEN** relevant build evidence lacks required revision, result, finish time,
  or URL data
- **THEN** the client returns an error without panicking

#### Scenario: No matching build

- **WHEN** a complete authorized query finds no matching successful build
- **THEN** the client returns an empty check-run collection
- **AND** existing raw collection and scoring determine the result
