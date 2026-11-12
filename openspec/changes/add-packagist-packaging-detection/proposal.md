# Proposal: Detect Packagist publishing in the Packaging check

## Why

Current behavior: the Packaging check recognizes publishing steps for npm,
Maven, Gradle, sbt, RubyGems, NuGet, Docker, PyPI, Go, Cargo, ko, and Hex,
but not for Composer (PHP). Packagist pulls releases from the repository
instead of receiving uploads, so there is no `push` or `publish` command.
Projects that do not use the Packagist GitHub hook trigger the pull from CI
through the Packagist update API, which is a publishing step the check can
recognize the same way it recognizes `gem push`.

Partially addresses [#4378](https://github.com/ossf/scorecard/issues/4378).
Projects that rely only on the Packagist GitHub hook leave nothing to detect
in the repository and remain inconclusive; detecting them needs a registry
lookup, which this change does not add.

## What Changes

- Target state: a GitHub workflow step whose `run` calls
  `packagist.org/api/update-package` counts as a publishing step.
- Target state: GitLab CI configuration that calls
  `packagist.org/api/update-package` counts as packaging.
- **Score change**: repositories with such a workflow get 10 on Packaging
  instead of an inconclusive result, under the existing successful-run rule.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `openspec/specs/raw-data-collection`: `Packaging Workflow Detection` adds
  the Packagist signals.

## Impact

- `checks/fileparser/github_workflow.go`: one job matcher.
- `checks/raw/gitlab/packaging.go`: one packaging string.
- `docs/checks/internal/checks.yaml`, `docs/checks.md`: Packaging
  description.
- No change to probes, evaluation, flags, or output formats.
