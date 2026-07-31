## 1. Project scaffolding

- [ ] 1.1 Initialize the Go module, matching the Scorecard repo's Go toolchain version to avoid CI friction
- [ ] 1.2 Add dependencies: `github.com/modelcontextprotocol/go-sdk` and `github.com/ossf/scorecard/v5`
- [ ] 1.3 Lay out packages so MCP-SDK usage is isolated under `cmd/` (drop-in as `cmd/mcp` upstream): e.g. `internal/server`, `internal/provider`, `internal/scorecarddata`, `cmd/scorecard-mcp`
- [ ] 1.4 Add `LICENSE` (Apache-2.0) and a `.golangci.yml` aligned with the Scorecard project

## 2. Core types and the provider seam

- [ ] 2.1 Define `Result`, `RepoRef`, and `Check` types mirroring Scorecard JSON v2 plus provenance (`source`, resolved commit SHA, date, Scorecard version), `caveats`, `attribution`, and optional `annotations`
- [ ] 2.2 Define the `ResultProvider` interface (`GetResult`, `Capabilities`) and `ProviderCapabilities`
- [ ] 2.3 Define the tool `outputSchema`(s) matching the result types, returned as `structuredContent` with a JSON `text` fallback

## 3. Repository reference parsing

- [ ] 3.1 Parse `platform/owner/repo` (default `github.com`; accept `gitlab.com`; tolerate trailing slash; clear validation errors)
- [ ] 3.2 Enforce a platform enum and a 40-hex commit pattern; unit-test valid and invalid inputs

## 4. CachedRESTProvider

- [ ] 4.1 Implement the REST client: `GET https://api.scorecard.dev/projects/{platform}/{org}/{repo}` with optional `?commit=` (host configurable)
- [ ] 4.2 Map the REST response to `Result`; set `source=cached-rest`, attach caveats (opted-in coverage; omits CI-Tests/Contributors/Dependency-Update-Tool) and CDLA Permissive 2.0 attribution
- [ ] 4.3 Error handling: 404 → "may not have opted into publishing" hint; 400 → validation; 429/5xx → bounded retry/backoff; surface as MCP tool errors
- [ ] 4.4 Unit tests with recorded fixtures for 200 / 404 / 400 responses

## 5. Check catalog (offline)

- [ ] 5.1 Integrate `github.com/ossf/scorecard/v5/docs/checks` (`checks.Read()`); map to catalog entries (name, short, risk, tags, supported platforms, doc URL, experimental flag)
- [ ] 5.2 Implement `list_checks` and `explain_check` tools, including the unknown-check-name error path
- [ ] 5.3 Implement documentation resources: `scorecard://checks` index and the `scorecard://checks/{name}` template
- [ ] 5.4 Unit tests for list/explain/resources and catalog-version consistency

## 6. MCP server runtime

- [ ] 6.1 Implement a transport-agnostic `newServer()` factory that registers all tools, resources, and `instructions`
- [ ] 6.2 Set the responsible-AI framing in `instructions`; keep behavioral guidance out of tool descriptions
- [ ] 6.3 Register tool annotations (`readOnlyHint`, `destructiveHint:false`, `idempotentHint`, `title`; `openWorldHint` on network tools); verify names ≤64 chars
- [ ] 6.4 Wire the stdio transport entrypoint with graceful shutdown on SIGINT/SIGTERM
- [ ] 6.5 Map recoverable failures to MCP tool errors with actionable hints without terminating the transport

## 7. scorecard-results tools

- [ ] 7.1 `get_repo_score` — compact aggregate + per-check summary; provenance; caveats; `-1` presented as inconclusive
- [ ] 7.2 `get_check_result` — single check detail; unknown-check-name error
- [ ] 7.3 `compare_repos` — multiple repos; bounded maximum with an explicit truncation note
- [ ] 7.4 Add the optional `commit` argument to the retrieval tools (pins the request when supplied)
- [ ] 7.5 Unit tests covering each `scorecard-results` scenario

## 8. Cross-cutting compliance

- [ ] 8.1 Verify every result carries `source`, resolved commit SHA, date, Scorecard version, caveats, and attribution
- [ ] 8.2 Verify tool descriptions are self-describing contracts that reference the Scorecard docs/API and contain no behavioral instructions
- [ ] 8.3 Verify all tool inputs use constrained, described schemas

## 9. Testing and verification

- [ ] 9.1 Map each spec scenario (mcp-server, scorecard-results, check-catalog) to a test case
- [ ] 9.2 Integration test: initialize over stdio, list tools/resources, invoke each tool against fixtures
- [ ] 9.3 Run `golangci-lint` and `go test ./...` clean
- [ ] 9.4 Manual smoke test with the MCP Inspector

## 10. Documentation and distribution

- [ ] 10.1 README: what it is, install, MCP client (stdio) configuration, example tool calls, and the cached-data caveats/attribution
- [ ] 10.2 Provide a client config example (e.g. `.mcp.json` / Claude Desktop connector entry)
- [ ] 10.3 Add goreleaser/ko config to build a single static binary (may defer if scope-limited)

## 11. Change closeout

- [ ] 11.1 `openspec validate add-scorecard-mcp-server --strict` passes against the implemented behavior
- [ ] 11.2 Update `AGENTS.md`/README if any convention changed during implementation
- [ ] 11.3 Archive the OpenSpec change (`openspec archive add-scorecard-mcp-server`) once implemented and merged
