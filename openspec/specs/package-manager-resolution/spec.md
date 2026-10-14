# package-manager-resolution Specification

## Purpose

Package manager resolution lets a user name a published package instead of a
repository. Scorecard looks the package up in its registry, extracts the
source repository URL from the package metadata, and scans that repository.
The supported registries are npm, PyPI, RubyGems, and NuGet.

## Requirements

### Requirement: Package Flags

The CLI SHALL accept `--npm`, `--pypi`, `--rubygems`, and `--nuget`, each
taking a package name. Exactly one of `--repo`, `--repos`, `--org`, `--npm`,
`--pypi`, `--rubygems`, `--nuget`, or `--local` MUST be set; otherwise option
validation MUST fail before any registry is queried.

#### Scenario: Package flag alone

- **WHEN** the user runs `scorecard --npm=left-pad`
- **THEN** option validation passes and the npm registry is queried

#### Scenario: Package flag combined with a repository

- **WHEN** the user runs `scorecard --npm=left-pad --repo=github.com/a/b`
- **THEN** validation fails stating exactly one of `repo`, `repos`, `org`,
  `npm`, `pypi`, `rubygems`, `nuget` or `local` must be set

### Requirement: Resolved Repository Replaces the Repo Option

When a package flag is set, the system SHALL use the repository URL resolved
from the registry as the repository to scan. The resolved URL SHALL then go
through the same repository URI resolution as a `--repo` value, so a URL that
no supported platform accepts MUST fail at that stage.

#### Scenario: Successful resolution

- **GIVEN** the npm package `foo` declares repository
  `git+https://github.com/foo/bar.git`
- **WHEN** the user runs `scorecard --npm=foo`
- **THEN** Scorecard scans `https://github.com/foo/bar`

#### Scenario: Resolved URL on an unsupported host

- **GIVEN** a RubyGems package whose `source_code_uri` is on an unsupported
  host
- **WHEN** the user runs `scorecard --rubygems=<name>`
- **THEN** the run fails because the URL cannot be parsed as a GitHub,
  GitLab, or Azure DevOps repository

### Requirement: Registry HTTP Requests

Registry lookups SHALL use unauthenticated HTTP GET requests with a 10-second
timeout. A transport failure MUST produce an internal error naming the
registry lookup that failed.

#### Scenario: Registry unreachable

- **WHEN** the npm registry request times out
- **THEN** the lookup fails with an error stating it failed to get the npm
  package json

### Requirement: npm Resolution

For `--npm`, the system SHALL request
`https://registry.npmjs.org/<name>/latest` and read `repository.url` from the
response. It SHALL strip a trailing `.git` and a leading `git+` from the URL
and return the result without further host validation. The lookup MUST fail
when the response is not valid JSON, when the status is HTTP 404, or when
`repository.url` is empty.

#### Scenario: HTTPS repository URL

- **GIVEN** `repository.url` is `git+https://github.com/pulumi/pulumi.git`
- **WHEN** the package is resolved
- **THEN** the result is `https://github.com/pulumi/pulumi`

#### Scenario: SSH repository URL

- **GIVEN** `repository.url` is `git+ssh://git@github.com/stevemao/left-pad.git`
- **WHEN** the package is resolved
- **THEN** the result is `ssh://git@github.com/stevemao/left-pad`

#### Scenario: Package without a repository

- **WHEN** the response has no `repository.url`
- **THEN** the lookup fails stating it could not find the source repo for the
  npm package

#### Scenario: Empty response body

- **WHEN** the registry returns an empty body
- **THEN** the lookup fails stating it failed to parse the npm package json

### Requirement: PyPI Resolution

For `--pypi`, the system SHALL request `https://pypi.org/pypi/<name>/json`
and consider every value in `info.project_urls` plus `info.project_url` as a
candidate. A candidate SHALL be accepted when it matches
`http(s)://github.com/<owner>/<repo>`, `http(s)://<owner>.github.io/<repo>`,
or `http(s)://gitlab.com/<owner>/<repo>`. Accepted candidates SHALL be
normalized to lowercase `https://github.com/<owner>/<repo>` or
`https://gitlab.com/<owner>/<repo>`. GitHub repository names MUST have a
trailing `.git` removed, and GitHub URLs whose owner is `sponsors` MUST be
ignored. The lookup MUST fail when no candidate is accepted or when accepted
candidates normalize to more than one distinct repository.

#### Scenario: Repository in project_url

- **GIVEN** `info.project_url` is `https://github.com/htaslan/color/`
- **WHEN** the package is resolved
- **THEN** the result is `https://github.com/htaslan/color`

#### Scenario: Equivalent URLs are deduplicated

- **GIVEN** `project_urls` contains `https://github.com/htaslan/color/`,
  `http://HTASLAN.github.io/cOLOr`, and
  `https://github.com/htaslan/color.git`
- **WHEN** the package is resolved
- **THEN** the result is `https://github.com/htaslan/color`

#### Scenario: GitHub Sponsors link ignored

- **GIVEN** `project_urls` contains `https://github.com/sponsors/htaslan` and
  `https://github.com/htaslan/color/`
- **WHEN** the package is resolved
- **THEN** the result is `https://github.com/htaslan/color`

#### Scenario: Conflicting repositories

- **GIVEN** `project_urls` contains `https://github.com/htaslan/color/` and
  `https://gitlab.com/htaslan/color`
