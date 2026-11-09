# Conformance mode

Conformance mode evaluates a repository against the requirements of a
security catalog, such as the
[OSPS Baseline](https://baseline.openssf.org), instead of scoring checks.
Each requirement gets a result, and the output records the probe findings
behind it.

> **Status: experimental.** The options and the output may change or be
> removed in a future release. Every run prints a warning to standard error,
> and conformance JSON output sets `"experimental": true`.

Conformance results are heuristic signals to inform a human decision. They
are not a certification, and a run that evaluates only some requirements says
nothing about the rest.

## Usage

```sh
scorecard --repo=github.com/ossf/scorecard \
  --catalog=osps-baseline --profile=maturity-1 --format=json
```

- `--catalog`: the catalog to evaluate, as `<id>` or `<id>@<version>`.
  Without a version, Scorecard uses the newest release it was built with and
  records which.
- `--profile`: the catalog's applicability group to evaluate. Required.
- `--format`: `json` for conformance JSON, or `gemara` for a Gemara
  Evaluation Log. Required.
- `--output`: write the results to a file instead of standard output.

`--repo` and the package-manager options work as they do for a normal scan.
`--checks`, `--probes`, `--policy`, `--repos`, and `--org` can't be used with
`--catalog`. Neither can `--local` or a `--commit` other than `HEAD`: every
requirement Scorecard evaluates so far needs branch protection settings,
which Scorecard can read only for the current state of a hosted repository.

Conformance mode runs only the probes that the catalog's mappings need, and
computes no check scores. Runs without `--catalog` are unchanged.

## Catalogs

| Catalog | Releases | Profiles |
| --- | --- | --- |
| `osps-baseline` | `v2026.08.28` | `maturity-1`, `maturity-2`, `maturity-3` |

Scorecard evaluates these OSPS Baseline requirements from the default
branch's probe findings:

- OSPS-AC-03.01, direct commits to the primary branch are prevented:
  `branchesAreProtected`, `requiresPRsToChangeCode`, and
  `branchProtectionAppliesToAdmins`. An unprotected default branch is
  `FAIL`.
- OSPS-AC-03.02, deleting the primary branch requires confirmation:
  `blocksDeleteOnBranches`. A rule that blocks deletion is `PASS`. A rule
  that allows it is `UNKNOWN`, because Scorecard can't see whether the
  platform asks for confirmation.

Every other requirement in the profile is listed as not evaluated.

With classic branch protection, whether the rules apply to administrators
is visible only to a token with admin access to the repository; repository
rulesets are visible to any token (see the
[Branch-Protection check](checks.md#branch-protection)). Without that
access, OSPS-AC-03.01 can be `UNKNOWN`.

Classic branch protection can also let specific users, teams, or apps
bypass required pull requests. Scorecard doesn't collect that list yet, so
an OSPS-AC-03.01 `PASS` doesn't rule it out; the result's explanation says
so. A repository ruleset with any bypass actor gives `NEEDS_REVIEW`.

## Results

| Result | Meaning |
| --- | --- |
| `PASS` | The evidence shows the requirement is met. |
| `FAIL` | The evidence shows the requirement is not met. |
| `UNKNOWN` | Evidence is missing, or a probe reported an error or no data. |
| `NOT_APPLICABLE` | The requirement doesn't apply to the selected profile. |
| `NEEDS_REVIEW` | Evidence exists, but deciding needs judgment. |

If Scorecard can't collect the data at all, for example because the token
lacks a permission the platform requires, the run fails with an error and
writes no results.

For OSPS-AC-03.01, a default branch that requires pull requests but lets
administrators bypass the rule is `NEEDS_REVIEW`: the requirement doesn't
say whether an administrator exception is allowed.

## Output formats

### Conformance JSON

`--format=json` writes one object:

```json
{
  "date": "2026-11-08T12:00:00Z",
  "repo": { "name": "github.com/owner/repo", "commit": "..." },
  "scorecard": { "version": "v5.7.0", "commit": "..." },
  "catalog": { "id": "osps-baseline", "version": "v2026.08.28" },
  "profile": "maturity-1",
  "results": [
    {
      "requirement": "OSPS-AC-03.02",
      "result": "PASS",
      "explanation": "The default branch \"main\" cannot be deleted.",
      "evidence": [
        {
          "values": { "branchName": "main" },
          "probe": "blocksDeleteOnBranches",
          "message": "'allow deletion' disabled on branch 'main'",
          "outcome": "True"
        }
      ]
    }
  ],
  "notEvaluated": ["OSPS-AC-01.01", "..."],
  "experimental": true
}
```

Each `evidence` entry is a probe finding, in the same shape as
`--format=probe`.

### Gemara Evaluation Log

`--format=gemara` writes a [Gemara](https://github.com/gemaraproj/gemara)
Layer 5 Evaluation Log, as JSON, that validates against the
`#EvaluationLog` definition of Gemara v1.6.0. To check one:

```sh
git clone --depth 1 --branch v1.6.0 https://github.com/gemaraproj/gemara
cd gemara
cue vet -c -d '#EvaluationLog' . /path/to/log.json
```

Results map to Gemara values as `PASS` to `Passed`, `FAIL` to `Failed`,
`UNKNOWN` to `Unknown`, `NOT_APPLICABLE` to `Not Applicable`, and
`NEEDS_REVIEW` to `Needs Review`. The catalog release is recorded in
`metadata.mapping-references`. Requirements that were not evaluated are left
out of the log; its description says how many there were.

The log's top-level `result` rolls up only the requirements in the log, as
Gemara defines it. A `Passed` log means every evaluated requirement passed,
not that the repository meets the whole catalog. For which requirements
were evaluated, read the log's description, or `notEvaluated` in
conformance JSON.
