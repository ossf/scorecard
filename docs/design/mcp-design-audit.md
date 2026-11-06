# scorecard-mcp — MCP Design Audit

- **Date:** 2026-07-31
- **Scope:** the server as merged to `main` (`b00e197`) — tools, resources,
  transport, error handling, and safety.
- **Method:** the `mcp-server-dev` skill's rubric (`tool-design`,
  `server-capabilities`, `resources-and-prompts`) and the Anthropic MCP
  Directory pass/fail review criteria, verified three ways: reading the Go
  source *and* the `modelcontextprotocol/go-sdk` v1.7.0 source (to confirm SDK
  behavior rather than trusting docs), live probes over real stdio (raw
  JSON-RPC and a `CommandTransport` client), and calls against the live
  `api.scorecard.dev` API.

This document is the single source of truth for the audit.

## Verdict

**Built correctly.** The server follows the recommended MCP shape end to end:
one-tool-per-action design, Directory-conformant annotations, structured output
with a text fallback, responsible-AI framing in the server `instructions` (not
in tool descriptions), resources exposed as an index plus a URI template, and
tool-level errors that carry recovery hints. The gaps below are refinements and
one explicit decision — none require structural rework.

## Rubric conformance

| Criterion (source) | Status | Evidence |
| --- | --- | --- |
| Deployment = local stdio, transport-agnostic factory | Pass (see D1) | `server.New()` builds transport-free; entrypoint binds stdio |
| Tool pattern A (fewer than ~15 actions) | Pass | 5 tools — the "one tool per action" sweet spot, not search+execute |
| Directory: `readOnlyHint` + `destructiveHint` + `title` on every tool | Pass | `readOnlyAnnotations` (`internal/server/server.go`); verified via `tools/list` |
| Directory: tool names <= 64 chars | Pass | longest is `get_check_result` (16) |
| Directory: read/write split | Pass | no write tools exist |
| Directory: no behavioral instructions in descriptions | Pass | descriptions disambiguate siblings — the recommended pattern, not imperative text |
| Directory: reference the upstream API in descriptions | Pass | each references the Scorecard REST API / docs |
| Structured output + text fallback | Pass | typed output; the SDK auto-populates the `TextContent` fallback when `Content` is unset |
| Errors as tool errors with actionable hints | Pass (see F1) | SDK wraps handler errors into `IsError`; sentinel errors name the fix |
| `instructions` used for cross-cutting framing | Pass | responsible-AI framing lives here, not in tool text |
| Resources: index + RFC 6570 template, offline | Pass | `scorecard://checks` + `scorecard://checks/{name}` via `docs/checks` |
| Response-size discipline | Pass (see F7) | `compare_repos` caps at 10 and self-reports truncation |
| No stdout pollution | Pass | `forbidigo` bans `fmt.Print*`; stdout is the protocol channel |
| `openWorldHint` scoped per tool | Exceeds | `true` for the 3 network tools, `false` for the 2 offline catalog tools |
| Prompts / sampling / roots / elicitation | Correctly omitted | documented non-goals |
| Progress / cancellation | N/A now | cached reads are fast; `ctx` propagates; planned for the live provider |

## Findings

### F1 — Medium: `get_check_result` conflates "unknown check" with "known-but-absent check"

The tool validates the requested check only against the *fetched result*, never
against the catalog it already ships. Two very different situations therefore
produce an identical, misleading error. Observed via live probes against
`ossf/scorecard`:

```text
get_check_result "SBOM"          -> isError: check not found in results: "SBOM" ... use list_checks ...
get_check_result "Totally-Bogus" -> isError: check not found in results: "Totally-Bogus" ... use list_checks ...
```

`SBOM` is a valid, *experimental* check that `list_checks` and `explain_check`
describe, so directing the user to `list_checks` is unhelpful — it is already
listed there. The same confusion affects `CI-Tests`, `Contributors`, and
`Dependency-Update-Tool` for repositories covered only by the weekly public
scan (they were *not* absent for `ossf/scorecard`, which self-publishes all
checks via the Scorecard Action).

**Fix:** cross-reference the catalog inside `get_check_result` and return three
distinct outcomes: (1) unknown name → "not a Scorecard check; see
`list_checks`"; (2) known but experimental / omitted from the cached scan →
"valid check, but not present in these cached results; live scanning will cover
it"; (3) present → return the result. The live provider (F-live below) resolves
outcome (2) at the source.

