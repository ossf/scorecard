# Design: OSPS Baseline conformance, first deliverable

## Context

The proposal in this change set the direction: Scorecard evaluates OSPS
Baseline requirements from its own probes, owns the mappings, and treats
Gemara as one output format. It left out specs, a design, and tasks.

This design narrows the first deliverable to something small enough to
review and ship as an experimental v5 feature. It also replaces the
implementation plan in [#4994](https://github.com/ossf/scorecard/pull/4994).
That plan grew too large to finish. The last section answers each review
comment on that PR.

## Goals

- Evaluate OSPS-AC-03.01 and OSPS-AC-03.02 (primary-branch protection)
  from existing probes, end to end: probe findings, then a requirement
  result, then a valid Gemara Evaluation Log.
- Leave native scores, checks, probes, and output formats unchanged.
- Report uncertainty honestly: never turn missing evidence into PASS or
  FAIL, and never present partial coverage as full conformance.
- Record enough identity (repository, commit, Scorecard version, catalog
  release, profile, time) to reproduce a result and compare it with
  another evaluator.

## Non-goals

- Full OSPS Baseline coverage, or new probes.
- Signing or attesting results. This waits for the open question on how
  Scorecard identifies itself as a signer (attestation identity, in
  `decisions.md`).
- Scorecard Action support, publishing to grc.store, or the weekly scan.
- A framework abstraction, capability registry, control catalog, or
  plugin model.
- OSCAL output, or a new in-toto predicate.

## Decisions

### Conformance is an execution mode

Setting `--catalog` switches the run into conformance mode. That mode
runs only the probes the catalog's mappings name and writes requirement
results; it computes no check scores. This mirrors Privateer, where the
run selects a catalog and an applicability group. Its output option only
chooses the serialization.

The proposal asked for one run to produce both check scores and
conformance results, after community feedback (recorded in
`decisions.md` as "Single run for all output"). This design defers that.
A review comment questioned whether the CLI needs it
([comment](https://github.com/ossf/scorecard/pull/4994#discussion_r3172125470)),
and `--probes` already returns before checks run. A combined run can
come back when a consumer such as the Action or api.scorecard.dev needs
it.

### Command-line options

```sh
scorecard --repo=github.com/ossf/scorecard \
  --catalog=osps-baseline --profile=maturity-1 --format=gemara
```

- `--catalog` takes a catalog ID with an optional `@<release>`. Without a
  release, Scorecard uses the newest one it was built with and records
  which.
- `--profile` takes one of the catalog's applicability group IDs. For
  OSPS these are `maturity-1` to `maturity-3`. It is required because
  OSPS applies requirements by maturity level and Gemara requires an
  applicability value on every result.
- `--format` must be `json` or `gemara` in this mode.

Values are checked against what is built into the binary, as `--format`
values are today, and the error lists the valid values.

Alternatives considered:

- A subcommand would duplicate `--repo`, `--local`, `--commit`,
  `--format`, and `--output`, which are registered only on the root
  command.
- `--level` assumes ordered levels. Gemara catalogs define their own
  applicability groups with no ordering, so it would fit OSPS only.
- `--scope` already means something else in this change (the scan scope
  of repo, repos, or org) and in Gemara's policy schema.
- `--output=baseline` is not available: `--output` names the results
  file.

IDs match Privateer's (`osps-baseline`, `maturity-1`), so results from
both evaluators line up without translation.

### Five results, plus a coverage list

| Result | Meaning |
| --- | --- |
| PASS | The evidence shows the requirement is met |
| FAIL | The evidence shows the requirement is not met |
| UNKNOWN | Evidence is missing, or a probe errored or couldn't observe it |
| NOT_APPLICABLE | The requirement doesn't apply to the selected profile |
| NEEDS_REVIEW | Evidence exists, but deciding needs judgment the evaluator doesn't make |

Requirements with no mapping are listed as not evaluated. That is
coverage, not a result.

NEEDS_REVIEW replaces the earlier name MANUAL_REVIEW, because the
judgment can come from a person or from another tool. Gemara has a
`Needs Review` value but no written definition of it. Privateer uses it
when its token can't read a setting; Scorecard reports that case as
UNKNOWN. The two evaluators can therefore label the same situation
differently, and the comparison work should expect that.

NOT_APPLICABLE comes only from the profile: a requirement whose
applicability groups don't include the selected profile. A probe that
reports `NotApplicable` (for example, "no branches found") does not make
a requirement not applicable; it gives UNKNOWN. This separates "could
not observe" from "does not apply", a distinction raised in review for
non-GitHub platforms
([comment](https://github.com/ossf/scorecard/pull/4994#discussion_r3036126361)).

### The two requirements

OSPS-AC-03.01 reads `requiresPRsToChangeCode` and
`branchProtectionAppliesToAdmins` for the default branch. If pull
requests are required but administrators can bypass the rule, the result
is NEEDS_REVIEW. The requirement says direct commits must be prevented.
It doesn't say whether an administrator exception is acceptable.

OSPS-AC-03.02 reads `blocksDeleteOnBranches`. Blocking deletion is
stricter than the requirement's "explicit confirmation of intent", so it
counts as a pass.

Both probes are stable and already feed the Branch-Protection check.
Findings for release branches are ignored, because both requirements
name the primary branch only.

### Go code in an internal package

The mappings and evaluation logic are Go, in `internal/conformance/`,
with one sub-package per OSPS release. A release package never changes
meaning after it ships; fixes go in the evaluator and are identified by
the Scorecard version.

Keeping the code internal means Allstar, the Action, and other importers
can't depend on it before it settles. The Gemara writer uses its own
structs, so `go.mod` gains no Gemara dependency. The code can move to
`pkg/` once it's stable and someone needs to import it.

There is no framework interface. Each requirement is a function over
probe findings, written in the style of `checks/evaluation`. A shared
abstraction can come when a second catalog shows what it needs to
cover.

### Gemara output

`--format=gemara` writes one Gemara Layer 5 Evaluation Log, as JSON,
valid against Gemara v1.6.0's `#EvaluationLog`. The catalog release goes
in `metadata.mapping-references`, the only field the schema has for it.
Requirements that weren't evaluated are left out of the log rather than
written as `Not Run`. A `Not Run` entry would need placeholder
descriptions and steps, while the conformance JSON already lists
unevaluated requirements.

### Experimental status

The new options are the only switch; no environment variable is needed.
While experimental, `--help` marks them, each run warns on standard
error, and the JSON output carries `"experimental": true`. A tracking
issue names the target release and the criteria to graduate. If the mode
hasn't graduated within 12 months of its first release, it is removed.

### Versioning

This ships as an additive v5 feature, not as part of a v6. A review
comment noted that a backwards-compatible change is still v5
([comment](https://github.com/ossf/scorecard/pull/4994#discussion_r3170295655)).

## Risks

- **Probe names become a dependency.** Each mapping names probes.
  Renaming a probe would break a mapping. Tests over the mappings catch
  this; maintainers still need to decide whether probe names are a
  versioned interface (see tasks).
- **The default branch isn't marked in findings.** Branch-protection
  findings carry a `branchName` value but don't say which branch is the
  default. The evaluator has to get the default branch name from the
  repository client.
- **Results will differ from Privateer's.** That is expected; the point
  is that the differences can be inspected.
- **No route to grc.store yet.** grc.store accepts Evaluation Logs only
  from registered hub plugins. Publishing Scorecard's own logs needs an
  agreement with the ORBIT WG before that milestone.

## Answers to the #4994 review

Comments are in the order they appear on the PR. "Superseded" means the
plan text the comment referred to is no longer part of this work.

| Comment | Answer |
| --- | --- |
| [Say how UNKNOWN works off GitHub](https://github.com/ossf/scorecard/pull/4994#discussion_r3036126361) | A probe's `NotApplicable`, `NotAvailable`, or `NotSupported` gives UNKNOWN, never NOT_APPLICABLE or FAIL. Specified in `conformance-evaluation`. |
| [The evaluation interface can't say a probe couldn't run](https://github.com/ossf/scorecard/pull/4994#discussion_r3036127731) | There is no interface; evaluation functions read finding outcomes, which already say this. |
| [How to tell "platform can't see it" from "project doesn't do it"](https://github.com/ossf/scorecard/pull/4994#discussion_r3036128363) | Same answer: finding outcomes already carry it (see the comment on unsupported features below). |
| [Release probes need per-platform versions](https://github.com/ossf/scorecard/pull/4994#discussion_r3036130882) | Release requirements are out of this deliverable. Noted for when they come in. |
| [What graduating Azure DevOps would look like](https://github.com/ossf/scorecard/pull/4994#discussion_r3036131267) | Superseded: this work doesn't change feature gates. |
| [Backwards compatible means v5](https://github.com/ossf/scorecard/pull/4994#discussion_r3170295655) | Agreed. Ships as additive v5; the v6 label is dropped. |
| [Does Level 1 include controls automation can't find?](https://github.com/ossf/scorecard/pull/4994#discussion_r3170300078) | Yes, some. The goal is no longer full Level 1 coverage; unmapped requirements are listed as not evaluated. |
| [Can't promote OSPO-style checks until the token problem is solved](https://github.com/ossf/scorecard/pull/4994#discussion_r3170312131) | Superseded: no check promotions here. |
| [Is this tied to a Level 1 control?](https://github.com/ossf/scorecard/pull/4994#discussion_r3170313826) | Superseded: no check promotions here. |
| [Raw format is unused; probes replaced it](https://github.com/ossf/scorecard/pull/4994#discussion_r3170316368) | Superseded: raw format is untouched. Conformance evidence uses probe findings. |
| [Full support (dropping the SARIF gate)](https://github.com/ossf/scorecard/pull/4994#discussion_r3170317657) | Out of scope here; can be its own change. |
| [Probes replaced raw results in check requests](https://github.com/ossf/scorecard/pull/4994#discussion_r3170325679) | Superseded: not part of this work. |
| [What does the feature-flag layer buy us?](https://github.com/ossf/scorecard/pull/4994#discussion_r3170335980) | Nothing yet, so it's dropped. The new options are the only switch. |
| [Probe groupings are more nuanced than the plan said](https://github.com/ossf/scorecard/pull/4994#discussion_r3170344329) | Superseded: this work doesn't change how probes are grouped or listed. |
| [How big is a capability?](https://github.com/ossf/scorecard/pull/4994#discussion_r3170361948) | No capability layer. Requirements map directly to probe findings. |
| [An `any` value type would hurt library users like Allstar](https://github.com/ossf/scorecard/pull/4994#discussion_r3170373488) | No public type. The code is internal, so library users see nothing. |
| [Security policy also falls back to file detection](https://github.com/ossf/scorecard/pull/4994#discussion_r3170380245) | Thanks; the standardization plan that sentence belonged to is dropped. |
| [The comment above isn't immediately relevant](https://github.com/ossf/scorecard/pull/4994#discussion_r3172102767) | Agreed. |
| [Findings already record unsupported features](https://github.com/ossf/scorecard/pull/4994#discussion_r3172113339) | Agreed; the design relies on these outcomes instead of a platform-capabilities input. |
| [The output format is the probe-to-output mapping](https://github.com/ossf/scorecard/pull/4994#discussion_r3172117173) | Close to this design: no framework interface, one function per requirement. The selection is a mode, though, because it also decides which probes run. |
| [Expected `--output=baseline` writing its own JSON, not one schema for everything](https://github.com/ossf/scorecard/pull/4994#discussion_r3172122820) | Agreed on separate JSON: conformance mode writes its own schema, and native JSON is unchanged. `--output` is the results-file flag, so the mode is `--catalog`. |
| [Single run is fine for the library, questionable for the CLI](https://github.com/ossf/scorecard/pull/4994#discussion_r3172125470) | Agreed; the single run is deferred. |
