## MODIFIED Requirements

### Requirement: Packaging Workflow Detection

For GitHub, the system SHALL check each workflow for a job that publishes a
package, recognized by: `actions/setup-node` with the npm registry URL plus
an `npm publish` step; `actions/setup-java` plus `mvn deploy` or
`gradle publish`; `sbt ci-release`; `gem push`; `nuget push`;
`docker push` or `docker/build-push-action`; `pypa/gh-action-pypi-publish`;
`relekang/python-semantic-release`; `goreleaser/goreleaser-action`;
`cargo publish`; `imjasonh/setup-ko` or `ko-build/setup-ko`;
`npx semantic-release`; `erlef/setup-elixir` plus `hex.publish`; or a
`run` step calling `packagist.org/api/update-package`. A matching workflow
SHALL count as packaging only when it has a successful run, unless the
platform cannot list runs.
For GitLab, the system SHALL search the flattened CI configuration for
`docker push`, `nuget push`, `poetry publish`, `twine upload`, or
`packagist.org/api/update-package`. For local directories, the system
SHALL apply both methods.

#### Scenario: Published with a successful run

- **GIVEN** a workflow using `pypa/gh-action-pypi-publish`
- **AND** the workflow has a successful run
- **WHEN** the Packaging raw data is collected
- **THEN** a package is recorded with the workflow file and run URLs

#### Scenario: Publishing workflow never ran

- **GIVEN** a publishing workflow with no successful runs
- **WHEN** the Packaging raw data is collected
- **THEN** the workflow is recorded with a message that it was not used in
  runs

#### Scenario: Packagist update from a workflow

- **GIVEN** a workflow step that runs `curl -X POST` against
  `https://packagist.org/api/update-package`
- **AND** the workflow has a successful run
- **WHEN** the Packaging raw data is collected
- **THEN** a package is recorded with the workflow file and run URLs

#### Scenario: Packagist mentioned without publishing

- **GIVEN** a workflow that only installs dependencies from Packagist or
  reads a package page
- **WHEN** the Packaging raw data is collected
- **THEN** no package is recorded for that workflow
