# Contributing to OpenSSF Scorecard

Thank you for contributing your time and expertise to the OpenSSF Scorecard
project. This document describes the contribution guidelines for the project.

**Note:** Before you start contributing, you must read and abide by our
**[Code of Conduct](./CODE_OF_CONDUCT.md)**.

<!-- vim-markdown-toc GFM -->

* [Contributing code](#contributing-code)
  * [Getting started](#getting-started)
  * [Environment Setup](#environment-setup)
* [Contributing steps](#contributing-steps)
* [Error handling](#error-handling)
* [How to build scorecard locally](#how-to-build-scorecard-locally)
* [PR Process](#pr-process)
* [What to do before submitting a pull request](#what-to-do-before-submitting-a-pull-request)
* [Permission for GitHub personal access tokens](#permission-for-github-personal-access-tokens)
* [Where the CI Tests are configured](#where-the-ci-tests-are-configured)
* [dailyscore-cronjob](#dailyscore-cronjob)
  * [Deploying the cron job](#deploying-the-cron-job)
* [How do I add additional GitHub repositories to be scanned by scorecard dailyscore?](#how-do-i-add-additional-github-repositories-to-be-scanned-by-scorecard-dailyscore)
* [Adding New Checks](#adding-new-checks)
* [Updating Docs](#updating-docs)
* [Choosing checks to run](#choosing-checks-to-run)

<!-- vim-markdown-toc -->

## Contributing code

### Getting started

1.  Create [a GitHub account](https://github.com/join)
1.  Create a
    [personal access token](https://docs.github.com/en/free-pro-team@latest/developers/apps/about-apps#personal-access-tokens)
1.  Set up your [development environment](#environment-setup)

### Environment Setup

You must install these tools:

1.  [`git`](https://help.github.com/articles/set-up-git/): For source control

1.  [`go`](https://golang.org/doc/install): You need go version
    [v1.17](https://golang.org/dl/) or higher.

1.  [`docker`](https://docs.docker.com/engine/install/): `v18.9` or higher.

## Contributing steps

1.  Submit an issue describing your proposed change to the repo in question.
1.  The repo owners will respond to your issue promptly.
1.  Fork the desired repo, develop and test your code changes.
1.  Submit a pull request.

## Error handling

See [errors](errors/errors.md).

## How to build scorecard locally

Note that, by building the scorecard from the source code we are allowed to test
the changes made locally.

1.  Run the following command to clone your fork of the project locally

```shell
git clone git@github.com:<user>/scorecard.git $GOPATH/src/github.com/<user>/scorecard.git
```

1.  Enter the project folder by running the command `cd ./scorecard`
1.  Ensure you activate module support before continue (`$ export
    GO111MODULE=on`)
1.  Install the build tools for the project by running the command `make install`
1.  Run the command `make build` to build the source code

## How to run scorecard locally

In the project folder, run the following command:

```shell
// Get scores for a repository
$ go run main.go --repo=github.com/ossf-tests/scorecard-check-branch-protection-e2e
```

You can input the repository you want to analyze using the `--repo=<your_repo>` flag. To view more Scorecard commands run:

```shell
// View scorecard help
$ go run main.go --help
```

### Choosing checks to run

You can use the `--checks` option to select which checks to run.
This is useful if, for example, you only want to run the check you're
currently developing.

```shell
// Get score for Pinned-Dependencies check
$ go run main.go --repo=github.com/ossf-tests/scorecard-check-branch-protection-e2e --checks=Pinned-Dependencies

// Get score for Pinned-Dependencies and Binary-Artifacts check
$ go run main.go --repo=github.com/ossf-tests/scorecard-check-branch-protection-e2e --checks=Pinned-Dependencies,Binary-Artifacts
```

## PR Process

Every PR should be annotated with an icon indicating whether it's a:

-   Breaking change: :warning: (`:warning:`)
-   Non-breaking feature: :sparkles: (`:sparkles:`)
-   Patch fix: :bug: (`:bug:`)
-   Docs: :book: (`:book:`)
-   Infra/Tests/Other: :seedling: (`:seedling:`)
-   No release note: :ghost: (`:ghost:`)

Use :ghost: (no release note) only for the PRs that change or revert unreleased
changes, which don't deserve a release note. Please don't abuse it.

You are free to use the `:xyz:` aliases or to use the equivalent emoji directly.

Individual commits should not be tagged separately, but will generally be
assumed to match the PR. For instance, if you have a bugfix in with a breaking
change, it's generally encouraged to submit the bugfix separately, but if you
must put them in one PR, you should mark the whole PR as breaking.

## What to do before submitting a pull request

Following the targets that can be used to test your changes locally.

| Command  | Description                                        | Is called in the CI? |
| -------- | -------------------------------------------------- | -------------------- |
| make all | Runs go test,golangci lint checks, fmt, go mod tidy| yes                  |
| make e2e-pat | Runs e2e tests                                     | yes                  |

Make sure to signoff your commits before submitting a pull request.

https://docs.pi-hole.net/guides/github/how-to-signoff/

## Permission for GitHub personal access tokens

The personal access token need the following scopes:

-   `repo:status` - Access commit status
-   `repo_deployment` - Access deployment status
-   `public_repo` - Access public repositories

## Where the CI Tests are configured

1.  See the [action files](.github/workflows) to check its tests, and the
    scripts used on it.

## How do I add additional GitHub repositories to be scanned by scorecard weekly?

Scorecard maintains the list of repositories in a file
https://github.com/ossf/scorecard/blob/main/cron/internal/data/projects.csv

Submit a PR for this file and scorecard would start scanning in subsequent runs.

## Adding New Checks

See [checks/write.md](checks/write.md).
When you add new checks, you need to also update the docs.

## Updating Docs

A summary for each check needs to be included in the `README.md`.
In most cases, to update the documentation simply edit the corresponding
`.md` file, with the notable exception of the auto-generated file `checks.md`.

Details about each check need to be  provided in
[docs/checks/internal/checks.yaml](docs/checks/internal/checks.yaml).
If you want to update its documentation, update that `checks.yaml` file.

Whenever you modify the `checks.yaml` file, run the following to
generate `docs/checks.md`:

~~~~
make generate-docs
~~~~

**DO NOT** edit `docs/checks.md` directly, as that is an
auto-generated file. Edit `docs/checks/internal/checks.yaml` instead.
