## ADDED Requirements

### Requirement: Azure DevOps Branch Policy Data

The Azure client SHALL return equivalent default and named branch data
using every applicable policy page. Only enabled, blocking, non-deleted PR
policies SHALL establish protection. Push restrictions and conditional
rules SHALL NOT receive unconditional credit. Unknown controls SHALL remain
unknown. Required-data failures SHALL preserve their context and cause.

#### Scenario: Required policies on the default branch

- **GIVEN** applicable required reviewer and build policies
- **WHEN** the default branch or that branch's explicit name is requested
- **THEN** both return equivalent populated protection controls
- **AND** repeated default access uses the same initialized snapshot
- **AND** queries identify the repository and fully qualified ref

#### Scenario: Unrelated or non-mandatory policies

- **GIVEN** only secrets scanning, optional, disabled, or deleted policies
- **WHEN** branch protection is requested
- **THEN** those policies do not establish PR protection

#### Scenario: A policy on a later page

- **GIVEN** an applicable policy appears after a continuation token
- **WHEN** the branch is read
- **THEN** the later policy contributes to the returned controls
- **AND** a repeated token fails explicitly rather than looping

#### Scenario: Self-votes and conditional review rules

- **GIVEN** a minimum-review policy permits the requestor's vote to count
- **WHEN** the independent-approval floor is calculated
- **THEN** at most one requestor vote is subtracted, with a floor of zero
- **AND** path-filtered rules do not establish a universal reviewer count
- **AND** overlapping policy counts are not added together
- **AND** missing creator-vote data does not invent an independent reviewer

#### Scenario: Incomplete freshness evidence

- **GIVEN** a policy excludes the last pusher but does not establish fresh
  approval, or a build expires only after a delay
- **WHEN** protection data is returned
- **THEN** it does not claim the corresponding unconditional freshness
  protection
- **AND** last-push approval requires both freshness and last-pusher exclusion
- **AND** up-to-date builds require immediate expiration on target updates

#### Scenario: Required policy data cannot be read

- **WHEN** an applicable-policy read fails or required settings are malformed
- **THEN** the branch request returns an error with its context and cause
- **AND** it does not return a successful empty policy result

#### Scenario: Required contexts and ownership limits

- **GIVEN** required build and status policies or a required-reviewer group
- **WHEN** protection data is returned
- **THEN** build contexts use `azure-pipelines/<definition-id>`
- **AND** status contexts preserve the configured genre/name
- **AND** contexts identify requirements rather than successful runs
- **AND** group reviewers do not establish CODEOWNERS enforcement

#### Scenario: Unknown policy semantics

- **GIVEN** an active policy type or optional setting cannot be interpreted
- **WHEN** protection data is returned
- **THEN** the corresponding controls remain unknown
- **AND** unknown data does not become a false setting

### Requirement: Azure DevOps Bounded Permission Evidence

The client SHALL evaluate effective caller permissions on the exact branch
without an administrator shortcut. Positive evaluations SHALL establish
only proven exceptions, and caller denials SHALL NOT establish global
protection. Administrator coverage SHALL remain unknown. Permission
authorization denials SHALL retain policy data and report the limitation.
Other permission failures SHALL remain explicit errors.

#### Scenario: Caller permissions do not prove a global restriction

- **GIVEN** the caller lacks forced-update authority
- **WHEN** branch permission evidence is evaluated
- **THEN** force-push and deletion protection remain unknown
- **AND** administrator enforcement is not inferred

#### Scenario: A proven forced-update exception

- **GIVEN** the caller can contribute and force updates
- **AND** applicable PR policies either do not block the update or the
  caller has pushing-bypass authority
- **WHEN** permission evidence is returned
- **THEN** forced updates are reported as allowed
- **AND** the caller-only scope is reported on standard error

#### Scenario: Permission access is denied

- **GIVEN** readable PR policies and a permission authorization denial
- **WHEN** branch protection is requested
- **THEN** mapped PR policy data remains available
- **AND** the permission limitation is reported without credential data
- **AND** permission-backed controls remain unknown

#### Scenario: Invalid permission evidence

- **WHEN** permission evaluation fails for a non-authorization reason or
  returns missing, duplicate, or mismatched evaluations
- **THEN** the branch request returns an explicit error
- **AND** missing data does not become a false permission value

#### Scenario: Case-sensitive and nested branch names

- **GIVEN** branch names differ by case or include folders or non-ASCII text
- **WHEN** their security tokens are constructed
- **THEN** the exact names remain distinct and slash hierarchy is preserved
- **AND** the client requests permissions for the matching branch
- **AND** non-slash characters use UTF-16 little-endian hexadecimal encoding

#### Scenario: Policy-protected deletion and unknown controls

- **GIVEN** a required PR policy and a caller permission evaluation
- **WHEN** the bounded evaluation cannot establish deletion authority
- **THEN** deletion protection remains unknown
- **AND** unknown controls use the existing `check-evaluation` behavior
- **AND** no global protection is invented from the caller's permissions
