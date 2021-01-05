# Check Documentation

This page contains information on how each check works and calculates scores.
All of these checks are basically "best-guesses" currently, and operate on a set of heuristics.

They are all subject to change, and have much room for improvement!
If you have ideas for things to add, or new ways to detect things, please contribute!

## Security-MD

This check tries to determine if a project has published security policies.
It works by looking for a file named `SECURITY.md` (case-insensitive) in a few well-known directories.

## Contributors

This check tries to determine if a project has a set of contributors from multiple companies.
It works by looking at the authors of recent commits and checking the `Organization` field on the GitHub user profile.

## Frozen-Deps

This check tries to determine if a project has declared and pinned its dependencies.
It works by looking for a set of well-known package management lock files.

## Signed-Tags

This check looks for cryptographically signed tags in the git history.

## Signed-Releases

This check tries to determine if a project cryptographically signs release artifacts.
It works by looking for well-known filenames within recently published GitHub releases.

## CI-Tests

This check tries to determine if the project run tests before pull requests are merged.
It works by looking for a set of well-known CI-system names in GitHub `CheckRuns` and `Statuses`.

## Code-Review

This check tries to determine if a project requires code review before pull requests are merged.
It works by looking for a set of well-known code review system results in GitHub Pull Requests.

## CII-Best-Practices

This check tries to determine if the project has a [CII Best Practices Badge](https://bestpractices.coreinfrastructure.org/en).
It uses the URL for the Git repo and the CII API.

## Pull-Requests

This check tries to determine if the project requires pull requests for all changes to the default branch.
It works by looking at recent commits and using the GitHub API to search for associated pull requests.

## Fuzzing

This check tries to determine if the project uses a fuzzing system.
It currently works by checking if the repo name is in the [OSS-Fuzz](https://github.com/google/oss-fuzz) project list.

## SAST

This check tries to determine if the project uses static code analysis systems.
It currently works by looking for well-known results ([CodeQL](https://securitylab.github.com/tools/codeql), etc.) in GitHub pull requests.

## Active

This check tries to determine if the project is still "actively maintained".
It currently works by looking for releases or commits within the last 90 days.

## Branch-Protection

This check tries to determine if the project has branch protection enabled.
