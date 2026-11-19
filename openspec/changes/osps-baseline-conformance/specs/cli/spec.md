## MODIFIED Requirements

### Requirement: Scan Flags

The root command SHALL register the scan flags below. Each flag's behavior
SHALL be the one defined by the capability named for it:

- `--checks` (comma-separated, default empty): check selection, defined in
  `check-registry`.
- `--probes` (comma-separated, default empty): probe execution, defined in
  `probes`.
- `--format` (default `default`): format names and gating, defined in
  `result-formats`.
- `--output` (shorthand `-o`, default empty): output destination, defined
  in `result-formats`.
- `--show-details` (boolean, default false): detail rendering, defined in
  `result-formats`.
- `--show-annotations` (boolean, default false): annotation rendering,
  defined in `result-formats`; annotation discovery is defined in
  `maintainer-annotations`.
- `--policy` (default empty): policy file and policy-driven selection,
  defined in `policy`.
- `--catalog` (default empty, experimental): conformance mode and catalog
  selection, defined in `conformance-evaluation`.
- `--profile` (default empty, experimental): the catalog applicability
  group to evaluate, defined in `conformance-evaluation`.

The system SHALL register `--policy` only when `ENABLE_SARIF` is set;
without it, `--policy` MUST be rejected as an unknown flag.

#### Scenario: Output flag shorthand

- **WHEN** a user runs `scorecard --repo=owner/repo -o results.json`
- **THEN** the system treats `-o results.json` as `--output=results.json`

#### Scenario: Policy flag without SARIF enabled

- **GIVEN** `ENABLE_SARIF` is not set
- **WHEN** a user runs `scorecard --repo=owner/repo --policy=policy.yml`
- **THEN** the command fails with `unknown flag: --policy`

#### Scenario: Conformance flags

- **WHEN** a user runs `scorecard --repo=owner/repo
  --catalog=osps-baseline --profile=maturity-1 --format=json`
- **THEN** the system runs in conformance mode, as defined in
  `conformance-evaluation`
