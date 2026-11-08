# public-go-api Specification

## Purpose

The `pkg/scorecard` package lets Go programs run Scorecard without the CLI.
This spec covers the exported `Run` function, its functional options and
defaults, how it creates repository and service clients, how it selects
checks or probes, the shape of the `Result` it returns, and the errors it
returns. Check selection rules, probe execution, client behavior, and
result encoding are defined in their own capabilities.

## Requirements

### Requirement: Repository-Driven Run

The system SHALL export
`Run(ctx context.Context, repo clients.Repo, opts ...Option) (Result, error)`.
`Run` SHALL call `InitRepo` on the selected repository client with the
configured commit and commit depth, SHALL close that client before
returning once `InitRepo` has succeeded, and SHALL return a zero `Result`
whenever it returns a non-nil error.

#### Scenario: Valid repository is supplied

- **GIVEN** a caller created a `*githubrepo.Repo` for `ossf/scorecard`
- **WHEN** the caller invokes `Run(ctx, repo)`
- **THEN** `Run` initializes the repository at `HEAD` and returns a
  `Result` and a nil error

#### Scenario: Run fails

- **WHEN** `Run` returns a non-nil error
- **THEN** the returned `Result` is the zero value

### Requirement: Default Run Configuration

Without options, `Run` SHALL use commit `HEAD`, log level `info`, commit
depth 0, no probes, and the default check set. Built-in repository clients
treat a commit depth of 0 as their default depth (see `repo-clients`). The
default check set SHALL be the checks returned by `checks.GetAll()`, which
include the experimental checks only when `SCORECARD_EXPERIMENTAL` is set,
filtered by the request types and repository type that `Run` requires.

#### Scenario: Caller supplies no options

- **WHEN** `Run` is invoked with only a context and a GitHub repository
- **THEN** it analyzes `HEAD` with every non-experimental check that
  supports GitHub repositories
- **AND** it runs no probes

#### Scenario: Experimental checks enabled

- **GIVEN** `SCORECARD_EXPERIMENTAL` is set
- **WHEN** `Run` is invoked with no options
- **THEN** the default check set also includes the experimental checks
  that apply to the repository

### Requirement: Functional Option Overrides

The system SHALL provide the options `WithCommitSHA`, `WithCommitDepth`,
`WithChecks`, `WithProbes`, `WithLogLevel`, and `WithFileModeGit`. Each
option SHALL modify only the configuration of the `Run` call it is passed
to. `WithCommitDepth` SHALL pass its value to `InitRepo` without
validation. `WithLogLevel` SHALL set the level of the logger `Run` passes
to the default local-directory client. `WithFileModeGit` SHALL take effect
only when `Run` creates the default GitHub client, which then fetches files
with git. An empty `WithChecks` list SHALL select the default check set.

#### Scenario: Options apply per call

- **GIVEN** one `Run` call with `WithCommitSHA("<sha>")`
- **WHEN** a second `Run` call is made without options
- **THEN** the second call analyzes `HEAD`

#### Scenario: Negative commit depth

- **WHEN** `Run` is invoked with `WithCommitDepth(-5)`
- **THEN** `Run` returns no option error and passes `-5` to `InitRepo`

#### Scenario: Git file mode with a caller-supplied client

- **WHEN** `Run` is invoked with `WithFileModeGit()` and `WithRepoClient`
- **THEN** the caller-supplied client is used unchanged

### Requirement: Repository-Specific Client Creation

When no repository client is supplied with `WithRepoClient`, `Run` SHALL
create the default client for the repository type: a local-directory
client for `*localdir.Repo`, a GitHub client for `*githubrepo.Repo`, a
GitLab client for the repository host for `*gitlabrepo.Repo`, and an Azure
DevOps client for `*azuredevopsrepo.Repo`.

#### Scenario: Default GitLab client

- **WHEN** `Run` is invoked with a `*gitlabrepo.Repo` and no
  `WithRepoClient`
- **THEN** `Run` creates a GitLab client for that repository's host

#### Scenario: Local directory client

- **WHEN** `Run` is invoked with a `*localdir.Repo` and no
  `WithRepoClient`
- **THEN** `Run` creates a local-directory client

### Requirement: Default Service Clients

When not overridden, `Run` SHALL create the default OpenSSF Best Practices
client, an OSS-Fuzz client reading `ossfuzz.StatusURL`, the default OSV
vulnerabilities client, and a deps.dev project client. The deps.dev
project client SHALL have no override option. Client endpoints and
behavior are defined in `repo-clients`.

#### Scenario: No service client options

- **WHEN** `Run` is invoked without service client options
- **THEN** checks receive the default Best Practices, OSS-Fuzz, OSV, and
  deps.dev clients

### Requirement: Service Client Override Options

The system SHALL provide `WithRepoClient`, `WithOSSFuzzClient`,
`WithVulnerabilitiesClient`, and `WithOpenSSFBestPraticesClient`, each of
which replaces the corresponding client for that `Run` call. The Best
Practices option SHALL be exported under the misspelled name
`WithOpenSSFBestPraticesClient` (SC-020).

#### Scenario: Caller supplies a repository client

- **WHEN** `Run` is invoked with `WithRepoClient(c)`
- **THEN** `Run` uses `c` and creates no default repository client

#### Scenario: Caller supplies a Best Practices client

