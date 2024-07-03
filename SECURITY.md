# OpenSSF Scorecard Security Policy

This document outlines security procedures and general policies for the
OpenSSF Scorecard project.

This policy adheres to the [vulnerability management guidance](https://www.linuxfoundation.org/security)
for Linux Foundation projects.

- [Disclosing a security issue](#disclosing-a-security-issue)
- [Vulnerability management](#vulnerability-management)
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

A maintainer will acknowledge the report within 72 hours, and will send a more
detailed response within an additional 72 hours indicating the next steps in
handling your report.

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

## Suggesting changes

If you have suggestions on how this process could be improved please submit an
issue or pull request.
