## ADDED Requirements

### Requirement: Conformance Mode

The system SHALL run in conformance mode when `--catalog` is set. In
conformance mode the system SHALL run the probes that the selected
catalog's mappings name, evaluate each mapped requirement from those
probes' findings, and write requirement results. It SHALL NOT compute
check scores or an aggregate score in that run.

Conformance mode MUST NOT change probe findings, check scores, or output
formats for runs without `--catalog`.

#### Scenario: Conformance run

- **WHEN** a user runs `scorecard --repo=github.com/owner/repo
  --catalog=osps-baseline --profile=maturity-1 --format=json`
- **THEN** the output contains requirement results and no check scores

#### Scenario: Native run unchanged

- **GIVEN** the same repository state
- **WHEN** a user runs `scorecard --repo=github.com/owner/repo
  --format=json` with and without this capability built in
- **THEN** both outputs report the same check scores and aggregate score

### Requirement: Catalog Selection

`--catalog` SHALL take a catalog ID, optionally followed by `@` and a
release version, for example `osps-baseline@v2026.08.28`. The system
SHALL accept only catalogs and releases built into the binary. Without a
version, the system SHALL use the newest built-in release of that
catalog and record which release it used.

The first built-in catalog SHALL be `osps-baseline` at release
`v2026.08.28`.

#### Scenario: Version omitted

- **WHEN** a user passes `--catalog=osps-baseline`
- **THEN** the system evaluates the newest built-in OSPS Baseline release
- **AND** the output records that release's version

#### Scenario: Unknown catalog

- **WHEN** a user passes `--catalog=osps`
- **THEN** option validation fails with an "unsupported catalog" error
  that lists the built-in catalog IDs

#### Scenario: Unknown release

- **WHEN** a user passes `--catalog=osps-baseline@v2025.01.01`
- **THEN** option validation fails with an "unsupported catalog version"
  error that lists the built-in releases

### Requirement: Catalog Release Immutability

A built-in catalog release MUST NOT change meaning after it ships. A
correction SHALL ship as a change to the evaluator, recorded by the
Scorecard version. It SHALL NOT edit the release's requirement text or
applicability.

#### Scenario: Evaluator fix

- **GIVEN** a mapping bug is fixed for `osps-baseline@v2026.08.28`
- **WHEN** the fix ships
- **THEN** the release's requirement text and applicability are unchanged
- **AND** the output's Scorecard version identifies the fixed evaluator

### Requirement: Profile Selection

`--profile` SHALL take the ID of one applicability group defined by the
selected catalog release, and MUST be set whenever `--catalog` is set.
For `osps-baseline` the IDs SHALL be `maturity-1`, `maturity-2`, and
`maturity-3`. The system SHALL NOT assume that a catalog's groups are
ordered.

#### Scenario: Profile missing

- **WHEN** a user passes `--catalog=osps-baseline` without `--profile`
- **THEN** option validation fails with "`--profile` is required with
  `--catalog`"

#### Scenario: Unknown profile

- **WHEN** a user passes `--catalog=osps-baseline --profile=maturity1`
- **THEN** option validation fails with an "unsupported profile" error
  that lists `maturity-1`, `maturity-2`, and `maturity-3`

#### Scenario: Profile without catalog

- **WHEN** a user passes `--profile=maturity-1` without `--catalog`
- **THEN** option validation fails with "`--profile` requires
  `--catalog`"

### Requirement: Conformance Option Constraints

In conformance mode the system MUST reject `--checks`, `--probes`, and
`--policy`, and MUST reject `--repos` and `--org`. It SHALL accept
`--format` values `json` and `gemara` only. These failures SHALL be
reported with the other validation errors, as defined in `cli`.

#### Scenario: Checks with catalog

- **WHEN** a user passes `--catalog=osps-baseline --profile=maturity-1
  --checks=Branch-Protection`
- **THEN** option validation fails with "`--checks` cannot be used with
  `--catalog`"

#### Scenario: Format omitted in conformance mode

- **WHEN** a user passes `--catalog=osps-baseline --profile=maturity-1`
  without `--format`
- **THEN** option validation fails with "`--catalog` requires
  `--format=json` or `--format=gemara`"

### Requirement: Requirement Results

The system SHALL give each evaluated requirement exactly one result:

- `PASS`: the evidence shows the requirement is met.
- `FAIL`: the evidence shows the requirement is not met.
- `UNKNOWN`: evidence needed to decide is missing, or a probe errored or
  could not observe it.
- `NOT_APPLICABLE`: the requirement does not apply to the selected
  profile.
- `NEEDS_REVIEW`: evidence exists, but deciding needs judgment that the
  evaluator does not make. A person or another tool can make it.

#### Scenario: One result per requirement

- **WHEN** a conformance run evaluates OSPS-AC-03.01
- **THEN** the output gives it exactly one of the five results

### Requirement: Probe Outcome Handling

A probe finding with outcome `Error`, `NotAvailable`, or `NotSupported`
MUST NOT produce `PASS` or `FAIL` for a requirement that needs that
finding. A probe finding with outcome `NotApplicable` MUST NOT by itself
produce `NOT_APPLICABLE` for a requirement.

#### Scenario: Probe could not read settings

