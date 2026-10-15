# policy Specification

## Purpose

A Scorecard policy is a YAML file that sets, per check, a minimum score and
whether the check is enforced or disabled. This spec defines the file format,
its validation, and how a policy affects check selection. The `--policy` flag
and its `ENABLE_SARIF` gate are defined in `cli`; how SARIF output uses the
policy is defined in `result-formats`.

## Requirements

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
`readPolicy` error and a non-zero exit status before any repository is
scanned, when any of the following holds: the file cannot be read; the YAML
is malformed; `version` is not `1`; a key under `policies` is not a known
check name (experimental checks included); a `mode` is not `enforced` or
`disabled`; a `score` is below 0 or above 10; or the same check is defined
more than once. A repeated check name is rejected by the YAML parser as a
duplicate mapping key, so the error is a YAML parse error rather than a
repeated-check error.

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

#### Scenario: Invalid policy stops the command

- **GIVEN** `ENABLE_SARIF` is set and the policy file has `version: 2`
- **WHEN** a user runs with `--policy=policy.yml`
- **THEN** the command fails with a `readPolicy` error and a non-zero exit
  status
- **AND** no repository is scanned

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
every check named in the policy, regardless of its mode, skipping any check
that does not support the required request types (for example `--local` or a
non-HEAD `--commit`) or the repository's type. When `--checks` is supplied,
it MUST take precedence over the policy for selecting checks.

#### Scenario: Policy selects checks

- **GIVEN** a policy naming Token-Permissions, Branch-Protection, and
  Vulnerabilities
- **WHEN** a user runs without `--checks`
- **THEN** exactly those three checks run, including the disabled one

#### Scenario: Unsupported check skipped

- **GIVEN** `ENABLE_SARIF` is set and a policy lists License and Maintained
- **WHEN** a user scans a local directory with `--local`
- **THEN** License runs
- **AND** Maintained is skipped because it does not support file-based
  requests

### Requirement: Every Selected Check Needs a Policy

When a policy is supplied, the system MUST fail check selection with
"checks don't have policies" if any selected check has no entry in the
policy, and MUST log the name of one such check. Selected checks are
visited in map iteration order, so when several lack an entry, which name is
logged is arbitrary.

#### Scenario: Extra check requested

- **GIVEN** a policy that names only Token-Permissions
- **WHEN** a user runs with `--policy` and `--checks Binary-Artifacts`
- **THEN** the run fails with "checks don't have policies"
- **AND** the log says "check Binary-Artifacts has no policy declared"

### Requirement: Policy Does Not Change Scores

The system MUST NOT use the policy to change check scores, the aggregate
score, or the contents of non-SARIF formats.

#### Scenario: Disabled check in JSON

- **GIVEN** a policy that disables Token-Permissions
- **WHEN** a user runs with `--policy` and `--format json`
- **THEN** Token-Permissions still appears in `checks` with its computed
  score
