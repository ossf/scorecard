# `scorecard-mcp`

[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/uwu-tools/scorecard-mcp/badge)](https://scorecard.dev/viewer/?uri=github.com/uwu-tools/scorecard-mcp)

An [MCP](https://modelcontextprotocol.io) (Model Context Protocol) server for
[OpenSSF Scorecard](https://github.com/ossf/scorecard). It exposes Scorecard's
security-posture data to MCP clients (Claude Desktop/Code, VS Code, and others)
as well-typed, read-only tools and resources.

> **Status: experimental.** This server currently reads pre-computed results from
> the public Scorecard REST API (`api.scorecard.dev`). On-demand, in-process
> scanning (via the Scorecard library) is designed for but not yet implemented —
> see [Roadmap](#roadmap). The intended endgame is to contribute this as an
> in-tree `scorecard mcp` subcommand in `ossf/scorecard`.

Results are **heuristic signals to inform a human decision, not a verdict.** The
server never asserts that a repository "is secure" or "is insecure."

## What it provides

**Tools**

| Tool | Description |
| --- | --- |
| `get_repo_score` | Aggregate score + per-check summary for a repository |
| `get_check_result` | Full result (score, reason, details) for one check on a repository |
| `compare_repos` | Aggregate scores across several repositories |
| `list_checks` | Catalog of Scorecard checks (offline) |
| `explain_check` | Methodology, risk, and remediation for one check (offline) |

**Resources**

- `scorecard://checks` — index of all checks
- `scorecard://checks/{name}` — documentation for a single check

Every result carries provenance (resolved commit SHA, scan date, Scorecard
version, and which provider produced it), cached-data caveats, and a
CDLA Permissive 2.0 data-license attribution.

## Install

Requires Go 1.25+.

```sh
go install github.com/uwu-tools/scorecard-mcp/cmd/scorecard-mcp@latest
```

Or build from source:

```sh
git clone https://github.com/uwu-tools/scorecard-mcp
cd scorecard-mcp
go build -o bin/scorecard-mcp ./cmd/scorecard-mcp
```

The server speaks MCP over **stdio**; no credentials are required for the cached
REST provider.

## Configure your MCP client

**Claude Desktop / Claude Code** (`.mcp.json` or the app's MCP config):

```json
{
  "mcpServers": {
    "scorecard": {
      "command": "scorecard-mcp"
    }
  }
}
```

**VS Code** (`.vscode/mcp.json`):

```json
{
  "servers": {
    "scorecard": {
      "type": "stdio",
      "command": "scorecard-mcp"
    }
  }
}
```

(Use an absolute path to the binary if it is not on your `PATH`.)

## Example

Asking a client "What's the OpenSSF Scorecard for ossf/scorecard?" calls
`get_repo_score` and returns structured content like:

```json
{
  "repo": { "platform": "github.com", "org": "ossf", "name": "scorecard" },
  "commit": "64febf8c5229...",
  "source": "cached-rest",
  "score": 8.7,
  "checks": [
    { "name": "Code-Review", "score": 10, "reason": "..." },
    { "name": "Fuzzing", "score": 10, "reason": "..." }
  ],
  "caveats": [
    "Cached results cover only projects that have opted in via publish_results: true.",
    "The weekly public scan omits the CI-Tests, Contributors, and Dependency-Update-Tool checks."
  ],
  "attribution": { "data_license": "CDLA-Permissive-2.0", "source_url": "https://api.scorecard.dev" }
}
```

## Caveats

The cached REST provider:

- covers only public repositories that have opted in via
  [`publish_results: true`](https://github.com/ossf/scorecard-action);
- omits the `CI-Tests`, `Contributors`, and `Dependency-Update-Tool` checks
  (excluded from the weekly public scan);
- reports a check or aggregate score of `-1` as **inconclusive** — not a
  failing score.

Data from the REST API is licensed under
[CDLA Permissive 2.0](https://cdla.dev/permissive-2-0).

## Development

This project is built spec-first with [OpenSpec](https://openspec.dev)
(`openspec/`) and with the MCP dev Agent Skills.

See [`AGENTS.md`](AGENTS.md) for the workflow and conventions.

```sh
go test ./...
golangci-lint run ./...
```

## Roadmap

- **Live scanning** — a `LocalRunProvider` using `pkg/scorecard.Run` for any
  repository a token can access (private repos, all checks, fresh results),
  behind the same `ResultProvider` seam.
- **Streamable HTTP transport** for hosted/mixed environments.
- **Upstreaming** as an in-tree `scorecard mcp` subcommand in `ossf/scorecard`.

## References

- [Initial implementation and design decisions](docs/init-impl.md)
- [MCP design audit and resolutions](docs/mcp-design-audit.md)

## Community

- [Code of Conduct](CODE_OF_CONDUCT.md)
- [Contributing](CONTRIBUTING.md)
- [Security Policy](SECURITY.md)
- [Support](SUPPORT.md)
- [Maintainers](MAINTAINERS.md)
- [Changelog](CHANGELOG.md)

## License

Apache 2.0 — see [`LICENSE`](LICENSE).
