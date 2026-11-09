# Design: Scorecard MCP Server

## Context

OpenSSF Scorecard evaluates a repository's security posture and emits structured
results (an aggregate score plus per-check scores, reasons, and details). Today an
AI agent can only reach that data by scraping the site, hand-rolling REST calls, or
shelling out to the CLI and parsing output. This change adds an **MCP server** — a
process that speaks the Model Context Protocol so MCP clients (Claude Desktop/Code,
etc.) can call well-typed _tools_ and read _resources_ instead.

Key facts established during discovery (both from the local `ossf/scorecard` and
`scorecard-webapp` clones and from `darnit`):

- **`pkg/scorecard.Run` is importable out-of-tree.** Go's `internal/` visibility
  rule only blocks an external module from importing Scorecard's `internal/…`
  packages _directly_; the public `pkg/scorecard` (which itself uses those internal
  packages) compiles fine when we import it. So live scanning does not _require_
  living in-tree — in-tree is the endgame for ecosystem/distribution reasons.
- **The check catalog is available offline.** `github.com/ossf/scorecard/v5/docs/checks`
  is a public package (`checks.Read()` → per-check name, risk, short, description,
  remediation, tags, supported platforms, doc URL). No network or scan is needed to
  list or explain checks.
- **The cached REST result is a subset of the CLI result.** REST returns score +
  checks + metadata; the library additionally exposes probe findings and maintainer
  annotations. Response fidelity therefore depends on which provider produced it.
- **Score `-1` means inconclusive**, not "zero"; the aggregate can be `-1` too.
- **Auth for live scans is handled by the library's roundtripper** (reads
  `GITHUB_AUTH_TOKEN`/`GH_TOKEN`, GitHub App, `GITLAB_AUTH_TOKEN`; waits out rate
  limits) — no custom auth code needed later.
- **Constraints:** OSSF conventions (Apache-2.0, DCO), Scorecard's minimal-dependency
  policy, the Anthropic MCP Directory review criteria (we conform now), and the CDLA
  Permissive 2.0 license on REST data.

## Goals / Non-Goals

**Goals:**

- Expose Scorecard signals to MCP clients as well-typed, structured tools + resources.
- Make the data backend swappable via a `ResultProvider` seam (cached REST now,
  in-process live scans later) with no rewrite.
- Emit deterministic, provenance-tagged JSON so downstream consumers (darnit) can
  cache and reproduce results.
- Keep the design upstreamable as an in-tree `scorecard mcp` subcommand.
- Conform to MCP Directory review criteria and responsible-AI framing from day one.

**Non-Goals** (see proposal for the full list): live/in-process scanning
implementation, MCP Prompts, hosted HTTP transport, publishing results, MCPB/registry
packaging, reimplementing Scorecard's scoring, and darnit-side adapter code.

## Decisions

### D1 — Language & SDK: Go + official `go-sdk`

Go matches the Scorecard ecosystem and lets us import `pkg/scorecard` directly.
Use `github.com/modelcontextprotocol/go-sdk`. _Alternatives:_ Python/FastMCP or the
TS SDK (rejected — a language mismatch with the in-tree endgame and a rewrite later).

### D2 — `ResultProvider` is the central extensibility seam

All tools depend on an interface, not a concrete backend. (An _interface_ in Go is a
contract: any type implementing these methods can be substituted.)

```go
type ResultProvider interface {
    // Result for a repo; commit "" means latest/HEAD.
    GetResult(ctx context.Context, repo RepoRef, commit string) (*Result, error)
    // Which checks/fidelity this provider can supply, for caveat reporting.
    Capabilities() ProviderCapabilities
}
```

- **`CachedRESTProvider`** (ships now): GETs `https://api.scorecard.dev/projects/
{platform}/{org}/{repo}` (optionally `?commit=`). Public + opted-in repos only.
- **`LocalRunProvider`** (specced, deferred): wraps `pkg/scorecard.Run` for any repo
  a token can access, all checks, fresh results.

