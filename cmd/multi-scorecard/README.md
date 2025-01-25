# Multi Scorecard

This program runs [OpenSSF Scorecard](https://github.com/ossf/scorecard) over
many repositories using a [GitHub
App](https://docs.github.com/en/apps/creating-github-apps/about-creating-github-apps/about-creating-github-apps)
credential. GitHub is queried to determine the orgs and repos the app is
installed on to determine which repos to run Scorecard over. Results are
printed to stdout in a JSON array.

## Usage

A [GitHub
App](https://docs.github.com/en/apps/creating-github-apps/about-creating-github-apps/about-creating-github-apps)
must be created and installed on the repositories you wish to scan.

To install:

```
go get github.com/jeffmendoza/multi-scorecard@latest
```

To run:

```
multi-scorecard -appid 1234 -keyfile my-app.private-key.pem > results.json
```

Where `1234` is the App ID of the app, and `my-app.private-key.pem` is the
private key file of the app.
