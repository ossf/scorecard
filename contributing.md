# Contributing to OSS Scorecards!

Thank you for contributing your time and expertise to the OSS Scorecards project.
This document describes the contribution guidelines for the project.

**Note:** Before you start contributing, you must read and abide by our **[Code of Conduct](./code-of-conduct.md)**.

## Contributing code

### Getting started

1.  Create [a GitHub account](https://github.com/join)
1.  Create a [personal access token](https://docs.github.com/en/free-pro-team@latest/developers/apps/about-apps#personal-access-tokens)
1.  Set up your [development environment](#environment-setup)

Then you can [iterate](#iterating).
    
## Environment Setup

You must install these tools:

1.  [`git`](https://help.github.com/articles/set-up-git/): For source control

1.  [`go`](https://golang.org/doc/install): The language Tekton Pipelines is
    built in. You need go version [v1.15](https://golang.org/dl/) or higher.

## Iterating

You can build the project with:

```shell
go build .
```

You can also use `go run` to iterate without a separate rebuild step:

```shell
go run . --repo=<repo>
```

You can run tests with:

```shell
go test .
```

## Adding New Checks

Each check is currently just a function of type `CheckFn`.
The signature is:

```golang
type CheckFn func(*checker.Checker) CheckResult
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
