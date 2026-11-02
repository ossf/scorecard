# Tasks: Serve Scorecard results to MCP clients from `scorecard mcp`

## 1. Specs

- [x] 1.1 Import the `mcp-server`, `scorecard-results`, and
  `check-catalog` specs and their archived change with history.
- [x] 1.2 Replace the placeholder Purpose sections and add the three
  capabilities to the ownership table in `openspec/README.md`.
- [x] 1.3 Record the subcommand, base URL, and catalog-source deltas.

## 2. Subcommand

- [ ] 2.1 Add `cmd/mcp.go` and register it on the root command.
- [ ] 2.2 Remove the standalone `main` package and move its integration
  test to drive `scorecard mcp`.
- [ ] 2.3 Replace the test that read the Scorecard version from `go.mod`.
- [ ] 2.4 Add the MCP dependencies to `go.mod` and `go.sum`.
- [ ] 2.5 Update `docs/mcp.md` for the `scorecard` binary.

## 3. Validation

- [ ] 3.1 Build, unit tests, and lint.
- [ ] 3.2 Validate the change and the imported specs with OpenSpec.
- [ ] 3.3 Smoke-test `scorecard mcp` with an MCP client.