- **GIVEN** `requiresPRsToChangeCode` reports `NotAvailable` for the
  default branch
- **WHEN** OSPS-AC-03.01 is evaluated
- **THEN** its result is `UNKNOWN`, not `FAIL`

#### Scenario: Probe found no branches

- **GIVEN** `requiresPRsToChangeCode` reports `NotApplicable` because no
  branches were found
- **WHEN** OSPS-AC-03.01 is evaluated
- **THEN** its result is `UNKNOWN`

### Requirement: Applicability

The system SHALL evaluate a requirement only when the selected profile
is among the requirement's applicability groups in the catalog release.
A mapped requirement that does not list the selected profile SHALL get
`NOT_APPLICABLE` and SHALL NOT run probes on its behalf.

#### Scenario: Requirement outside the profile

- **GIVEN** a mapped requirement whose applicability is only `maturity-3`
- **WHEN** a user selects `--profile=maturity-1`
- **THEN** that requirement's result is `NOT_APPLICABLE`

### Requirement: Coverage

The output SHALL list every requirement in the catalog release that the
selected profile includes and the system has no mapping for, as not
evaluated. Not evaluated is a coverage state, not a result. The output
MUST NOT describe a run as conformance with the whole catalog unless
every requirement in the profile was evaluated.

#### Scenario: Partial coverage

- **GIVEN** mappings exist only for OSPS-AC-03.01 and OSPS-AC-03.02
- **WHEN** a user selects `--profile=maturity-1`
- **THEN** the output reports results for those two requirements
- **AND** lists every other `maturity-1` requirement as not evaluated

### Requirement: Evidence

Each requirement result SHALL record the probe findings it used. Each
recorded finding SHALL include its probe name, outcome, message, and
values. The result SHALL also include a short explanation of how the
findings led to it.

#### Scenario: Evidence for a failed requirement

- **GIVEN** `requiresPRsToChangeCode` reports `False` for the default
  branch
- **WHEN** OSPS-AC-03.01 is evaluated
- **THEN** its result is `FAIL`
- **AND** its evidence includes that finding and the branch name

### Requirement: Assessment Identity

Each conformance result SHALL record the repository, the commit
assessed, the Scorecard version and commit, the catalog ID and release
version, the profile, and the time of the assessment.

#### Scenario: Identity fields

- **WHEN** a conformance run completes for `github.com/owner/repo`
- **THEN** the output records the repository, the commit, the Scorecard
  version, `osps-baseline`, its release version, the profile, and a
  timestamp

### Requirement: OSPS-AC-03.01 Evaluation

The system SHALL evaluate OSPS-AC-03.01 (direct commits to the primary
branch are prevented) from the default branch's findings for
`requiresPRsToChangeCode` and `branchProtectionAppliesToAdmins`:

- `PASS` when both are `True`.
- `FAIL` when `requiresPRsToChangeCode` is `False`.
- `UNKNOWN` otherwise, except as defined in OSPS-AC-03.01 Administrator
  Bypass.

Findings for branches other than the default branch SHALL be ignored.

#### Scenario: Protected for everyone

- **GIVEN** both probes report `True` for the default branch
- **WHEN** OSPS-AC-03.01 is evaluated
- **THEN** its result is `PASS`

### Requirement: OSPS-AC-03.01 Administrator Bypass

The system SHALL give OSPS-AC-03.01 `NEEDS_REVIEW` when
`requiresPRsToChangeCode` is `True` and `branchProtectionAppliesToAdmins`
is `False` for the default branch. The requirement text does not say
whether an administrator exception is allowed, so the evaluator does not
decide it.

#### Scenario: Administrators can bypass

- **GIVEN** `requiresPRsToChangeCode` is `True` and
  `branchProtectionAppliesToAdmins` is `False` for the default branch
- **WHEN** OSPS-AC-03.01 is evaluated
- **THEN** its result is `NEEDS_REVIEW`

### Requirement: OSPS-AC-03.02 Evaluation

The system SHALL evaluate OSPS-AC-03.02 (deleting the primary branch
requires explicit confirmation) from the default branch's finding for
`blocksDeleteOnBranches`:

- `PASS` when the finding is `True`.
- `FAIL` when the finding is `False`.
- `UNKNOWN` otherwise.

A block on deletion SHALL count as meeting the requirement, since it is
stricter than requiring confirmation.

#### Scenario: Deletion blocked

- **GIVEN** `blocksDeleteOnBranches` reports `True` for the default
  branch
- **WHEN** OSPS-AC-03.02 is evaluated
- **THEN** its result is `PASS`

#### Scenario: Deletion allowed

- **GIVEN** `blocksDeleteOnBranches` reports `False` for the default
  branch
- **WHEN** OSPS-AC-03.02 is evaluated
- **THEN** its result is `FAIL`

### Requirement: Experimental Status

While conformance mode is experimental, the system SHALL label
`--catalog` and `--profile` as experimental in `--help`. It SHALL print a
warning to standard error on every conformance run, and SHALL mark
conformance JSON output as experimental. No environment variable SHALL
be needed to use conformance mode.

#### Scenario: Warning on each run

- **WHEN** a user runs a conformance scan
- **THEN** standard error includes a warning that conformance mode is
  experimental and its output may change
