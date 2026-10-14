# remediation Specification

## Purpose

Scorecard attaches remediation guidance to results so maintainers know how
to fix what a check or probe flagged. Guidance comes from three sources:
the remediation section of each probe's `def.yml`, finding-specific
suggestions generated for unpinned dependencies, and the documented
remediation list of each check. This spec defines how that guidance is
attached, which placeholders are substituted, and what happens when
guidance cannot be produced.

## Requirements

### Requirement: Probe Remediation Definition

Each probe's `def.yml` SHALL define a `remediation` section with
`onOutcome`, `effort`, `text`, and `markdown`. `effort` MUST be one of
`Low`, `Medium`, or `High`, and `onOutcome` MUST be a valid finding outcome
(`True`, `False`, `NotApplicable`, `NotAvailable`, `NotSupported`, or
`Error`). A probe definition that violates either rule MUST be rejected as
invalid. The `text` and `markdown` lists MUST each be joined with newlines
to form the remediation text and markdown.

#### Scenario: Invalid effort

- **GIVEN** a probe `def.yml` with `effort: Huge`
- **WHEN** the probe definition is loaded
- **THEN** loading fails with an "invalid" error

#### Scenario: Multi-line remediation

- **GIVEN** a probe whose `remediation.text` has three list items
- **WHEN** a finding is created for that probe
- **THEN** the finding's remediation text is the three items separated by
  newlines

### Requirement: Remediation Only on the Bad Outcome

The system SHALL attach a probe's remediation to a finding only when the
finding's outcome equals the probe's `onOutcome`. A finding with any other
outcome MUST carry no remediation, and the remediation MUST NOT be
restored if the outcome is later changed.

#### Scenario: Passing finding

- **GIVEN** a probe with `onOutcome: False`
- **WHEN** the probe produces a finding with outcome `True`
- **THEN** the finding has no `remediation`

#### Scenario: Failing finding

- **GIVEN** a probe with `onOutcome: False`
- **WHEN** the probe produces a finding with outcome `False`
- **THEN** the finding's `remediation` contains the probe's `text`,
  `markdown`, and `effort`

### Requirement: Remediation Effort Encoding

The system SHALL serialize remediation effort in JSON findings as an
integer: 1 for `Low`, 2 for `Medium`, 3 for `High`, and 0 when no effort is
set.

#### Scenario: Medium effort

- **GIVEN** a probe with `effort: Medium`
- **WHEN** a failing finding is written with `--format probe`
- **THEN** its `remediation.effort` is `2`

### Requirement: Location Placeholder Substitution

When a finding with remediation is given a location, the system SHALL
replace every `${{ finding.location.path }}` in the remediation text and
markdown with the location's path.

#### Scenario: Path substituted

- **GIVEN** a remediation text "Edit ${{ finding.location.path }}"
- **WHEN** the finding is located at `.github/workflows/ci.yml`
- **THEN** the remediation text is "Edit .github/workflows/ci.yml"

### Requirement: Metadata Placeholder Substitution

The system SHALL replace `${{ metadata.<key> }}` in remediation text and
markdown with the value supplied for `<key>` by the probe. The repository
metadata available to probes MUST include `repository.host`,
`repository.name`, `repository.uri`, `repository.sha1`,
`repository.defaultBranch` (or `unknown` when the platform cannot report
it), and `localPath`. Probes additionally supply `workflow` (the workflow
path with `.github/workflows/` removed) for token-permission findings,
`osvid` for vulnerability findings, and `patch` for script-injection
findings. A placeholder whose key is not supplied MUST be left unchanged.

#### Scenario: Token permissions link

- **GIVEN** a top-level permissions finding in
  `.github/workflows/ci.yml` of `github.com/org/repo` on branch `main`
- **WHEN** the finding's remediation is built
- **THEN** the text links to
  `https://app.stepsecurity.io/secureworkflow/github.com/org/repo/ci.yml/main?enable=permissions`

#### Scenario: Vulnerability identifier

- **GIVEN** a vulnerability finding for `GHSA-xxxx-yyyy-zzzz`
- **WHEN** the finding's remediation is built
- **THEN** the text refers to `https://osv.dev/GHSA-xxxx-yyyy-zzzz`

#### Scenario: Missing metadata

- **GIVEN** a script-injection finding for which no patch could be
  generated
- **WHEN** the finding's remediation is built
- **THEN** the markdown still contains the literal `${{ metadata.patch }}`

### Requirement: Script Injection Patch

For a dangerous-workflow script-injection finding, the system SHALL try to
generate a patch for the affected workflow file. When generation succeeds
the patch MUST be set as the remediation's `patch` field and substituted
for `${{ metadata.patch }}`. When the workflow file cannot be read or
parsed, the finding MUST be emitted without a patch.

#### Scenario: Patch generated

