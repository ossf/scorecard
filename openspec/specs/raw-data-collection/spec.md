# raw-data-collection Specification

## Purpose

Raw data collection is the first tier of every Scorecard check. For each
check it gathers evidence from the repository's files and from the hosting
platform's APIs, without scoring it, so that probes can turn the evidence
into findings. This spec describes which files, API data, and time windows
each check reads, how patterns are detected, and how missing data is
handled.

## Requirements

### Requirement: Repository File Matching

The system SHALL select repository files for analysis by matching a glob
pattern against the file's full path and, when that fails, against the
file's base name. Matching SHALL be case-insensitive unless a check states
otherwise. Files SHALL be read at the commit being analyzed. When the git
file mode is selected, the file list SHALL come from a shallow clone of the
repository at that commit.

#### Scenario: Base-name match in a subdirectory

- **GIVEN** a repository containing `build/ci/Dockerfile.release`
- **WHEN** a check searches for files matching `*Dockerfile*`
- **THEN** the file is selected even though the pattern has no directory
  part

#### Scenario: Case-insensitive match

- **GIVEN** a repository containing `.GITHUB/WORKFLOWS/CI.YML`
- **WHEN** a check searches case-insensitively for `.github/workflows/*`
- **THEN** the file is selected

### Requirement: Test Data Exclusion

The system SHALL skip files whose path starts with `testdata/` or
`src/test/`, or contains `/testdata/` or `/src/test/`, whenever a check
selects files by pattern. Checks that walk the full file list by exact name
(License, Security-Policy, Dependency-Update-Tool) SHALL NOT apply this
exclusion.

#### Scenario: Binary in test data is ignored

- **GIVEN** a repository containing `pkg/testdata/sample.exe`
- **WHEN** the Binary-Artifacts raw data is collected
- **THEN** `pkg/testdata/sample.exe` is not reported

#### Scenario: Java test sources are ignored

- **GIVEN** a repository containing `module/src/test/resources/run.sh`
- **WHEN** the Pinned-Dependencies raw data is collected
- **THEN** the script is not analyzed

### Requirement: GitHub Workflow File Recognition

The system SHALL treat a file as a GitHub Actions workflow only when it sits
directly in `.github/workflows/` and has a `.yml` or `.yaml` extension.
Workflows that contain only blank lines and `#` comments SHALL be skipped by
the Dangerous-Workflow, Token-Permissions, and Pinned-Dependencies
collectors. A workflow that cannot be parsed into any structure SHALL cause
the collecting check to fail with an error.

#### Scenario: Nested file is not a workflow

- **GIVEN** a file `.github/workflows/templates/build.yml`
- **WHEN** workflow-based raw data is collected
- **THEN** the file is not treated as a workflow

#### Scenario: Comment-only workflow is skipped

- **GIVEN** a workflow file that contains only `#` comments
- **WHEN** the Token-Permissions raw data is collected
- **THEN** the workflow is not counted and produces no permission entries

#### Scenario: Unparseable workflow

- **GIVEN** a workflow file whose YAML cannot be parsed
- **WHEN** the Dangerous-Workflow raw data is collected
- **THEN** the collection fails with a workflow parse error

### Requirement: Commit and Issue Windows

The system SHALL analyze the most recent commits on the default branch,
30 commits by default, adjustable with the commit-depth option. On GitHub,
the system SHALL read the 30 most recently updated issues together with the
last 30 comments on each. Data that depends on merged pull requests
(Code-Review, CI-Tests, SAST) SHALL be derived from the pull request
associated with each commit in that window.

#### Scenario: Default commit window

- **WHEN** a GitHub repository is scanned without a commit-depth option
- **THEN** commit-based raw data covers the 30 most recent default-branch
  commits

#### Scenario: Custom commit window

- **WHEN** a scan sets the commit depth to 100
- **THEN** commit-based raw data covers up to 100 commits

### Requirement: Unavailable Data Handling

The system SHALL return an error for a check when a required data source
fails. When a client reports that a feature is unsupported, the system
SHALL continue without that data in these cases: release listings for
Branch-Protection and SBOM, workflow run listings for Binary-Artifacts and
Packaging, the license API, the organization `.github` repository for
Security-Policy, commit search for Dependency-Update-Tool, and commit
listing for SAST.