### F2 — Medium: no config surface for `BaseURL` (and a dead-code consequence)

`cmd/scorecard-mcp/main.go` hardcodes `provider.NewCachedREST("")`, so nothing
overrides the REST base URL. Two consequences:

1. The trailing-slash normalization in `NewCachedREST`
   (`strings.TrimSuffix(baseURL, "/")`, `internal/provider/rest.go`) is
   currently **unreachable in the running binary** — only unit tests exercise
   it, since the sole caller always passes `""`.
2. There is no way to point at a staging or private mirror of the Scorecard API
   without a code change and rebuild.

**Fix:** add a `-base-url` flag and/or a `SCORECARD_MCP_BASE_URL` env var in
`main.go`.

### F3 — Low: no stdio-subprocess integration test

Only the in-memory transport is exercised (`internal/server/server_test.go`),
so a regression in the stdio wiring itself (`cmd/scorecard-mcp/main.go`) would
not be caught. A `CommandTransport`-based test (spawn the built binary, connect
a real MCP client over real stdio) was manually verified during the audit and
would be cheap to keep as a permanent regression test.

### F4 — Low: `experimentalChecks` map can go stale

`internal/catalog/catalog.go`'s `experimentalChecks` map (`Webhooks`, `SBOM`)
is hand-maintained against `ossf/scorecard`'s `checks/all_checks.go` and will
silently drift on a `scorecard` dependency upgrade that adds or promotes a
check. Not an MCP-protocol issue, but a real correctness risk for catalog
output (and it feeds F1). Revisit on every `scorecard` bump, or derive it from
the vendored data instead of hand-maintaining it.

### F5 — Low: input constraints live in code, not in the JSON schema

Commit format and the `compare_repos` cap are enforced in handlers, but the
declared `inputSchema` does not express them — no `pattern` on `commit`, no
`maxItems` on `repos`. Expressing these in the schema lets the host and model
validate before a round-trip and self-correct.

### F6 — Low: `compare_repos` fetches sequentially

Up to ten serial HTTP calls means up to 10x latency. Cancellation is already
respected (`ctx` threads through each call); a bounded concurrent fetch would
cut wall-clock time materially.

### F7 — Low: `get_check_result` payload can be large

It returns the full `details[]`, which for checks such as `Pinned-Dependencies`
can be long. `get_repo_score` is correctly compact; consider a `details:false`
option or a truncation note here.

## Decision needed

### D1 — Deployment model / Anthropic Directory-submission fit

Local **stdio is the correct choice** for the stated endgame — an in-tree
`scorecard mcp` subcommand of a CLI security tool developers already install
(the same distribution shape as a language server or a `kubectl` plugin), with
no OAuth since the cached API is public. The skill's default bias toward remote
HTTP assumes a hosted-service model this project does not have.

The one place it matters: **Anthropic Directory submission** generally expects
remote HTTP or an MCPB bundle, not bare stdio. Decide explicitly whether
Directory submission is a goal, rather than discovering the mismatch later.

## Informational / optional

- **Version:** `Version` is still `dev`; wire it via `-ldflags` at release time.
- **Prompts:** none registered (a documented non-goal). A
  `/scorecard-summary <repo>` or repo-comparison prompt would be low-code,
  high-leverage if a common multi-step pattern emerges.
- **`goheader`:** enabled in `.golangci.yml` but a no-op (no template
  configured), so Apache license headers are not actually enforced.

## What is exemplary

- **Annotation hygiene:** per-tool `openWorldHint` (network vs offline) — a
  subtlety many servers miss.
- **Framing placement:** responsible-AI guidance is in `instructions` and the
  result `caveats[]`, never as imperative text in descriptions — the
  Directory-safe split, with a code comment stating the reasoning.
- **Provenance on every result:** resolved commit SHA, scan date, Scorecard
  version, `source`, plus CDLA attribution — results are cacheable and
  reproducible (the darnit contract).
- **Provider seam:** `ResultProvider` keeps "live later" an addition, not a
  rewrite.
- **Determinism guardrail:** the `forbidigo` ban on `fmt.Print*` protects the
  stdio channel — a correctness property, not just style.

## Recommended order