- **GIVEN** a workflow that interpolates an untrusted input in a `run`
  step
- **WHEN** the script-injection probe runs on a local checkout
- **THEN** the finding's `remediation.patch` contains the proposed change

### Requirement: GitHub Actions Pinning Suggestion

For each unpinned GitHub Action dependency, the system SHALL attach the
remediation text `update your workflow using
https://app.stepsecurity.io/secureworkflow/<owner>/<repo>/<workflow>/<branch>?enable=pin`
and the equivalent markdown link, where `<workflow>` is the workflow path
with `.github/workflows/` removed and `<branch>` is the default branch.
When the default branch or the `<owner>/<repo>` cannot be determined (the
repository URI does not have exactly three `/`-separated parts), the
dependency MUST receive no suggestion.

#### Scenario: Unpinned action

- **GIVEN** `github.com/org/repo` with default branch `main`
- **AND** `.github/workflows/ci.yml` uses `actions/checkout@v4`
- **WHEN** Pinned-Dependencies runs
- **THEN** the remediation is "update your workflow using
  `https://app.stepsecurity.io/secureworkflow/org/repo/ci.yml/main?enable=pin`"

#### Scenario: Repository path cannot be split

- **GIVEN** a repository URI with a nested group path
- **WHEN** an unpinned action is found
- **THEN** no workflow pinning suggestion is attached

### Requirement: Docker Image Digest Suggestion

For each unpinned Dockerfile container image, the system SHALL look up the
image's current digest in its container registry and attach the
remediation text and markdown
`pin your Docker image by updating <image> to <image>@<digest>`, where
`<image>` is `<name>:<tag>` when a tag is present and `<name>` otherwise.
When the image has no name or the digest lookup fails, the dependency MUST
receive no suggestion.

#### Scenario: Tagged image

- **GIVEN** a Dockerfile with `FROM golang:1.22`
- **WHEN** the digest lookup returns `sha256:abc`
- **THEN** the remediation is "pin your Docker image by updating
  golang:1.22 to golang:1.22@sha256:abc"

#### Scenario: Lookup failure

- **GIVEN** an image that does not exist in any reachable registry
- **WHEN** Pinned-Dependencies runs
- **THEN** no Docker pinning suggestion is attached

### Requirement: NuGet Pinning Suggestion

For each unpinned NuGet install or restore command, the system SHALL
attach remediation text recommending central package management or a
lockfile, with links to the Microsoft documentation. When some project or
central-package-management files lock or fix versions and others do not,
the system MUST append the list of offending files or versions to that
text. When all NuGet dependencies are locked or fixed, they MUST be
treated as pinned and carry no remediation.

#### Scenario: Partially locked projects

- **GIVEN** two csproj files, only one of which sets
  `RestoreLockedMode` to true
- **WHEN** an unpinned `nuget restore` is found
- **THEN** the remediation text ends with ": some of your csproj files set
  the RestoreLockedMode property to true, while other do not set it:"
  followed by the unlocked file paths

### Requirement: Dependency Suggestions Replace Probe Text

The system SHALL use a dependency-specific suggestion, when one exists, as
the remediation of the corresponding unpinned-dependency finding. An
unpinned-dependency finding without a dependency-specific suggestion MUST
carry no remediation.

#### Scenario: No suggestion available

- **GIVEN** an unpinned `pip install` command in a shell script
- **WHEN** the pinsDependencies probe runs
- **THEN** the finding has outcome `False` and no `remediation`

### Requirement: Remediation in Rendered Output

The system SHALL surface remediation as follows: the `probe` format MUST
include each finding's `remediation` object; for Pinned-Dependencies
details in the `default`, `json`, and `intoto` formats, the remediation
text MUST be appended to the detail line; in SARIF, a location's message
MUST include `Remediation tip: <remediation markdown>`. Details produced
directly from other probe findings MUST NOT include remediation text in
the `default`, `json`, or `intoto` formats.

#### Scenario: Pinned dependency detail

- **GIVEN** an unpinned action with a pinning suggestion
- **WHEN** results are rendered with `--show-details`
- **THEN** the Warn detail ends with ": update your workflow using"
  followed by the StepSecurity `secureworkflow` URL

### Requirement: Check-Level Remediation

The system SHALL take each check's remediation bullets from the check
documentation and render them in SARIF output as the
`**Remediation (click "Show more" below)**` section of the rule's
`help.markdown`, as one Markdown list item per bullet. The `default` format MUST
instead point to the check's documentation URL in its
`Documentation/Remediation` column.

#### Scenario: SARIF rule help

- **GIVEN** a check documented with remediation items "a" and "b"
- **WHEN** results are rendered as SARIF
- **THEN** the rule's `help.markdown` begins with the remediation heading
  followed by bullets for "a" and "b"