#### Scenario: API failure

- **GIVEN** the webhook listing API returns an error
- **WHEN** the Webhooks raw data is collected
- **THEN** the check reports an internal error

#### Scenario: Unsupported release listing

- **GIVEN** a client that does not support listing releases
- **WHEN** the SBOM raw data is collected
- **THEN** release assets are skipped and source files are still searched

### Requirement: Result Repository Metadata

For a repository with at least one commit, the system SHALL record, before
any check or probe runs, raw metadata with these keys: `repository.host`,
`repository.name` (the URI without the host prefix), `repository.uri`,
`repository.sha1` (the SHA of the first commit the client lists, or `unknown`
when the client cannot list commits, as for local directories),
`repository.defaultBranch`, and `localPath` (the client's local checkout
path, which may be empty). When the client reports that default-branch lookup
is unsupported, `repository.defaultBranch` SHALL be `unknown`; any other
default-branch lookup error SHALL end the run with an internal error
beginning `GetDefaultBranchName:`.

#### Scenario: Remote repository metadata

- **GIVEN** `github.com/org/repo` with default branch `main`
- **WHEN** Scorecard scans it
- **THEN** the raw metadata has `repository.host` `github.com`,
  `repository.name` `org/repo`, and `repository.defaultBranch` `main`
- **AND** `repository.sha1` is the SHA of the first listed commit

#### Scenario: Local directory metadata

- **WHEN** Scorecard scans a local directory
- **THEN** the raw metadata has `repository.sha1` and
  `repository.defaultBranch` set to `unknown`
- **AND** `localPath` is the directory's absolute path

### Requirement: Binary Artifact Detection

The system SHALL scan every non-excluded file and read up to its first 1024
bytes. A file SHALL be reported as binary when its content signature maps
to one of these types, or when its extension is one of them and the bytes
are not plain text: `crx`, `deb`, `dex`, `dey`, `elf`, `o`, `a`, `so`,
`macho`, `iso`, `class`, `jar`, `bundle`, `dylib`, `lib`, `msi`, `dll`,
`drv`, `efi`, `exe`, `ocx`, `pyc`, `pyo`, `par`, `rpm`, `wasm`, `whl`.
Empty files SHALL be ignored. A `gradle-wrapper.jar` SHALL be marked as a
verified binary when a workflow uses `gradle/wrapper-validation-action` or
`gradle/actions/wrapper-validation` and a successful run of that workflow
exists for the latest commit.

#### Scenario: Executable detected by content

- **GIVEN** a file `tools/helper` whose bytes are an ELF executable
- **WHEN** the Binary-Artifacts raw data is collected
- **THEN** `tools/helper` is reported as a binary file

#### Scenario: Text file with binary extension

- **GIVEN** a file `notes.lib` that contains plain text
- **WHEN** the Binary-Artifacts raw data is collected
- **THEN** `notes.lib` is not reported

#### Scenario: Validated Gradle wrapper

- **GIVEN** `gradle/wrapper/gradle-wrapper.jar` exists
- **AND** a workflow uses `gradle/actions/wrapper-validation`
- **AND** that workflow has a successful run on the latest commit
- **WHEN** the Binary-Artifacts raw data is collected
- **THEN** the jar is reported as a verified binary

### Requirement: Branch Selection for Branch Protection

The system SHALL collect protection data for the default branch and for
each branch that a release targets. A release target that is a 40-character
hexadecimal commit SHA SHALL be skipped. When a release targets `master`
and no such branch exists, the system SHALL use `main` instead. Each branch
SHALL appear only once. A release without a target SHALL cause an error.

#### Scenario: Release branch included

- **GIVEN** a default branch `main` and a release targeting `release-1.x`
- **WHEN** the Branch-Protection raw data is collected
- **THEN** protection data is collected for `main` and `release-1.x`

#### Scenario: Release pinned to a commit

- **GIVEN** a release whose target is a full commit SHA
- **WHEN** the Branch-Protection raw data is collected
- **THEN** no extra branch is added for that release

### Requirement: Branch Protection Settings

