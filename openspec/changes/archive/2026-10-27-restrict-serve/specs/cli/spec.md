## MODIFIED Requirements

### Requirement: Command Tree

The system SHALL provide a root `scorecard` command that runs a scan, a
`serve` subcommand that starts the HTTP server, and a `version` subcommand
that prints build information. The scan flags SHALL be accepted only by the
root command; subcommands SHALL reject them as unknown flags. `serve` SHALL
accept `--listen`, whose behavior is defined in `http-server`.

#### Scenario: Root command runs a scan

- **WHEN** a user runs `scorecard --repo=github.com/owner/repo`
- **THEN** the system scans that repository and prints the results

#### Scenario: Scan flags are rejected by subcommands

- **WHEN** a user runs `scorecard serve --verbosity=debug`
- **THEN** the system fails with `unknown flag: --verbosity`

#### Scenario: Serve accepts listen

- **WHEN** a user runs `scorecard serve --listen 0.0.0.0`
- **THEN** the system starts the HTTP server on all interfaces
