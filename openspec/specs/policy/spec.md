# policy Specification

## Purpose

A Scorecard policy is a YAML file that sets, per check, a minimum score and
whether the check is enforced or disabled. The CLI accepts it with
`--policy`, uses it to choose which checks run, and uses it to decide
which checks produce SARIF results. This spec defines the file format,
its validation, and how it affects check selection and SARIF output.

## Requirements

### Requirement: Policy Flag Gating

The system SHALL register the `--policy` flag only when the `ENABLE_SARIF`
environment variable is set. When a policy file is supplied while
`ENABLE_SARIF` is not set, option validation MUST fail with "policy file is
not supported yet".

#### Scenario: Flag unavailable

- **GIVEN** `ENABLE_SARIF` is not set
- **WHEN** a user passes `--policy policy.yml`
- **THEN** the command rejects the flag

#### Scenario: Flag available

- **GIVEN** `ENABLE_SARIF=1` is set
- **WHEN** a user passes `--policy policy.yml`
- **THEN** the policy file is read before any repository is scanned

### Requirement: Policy File Format

The system SHALL read a policy as YAML with a top-level integer `version`
and a `policies` mapping from check name to an object with `mode` and
`score`. `mode` MUST be `enforced` or `disabled`; `score` is an integer
minimum score and MUST default to 0 when omitted.

#### Scenario: Valid policy

- **GIVEN** the policy

  ```yaml
  version: 1
  policies:
    Token-Permissions:
      score: 3
      mode: disabled
    Branch-Protection:
      score: 5
      mode: enforced
  ```

- **WHEN** it is parsed
- **THEN** Token-Permissions is disabled with score 3
- **AND** Branch-Protection is enforced with score 5

#### Scenario: Score omitted

- **GIVEN** a policy entry with `mode: disabled` and no `score`
- **WHEN** it is parsed
- **THEN** the entry's score is 0

### Requirement: Policy Validation

The system MUST reject a policy file, and abort the run with a
`readPolicy` error, when any of the following holds: the file cannot be
read; the YAML is malformed; `version` is not `1`; a key under `policies`
is not a known check name (experimental checks included); a `mode` is not
`enforced` or `disabled`; a `score` is below 0 or above 10; or the same
check is defined more than once.

#### Scenario: Unsupported version

- **GIVEN** a policy with `version: 2`
- **WHEN** it is parsed
- **THEN** parsing fails with "invalid version"

#### Scenario: Unknown check

- **GIVEN** a policy with an entry for `Unknown-Check`
- **WHEN** it is parsed
- **THEN** parsing fails with "invalid check name: Unknown-Check"

#### Scenario: Invalid mode

- **GIVEN** a policy entry with `mode: unknown`
- **WHEN** it is parsed
- **THEN** parsing fails with "invalid mode: unknown"

#### Scenario: Score out of range

- **GIVEN** a policy entry with `score: 11` or `score: -1`
- **WHEN** it is parsed
- **THEN** parsing fails with "invalid score"

#### Scenario: Duplicate check

- **GIVEN** a policy that lists `Token-Permissions` twice
- **WHEN** it is parsed
- **THEN** parsing fails

### Requirement: Case-Sensitive Check Names in the File

The system SHALL require check names under `policies` to match a known
check name exactly, including case.

#### Scenario: Lowercase check name

- **GIVEN** a policy with an entry for `token-permissions`
- **WHEN** it is parsed
- **THEN** parsing fails with "invalid check name"

### Requirement: No Policy by Default

When no policy file is supplied, the system SHALL run without a policy:
check selection MUST fall back to `--checks` or to all supported checks,
and every format except `sarif` MUST behave as if no policy exists.

#### Scenario: JSON without a policy

- **WHEN** a user runs `--format json` without `--policy`
- **THEN** all supported checks run and results are written normally

### Requirement: Check Selection From a Policy

When a policy is supplied and `--checks` is not, the system SHALL run
every check named in the policy, regardless of its mode, that supports the
current request type. When `--checks` is supplied, it MUST take precedence
over the policy for selecting checks.

#### Scenario: Policy selects checks

- **GIVEN** a policy naming Token-Permissions, Branch-Protection, and
  Vulnerabilities
- **WHEN** a user runs without `--checks`
- **THEN** exactly those three checks run, including the disabled one

#### Scenario: Unsupported check skipped

- **GIVEN** a policy that names a check requiring the GitHub API
- **WHEN** a user scans with `--local`
- **THEN** that check is skipped

### Requirement: Every Selected Check Needs a Policy

When a policy is supplied, the system MUST fail check selection with
"checks don't have policies" if any selected check has no entry in the
policy, and MUST log the name of the first such check.

#### Scenario: Extra check requested

- **GIVEN** a policy that names only Token-Permissions
- **WHEN** a user runs with `--policy` and `--checks Binary-Artifacts`
- **THEN** the run fails with "checks don't have policies"
- **AND** the log says "check Binary-Artifacts has no policy declared"

### Requirement: Policy Drives SARIF Results

The system SHALL use the policy when rendering `--format sarif`: a check
whose mode is `disabled` MUST produce no SARIF results, and an `enforced`
check MUST produce results only when its score is below the policy score
and is not inconclusive. Rendering SARIF for a check with no policy entry,
including any run without `--policy`, MUST fail with "missing policy for
check".

#### Scenario: Enforced check meets the threshold

- **GIVEN** a policy enforcing Branch-Protection with score 5
- **AND** Branch-Protection scored 5
- **WHEN** results are rendered as SARIF
- **THEN** the run contains the Branch-Protection rule but no result for
  it

#### Scenario: Enforced check misses the threshold

- **GIVEN** a policy enforcing Branch-Protection with score 5
- **AND** Branch-Protection scored 3
- **WHEN** results are rendered as SARIF
- **THEN** at least one Branch-Protection result is emitted

### Requirement: Policy Does Not Change Scores

The system MUST NOT use the policy to change check scores, the aggregate
score, or the contents of non-SARIF formats.

#### Scenario: Disabled check in JSON

- **GIVEN** a policy that disables Token-Permissions
- **WHEN** a user runs with `--policy` and `--format json`
- **THEN** Token-Permissions still appears in `checks` with its computed
  score
