# Supported Tools
* [CodeQL](https://docs.github.com/code-security/code-scanning/enabling-code-scanning/configuring-default-setup-for-code-scanning)
  * Detection is based on GitHub workflows using `github/codeql-action/analyze`, or GitHub Action checks run against PRs.
* [Qodana](https://github.com/JetBrains/qodana-action)
  * Detection based on GitHub workflows using `JetBrains/qodana-action`.
* [Snyk](https://github.com/snyk/actions)
  * Detection based on GitHub workflows using one of the actions from the set at https://github.com/snyk/actions
* [Sonar](https://docs.sonarsource.com/sonarqube/latest/setup-and-upgrade/overview/)
  * Detection based on the presence of a `pom.xml` file specifying a `sonar.host.url`, or GitHub Action checks run against PRs.

# Add Support

Don't see your SAST tool listed? 
Search for an existing issue, or create one, to discuss adding support.
