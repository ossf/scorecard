# Open Source Scorecards

Motivation: <TODO>

## Usage

The program only requires one argument to run, the name of the repo:

```shell
$ scorecards --repo=github.com/kubernetes/kubernetes
Security-MD 10 true
Contributors 10 true
Signed-Tags 7 false
Signed-Releases 0 false
Code-Review 10 true
CI-Tests 10 true
Frozen-Deps 10 true
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