1. **F1** — fast, user-facing correctness fix.
2. **F2** — add the `BaseURL` config surface (also makes the trailing-slash fix
   reachable).
3. **Live provider (`LocalRunProvider`)** — the deferred capability that also
   dissolves F1/F4 at the source (full check set, private repos).
4. **F3–F7** and the informational items — polish, folded into the release and
   hygiene work.
5. **D1** — make the Directory-submission call explicitly.

*(Superseded by [Resolutions](#resolutions) below, which reflects the order and
decisions actually agreed on.)*

## Resolutions

**Date:** 2026-08-01. Every finding and the D1 decision were discussed and
resolved one at a time, in the order below — revised from "Recommended order"
above after closer review of the code behind each finding. This section is
authoritative going forward.

1. **D1 — resolved.** stdio + in-tree `scorecard mcp` subcommand is the
   approach for the foreseeable future. Anthropic Directory submission is
   explicitly *not* a current goal (Directory expects remote HTTP or an MCPB
   bundle, not bare stdio) but is not permanently ruled out.
2. **F1 — resolved.** `get_check_result` will return four distinct outcomes
   instead of one generic "not found": (1) unknown check name, (2) known but
   experimental, (3) known but omitted from this provider's scan (e.g.
   `CachedRESTProvider.Capabilities().OmittedChecks`), (4) present. Requires
   wiring `*catalog.Catalog` into the handler alongside `provider.Provider`.
3. **F4 — resolved, promoted ahead of F2/F3** since it is load-bearing for
   F1's case (2). Replace the hand-maintained `experimentalChecks` map with a
   computed diff of `checks.GetAll()` vs. `checks.GetAllWithExperimental()`
   (scorecard v5.5.0 `checks/all_checks.go`), evaluated once in
   `catalog.New()`. Accepts a new import of
   `github.com/ossf/scorecard/v5/checks` (same module/version already
   depended on via `docs/checks`).
4. **F2 + F3 — merged into one unit; explicit design review required before
   implementation.** F3's `CommandTransport` stdio integration test will be
   written against F2's new `-base-url`/`SCORECARD_MCP_BASE_URL` flag
   (pointed at an `httptest.Server`) rather than live `api.scorecard.dev`.
   Open questions for that review: `go build`-per-test vs. a `TestMain` that
   builds once; test location (`cmd/scorecard-mcp/` vs. a tagged
   `internal/server/integration_test.go`).
5. **F7 — resolved.** Rejected a `details:false` input (it fights the tool's
   purpose). Instead: a new `max_details` input (default 50, clamped 1–500),
   and a new `checkDetail` output type (not the shared `model.Check`)
   carrying `details_total`/`details_truncated` alongside a capped
   `details[]`. Truncation is applied only in the `get_check_result` handler,
   not the provider — `Provider.GetResult` keeps returning the fullest result
   it can, which scopes the truncation knob to the one tool that needs it and
   preserves darnit's full-JSON contract. Verified against live data before
   picking 50: `Pinned-Dependencies` details ranged 5–186 across sampled repos
   (`ossf/scorecard`, `kubernetes/kubernetes`, `tensorflow/tensorflow`), and
   `Token-Permissions` hit 377 on `envoyproxy/envoy` — confirming truncation
   must be generic across any check, not specific to one.
6. **F5 — resolved.** Add `jsonschema` `pattern` (commit, 40-hex) and
   `maxItems` (`compare_repos` `repos`, 10) as client-side, fail-fast hints
   only. All existing handler-side validation (`scorecardref.ValidateCommit`,
   the `maxCompareRepos` truncation path) stays as the authoritative
   enforcement, since MCP schemas are advisory and not every client validates
   before calling.
7. **F6 — resolved.** Bounded concurrent fetch in `compare_repos` via
   `golang.org/x/sync/errgroup` with `SetLimit(5)`, preserving input order in
   `results[]`.
8. **Live provider (`LocalRunProvider`) — explicitly deferred** to its own
   follow-up design conversation (auth-token delivery, sync-vs-async
   invocation with progress/cancellation, tool-surface impact).
9. **Informational items — backlogged as-is:** `Version` via `-ldflags` at
   release time; a `/scorecard-summary` prompt only if a usage pattern
   emerges; `goheader` fixed opportunistically next time `.golangci.yml` is
   touched.
