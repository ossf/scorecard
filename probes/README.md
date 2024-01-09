# Scorecard probes

This directory contains all the Scorecard probes.

A probe is an assessment of a focused, specific supply-chain security risk typically isolated to a particular ecosystem. For example, Scorecards fuzzing check consists of many different probes that assess particular ecosystems or aspects of fuzzing.

Each probe has its own directory in `scorecard/probes`. The probes follow a camelcase naming convention that describe the exact supply-chain security risk a particular probe assesses. 

Probes can return multiple or a single finding, where a finding is a piece of data with an outcome, message, and optionally a location. Probes should be designed in such a way that a `finding.OutcomePositive` reflects a positive result, and `finding.OutcomeNegative` reflects a negative result. Scorecard has other `finding.Outcome` types available for other results; For example, the `finding.OutcomeNotAvailable` is often used for scenarios, where Scorecard cannot assess a project with a given probe. In addition, probes should also be named in such a way that they answer "yes" or "no", and where "yes" answers positively to the supply-chain security risk, and "no" answers negatively. For example, probes that check for SAST tools in the CI are called `toolXXXInstalled` so that `finding.OutcomePositive` reflects that it is positive to use the given tool, and that "yes" reflects what Scorecard considers the positive outcome. For some probes, this can be a bit trickier to do; The `notArchived` probe checks whether a project is archived, however, Scorecard considers archived projects to be negative, and the probe cannot be called `isArchived`. These naming conventions are not hard rules but merely guidelines. Note that probes do not do any formal evaluation such a scoring; This is left to the evaluation part once the outcomes have been produced by the probes.

A probe consists of three files: 

- `def.yml`: The documentation of the probe. 
- `impl.go`: The actual implementation of the probe.
- `impl_test.go`: The probes test.

## Reusing code in probes

When multiple probes use the same code, the reused code can be placed on `scorecard/probes/internal/utils`

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
	finding.OutcomePositive)
f = f.WithRemediationMetadata(map[string]string{
	"dataToDisplay": "this is the text we will display",
})
```

To display the content to the user, Scorecard needs to print out the findings metadata somewhere in the evaluation part. This can be done with logging.