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

## Lifecycle

Probes can exist in several different lifecycle states:
* `Experimental`: The semantics of the probe may change, and there are no stability guarantees.
* `Stable`: The probe behavior and semantics will not change. There may be bug fixes as needed.
* `Deprecated`: The probe is no longer supported and callers should not expect it to be maintained.

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
```
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
```
The project is integrated into OSS-Fuzz.
```

### Should the changes be in the probe or the evaluation?
The remediation data must be set in the probe. 