# repo-clients Specification

## Purpose

Repository clients give Scorecard checks a uniform, read-only view of a
repository regardless of where it is hosted: GitHub (including GitHub
Enterprise), GitLab (including self-managed instances), Azure DevOps, a local
directory, or a plain git clone. Auxiliary clients answer questions that live
outside the repository itself: OSS-Fuzz enrollment, known vulnerabilities from
OSV, and OpenSSF Best Practices badge level.

## Requirements

### Requirement: Repository URI Resolution Order

The CLI SHALL resolve a repository argument by trying GitHub first, then Azure
DevOps (only when the `SCORECARD_EXPERIMENTAL` environment variable is set),
then GitLab, and SHALL use the first platform that accepts the URI. When no
platform accepts it, the CLI MUST return an error that combines every
platform's parse error.

#### Scenario: GitHub URI is accepted first

- **WHEN** the user runs Scorecard with `--repo=github.com/ossf/scorecard`
- **THEN** the repository is treated as a GitHub repository
- **AND** GitLab and Azure DevOps parsing are not attempted

#### Scenario: Non-GitHub host falls through to GitLab

- **GIVEN** `GH_HOST` is unset
- **WHEN** the user runs Scorecard with `--repo=gitlab.com/gitlab-org/gitlab`
- **THEN** GitHub parsing fails with an unsupported host error
- **AND** the repository is treated as a GitLab repository

#### Scenario: Unparseable URI

- **WHEN** the repository argument is rejected by every platform
- **THEN** the CLI fails with an error stating it is unable to parse the URI
  as github, gitlab, or azuredevops, including each underlying error

#### Scenario: Azure DevOps URI without the experimental flag

- **GIVEN** `SCORECARD_EXPERIMENTAL` is unset
- **WHEN** the repository argument is a `dev.azure.com` URI
- **THEN** Azure DevOps parsing is not attempted

### Requirement: GitHub Repository URI Parsing

The GitHub client SHALL accept `owner/repo` shorthand and full URIs with or
without a scheme, defaulting the scheme to `https`. Shorthand MUST resolve to
the host in `GH_HOST` when set, otherwise `github.com`. The host MUST equal
`github.com` or the value of `GH_HOST`; any other host MUST be rejected with an
unsupported host error. A URI without both an owner and a repository segment
MUST be rejected with an invalid URL error.

#### Scenario: Shorthand defaults to github.com

- **GIVEN** `GH_HOST` is unset
- **WHEN** the input is `foo/kubeflow`
- **THEN** the repository host is `github.com`, owner `foo`, repo `kubeflow`

#### Scenario: Trailing slash is ignored

- **WHEN** the input is `https://github.com/foo/kubeflow/`
- **THEN** the owner is `foo` and the repo is `kubeflow`

#### Scenario: Shorthand with an enterprise host

- **GIVEN** `GH_HOST=github.corp.com`
- **WHEN** the input is `corpfoo/kubeflow`
- **THEN** the repository host is `github.corp.com`

#### Scenario: Host other than GitHub

- **GIVEN** `GH_HOST` is unset
- **WHEN** the input is `https://gitlab.com/foo/kubeflow`
- **THEN** GitHub parsing fails with an unsupported host error

### Requirement: GitHub Enterprise Endpoints

When `GH_HOST` is set to a value other than `github.com`, the GitHub client
SHALL send REST requests to `https://<GH_HOST>/api/v3` and GraphQL requests to
`https://<GH_HOST>/api/graphql`. The client's reported repository URI MUST use
`GH_HOST` as its host.

#### Scenario: Enterprise API base URLs

- **GIVEN** `GH_HOST=github.corp.com`
- **WHEN** the GitHub client is created
- **THEN** REST calls target `https://github.corp.com/api/v3`
- **AND** GraphQL calls target `https://github.corp.com/api/graphql`

#### Scenario: Default public endpoints

- **GIVEN** `GH_HOST` is unset or equals `github.com`
- **WHEN** the GitHub client is created
- **THEN** calls target the public github.com API endpoints

### Requirement: GitHub Authentication

