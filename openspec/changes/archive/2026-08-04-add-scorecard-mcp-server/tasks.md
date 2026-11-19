# Tasks: Scorecard MCP Server

## 1. Project scaffolding

- [x] 1.1 Initialize the Go module, matching the Scorecard repo's Go toolchain version to avoid CI friction
- [x] 1.2 Add dependencies: `github.com/modelcontextprotocol/go-sdk` v1.7.0 and `github.com/ossf/scorecard/v5` v5.5.0
- [x] 1.3 Lay out packages so MCP-SDK usage is isolated under `cmd/` and `internal/` (drop-in as `cmd/mcp` upstream):
      `internal/model`, `internal/provider`, `internal/scorecardref`, `internal/server`, `cmd/scorecard-mcp`
- [x] 1.4 Add `LICENSE` (Apache-2.0) and a `.golangci.yml` aligned with the Scorecard project

## 2. Core types and the provider seam

- [x] 2.1 Define `Result`, `RepoRef`, and `Check` types mirroring Scorecard JSON v2 plus provenance (`source`,
      resolved commit SHA, date, Scorecard version), `caveats`, `attribution`, and optional `annotations`
- [x] 2.2 Define the `ResultProvider` interface (`GetResult`, `Capabilities`) and `ProviderCapabilities`
- [x] 2.3 Define the tool `outputSchema`(s) matching the result types, returned as `structuredContent` with a JSON
      `text` fallback

## 3. Repository reference parsing

- [x] 3.1 Parse `platform/owner/repo` (default `github.com`; accept `gitlab.com`; tolerate trailing slash; clear
      validation errors)
- [x] 3.2 Enforce a platform enum and a 40-hex commit pattern; unit-test valid and invalid inputs

## 4. CachedRESTProvider

- [x] 4.1 Implement the REST client: `GET https://api.scorecard.dev/projects/{platform}/{org}/{repo}` with optional
      `?commit=` (host configurable)
- [x] 4.2 Map the REST response to `Result`; set `source=cached-rest`, attach caveats (opted-in coverage; omits
      CI-Tests/Contributors/Dependency-Update-Tool) and CDLA Permissive 2.0 attribution
- [x] 4.3 Error handling: 404 → "may not have opted into publishing" hint; 400 → validation; other statuses → error;
      surface as MCP tool errors
- [x] 4.4 Unit tests with an httptest mock for 200 / 404 / 400 responses

## 5. Check catalog (offline)

- [x] 5.1 Integrate `github.com/ossf/scorecard/v5/docs/checks` (`checks.Read()`); map to catalog entries (name, short,
      risk, tags, supported platforms, doc URL, experimental flag)
- [x] 5.2 Implement `list_checks` and `explain_check` tools, including the unknown-check-name error path
- [x] 5.3 Implement documentation resources: `scorecard://checks` index and the `scorecard://checks/{name}` template
- [x] 5.4 Unit tests for list/explain and catalog-version consistency (resource handlers covered via the integration
      smoke test)

## 6. MCP server runtime

- [x] 6.1 Implement a transport-agnostic `newServer()` factory that registers all tools, resources, and `instructions`
- [x] 6.2 Set the responsible-AI framing in `instructions`; keep behavioral guidance out of tool descriptions
- [x] 6.3 Register tool annotations (`readOnlyHint`, `destructiveHint:false`, `idempotentHint`, `title`;
      `openWorldHint` on network tools); verify names ≤64 chars
- [x] 6.4 Wire the stdio transport entrypoint with graceful shutdown on SIGINT/SIGTERM
- [x] 6.5 Map recoverable failures to MCP tool errors with actionable hints without terminating the transport

## 7. scorecard-results tools

- [x] 7.1 `get_repo_score` — compact aggregate + per-check summary; provenance; caveats; `-1` presented as inconclusive
- [x] 7.2 `get_check_result` — single check detail; unknown-check-name error
- [x] 7.3 `compare_repos` — multiple repos; bounded maximum with an explicit truncation note
- [x] 7.4 Add the optional `commit` argument to the retrieval tools (pins the request when supplied)
- [x] 7.5 Tests covering each `scorecard-results` scenario (via the in-memory MCP integration test)

## 8. Cross-cutting compliance

- [x] 8.1 Verify every result carries `source`, resolved commit SHA, date, Scorecard version, caveats, and attribution
      (confirmed via end-to-end stdio smoke test)
- [x] 8.2 Verify tool descriptions are self-describing contracts that reference the Scorecard docs/API and contain no
      behavioral instructions
- [x] 8.3 Verify all tool inputs use constrained, described schemas

## 9. Testing and verification

- [x] 9.1 Map each spec scenario (mcp-server, scorecard-results, check-catalog) to a test case
- [x] 9.2 Integration test: initialize over an in-memory transport, list tools/resources, invoke each tool with a fake provider
- [x] 9.3 Run `golangci-lint` (0 issues) and `go test ./...` clean
- [x] 9.4 Manual smoke test with the MCP Inspector (Inspector CLI over stdio: initialize; `tools/list` (5 tools with
      schemas + annotations); `resources/list` + `resources/templates/list`; `list_checks` returned 20 checks;
      `explain_check` OK path + unknown-check error path; `resources/read scorecard://checks`; `get_repo_score`
      returned a live score over the network)

## 10. Documentation and distribution

- [x] 10.1 README: what it is, install, MCP client (stdio) configuration, example tool calls, and the cached-data caveats/attribution
- [x] 10.2 Provide a client config example (Claude `.mcp.json` and VS Code `.vscode/mcp.json` entries in the README)
- [ ] 10.3 Add goreleaser/ko config to build a single static binary (deferred; carried forward to the upstreaming
      effort — release tooling lands with the in-tree `cmd/mcp` move, not in this change)

## 11. Change closeout

- [x] 11.1 `openspec validate add-scorecard-mcp-server --strict` passes against the implemented behavior
- [x] 11.2 Update `AGENTS.md`/README if any convention changed during implementation (README added)
- [x] 11.3 Archive the OpenSpec change (`openspec archive add-scorecard-mcp-server`) once implemented and merged
