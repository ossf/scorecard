# Security Scorecards

![build](https://github.com/ossf/scorecard/workflows/build/badge.svg?branch=main)
![CodeQL](https://github.com/ossf/scorecard/workflows/CodeQL/badge.svg?branch=main)

<img align="right" src="artwork/openssf_security.png" width="200" height="400">

## Overview 

-  [What Is Scorecards?](#what-is-scorecards)

## Using Scorecards

-  [Prerequisites](#prerequisites) 
-  [Authentication and Setup](#authentication-and-setup)
-  [Basic Usage](#basic-usage)
-  [Report Problems](#report-problems) 
-  [Scorecards' Public Data](#public-data)

## Checks 

-  [Default Scorecards Checks ](#scorecard-checks)
-  [Detailed Check Documentation](docs/checks.md) (Scoring Criteria, Risks, and Remediation)

## Contribute

-  [Code of Conduct](CODE_OF_CONDUCT.md)
-  [Contribute to Scorecards  ](CONTRIBUTING.md)
-  [Add a New Check](checks/write.md)
-  [Connect with the Scorecards Community](#connect-with-the-scorecards-community)
-  [Report a Security Issue](SECURITY.md)
   
________
________
## Overview 
### What is Scorecards?

We created Scorecards to give consumers of open-source projects an easy way to judge whether their dependencies are safe.

Scorecards is an automated tool that assesses a number of important heuristics [("checks")](#scorecard-checks) associated with software security and assigns each check a score of 0-10. You can use these scores to understand specific areas to improve in order to strengthen the security posture of your project. You can also assess the risks that dependencies introduce, and make informed decisions about accepting these risks, evaluating alternative solutions, or working with the maintainers to make improvements.

The inspiration for Scorecards’ logo: ["You passed! All D's ... and an A!"](https://youtu.be/rDMMYT3vkTk)

#### Project Goals

1.  Automate analysis and trust decisions on the security posture of open source
    projects.

1.  Use this data to proactively improve the security posture of the critical
    projects the world depends on.


## Using Scorecards

### Prerequisites

Platforms: Currently, Scorecards supports OSX and Linux platforms. If you are using a Windows OS you may experience issues. Contributions towards supporting Windows are welcome.

Language: You must have GoLang installed to run Scorecards (https://golang.org/doc/install)

### Authentication and Setup

Before running Scorecard, you need to either:

-   [create a GitHub access token](https://docs.github.com/en/free-pro-team@latest/developers/apps/about-apps#personal-access-tokens)
    and set it in an environment variable called `GITHUB_AUTH_TOKEN`,
    `GITHUB_TOKEN`, `GH_AUTH_TOKEN` or `GH_TOKEN`. This helps to avoid the
    GitHub's [api rate limits](https://developer.github.com/v3/#rate-limiting)
    with unauthenticated requests.

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

-   create a GitHub App Installations for higher rate-limit quotas. If you have
    an installed GitHub App and key file, you can use the three environment
    variables below, following the commands shown above for your platform.

```
GITHUB_APP_KEY_PATH=<path to the key file on disk>
GITHUB_APP_INSTALLATION_ID=<installation id>
GITHUB_APP_ID=<app id>
```

These variables can be obtained from the GitHub
[developer settings](https://github.com/settings/apps) page.


#### Docker

`scorecard` is available as a Docker container:

The `GITHUB_AUTH_TOKEN` has to be set to a valid [token](#Authentication)

```shell
docker run -e GITHUB_AUTH_TOKEN=token gcr.io/openssf/scorecard:stable --show-details --repo=https://github.com/ossf/scorecard
```

### Basic Usage
#### Using repository URL

Scorecards can run using just one argument, the URL of the target repo:

```shell
$ go install github.com/ossf/scorecard/v2@latest
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

### Showing Detailed Results 
For more details why a check fails, use the `--show-details` option:

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

#### Using a Package manager

For projects in the `--npm`, `--pypi`, or `--rubygems` ecosystems, you have the option to run Scorecards using a package manager. Provide the package name to run the checks on the corresponding GitHub source code.

For example, `--npm=angular`.

#### Running specific checks

To run only specific check(s), add the `--checks` argument with a list of check
names.

For example, `--checks=CI-Tests,Code-Review`.

#### Formatting Results

There are three formats currently: `default`, `json`, and `csv`. Others may be
added in the future.

These may be specified with the `--format` flag. For example, `--format=json`.

### Report Problems

If you have what looks like a bug, please use the
[Github issue tracking system.](https://github.com/ossf/scorecard/issues)
Before you file an issue, please search existing issues to see if your issue
is already covered.

### Public Data

If you're interested in seeing a list of projects with their Scorecard
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
openssf.scorecardcron.INFORMATION_SCHEMA.PARTITIONS WHERE table_name="scorecard"
ORDER BY partition_id DESC
LIMIT 1'

# Extract to GCS
bq extract --destination_format=NEWLINE_DELIMITED_JSON
'openssf:scorecardcron.scorecard$<partition_id>' gs://bucket-name/filename.json

```

The list of projects that are checked is available in the
[`cron/data/projects.csv`](https://github.com/ossf/scorecard/blob/main/cron/data/projects.csv)
file in this repository. If you would like us to track more, please feel free to
send a Pull Request with others. Currently, this list is derived from **projects hosted on GitHub
ONLY**. We do plan to expand them in near future to account for projects hosted
on other source control systems.

**NOTE**: The public dataset uses a Pass/Fail scoring system with a confidence score
between **0 and 10**. A confidence of 0 indicates that the check was unable to
achieve any real signal, and that the result should be ignored. A confidence of 10
indicates the check was completely sure of the result. 

## Checks
### Scorecard Checks

The following checks are all run against the target project by default:

Name                        | Description
--------------------------- | -----------
Binary-Artifacts            | Is the project free of checked-in binaries?
Branch-Protection           | Does the project use [Branch Protection](https://docs.github.com/en/free-pro-team@latest/github/administering-a-repository/about-protected-branches) ?
CI-Tests                    | Does the project run tests in CI, e.g. [GitHub Actions](https://docs.github.com/en/free-pro-team@latest/actions), [Prow](https://github.com/kubernetes/test-infra/tree/master/prow)?
CII-Best-Practices          | Does the project have an [OpenSSF (formerly CII) Best Practices Badge](https://bestpractices.coreinfrastructure.org/)?
Code-Review                 | Does the project require code review before code is merged?
Contributors                | Does the project have contributors from at least two different organizations?
Dependency-Update-Tool      | Does the project use tools to help update its dependencies?
Fuzzing                     | Does the project use fuzzing tools, e.g. [OSS-Fuzz](https://github.com/google/oss-fuzz)?
Maintained                  | Is the project maintained?
Pinned-Dependencies         | Does the project declare and pin [dependencies](https://docs.github.com/en/free-pro-team@latest/github/visualizing-repository-data-with-graphs/about-the-dependency-graph#supported-package-ecosystems)?
Packaging                   | Does the project build and publish official packages from CI/CD, e.g. [GitHub Publishing](https://docs.github.com/en/free-pro-team@latest/actions/guides/about-packaging-with-github-actions#workflows-for-publishing-packages) ?
SAST                        | Does the project use static code analysis tools, e.g. [CodeQL](https://docs.github.com/en/free-pro-team@latest/github/finding-security-vulnerabilities-and-errors-in-your-code/enabling-code-scanning-for-a-repository#enabling-code-scanning-using-actions), [SonarCloud](https://sonarcloud.io)?
Security-Policy             | Does the project contain a [security policy](https://docs.github.com/en/free-pro-team@latest/github/managing-security-vulnerabilities/adding-a-security-policy-to-your-repository)?
Signed-Releases             | Does the project cryptographically [sign releases](https://wiki.debian.org/Creating%20signed%20GitHub%20releases)?
Token-Permissions           | Does the project declare GitHub workflow tokens as [read only](https://docs.github.com/en/actions/reference/authentication-in-a-workflow)?
Vulnerabilities             | Does the project have unfixed vulnerabilities? Uses the [OSV service](https://osv.dev).

### Detailed Checks Documentation
To see detailed information about each check, its scoring criteria, and remediation steps, check out
the [checks documentation page](docs/checks.md).


## Contribute
### Code of Conduct
Before contributing, please follow our [Code of Conduct](main/CODE_OF_CONDUCT.md). 

### Contribute to Scorecards
See the [Contributing](CONTRIBUTING.md) documentation for guidance on how to
contribute to the project.

### Adding a Scorecard Check

If you'd like to add a check, please see guidance [here](checks/write.md).

### Connect with the Scorecards Community

If you want to get involved in the Scorecards community or have ideas you'd like to chat about, we discuss
this project in the
[OSSF Best Practices Working Group](https://github.com/ossf/wg-best-practices-os-developers)
meetings.

See the
[Community Calendar](https://calendar.google.com/calendar?cid=czYzdm9lZmhwNWk5cGZsdGI1cTY3bmdwZXNAZ3JvdXAuY2FsZW5kYXIuZ29vZ2xlLmNvbQ)
for the schedule and meeting invitations. The meetings happen biweekly
https://calendar.google.com/calendar/embed?src=s63voefhp5i9pfltb5q67ngpes%40group.calendar.google.com&ctz=America%2FLos_Angeles

For realtime discussion, you can join the
[#security_scorecards](https://slack.openssf.org/#security_scorecards) slack
channel. Slack requires registration, but the openssf team is open
invitation to anyone to register here. Feel free to come and ask any
questions.

### Report a Security Issue
To report a security issue, please follow instructions [here](main/SECURITY.md).
