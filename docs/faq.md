# Frequently Asked Questions

This page describes the most frequently asked questions about the Scorecard use, checks or relevance. These answers are continually changing. If you know any other question that could be interesting to be covered here, please [contribute](../CONTRIBUTING.md)!

## Installation / Usage
  - [Can I have a preview of the score?](#can-i-have-a-preview-of-the-score)
  - [What is the difference between Scorecard and other Code Scanning Tools?](#what-is-the-difference-between-scorecard-and-other-code-scanning-tools)

## Check-Specific Questions
  - [Binary-Artifacts: Is it possible to set up a blocklist to check Binary-Artifacts?](#binary-artifacts-is-it-possible-to-set-up-a-blocklist-to-check-binary-artifacts)
  - [Code-Review: Can I set Code-Review check to ignore bot commits?](#code-review-can-i-set-code-review-check-to-ignore-bot-commits)
  - [Fuzzing: Scorecard accepts custom fuzzers and libfuzzer?](#fuzzing-scorecard-accepts-custom-fuzzers-and-libfuzzer)
  - [Pinned-Dependencies: Will the scorecard see not-pinned dependencies in tests with Dockerfiles?](#pinned-dependencies-will-the-scorecard-see-not-pinned-dependencies-in-tests-with-dockerfiles)
  - [Pinned-Dependencies: Can I use version pinning instead of hash pinning?](#pinned-dependencies-can-i-use-version-pinning-instead-of-hash-pinning)
  - [Signed-Releases: Why would I sign releases?](#signed-releases-why-would-i-sign-releases)

________________________________________________________________________________
________________________________________________________________________________

## Installation / Usage

### Can I have a preview of the score?

Yes, a preview of the Scorecard scores can be seen at https://api.securityscorecards.dev/projects/github.com/<username_or_org>/<repository_name>/ for the repositories tracked by the Scorecard Project for being considered relevant in the Open Source scenario.

Anyhow, you can generate the scores for any public repository through the CLI following the steps bellow:

1. [Installation](https://github.com/joycebrum/scorecard#installation)
1. [Authentication](https://github.com/joycebrum/scorecard#authentication)
1. [Basic Usage](https://github.com/joycebrum/scorecard#basic-usage)

### What is the difference between Scorecard and other Code Scanning Tools?

To determine exactly what are the differences it is important to analyse both tools and compare them. But, usually, the code scanning tools are focused on one or two specific types of vulnerabilities, while the Scorecard's Checks are focused on the overall security posture of the project. That's because the Scorecard is related to the Security Best Practices and whether the project is following them or not.

## Check-Specific Questions

### Binary-Artifacts: Is it possible to set up a blocklist to check Binary-Artifacts?

It is still not possible to do that. However, the Scorecard team is working on this feature in the issue [ossf/scorecard#1270](https://github.com/ossf/scorecard/issues/1270).


### Code-Review: Can I set Code-Review check to ignore bot commits?

This is quite a complex question to be answered. Right now, there is no way to do that and here are some pros and cons on allowing the users to set up a ignore list with bots.

- Pros: Some bots have a very frequent and automated job and, for some projects, reviewing every change is not feasible or reasonable.
- Cons: Any bot can be compromised (its credentials can be compromised, for example), or considering that the commits are not being signed, an attacker could easily send a commit spoofing the bot. This means that the bot having a not supervised access to the main branch could potentially be a security risk.

Anyhow, this is being discussed by the Scorecard Team, for more informations about this topic please see the issue [Code Review Check handle commits made by version bump bots](https://github.com/ossf/scorecard/issues/2302).

### Fuzzing: Scorecard accepts custom fuzzers and libfuzzer?

The Fuzzing Check detects OSS Fuzz, ClusterFuzzLite, OneFuzz and Go custom checks, thus it only catches custom fuzzing for GoLang. So, the check doesn’t detect custom use of libfuzzer, but some of these fuzzing tools might be using libfuzzers under the hood.

To see more about how the Fuzzing Check determines whether the project uses fuzzing or not, see [Fuzzing Check](https://github.com/ossf/scorecard/blob/main/docs/checks.md#fuzzing). 

### Pinned-Dependencies: Will the scorecard see not-pinned dependencies in tests with Dockerfiles?

Scorecard can show the dependencies that are referred to in tests like Dockerfiles, so it could be a great way for you to fix those dependencies and avoid the vulnerabilities related to version pinning dependencies. To see more about the benefits of hash pinning instead of version pinning, please see the [Pinned-Dependencies Check Description](/checks.md#pinned-dependencies)

### Pinned-Dependencies: Can I use version pinning instead of hash pinning?
It is not encouraged. The OpenSSF recommends the use of hash pinning instead of version pinning declarations in order to reduce several security risks. Please take a look at the [Pinned-Dependencies Check Description](/checks.md#pinned-dependencies) to a better understanding of the benefits of the Hash Pinning.


### Signed-Releases: Why would I sign releases?

The main benefit that the [signed releases](/checks.md#signed-releases) could bring for now is the guarantee that a specific artifact was released by a source that you approve or you say is reliable.

Although, there are already moves to make it even more influential on the download process. The OpenSSF is working on [Implementing the signature verification with NPM packages](https://github.blog/2022-08-08-new-request-for-comments-on-improving-npm-security-with-sigstore-is-now-open/) which would allow a consumer to automatically verify if the package they are downloading was generated through a reliable builder and if it is correctly signed.

The Releases Signature already has some benefits and it is moving to a future with even more security benefits both for consumers and maintainers.