The GitHub client SHALL authenticate using the first non-empty token found in
`GITHUB_AUTH_TOKEN`, `GITHUB_TOKEN`, `GH_TOKEN`, `GH_AUTH_TOKEN`, in that
order, sent as a `Bearer` authorization header. A token value containing
commas MUST be split into multiple tokens used in round-robin order. When no
token variable is set, the client SHALL obtain tokens from the RPC token server
at `GITHUB_AUTH_SERVER`, and failing that SHALL authenticate as a GitHub App
using `GITHUB_APP_KEY_PATH`, `GITHUB_APP_ID`, and `GITHUB_APP_INSTALLATION_ID`.
When no credentials are configured, the client MUST log an error referring to
the authentication documentation and continue without credentials.

#### Scenario: Token precedence

- **GIVEN** `GITHUB_AUTH_TOKEN=a` and `GITHUB_TOKEN=b`
- **WHEN** a GitHub API request is made
- **THEN** token `a` is used
- **AND** a warning states that the two variables differ and the former is
  used

#### Scenario: Multiple comma-separated tokens

- **GIVEN** `GITHUB_AUTH_TOKEN=t1,t2`
- **WHEN** successive GitHub API requests are made
- **THEN** the tokens are rotated in round-robin order

#### Scenario: GitHub App credentials

- **GIVEN** no token variable or `GITHUB_AUTH_SERVER` is set
- **AND** `GITHUB_APP_KEY_PATH`, `GITHUB_APP_ID`, and
  `GITHUB_APP_INSTALLATION_ID` are set
- **WHEN** a GitHub API request is made
- **THEN** the request is authenticated as the GitHub App installation

#### Scenario: No credentials

- **GIVEN** none of the GitHub credential variables are set
- **WHEN** the GitHub client is created
- **THEN** an error is logged saying the GitHub token env var is not set
- **AND** requests proceed unauthenticated

### Requirement: GitHub Rate Limit Handling

The GitHub client SHALL wait and retry a request when the response carries an
integer `Retry-After` header, sleeping for that many seconds. When
`X-RateLimit-Remaining` is zero or less, the client SHALL sleep until the time
given by `X-RateLimit-Reset` and then retry. Retries MUST NOT be capped by an
attempt count. Responses with missing or non-numeric rate limit headers MUST be
returned unchanged.

#### Scenario: Retry-After honored

- **GIVEN** the first response is HTTP 429 with `Retry-After: 1`
- **AND** the second response is HTTP 200
- **WHEN** the request is sent
- **THEN** the client waits one second, retries, and returns the HTTP 200
  response after exactly two requests

#### Scenario: Rate limit exhausted

- **GIVEN** a response with `X-RateLimit-Remaining: 0` and a future
  `X-RateLimit-Reset`
- **WHEN** the request is sent
- **THEN** the client logs that the rate limit was exceeded, waits until the
  reset time, and retries

### Requirement: Repository Initialization

Each remote client SHALL verify the repository exists during initialization
and MUST fail with a repository-unreachable error when the platform API cannot
return it. After initialization the client SHALL use the canonical owner,
name, and default branch reported by the platform. When initialization fails,
the scan of that repository SHALL stop before any check or probe runs and the
initialization error SHALL be returned in place of a result.

#### Scenario: Missing GitHub repository

- **WHEN** the GitHub repository lookup fails
- **THEN** initialization fails with a repo unreachable error

#### Scenario: GitLab repository with repository access disabled

- **GIVEN** the GitLab project's repository access level is disabled
- **WHEN** the GitLab client is initialized
- **THEN** initialization fails with a repo unreachable error

#### Scenario: Azure DevOps repository lookup fails

- **WHEN** the Azure DevOps repository cannot be fetched
- **THEN** initialization fails with a "could not get repository" error

#### Scenario: Initialization failure ends the scan

- **GIVEN** a repository whose client initialization fails
- **WHEN** Scorecard scans it
- **THEN** no check or probe runs for that repository
- **AND** the initialization error is returned instead of a result

### Requirement: Commit Depth

The GitHub, GitLab, Azure DevOps, and local directory clients SHALL treat a
commit depth of zero or less as 30. Commit listing SHALL return at most the
configured number of commits, starting from the requested commit (or the
default branch head when the commit is `HEAD`). The GitHub client SHALL
request commits in pages of at most 100 and, when the API returns
`502 Bad Gateway`, SHALL halve the page size and retry up to three times.

