## ADDED Requirements

### Requirement: Same Raw Data for Checks and Probes

For every check except Packaging, naming probes (see `probes`, Running
Selected Probes) SHALL collect the same raw data as running the check. It
SHALL make the same repository client calls and fail with the same error
text. Experimental gating applies only when checks run (see
`check-registry`, Experimental Checks). Probe mode therefore collects SBOM
and Webhooks raw data without `SCORECARD_EXPERIMENTAL`.

#### Scenario: Maintained in both modes

- **GIVEN** a GitHub repository
- **WHEN** the Maintained check runs, and separately a user names the
  `archived` probe
- **THEN** both collect the same Maintained raw data with the same
  repository client calls

#### Scenario: Collection error in probe mode

- **GIVEN** listing commits fails with an error
- **WHEN** a user names a probe that needs Maintained raw data
- **THEN** the run fails with the error text the Maintained check reports
  for the same failure

### Requirement: Packaging Raw Data by Mode

When the Packaging check runs, GitHub, GitLab, and local directories SHALL
follow Packaging Workflow Detection, and any other client SHALL yield no
packaging data and no error. When probes are named, Packaging raw data SHALL
be available only for GitHub and GitLab.

#### Scenario: Packaging on a local directory

- **GIVEN** a local directory with a GitHub workflow that publishes a
  package
- **WHEN** the Packaging check runs against it
- **THEN** the workflow is recorded as packaging
- **AND** naming `packagedWithAutomatedWorkflow` against the same directory
  fails because only GitHub and GitLab are supported