For each selected branch, the system SHALL record whether it is protected
and, where the platform provides them, these settings: deletions allowed,
force pushes allowed, linear history required, rules enforced for
administrators, approval of the last push required, pull requests required,
required approving review count, stale reviews dismissed, code owner review
required, status checks required, branches required to be up to date, and
the required status check contexts. Settings the token cannot read SHALL be
recorded as unknown rather than false. The system SHALL also record the
path of every file named exactly `CODEOWNERS`, matched case-sensitively.

#### Scenario: Admin-only settings without admin token

- **GIVEN** a GitHub token without administrator access to the repository
- **WHEN** the Branch-Protection raw data is collected
- **THEN** settings that need administrator access are recorded as unknown

#### Scenario: CODEOWNERS file recorded

- **GIVEN** a repository containing `.github/CODEOWNERS`
- **WHEN** the Branch-Protection raw data is collected
- **THEN** `.github/CODEOWNERS` is listed as a code owners file

### Requirement: CII Best Practices Badge

The system SHALL look up the repository URI in the OpenSSF Best Practices
service and record the badge level returned. The check SHALL fail with an
error when no Best Practices client is configured or the lookup fails.

#### Scenario: Badge level recorded

- **GIVEN** a repository registered with a passing badge
- **WHEN** the CII-Best-Practices raw data is collected
- **THEN** the raw data records the passing badge level

### Requirement: CI Test Evidence

For each commit in the commit window that has a merged pull request, the
system SHALL collect the check runs and commit statuses reported for the
pull request's head commit, and record them with the pull request number.
Commits without a merged pull request SHALL be skipped. A failure to list
commits, check runs, or statuses SHALL cause an error.

#### Scenario: Merged pull request with checks

- **GIVEN** a recent commit merged through pull request 42
- **WHEN** the CI-Tests raw data is collected
- **THEN** the check runs and statuses for the head commit of pull request
  42 are recorded

#### Scenario: Direct push

- **GIVEN** a recent commit with no associated merged pull request
- **WHEN** the CI-Tests raw data is collected
- **THEN** no CI entry is recorded for that commit

### Requirement: Code Review Changesets

The system SHALL group commits in the commit window into changesets by
review platform, checking each commit in this order: Prow (merged pull
request with an `lgtm` or `approved` label), GitHub (merged pull request),
Phabricator (`Differential Revision:` with a `D<number>` id), Gerrit (both
`Reviewed-on:` and `Reviewed-by:` in the message), and Piper
(`PiperOrigin-RevId:` with at least three digits). A commit matching none
SHALL form its own changeset with an unknown platform. For GitHub
changesets, the system SHALL record the pull request author and reviews,
and SHALL count whoever merged the pull request as an approving reviewer.

#### Scenario: Commits from one pull request

- **GIVEN** two recent commits that belong to the same merged pull request
- **WHEN** the Code-Review raw data is collected
- **THEN** both commits are grouped into one GitHub changeset

#### Scenario: Gerrit-reviewed commit

- **GIVEN** a commit whose message contains `Reviewed-on:` and
  `Reviewed-by:` lines
- **WHEN** the Code-Review raw data is collected
- **THEN** the commit forms a Gerrit changeset identified by its SHA

### Requirement: Contributor Data

The system SHALL list the repository's contributors with their
contribution counts, code owner status, organization memberships, and
company names. Company names SHALL be normalized by lowercasing, removing
`inc.`, `llc`, and commas, trimming a leading `@` and surrounding spaces,
and dropping duplicates. On GitHub, contributor data SHALL be available
only when the analyzed commit is HEAD.

#### Scenario: Company normalization

- **GIVEN** a contributor whose profile company is `@Example Inc.`
- **WHEN** the Contributors raw data is collected
- **THEN** the recorded company is `example`

#### Scenario: Non-HEAD commit

- **WHEN** a GitHub repository is scanned at a specific older commit
- **THEN** the Contributors raw data collection fails as unsupported

### Requirement: Untrusted Checkout Detection

The system SHALL report an untrusted checkout for each `actions/checkout`
step whose `ref` input contains `github.event.pull_request` or
`github.event.workflow_run`, but only in workflows triggered by
`pull_request_target` or `workflow_run`. Each report SHALL include the
file, the step's line, the `ref` value, and the job.

#### Scenario: Checkout of pull request head on pull_request_target

- **GIVEN** a workflow triggered by `pull_request_target`
- **AND** a step using `actions/checkout` with
  `ref: ${{ github.event.pull_request.head.sha }}`
