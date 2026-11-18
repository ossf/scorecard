# Adding a check

A check scores one security practice from 0 to 10. This guide covers what a
check must meet, how its parts fit together, and the steps to add one.

## Requirements for a check

If you'd like to add a check, make sure it meets the following criteria and then
create a new GitHub Issue to discuss with the team:

- The scorecard must only be composed of automate-able, objective data. For
  example, a project having 10 contributors doesn’t necessarily mean it’s more
  secure than a project with 50 contributors. But, having two maintainers
  might be preferable to only having one - the larger bus factor and ability
  to provide code reviews is objectively better.
- The scorecard criteria can be as specific as possible and are not limited to
  general recommendations. For example, for Go, we can recommend/require
  specific linters and analyzers to be run on the codebase.
- The scorecard can be populated for any open source project without any work
  or interaction from maintainers.
- Maintainers must be provided with a mechanism to correct any automated
  scorecard findings they feel were made in error, provide "hints" for
  anything we can't detect automatically, and even dispute the applicability
  of a given scorecard finding for that repository.
- Any criteria in the scorecard must be actionable. It should be possible,
  with help, for any project to "check all the boxes".
- Any solution to compile a scorecard should be usable by the greater open
  source community to monitor upstream security.

Many ideas fit better as a probe added to an existing check. See
[probes/README.md](../probes/README.md).

## How a check works

A check runs in three stages:

1. A **collector** in [`checks/raw/`](raw/) reads repository files and the
   host's API, and returns the check's raw data.
2. **Probes** in [`probes/`](../probes/) each answer one yes-or-no question
   about the raw data and return findings.
3. An **evaluation** in [`checks/evaluation/`](evaluation/) turns the findings
   into a score and a reason.

A **definition** in
[`internal/checkdef/definitions.go`](../internal/checkdef/definitions.go)
wires the three together. It names the collector, the field of
`checker.RawResults` that stores the raw data, the probes, and the evaluation.
The same definition serves both ways of running Scorecard: running the check,
and running its probes by name with `--probes`.

Maintained is a short example to read alongside this guide:

- collector: [`checks/raw/maintained.go`](raw/maintained.go)
- probes: the `Maintained` group in [`probes/entries.go`](../probes/entries.go)
- evaluation: [`checks/evaluation/maintained.go`](evaluation/maintained.go)
- definition: `Maintained` in
  [`internal/checkdef/definitions.go`](../internal/checkdef/definitions.go)
- exported function: [`checks/maintained.go`](maintained.go)

## Steps

The steps use a check named `My-Check`. Replace the name with yours.

1. **Name the check.** In
   [`internal/checknames/checknames.go`](../internal/checknames/checknames.go),
   add a constant and append it to `AllValidChecks`:

   ```go
   MyCheck CheckName = "My-Check"
   ```

2. **Add the raw data type.** In
   [`checker/raw_result.go`](../checker/raw_result.go), define a type for what
   the collector finds, and add a field for it to `RawResults`:

   ```go
   // MyCheckData contains the raw results for the My-Check check.
   type MyCheckData struct {
       Files []File
   }
   ```

   ```go
   MyCheckResults MyCheckData
   ```

3. **Write the collector** in `checks/raw/my_check.go`. It takes the check
   request and returns the raw data:

   ```go
   func MyCheck(c *checker.CheckRequest) (checker.MyCheckData, error)
   ```

   Use `c.RepoClient` to list and read files and to call the host's API. A
   collector only gathers data; it doesn't decide what is good or bad. Test it
   in `checks/raw/my_check_test.go` with the mocks in
   [`clients/mockclients`](../clients/mockclients/).

4. **Add probes.** Each probe answers one question about `MyCheckResults`.
   Follow [probes/README.md](../probes/README.md), and register each probe
   with `checknames.MyCheck` as the check whose raw data it needs. The
   probes' own tests pass, but `make unit-test` fails until step 8 registers
   the check. Then list the probes as a group in
   [`probes/entries.go`](../probes/entries.go):

   ```go
   MyCheck = []ProbeImpl{
       myFirstProbe.Run,
       mySecondProbe.Run,
   }
   ```

5. **Write the evaluation** in `checks/evaluation/my_check.go`. It receives
   the findings from the check's probes and returns the result:

   ```go
   func MyCheck(name string, findings []finding.Finding, dl checker.DetailLogger) checker.CheckResult
   ```

   - First confirm that each expected probe returned a finding, with
     `finding.UniqueProbesEqual`, passing the `Probe` names of the probes in
     the group. If not, return an "invalid probe results"
     runtime error, as `evaluation.Maintained` does.
   - Log the findings that explain the score with
     `checker.LogFinding(dl, f, level)`:
     - `checker.DetailWarn` for findings that lower the score. These show
       with the `--show-details` option.
     - `checker.DetailInfo` for findings that raise the score. These also
       show with `--show-details`.
     - `checker.DetailDebug` for anything else. These show only with
       `--verbosity Debug`.
   - Return a result with a short sentence explaining the score:
     - `checker.CreateMaxScoreResult()` and `checker.CreateMinScoreResult()`
       for the highest and lowest scores.
     - `checker.CreateProportionalScoreResult()` for a score in proportion
       to a count.
     - `checker.CreateResultWithScore()` for any other score.
     - `checker.CreateInconclusiveResult()` when the check ran but can't
       determine a score.

   Test it in `checks/evaluation/my_check_test.go` with findings built by
   hand; the tests don't need a repository.

