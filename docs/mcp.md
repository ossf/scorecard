# `scorecard mcp`

`scorecard mcp` runs an [MCP](https://modelcontextprotocol.io) (Model Context
Protocol) server for Scorecard. It exposes Scorecard's
security-posture data to MCP clients (Claude Desktop/Code, VS Code, and others)
as well-typed, read-only tools and resources.

> **Status: experimental.** This server reads pre-computed results from the
> public Scorecard REST API (`api.scorecard.dev`). On-demand, in-process
> scanning is designed for but not yet implemented — see [Roadmap](#roadmap).

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

`scorecard mcp` is part of the `scorecard` binary. Install `scorecard` as
described in the [README](../README.md#installation).

The server speaks MCP over **stdio**; no credentials are required for the cached
REST provider.

## Configure your MCP client

> If `scorecard` is not on your `PATH`, use its absolute path in the configs
> below.

**Claude Code** — register globally with the CLI instead of hand-editing
JSON:

```sh
claude mcp add scorecard --scope user -- scorecard mcp
```

**Claude Desktop / Claude Code** (`.mcp.json` or the app's MCP config):

```json
{
  "mcpServers": {
    "scorecard": {
      "command": "scorecard",
      "args": ["mcp"]
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
      "command": "scorecard",
      "args": ["mcp"]
    }
  }
}
```

(Use an absolute path to the binary if it is not on your `PATH`.)

## Try it with the MCP Inspector

The [MCP Inspector](https://github.com/modelcontextprotocol/inspector) is the
quickest way to exercise the server directly, without wiring it into a client.
It needs [Node.js](https://nodejs.org/) (for `npx`); nothing to install ahead of
time.

Point it at `scorecard mcp`, using an absolute path to `scorecard` if it is
not on your `PATH`.

### Interactive UI

```sh
npx @modelcontextprotocol/inspector scorecard mcp
```

This launches a local web UI and prints a URL with a session token. Open it,
click **Connect**, then use the **Tools** and **Resources** tabs to browse and
invoke everything — for example, call `get_repo_score` with
`repo = github.com/ossf/scorecard`.

### Command line (scriptable)

The same Inspector has a `--cli` mode for non-interactive checks — handy for a
quick smoke test or CI:

```sh
# List the registered tools, resources, and resource templates
npx @modelcontextprotocol/inspector --cli scorecard mcp --method tools/list
npx @modelcontextprotocol/inspector --cli scorecard mcp --method resources/list
npx @modelcontextprotocol/inspector --cli scorecard mcp --method resources/templates/list

# Offline tools (no network)
npx @modelcontextprotocol/inspector --cli scorecard mcp \
  --method tools/call --tool-name list_checks
npx @modelcontextprotocol/inspector --cli scorecard mcp \
  --method tools/call --tool-name explain_check --tool-arg check=Branch-Protection

# Read a documentation resource
npx @modelcontextprotocol/inspector --cli scorecard mcp \
  --method resources/read --uri scorecard://checks

# Live query against api.scorecard.dev
npx @modelcontextprotocol/inspector --cli scorecard mcp \
  --method tools/call --tool-name get_repo_score --tool-arg repo=github.com/ossf/scorecard
```

Pass a tool's arguments with repeated `--tool-arg key=value` flags (for example,
add `--tool-arg commit=<sha>` to pin a query to a specific commit). An unknown
check name returns a tool error suggesting `list_checks`, which is a quick way to
confirm error handling works.

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

The server lives in [`cmd/mcp.go`](../cmd/mcp.go) and
[`cmd/internal/mcp/`](../cmd/internal/mcp/). Its specs are the `mcp-server`,
`scorecard-results`, and `check-catalog` capabilities in
[`openspec/specs/`](../openspec/specs/).

Point the server at another Scorecard REST API with `--base-url` or the
`SCORECARD_MCP_BASE_URL` environment variable.

## Roadmap

- **Live scanning** — a `LocalRunProvider` using `pkg/scorecard.Run` for any
  repository a token can access (private repos, all checks, fresh results),
  behind the same `ResultProvider` seam.
- **Streamable HTTP transport** for hosted/mixed environments.

## References

- [MCP design audit and resolutions](design/mcp-design-audit.md)
- [ossf/scorecard#4635](https://github.com/ossf/scorecard/issues/4635), the
  request for MCP support

## License

Apache 2.0 — see [`LICENSE`](../LICENSE).

Data from the REST API is licensed under
[CDLA Permissive 2.0](https://github.com/ossf/scorecard#scorecard-rest-api).