#### Scenario: Default depth

- **WHEN** a client is initialized with commit depth 0
- **THEN** commit listing returns at most 30 commits

#### Scenario: Custom depth via the CLI

- **WHEN** the user passes `--commit-depth=50`
- **THEN** commit listing returns at most 50 commits

#### Scenario: GitHub gateway timeout

- **GIVEN** a GitHub GraphQL commit query fails with `502 Bad Gateway`
- **WHEN** fewer than three retries have been made
- **THEN** the client halves the number of commits requested and retries

### Requirement: RepoClient Interface Contract

Every repository client SHALL implement the RepoClient library interface. A
method the platform cannot serve MUST return an error that wraps the exported
unsupported feature error so callers can detect it. The interface covers:

- initialization for a repository, a commit (`HEAD` means the default
  branch), and a commit depth
- the repository URI, archived status, and creation time
- file listing, file reading, and the local filesystem path of the files
- a named branch, the default branch, and the default branch name
- a client for the owner's `.github` repository
- commits, issues, licenses, releases, contributors, and webhooks
- successful workflow runs for a workflow file
- check runs and statuses for a ref
- programming languages
- code search and commit search by author
- releasing resources on close

#### Scenario: Caller detects unsupported operation

- **WHEN** a check calls a method that the client does not support
- **THEN** the returned error matches the unsupported feature error

#### Scenario: Empty search query

- **WHEN** code search is called with an empty query on GitHub, GitLab, Azure
  DevOps, or the git client
- **THEN** an error stating the search query is empty is returned

### Requirement: GitHub HEAD-Only Operations

When initialized for a commit other than `HEAD`, the GitHub client MUST return
an unsupported feature error from issue listing, archived status, branch
lookups, webhook listing, release listing, contributor listing, workflow run
listing, code search, and commit search. File access and commit listing SHALL
remain available for any commit.

#### Scenario: Releases at a pinned commit

- **GIVEN** the GitHub client was initialized with a specific commit SHA
- **WHEN** releases are listed
- **THEN** the error states ListReleases is only supported for HEAD queries

#### Scenario: Files at a pinned commit

- **GIVEN** the GitHub client was initialized with a specific commit SHA
- **WHEN** files are listed
- **THEN** the files reflect the repository at that commit

### Requirement: Repository File Retrieval

Remote clients SHALL fetch repository contents once, on first file access, as
an archive extracted to a temporary directory: a gzipped tarball for GitHub
and GitLab, and a zip for Azure DevOps. The GitHub client SHALL instead
perform a shallow, single-branch, tagless git clone of depth 1 when git file
mode is selected (`--file-mode=git` in the CLI, `WithFileModeGit` in the Go
API); otherwise it SHALL use the archive. Other clients ignore the file mode.
Closing the GitHub and Azure DevOps clients SHALL remove the temporary
directory. Validation of the `--file-mode` value is defined in `cli`.

#### Scenario: Default archive mode

- **WHEN** a GitHub repository is scanned without `--file-mode`
- **THEN** the repository tarball for the requested commit is downloaded

#### Scenario: Git file mode

- **WHEN** a GitHub repository is scanned with `--file-mode=git`
- **THEN** files are read from a depth-1 clone of the repository

### Requirement: Archive Extraction Safety

When extracting a tarball, the client SHALL strip the archive's top-level
directory, skip symlinks and zero-length files, and MUST abort extraction when
an entry path would resolve outside the destination directory. A tarball that
returns HTTP 400 or 404, or that is corrupted, SHALL be logged and skipped so
the repository appears to have no files, rather than failing the scan. When
extracting an Azure DevOps zip, the client MUST abort on paths outside the
destination or on any entry larger than 100 MB.

#### Scenario: Path traversal entry

- **WHEN** a tarball entry resolves outside the extraction directory
- **THEN** extraction fails with a ZipSlip path error

#### Scenario: Tarball not found

- **WHEN** the tarball download returns HTTP 404
- **THEN** a message "unable to get tarball ... Skipping..." is logged
- **AND** file listing returns no files without an error

#### Scenario: Missing tarball for an org .github repository

