# Scorecarddata

## Overview
- [What Is scorecarddata?](#what-is-Scorecarddata)
- [Why should I use this over running the scorecard CLI?](#Why-should-I-use-this-over-running-the-scorecard-CLI)
- [What does Scorecarddata do?](#what-does-scorecarddata-do)

## Using Scorecardsdata
- [How do I run scorecarddata?](#How-do-I-run-scorecarddata)
- [What are the prerequisites to run this?](#What-are-the-prerequisites-to-run)
- [Can I get additional checks other than the default?](#Can-I-get-additional-checks-other-than-the-default)
- [What are the results from the tool look like?](#What-are-the-results-from-the-tool-look-like)
- [Can I use this in CI and exclude results that I think are false positives?](#Can-I-use-this-in-CI-and-exclude-results-that-I-think-are-false-positives)
- [Using it as a server](#Using-it-as-a-server)


## Miscellaneous
- [How are the `go` dependencies parsed?](#How-are-the-go-dependencies-parsed)
- [How about support for other languages?](#How-about-support-for-other-languages)

## Overview
### what is scorecarddata?

Scorecardsdata is a tool that will parse the `go.mod/go.sum` project dependencies and fetches the [scorecard](https://github.com/ossf/scorecard) data for its dependencies. 
It uses the Google Bigquery data https://github.com/ossf/scorecard#public-data`openssf:scorecardcron.scorecard-v2_latest` to fetch the results.

### Why should I use this over running the scorecard CLI?
The scorecard CLI would take time to fetch hundreds of repositories, and the GitHub's API will be throttled. This project helps solve the problem by bringing the data from the BigQuery table, which scorecard runs as part of a weekly cron job.

### what does scorecarddata do?
```
- parses go.mod/go.sum for your project
- get the dependecies github URL's
- use the above dependencies to filter the data from Bigquery which scorecard cron jobs updates every week.
- export the results as json
```
## Using Scorecardsdata

### How do I run scorecarddata?
`scorecarddata go  -m /home/sammy/go/src/github.com/naveensrinivasan/kubernetes --GOOGLE_CLOUD_PROJECT openssf| jq`

### What are the prerequisites to run? 

- Google cloud account
- https://cloud.google.com/bigquery/public-data

### Can I get additional checks other than the default?
Yes, these are options within command line.
```
./scorecarddata --help
scorecarddata uses the scorecard bigquery to fetch results for dependecies.

Usage:
  scorecarddata [flags]
  scorecarddata [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  go          Parses go.mod dependecies and fetches the data from scorecard bigquery for those repositories.
  help        Help about any command
  server      A brief description of your command

Flags:
      --GOOGLE_CLOUD_PROJECT string    The ENV variable that will be used in the BigQuery for querying.
      --config string                  config file (default is $HOME/.scorecarddata.yaml)
      --exclusions-file string         A file with exclusions comma separated by check and value. Example Code-Review,github.com/ossf/scorecard
  -h, --help                           help for scorecarddata
      --scorecard_checks stringArray   The scorecard checks to filter by.Example CI-Tests,Binary-Artifacts etc.https://github.com/ossf/scorecard/blob/main/docs/checks.md (default [Code-Review,Branch-Protection,Pinned-Dependencies,Dependency-Update-Tool,Fuzzing])
  -t, --toggle                         Help message for toggle

Use "scorecarddata [command] --help" for more information about a command.
```

### What are the results from the tool look like?
```json=
[
 {
    "Name": "github.com/containerd/ttrpc",
    "Check": "Pinned-Dependencies",
    "Score": 7,
    "Details": "Warn: dependency not pinned by hash (job 'Run Protobuild'): .github/workflows/ci.yml:98",
    "Reason": "dependency not pinned by hash detected -- score normalized to 7"
  },
  {
    "Name": "github.com/kisielk/errcheck",
    "Check": "Pinned-Dependencies",
    "Score": 7,
    "Details": "Info: Third-party actions are pinned",
    "Reason": "dependency not pinned by hash detected -- score normalized to 7"
  }
]
```
### Can I use this in CI and exclude results that I think are false positives?
Yes, results can be excluded by providing an `--exclusions-file` with `check` and `repository`
```
Pinned-Dependencies,github.com/godbus/dbus
Pinned-Dependencies,github.com/kisielk/errcheck
```
`scorecarddata go -m . --GOOGLE_CLOUD_PROJECT openssf --exclusions-file ./exclusions --scorecard_checks Pinned-Dependencies`
### Using it as a server

The Scorecarddata can be used as a server to serve request for `checks` and `repositories`. 

`./scorecarddata server --GOOGLE_CLOUD_PROJECT openssf`

This will start a `HTTP` server in port `8080`.

This can be posted to the server to get results.
```json=
{
  "repositories": [
    "github.com/stretchr/objx"
  ],
  "checks": [
    "Pinned-Dependencies"
    ]
}
```

## Miscellaneous

### How are the `go` dependencies parsed?
It is bash goo :face_palm: More on this [explainshell](https://explainshell.com/explain?cmd=go+list+-m+-f+%27%7B%7Bif+not+%28or++.Main%29%7D%7D%7B%7B.Path%7D%7D%7B%7Bend%7D%7D%27+all+++%7C+grep+%22%5Egithub%22+%7C+sort+-u+%7C+cut+-d%2F+-f1-3+%7Cawk+%27%7Bprint+%241%7D%27%7C+sed+%22s%2F%5E%2F%5C%22%2F%3Bs%2F%24%2F%5C%22%2F%22%7C++tr+%27%5Cn%27+%27%2C%27+%7C+head+-c+-1)

### How about support for other languages? 
If you know how to parse the deps for other languages please do a PR.
The interface that needs to be implemented is

https://github.com/naveensrinivasan/scorecarddata/blob/70e7880c642d219f28d3d3738506c2a34d9a0882/pkg/deps/deps.go#L3 