- **WHEN** the Dangerous-Workflow raw data is collected
- **THEN** an untrusted checkout is reported at that step

#### Scenario: Same checkout on pull_request

- **GIVEN** the same checkout step in a workflow triggered only by
  `pull_request`
- **WHEN** the Dangerous-Workflow raw data is collected
- **THEN** no untrusted checkout is reported

### Requirement: Script Injection Detection

The system SHALL inspect every `${{ ... }}` expression in workflow `run`
scripts and report a script injection when the expression contains
`github.head_ref`, applies `toJSON` to `github` or `github.event`, or
references a `github.event.` field that attackers control: issue,
pull request, and discussion titles and bodies; comment, review, and review
comment bodies; page names; commit messages and author names and emails;
blocked user names and emails; and pull request head ref, label, and repo
default branch. The system SHALL also record the number of workflows
analyzed.

#### Scenario: Issue title in a run script

- **GIVEN** a `run` step containing `echo "${{ github.event.issue.title }}"`
- **WHEN** the Dangerous-Workflow raw data is collected
- **THEN** a script injection is reported with that expression as the
  snippet

#### Scenario: Trusted context

- **GIVEN** a `run` step containing `echo "${{ github.sha }}"`
- **WHEN** the Dangerous-Workflow raw data is collected
- **THEN** no script injection is reported

### Requirement: Dependency Update Tool Detection

The system SHALL compare every repository path, case-insensitively, with
known configuration files: Dependabot (`.github/dependabot.yml`,
`.github/dependabot.yaml`); Renovate (`renovate.json`, `renovate.json5`,
and those names under `.github/` or `.gitlab/`, `.renovaterc`,
`.renovaterc.json`, `.renovaterc.json5`); and Scala Steward
(`.scala-steward.conf` or `scala-steward.conf` at the root or under
`.github/` or `.config/`). When no file is found, the system SHALL search
for commits authored by `dependabot[bot]` and record Dependabot, without a
file, if one was committed by the Dependabot account.

#### Scenario: Renovate configuration

- **GIVEN** a repository containing `.github/renovate.json5`
- **WHEN** the Dependency-Update-Tool raw data is collected
- **THEN** RenovateBot is recorded with that file

#### Scenario: Dependabot commits without config

- **GIVEN** no configuration file but recent commits by Dependabot
- **WHEN** the Dependency-Update-Tool raw data is collected
- **THEN** Dependabot is recorded as the tool

### Requirement: Continuous Fuzzing Service Detection

The system SHALL record OSS-Fuzz when the OSS-Fuzz project status list
contains the repository URI, and ClusterFuzzLite when the repository has a
`.clusterfuzzlite/Dockerfile` (matched case-sensitively) with at least one
non-comment line. OSS-Fuzz detection SHALL be skipped when no OSS-Fuzz
client is configured.

#### Scenario: ClusterFuzzLite configured

- **GIVEN** a `.clusterfuzzlite/Dockerfile` containing build instructions
- **WHEN** the Fuzzing raw data is collected
- **THEN** ClusterFuzzLite is recorded as a fuzzer

### Requirement: Language Fuzz Function Detection

The system SHALL search source files of each prominent repository language
for a language-specific fuzzing or property-testing marker, recording each
matching line. A language SHALL be prominent when its line count is at
least a quarter of the average line count per language; when the client
reports all languages, every supported language SHALL be searched. The
markers SHALL be: Go native fuzz functions in `*_test.go`; libFuzzer entry
points in C, C++, and Swift; `libfuzzer_sys` in Rust; Jazzer's
`FuzzedDataProvider` import in Java; `import atheris` in Python; fast-check
imports in JavaScript and TypeScript; QuickCheck, Hedgehog, Validity, or
SmallCheck imports in Haskell; EQC or PropEr includes in Erlang; PropCheck
or ExUnitProperties in Elixir; `qcheck` in Gleam; and FsCheck in C# and F#.

#### Scenario: Go fuzz test

- **GIVEN** a Go repository with `func FuzzParse(f *testing.F)` in
  `parse_test.go`
- **WHEN** the Fuzzing raw data is collected
- **THEN** the Go built-in fuzzer is recorded with that file and line

#### Scenario: Minor language not searched

- **GIVEN** a repository where Python makes up far less than a quarter of
  the average lines per language
