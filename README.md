# Security Scorecards

![build](https://github.com/ossf/scorecard/workflows/build/badge.svg?branch=main)
![CodeQL](https://github.com/ossf/scorecard/workflows/CodeQL/badge.svg?branch=main)
[![codecov](https://codecov.io/gh/ossf/scorecard/branch/main/graph/badge.svg?token=PMJ6NAN9J3)](https://codecov.io/gh/ossf/scorecard)

<img align="right" src="artwork/openssf_security.png" width="200" height="400">

<!-- vim-markdown-toc GFM -->

*   [Motivation](#motivation)
*   [Goals](#goals)
*   [Scorecard Checks](#scorecard-checks)
*   [Usage](#usage)
    *   [Using repository URL](#using-repository-url)
    *   [Using a package manager](#using-a-package-manager)
    *   [Running specific checks](#running-specific-checks)
    *   [Authentication](#authentication)
    *   [Understanding Scorecard results](#understanding-scorecard-results)
    *   [Formatting Results](#formatting-results)
*   [Public Data](#public-data)
*   [Adding a Scorecard Check](#adding-a-scorecard-check)
*   [Troubleshooting](#troubleshooting)
*   [Supportability](#supportability)
*   [Contributing](#contributing)

<!-- vim-markdown-toc -->

## Motivation

A short motivational video clip to inspire us: https://youtu.be/rDMMYT3vkTk "You
passed! All D's ... and an A!"

## Goals

1.  Automate analysis and trust decisions on the security posture of open source
    projects.

1.  Use this data to proactively improve the security posture of the critical
    projects the world depends on.

## Scorecard Checks

The following checks are all run against the target project by default:

Name               | Description
------------------ | -----------
Security-Policy    | Does the project contain a [security policy](https://docs.github.com/en/free-pro-team@latest/github/managing-security-vulnerabilities/adding-a-security-policy-to-your-repository)?
Contributors       | Does the project have contributors from at least two different organizations?
Frozen-Deps        | Does the project declare and freeze [dependencies](https://docs.github.com/en/free-pro-team@latest/github/visualizing-repository-data-with-graphs/about-the-dependency-graph#supported-package-ecosystems)?
Signed-Releases    | Does the project cryptographically [sign releases](https://wiki.debian.org/Creating%20signed%20GitHub%20releases)?
Signed-Tags        | Does the project cryptographically sign release tags?
CI-Tests           | Does the project run tests in CI, e.g. [GitHub Actions](https://docs.github.com/en/free-pro-team@latest/actions), [Prow](https://github.com/kubernetes/test-infra/tree/master/prow)?
Code-Review        | Does the project require code review before code is merged?
CII-Best-Practices | Does the project have a [CII Best Practices Badge](https://bestpractices.coreinfrastructure.org/en)?
Pull-Requests      | Does the project use [Pull Requests](https://docs.github.com/en/free-pro-team@latest/github/collaborating-with-issues-and-pull-requests/about-pull-requests) for all code changes?
Fuzzing            | Does the project use fuzzing tools, e.g. [OSS-Fuzz](https://github.com/google/oss-fuzz)?
SAST               | Does the project use static code analysis tools, e.g. [CodeQL](https://docs.github.com/en/free-pro-team@latest/github/finding-security-vulnerabilities-and-errors-in-your-code/enabling-code-scanning-for-a-repository#enabling-code-scanning-using-actions), [SonarCloud](https://sonarcloud.io)?
Active             | Did the project get any commits in the last 90 days?
Branch-Protection  | Does the project use [Branch Protection](https://docs.github.com/en/free-pro-team@latest/github/administering-a-repository/about-protected-branches) ?
Packaging          | Does the project build and publish official packages from CI/CD, e.g. [GitHub Publishing](https://docs.github.com/en/free-pro-team@latest/actions/guides/about-packaging-with-github-actions#workflows-for-publishing-packages) ?
Vulnerabilities    | Does the project have unfixed vulnerabilities? Uses the [OSV service](https://osv.dev).

To see detailed information about each check and remediation steps, check out
the [checks documentation page](checks/checks.md).

## Usage

### Using repository URL

The program can run using just one argument, the URL of the repo:

```shell
$ go build
$ ./scorecard --repo=github.com/kubernetes/kubernetes
Starting [Active]
Starting [Branch-Protection]
Starting [CI-Tests]
Starting [CII-Best-Practices]
Starting [Code-Review]
Starting [Contributors]
Starting [Frozen-Deps]
Starting [Fuzzing]
Starting [Packaging]
Starting [Pull-Requests]
Starting [SAST]
Starting [Security-Policy]
Starting [Signed-Releases]
Starting [Signed-Tags]
Finished [Fuzzing]
Finished [CII-Best-Practices]
Finished [Branch-Protection]
Finished [Packaging]
Finished [Security-Policy]
Finished [Frozen-Deps]
Finished [Signed-Tags]
Finished [Signed-Releases]
Finished [SAST]
Finished [CI-Tests]
Finished [Active]
Finished [Contributors]
Finished [Pull-Requests]
Finished [Code-Review]

RESULTS
-------
Active: Pass 10
Branch-Protection: Fail 10
CI-Tests: Pass 10
CII-Best-Practices: Pass 10
Code-Review: Pass 10
Contributors: Pass 10
Frozen-Deps: Pass 10
Fuzzing: Pass 10
Packaging: Fail 0
Pull-Requests: Pass 10
SAST: Fail 10
Security-Policy: Pass 10
Signed-Releases: Fail 10
Signed-Tags: Fail 10
```

### Using a Package manager

scorecard has an option to provide either `--npm` / `--pypi` / `--rubygems`
package name and it would run the checks on the corresponding GitHub source
code.

For example:

```shell
./scorecard --npm=angular
Starting [Active]
Starting [Branch-Protection]
Starting [CI-Tests]
Starting [CII-Best-Practices]
Starting [Code-Review]
Starting [Contributors]
Starting [Frozen-Deps]
Starting [Fuzzing]
Starting [Packaging]
Starting [Pull-Requests]
Starting [SAST]
Starting [Security-Policy]
Starting [Signed-Releases]
Starting [Signed-Tags]
Finished [Signed-Releases]
Finished [Fuzzing]
Finished [CII-Best-Practices]
Finished [Security-Policy]
Finished [CI-Tests]
Finished [Packaging]
Finished [SAST]
Finished [Code-Review]
Finished [Branch-Protection]
Finished [Frozen-Deps]
Finished [Signed-Tags]
Finished [Active]
Finished [Pull-Requests]
Finished [Contributors]

RESULTS
-------
Active: Fail 10
Branch-Protection: Fail 0
CI-Tests: Pass 10
CII-Best-Practices: Fail 10
Code-Review: Pass 10
Contributors: Pass 10
Frozen-Deps: Fail 0
Fuzzing: Fail 10
Packaging: Fail 0
Pull-Requests: Fail 9
SAST: Fail 10
Security-Policy: Pass 10
Signed-Releases: Fail 0
Signed-Tags: Fail 10
```

### Running specific checks

To use a particular check(s), add the `--checks` argument with a list of check
names.

For example, `--checks=CI-Tests,Code-Review`.

### Authentication

Before running Scorecard, you need to, either:

-   [create a GitHub access token](https://docs.github.com/en/free-pro-team@latest/developers/apps/about-apps#personal-access-tokens)
    and set it in environment variable `GITHUB_AUTH_TOKEN`. This helps to avoid
    the GitHub's
    [api rate limits](https://developer.github.com/v3/#rate-limiting) with
    unauthenticated requests.

```shell
# For posix platforms, e.g. linux, mac:
export GITHUB_AUTH_TOKEN=<your access token>

# For windows:
set GITHUB_AUTH_TOKEN=<your access token>
```

Multiple `GITHUB_AUTH_TOKEN` can be provided separated by comma to be utilized
in a round robin fashion.

-   create a GitHub App Installations for higher rate-limit quotas. If you have
    an installed GitHub App and key file, you can use these three environment
    variables, following the commands shown above for your platform.

```
GITHUB_APP_KEY_PATH=<path to the key file on disk>
GITHUB_APP_INSTALLATION_ID=<installation id>
GITHUB_APP_ID=<app id>
```

These can be obtained from the GitHub
[developer settings](https://github.com/settings/apps) page.

### Understanding Scorecard results

Each check returns a **Pass / Fail** decision, as well as a confidence score
between **0 and 10**. A confidence of 0 should indicate the check was unable to
achieve any real signal, and the result should be ignored. A confidence of 10
indicates the check is completely sure of the result.

### Formatting Results

There are three formats currently: `default`, `json`, and `csv`. Others may be
added in the future.

These may be specified with the `--format` flag.

## Public Data

If you're only interested in seeing a list of projects with their Scorecard
check results, we publish these results in a
[BigQuery public dataset](https://cloud.google.com/bigquery/public-data).

This data is available in the public BigQuery dataset
`openssf:scorecardcron.scorecard`. The latest results are available in the
BigQuery view `openssf:scorecardcron.scorecard_latest`.

You can extract the latest results to Google Cloud storage in JSON format using
the [`bq`](https://cloud.google.com/bigquery/docs/bq-command-line-tool) tool:

```
# Get the latest PARTITION_ID
bq query --nouse_legacy_sql 'SELECT partition_id FROM
openssf.scorecardcron.INFORMATION_SCHEMA.PARTITIONS ORDER BY partition_id DESC
LIMIT 1'

# Extract to GCS
bq extract --destination_format=NEWLINE_DELIMITED_JSON
'openssf:scorecardcron.scorecard$<partition_id>' gs://bucket-name/filename.json

```

The list of projects that are checked is available in the
[`cron/data/projects.csv`](https://github.com/ossf/scorecard/blob/main/cron/data/projects.csv)
file in this repository. If you would like us to track more, please feel free to
send a Pull Request with others.

**NOTE**: Currently, these lists are derived from **projects hosted on GitHub
ONLY**. We do plan to expand them in near future to account for projects hosted
on other source control systems.

## Adding a Scorecard Check

If you'd like to add a check, make sure it is something that meets the following
criteria and then create a new GitHub Issue:

-   The scorecard must only be composed of automate-able, objective data. For
    example, a project having 10 contributors doesn’t necessarily mean it’s more
    secure than a project with say 50 contributors. But, having two maintainers
    might be preferable to only having one - the larger bus factor and ability
    to provide code reviews is objectively better.
-   The scorecard criteria can be as specific as possible and not limited
    general recommendations. For example, for Go, we can recommend/require
    specific linters and analyzers to be run on the codebase.
-   The scorecard can be populated for any open source project without any work
    or interaction from maintainers.
-   Maintainers must be provided with a mechanism to correct any automated
    scorecard findings they feel were made in error, provide "hints" for
    anything we can't detect automatically, and even dispute the applicability
    of a given scorecard finding for that repository.
-   Any criteria in the scorecard must be actionable. It should be possible,
    with help, for any project to "check all the boxes".
-   Any solution to compile a scorecard should be usable by the greater open
    source community to monitor upstream security.

## Troubleshooting

-   ### Bugs and Feature Requests:

    If you have what looks like a bug, or you would like to make a feature
    request, please use the
    [Github issue tracking system.](https://github.com/ossf/scorecard/issues)
    Before you file an issue, please search existing issues to see if your issue
    is already covered.

-   ### Slack

    For realtime discussion, you can join the
    [#security_scorecards](https://slack.openssf.org/#security_scorecards) slack
    channel. Slack requires registration, but the openssf team is open
    invitation to anyone to register here. Feel free to come and ask any
    questions.

## Supportability

Currently, scorecard officially supports OSX and Linux platforms. So, if you are
using a Windows OS you may find issues. Contributions towards supporting Windows
are welcome.

## Contributing

If you want to get involved or have ideas you'd like to chat about, we discuss
this project in the
[OSSF Best Practices Working Group](https://github.com/ossf/wg-best-practices-os-developers)
meetings.

See the
[Community Calendar](https://calendar.google.com/calendar?cid=czYzdm9lZmhwNWk5cGZsdGI1cTY3bmdwZXNAZ3JvdXAuY2FsZW5kYXIuZ29vZ2xlLmNvbQ)
for the schedule and meeting invitations. The meetings happen biweekly
https://calendar.google.com/calendar/embed?src=s63voefhp5i9pfltb5q67ngpes%40group.calendar.google.com&ctz=America%2FLos_Angeles

See the [Contributing](CONTRIBUTING.md) documentation for guidance on how to
contribute.
