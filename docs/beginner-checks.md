# Getting Started with Scorecard Checks for Supply Chain Security

If you are an open-source developer, you know that maintaining the security posture of a project can feel like a full-time job, but it can be easier with a little guidance. [Scorecard](https://github.com/ossf/scorecard#openssf-scorecard), an automated tool created [by the Open Source Security Foundation](https://openssf.org/) (OpenSSF), checks projects for security best practices and scores projects on a scale between 1 and 10. Maintainers can use the Scorecard checks as guidelines to improve their project's security practices.

Choosing which Scorecard checks to get started with as a project maintainer can be overwhelming. This page walks through some of the most important checks to start with for project improvement, focusing on the ones that give you the biggest payoff versus effort. They're broken down into three categories based on stages of the development process: setting up your project, accepting contributions from others, and packaging the project to release to the world.

## 1. Setting up your project

The setup phase of an open-source project can be hectic, which can make it difficult to prioritize security concerns. Start your project off strong by focusing on Scorecard checks that help you secure your project dependencies and workflows.

- Secure your dependencies with the Vulnerabilities check and the Dependency-Update-Tool check
- Secure your workflows with the Token-Permissions check

### The Vulnerabilities and Dependency-Update-Tool checks secure your dependencies

Vulnerabilities are probably the most familiar security risk. By running Scorecard’s [Vulnerabilities check](https://github.com/ossf/scorecard/blob/main/docs/checks.md#vulnerabilities), you’ll get information about known vulnerabilities in your project, either through your codebase or through your direct and (for most languages) indirect dependencies. Tracking these can be tricky, but it’s important: [95% of vulnerable dependencies are transitive](https://www.endorlabs.com/state-of-dependency-management).

If vulnerabilities are found in your dependencies, there are a few options:

- Update the dependency to a non-vulnerable version (if available)
- Submit a patch to the vulnerable project
- Replace the dependency with a non-vulnerable dependency
- Remove the dependency and write code to take its place
- If you are sure a vulnerability does not impact your project, you may ignore the dependency by creating an [osv-scanner.toml](https://google.github.io/osv-scanner/configuration/#ignore-vulnerabilities-by-id) file in your project's root directory.

If you have handled the vulnerabilities in your dependencies and are still not satisfied with your score for this check, make sure there are no open, unfixed vulnerabilities in your project’s own codebase. Once you have dealt with those, your score should improve.

Next, Scorecard’s [Dependency-Update-Tool check](https://github.com/ossf/scorecard/blob/main/docs/checks.md#dependency-update-tool) encourages developers to keep their dependencies up to date, which is a great way to stay on top of security updates. This check awards a high score to a project if it uses a dependency update tool such as [Dependabot](https://docs.github.com/code-security/dependabot), [Renovate bot](https://docs.renovatebot.com/), or [PyUp](https://github.com/pyupio/pyup#readme). Using one of these tools helps streamline security processes by notifying you when vulnerabilities have surfaced in your dependencies or when new versions of your dependencies become available.

Automated processes like these save you time and are highly configurable; for example, you can set your bot to update dependencies every day or every week at the same time.

If you want to increase your score in this category, sign up for automatic updates with a dependency update tool. Do note, however, that this check can only assess whether the dependency update tool is enabled, not that it is running. To maximize your benefit from this check, be sure that you consistently run your dependency update tool.

### Token-Permissions check helps you secure your workflows

We suggest addressing the [Token-Permissions check](https://github.com/ossf/scorecard/blob/main/docs/checks.md#token-permissions) next because it takes just a few minutes to “set it and forget it” and secure your workflows. The check warns you when your project’s top-level tokens have `write` access instead of the more restrictive `read` access. Not all `write` access permissions need to be eliminated; some workflows may genuinely require them. However, ensuring your top-level permissions have `read` access helps your project adhere to the principle of least privilege, which means that permissions are granted based on the minimal necessary access to perform a function. Projects that have top-level tokens set to `write` access for their automated workflows may be vulnerable to malicious code execution. By simply switching `write` to `read` you’ll be able to close off a powerful attack vector (exploiting compromised tokens that have write permissions and using them to inject malicious code into the project).

To change the default setting for token permissions, add the following to the top of your workflow:

```
permissions:
  contents: read
```

When you add a GitHub Action, be sure to read the Action’s docs to see if it needs any additional permissions; this information is usually prominent in the Action’s README.

## 2. Manage contributions to your project

As projects grow, they generally start including contributions from others. Contributors can expand your project’s scope and maturity, but they can also introduce security risk. To protect your project at this stage, we recommend improving the Branch Protection check, which allows you, the maintainer, to define rules that require certain workflows for certain branches.

### Branch Protection reduces the risks of errors and hacks

The [Branch Protection check](https://github.com/ossf/scorecard/blob/main/docs/checks.md#branch-protection) can help protect your code from unvetted changes. You can choose either or both of the following options:

- Require code review. Select this if your project has more than one maintainer. Requiring review before changes are merged is one of the strongest protections you can give your code. This will also improve your [Code Review check](https://github.com/ossf/scorecard/blob/main/docs/checks.md#code-review) score.
- Require [status checks](https://docs.github.com/pull-requests/collaborating-with-pull-requests/collaborating-on-repositories-with-code-quality-features/about-status-checks). All projects would benefit from selecting this option. It ensures that all Continuous Integration (CI) tests chosen by the maintainer have passed before a change is accepted, helping you catch mistakes early on in the development process. This will also improve your [CI Test check](https://github.com/ossf/scorecard/blob/main/docs/checks.md#ci-tests) score.

## 3. Package and release your project

Deciding how to securely share your code can be difficult. Building locally on your laptop may seem simpler at first, but using an automated build process to create your package on your CI/CD system provides you with security benefits that pay off in the long run. Scorecard’s Packaging check helps guide you through this process.

### Packaging check verifies if a project is published as a package

The [Packaging check](https://github.com/ossf/scorecard/blob/main/docs/checks.md#packaging) assesses whether a project has been published as a package. The check is currently limited to repositories hosted on GitHub. It looks for [GitHub packaging workflows](https://docs.github.com/packages/learn-github-packages/publishing-a-package) and language-specific GitHub Actions that upload the package to a corresponding hub, like npm or PyPI.

Packaging your projects makes it easier for users to receive security patches as updates. It also provides information about the release details to your users, which opens the door to more collaboration from your open-source peers.

Another benefit to releasing projects as packages is reproducibility—the version that new users can download and execute is identical to the one that you and other contributors have already reviewed. Packages also have clear versioning documentation; this makes it easier to track whether any newly discovered security issues are applicable to your project.

## Customize your checks to your project’s needs

Based on the specifics of your project, not all the checks offered by Scorecard, or even not all of the checks discussed on this page, may apply to you. For example, if you are the sole maintainer of an open-source project, the “Code Review” check would not be usable for your project.

The languages you use also influence which checks will be useful to you. For example, if your project is built in C or C++, the Packaging and Dependency-Update-Tool checks will not be applicable because the C/C++ ecosystem does not have a centralized package manager.

To learn more about all the checks Scorecard offers, see the [checks documentation](https://github.com/ossf/scorecard/blob/main/docs/checks.md#check-documentation).