- **WHEN** the package is resolved
- **THEN** the lookup fails stating it found too many possible source repos
  for the pypi package

#### Scenario: No recognizable repository

- **GIVEN** no candidate URL is on github.com, github.io, or gitlab.com
- **WHEN** the package is resolved
- **THEN** the lookup fails stating it could not find the source repo for the
  pypi package

### Requirement: RubyGems Resolution

For `--rubygems`, the system SHALL request
`https://rubygems.org/api/v1/gems/<name>.json` and return `source_code_uri`
from the response unchanged. The lookup MUST fail when the response is not
valid JSON or when `source_code_uri` is empty.

#### Scenario: Gem with a source URI

- **GIVEN** the response has `source_code_uri` set to
  `https://github.com/rails/rails`
- **WHEN** the gem is resolved
- **THEN** the result is `https://github.com/rails/rails`

#### Scenario: Gem without a source URI

- **WHEN** the response has an empty `source_code_uri`
- **THEN** the lookup fails stating it could not find the source repo for the
  ruby gem

### Requirement: NuGet Service Discovery

For `--nuget`, the system SHALL first request the NuGet service index at
`https://api.nuget.org/v3/index.json` and read the `@id` of the resources
whose `@type` is `PackageBaseAddress/3.0.0` and `RegistrationsBaseUrl/3.6.0`.
The lookup MUST fail when the index cannot be fetched or parsed, or when
either resource is missing.

#### Scenario: Index unreachable

- **WHEN** the service index request fails
- **THEN** the lookup fails stating it failed to get the nuget index json

#### Scenario: Missing resource type

- **WHEN** the index has no `RegistrationsBaseUrl/3.6.0` resource
- **THEN** the lookup fails stating it failed to find that URI in the nuget
  index json

### Requirement: NuGet Latest Listed Version

The system SHALL lowercase the package name and request
`<RegistrationsBaseUrl><name>/index.json`. It SHALL walk registration pages
from last to first and entries within a page from last to first. A page whose
entries are not inlined SHALL be fetched from the page's `@id`. The system
SHALL select the first entry that is listed and has no prerelease suffix,
treating a missing `listed` field as listed. Build metadata after `+` MUST be
ignored when parsing the version. The lookup MUST fail when no entry
qualifies.

#### Scenario: Prerelease skipped

- **GIVEN** the newest entry is `2.0.0-beta` and the next is `1.9.0`
- **WHEN** the latest version is determined
- **THEN** version `1.9.0` is selected

#### Scenario: Unlisted version skipped

- **GIVEN** the newest entry has `listed` set to false
- **WHEN** the latest version is determined
- **THEN** the next older listed, non-prerelease version is selected

#### Scenario: Build metadata

- **GIVEN** the newest entry is `1.2.3+abc`
- **WHEN** the latest version is determined
- **THEN** version `1.2.3` is selected

#### Scenario: No listed version

- **WHEN** every entry is unlisted or a prerelease
- **THEN** the lookup fails stating it failed to get a listed version for the
  package

### Requirement: NuGet Source Repository Extraction

The system SHALL request
`<PackageBaseAddress><name>/<version>/<name>.nuspec` using the lowercased
name and selected version. It SHALL take the `repository` element's `url`
attribute, or failing that the `projectUrl` element. The URL MUST match a
`github.com/<owner>/<repo>` or `gitlab.com/<owner>/<repo>` pattern, with
optional scheme and `www.`. The system SHALL trim a trailing `/` and then a
trailing `.git` from the selected URL. The lookup MUST fail when the nuspec
metadata is empty or when neither URL is a supported GitHub or GitLab URL.

#### Scenario: Repository URL with .git suffix

- **GIVEN** the nuspec repository URL is `https://github.com/foo/foo.net.git`
- **WHEN** the source repository is extracted
- **THEN** the result is `https://github.com/foo/foo.net`

#### Scenario: Fallback to projectUrl on GitLab

- **GIVEN** the nuspec has no repository URL
- **AND** `projectUrl` is `https://gitlab.com/foo/foo.net`
- **WHEN** the source repository is extracted
- **THEN** the result is `https://gitlab.com/foo/foo.net`

#### Scenario: Unsupported project URL

- **GIVEN** neither URL is on github.com or gitlab.com
- **WHEN** the source repository is extracted
- **THEN** the lookup fails stating the source repo is not defined for the
  nuget package

### Requirement: NuGet HTTP Status Handling

Every NuGet request SHALL require an HTTP 200 response. Any other status MUST
fail the lookup with an error naming the request and the returned status.

#### Scenario: Registration index not found

- **WHEN** the registration index request returns HTTP 404
- **THEN** the lookup fails stating it failed to get the nuget package
  registration index json with that status

### Requirement: Resolution Failure Behavior

A failed package lookup MUST stop the run before any repository is scanned.
The CLI SHALL return an error prefixed with
`fetchGitRepositoryFromPackageManagers`. The `serve` HTTP endpoint SHALL
accept the same package names as `npm`, `pypi`, `rubygems`, and `nuget`
query parameters or JSON body fields, and MUST respond with HTTP 500 when the
lookup fails.

#### Scenario: CLI lookup failure

- **WHEN** `scorecard --pypi=<name>` cannot resolve a source repository
- **THEN** the command exits with an error and no checks run

#### Scenario: Serve endpoint lookup failure

- **WHEN** a request to the serve endpoint sets `npm=<name>` and the lookup
  fails
- **THEN** the response status is HTTP 500