- **WHEN** the tarball for a repository named `.github` is not found
- **THEN** no message is logged

#### Scenario: Oversized zip entry

- **WHEN** an Azure DevOps zip entry exceeds 100 MB
- **THEN** extraction fails with a "file too large, possible zip bomb" error

### Requirement: GitHub Branch Protection Data

The GitHub client SHALL read repository rulesets, which require only read
access, and SHALL attempt to read classic branch protection rules, which
require admin access. A permissions error on classic branch protection MUST be
ignored when at least one active ruleset exists, and otherwise MUST be
returned with a message that some GitHub tokens cannot read classic branch
protection rules.

#### Scenario: Non-admin token with rulesets

- **GIVEN** the token cannot read classic branch protection
- **AND** the repository has an active ruleset
- **WHEN** the default branch is requested
- **THEN** the branch is returned with ruleset-derived protection data

#### Scenario: Non-admin token without rulesets

- **GIVEN** the token cannot read classic branch protection
- **AND** the repository has no rulesets
- **WHEN** the default branch is requested
- **THEN** an error referencing the fine-grained token documentation is
  returned

### Requirement: GitHub Contributors

The GitHub client SHALL build the contributor list from the repository
contributors API merged with code owners from the first readable file among
`CODEOWNERS`, `.github/CODEOWNERS`, and `docs/CODEOWNERS`. Team owners SHALL
be expanded to their members when the token permits, and code owner expansion
MUST stop at 100 owners. Each contributor SHALL include organization
memberships (errors ignored) and company from the user profile.

#### Scenario: Code owner who is also a contributor

- **GIVEN** a user appears in both the contributors API and `CODEOWNERS`
- **WHEN** contributors are listed
- **THEN** that user appears once, marked as a code owner

#### Scenario: Unparseable CODEOWNERS

- **WHEN** the CODEOWNERS file cannot be parsed
- **THEN** a message is logged and contributors are returned without code
  owner data

### Requirement: GitHub Organization Repository

The GitHub client SHALL provide a client for the `<owner>/.github` repository
at `HEAD`, reusing the same transport and file mode. GitLab, Azure DevOps,
local directory, and git clients MUST return an unsupported feature error for
this operation.

#### Scenario: Org repository on GitHub

- **WHEN** a check requests the org repository client for `ossf/scorecard`
- **THEN** a client initialized for `ossf/.github` is returned

#### Scenario: Org repository on GitLab

- **WHEN** a check requests the org repository client for a GitLab project
- **THEN** an unsupported feature error is returned

### Requirement: GitLab Repository URI Parsing

The GitLab client SHALL require a host in the URI and MUST reject `owner/repo`
shorthand. The first path segment SHALL be the owner and the remainder the
project. When `GL_HOST` is set and its host matches the URI host, any path
prefix in `GL_HOST` SHALL be treated as part of the host. A URI SHALL be
accepted when its host contains `gitlab.`, MUST be rejected when its host is
`github.com`, and otherwise SHALL be accepted only if an unauthenticated
request to the instance's project list API returns HTTP 200.

#### Scenario: Shorthand rejected

- **WHEN** the input is `ossf/scorecard`
- **THEN** GitLab parsing fails stating a gitlab repo must specify host

#### Scenario: GL_HOST with a path prefix

- **GIVEN** `GL_HOST=https://foo.com/gitlab/`
- **WHEN** the input is `foo.com/gitlab/ssdlc/scorecard-scanner`
- **THEN** the host is `foo.com/gitlab`, owner `ssdlc`, project
  `scorecard-scanner`

#### Scenario: Self-managed instance probe

- **WHEN** the input host does not contain `gitlab.` and is not `github.com`
- **AND** the instance's project list API does not return HTTP 200
- **THEN** parsing fails with a repo unreachable error

#### Scenario: GitHub project named like GitLab

- **WHEN** the input is `http://github.com/foo/gitlab.test`
- **THEN** GitLab parsing fails because the host is not a GitLab host

### Requirement: GitLab Authentication and Endpoints

The GitLab client SHALL authenticate with the token in `GITLAB_AUTH_TOKEN`,
using `https://<host>` as the API base URL, `https://<host>/api/graphql` for
GraphQL, and the `PRIVATE-TOKEN` header for archive downloads. An unset token
SHALL result in unauthenticated requests.

