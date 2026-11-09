# Proposal: Scorecard MCP Server

## Why

OpenSSF Scorecard produces rich, structured security-posture data for open source
repositories, but today an AI agent can only get at it by scraping the site,
hand-rolling REST calls, or shelling out to the CLI and parsing output. There is
no first-class, agent-native way to ask "what are this repo's security signals?"
and get back structured, provenance-tagged data. An MCP server closes that gap —
and doing it in Go, importing the Scorecard library directly, sets up a clean path
to upstreaming it as an in-tree `scorecard mcp` subcommand alongside the existing
`scorecard serve`.

## What Changes

- Introduce a new Go MCP server (via the official `modelcontextprotocol/go-sdk`)
  that exposes OpenSSF Scorecard data to MCP clients over **stdio**.
- Source results through a **`ResultProvider` seam** so the data backend is
  swappable. This change ships **one** provider: `CachedRESTProvider`, reading
  pre-computed public results from `api.scorecard.dev`.
- **Live/in-process scanning (`LocalRunProvider`, via `pkg/scorecard.Run`) is
  specced as part of the provider seam in this change but its implementation is
  deferred to a later change** ("cached now, live later"). No live scans ship here.
- Provide a **decomposed tool surface** rather than one monolithic tool:
  retrieve a repo's aggregate score, retrieve a single check's result, compare
  repos, list checks, and explain a check.
- Expose **read-only resources** for Scorecard check documentation.
- Every result carries **provenance metadata** (resolved commit SHA, scan/
  generation date, Scorecard version, and which provider/source produced it) so
  downstream consumers can cache and reproduce it — notably darnit
  (issue <https://github.com/darnitdevorg/darnit/issues/194>),
  which reuses Scorecard JSON across many OpenSSF Baseline controls. Retrieval
  tools accept an optional `commit` argument to fetch immutable results for a
  specific commit (supported by both the REST API and the library).
- Apply **responsible-AI framing**: results are surfaced as signals with explicit
  cached-data caveats (opted-in projects only; the weekly scan omits `CI-Tests`,
  `Contributors`, `Dependency-Update-Tool`); the server never asserts a repo
  "is secure/insecure." Framing is derived from Scorecard's own documented
  non-goals (heuristics with false positives/negatives; aggregate scores say
  nothing about individual behaviors; not a guarantee of security or regulatory
  compliance).
- Attach **licensing/attribution**: cached REST data is licensed CDLA Permissive
  2.0 and Scorecard code is Apache-2.0; results surface a data-license/attribution
  note so downstream redistribution stays compliant.

## Capabilities

### New Capabilities

- `mcp-server`: MCP server runtime — protocol lifecycle over stdio (Streamable
  HTTP left to a later change), tool and resource registration, parsing of
  `platform/owner/repo` package references, uniform structured-JSON responses,
  consistent error handling, and the responsible-AI framing applied to all output.
- `scorecard-results`: retrieve a repository's Scorecard results (aggregate score
  and per-check detail) and compare results across repositories, sourced through
  the swappable `ResultProvider` (cached REST now; live in-process deferred), with
  provenance metadata (commit SHA, scan date, source) and explicit cached-data
  caveats attached to every response. Response fidelity is provider-dependent
  (probe findings and maintainer annotations are available only from the live
  provider); every result declares its `source` and completeness.
- `check-catalog`: list the available Scorecard checks and explain an individual
  check (purpose, scoring rationale, risk, remediation), exposed both as tools
  and as read-only documentation resources.

### Modified Capabilities

<!-- None — this is a greenfield change; no existing specs are modified. -->

## Impact

- **New code:** a new Go module (this repo) with the MCP server, the
  `ResultProvider` interface + `CachedRESTProvider`, a REST client for
  `api.scorecard.dev`, and tool/resource handlers. Packaged so the server core
  drops into `github.com/ossf/scorecard`'s `cmd/` as a subcommand later.
- **Dependencies:** `github.com/modelcontextprotocol/go-sdk`;
  `github.com/ossf/scorecard/v5` (for shared result types now, `pkg/scorecard.Run`
  later); standard-library `net/http`. MCP-SDK deps are kept under `cmd/` to limit
  the module's dependency surface for downstream library importers.
- **External systems:** read-only calls to `api.scorecard.dev` (data licensed
  CDLA Permissive 2.0; attribution surfaced in results). No credentials required
  for the cached path; token auth (`GITHUB_AUTH_TOKEN`/`GH_TOKEN`, GitHub App, or
  `GITLAB_AUTH_TOKEN`) becomes relevant only when the deferred live provider lands
  — handled automatically by the Scorecard library's roundtripper, so no custom
  auth code is required.
- **Consumers:** MCP clients (Claude Desktop/Code, etc.) locally via stdio;
  darnit as a JSON consumer once its adapter
  (<https://github.com/darnitdevorg/darnit/issues/194>) is wired.
- **Compatibility:** greenfield — no breaking changes.

## Non-goals

- **Live/in-process scanning.** The `LocalRunProvider` implementation, private-repo
  scans, the full check set, and `GITHUB_AUTH_TOKEN` handling are deferred to a
  later change; only the provider seam is specced now.
- **Publishing/writing results.** The REST `postResult` (OIDC) publish path is out
  of scope; this server is read-only.
- **Remote/hosted deployment and auth.** stdio only in this change; Streamable
  HTTP transport and any auth are a later change.
- **Reimplementing scoring or checks.** We consume Scorecard's results and library;
  we do not reinvent check logic or the aggregate score.
- **MCP apps / UI widgets / elicitation.** Plain structured results only.
- **MCP Prompts (slash-command workflows).** Deferred; this change ships tools +
  resources only. Add later if a repeated workflow emerges.
- **Distribution packaging.** MCPB bundling and MCP Registry publication are out of
  scope here.
- **darnit-side integration code.** We guarantee a stable JSON contract; wiring it
  into darnit's `ExecutionContext` is the scope of darnit issue
  <https://github.com/darnitdevorg/darnit/issues/194>, not this change.