- **WHEN** `Run` is invoked with `WithOpenSSFBestPraticesClient(c)`
- **THEN** checks query `c` instead of the default Best Practices client

### Requirement: Check Selection in Run

`Run` SHALL require file-based support for a `*localdir.Repo` and
commit-based support when the commit is not `HEAD`, compared
case-insensitively. It SHALL resolve the checks to run by calling
`policy.GetEnabled` with no policy, the `WithChecks` names, those request
types, and the repository type, before initializing the repository. A
selection error MUST be returned wrapped as `getting enabled checks:`. The
selection rules are defined in `check-registry`.

#### Scenario: Unknown check name

- **WHEN** `Run` is invoked with `WithChecks([]string{"Nope"})`
- **THEN** it returns an error beginning `getting enabled checks:`
- **AND** the repository is not initialized

#### Scenario: Lowercase head

- **WHEN** `Run` is invoked with `WithCommitSHA("head")`
- **THEN** commit-based support is not required

### Requirement: Probe Mode

When `WithProbes` is given a non-empty list, `Run` SHALL run only those
probes after initializing the repository and SHALL return their findings
in `Result.Findings` with no check results and no maintainer annotation
configuration. Check selection SHALL still run first, so an invalid
`WithChecks` name fails the call even in probe mode. Probe execution is
defined in `probes`.

#### Scenario: Probes requested

- **WHEN** `Run` is invoked with `WithProbes([]string{"hasLicenseFile"})`
- **THEN** `Result.Findings` holds that probe's findings
- **AND** `Result.Checks` is empty

#### Scenario: Invalid check name in probe mode

- **WHEN** `Run` is invoked with `WithProbes([]string{"hasLicenseFile"})`
  and `WithChecks([]string{"Nope"})`
- **THEN** it returns an error beginning `getting enabled checks:`

### Requirement: Empty Repository Result

When the repository client lists no commits, `Run` SHALL return a nil
error and a `Result` that carries the identity fields with
`Repo.CommitSHA` set to the requested commit, and SHALL run no checks or
probes.

#### Scenario: Repository has no commits

- **GIVEN** a repository with no commits
- **WHEN** `Run` is invoked with no options
- **THEN** it returns a nil error
- **AND** `Result.Repo.CommitSHA` is `HEAD` and `Result.Checks` is empty

### Requirement: Result Identity Fields

`Run` SHALL set `Result.Repo.Name` to the repository URI and
`Result.Repo.CommitSHA` to the SHA of the first commit the repository
client lists, or `unknown` when the client does not support listing
commits. `Result.Scorecard.Version` and `Result.Scorecard.CommitSHA` SHALL
hold the build's version and commit, and `Result.Date` SHALL be the time
after repository initialization and before checks run. Encoding of these
fields is defined in `result-formats`.

#### Scenario: Remote repository

- **WHEN** `Run` analyzes `github.com/ossf/scorecard` at `HEAD`
- **THEN** `Result.Repo.Name` is `github.com/ossf/scorecard`
- **AND** `Result.Repo.CommitSHA` is the SHA of the latest listed commit

#### Scenario: Commit listing unsupported

- **GIVEN** a repository client whose `ListCommits` reports an unsupported
  feature
- **WHEN** `Run` completes
- **THEN** `Result.Repo.CommitSHA` is `unknown`

### Requirement: Formatting Separation

`Run` SHALL return a format-independent `Result` with separate fields for
repository identity, date, Scorecard identity, check results, raw results,
findings, metadata, and maintainer annotation configuration. `Run` SHALL
leave `Metadata` empty and SHALL append check results in completion order,
without sorting. Text, JSON, SARIF, probe, raw, and in-toto output SHALL be
produced only by the separate `As*` methods and `FormatResults`, whose
rules are defined in `result-formats`.

#### Scenario: Library caller inspects the result

- **WHEN** `Run` succeeds in check mode
- **THEN** the caller can read `Result.Checks`, `Result.RawResults`, and
  `Result.Findings` without calling a formatter

#### Scenario: Result order

- **WHEN** `Run` returns several check results
- **THEN** their order in `Result.Checks` is not sorted by name

### Requirement: Returned Errors

`Run` SHALL apply options in order and return the first error an option
returns, before creating any client; no exported option returns an error.
`Run` MUST return client creation errors wrapped as
`creating github client:`, `creating gitlab client:`, or
`creating azure devops client:`, and `InitRepo` errors unwrapped. A
`ListCommits` or `GetDefaultBranchName` failure other than an unsupported
feature MUST be returned as an `internal error` whose message begins
`ListCommits:` or `GetDefaultBranchName:`. A `LocalPath` failure MUST be
returned wrapped as `RepoClient.LocalPath:`.

#### Scenario: GitHub client cannot be created

- **GIVEN** the default GitHub client cannot be created
- **WHEN** `Run` is invoked with a `*githubrepo.Repo`
- **THEN** it returns an error beginning `creating github client:`

#### Scenario: Repository unreachable

- **WHEN** `InitRepo` fails
- **THEN** `Run` returns that error unchanged

#### Scenario: Default branch lookup fails

- **GIVEN** `GetDefaultBranchName` fails with an error other than an
  unsupported feature
- **WHEN** `Run` is invoked
- **THEN** it returns an `internal error` containing
  `GetDefaultBranchName:`
