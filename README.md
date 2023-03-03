# OpenSSF Scorecard

[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/ossf/scorecard/badge)](https://api.securityscorecards.dev/projects/github.com/ossf/scorecard)
[![OpenSSF Best Practices](https://bestpractices.coreinfrastructure.org/projects/5621/badge)](https://bestpractices.coreinfrastructure.org/projects/5621)
![build](https://github.com/ossf/scorecard/workflows/build/badge.svg?branch=main)
![CodeQL](https://github.com/ossf/scorecard/workflows/CodeQL/badge.svg?branch=main)
[![Go Reference](https://pkg.go.dev/badge/github.com/ossf/scorecard/v4.svg)](https://pkg.go.dev/github.com/ossf/scorecard/v4)
[![Go Report Card](https://goreportcard.com/badge/github.com/ossf/scorecard/v4)](https://goreportcard.com/report/github.com/ossf/scorecard/v4)
[![codecov](https://codecov.io/gh/ossf/scorecard/branch/main/graph/badge.svg?token=PMJ6NAN9J3)](https://codecov.io/gh/ossf/scorecard)
[![Slack](https://slack.babeljs.io/badge.svg)](https://slack.openssf.org/#security_scorecards)
[![SLSA 3](https://slsa.dev/images/gh-badge-level3.svg)](https://slsa.dev)

<img align="right" src="artwork/openssf_security_compressed.png" width="200" height="400">

## Overview

-   [What Is Scorecard?](#what-is-scorecard)
-   [Prominent Scorecard Users](#prominent-scorecard-users)
-   [Scorecard's Public Data](#public-data)

## Using Scorecard

-   [Scorecard GitHub Action](#scorecard-github-action)
-   [Scorecard REST API](#scorecard-rest-api)
-   [Scorecard Badges](#scorecard-badges)
-   [Scorecard Command Line Interface](#scorecard-command-line-interface)
    -   [Prerequisites](#prerequisites)
    -   [Installation](#installation)
    -   [Authentication](#authentication)
    -   [Basic Usage](#basic-usage)

## Checks

-   [Default Scorecard Checks](#scorecard-checks)
-   [Detailed Check Documentation](docs/checks.md) (Scoring Criteria, Risks, and
    Remediation)

## Other Important Recommendations
-   [Two-factor Authentication (2FA)](#two-factor-authentication-2fa)

## Scoring
-   [Aggregate Score](#aggregate-score)

## Contribute

-   [Report Problems](#report-problems)
-   [Code of Conduct](CODE_OF_CONDUCT.md)
-   [Contribute to Scorecard ](CONTRIBUTING.md)
-   [Add a New Check](checks/write.md)
-   [Connect with the Scorecard Community](#connect-with-the-scorecard-community)
-   [Report a Security Issue](SECURITY.md)

## FAQ

- [FAQ](docs/faq.md)

## Overview

### What is Scorecard?
We created Scorecard to help open source maintainers improve their security
best practices and to help open source consumers judge whether their dependencies
are safe.

Scorecard is an automated tool that assesses a number of important heuristics
[("checks")](#scorecard-checks) associated with software security and assigns
each check a score of 0-10. You can use these scores to understand specific
areas to improve in order to strengthen the security posture of your project.
You can also assess the risks that dependencies introduce, and make informed
decisions about accepting these risks, evaluating alternative solutions, or
working with the maintainers to make improvements.

The inspiration for Scorecard’s logo:
["You passed! All D's ... and an A!"](https://youtu.be/rDMMYT3vkTk)

#### Project Goals

1.  Automate analysis and trust decisions on the security posture of open source
    projects.

1.  Use this data to proactively improve the security posture of the critical
    projects the world depends on.

### Prominent Scorecard Users

Scorecard has been run on thousands of projects to monitor and track security
metrics. Prominent projects that use Scorecard include:

-   [Tensorflow](https://github.com/tensorflow/tensorflow)
-   [Angular](https://github.com/angular/angular)
-   [Flutter](https://github.com/flutter/flutter)
-   [sos.dev](https://sos.dev)
-   [deps.dev](https://deps.dev)

### Public Data

We run a weekly Scorecard scan of the 1 million most critical open source
projects judged by their direct dependencies and publish the results in a
[BigQuery public dataset](https://cloud.google.com/bigquery/public-data).

This data is available in the public BigQuery dataset
`openssf:scorecardcron.scorecard-v2`. The latest results are available in the
BigQuery view `openssf:scorecardcron.scorecard-v2_latest`.

You can query the data using [BigQuery Explorer](http://console.cloud.google.com/bigquery) by navigating to Add Data > Star a project by name > 'openssf'.
For example, you may be interested in how a project's score has changed over time:

```sql
SELECT date, score FROM `openssf.scorecardcron.scorecard-v2` WHERE repo.name="github.com/ossf/scorecard" ORDER BY date ASC
```

You can extract the latest results to Google Cloud storage in JSON format using
the [`bq`](https://cloud.google.com/bigquery/docs/bq-command-line-tool) tool:

```
# Get the latest PARTITION_ID
bq query --nouse_legacy_sql 'SELECT partition_id FROM
openssf.scorecardcron.INFORMATION_SCHEMA.PARTITIONS WHERE table_name="scorecard-v2"
AND partition_id!="__NULL__" ORDER BY partition_id DESC
LIMIT 1'

# Extract to GCS
bq extract --destination_format=NEWLINE_DELIMITED_JSON
'openssf:scorecardcron.scorecard-v2$<partition_id>' gs://bucket-name/filename-*.json

```

The list of projects that are checked is available in the
[`cron/internal/data/projects.csv`](https://github.com/ossf/scorecard/blob/main/cron/internal/data/projects.csv)
file in this repository. If you would like us to track more, please feel free to
send a Pull Request with others. Currently, this list is derived from **projects
hosted on GitHub ONLY**. We do plan to expand them in near future to account for
projects hosted on other source control systems.

## Using Scorecard

### Scorecard GitHub Action

The easiest way to use Scorecard on GitHub projects you own is with the
[Scorecard GitHub Action](https://github.com/ossf/scorecard-action). The Action
runs on any repository change and issues alerts that maintainers can view in the
repository’s Security tab. For more information, see the Scorecard GitHub
Action
[installation instructions](https://github.com/ossf/scorecard-action#installation).

### Scorecard REST API

To query pre-calculated scores of OSS projects, use the [REST API](https://api.securityscorecards.dev).

To enable your project to be available on the REST API, set
[`publish_results: true`](https://github.com/ossf/scorecard-action/blob/dd5015aaf9688596b0e6d11e7f24fff566aa366b/action.yaml#L35)
in the Scorecard GitHub Action setting.

### Scorecard Badges

Enabling [`publish_results: true`](https://github.com/ossf/scorecard-action/blob/dd5015aaf9688596b0e6d11e7f24fff566aa366b/action.yaml#L35)
in Scorecard GitHub Actions also allows maintainers to display a Scorecard badge on their repository to show off their
hard work. This badge also auto-updates for every change made to the repository.
To include a badge on your project's repository, simply add the following markdown to your README:

```
[![OpenSSF
Scorecard](https://api.securityscorecards.dev/projects/github.com/{owner}/{repo}/badge)](https://api.securityscorecards.dev/projects/github.com/{owner}/{repo})
```

### Scorecard Command Line Interface

To run a Scorecard scan on projects you do not own, use the command line
interface installation option.

#### Prerequisites

Platforms: Currently, Scorecard supports OSX and Linux platforms. If you are
using a Windows OS you may experience issues. Contributions towards supporting
Windows are welcome.

Language: You must have GoLang installed to run Scorecard
(https://golang.org/doc/install)

#### Installation

##### Docker

`scorecard` is available as a Docker container:

```shell
docker pull gcr.io/openssf/scorecard:stable
```

To use a specific scorecard version (e.g., v3.2.1), run:

```shell
docker pull gcr.io/openssf/scorecard:v3.2.1
```

##### Standalone

To install Scorecard as a standalone:

Visit our latest [release page](https://github.com/ossf/scorecard/releases/latest) and
download the correct zip file for your operating system.

Add the binary to your `GOPATH/bin` directory (use `go env GOPATH` to identify your directory if necessary).

###### Verifying SLSA provenance for downloaded releases

We generate [SLSA3 signatures](https://slsa.dev) using the OpenSSF's [slsa-framework/slsa-github-generator](https://github.com/slsa-framework/slsa-github-generator) during the release process. To verify a release binary:
1. Install the verification tool from [slsa-framework/slsa-verifier#installation](https://github.com/slsa-framework/slsa-verifier#installation).
2. Download the signature file `attestation.intoto.jsonl` from the [GitHub releases page](https://github.com/GoogleContainerTools/jib/releases/latest).
3. Run the verifier:

```shell
slsa-verifier -artifact-path <the-zip> -provenance attestation.intoto.jsonl -source github.com/ossf/scorecard -tag <the-tag>
```

##### Using package managers

Package Manager                                            | Supported Distribution | Command
---------------------------------------------------------- | ---------------------- | -------
Nix                                                        | NixOS                  | `nix-shell -p nixpkgs.scorecard`
[AUR helper](https://wiki.archlinux.org/title/AUR_helpers) | Arch Linux             | Use your AUR helper to install `scorecard`
[Homebrew](https://brew.sh/)                               | macOS or Linux         | `brew install scorecard`

#### Authentication

GitHub imposes [api rate limits](https://developer.github.com/v3/#rate-limiting)
on unauthenticated requests. To avoid these limits, you must authenticate your
requests before running Scorecard. There are two ways to authenticate your
requests: either create a GitHub personal access token, or create a GitHub App
Installation.

-   [Create a classic GitHub personal access token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token#creating-a-personal-access-token-classic).
    When creating the personal access token, we suggest you choose the
    `public_repo` scope. Set the token in an environment variable called
    `GITHUB_AUTH_TOKEN`, `GITHUB_TOKEN`, `GH_AUTH_TOKEN` or `GH_TOKEN` using the
    commands below according to your platform.

```shell
# For posix platforms, e.g. linux, mac:
export GITHUB_AUTH_TOKEN=<your access token>
# Multiple tokens can be provided separated by comma to be utilized
# in a round robin fashion.
export GITHUB_AUTH_TOKEN=<your access token1>,<your access token2>

# For windows:
set GITHUB_AUTH_TOKEN=<your access token>
set GITHUB_AUTH_TOKEN=<your access token1>,<your access token2>
```

OR

-   [Create a GitHub App Installation](https://docs.github.com/en/developers/apps/building-github-apps/creating-a-github-app)
    for higher rate-limit quotas. If you have an installed GitHub App and key
    file, you can use the three environment variables below, following the
    commands (`set` or `export`) shown above for your platform.

```
GITHUB_APP_KEY_PATH=<path to the key file on disk>
GITHUB_APP_INSTALLATION_ID=<installation id>
GITHUB_APP_ID=<app id>
```

These variables can be obtained from the GitHub
[developer settings](https://github.com/settings/apps) page.

#### Basic Usage

##### Using repository URL

Scorecard can run using just one argument, the URL of the target repo:

```shell
$ scorecard --repo=github.com/ossf-tests/scorecard-check-branch-protection-e2e
Starting [CII-Best-Practices]
Starting [Fuzzing]
Starting [Pinned-Dependencies]
Starting [CI-Tests]
Starting [Maintained]
Starting [Packaging]
Starting [SAST]
Starting [Dependency-Update-Tool]
Starting [Token-Permissions]
Starting [Security-Policy]
Starting [Signed-Releases]
Starting [Binary-Artifacts]
Starting [Branch-Protection]
Starting [Code-Review]
Starting [Contributors]
Starting [Vulnerabilities]
Finished [CI-Tests]
Finished [Maintained]
Finished [Packaging]
Finished [SAST]
Finished [Signed-Releases]
Finished [Binary-Artifacts]
Finished [Branch-Protection]
Finished [Code-Review]
Finished [Contributors]
Finished [Dependency-Update-Tool]
Finished [Token-Permissions]
Finished [Security-Policy]
Finished [Vulnerabilities]
Finished [CII-Best-Practices]
Finished [Fuzzing]
Finished [Pinned-Dependencies]

RESULTS
-------
Aggregate score: 7.9 / 10

Check scores:
|---------|------------------------|--------------------------------|---------------------------------------------------------------------------|
|  SCORE  |          NAME          |             REASON             |                         DOCUMENTATION/REMEDIATION                         |
|---------|------------------------|--------------------------------|---------------------------------------------------------------------------|
| 10 / 10 | Binary-Artifacts       | no binaries found in the repo  | github.com/ossf/scorecard/blob/main/docs/checks.md#binary-artifacts       |
|---------|------------------------|--------------------------------|---------------------------------------------------------------------------|
| 9 / 10  | Branch-Protection      | branch protection is not       | github.com/ossf/scorecard/blob/main/docs/checks.md#branch-protection      |
|         |                        | maximal on development and all |                                                                           |
|         |                        | release branches               |                                                                           |
|---------|------------------------|--------------------------------|---------------------------------------------------------------------------|
| ?       | CI-Tests               | no pull request found          | github.com/ossf/scorecard/blob/main/docs/checks.md#ci-tests               |
|---------|------------------------|--------------------------------|---------------------------------------------------------------------------|
| 0 / 10  | CII-Best-Practices     | no badge found                 | github.com/ossf/scorecard/blob/main/docs/checks.md#cii-best-practices     |
|---------|------------------------|--------------------------------|---------------------------------------------------------------------------|
| 10 / 10 | Code-Review            | branch protection for default  | github.com/ossf/scorecard/blob/main/docs/checks.md#code-review            |
|         |                        | branch is enabled              |                                                                           |
|---------|------------------------|--------------------------------|---------------------------------------------------------------------------|
| 0 / 10  | Contributors           | 0 different companies found -- | github.com/ossf/scorecard/blob/main/docs/checks.md#contributors           |
|         |                        | score normalized to 0          |                                                                           |
|---------|------------------------|--------------------------------|---------------------------------------------------------------------------|
| 0 / 10  | Dependency-Update-Tool | no update tool detected        | github.com/ossf/scorecard/blob/main/docs/checks.md#dependency-update-tool |
|---------|------------------------|--------------------------------|---------------------------------------------------------------------------|
| 0 / 10  | Fuzzing                | project is not fuzzed in       | github.com/ossf/scorecard/blob/main/docs/checks.md#fuzzing                |
|         |                        | OSS-Fuzz                       |                                                                           |
|---------|------------------------|--------------------------------|---------------------------------------------------------------------------|
| 1 / 10  | Maintained             | 2 commit(s) found in the last  | github.com/ossf/scorecard/blob/main/docs/checks.md#maintained             |
|         |                        | 90 days -- score normalized to |                                                                           |
|         |                        | 1                              |                                                                           |
|---------|------------------------|--------------------------------|---------------------------------------------------------------------------|
| ?       | Packaging              | no published package detected  | github.com/ossf/scorecard/blob/main/docs/checks.md#packaging              |
|---------|------------------------|--------------------------------|---------------------------------------------------------------------------|
| 8 / 10  | Pinned-Dependencies    | unpinned dependencies detected | github.com/ossf/scorecard/blob/main/docs/checks.md#pinned-dependencies    |
|         |                        | -- score normalized to 8       |                                                                           |
|---------|------------------------|--------------------------------|---------------------------------------------------------------------------|
| 0 / 10  | SAST                   | no SAST tool detected          | github.com/ossf/scorecard/blob/main/docs/checks.md#sast                   |
|---------|------------------------|--------------------------------|---------------------------------------------------------------------------|
| 0 / 10  | Security-Policy        | security policy file not       | github.com/ossf/scorecard/blob/main/docs/checks.md#security-policy        |
|         |                        | detected                       |                                                                           |
|---------|------------------------|--------------------------------|---------------------------------------------------------------------------|
| ?       | Signed-Releases        | no releases found              | github.com/ossf/scorecard/blob/main/docs/checks.md#signed-releases        |
|---------|------------------------|--------------------------------|---------------------------------------------------------------------------|
| 10 / 10 | Token-Permissions      | tokens are read-only in GitHub | github.com/ossf/scorecard/blob/main/docs/checks.md#token-permissions      |
|         |                        | workflows                      |                                                                           |
|---------|------------------------|--------------------------------|---------------------------------------------------------------------------|
| 10 / 10 | Vulnerabilities        | no vulnerabilities detected    | github.com/ossf/scorecard/blob/main/docs/checks.md#vulnerabilities        |
|---------|------------------------|--------------------------------|---------------------------------------------------------------------------|
```

###### Docker

The `GITHUB_AUTH_TOKEN` has to be set to a valid [token](#Authentication)

```shell
docker run -e GITHUB_AUTH_TOKEN=token gcr.io/openssf/scorecard:stable --show-details --repo=https://github.com/ossf/scorecard
```

To use a specific scorecard version (e.g., v3.2.1), run:

```shell
docker run -e GITHUB_AUTH_TOKEN=token gcr.io/openssf/scorecard:v3.2.1 --show-details --repo=https://github.com/ossf/scorecard
```

##### Showing Detailed Results

For more details about why a check fails, use the `--show-details` option:

```
./scorecard --repo=github.com/ossf-tests/scorecard-check-branch-protection-e2e --checks Branch-Protection --show-details
Starting [Pinned-Dependencies]
Finished [Pinned-Dependencies]

RESULTS
-------
|---------|------------------------|--------------------------------|--------------------------------|---------------------------------------------------------------------------|
|  SCORE  |          NAME          |             REASON             |            DETAILS             |                         DOCUMENTATION/REMEDIATION                         |
|---------|------------------------|--------------------------------|--------------------------------|---------------------------------------------------------------------------|
| 9 / 10  | Branch-Protection      | branch protection is not       | Info: 'force pushes' disabled  | github.com/ossf/scorecard/blob/main/docs/checks.md#branch-protection      |
|         |                        | maximal on development and all | on branch 'main' Info: 'allow  |                                                                           |
|         |                        | release branches               | deletion' disabled on branch   |                                                                           |
|         |                        |                                | 'main' Info: linear history    |                                                                           |
|         |                        |                                | enabled on branch 'main' Info: |                                                                           |
|         |                        |                                | strict status check enabled    |                                                                           |
|         |                        |                                | on branch 'main' Warn: status  |                                                                           |
|         |                        |                                | checks for merging have no     |                                                                           |
|         |                        |                                | specific status to check on    |                                                                           |
|         |                        |                                | branch 'main' Info: number     |                                                                           |
|         |                        |                                | of required reviewers is 2     |                                                                           |
|         |                        |                                | on branch 'main' Info: Stale   |                                                                           |
|         |                        |                                | review dismissal enabled on    |                                                                           |
|         |                        |                                | branch 'main' Info: Owner      |                                                                           |
|         |                        |                                | review required on branch      |                                                                           |
|         |                        |                                | 'main' Info: 'admininistrator' |                                                                           |
|         |                        |                                | PRs need reviews before being  |                                                                           |
|         |                        |                                | merged on branch 'main'        |                                                                           |
|---------|------------------------|--------------------------------|--------------------------------|---------------------------------------------------------------------------|
```

##### Using a Package manager

For projects in the `--npm`, `--pypi`, or `--rubygems` ecosystems, you have the
option to run Scorecard using a package manager. Provide the package name to
run the checks on the corresponding GitHub source code.

For example, `--npm=angular`.

##### Running specific checks

To run only specific check(s), add the `--checks` argument with a list of check
names.

For example, `--checks=CI-Tests,Code-Review`.

##### Formatting Results

The currently supported formats are `default` (text) and `json`.

These may be specified with the `--format` flag. For example, `--format=json`.



## Checks

### Scorecard Checks

The following checks are all run against the target project by default:

Name        | Description                               | Risk Level | Token Required | Note
----------- | ----------------------------------------- | ---------- | --------------- | -----------------
[Binary-Artifacts](docs/checks.md#binary-artifacts)             | Is the project free of checked-in binaries?                                                                                                                                                                                                                                                                                  | High               | PAT, GITHUB_TOKEN   |
[Branch-Protection](docs/checks.md#branch-protection)           | Does the project use [Branch Protection](https://docs.github.com/en/free-pro-team@latest/github/administering-a-repository/about-protected-branches) ?                                                                                                                                                                       | High | PAT (`repo` or `repo> public_repo`), GITHUB_TOKEN    | certain settings are only supported with a maintainer PAT
[CI-Tests](docs/checks.md#ci-tests)                             | Does the project run tests in CI, e.g. [GitHub Actions](https://docs.github.com/en/free-pro-team@latest/actions), [Prow](https://github.com/kubernetes/test-infra/tree/master/prow)?                                                                                                                                         | Low | PAT, GITHUB_TOKEN   |
[CII-Best-Practices](docs/checks.md#cii-best-practices)         | Does the project have an [OpenSSF (formerly CII) Best Practices Badge](https://bestpractices.coreinfrastructure.org/en)?                                                                                                                                                                                                                         | Low  | PAT, GITHUB_TOKEN   |
[Code-Review](docs/checks.md#code-review)                       | Does the project practice code review before code is merged?                                                                                                                                                                                                                                                                 | High | PAT, GITHUB_TOKEN   |
[Contributors](docs/checks.md#contributors)                     | Does the project have contributors from at least two different organizations?                                                                                                                                                                                                                                                | Low | PAT, GITHUB_TOKEN   |
[Dangerous-Workflow](docs/checks.md#dangerous-workflow)         | Does the project avoid dangerous coding patterns in GitHub Action workflows?                                                                                                                                                                                                                                                 | Critical | PAT, GITHUB_TOKEN   |
[Dependency-Update-Tool](docs/checks.md#dependency-update-tool) | Does the project use tools to help update its dependencies?                                                                                                                                                                                                                                                                  | High | PAT, GITHUB_TOKEN   |
[Fuzzing](docs/checks.md#fuzzing)                               | Does the project use fuzzing tools, e.g. [OSS-Fuzz](https://github.com/google/oss-fuzz)?                                                                                                                                                                                                                                     | Medium | PAT, GITHUB_TOKEN   |
[License](docs/checks.md#license)                               | Does the project declare a license?                                                                                                                                                                                                                                                                                          | Low | PAT, GITHUB_TOKEN   |
[Maintained](docs/checks.md#maintained)                         | Is the project at least 90 days old, and maintained?                                                                                                                                                                                                                                                                                                   | High | PAT, GITHUB_TOKEN   |
[Pinned-Dependencies](docs/checks.md#pinned-dependencies)       | Does the project declare and pin [dependencies](https://docs.github.com/en/free-pro-team@latest/github/visualizing-repository-data-with-graphs/about-the-dependency-graph#supported-package-ecosystems)?                                                                                                                     | Medium | PAT, GITHUB_TOKEN   |
[Packaging](docs/checks.md#packaging)                           | Does the project build and publish official packages from CI/CD, e.g. [GitHub Publishing](https://docs.github.com/en/free-pro-team@latest/actions/guides/about-packaging-with-github-actions#workflows-for-publishing-packages) ?                                                                                            | Medium | PAT, GITHUB_TOKEN   |
[SAST](docs/checks.md#sast)                                     | Does the project use static code analysis tools, e.g. [CodeQL](https://docs.github.com/en/free-pro-team@latest/github/finding-security-vulnerabilities-and-errors-in-your-code/enabling-code-scanning-for-a-repository#enabling-code-scanning-using-actions), [LGTM (deprecated)](https://lgtm.com), [SonarCloud](https://sonarcloud.io)? | Medium | PAT, GITHUB_TOKEN   |
[Security-Policy](docs/checks.md#security-policy)               | Does the project contain a [security policy](https://docs.github.com/en/free-pro-team@latest/github/managing-security-vulnerabilities/adding-a-security-policy-to-your-repository)?                                                                                                                                          | Medium | PAT, GITHUB_TOKEN   |
[Signed-Releases](docs/checks.md#signed-releases)               | Does the project cryptographically [sign releases](https://wiki.debian.org/Creating%20signed%20GitHub%20releases)?                                                                                                                                                                                                           | High | PAT, GITHUB_TOKEN   |
[Token-Permissions](docs/checks.md#token-permissions)           | Does the project declare GitHub workflow tokens as [read only](https://docs.github.com/en/actions/reference/authentication-in-a-workflow)?                                                                                                                                                                                   | High | PAT, GITHUB_TOKEN   |
[Vulnerabilities](docs/checks.md#vulnerabilities)               | Does the project have unfixed vulnerabilities? Uses the [OSV service](https://osv.dev).                                                                                                                                                                                                                                      | High | PAT, GITHUB_TOKEN   |
[Webhooks](docs/checks.md#webhooks)               | Does the webhook defined in the repository have a token configured to authenticate the origins of requests?                                                                                                                                                                                                                                      | High | maintainer PAT (`admin: repo_hook` or `admin> read:repo_hook` [doc](https://docs.github.com/en/rest/webhooks/repo-config#get-a-webhook-configuration-for-a-repository)  | EXPERIMENTAL

### Detailed Checks Documentation

To see detailed information about each check, its scoring criteria, and
remediation steps, check out the [checks documentation page](docs/checks.md).

## Other Important Recommendations

### Two-factor Authentication (2FA)

[Two-factor Authentication (2FA)](https://docs.github.com/en/authentication/securing-your-account-with-two-factor-authentication-2fa/about-two-factor-authentication) adds an extra layer of security when logging into websites or apps. 2FA protects your account if your password is compromised by requiring a second form of authentication, such as codes sent via SMS or authentication app, or touching a physical security key.

We strongly recommend that you enable 2FA on GitHub and any important account where it is available. 2FA is not a Scorecard check because GitHub does not make that data about user accounts public. Arguably, this data should always remain private, since accounts without 2FA are so vulnerable to attack.

Though it is not an official check, we urge all project maintainers to enable 2FA to protect their projects from compromise.

#### Enabling 2FA

##### For users

Follow the steps described at [Configuring two-factor authentication](https://docs.github.com/en/authentication/securing-your-account-with-two-factor-authentication-2fa/configuring-two-factor-authentication)

If possible, use either:

- physical security key (preferred), such as Titan or Yubikey
- recovery codes, stored in an access protected and encrypted vault

As a last option, use SMS. Beware: 2FA using SMS is vulnerable to [SIM swap attack](https://en.wikipedia.org/wiki/SIM_swap_scam).

##### For an organization

1. [Prepare to require 2FA in your organization](https://docs.github.com/en/organizations/keeping-your-organization-secure/managing-two-factor-authentication-for-your-organization/preparing-to-require-two-factor-authentication-in-your-organization)
2. [Require 2FA in your organization](https://docs.github.com/en/organizations/keeping-your-organization-secure/managing-two-factor-authentication-for-your-organization/requiring-two-factor-authentication-in-your-organization)

## Scoring

### Aggregate Score
Each individual check returns a score of 0 to 10, with 10 representing the best
possible score. Scorecard also produces an aggregate score, which is a
weight-based average of the individual checks weighted by risk.

*   “Critical” risk checks are weighted at 10
*   “High” risk checks are weighted at 7.5
*   “Medium” risk checks are weighted at 5
*   “Low” risk checks are weighted at 2.5

See the [list of current Scorecard checks](#scorecard-checks) for each check's
risk level.

## Contribute

### Report Problems

If you have what looks like a bug, please use the
[Github issue tracking system.](https://github.com/ossf/scorecard/issues) Before
you file an issue, please search existing issues to see if your issue is already
covered.

### Contribute to Scorecard

Before contributing, please follow our [Code of Conduct](CODE_OF_CONDUCT.md).

See the [Contributing](CONTRIBUTING.md) documentation for guidance on how to
contribute to the project.

### Adding a Scorecard Check

If you'd like to add a check, please see guidance [here](checks/write.md).

### Connect with the Scorecard Community

If you want to get involved in the Scorecard community or have ideas you'd like
to chat about, we discuss this project in the
[OSSF Best Practices Working Group](https://github.com/ossf/wg-best-practices-os-developers)
meetings.

Artifact                      | Link
----------------------------- | ----
Scorecard Dev Forum           | [ossf-scorecard-dev@](https://groups.google.com/g/ossf-scorecard-dev)
Scorecard Announcements Forum | [ossf-scorecard-announce@](https://groups.google.com/g/ossf-scorecard-announce)
Community Meeting VC          | [Link to z o o m meeting](https://zoom.us/j/98835923979)
Community Meeting Calendar    | Biweekly Thursdays, 1:00pm-2:00pm PST <br>[Calendar](https://calendar.google.com/calendar?cid=czYzdm9lZmhwNWk5cGZsdGI1cTY3bmdwZXNAZ3JvdXAuY2FsZW5kYXIuZ29vZ2xlLmNvbQ)
Meeting Notes                 | [Notes](https://docs.google.com/document/d/1dB2U7_qZpNW96vtuoG7ShmgKXzIg6R5XT5Tc-0yz6kE/edit#heading=h.4k8ml0qkh7tl)
Slack Channel                 | [#security_scorecards](https://slack.openssf.org/#security_scorecards)

&nbsp;                                                           | Facilitators      | Company | Profile
---------------------------------------------------------------- | ----------------- | ------- | -------
<img width="30px" src="https://github.com/azeemshaikh38.png">    | Azeem Shaikh      | Google  | [azeemshaikh38](https://github.com/azeemshaikh38)
<img width="30px" src="https://github.com/laurentsimon.png">     | Laurent Simon     | Google  | [laurentsimon](https://github.com/laurentsimon)
<img width="30px" src="https://github.com/naveensrinivasan.png"> | Naveen Srinivasan | Endor Labs | [naveensrinivasan](https://github.com/naveensrinivasan)
<img width="30px" src="https://github.com/chrismcgehee.png">     | Chris McGehee     | Datto   | [chrismcgehee](https://github.com/chrismcgehee)
<img width="30px" src="https://github.com/justaugustus.png">     | Stephen Augustus  | Cisco   | [justaugustus](https://github.com/justaugustus)

### Report a Security Issue

To report a security issue, please follow instructions [here](SECURITY.md).

## Stargazers over time

[![Stargazers over time](https://starchart.cc/ossf/scorecard.svg)](https://starchart.cc/ossf/scorecard)


## FAQ

### FAQ

See the [FAQ](docs/faq.md) for answers to Frequently Asked Questions about Scorecard.
