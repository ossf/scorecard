# `multi-scorecard`

This program runs OpenSSF Scorecard over many repositories using a [GitHub App](https://docs.github.com/en/apps/creating-github-apps/about-creating-github-apps/about-creating-github-apps) credential.
GitHub is queried to determine the orgs and repos the app is installed on to determine which repos to run Scorecard over.

Results are printed to stdout in a JSON array.

*`multi-scorecard` was originally featured as part of [Jeff Mendoza](https://github.com/jeffmendoza) and [Stephen Augustus](https://github.com/justaugustus)' SOSS Fusion talk, "Scorecard at Scale: Old and New Possibilities for Lifting Security on All Repositories".*

- [Session page with slides](https://sched.co/1hcPq)
- [Session recording](https://youtu.be/-XZqbO3hGcw?si=eGicz0sjgiIRhol4)
- [Previous source repository](https://github.com/jeffmendoza/multi-scorecard)

## Usage

A [GitHub App](https://docs.github.com/en/apps/creating-github-apps/about-creating-github-apps/about-creating-github-apps) must be created and installed on the repositories you wish to scan.

To install:

```console
go get github.com/ossf/scorecard/cmd/multi-scorecard@multi-scorecard
```

To run:

```console
multi-scorecard -appid 1234 -keyfile my-app.private-key.pem > results.json
```

Where `1234` is the App ID of the app, and `my-app.private-key.pem` is the private key file of the app.
