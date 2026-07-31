## ADDED Requirements

### Requirement: List available checks
The server SHALL provide a tool that lists the available Scorecard checks, returning
for each check its name, short description, risk level, tags, supported platforms, and
documentation URL, sourced from Scorecard's own check documentation without requiring
a network call or a scan.

#### Scenario: List checks
- **WHEN** a client calls the list-checks tool
- **THEN** the server SHALL return the set of checks, each with name, short description, risk, tags, supported platforms, and documentation URL

#### Scenario: Experimental checks are indicated
- **WHEN** the catalog includes an experimental check
- **THEN** the server SHALL indicate that the check is experimental

### Requirement: Explain a single check
The server SHALL provide a tool that returns detailed documentation for a single named
check, including its description, risk level, tags, remediation guidance, and
documentation URL.

#### Scenario: Explain an existing check
- **WHEN** a client requests an explanation of a valid check name
- **THEN** the server SHALL return the check's description, risk, tags, remediation guidance, and documentation URL

#### Scenario: Unknown check name
- **WHEN** the requested check name is not recognized
- **THEN** the server SHALL return an error indicating the name is unknown and how to list valid checks

### Requirement: Check documentation resources
The server SHALL expose check documentation as read-only MCP resources: an index
resource listing all checks and a per-check resource addressed by a URI template.

#### Scenario: Read the index resource
- **WHEN** a client reads the checks index resource
- **THEN** the server SHALL return the list of checks with their identifiers

#### Scenario: Read a per-check resource
- **WHEN** a client reads the resource for a specific check by name
- **THEN** the server SHALL return that check's documentation

### Requirement: Catalog consistency with the vendored Scorecard version
The check catalog and its documentation SHALL be sourced from the Scorecard library
version the server is built against, so that catalog content is consistent with that
version.

#### Scenario: Catalog matches the vendored version
- **WHEN** the catalog is served
- **THEN** its contents SHALL correspond to the Scorecard version the server was built with
