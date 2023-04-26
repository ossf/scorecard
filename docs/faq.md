# Frequently Asked Questions

This page answers frequently asked questions about Scorecard, including its purpose, usage, and checks. This page is continually updated. If you would like to add a question, please [contribute](../CONTRIBUTING.md)!

## Installation / Usage
  - [Can I preview my project's score?](#can-i-preview-my-projects-score)
  - [What is the difference between Scorecard and other Code Scanning tools?](#what-is-the-difference-between-scorecard-and-other-code-scanning-tools)
  - [Wasn't this project called "Scorecards" (plural)?](#wasnt-this-project-called-scorecards-plural)

## Check-Specific Questions
  - [Binary-Artifacts: Can I allowlist testing artifacts?](#binary-artifacts-can-i-allowlist-testing-artifacts)
  - [Code-Review: Can it ignore bot commits?](#code-review-can-it-ignore-bot-commits)
  - [Fuzzing: Does Scorecard accept custom fuzzers?](#fuzzing-does-scorecard-accept-custom-fuzzers)
  - [Pinned-Dependencies: Will Scorecard detect unpinned dependencies in tests with Dockerfiles?](#pinned-dependencies-will-scorecard-detect-unpinned-dependencies-in-tests-with-dockerfiles)
  - [Pinned-Dependencies: Can I use version pinning instead of hash pinning?](#pinned-dependencies-can-i-use-version-pinning-instead-of-hash-pinning)
  - [Signed-Releases: Why sign releases?](#signed-releases-why-sign-releases)
  - [Dependency-Update-Tool: Why should I trust recommended updates are safe?](#dependency-Update-Tool-why-should-i-trust-recommended-updates-are-safe) 

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

Scorecard can show the dependencies that are referred to in tests like Dockerfiles, so it could be a great way for you to fix those dependencies and avoid the vulnerabilities related to version pinning dependencies. To see more about the benefits of hash pinning instead of version pinning, please see the [Pinned-Dependencies check description](checks.md#pinned-dependencies)

### Pinned-Dependencies: Can I use version pinning instead of hash pinning?
Version pinning is a significant improvement over not pinning your dependencies. However, it still leaves your project vulnerable to tag-renaming attacks (where a dependency's tags are deleted and recreated to point to a malicious commit).

The OpenSSF therefore recommends hash pinning instead of version pinning, along with the use of dependency update tools such as dependabot to keep your dependencies up-to-date.

Please see the [Pinned-Dependencies check description](checks.md#pinned-dependencies) for a better understanding of the benefits of the Hash Pinning.

### Signed-Releases: Why sign releases?

Currently, the main benefit of [signed releases](checks.md#signed-releases) is the guarantee that a specific artifact was released by a source that you approve or attest is reliable.

However, there are already moves to make it even more relevant. For example, the OpenSSF is working on [implementing signature verification for NPM packages](https://github.blog/2022-08-08-new-request-for-comments-on-improving-npm-security-with-sigstore-is-now-open/) which would allow a consumer to automatically verify if the package they are downloading was generated through a reliable builder and if it is correctly signed.

Signing releases already has some relevance and it will soon offer even more security benefits for both consumers and maintainers.

### Dependency-Update-Tool: Why should I trust recommended updates are safe?

Both dependabot and renovatebot won't update your dependencies immediately. They have some precautions to make sure a release is reasonable / won't break your build (see [dependabot documentation](https://docs.github.com/en/code-security/dependabot/dependabot-security-updates/about-dependabot-security-updates#about-compatibility-scores)).

You can either configure the tools to only update your dependencies once a week or once a month. This way, if a malicious version is released, it's very likely that it'll be reported and removed before it even gets suggested to you. Besides, there's also the benefit that it gives you the chance to validate the new release before merging if you want to.

