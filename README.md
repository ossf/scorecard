# Open Source Scorecards

## Motivation

A short motivational video clip to inspire us: https://youtu.be/rDMMYT3vkTk "You passed! All D's ... and an A!"

## Goals
1. Fill the gaps that prevent automated analysis and trust decisions for measuring and reporting on the security posture of open source projects. 

1. Use this data to proactively improve the security posture of the critical projects the world depends on.

## Usage

The program only requires one argument to run, the name of the repo:

```shell
$ go build
$ ./scorecard --repo=github.com/kubernetes/kubernetes
Starting [Active]
Starting [CI-Tests]
Starting [CII-Best-Practices]
Starting [Code-Review]
Starting [Contributors]
Starting [Frozen-Deps]
Starting [Fuzzing]
Starting [Pull-Requests]
Starting [SAST]
Starting [Security-Policy]
Starting [Signed-Releases]
Starting [Signed-Tags]
Finished [Fuzzing]
Finished [CII-Best-Practices]
Finished [Frozen-Deps]
Finished [Security-Policy]
Finished [Contributors]
Finished [Signed-Releases]
Finished [Signed-Tags]
Finished [CI-Tests]
Finished [SAST]
Finished [Code-Review]
Finished [Pull-Requests]
Finished [Active]

RESULTS
-------
Active: Pass 10
CI-Tests: Pass 10
CII-Best-Practices: Pass 10
Code-Review: Pass 10
Contributors: Pass 10
Frozen-Deps: Pass 10
Fuzzing: Pass 10
Pull-Requests: Pass 10
SAST: Fail 0
Security-Policy: Pass 10
Signed-Releases: Fail 10
Signed-Tags: Fail 5
```

It is recommended to use an OAuth token to avoid rate limits.
You can create one by the following the instructions
[here](https://docs.github.com/en/free-pro-team@latest/developers/apps/about-apps#personal-access-tokens).
Set the access token as an environment variable:

```shell
export GITHUB_AUTH_TOKEN=<your access token>
```

## Checks

The following checks are all run against the target project:

| Name  | Description |
|---|---|
| Security-MD | Does the project contain security policies? |
| Contributors  | Does the project have contributors from at least two different organizations? |
| Frozen-Deps | Does the project declare and freeze dependencies? |
| Signed-Tags | Does the project cryptographically sign release tags? |
| Signed-Releases | Does the project cryptographically sign releases? |
| CI-Tests | Does the project run tests in CI? |
| Code-Review | Does the project require code review before code is merged? |
| CII-Best-Practices | Does the project have a CII Best Practices Badge? |
| Pull-Requests | Does the project use Pull Requests for all changes? |
| Fuzzing | Does the project use OSS-Fuzz? |
| SAST | Does the project use static code analysis tools, e.g. CodeQL? |
| Active | Did the project get any commits and releases in last 90 days? |

To see detailed information on how each check works, see the check-specific documentation pages.
To use a particular check, add the `--checks` argument with a list of check
names (for example `--checks=CI-Tests,Code-Review`).

If you'd like to add a check, make sure it is something that meets the following criteria:
* automate-able 
* objective
* actionable

and then create a new GitHub Issue.

## Results

Each check returns a Pass/Fail decision, as well as a confidence score between 0 and 10.
A confidence of 0 should indicate the check was unable to achieve any real signal, and the result
should be ignored.
A confidence of 10 indicates the check is completely sure of the result.

Many of the checks are based on heuristics, contributions are welcome to improve the detection!

### Requirements
* The scorecard must only be composed of automate-able, objective data. For example, a project having 10 contributors doesn’t necessarily mean it’s more secure than a project with say 50 contributors. But, having two maintainers might be preferable to only having one -  the larger bus factor and ability to provide code reviews is objectively better.
* The scorecard criteria can be as specific as possible and not limited general recommendations. For example, for Go, we can recommend/require specific linters and analyzers to be run on the codebase.
* The scorecard can be populated for any open source project without any work or interaction from maintainers. 
* Maintainers must be provided with a mechanism to correct any automated scorecard findings they feel were made in error, provide "hints" for anything we can't detect automatically, and even dispute the applicability of a given scorecard finding for that repository.
* Any criteria in the scorecard must be actionable. It should be possible, with help, for any project to "check all the boxes".
* Any solution to compile a scorecard should be usable by the greater open source community to monitor upstream security.

## Contributing

See the [Contributing](contributing.md) documentation for guidance on how to contribute.
