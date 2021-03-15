# Contributing to Security Scorecards

Thank you for contributing your time and expertise to the Security Scorecards project.
This document describes the contribution guidelines for the project.

**Note:** Before you start contributing, you must read and abide by our **[Code of Conduct](./CODE_OF_CONDUCT.md)**.


<!-- vim-markdown-toc GFM -->

* [Contributing code](#contributing-code)
  * [Getting started](#getting-started)
  * [Environment Setup](#environment-setup)
* [Contributing steps](#contributing-steps)
* [How to build scorecard locally](#how-to-build-scorecard-locally)
* [What to do before submitting a pull request](#what-to-do-before-submitting-a-pull-request)
* [Permission for GitHub personal access tokens](#permission-for-github-personal-access-tokens)
* [Where the CI Tests are configured](#where-the-ci-tests-are-configured)
* [dailyscore-cronjob](#dailyscore-cronjob)
  * [Deploying the cron job](#deploying-the-cron-job)
* [How do I add additional GitHub repositories to be scanned by scorecard dailyscore?](#how-do-i-add-additional-github-repositories-to-be-scanned-by-scorecard-dailyscore)
* [Adding New Checks](#adding-new-checks)

<!-- vim-markdown-toc -->

## Contributing code

### Getting started

1. Create [a GitHub account](https://github.com/join)
1. Create a [personal access token](https://docs.github.com/en/free-pro-team@latest/developers/apps/about-apps#personal-access-tokens)
1. Set up your [development environment](#environment-setup)

### Environment Setup

You must install these tools:

1. [`git`](https://help.github.com/articles/set-up-git/): For source control

1. [`go`](https://golang.org/doc/install): You need go version [v1.16](https://golang.org/dl/) or higher.

1. [`docker`](https://docs.docker.com/engine/install/): `v18.9` or higher.

## Contributing steps

1. Submit an issue describing your proposed change to the repo in question.
1. The repo owners will respond to your issue promptly.
1. Fork the desired repo, develop and test your code changes.
1. Submit a pull request.

## How to build scorecard locally

Note that, by building the scorecard from the source code we are allowed to test the changes made locally.

1. Run the following command to clone your fork of the project locally

```shell
git clone git@github.com:<user>/scorecard.git $GOPATH/src/github.com/<user>/scorecard.git
```

1. Ensure you activate module support before continue (`$ export GO111MODULE=on`)
1. Run the command `make build` to build the source code

## What to do before submitting a pull request

Following the targets that can be used to test your changes locally.

| Command    | Description                                         | Is called in the CI? |
| ---------- | --------------------------------------------------- | -------------------- |
| make all   | Runs go test,golangci lint checks, fmt, go mod tidy | yes                  |
| make e2e   | Runs e2e tests                                      | yes                  |

## Permission for GitHub personal access tokens

The personal access token need the following scopes:

- `repo:status` - Access commit status
- `repo_deployment` - Access deployment status
- `public_repo` - Access public repositories

## Where the CI Tests are configured

1. See the [action files](.github/workflows) to check its tests, and the scripts used on it.

## dailyscore-cronjob

scorecard scans https://github.com/ossf/scorecard/blob/main/cron/projects.txt repositories in GCP k8s and publishes the results in GCS bucket `ossf-scorecards`.

The `cron` definition ./scorecard/cron/cron.sh and the `k8s` for the `cron` is in ./scorecard/k8s/cron.yaml

The logs for the cron are available at https://console.cloud.google.com/kubernetes/cronjob/us-central1-c/openssf/default/daily-score/logs?project=openssf

### Deploying the cron job

The cronjob can be deployed into k8s by running `kubectl apply -f ./scorecard/k8s/cron.yaml`. 
This will deploy a k8s cron job.

Any updates to the ./scorecard/cron/cron.sh will be deployed by the docker container ./scorecard/Dockerfile.gsutil


## How do I add additional GitHub repositories to be scanned by scorecard dailyscore?

Scorecard maintains the list of repositories in a file https://github.com/ossf/scorecard/blob/main/cron/projects.txt 

Submit a PR for this file and scorecard would start scanning in subsequent runs.

## Adding New Checks

Each check is currently just a function of type `CheckFn`.
The signature is:

```golang
type CheckFn func(c.Checker) CheckResult
```

Checks are registered in an init function:

```golang
	AllChecks = append(AllChecks, NamedCheck{
		Name: "Code-Review",
		Fn:   DoesCodeReview,
	})
```

Currently only one set of checks can be run.
In the future, we'll allow declaring multiple suites and configuring which checks get run.
