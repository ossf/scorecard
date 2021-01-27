# Check Documentation

This page contains information on how each check works and provide remediation
steps to fix the failure. All of these checks are basically "best-guesses"
currently, and operate on a set of heuristics.

They are all subject to change, and have room for improvement!
If you have ideas for things to add, or new ways to detect things,
please contribute!

## Security-MD

This check tries to determine if a project has published a security policy.
It works by looking for a file named `SECURITY.md` (case-insensitive) in a
few well-known directories.

**Remediation steps**:
- Place a security policy file `SECURITY.md` in the root directory of your
  repository. This makes it easily discoverable by a vulnerability reporter.
- The file should contain information on what constitues a vulnerability and
  a way to report it securely (e.g. issue tracker with private issue support,
  encrypted email with a published public key).

## Contributors

This check tries to determine if a project has a set of contributors from
multiple companies. It works by looking at the authors of recent commits and
checking the `Organization` field on the GitHub user profile.

**Remediation steps**:
- There is *NO* remediation work needed here. This is just to provide some
  insights on which organization(s) have contributed to the project and
  making trust decision based on that.

## Frozen-Deps

This check tries to determine if a project has declared and pinned its
dependencies. It works by looking for a set of well-known package management
lock files.

**Remediation steps**:
- Declare all your dependencies with specific versions in your package format
  file (e.g. `package.json` for npm, `requirements.txt` for python). For C/C++,
  check in the code from a trusted source and add a `README` on the specific
  version used (and the archive SHA hashes).
- If the package manager supports lock files (e.g. `package-lock.json` for npm),
  make sure to check these in the source code as well. These files maintain
  signatures for the entire dependency tree and saves from future exploitation
  in case the package is compromised.

## Signed-Tags

This check looks for cryptographically signed tags in the git history.

**Remediation steps**:
- Generate a new signing key.
- Add your key to your source hosting provider.
- Configure your key and email in git.
- Publish the tag and then sign it with this key.
- For GitHub, check out the steps
  [here](https://docs.github.com/en/github/authenticating-to-github/signing-tags#further-reading). 

## Signed-Releases

This check tries to determine if a project cryptographically signs release
artifacts. It works by looking for well-known filenames within recently
published GitHub releases.

**Remediation steps**:
- Publish the release.
- Generate a signing key.
- Download the release as an archive locally.
- Sign the release archive with this key (should output a signature file).
- Attach the signature file next to the release archive.
- For GitHub, check out the steps
  [here](https://wiki.debian.org/Creating%20signed%20GitHub%20releases).

## CI-Tests

This check tries to determine if the project run tests before pull requests are
merged. It works by looking for a set of well-known CI-system names in GitHub
`CheckRuns` and `Statuses`.

**Remediation steps**:
- Check-in scripts that run all the tests in your repository.
- Integrate those scripts with a CI/CD platform that runs it on every pull
  request (e.g.
  [GitHub Actions](https://docs.github.com/en/actions/learn-github-actions/introduction-to-github-actions),
  [Prow](https://github.com/kubernetes/test-infra/tree/master/prow),
  etc).

## Code-Review

This check tries to determine if a project requires code review before
pull requests are merged. It works by looking for a set of well-known code
review system results in GitHub Pull Requests.

**Remediation steps**:
- Follow security best practices by performing strict code reviews for every
  new pull request.
- Make "code reviews" mandatory in your repository configuration.
  E.g. [GitHub](https://docs.github.com/en/github/administering-a-repository/about-protected-branches#require-pull-request-reviews-before-merging).
- Enforce the rule for administrators / code owners as well.
  E.g. [GitHub](https://docs.github.com/en/github/administering-a-repository/about-protected-branches#include-administrators)

## CII-Best-Practices

This check tries to determine if the project has a [CII Best Practices Badge](https://bestpractices.coreinfrastructure.org/en).
It uses the URL for the Git repo and the CII API.

**Remediation steps**:
- Sign up for the [CII Best Practices program](https://bestpractices.coreinfrastructure.org/en).

## Pull-Requests

This check tries to determine if the project requires pull requests for all
changes to the default branch. It works by looking at recent commits and using
the GitHub API to search for associated pull requests.

**Remediation steps**:
- Always open a pull request for any change you intend to make, big or small.
- Make "pull requests" mandatory in your repository configuration.
  E.g. [GitHub](https://docs.github.com/en/github/administering-a-repository/about-protected-branches#require-pull-request-reviews-before-merging)
- Enforce the rule for administrators / code owners as well.
  E.g. [GitHub](https://docs.github.com/en/github/administering-a-repository/about-protected-branches#include-administrators)

## Fuzzing

This check tries to determine if the project uses a fuzzing system.
It currently works by checking if the repo name is in the
[OSS-Fuzz](https://github.com/google/oss-fuzz) project list.

**Remediation steps**:
- Integrate the project with OSS-Fuzz by following the instructions
  [here](https://google.github.io/oss-fuzz/).

## SAST

This check tries to determine if the project uses static code analysis systems.
It currently works by looking for well-known results ([CodeQL](https://securitylab.github.com/tools/codeql), etc.) in GitHub pull requests.

**Remediation steps**:
- Run CodeQL checks in your CI/CD by following the instructions
  [here](https://github.com/github/codeql-action#usage).

## Active

This check tries to determine if the project is still "actively maintained".
It currently works by looking for commits within the last 90 days.

**Remediation steps**:
- There is *NO* remediation work needed here. This is just to indicate your
  project activity and maintenance commitment.

## Branch-Protection

This check tries to determine if the project has branch protection enabled.

**Remediation steps**:
- Enable branch protection settings in your source hosting provider to avoid
  force pushes or deletion of your important branches.
- For GitHub, check out the steps
  [here](https://docs.github.com/en/github/administering-a-repository/managing-a-branch-protection-rule).