- **WHEN** the Fuzzing raw data is collected
- **THEN** Python files are not searched for Atheris

### Requirement: License Detection

The system SHALL first ask the platform's license API for license files and,
if it returns any, record each with its SPDX identifier. When the API
returns none or is unsupported, the system SHALL record the first top-level
file whose name matches the license pattern: optionally an SPDX-like prefix
or suffix around `LICENSE`, `LICENCE`, `COPYING`, `COPYRIGHT`, or `PATENTS`
(case-insensitive), with no extension or one of `.adoc`, `.asc`, `.doc`,
`.docx`, `.ext`, `.html`, `.markdown`, `.md`, `.rst`, `.txt`, `.xml`. A
license SHALL be marked approved when its SPDX identifier is on the
built-in FSF/OSI list (case-insensitive). A filename match without an SPDX
identifier SHALL be recorded as `NOASSERTION` named `Other`.

#### Scenario: Plain license file

- **GIVEN** a local repository with a top-level `LICENSE` file
- **WHEN** the License raw data is collected
- **THEN** `LICENSE` is recorded with SPDX identifier `NOASSERTION`

#### Scenario: SPDX identifier in the file name

- **GIVEN** a local repository with a top-level `LICENSE_LPPL-1.3c` file
- **WHEN** the License raw data is collected
- **THEN** the license is recorded as `LPPL-1.3c` and marked approved

#### Scenario: Nested file is not a license

- **GIVEN** a repository with only `.github/workflows/license-check.yml`
- **WHEN** the License raw data is collected
- **THEN** no license file is recorded

### Requirement: Maintenance Activity Data

The system SHALL record whether the repository is archived, the commits in
the commit window, the recent issues with their comments, and the
repository creation time. A failure to read any of these SHALL cause an
error.

#### Scenario: Archived repository

- **GIVEN** an archived repository
- **WHEN** the Maintained raw data is collected
- **THEN** the archived status is recorded as true

### Requirement: Packaging Workflow Detection

For GitHub, the system SHALL check each workflow for a job that publishes a
package, recognized by: `actions/setup-node` with the npm registry URL plus
an `npm publish` step; `actions/setup-java` plus `mvn deploy` or
`gradle publish`; `sbt ci-release`; `gem push`; `nuget push`;
`docker push` or `docker/build-push-action`; `pypa/gh-action-pypi-publish`;
`relekang/python-semantic-release`; `goreleaser/goreleaser-action`;
`cargo publish`; `imjasonh/setup-ko` or `ko-build/setup-ko`;
`npx semantic-release`; or `erlef/setup-elixir` plus `hex.publish`. A
matching workflow SHALL count as packaging only when it has a successful
run, unless the platform cannot list runs. For GitLab, the system SHALL
search the flattened CI configuration for `docker push`, `nuget push`,
`poetry publish`, or `twine upload`. For local directories, the system
SHALL apply both methods.

#### Scenario: Published with a successful run

- **GIVEN** a workflow using `pypa/gh-action-pypi-publish`
- **AND** the workflow has a successful run
- **WHEN** the Packaging raw data is collected
- **THEN** a package is recorded with the workflow file and run URLs

#### Scenario: Publishing workflow never ran

- **GIVEN** a publishing workflow with no successful runs
- **WHEN** the Packaging raw data is collected
- **THEN** the workflow is recorded with a message that it was not used in
  runs

### Requirement: GitHub Action Pinning

The system SHALL record every `uses:` reference in workflow steps and every
job-level reusable workflow call as a GitHub Action dependency, matching
workflow paths case-sensitively. References starting with `./` or `$/`
SHALL be skipped. A reference SHALL be pinned when it points to a local
path, ends in `@` followed by at least 40 hexadecimal characters, or is a
`docker://` image with an `@sha256:` digest. Unpinned actions SHALL carry a
remediation that suggests pinning to a commit SHA.

#### Scenario: Tag reference

- **GIVEN** a step with `uses: actions/checkout@v4`
- **WHEN** the Pinned-Dependencies raw data is collected
- **THEN** an unpinned GitHub Action dependency is recorded

#### Scenario: Commit SHA reference

- **GIVEN** a step using `actions/checkout@` followed by a 40-character SHA
- **WHEN** the Pinned-Dependencies raw data is collected
- **THEN** the dependency is recorded as pinned

