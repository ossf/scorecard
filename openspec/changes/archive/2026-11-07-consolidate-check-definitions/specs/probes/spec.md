## MODIFIED Requirements

### Requirement: Probe Registration

A probe SHALL be directly selectable by name only when it is registered. A
probe registers itself by name when its package loads, either with an
implementation and the checks whose raw data it needs, or as an independent
probe with an implementation that takes the scan request. Registration SHALL
fail if the name is empty, no implementation is given, or a raw-data probe
lists no required checks. Looking up a name that is not registered SHALL fail
with a "probe not found" error. Exactly three probes,
`hasNoGitHubWorkflowPermissionUnknown`, `jobLevelPermissions`, and
`topLevelPermissions`, SHALL NOT be registered; they run only inside the
Token-Permissions check.

Every check a registered probe lists as required SHALL be a registered
check (see `check-registry`, Registered Checks). Naming any registered probe
therefore never fails for lack of a raw-data collector.

#### Scenario: Probe declares its raw data

- **WHEN** the `hasPermissiveLicense` probe is registered
- **THEN** it declares that it needs the License raw data

#### Scenario: Every registered probe has a collector

- **WHEN** each registered raw-data probe is named on its own
- **THEN** every check it lists as required is a registered check
- **AND** no run fails with `unknown check`

#### Scenario: Unknown probe name

- **WHEN** a lookup is made for probe `noSuchProbe`
- **THEN** the system fails with `probe "noSuchProbe" not found`

#### Scenario: Unregistered token-permission probe

- **WHEN** a user runs `scorecard --repo=github.com/owner/repo
  --probes=topLevelPermissions`
- **THEN** the run fails with `probe "topLevelPermissions" not found`
