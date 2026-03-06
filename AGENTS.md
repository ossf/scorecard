# AGENTS.md

This file provides guidance for AI coding agents working on the OpenSSF Scorecard project.

## Project Overview

OpenSSF Scorecard is a security health metrics tool for open source projects. It runs automated checks and probes against repositories (GitHub, GitLab, Azure DevOps) to produce security scores (0-10) across areas like branch protection, code review, dependency management, and more.

The project is written in Go and follows a three-tier architecture:
1. **Raw data collection** (`checks/raw/`) - gathers data from repository APIs
2. **Probe execution** (`probes/`) - analyzes raw data with granular heuristics
3. **Evaluation/scoring** (`checks/evaluation/`) - converts probe findings into numeric scores

## Spec-Driven Development

This project uses [OpenSpec](https://openspec.dev/) for planning non-trivial changes. Before implementing significant features or architectural changes, follow the OpenSpec workflow:

1. **Explore** the codebase and existing specs in `openspec/specs/`
2. **Propose** a change in `openspec/changes/<change-id>/proposal.md`
3. **Design** the technical approach in `design.md`
4. **Spec** the requirements and scenarios in `specs/<feature>/spec.md`
5. **Implement** the change
6. **Verify** against the spec

Do not restructure or make architectural changes without a spec.

## Build and Test

### Prerequisites

- Go v1.23.0+
- `protoc` v3+
- `make`
- A GitHub personal access token (`GITHUB_AUTH_TOKEN` env var) for e2e tests

### Key Commands

```bash
# Install build tools
make install

# Build the scorecard binary
make build-scorecard

# Run ALL checks (build, test, lint, docs validation)
make all

# Run unit tests only
make unit-test

# Run linter only
make check-linter

# Auto-fix linter issues
make fix-linter

# Run e2e tests (requires GITHUB_AUTH_TOKEN)
make e2e-pat

# Generate documentation
make generate-docs

# Run scorecard locally
go run main.go --repo=github.com/<owner>/<repo> --format json
```

### Useful CLI Flags

- `--repo` / `--local` - specify a remote or local repository to scan
- `--checks` - comma-separated list of specific checks to run
- `--probes` - comma-separated list of specific probes to run
- `--format` - output format: `default`, `json`, `sarif`, `probe`, `raw`
- `--show-details` - show detailed findings

## Code Style and Conventions

- Linting is enforced via `golangci-lint` with the config in `.golangci.yml`
- Always run `make check-linter` before submitting changes
- Use `make fix-linter` to auto-fix most issues
- Run `go mod tidy && go mod verify` after changing dependencies
- **Follow existing patterns.** Do not introduce new frameworks, libraries, or patterns without strong justification.
- **Do not introduce new dependencies** without justification. The project favors minimal dependency trees.
- **Prefer editing existing files** over creating new ones. Do not create documentation files unless explicitly asked.
- **Do not add docstrings, comments, or type annotations** to code you did not change.

## Project Structure

```
scorecard/
├── main.go                          # CLI entry point
├── cmd/                             # CLI command definitions
├── pkg/scorecard/                   # Core scoring engine (scorecard.Run())
├── checker/                         # Check infrastructure and result types
├── checks/                          # Check implementations
│   ├── raw/                         # Raw data collectors
│   └── evaluation/                  # Scoring/evaluation logic
├── probes/                          # Granular probe implementations
│   └── <probeName>/
│       ├── def.yml                  # Probe documentation/definition
│       ├── impl.go                  # Probe implementation
│       └── impl_test.go             # Probe tests
├── clients/                         # Platform API clients
│   ├── githubrepo/                  # GitHub (REST + GraphQL)
│   ├── gitlabrepo/                  # GitLab
│   ├── azuredevopsrepo/             # Azure DevOps (experimental)
│   └── localdir/                    # Local directory scanning
├── finding/                         # Finding data structures
├── policy/                          # Policy enforcement
├── cron/                            # Batch scanning infrastructure
├── attestor/                        # Attestation support
├── docs/                            # Documentation
│   └── checks/internal/checks.yaml # Check descriptions (source of truth)
└── e2e/                             # End-to-end tests
```

## Adding New Checks

See `checks/write.md` for the full guide. In summary:

1. Create `checks/mycheck.go` with a `CheckMyCheckName` constant
2. Register with `registerCheck()` in an `init()` function
3. Implement raw data collection in `checks/raw/`
4. Create probes in `probes/` (see below)
5. Add evaluation logic in `checks/evaluation/`
6. Write unit tests in `checks/mycheck_test.go`
7. Write e2e tests in `e2e/mycheck_test.go`
8. Update `docs/checks/internal/checks.yaml` and run `make generate-docs`

## Adding New Probes

See `probes/README.md` for details. Each probe has three files:

- `def.yml` - defines the probe's id, lifecycle (Experimental/Stable/Deprecated), description, motivation, implementation details, outcomes, and remediation
- `impl.go` - implementation; must return `[]finding.Finding`
- `impl_test.go` - tests

Register probes via `probes.MustRegister()` (for check-associated probes) or `probes.MustRegisterIndependent()`. The probe catalog is in `probes/entries.go`.

Use `make setup-probe probeName=myProbeName` to scaffold a new probe.

Probe names use camelCase and should be phrased as boolean statements (e.g., `hasUnverifiedBinaryArtifacts`, `branchesAreProtected`).

## Key Interfaces

### `clients.RepoClient`

The main interface for interacting with repository platforms. Implementations exist for GitHub, GitLab, Azure DevOps, and local directories. Key methods include `InitRepo()`, `ListCommits()`, `ListFiles()`, `GetBranch()`, `ListCheckRunsForRef()`, `Search()`, etc.

### `checker.CheckFn`

The function signature for check entry points: `func(c *checker.CheckRequest) checker.CheckResult`

## External Service Integrations

Scorecard integrates with several external services:
- **OSV** (Open Source Vulnerabilities) - vulnerability data
- **OSS-Fuzz** - fuzzing status
- **OpenSSF Best Practices** (formerly CII) - badge status
- **deps.dev** - dependency/package data

## Git, PR, and Commit Guidelines

- **Never commit directly to `main`.** Always use feature branches and PRs.
- **Do not create PRs unless explicitly asked.** Commit and push to feature branches, then wait for instruction.
- **Do not force-push** without explicit permission.
- **Stage specific files** when committing. Do not use `git add -A` or `git add .` — this avoids accidentally committing secrets or binaries.
- Every PR title must be prefixed with an emoji indicating the change type:
  - `:warning:` - breaking change
  - `:sparkles:` - new feature
  - `:bug:` - bug fix
  - `:book:` - documentation
  - `:seedling:` - infra/tests/other
  - `:ghost:` - no release note
- **All commits MUST include a DCO sign-off.** Always pass the `-s` flag to `git commit`. This adds the `Signed-off-by` trailer using the committer's git config `user.name` and `user.email`. DCO checks will fail without it.
- **Commit messages should be detailed.** Explain *why*, not just *what*. Use bullet points for multi-part changes.
- **AI co-authorship trailer is required** on all AI-assisted commits.
- **Commit message format example:**
  ```bash
  git commit -s -m "$(cat <<'EOF'
  :seedling: Short description of the change

  Detailed explanation of why this change was made.

  Co-Authored-By: Claude <noreply@anthropic.com>
  EOF
  )"
  ```
  The `-s` flag adds the `Signed-off-by` line automatically. Use just "Claude" in the co-authorship trailer — do not include the model version.
- Address review feedback without rebasing; use merge commits to sync with `main`

## Documentation

- **DO NOT** edit `docs/checks.md` directly. Edit `docs/checks/internal/checks.yaml` and run `make generate-docs`
- Probe documentation lives in each probe's `def.yml`
- Run `make validate-docs` to check documentation validity

## Security Considerations

- Do not introduce OWASP top-10 vulnerabilities
- Be cautious with GitHub API token handling - tokens are passed via environment variables
- The `checks/dangerous_workflow.go` check specifically looks for dangerous patterns in GitHub Actions workflows; be careful when modifying its detection logic
- Score changes require discussion in a GitHub issue before implementation

## Testing

- Unit tests: `make unit-test` (runs `go test` with race detection and coverage)
- E2e tests use dedicated test repositories (e.g., `ossf-tests/*`) that should not change over time
- E2e tests use the Ginkgo framework
- Mock clients are in `clients/mockclients/` (generated via `mockgen`)
- Generate mocks with `make generate-mocks`
