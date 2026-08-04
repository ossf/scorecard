# scorecard-results Specification

## Purpose

TBD - created by archiving change add-scorecard-mcp-server. Update Purpose after archive.

## Requirements

### Requirement: Retrieve a repository's Scorecard results

The server SHALL provide a tool that returns a repository's aggregate score and a
per-check summary (check name, score, and short reason) for the requested repository,
returning a compact summary by default.

#### Scenario: Aggregate score and per-check summary

- **WHEN** a client calls the repository-score tool with a valid repository reference
- **THEN** the server SHALL return the aggregate score and a per-check summary produced by a result provider

#### Scenario: Repository has no published results

- **WHEN** the cached provider has no results for the repository
- **THEN** the server SHALL return an MCP tool error indicating the repository may not have opted into publishing
  results and that live scanning will be available in a future version

### Requirement: Retrieve a single check's result

The server SHALL provide a tool that returns the full detail (score, reason, and
details) for a single named check on a repository.

#### Scenario: Existing check

- **WHEN** a client requests a specific check for a repository
- **THEN** the server SHALL return that check's score, reason, and details

#### Scenario: Unknown check name

- **WHEN** the requested check name is not a known Scorecard check
- **THEN** the server SHALL return an error explaining how to discover valid check names

### Requirement: Compare repositories

The server SHALL provide a tool that returns the aggregate scores of multiple
repositories for comparison and SHALL indicate when the number of requested
repositories exceeds a supported maximum.

#### Scenario: Compare multiple repositories

- **WHEN** a client provides several repository references
- **THEN** the server SHALL return each repository's aggregate score and provenance

#### Scenario: Requested list exceeds the maximum

- **WHEN** the number of repositories exceeds the supported maximum
- **THEN** the server SHALL return results up to the cap and note that the list was truncated

### Requirement: Swappable result provider

Results SHALL be produced through a result-provider abstraction, with a cached REST
provider used in this change, and every result SHALL declare which source produced it.

#### Scenario: Cached REST source is declared

- **WHEN** a result is produced by the cached REST provider
- **THEN** the result SHALL declare its source as the cached provider

### Requirement: Provenance metadata

Every result SHALL include the repository reference, the resolved commit SHA, the
scan or generation date, and the Scorecard version that produced it.

#### Scenario: Provenance present on every result

- **WHEN** any result is returned
- **THEN** it SHALL include the repository reference, resolved commit SHA, scan date, and Scorecard version

### Requirement: Commit pinning

Retrieval tools SHALL accept an optional commit argument and, when it is provided,
SHALL request results for that specific commit; when it is omitted, SHALL request the
latest available results.

#### Scenario: Commit specified

- **WHEN** a client supplies a commit SHA
- **THEN** the server SHALL request results pinned to that commit

#### Scenario: Commit omitted

- **WHEN** no commit is supplied
- **THEN** the server SHALL request the latest available results

### Requirement: Cached-data caveats and fidelity

Results from the cached provider SHALL include caveats stating that coverage is
limited to opted-in projects and that the weekly public scan omits the CI-Tests,
Contributors, and Dependency-Update-Tool checks, and every result SHALL declare its
completeness or fidelity.

#### Scenario: Cached caveats attached

- **WHEN** a result is produced by the cached provider
- **THEN** the result SHALL include the opted-in-coverage and omitted-checks caveats and declare its fidelity

### Requirement: Score semantics preserved

The server SHALL surface Scorecard scores unchanged, SHALL represent a score of -1 as
inconclusive rather than as a failing score, and SHALL NOT recompute the aggregate
score.

#### Scenario: Inconclusive check score

- **WHEN** a check has a score of -1
- **THEN** the server SHALL present it as inconclusive rather than as a zero or failing score

### Requirement: Data-license attribution

Every cached result SHALL include attribution stating that the data is provided under
CDLA Permissive 2.0, together with the source URL.

#### Scenario: Attribution present

- **WHEN** a cached result is returned
- **THEN** it SHALL include the CDLA Permissive 2.0 data-license attribution and the source URL