### Requirement: Container Image Pinning

The system SHALL parse files matching `*Dockerfile*` and record each `FROM`
image as a container image dependency. Files SHALL be skipped when they are
under a `vendor` or `third_party` directory, have a Go, C, C++, Rust,
JavaScript, Python, or Java source extension, are shell scripts, contain
only comments, or have `template`, `tmpl`, or `tpl` as a name token. An
image SHALL be pinned when it carries an `@sha256:` digest of 64 hex
characters or a variable reference, after substituting `ARG` defaults. A
stage alias SHALL inherit the pinned status of its source, and `scratch`
SHALL never be recorded. A Dockerfile that cannot be parsed SHALL cause an
error.

#### Scenario: Image by tag

- **GIVEN** a Dockerfile with `FROM python:3.12`
- **WHEN** the Pinned-Dependencies raw data is collected
- **THEN** an unpinned container image `python` at `3.12` is recorded

#### Scenario: Stage built from a pinned image

- **GIVEN** `FROM golang@sha256:<digest> AS build` followed by `FROM build`
- **WHEN** the Pinned-Dependencies raw data is collected
- **THEN** both `FROM` lines are recorded as pinned

### Requirement: Insecure Download Detection

The system SHALL parse shell commands from Dockerfile `RUN` instructions
(including heredocs), from shell script files, and from workflow `run` steps
whose shell resolves to `sh`, `bash`, or `mksh`. A file SHALL count as a
shell script when its shebang names one of those shells or, with no
shebang, when its extension is `.sh`, `.bash`, or `.mksh`. Workflow steps
SHALL default to `bash`, except steps or jobs that run only on Windows,
which SHALL be skipped. The system SHALL record an unpinned download-then-run
dependency when `curl`, `wget`, `gsutil`, or `aws s3api get-object` is
piped into an interpreter, when a downloaded file is later executed, or
when a download is run through process substitution. A piped download SHALL
not be recorded when every URL is a `raw.githubusercontent.com` URL pinned
to a commit SHA. Shell snippets that cannot be parsed SHALL be recorded as
processing errors.

#### Scenario: Curl piped to bash

- **GIVEN** a script with `curl -sSL https://example.com/install.sh | bash`
- **WHEN** the Pinned-Dependencies raw data is collected
- **THEN** an unpinned download-then-run dependency is recorded

#### Scenario: Download pinned to a commit

- **GIVEN** a command piping a `raw.githubusercontent.com` URL that uses a
  40-character commit SHA into `sh`
- **WHEN** the Pinned-Dependencies raw data is collected
- **THEN** no download-then-run dependency is recorded

#### Scenario: Windows-only job

- **GIVEN** a workflow job with `runs-on: windows-latest` and no explicit
  shell
- **WHEN** the Pinned-Dependencies raw data is collected
- **THEN** the job's `run` steps are not analyzed for downloads

### Requirement: Package Manager Install Pinning

Within the parsed shell commands, the system SHALL record package manager
installs as dependencies with these pinning rules:

- `go get` or `go install` SHALL be pinned when the module has `@none`, a
  commit hash, or a full semantic version without `-insecure`.
- `pip install` and `python -m pip install` SHALL be pinned with
  `--require-hashes`, when installing only `.whl` files, or for editable
  installs from a VCS URL at a 40-character commit with `--no-deps`.
- `npm install`, `i`, `install-test`, and `update` SHALL be unpinned unless
  they install a git dependency at a commit hash; `npm ci` SHALL be pinned.
- `choco install` SHALL be pinned with `--requirechecksum`,
  `--requirechecksums`, or `--require-checksums`.
- `dotnet add package` SHALL be pinned with `-v` or `--version`;
  `nuget install` SHALL be pinned with `-Version` or a `packages.config`
  argument; `nuget restore -LockedMode`, `dotnet restore --locked-mode`,
  and `msbuild /t:restore /p:RestoreLockedMode=true` SHALL be pinned.

#### Scenario: Unpinned pip install

- **GIVEN** a `run` step with `pip install requests`
- **WHEN** the Pinned-Dependencies raw data is collected
- **THEN** an unpinned pip command dependency is recorded

#### Scenario: Hash-checked pip install