#### Scenario: Token set

- **GIVEN** `GITLAB_AUTH_TOKEN` is set
- **WHEN** the GitLab client downloads the repository archive
- **THEN** the request includes the token in the `PRIVATE-TOKEN` header

### Requirement: GitLab Platform Behavior

The GitLab client SHALL download the archive from
`https://<host>/api/v4/projects/<id>/repository/archive.tar.gz` and SHALL add
a synthetic file `gitlabscorecard_flattened_ci.yaml` containing the merged CI
configuration from the project's CI lint API. It SHALL report an empty local
path, and closing it SHALL NOT remove downloaded files. Contributor listing,
branch lookups, release listing, code search, and commit search MUST return an
unsupported feature error when the commit is not `HEAD`. When listing commits
for a non-`HEAD` commit, the client SHALL warn on stderr that merge requests
may be missing. Creating an OSS-Fuzz client for GitLab MUST return an
unsupported feature error.

#### Scenario: Merged CI file

- **WHEN** files are listed for a GitLab project
- **THEN** the list includes `gitlabscorecard_flattened_ci.yaml`

#### Scenario: Pipeline status mapping

- **WHEN** check runs are listed for a ref
- **THEN** GitLab pipeline statuses are mapped to check run statuses, for
  example `running` to `in_progress` and `failed` to `completed` with
  conclusion `failure`

### Requirement: Azure DevOps Repository URI Parsing

The Azure DevOps client SHALL accept, with or without an `https` scheme,
`dev.azure.com/<org>/<project>/_git/<repo>`,
`<org>.visualstudio.com/<project>/_git/<repo>`, and
`<org>.visualstudio.com/DefaultCollection/<project>/_git/<repo>`, and SHALL
normalize all forms to host `dev.azure.com`. A non-`https` scheme, an explicit
port, a wrong segment count, or a missing `_git` segment MUST be rejected.
When the URI's host is an Azure DevOps host but parsing fails, the CLI MUST
report the Azure DevOps error instead of falling through to GitLab.

#### Scenario: visualstudio.com form

- **GIVEN** `SCORECARD_EXPERIMENTAL` is set
- **WHEN** the input is `contoso.visualstudio.com/proj/_git/repo`
- **THEN** the repository URI is `dev.azure.com/contoso/proj/_git/repo`

#### Scenario: Malformed Azure DevOps URI

- **GIVEN** `SCORECARD_EXPERIMENTAL` is set
- **WHEN** the input is `dev.azure.com/contoso/proj/repo`
- **THEN** the CLI fails with "unable to parse Azure DevOps repository URI"

### Requirement: Azure DevOps Platform Behavior

The Azure DevOps client SHALL authenticate with the personal access token in
`AZURE_DEVOPS_AUTH_TOKEN` against `https://dev.azure.com/<org>`. It SHALL
report archived status from the repository's disabled flag, releases from git
tags, issues from the project's work items, and builds for the given YAML file
as workflow runs. Repository creation time SHALL come from the audit log,
falling back to the first commit date when the audit log is inaccessible or
empty. License listing and the org repository client MUST return an
unsupported feature error.

#### Scenario: Audit log inaccessible

- **WHEN** the audit log cannot be read
- **THEN** the creation time is the date of the first commit

#### Scenario: Licenses on Azure DevOps

- **WHEN** licenses are listed
- **THEN** an unsupported feature error is returned

### Requirement: Local Directory Client

The local directory client SHALL accept a path only if it exists and is a
directory, and SHALL report its URI as `file://<path>` with an empty host. It
SHALL list files recursively relative to that path, skipping `.git`
directories and dangling symlinks, and MUST refuse to open a file whose
resolved path escapes the directory. Archived status, branches, default
branch, commits, issues, releases, contributors, workflow runs, check runs,
statuses, webhooks, search, commit search, licenses, creation time, and the
org repository client MUST return an unsupported feature error. Programming
language listing SHALL return a single placeholder entry meaning all
languages. Local scans SHALL NOT use OSS-Fuzz or Best Practices clients.

#### Scenario: Path is a file

- **WHEN** `--local` points at a regular file
- **THEN** repository creation fails with a not a directory error

