# OpenSSF Scorecard Probes

This directory contains all the Scorecard probes. Each probe has its own subdirectory.

A probe is an individual heuristic, which provides information about a distinct behavior a project under analysis may or may not be doing.
The probes follow a camelcase naming convention that describe the exact heuristic a particular probe assesses.
The name should be phrased in a way that can be answered by "true" or "false".

Probes can return one or more findings, where a finding is a piece of data with an outcome, message, and optionally a location where the behavior was observed.
The primary outcomes are `finding.OutcomeTrue` and `finding.OutcomeFalse`, but other outcomes are available as well.
For example, the `finding.OutcomeNotAvailable` is often used for scenarios, where Scorecard cannot assess a behavior because there is no data to analyze.

A probe consists of three files:

- `def.yml`: The documentation of the probe.
- `impl.go`: The actual implementation of the probe.
- `impl_test.go`: The probe's test.

Most probes read raw data that a check collects, such as the commits the
Maintained check lists. A probe can also collect its own data; these are
called independent probes. To add a check as well, see
[checks/write.md](../checks/write.md).

## Adding a probe

The steps use a probe named `myProbe` that reads the Maintained check's raw
data. Replace both with yours.

1. **Create the files.** From the repository root, run:

   ```sh
   make setup-probe probeName=myProbe
   ```

   This creates `probes/myProbe/` with `def.yml`, `impl.go`, and
   `impl_test.go`.

2. **Fill in `def.yml`.** Scorecard checks `id`, `lifecycle`, `remediation`,
   and `ecosystem` when the probe creates its first finding, so a mistake
   there shows up in the probe's own tests:

   - `id`: the probe's name, `myProbe`.
   - `lifecycle`: `experimental` for a new probe. See [Lifecycle](#lifecycle).
   - `short`, `motivation`, `implementation`: what the probe checks, why it
     matters, and how it decides. These appear in
     [docs/probes.md](../docs/probes.md).
   - `outcome`: a list with one entry per outcome the probe can return.
   - `remediation`:
     - `onOutcome`: the outcome that needs fixing, `True` or `False`.
     - `effort`: `Low`, `Medium`, or `High`.
     - `text`, and optionally `markdown`: the steps to fix it.
   - `ecosystem`:
     - `languages`: the languages the probe applies to, from
       [`clients/languages.go`](../clients/languages.go), or `all`.
     - `clients`: where the probe works: `github`, `gitlab`, `localdir`.

   [`hasRecentCommits/def.yml`](hasRecentCommits/def.yml) is a complete
   example.

3. **Write `Run` in `impl.go`.** It reads the raw data and returns findings,
   the probe's name, and an error. The scaffold already declares `fs` (the
   embedded `def.yml`) and the `Probe` name constant; `uerror` is
   `github.com/ossf/scorecard/v5/probes/internal/utils/uerror`:

   ```go
   func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
       if raw == nil {
           return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
       }
       outcome := finding.OutcomeFalse
       if len(raw.MaintainedResults.DefaultBranchCommits) > 0 {
           outcome = finding.OutcomeTrue
       }
       f, err := finding.NewWith(fs, Probe, "explain the outcome here", nil, outcome)
       if err != nil {
           return nil, Probe, fmt.Errorf("create finding: %w", err)
       }
       return []finding.Finding{*f}, Probe, nil
   }
   ```

   Pass a `*finding.Location` instead of `nil` when the finding points to a
   file. Attach data for the check's evaluation with `f.WithValues(...)`.
   [`hasRecentCommits/impl.go`](hasRecentCommits/impl.go) is a complete
   example.

4. **Register the probe.** Uncomment the `init()` in `impl.go` that fits, and
   add the imports its comment names. A probe that reads a check's raw data
   names that check:

   ```go
   func init() {
       probes.MustRegister(Probe, Run, []checknames.CheckName{checknames.Maintained})
   }
   ```

   An independent probe has
   `func Run(c *checker.CheckRequest) ([]finding.Finding, string, error)`,
   reads its data through `c.RepoClient` as a collector does, and registers
   with `probes.MustRegisterIndependent(Probe, Run)`. Registered
   probes can be run by name with `scorecard --probes=myProbe`.

5. **List the probe in [`entries.go`](entries.go).** Import the package, then
   add `myProbe.Run` to one list:

   - a check's group, such as `Maintained`, if the check should use it to
     score. Also add `myProbe.Probe` to the list of probe names that the
     check's evaluation in [`checks/evaluation/`](../checks/evaluation/)
     passes to `finding.UniqueProbesEqual`, and score its findings there.
   - `Uncategorized`, if no check uses it yet.
   - `Independent`, for an independent probe.

   Without this step the probe's package is never loaded, and
   `--probes=myProbe` fails with `probe "myProbe" not found`. No test
   catches it.

6. **Test the probe** in `impl_test.go`. Build `checker.RawResults` by hand
   for each case and compare the outcomes. `test.AssertOutcomes` in
   [`internal/utils/test`](internal/utils/test/) does the comparison;
   [`hasRecentCommits/impl_test.go`](hasRecentCommits/impl_test.go) is an
   example.

7. **Regenerate the docs and check your work.** Run `make generate-docs` and
   commit the updated `docs/probes.md`; CI fails if it is out of date. Then
   run `make unit-test` and `make check-linter`.

## Lifecycle

Probes can exist in several different lifecycle states:

- `experimental`: The semantics of the probe may change, and there are no stability guarantees.
- `stable`: The probe behavior and semantics will not change. There may be bug fixes as needed.
- `deprecated`: The probe is no longer supported and callers should not expect it to be maintained.

## Reusing code in probes

When multiple probes use the same code, the reused code can be placed in a package under `probes/internal/`

## How do I know which probes to add?

In general, browsing through the Scorecard GitHub issues is the best way to find new probes to add. Requests for support for new tools, fuzzing engines or other heuristics can often be converted into specific probes.

## Probe definition formatting

Probe definitions can display links following standard markdown format.

Probe definitions can display dynamic content. This requires modifications in `def.yml` and `impl.go` and in the evaluation steps.

The following snippet in `def.yml` will display dynamic data provided by `impl.go`:

```md
${{ metadata.dataToDisplay }}
```

And then in `impl.go` add the following metadata:

```golang
f, err := finding.NewWith(fs, Probe,
	"Message", nil,
	finding.OutcomeTrue)
f = f.WithRemediationMetadata(map[string]string{
	"dataToDisplay": "this is the text we will display",
})
```

### Example

Consider a probe with following line in its `def.yml`:

```text
The project ${{ metadata.oss-fuzz-integration-status }} integrated into OSS-Fuzz.
```

and the probe sets the following metadata:

```golang
f, err := finding.NewWith(fs, Probe,
	"Message", nil,
	finding.OutcomeTrue)
f = f.WithRemediationMetadata(map[string]string{
	"oss-fuzz-integration-status": "is",
})
```

The probe will then output the following text:

```text
The project is integrated into OSS-Fuzz.
```

### Should the changes be in the probe or the evaluation?

The remediation data must be set in the probe.
