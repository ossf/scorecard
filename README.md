# Open Source Scorecards

### Motivation
Consumers of open source projects need to understand the security posture of the software they depend on, and in many cases, have the ability to automate checks or policies using this data. One proposed solution is to define a set of auomate-able and objective data to produce a security "scorecard" for projects. An organization could then create an internal policy such as "projects with a score less than X, need to be further reviewed."

In addition, this score is something that is actionable and can provide maintainers, contributors and other stakeholders concrete ways to improve the security posture of the projects they work or depend on. 

### Goals
1. Fill the gaps that prevent automated analysis and trust decisions for measuring and reporting on the security posture of open source projects. 

1. Use this data to proactively improve the security posture of the critical projects the world depends on.

### Requirements
* The scorecard must only be composed of automate-able, objective data. For example, a project having 10 contributors doesn’t necessarily mean it’s more secure than a project with say 50 contributors. But, having two maintainers might be preferable to only having one -  the larger bus factor and ability to provide code reviews is objectively better.
* The scorecard criteria can be as specific as possible and not limited general recommendations. For example, for Go, we can recommend/require specific linters and analyzers to be run on the codebase.
* The scorecard can be populated for any open source project without any work or interaction from maintainers. 
* Maintainers must be provided with a mechanism to correct any automated scorecard findings they feel were made in error, provide "hints" for anything we can't detect automatically, and even dispute the applicability of a given scorecard finding for that repository.
* Any criteria in the scorecard must be actionable. It should be possible, with help, for any project to "check all the boxes".
* Any solution to compile a scorecard should be usable by the greater open source community to monitor upstream security.
* Skeleton repos should be provided that set new projects up for success.

## Usage

The program only requires one argument to run, the name of the repo:

```shell
$ scorecards --repo=github.com/kubernetes/kubernetes
2020/10/09 10:25:12 Starting [Code-Review]
2020/10/09 10:25:12 Starting [Contributors]
2020/10/09 10:25:12 Starting [Frozen-Deps]
2020/10/09 10:25:12 Starting [Signed-Releases]
2020/10/09 10:25:12 Starting [Security-MD]
2020/10/09 10:25:12 Starting [Signed-Tags]
2020/10/09 10:25:12 Starting [CI-Tests]
2020/10/09 10:25:12 Finished [Security-MD]
2020/10/09 10:25:14 Finished [Contributors]
2020/10/09 10:25:16 Finished [Signed-Tags]
2020/10/09 10:25:16 Finished [Signed-Releases]
2020/10/09 10:25:25 Finished [Code-Review]
2020/10/09 10:25:28 Finished [CI-Tests]
2020/10/09 10:25:38 Finished [Frozen-Deps]

RESULTS
-------
CI-Tests true 10
Code-Review true 10
Contributors true 10
Frozen-Deps true 10
Security-MD true 10
Signed-Releases false 0
Signed-Tags false 7
```

You'll probably also need to set an Oauth token to avoid rate limits.
You can create a personal access token by following these steps: https://docs.github.com/en/free-pro-team@latest/developers/apps/about-apps#personal-access-tokens

Set that as an environment variable:

```shell
export GITHUB_OAUTH_TOKEN=<your token>
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

To see detailed information on how each check works, see the check-specific documentation pages.

## Results

Each check returns a pass/fail decision, as well as a confidence score between 0 and 10.
A confidence of 0 should indicate the check was unable to achieve any real signal, and the result
should be ignored.
A confidence of 10 indicates the check is completely sure of the result.

Many of the checks are based on heuristics, contributions are welcome to improve the detection!

## Contributing

See the [Contributing](contributing.md) documentation for guidance on how to contribute.