- **GIVEN** `pip install --require-hashes -r requirements.txt`
- **WHEN** the Pinned-Dependencies raw data is collected
- **THEN** the pip command dependency is recorded as pinned

### Requirement: NuGet Lock Configuration Override

When unpinned NuGet commands are found, the system SHALL read
`Directory.*.props` and `*.csproj` files. If central package management is
enabled and every `PackageVersion` uses an exact bracketed version such as
`[1.2.3]`, or if central package management is off and every project file
sets `RestoreLockedMode`, the system SHALL mark all NuGet commands as
pinned. The remediation for partially locked configurations is defined in
`remediation` (NuGet Pinning Suggestion). Malformed props or project files
SHALL be logged as warnings and skipped.

#### Scenario: All projects locked

- **GIVEN** an unpinned `dotnet restore` command
- **AND** every `.csproj` file sets `RestoreLockedMode` to true
- **WHEN** the Pinned-Dependencies raw data is collected
- **THEN** the NuGet command is recorded as pinned

### Requirement: SAST Tool Evidence

The system SHALL record, for each commit in the commit window with a merged
pull request, whether the head commit has a completed check run with a
`success` or `neutral` conclusion from `github-advanced-security`,
`github-code-scanning`, `lgtm-com`, `sonarcloud`, or `sonarqubecloud`. The
system SHALL also record workflows that use
`github/codeql-action/analyze`, any `snyk/actions/` action,
`facebook/pysa-action`, `JetBrains/qodana-action`, or
`hadolint/hadolint-action`, and every `pom.xml` that sets
`sonar.host.url`, with the URL and line range.

#### Scenario: CodeQL workflow

- **GIVEN** a workflow step using `github/codeql-action/analyze@v3`
- **WHEN** the SAST raw data is collected
- **THEN** the workflow is recorded as a CodeQL workflow

#### Scenario: Failed scan does not count

- **GIVEN** a merged pull request whose `github-code-scanning` check run
  concluded `failure`
- **WHEN** the SAST raw data is collected
- **THEN** that commit is recorded as not checked by a SAST tool

### Requirement: SBOM Detection

The system SHALL record SBOM files that match `.cdx.json`, `.cdx.xml`,
`.spdx`, `.spdx.json`, `.spdx.xml`, `.spdx.yaml`, `.spdx.yml`, `.spdx.rdf`,
or `.spdx.rdf.xml` (case-insensitive), both as top-level repository files
that do not start with a dot and as release assets. For releases, the
system SHALL inspect only the 5 most recent releases and record at most one
SBOM asset per release. Experimental gating of the SBOM check is defined in
`check-registry`.

#### Scenario: Top-level SBOM

- **GIVEN** a repository with `sbom.spdx.json` at the root
- **WHEN** the SBOM raw data is collected
- **THEN** `sbom.spdx.json` is recorded as a source SBOM

#### Scenario: Nested SBOM ignored

- **GIVEN** a repository with only `docs/sbom.cdx.json`
- **WHEN** the SBOM raw data is collected
- **THEN** no source SBOM is recorded

### Requirement: Security Policy File Location

The system SHALL look, case-insensitively, for `SECURITY.md`,
`SECURITY.markdown`, `SECURITY.adoc`, or `SECURITY.rst` at the repository
root, in `.github/`, or in `docs/`, and for `SECURITY.rst` in `doc/`. It
SHALL record the first match only. When none is found, the system SHALL
look in the same locations in the organization's `.github` repository and
record a match there as a URL. An unreachable or unsupported organization
repository SHALL be treated as having no policy.

#### Scenario: Policy in docs folder

- **GIVEN** a repository containing `docs/security.md`
- **WHEN** the Security-Policy raw data is collected
- **THEN** `docs/security.md` is recorded as the policy file

#### Scenario: Organization-level policy

- **GIVEN** a repository with no policy file
- **AND** its organization's `.github` repository contains `SECURITY.md`
- **WHEN** the Security-Policy raw data is collected
- **THEN** the organization's `SECURITY.md` is recorded as a URL

### Requirement: Security Policy Content Analysis

For each recorded policy file, the system SHALL record the file size and,
line by line, every `http` or `https` URL, every email address (including
the `[at]` form), and every text hit: a one- to four-digit number or the
case-insensitive strings `Disclos` or `Vuln`. Each hit SHALL include its
line number and offset within the line. An empty file SHALL produce no
hits.