Every `Result` records which provider/source produced it, so responses can carry
accurate caveats. _Alternative:_ hardcode the REST client (rejected — forces a rewrite
for "live later," the exact trap we're avoiding).

### D3 — Result data model: a provenance-rich superset of Scorecard JSON v2

Tools return **structured output** — MCP lets a tool declare an `outputSchema` (the
JSON shape it promises) and return `structuredContent` (the typed object) alongside a
`text` fallback (a JSON string, for hosts that don't yet read structured content). Our
schema mirrors Scorecard's canonical **JSON v2** and adds provenance/caveats:

```text
repo:        { platform, org, name }
commit:      resolved commit SHA (40-hex)
date:        scan/generation timestamp (RFC3339)
scorecard:   { version, commit }
source:      "cached-rest" | "live-local"
score:       number            # aggregate 0–10, or -1 = inconclusive
checks:      [ { name, score(0–10 or -1), reason, details[], documentation{short,url} } ]
annotations: [...]             # maintainer justifications, live provider only (optional)
caveats:     [ ... ]           # e.g. "cached: omits CI-Tests/Contributors/Dependency-Update-Tool"
attribution: { data_license: "CDLA-Permissive-2.0", source_url }
```

`checks[].score = -1` is surfaced explicitly as _inconclusive_. We never recompute
scores — we pass through what Scorecard reports.

### D4 — Determinism / caching contract (for darnit)

darnit's `ExecutionContext.get_or_run_tool(tool_key, run_func)` memoizes a tool's JSON
under a string key so N Baseline controls trigger exactly one run
(<https://github.com/darnitdevorg/darnit/issues/194>). To make our output a safe cache
entry and reproducible for attestation, **every result includes the resolved commit
SHA, scan date, and Scorecard version**, and retrieval tools accept an optional
`commit` argument for immutable historical results. Recommended cache key shape for
consumers: `scorecard:{platform}/{org}/{repo}@{commit}`. (darnit's `cache_key` is a
static per-run string today; our commit SHA lets them upgrade to per-repo+commit keys.
That change is darnit's, not ours.)

### D5 — Tool surface: five read-only tools, compact-by-default

| Tool               | Purpose                                            |
| ------------------ | -------------------------------------------------- |
| `get_repo_score`   | Aggregate score + per-check summaries for one repo |
| `get_check_result` | One check's full detail for one repo               |
| `compare_repos`    | Aggregate scores across several repos              |
| `list_checks`      | Catalog of checks (offline, from `docs/checks`)    |
| `explain_check`    | One check's methodology/risk/remediation (offline) |

- Each tool sets MCP **annotations** (host hints): `readOnlyHint: true`, `title`, and
  `openWorldHint: true` on the network-touching tools. All tools are read-only, so the
  Directory's read/write-split rule is satisfied by construction.
- **Descriptions are contracts** (say what it does, returns, and does _not_ do; point
  to sibling tools) but contain **no behavioral instructions** ("always call X") —
  those are treated as prompt injection at Directory review.
- `get_repo_score` is **compact by default** (score + per-check name/score/short
  reason); callers drill into `get_check_result`/`explain_check` for detail. Large
  outputs are truncated with an explicit note.

### D6 — Responsible-AI framing lives in `instructions` + payload, never in tool text

MCP servers can set an `instructions` string that the host places in the model's
system prompt. That is where cross-cutting framing goes: "Scorecard results are
heuristic signals, not a verdict; aggregate scores say nothing about individual
behaviors; cached data covers opted-in repos and omits CI-Tests/Contributors/
Dependency-Update-Tool." The same caveats ride in each result's `caveats[]`. Framing is
derived from Scorecard's own documented non-goals.

### D7 — Check catalog as resources (offline via `docs/checks`)

MCP **resources** are read-only data the host can pull into context (as opposed to
tools the model calls). Expose the catalog as `scorecard://checks` (index) and a
template `scorecard://checks/{name}` (one check), sourced from the importable
`docs/checks` package — no network, always consistent with the vendored Scorecard
version. `list_checks`/`explain_check` return the same data for model-initiated use.

### D8 — Transport structuring: one factory, stdio now, HTTP later

A single transport-agnostic `newServer()` registers all tools/resources and the
`instructions`. The entrypoint chooses the transport binding — stdio now, Streamable
HTTP behind a flag later. This is exactly the shape that drops into `ossf/scorecard`
as a subcommand (D11).

### D9 — Platform + reference parsing

Accept `platform/owner/repo` (platform optional, default `github.com`; `gitlab.com`
supported), mirroring the coverage of the cached REST API and the parsing of the
steiza reference. Azure DevOps and local dirs are possible once the live provider
lands.

### D10 — Licensing / attribution

Surface a data-license/attribution note (`CDLA-Permissive-2.0` + source URL) in
results and in the catalog resource. Repo code is Apache-2.0.

### D11 — Upstreaming path to an in-tree `scorecard mcp` subcommand

`cmd/serve.go` (the existing HTTP interface) is the template. The server core lives in
a package that the eventual `cmd/mcp.go` wires as `mcpCmd(o *options.Options)` +
`AddCommand(...)` in `cmd/root.go`, reusing Scorecard's `makeRepo()` routing and
`options.Validate()`. To honor Scorecard's minimal-dependency policy, MCP-SDK imports
stay isolated under `cmd/` (a build tag or a nested module is decided with maintainers
at upstream time). Match the Scorecard Go toolchain version to avoid CI friction.

### D12 — Auth for the deferred live provider

`LocalRunProvider` will rely entirely on the library's roundtripper (env-var tokens,
GitHub App, GitLab, automatic rate-limit handling). Caution for later: a single token
is not safe across concurrent scans.

### D13 — Development methodology (recorded so future sessions follow it)

This project is built spec-first with **OpenSpec** and with the **MCP dev Agent
Skills** (`build-mcp-server` and friends;
<https://modelcontextprotocol.io/docs/2026-07-28/develop/build-with-agent-skills>).
Future agents should install and use them (see `AGENTS.md` for the install fallback
when the plugin marketplace is blocked) rather than hand-rolling MCP structure.

## Risks / Trade-offs

- **Cached data is stale/incomplete** (opted-in repos only; omits three checks) →
  attach caveats to every result; support `commit` pinning; the deferred live provider
  fills the gap.
- **`go-sdk` is young and may churn** → pin the version, isolate its use behind
  `newServer()`, verify the exact API at implementation time.
- **Scorecard v5 API drift** → pin the module; keep providers a thin adapter over
  `pkg/scorecard` and the REST schema.
- **Minimal-dependency policy vs. adding the MCP SDK to core** → isolate under `cmd/`
  (build tag / nested module) and justify in the upstream PR.
- **Payload size vs. context budget** → compact-by-default + drill-down + truncation
  notes.
- **`-1` misread as a failing score** → represent and label it as _inconclusive_.
- **Concurrent live scans reusing one token** (later) → serialize or use a token pool.

## Migration Plan

Greenfield — no data migration or rollback needed. Phased rollout:

1. **This change:** `CachedRESTProvider`, stdio transport, five tools + catalog
   resources, structured output with provenance + caveats.
2. **Later:** `LocalRunProvider` (`pkg/scorecard.Run`) with progress + cancellation;
   private repos and the full check set.
3. **Later:** Streamable HTTP transport; upstream as the in-tree `scorecard mcp`
   subcommand.

## Open Questions

- Exact `go-sdk` API surface and version to pin (resolve at implementation).
- Dependency-isolation mechanism for upstreaming (build tag vs. nested module) —
  decide with Scorecard maintainers.
- Whether `explain_check` should also list a check's probe IDs in v1 (leaning no).
- Canonical REST host (`api.scorecard.dev` vs `api.securityscorecards.dev`) — default
  to `api.scorecard.dev`, make it configurable.
- `compare_repos` output shape and max repo count.
