# Proposal: Serve Scorecard results to MCP clients from `scorecard mcp`

## Why

Users have asked for Model Context Protocol (MCP) support so that LLM
clients can read Scorecard results (`ossf/scorecard#4635`). That
discussion preferred local integration with the Scorecard binary over a
hosted service, to avoid new infrastructure.

A standalone MCP server for Scorecard already exists in
`uwu-tools/scorecard-mcp`, with OpenSpec specs for its three capabilities.
This change moves that server into Scorecard as a subcommand and keeps
its history and specs. The server then ships and versions with the CLI
and reads the check documentation from the same module.

## What Changes

- Add a `scorecard mcp` subcommand that runs the MCP server over stdio.
- `scorecard mcp` accepts `--base-url`, which selects the Scorecard REST
  API that results are read from. `SCORECARD_MCP_BASE_URL` supplies the
  default; otherwise the public `https://api.scorecard.dev` is used.
- The server reads published results only. Live in-process scanning is
  out of scope; the provider abstraction leaves room for it.
- The check catalog is built from the same module's `docs/checks`, so it
  always matches the binary.
- The server advertises the `scorecard` build version.
- Each result reports which catalog checks it lacks, instead of a fixed
  list, and its caveats describe both sources the REST API serves:
  results projects publish and results from the weekly public scan.
- Errors for missing results and missing checks point to
  `scorecard --repo`, which computes results locally.
- `compare_repos` returns the same provenance, caveats, and attribution
  as the other result tools.
- Catalog documentation URLs reference the build commit, and REST API
  requests carry a `scorecard/<version>` User-Agent.

## Capabilities

### New Capabilities

None. `mcp-server`, `scorecard-results`, and `check-catalog` arrive with
the imported history, as archived by
`changes/archive/2026-08-04-add-scorecard-mcp-server/`.

### Modified Capabilities

- `openspec/specs/cli`: the Command Tree requirement adds `mcp` and its
  `--base-url` flag.
- `openspec/specs/mcp-server`: the server is started by `scorecard mcp`.
- `openspec/specs/scorecard-results`: adds the REST API base URL
  selection and client identification; computes each result's coverage
  from the catalog; points missing-result errors to `scorecard --repo`;
  adds full metadata to comparisons.
- `openspec/specs/check-catalog`: the catalog comes from the module the
  binary is built from, not a vendored dependency, and its documentation
  URLs reference the build commit.

## Impact

- `cmd/mcp.go`: the cobra subcommand, registered beside `serve`.
- `cmd/internal/mcp/`: the imported server, provider, catalog, model, and
  repository-reference packages.
- `go.mod`: adds `github.com/modelcontextprotocol/go-sdk` and
  `github.com/google/jsonschema-go`.
- `docs/mcp.md`: install and client configuration for `scorecard mcp`.
- No changes to scan flags, scoring, output formats, or `scorecard serve`.
