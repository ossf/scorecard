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

### Tools

| Tool               | Description                                                        |
| ------------------ | ------------------------------------------------------------------ |
| `get_repo_score`   | Aggregate score + per-check summary for a repository               |
| `get_check_result` | Full result (score, reason, details) for one check on a repository |
| `compare_repos`    | Aggregate scores across several repositories                       |
| `list_checks`      | Catalog of Scorecard checks (offline)                              |
| `explain_check`    | Methodology, risk, and remediation for one check (offline)         |

### Resources

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

> **Troubleshooting:** if `go install` or `go build` fails with
> `compile: version "X" does not match go tool version "Y"` across many
> stdlib packages, you have a stray `GOROOT` environment variable pointing at
> a different Go toolchain than the one on `PATH` (common with multiple Go
> version managers, e.g. gimme). Run `env -u GOROOT go install ...` (or
> `env -u GOROOT go build ...`) instead, or unset `GOROOT` in your shell
> profile.

The server speaks MCP over **stdio**; no credentials are required for the cached
REST provider.

## Configure your MCP client

> `go install` places the binary in `$(go env GOPATH)/bin` (or `$GOBIN` if
> set), which is **not on `PATH` by default** on most systems. Either add
> that directory to `PATH`, or use its absolute path
> (`$(go env GOPATH)/bin/scorecard-mcp`) in the configs below.

**Claude Code** — register globally with the CLI instead of hand-editing
JSON:

```sh
claude mcp add scorecard --scope user -- "$(go env GOPATH)/bin/scorecard-mcp"
```

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
  "date": "2026-08-01T02:19:41Z",
  "scorecard": { "version": "v5.3.0", "commit": "c22063e786c1..." },
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
  "attribution": {
    "data_license": "CDLA-Permissive-2.0",
    "source_url": "https://api.scorecard.dev"
  },
  "complete": false
}
```

`commit` is the _target repository's_ resolved commit; `scorecard.commit` and
`scorecard.version` identify the build of the Scorecard tool that produced the
result — the two are unrelated and easy to confuse. `complete` reports whether
the provider ran the full check set (the cached REST provider always reports
`false`, since it omits three checks — see [Caveats](#caveats)).

## Caveats

The cached REST provider:

- covers only public repositories that have opted in via
  [`publish_results: true`](https://github.com/ossf/scorecard-action);
- omits the `CI-Tests`, `Contributors`, and `Dependency-Update-Tool` checks
  (excluded from the weekly public scan);
- reports a check or aggregate score of `-1` as **inconclusive** — not a
  failing score. `reason` on an inconclusive check is passed through verbatim
  from the upstream API and can be a raw internal-error string (e.g. a
  GitHub token/permissions failure inside Scorecard itself) rather than a
  descriptive explanation;
- passes `date` through as returned by `api.scorecard.dev`, which is not
  normalized to a single format (observed as both a full RFC 3339 timestamp
  and a bare `YYYY-MM-DD` date depending on the repository).

## Development

This project is built spec-first with [OpenSpec](https://openspec.dev)
(`openspec/`) and with the MCP dev Agent Skills.

See [`AGENTS.md`](AGENTS.md) for the workflow and conventions.

```sh
go test ./...
golangci-lint run ./...
```

(See the [Install](#install) troubleshooting note if these fail with a Go
toolchain version mismatch.)

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

Data from the REST API is licensed under
[CDLA Permissive 2.0](https://github.com/ossf/scorecard#scorecard-rest-api).
