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

### Requirement: REST API client identification

The cached REST provider SHALL identify itself to the REST API with a
`User-Agent` of `scorecard/<version>`, where `<version>` is the version
printed by `scorecard version`.

#### Scenario: Requests carry the scorecard version

- **WHEN** the cached REST provider requests results
- **THEN** the request SHALL carry the `User-Agent` `scorecard/<version>`

## MODIFIED Requirements

### Requirement: Retrieve a repository's Scorecard results

The server SHALL provide a tool that returns a repository's aggregate score and a
per-check summary (check name, score, and short reason) for the requested repository,
returning a compact summary by default.

#### Scenario: Aggregate score and per-check summary

- **WHEN** a client calls the repository-score tool with a valid repository reference
- **THEN** the server SHALL return the aggregate score and a per-check summary produced by a result provider

#### Scenario: Repository has no published results

- **WHEN** the cached provider has no results for the repository
- **THEN** the server SHALL return an MCP tool error stating that the
  repository has neither published results nor results from the weekly
  public scan, and that `scorecard --repo` computes results locally

### Requirement: Retrieve a single check's result

The server SHALL provide a tool that returns the full detail (score, reason, and
details) for a single named check on a repository.

#### Scenario: Existing check

- **WHEN** a client requests a specific check for a repository
- **THEN** the server SHALL return that check's score, reason, and details

#### Scenario: Unknown check name

- **WHEN** the requested check name is not a known Scorecard check
- **THEN** the server SHALL return an error explaining how to discover valid check names

#### Scenario: Check does not run on the platform

- **WHEN** the requested check is a known, non-experimental check that does
  not run on the repository's platform
- **THEN** the server SHALL return an error stating that the check does not
  run on that platform

#### Scenario: Known check missing from the result

- **WHEN** the requested check is a known, non-experimental check that the
  repository's result does not include
- **THEN** the server SHALL return an error stating that the result lacks
  the check and that `scorecard --repo <repo> --checks <check>` computes it
  locally

### Requirement: Compare repositories

The server SHALL provide a tool that returns the aggregate scores of multiple
repositories for comparison and SHALL indicate when the number of requested
repositories exceeds a supported maximum.

#### Scenario: Compare multiple repositories

- **WHEN** a client provides several repository references
- **THEN** the server SHALL return, for each repository, its aggregate
  score, provenance, caveats, completeness, and data-license attribution

#### Scenario: Requested list exceeds the maximum

- **WHEN** the number of repositories exceeds the supported maximum
- **THEN** the server SHALL return results up to the cap and note that the list was truncated

### Requirement: Cached-data caveats and fidelity

Results from the cached provider SHALL include caveats stating that the
REST API serves results that projects publish with `publish_results` and
results from the weekly public scan. Every result SHALL list the
non-experimental checks in the `scorecard` binary's catalog that run on
the repository's platform and that the result does not include, and SHALL
declare itself complete only when it lacks none.

#### Scenario: Cached caveats attached

- **WHEN** a result is produced by the cached provider
- **THEN** the result SHALL include the coverage caveats and declare its
  completeness

#### Scenario: Result lacks checks

- **WHEN** a result does not include one or more non-experimental catalog
  checks that run on the repository's platform
- **THEN** the result SHALL list those checks as missing, SHALL declare
  itself incomplete, and SHALL include a caveat naming them

#### Scenario: Result includes every check

- **WHEN** a result includes every non-experimental catalog check that
  runs on the repository's platform
- **THEN** the result SHALL declare itself complete and list no missing
  checks
