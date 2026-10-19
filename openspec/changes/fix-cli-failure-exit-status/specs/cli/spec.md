## MODIFIED Requirements

### Requirement: Exit Status

The system MUST exit with status 1 and print
`error during command execution: <error>` when validation fails, the policy
file is invalid, check selection fails, the organization cannot be listed,
or package resolution fails.

When a repository cannot be parsed or scanned, the system SHALL print
`Skipping <uri>: <error>` to standard error, record the failure, and
continue with the remaining repositories. When formatting or writing a
repository's results fails, the system SHALL print
`Failed to format results for <uri>: <error>` to standard error, record the
failure, and continue. A completed scan whose output fails SHALL NOT be
reported as skipped. A formatting failure SHALL NOT prevent the system
from detecting runtime errors in that repository's completed checks.

When any check reports a runtime error, the system SHALL print
`Check <name> failed for <uri>: <error>` to standard error and finish all
repositories. If check runtime errors are the only failures, the final
error SHALL include "one or more checks failed during execution".

After all requested repositories have been processed, the system MUST
exit with status 1 if any repository, output, or check runtime error
occurred. For repository-scan and result-output failures, the final error
SHALL preserve the failed repository's context and original cause.
It SHALL also include the check-failure summary when a check runtime error
occurred. This rule SHALL apply to both `--repos` and repositories
selected through `--org`. Archived repositories excluded during
organization enumeration SHALL NOT count as scan failures.

The stage-specific messages SHALL be the canonical immediate diagnostics.
The final command error SHALL be the canonical aggregate failure summary
that determines exit status. Repeated details in command error reporting
SHALL NOT replace these diagnostics or change the exit-status rule.

Check scores and policy thresholds SHALL NOT affect exit status. When
execution completes without an error, the system MUST exit with status 0,
including when check results are inconclusive.

#### Scenario: Check runtime error

- **GIVEN** a check returns a runtime error for one repository
- **WHEN** the scan of all repositories completes
- **THEN** the system exits with status 1 and
  "one or more checks failed during execution"

#### Scenario: Skipped repository

- **WHEN** a user runs `scorecard --repo=::bad::`
- **THEN** the system prints a `Skipping` message and exits with status 1
- **AND** the final error includes the repository and its parse failure

#### Scenario: Unknown probe skips the repository

- **WHEN** a user runs `scorecard --local=. --probes=nope`
- **THEN** the system prints `Skipping .: run: getting probe "nope": ...`
  to standard error and exits with status 1

#### Scenario: Formatting error does not stop a multi-repo scan

- **GIVEN** a scan of two repositories with `--repos`
- **WHEN** formatting the first repository's results fails
- **THEN** `Failed to format results for <uri>: <error>` is printed on
  standard error
- **AND** the first repository is not reported as skipped
- **AND** the second repository is still scanned and formatted
- **AND** the system exits with status 1 after both repositories finish

#### Scenario: Low score

- **WHEN** every check succeeds but scores 0
- **THEN** the system exits with status 0

#### Scenario: Inconclusive score

- **WHEN** execution succeeds and all check scores are inconclusive
- **THEN** the system exits with status 0

#### Scenario: Multiple repository failures

- **GIVEN** three repositories selected through `--repos`
- **WHEN** the first cannot be parsed and the second cannot be scanned
- **THEN** the third repository is still scanned and formatted
- **AND** the final error includes both failed repositories and their causes
- **AND** the system exits with status 1

#### Scenario: Formatting and check runtime errors together

- **GIVEN** a repository has completed checks with a runtime error
- **WHEN** formatting that repository's results also fails
- **THEN** the system reports both the check and formatting diagnostics
- **AND** continues processing the remaining repositories
- **AND** the final error includes the formatting failure and
  "one or more checks failed during execution"
- **AND** the system exits with status 1

#### Scenario: A repository in an organization scan fails

- **GIVEN** `--org` successfully lists multiple non-archived repositories
- **WHEN** one of those repositories cannot be scanned
- **THEN** the system reports its scan failure and processes the remaining
  repositories
- **AND** the system exits with status 1 after the organization scan finishes
