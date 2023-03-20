# Frequently Asked Questions

This page answers frequently asked questions about Scorecard, including its purpose, usage, and checks. This page is continually updated. If you would like to add a question, please [contribute](../CONTRIBUTING.md)!

## Installation / Usage
  - [Can I preview my project's score?](#can-i-preview-my-projects-score)
  - [What is the difference between Scorecard and other Code Scanning tools?](#what-is-the-difference-between-scorecard-and-other-code-scanning-tools)
  - [Wasn't this project called "Scorecards" (plural)?](#wasnt-this-project-called-scorecards-plural)

## Check-Specific Questions
- [Frequently Asked Questions](#frequently-asked-questions)
  - [Installation / Usage](#installation--usage)
  - [Check-Specific Questions](#check-specific-questions)
  - [Installation / Usage](#installation--usage-1)
    - [Can I preview my project's score?](#can-i-preview-my-projects-score)
    - [What is the difference between Scorecard and other Code Scanning tools?](#what-is-the-difference-between-scorecard-and-other-code-scanning-tools)
    - [Wasn't this project called "Scorecards" (plural)?](#wasnt-this-project-called-scorecards-plural)
  - [Check-specific Questions](#check-specific-questions-1)
    - [Binary-Artifacts: Can I allowlist testing artifacts?](#binary-artifacts-can-i-allowlist-testing-artifacts)
    - [Code-Review: Can it ignore bot commits?](#code-review-can-it-ignore-bot-commits)
    - [Fuzzing: Does Scorecard accept custom fuzzers?](#fuzzing-does-scorecard-accept-custom-fuzzers)
    - [Pinned-Dependencies: Will Scorecard detect unpinned dependencies in tests with Dockerfiles?](#pinned-dependencies-will-scorecard-detect-unpinned-dependencies-in-tests-with-dockerfiles)
    - [Pinned-Dependencies: Can I use version pinning instead of hash pinning?](#pinned-dependencies-can-i-use-version-pinning-instead-of-hash-pinning)
    - [Signed-Releases: Why sign releases?](#signed-releases-why-sign-releases)
    - [Branch-Protection: How to setup a 10/10 branch protection on GitHub?](#branch-protection-how-to-setup-a-1010-branch-protection-on-github)

---

## Installation / Usage

### Can I preview my project's score?

Yes.

Over a million projects are automatically tracked by the Scorecard project. These projects' scores can be seen at https://api.securityscorecards.dev/projects/github.com/<username_or_org>/<repository_name>.

You can also use the CLI to generate scores for any public repository by following these steps:

1. [Installation](https://github.com/ossf/scorecard#installation)
2. [Authentication](https://github.com/ossf/scorecard#authentication)
3. [Basic Usage](https://github.com/ossf/scorecard#basic-usage)

### What is the difference between Scorecard and other Code Scanning tools?

Most code scanning tools are focused on detecting specific vulnerabilities already existing in your codebase. Scorecard, however, is focused on improving the project's overall security posture by helping it adopt best practices. The best solution for your project may well be to adopt Scorecard along with other tools!

### Wasn't this project called "Scorecards" (plural)?

Yes, kind of. The project was initially called "Security Scorecards" but that form wasn't used consistently. In particular, the repo was named "scorecard" and so was the program. Over time people started referring to either form (singular and plural) and the inconsitency became prevalent. To end this situation the decision was made to consolidate over the use of the singular form in keeping with the repo and program name, drop the "Security" part and use "OpenSSF" instead to ensure uniqueness. One should therefore refer to this project as "OpenSSF Scorecard" or "Scorecard" for short.

## Check-specific Questions

### Binary-Artifacts: Can I allowlist testing artifacts?

Scorecard lowers projects' scores whenever it detects binary artifacts. However, many projects use binary artifacts strictly for testing purposes.

While it isn't currently possible to allowlist such binaries, the Scorecard team is working on this feature ([#1270](https://github.com/ossf/scorecard/issues/1270)).

### Code-Review: Can it ignore bot commits?

This is quite a complex question. Right now, there is no way to do that. Here are some pros and cons on allowing users to set up an ignore-list for bots.

- Pros: Some bots run very frequently; for some projects, reviewing every change is therefore not feasible or reasonable.
- Cons: Bots can be compromised (their credentials can be compromised, for example). Or if commits are not signed, an attacker could easily send a commit spoofing the bot. This means that a bot having unsupervised write access to the repository could be a security risk.

However, this is being discussed by the Scorecard Team ([#2302](https://github.com/ossf/scorecard/issues/2302)).

### Fuzzing: Does Scorecard accept custom fuzzers?

Currently only for projects written in Go.

For more information, see the [Fuzzing check description](https://github.com/ossf/scorecard/blob/main/docs/checks.md#fuzzing). 

### Pinned-Dependencies: Will Scorecard detect unpinned dependencies in tests with Dockerfiles?

Scorecard can show the dependencies that are referred to in tests like Dockerfiles, so it could be a great way for you to fix those dependencies and avoid the vulnerabilities related to version pinning dependencies. To see more about the benefits of hash pinning instead of version pinning, please see the [Pinned-Dependencies check description](/checks.md#pinned-dependencies)

### Pinned-Dependencies: Can I use version pinning instead of hash pinning?
Version pinning is a significant improvement over not pinning your dependencies. However, it still leaves your project vulnerable to tag-renaming attacks (where a dependency's tags are deleted and recreated to point to a malicious commit).

The OpenSSF therefore recommends hash pinning instead of version pinning, along with the use of dependency update tools such as dependabot to keep your dependencies up-to-date.

Please see the [Pinned-Dependencies check description](/checks.md#pinned-dependencies) for a better understanding of the benefits of the Hash Pinning.

### Signed-Releases: Why sign releases?

Currently, the main benefit of [signed releases](/checks.md#signed-releases) is the guarantee that a specific artifact was released by a source that you approve or attest is reliable.

However, there are already moves to make it even more relevant. For example, the OpenSSF is working on [implementing signature verification for NPM packages](https://github.blog/2022-08-08-new-request-for-comments-on-improving-npm-security-with-sigstore-is-now-open/) which would allow a consumer to automatically verify if the package they are downloading was generated through a reliable builder and if it is correctly signed.

Signing releases already has some relevance and it will soon offer even more security benefits for both consumers and maintainers.

### Branch-Protection: How to setup a 10/10 branch protection on GitHub?

To get a 10/10 score for Branch-Protection check using a non-admin token, you should have the following settings for your main branch:

![image](/docs/design/images/branch-protection-settings-non-admin-token.png)

When using an admin token, Scorecard can verify if a few other important settings are ensured:

![image](/docs/design/images/branch-protection-settings-admin-token.png)

It's important to reiterate that Branch-Protection score is Titer-based. If a setting from Tier 1 is not satisfied, it does not matter that all other settings are met, the score will be truncated up the Tier's maximum. In this case, 3/10. The following table shows the relation between branch protection settings on GitHub and the score Tier:

| Name                                                                                                     | Status                          | Required only for admin token | Tier |
| -------------------------------------------------------------------------------------------------------- | ------------------------------- | ----------------------------- | ---- |
| Allow force pushes                                                                                       | Disabled                        | -                             | 1    |
| Allow deletions                                                                                          | Disabled                        | -                             | 1    |
| Do not allow bypassing the above settings                                                                | Enabled                         | Yes                           | 1    |
| Require a pull request before merging > Require Approvals                                                | Enabled with at least 1         | -                             | 2    |
| Require status checks to pass before merging > Require branches to be up to date before merging          | Enabled                         | Yes                           | 2    |
| Require a pull request before merging > Require approval of the most recent reviewable push              | Enabled                         | Yes                           | 2    |
| Require status checks to pass before merging > Status Checks                                             | At least 1                      | -                             | 3    |
| Require a pull request before merging > Require Approvals                                                | Enabled with at least 2         | -                             | 4    |
| Require a pull request before merging > Require review from Code Owners                                  | Enabled and has CODEOWNERS file | Yes                           | 4    |
| Require a pull request before merging > Dismiss stale pull request approvals when new commits are pushed | Enabled                         | Yes                           | 5    |
