# OpenSpec

This directory holds [OpenSpec](https://openspec.dev/) specifications for
Scorecard.

- `specs/` describes Scorecard v5 behavior as built, including known
  defects. A requirement that records a defect is not a compatibility
  promise; fixing it is an ordinary spec delta.
- `changes/` holds proposed changes. `changes/osps-baseline-conformance/`
  describes a v6 target in which one run produces both check scores and
  conformance results. In v5, running probes directly returns probe
  findings and skips checks.

## Capability ownership

Each behavior is stated in one capability. Others refer to it by path
rather than restating it.

| Capability | Owns |
| --- | --- |
| `aggregate-scoring` | Aggregate formula, weights, inconclusive exclusion |
| `attestor` | Attestor commands, attestor policy, signing |
| `check-evaluation` | Per-check score rules and shared evaluation behavior |
| `check-registry` | Check inventory, metadata, selection, experimental gating |
| `cli` | Command tree, flags, progress output, result ordering, exit status |
| `http-server` | `scorecard serve` requests, statuses, CORS, lifecycle |
| `maintainer-annotations` | Annotation discovery, parsing, applicability |
| `package-manager-resolution` | npm, PyPI, RubyGems, and NuGet lookup |
| `policy` | Policy file parsing, policy-driven selection, policy errors |
| `probes` | Probe registration, direct probe runs, findings and outcomes |
| `public-go-api` | Exported `Run`, its options, defaults, and errors |
| `raw-data-collection` | Evidence gathered for checks and probes |
| `remediation` | Remediation definitions and generated suggestions |
| `repo-clients` | Remote and local repository access and host support |
| `result-formats` | Output formats, schemas, destinations, score presentation |

A flag is defined in `cli`; what it turns on is defined by the capability
that owns that behavior. A computed value is defined where it is computed;
how it is displayed is defined in `result-formats`.

## Validation

```sh
openspec validate --specs --strict
```