6. **Define the check.** In
   [`internal/checkdef/definitions.go`](../internal/checkdef/definitions.go),
   add a definition and append it to the `definitions` list:

   ```go
   // MyCheck is the My-Check check.
   var MyCheck = define(
       checknames.MyCheck,
       raw.MyCheck,
       func(r *checker.RawResults, data checker.MyCheckData) { r.MyCheckResults = data },
       probes.MyCheck,
       evaluation.MyCheck,
   ).supports(checker.CommitBased, checker.FileBased)
   ```

   The collector's return type and the store function's type must match, so
   the compiler catches a collector paired with the wrong field. If the
   collector takes a `clients.RepoClient` instead of the request, wrap it, as
   `CITests` does. See [Request types](#request-types) for `.supports` and
   [Experimental checks](#experimental-checks) for `.experimental`.

7. **Export the check.** Create `checks/my_check.go` with the name constant
   and a function that runs the definition:

   ```go
   // CheckMyCheck is the exported check name for My-Check.
   const CheckMyCheck = "My-Check"

   // MyCheck runs the My-Check check.
   func MyCheck(c *checker.CheckRequest) checker.CheckResult {
       return checkdef.MyCheck.Run(c)
   }
   ```

8. **Register the check.** Add `{checkdef.MyCheck, MyCheck}` to the
   `registered` table in [`checks/all_checks.go`](all_checks.go). Then add
   the check to `wantRegistry` in [`checks/registry_test.go`](registry_test.go),
   and `CheckMyCheck` from step 7 to `TestCheckNameConstants` in the same
   file, which keeps it equal to `checknames.MyCheck`.

9. **Document the check.**
   - Add an entry to
     [`docs/checks/internal/checks.yaml`](../docs/checks/internal/checks.yaml)
     with its risk, tags, supported repository types (`repos`), a short
     description, a full description, and remediation steps. Copy the format
     of an existing entry, such as `Maintained`. Risk is `Critical`, `High`,
     `Medium`, or `Low`.
   - Run `make generate-docs` to regenerate `docs/checks.md`. Commit both
     files. Don't edit `docs/checks.md` by hand.
   - Add a row to the checks table in the [README](../README.md#scorecard-checks).
   - Add the check's name to the `check` dropdown in
     [`.github/ISSUE_TEMPLATE/1-bug.yml`](../.github/ISSUE_TEMPLATE/1-bug.yml).

10. **Test and lint.** Run `make unit-test` and `make check-linter`. For a
    test against a real repository, add one in `e2e/my_check_test.go` using
    a dedicated repository that won't change over time. End-to-end tests need
    a GitHub token; CI runs them after merge and in the gated
    `integration-test` job on pull requests.

## What catches a missed step

| If you forget to | This fails |
| --- | --- |
| add the name to `AllValidChecks` | `TestDefinitionNames` (`internal/checkdef`), `TestRegistryNames` (`checks`) |
| add the definition, or add it to `definitions` | `TestDefinitionNames` |
| add the `registered` table entry | `TestRegistryNames`, `make validate-docs` |
| update `wantRegistry` | `TestRegistry` |
| name a registered check when registering a probe | `TestRequiredRawDataIsRegistered` (`internal/probes`) |
| list a probe in the evaluation's expected probes | the evaluation returns "invalid probe results"; the check's tests fail |
| add the `checks.yaml` entry | `make validate-docs` |
| regenerate `docs/checks.md` | the `generate-docs` CI job |

Nothing catches a missing README row, bug form entry, or `checks/my_check.go`
constant in `TestCheckNameConstants`; reviewers check those.

## Request types

A check declares the kinds of request it supports with `.supports(...)`:

- `checker.FileBased`: the check works on files alone, so it runs on a local
  directory (`--local`).
- `checker.CommitBased`: the check works at any commit, so it runs with
  `--commit` set to something other than the default branch's HEAD.

A check without `.supports(...)` runs only against a repository host at
HEAD, and is skipped for `--local` and other commits.

## Experimental checks

A new check can start as experimental. Add `.experimental("My-Check")` to its
definition. Scorecard then runs it only when the `SCORECARD_EXPERIMENTAL`
environment variable is set, to any value. An experimental check also needs:

- its name in `experimentalChecks` in
  [`checks/registry_test.go`](registry_test.go) and in
  [`cmd/internal/mcp/catalog/catalog.go`](../cmd/internal/mcp/catalog/catalog.go)
  (`TestExperimentalChecksMatch` fails otherwise);
- `EXPERIMENTAL` in the Note column of its README row.

## Errors

- **Collector:** return the error with context, for example
  `fmt.Errorf("listing workflows: %w", err)`. The definition reports any
  collector error as an internal error, and the check returns a runtime
  error result.
- **Probe:** return the error with context, as in the example in
  [probes/README.md](../probes/README.md); the check returns a runtime error
  result.
- **Evaluation:** return `checker.CreateRuntimeErrorResult()` with an error
  from [`errors`](../errors/), such as `sce.ErrScorecardInternal`.

See [errors/errors.md](../errors/errors.md) for the error values.
