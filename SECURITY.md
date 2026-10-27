# OpenSSF Scorecard Security Policy

This document outlines security procedures and general policies for the
OpenSSF Scorecard project.

This policy adheres to the [vulnerability management guidance](https://www.linuxfoundation.org/security)
for Linux Foundation projects.

CRA stewardship: This project is supported under the Linux Foundation CRA
stewardship framework, as described at https://www.linuxfoundation.org/security.
Security vulnerabilities should be reported through the mechanisms described
below, which we will coordinate with our CRA steward. For actively exploited
vulnerabilities and severe incidents that may require CRA escalation, please
use the project's emergency security reporting mechanisms as appropriate.

- [Disclosing a security issue](#disclosing-a-security-issue)
- [Vulnerability management](#vulnerability-management)
- [Privileged workflows](#privileged-workflows)
- [Binary files in the repository](#binary-files-in-the-repository)
- [Suggesting changes](#suggesting-changes)

## Disclosing a security issue

The OpenSSF Scorecard maintainers take all security issues in the project
seriously. Thank you for improving the security of OpenSSF Scorecard. We
appreciate your dedication to responsible disclosure and will make every effort
to acknowledge your contributions.

OpenSSF Scorecard leverages GitHub's private vulnerability reporting.

To learn more about this feature and how to submit a vulnerability report,
review [GitHub's documentation on private reporting](https://docs.github.com/code-security/security-advisories/guidance-on-reporting-and-writing-information-about-vulnerabilities/privately-reporting-a-security-vulnerability).

Here are some helpful details to include in your report:

- a detailed description of the issue
- the steps required to reproduce the issue
- versions of the project that may be affected by the issue
- if known, any mitigations for the issue

A maintainer will acknowledge the report within three (3) business days, and
will send a more detailed response within an additional three (3) business days
indicating the next steps in handling your report.

If you've been unable to successfully draft a vulnerability report via GitHub
or have not received a response during the alloted response window, please
reach out via the [OpenSSF security contact email](mailto:security@openssf.org).

After the initial reply to your report, the maintainers will endeavor to keep
you informed of the progress towards a fix and full announcement, and may ask
for additional information or guidance.

## Vulnerability management

When the maintainers receive a disclosure report, they will assign it to a
primary handler.

This person will coordinate the fix and release process, which involves the
following steps:

- confirming the issue
- determining affected versions of the project
- auditing code to find any potential similar problems
- preparing fixes for all releases under maintenance

Fixed vulnerabilities are published as
[GitHub Security Advisories](https://github.com/ossf/scorecard/security/advisories)
on the affected repository, which feed the GitHub Advisory Database and OSV.
The release notes of the release that contains the fix also mention it.

## Privileged workflows

Two workflows run in a privileged context on pull requests. The maintainers
reviewed both:

- [`verify.yml`](.github/workflows/verify.yml) runs on `pull_request_target`
  to check the PR title. It checks out the base branch, never the PR, passes
  the title through an environment variable, and its token can only write
  check results.
- [`scdiff.yml`](.github/workflows/scdiff.yml) runs on PR comments. It runs
  the PR's code to compare Scorecard results before and after the change,
  using the repository's GitHub and GitLab tokens. It runs only when the
  commenter has write access to this repository, and only when the PR hasn't
  changed since the comment. The comparison script comes from `main`, not from
  the PR. Maintainers review the PR before asking for a run.

No other workflow uses `pull_request_target`, `issue_comment`, or
`workflow_run`.

## Binary files in the repository

The repository contains no release binaries. The binary files it does
contain are test fixtures:

- [`checks/testdata/binaryartifacts/`](checks/testdata/binaryartifacts/):
  a Mach-O executable, a WebAssembly module, and two jars. They test the
  Binary-Artifacts check, which has to detect files like these.
- `basic.tar.gz` in [`clients/githubrepo/testdata/`](clients/githubrepo/testdata/)
  and [`clients/gitlabrepo/testdata/`](clients/gitlabrepo/testdata/), and
  `basic.zip` in [`clients/azuredevopsrepo/testdata/`](clients/azuredevopsrepo/testdata/):
  repository archives for the tests of each client's archive handling.

Images under [`artwork/`](artwork/) and [`docs/`](docs/), and
[`CHARTER.pdf`](CHARTER.pdf), are documentation.

## Suggesting changes

If you have suggestions on how this process could be improved please submit an
issue or pull request.
