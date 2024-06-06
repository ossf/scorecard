# Repository Guidelines

This document attempts to outline a structure for creating and associating
GitHub repositories with the OpenSSF Scorecard project.
<!-- TODO: It also describes how and when repositories are removed. -->

<!-- TODO: Do we need a separate issue template for these requests? e.g.,
Requests for creating, transferring, modifying, or archiving repositories can be made by [opening a request](https://github.com/ossf/scorecard/issues/new/choose) against this repository.
-->

- [Approval](#approval)
- [Requirements](#requirements)
- [Donated repositories](#donated-repositories)
  - [Copyright headers](#copyright-headers)
- [Attribution](#attribution)

## Approval

New repositories require approval from the OpenSSF Scorecard Steering Committee.

## Requirements

The following requirements apply to all OpenSSF Scorecard repositories:

- Must be identified in the OpenSSF Scorecard project documentation
- Must reside in the [OpenSSF GitHub organization](https://github.com/ossf)
- Must utilize the topic [`openssf-scorecard`](https://github.com/topics/openssf-scorecard) (ref: [managing topics](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/classifying-your-repository-with-topics))
- Must adopt the OpenSSF Scorecard [Code of Conduct](/CODE_OF_CONDUCT.md)
- Must adopt an appropriate license, in compliance with the Intellectual Property Policy of OpenSSF Scorecard [charter](/CHARTER.md)
- Must include headers across all files that attribute copyright as follows:

  ```text
  Copyright [YYYY] OpenSSF Scorecard Authors
  ```

- Must enforce usage of the Developer Certificate of Origin (DCO) via the [DCO GitHub Application](https://github.com/apps/dco)
- All privileges to the repository must be defined via [GitHub teams](https://docs.github.com/en/organizations/organizing-members-into-teams/about-teams), [instead of individuals](https://github.com/ossf/tac/blob/main/policies/access.md#teams-not-individuals)
- All code review permissions must be defined via [CODEOWNERS](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners)
- All contributors with privileges to the repository must also be active members of the OpenSSF Scorecard project

## Donated repositories

The OpenSSF Scorecard project may at times accept repository donations.

Donated repositories must:

- Adhere to the [requirements for all project repositories](#requirements)
<!-- TODO: Need documentation on license scans and acceptable licenses for dependencies e.g.,
- Licenses of dependencies are acceptable; please review the [allowed-third-party-license-policy.md](https://github.com/cncf/foundation/blob/main/allowed-third-party-license-policy.md) and [exceptions](https://github.com/cncf/foundation/tree/main/license-exceptions). If your dependencies are not covered, then please open a `License Exception Request` issue in [cncf/foundation](https://github.com/cncf/foundation/issues) repository.
-->

### Copyright headers

The addition of required copyright headers to code created by the contributors
can occur post-transfer, but should ideally occur shortly thereafter.

***Note that copyright notices should only be modified or removed by the people or organizations named in the notice.***

## Attribution

These guidelines were drafted with inspiration from the [Kubernetes project's repository guidelines](https://github.com/kubernetes/community/blob/e65e7141f8c2bb82f33762c35e19059e9c5d034e/github-management/kubernetes-repositories.md).