#### Scenario: Escaping symlink

- **GIVEN** a symlink inside the directory points outside it
- **WHEN** the file is opened through the client
- **THEN** opening fails

#### Scenario: Commits unavailable

- **WHEN** commits are listed for a local directory
- **THEN** an unsupported feature error is returned

### Requirement: Git Client

The git client SHALL clone the repository into a temporary directory,
prefixing `https://` when no `https://` or `ssh://` scheme is present and
appending `.git` when absent; a `file://` URI SHALL be copied and opened
instead. A non-`HEAD` commit SHALL be checked out after cloning. The client
SHALL list commits in committer-time order up to the commit depth, search with
the query as a regular expression over the worktree, report branches as
unprotected, and derive creation time from history. Archived status, default
branch details, org client, issues, licenses, releases, contributors, workflow
runs, check runs, statuses, webhooks, languages, and commit search MUST return
an unsupported feature error.

#### Scenario: Regex search with filename

- **WHEN** search is called with a query and a filename
- **THEN** results include paths whose contents match the query regex and
  whose path ends with the filename

### Requirement: OSS-Fuzz Client

The OSS-Fuzz client SHALL download the OSS-Fuzz status file (default
`https://oss-fuzz-build-logs.storage.googleapis.com/status.json`), normalize
each project's `main_repo` to lowercase `host/owner/repo` with any `.git`
suffix removed, and answer a search whose query is a repository URI with one
hit when the lowercased query exactly matches a normalized entry and zero hits
otherwise. A status file that returns HTTP 400 or higher or cannot be parsed
MUST cause search to return an error. All other repository methods MUST return
an unsupported feature error.

#### Scenario: Enrolled project, case-insensitive

- **GIVEN** the status file lists `https://github.com/Foo/Bar.git`
- **WHEN** search is called with query `github.com/foo/bar`
- **THEN** the response reports one hit

#### Scenario: Substring is not a match

- **GIVEN** the status file lists `github.com/foo/barbaz`
- **WHEN** search is called with query `github.com/foo/bar`
- **THEN** the response reports zero hits

#### Scenario: Status file unavailable

- **WHEN** the status file request returns HTTP 404
- **THEN** search returns an error

### Requirement: Vulnerabilities Client

The vulnerabilities client SHALL scan the repository's local path recursively
and its latest commit using OSV-Scanner and SHALL return each unfixed
vulnerability's ID and aliases, deduplicated by ID. It MUST ignore Go
standard library findings and Maven findings whose version is an unresolved
placeholder containing `@` or `${`. Finding no packages MUST yield an empty
result rather than an error. A scanner panic MUST be converted into an
internal error. The CLI SHALL send the user agent `scorecard-cli/<version>`.

#### Scenario: No packages detected

- **WHEN** OSV-Scanner reports that no packages were found
- **THEN** an empty vulnerability list is returned without error

#### Scenario: Duplicate findings

- **WHEN** the same vulnerability ID is found in two manifests
- **THEN** the ID appears once in the result

#### Scenario: Maven template placeholder

- **WHEN** a Maven package version is `@project.version@`
- **THEN** findings for that package are excluded

### Requirement: Best Practices Badge Client

The Best Practices client SHALL query
`<base>/projects.json?url=https://<repo-uri>`, where the base URL is
`CII_BEST_PRACTICES_URL` or `https://www.bestpractices.dev` by default. An
empty result list MUST map to not found. Otherwise the first result's
`badge_level` SHALL map to in progress, passing, silver, or gold by substring
match, and any other value MUST map to unknown with an unsupported badge
error. HTTP 429 responses SHALL be retried with exponential backoff (1, 2,
then 4 seconds) for up to three attempts before failing. A bucket-backed
variant SHALL read `<repo-uri>/result.json` from a blob bucket and report not
found when the object is absent.

#### Scenario: Passing project

- **WHEN** the API returns `[{"badge_level":"passing"}]`
- **THEN** the badge level is passing

#### Scenario: Unknown project

- **WHEN** the API returns `[]`
- **THEN** the badge level is not found

#### Scenario: Persistent rate limiting

- **WHEN** all three attempts return HTTP 429
- **THEN** the request fails with an error stating it failed after
  exponential backoff