#### Scenario: Policy with contact details

- **GIVEN** a policy containing `Report vulnerabilities to sec@example.com`
- **WHEN** the Security-Policy raw data is collected
- **THEN** an email hit and a `Vuln` text hit are recorded on that line

### Requirement: Release and Provenance Data

The system SHALL list the repository's releases with their assets and look
up the project's package versions in deps.dev, recording each version's
package system, name, version, and first SLSA provenance with its commit
and verification status. A failed deps.dev lookup SHALL be logged and
treated as having no packages. A failure to list releases SHALL cause an
error.

#### Scenario: deps.dev unavailable

- **GIVEN** deps.dev returns an error for the repository
- **WHEN** the Signed-Releases raw data is collected
- **THEN** releases are recorded and the package list is empty

### Requirement: Token Permission Declarations

For each workflow, the system SHALL count the workflow and record its
top-level permissions and each job's permissions. A missing `permissions`
block SHALL be recorded as undeclared. An empty block SHALL be recorded as
`none`. `read-all` and `read` SHALL be recorded as read, `none` as none,
and any other value for the whole token, such as `write-all`, as write. A
scope set to `write` SHALL be recorded as write when it is `statuses`,
`checks`, `security-events`, `deployments`, `contents`, `packages`, or
`actions`, and as unknown otherwise. Each entry SHALL include the file,
line, location (top or job), and scope name.

#### Scenario: Undeclared top-level permissions

- **GIVEN** a workflow with no top-level `permissions` key
- **WHEN** the Token-Permissions raw data is collected
- **THEN** an undeclared top-level permission is recorded for the workflow

#### Scenario: Write-all at the top level

- **GIVEN** a workflow with `permissions: write-all`
- **WHEN** the Token-Permissions raw data is collected
- **THEN** a top-level write permission is recorded with value `write-all`

### Requirement: Token Permission Exemptions

At the job level, the system SHALL not treat `packages: write` as write in a
packaging workflow. It SHALL not treat `contents: write` as write in a
releasing workflow (python-semantic-release, semantic-release through
`npx`, `pnpm`, or `yarn`, `actions/setup-go` with GoReleaser, the SLSA Go or
generic generator, or `mvn release:prepare`) or in a
`peaceiris/actions-gh-pages` deployment. It SHALL not treat
`security-events: write` as write in a workflow that uses
`github/codeql-action/analyze`, `github/codeql-action/upload-sarif`,
`ossf/scorecard-action`, `haskell-actions/hlint-scan`, or
`zizmorcore/zizmor-action`. Exempted write scopes SHALL be recorded at the
unknown level, and each workflow SHALL also get informational unknown-level
entries noting which of these classifications matched.

#### Scenario: SARIF upload job

- **GIVEN** a job with `security-events: write` that uses
  `github/codeql-action/upload-sarif`
- **WHEN** the Token-Permissions raw data is collected
- **THEN** the scope is not recorded as a write permission

#### Scenario: Same scope at the top level

- **GIVEN** the same workflow declares `security-events: write` at the top
  level
- **WHEN** the Token-Permissions raw data is collected
- **THEN** a top-level write permission is recorded

### Requirement: Vulnerability Data

The system SHALL run the OSV scanner against the local checkout of the
repository and against the latest commit SHA, unless the commit list is
unavailable or every commit lacks a SHA. The system SHALL record the
unfixed vulnerabilities it reports. When the scanner finds no packages,
the system SHALL record no vulnerabilities instead of failing.

#### Scenario: Repository without manifests

- **GIVEN** a repository with no dependency manifests
- **WHEN** the Vulnerabilities raw data is collected
- **THEN** an empty vulnerability list is recorded without error

### Requirement: Webhook Data

The system SHALL list the repository's webhooks and record, for each, its
identity and whether it uses an authentication secret. When run as a check,
webhook data SHALL be collected only if the `SCORECARD_EXPERIMENTAL`
environment variable is set.

#### Scenario: Webhooks listed

- **GIVEN** a repository with two webhooks, one with a secret
- **AND** `SCORECARD_EXPERIMENTAL` is set
- **WHEN** the Webhooks raw data is collected
- **THEN** both webhooks are recorded with their secret status
