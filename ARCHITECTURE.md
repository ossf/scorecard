# Scorecard architecture

This document describes the actors in a Scorecard scan, what each one does,
and the external interfaces of the software this repository releases. The
behavior of each component is specified in more detail under
[`openspec/specs/`](openspec/specs/).

## Actors

- **User:** runs Scorecard against a repository or package, through the CLI,
  the container image, the Go API, or `scorecard serve`.
- **Scorecard:** the `scorecard` binary or the `pkg/scorecard` library. It
  collects data about a repository, evaluates it, and reports the result.
- **Repository host:** GitHub (including GitHub Enterprise), GitLab
  (including self-managed instances), or Azure DevOps. Scorecard reads
  repository contents and metadata through the host's API, authenticated with
  a token the user provides. Scorecard can also scan a local directory.
- **Auxiliary services:** OSV supplies known vulnerabilities, through
  OSV-Scanner. OSS-Fuzz reports fuzzing enrollment, and the OpenSSF Best
  Practices API reports badge level. The npm, PyPI, RubyGems, and NuGet
  registries map a package to its source repository.
- **Policy author:** writes an optional policy file that selects which checks
  run and sets a minimum score for each.
- **Repository maintainer:** can add a Scorecard configuration file to their
  repository to annotate check results.

## How a scan works

1. The user names a repository, a package, or a local directory, plus the
   checks or probes to run, either directly or through a policy file.
2. For a package, Scorecard looks up its source repository in the package
   registry.
3. Scorecard creates a read-only client for the repository host and the
   auxiliary clients it needs.
4. Each check collects raw data from repository files and the host's API
   ([`checks/raw/`](checks/raw/)). Scorecard parses files such as workflows,
   Dockerfiles, and dependency manifests; it doesn't build or run the scanned
   repository's code.
5. Probes turn the raw data into findings, each a yes-or-no answer with a
   location and remediation ([`probes/`](probes/)).
6. Each check scores its findings from 0 to 10, or -1 when inconclusive
   ([`checks/evaluation/`](checks/evaluation/)). The aggregate score weights
   checks by risk.
7. Scorecard attaches any maintainer annotations from the scanned repository
   and writes the result in the requested format. A policy never changes
   scores; SARIF output reports only enforced checks that score below the
   policy's minimum.

## Related systems in other repositories

- [`ossf/scorecard-action`](https://github.com/ossf/scorecard-action) runs
  Scorecard in GitHub Actions and can publish results to the REST API.
- [`ossf/scorecard-infra`](https://github.com/ossf/scorecard-infra) runs the
  weekly batch scan of over a million repositories and the
  [REST API](https://api.scorecard.dev) that serves its results and the
  results published by the Action.
- [`ossf/scorecard-webapp`](https://github.com/ossf/scorecard-webapp) is the
  [scorecard.dev](https://scorecard.dev) website.

## External interfaces

Inputs:

- **CLI:** the `scorecard` command, its flags, and its environment variables
  ([spec](openspec/specs/cli/spec.md), [usage](README.md#using-scorecard)).
- **Authentication:** `GITHUB_AUTH_TOKEN` (or `GITHUB_TOKEN`, `GH_TOKEN`,
  `GH_AUTH_TOKEN`), or a GitHub App through `GITHUB_APP_KEY_PATH`,
  `GITHUB_APP_ID`, and `GITHUB_APP_INSTALLATION_ID`; `GITLAB_AUTH_TOKEN`;
  `AZURE_DEVOPS_AUTH_TOKEN` ([spec](openspec/specs/repo-clients/spec.md)).
- **Go API:** the `pkg/scorecard` package
  ([spec](openspec/specs/public-go-api/spec.md),
  [reference](https://pkg.go.dev/github.com/ossf/scorecard/v5/pkg/scorecard)).
- **HTTP server:** `scorecard serve`, which listens on `PORT` (default 8080)
  ([spec](openspec/specs/http-server/spec.md)).
- **Policy file:** `--policy`
  ([spec](openspec/specs/policy/spec.md)).
- **Maintainer annotations:** a Scorecard configuration file in the scanned
  repository ([spec](openspec/specs/maintainer-annotations/spec.md)).
- **Package names:** `--npm`, `--pypi`, `--rubygems`, `--nuget`
  ([spec](openspec/specs/package-manager-resolution/spec.md)).

Outputs:

- **Result formats:** text, JSON, SARIF, probe findings, raw data, and in-toto
  ([spec](openspec/specs/result-formats/spec.md)). The JSON schemas are
  [`json.v2.schema`](pkg/scorecard/json.v2.schema) and
  [`json.raw.schema`](pkg/scorecard/json.raw.schema).
- **Release artifacts:** binaries and archives on
  [GitHub Releases](https://github.com/ossf/scorecard/releases) with SLSA
  provenance, and the `ghcr.io/ossf/scorecard` container image.

Outbound calls: the repository host APIs and auxiliary services listed under
[Actors](#actors).
