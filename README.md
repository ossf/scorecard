# Open Source Scorecards

Motivation: <TODO>

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
