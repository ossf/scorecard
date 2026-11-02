## ADDED Requirements

### Requirement: REST API base URL

The cached REST provider SHALL read results from the base URL given by
`scorecard mcp --base-url`. When the flag is not set, it SHALL use the
`SCORECARD_MCP_BASE_URL` environment variable, and when that is empty, it
SHALL use `https://api.scorecard.dev`.

#### Scenario: Flag sets the base URL

- **WHEN** a user runs `scorecard mcp --base-url http://localhost:8080`
- **THEN** result requests SHALL go to `http://localhost:8080`

#### Scenario: Environment variable sets the default

- **WHEN** `SCORECARD_MCP_BASE_URL` is `http://localhost:8080` and
  `--base-url` is not set
- **THEN** result requests SHALL go to `http://localhost:8080`

#### Scenario: Public API by default

- **WHEN** neither `--base-url` nor `SCORECARD_MCP_BASE_URL` is set
- **THEN** result requests SHALL go to `https://api.scorecard.dev`
