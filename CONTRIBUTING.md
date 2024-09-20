# Contributing to OpenSSF Scorecard

Thank you for contributing your time and expertise to the OpenSSF Scorecard
project. This document describes the contribution guidelines for the project.

> [!IMPORTANT]
> Before you start contributing, you must read and abide by our
**[Code of Conduct](./CODE_OF_CONDUCT.md)**.
>
> Additionally the Linux Foundation (LF) requires all contributions include per-commit sign-offs.
> Ensure you use the `-s` or `--signoff` flag for every commit.
>
> For more details, see the [LF DCO wiki](https://wiki.linuxfoundation.org/dco)
> or [this Pi-hole signoff guide](https://docs.pi-hole.net/guides/github/how-to-signoff/).

<!-- vim-markdown-toc GFM -->

* [Contributing code](#contributing-code)
  * [Getting started](#getting-started)
  * [Environment Setup](#environment-setup)
  * [New to Go?](#new-to-go)
* [Contributing steps](#contributing-steps)
* [How to build scorecard locally](#how-to-build-scorecard-locally)
* [PR Process](#pr-process)
* [What to do before submitting a pull request](#what-to-do-before-submitting-a-pull-request)
* [Changing Score Results](#changing-score-results)
* [Linting](#linting)
* [Permission for GitHub personal access tokens](#permission-for-github-personal-access-tokens)
* [Adding New Probes](#adding-new-probes)
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
    [personal access token](https://docs.github.com/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens)
1.  Set up your [development environment](#environment-setup)

### Environment Setup

You must install these tools:

1.  [`git`](https://help.github.com/articles/set-up-git/): For source control

1.  [`go`](https://golang.org/doc/install): You need go version
    [v1.22.0](https://golang.org/dl/) or higher.

1.  [`protoc`](https://grpc.io/docs/protoc-installation/): `v3` or higher

1. [`make`](https://www.gnu.org/software/make/): You can build and run Scorecard without it, but some tasks are easier if you have it.

You may need these tools for some tasks:

1.  [`docker`](https://docs.docker.com/engine/install/): `v18.9` or higher.

### New to Go?

If you're unfamiliar with the language, there are plenty of articles, resources, and books.
We recommend starting with several resources from the official Go website:

* [How to Write Go Code](https://go.dev/doc/code)
* [A Tour of Go](https://go.dev/tour/)
* [Effective Go](https://go.dev/doc/effective_go)

## Contributing steps

1.  Identify an existing issue you would like to work on, or submit an issue describing your proposed change to the repo in question.
1.  The repo owners will respond to your issue promptly.
1.  Fork the desired repo, develop and test your code changes.
1.  Submit a pull request.

## How to build Scorecard locally

Note that, by building Scorecard from the source code we are allowed to test
the changes made locally.

1.  Clone your fork of the project locally. ([Detailed instructions](https://docs.github.com/repositories/creating-and-managing-repositories/cloning-a-repository#cloning-a-repository))
1.  Enter the project folder by running the command `cd ./scorecard`
1.  Install the build tools for the project by running the command `make install`
1.  Run the command `make build` to build the source code

## How to run scorecard locally

In the project folder, run the following command:

```shell
# Get scores for a repository
go run main.go --repo=github.com/ossf-tests/scorecard-check-branch-protection-e2e
```

Many developers prefer working with the JSON output format, although you may need to pretty print it.
Piping the output to [jq](https://jqlang.github.io/jq/) is one way of doing this.
```shell
# Get scores for a repository
go run main.go --repo=github.com/ossf-tests/scorecard-check-branch-protection-e2e --format json | jq
```

To view all Scorecard commands and flags run:

```shell
# View scorecard help
go run main.go --help
```

You should familiarize yourself with:
* `--repo` and `--local` to specify a repository
* `--checks` and `--probes` to specify which analyses run
* `--format` to change the result output format
* `--show-details` is pretty self explanatory

### Choosing checks to run

You can use the `--checks` option to select which checks to run.
This is useful if, for example, you only want to run the check you're
currently developing.

```shell
# Get score for Pinned-Dependencies check
go run main.go --repo=github.com/ossf-tests/scorecard-check-branch-protection-e2e --checks=Pinned-Dependencies

# Get score for Pinned-Dependencies and Binary-Artifacts check
go run main.go --repo=github.com/ossf-tests/scorecard-check-branch-protection-e2e --checks=Pinned-Dependencies,Binary-Artifacts
```

## PR Process

Every PR should be annotated with an icon indicating whether it's a:

-   Breaking change: :warning: (`:warning:`)
-   Non-breaking feature: :sparkles: (`:sparkles:`)
-   Patch fix: :bug: (`:bug:`)
-   Documentation changes (user or developer): :book: (`:book:`)
-   Infra/Tests/Other: :seedling: (`:seedling:`)
-   No release note: :ghost: (`:ghost:`)

Use :ghost: (no release note) only for the PRs that change or revert unreleased
changes, which don't deserve a release note. Please don't abuse it.

Prefer using the `:xyz:` aliases over the equivalent emoji directly when possible.

Individual commits should not be tagged separately, but will generally be
assumed to match the PR. For instance, if you have a bugfix in with a breaking
change, it's generally encouraged to submit the bugfix separately, but if you
must put them in one PR, you should mark the whole PR as breaking.

> [!NOTE]
> Once a maintainer reviews your code, please address feedback without rebasing when possible.
> This includes [synchronizing your PR](https://docs.github.com/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/keeping-your-pull-request-in-sync-with-the-base-branch)
> with `main`. The GitHub review experience is much nicer with traditional merge commits.

## What to do before submitting a pull request

Following the targets that can be used to test your changes locally.

| Command  | Description                                        | Is called in the CI? |
| -------- | -------------------------------------------------- | -------------------- |
| `make all` | Runs go test,golangci lint checks, fmt, go mod tidy| yes                  |
| `make e2e-pat` | Runs e2e tests                                     | yes                  |

When developing locally, the following targets are useful to run frequently.
While they are included in `make all`, running them individually is faster.

| Command  | Description | Called in the CI? |
|----------|-------------|-------------------|
| `make unit-test` | Runs unit tests only | yes |
| `make check-linter` | Checks linter issues only | yes |

## Changing Score Results

As a general rule of thumb, pull requests that change Scorecard score results will need a good reason to do so to get merged. 
It is a good idea to discuss such changes in a GitHub issue before implementing them.

## Linting

Most linter issues can be fixed with `golangci-lint` with the following command:

```
make fix-linter
```

## Permission for GitHub personal access tokens

For public repos, classic personal access tokens need the following scopes:

- `public_repo` - Read/write access to public repositories. Needed for branch protection

## Where the CI Tests are configured

1.  See the [action files](.github/workflows) to check its tests, and the
    scripts used on it.

## How do I add additional GitHub repositories to be scanned by scorecard weekly?

Scorecard maintains the list of GitHub repositories in a file
https://github.com/ossf/scorecard/blob/main/cron/internal/data/projects.csv

GitLab repositories are listed in:
https://github.com/ossf/scorecard/blob/main/cron/internal/data/gitlab-projects.csv

Append your desired repositories to the end of these files, then run `make add-projects`.
Commit the changes, and submit a PR and scorecard would start scanning in subsequent runs.

## Adding New Checks

See [checks/write.md](checks/write.md).
When you add new checks, you need to also update the docs.

## Adding New Probes

See [probes/README.md](probes/README.md) for information about the probes.

## Updating Docs

A summary for each check needs to be included in the `README.md`.
In most cases, to update the documentation simply edit the corresponding
`.md` file, with the notable exception of the auto-generated file `checks.md`.

> [!IMPORTANT]  
> **DO NOT** edit `docs/checks.md` directly, as that is an
> auto-generated file. Edit `docs/checks/internal/checks.yaml` instead.

Details about each check need to be  provided in
[docs/checks/internal/checks.yaml](docs/checks/internal/checks.yaml).
If you want to update its documentation:
1. Make your edits in `docs/checks/internal/checks.yaml`.
2. Regenerate `docs/checks.md` by running `make generate-docs`
3. Commit both `docs/checks/internal/checks.yaml` and `docs/checks.md`
