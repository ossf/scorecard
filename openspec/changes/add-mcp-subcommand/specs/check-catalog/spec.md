## REMOVED Requirements

### Requirement: Catalog consistency with the vendored Scorecard version

**Reason**: The server is now built inside the Scorecard module, so there
is no separately vendored Scorecard version to stay consistent with.

**Migration**: Replaced by "Catalog consistency with the built Scorecard
version".

## ADDED Requirements

### Requirement: Catalog consistency with the built Scorecard version

The check catalog and its documentation SHALL be sourced from the
`docs/checks` package of the module the `scorecard` binary is built from,
so that catalog content matches the checks that binary runs.

#### Scenario: Catalog matches the built version

- **WHEN** the catalog is served
- **THEN** its contents SHALL correspond to the check documentation
  compiled into the `scorecard` binary
