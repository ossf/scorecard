# Scorecard Attestor

## What is scorecard-attestor?

scorecard-attestor is a tool that runs scorecard on a software source repo, and based on certain policies about those results, produces a Google Cloud binary authorization attestation.

scorecard-attestor helps users secure their software deployment systems by ensuring the code that they deploy passes certain criteria.

## Building and using scorecard-attestor

scorecard-attestor can be built as a standalone binary from source using `make build-attestor`, or with Docker, using `make build-attestor-docker`. scorecard-attestor is intended to be used as part of a Google Cloud Build pipeline, and inherits environment variables based on [build substitutions](https://cloud.google.com/build/docs/configuring-builds/substitute-variable-values).

Unless there's an internal error, scorecard-attestor will always return a successful status code, but will only produce a binary authorization attestation if the policy check passes.

## Configuring policies for scorecard-attestor

Policies for scorecard attestor can be passed through the CLI using the `--policy` flag. Examples of policies can be seen in [attestor/policy/testdata](/attestor/policy/testdata).

### Policies

* `PreventBinaryArtifacts`: Ensure that a repository is free from binary artifacts, which can link against the final repo artifact but isn't reviewable.
  * `AllowedBinaryArtifacts`: A list of binary artifacts, by repo path, to ignore. If not specified, no binary artifacts will be allowed
* `PreventKnownVulnerabilities`: Ensure that the project is free from security vulnerabilities/advisories, as registered in osv.dev.
* `PreventUnpinnedDependencies`: Ensure that a project's dependencies are pinned by hash. Dependency pinning makes builds more predictable, and prevents the consumption of malicious package versions from a compromised upstream.
  * `AllowedUnpinnedDependencies`: Ignore some dependencies, either by the filepath of the dependency management file (`filepath`, e.g. requirements.txt or package.json) or the dependency name (`packagename`, the specific package being ignored). If multiple filepaths/names, or a combination of filepaths and names are specified, all of them will be used. If not specified, no unpinned dependencies will be allowed.
* `RequireCodeReviewed`: Require that If `CodeReviewRequirements` is not specified, at least one reviewer will be required on all changesets. Scorecard-attestor inherits scorecard's default commit window (i.e. will only look at the last 30 commits to determine if they are reviewed or not).
  * `CodeReviewRequirements.MinReviewers`: The minimum number of distinct approvals required.
  * `CodeReviewRequirements.RequiredApprovers`: A set of approvers, any of whom must be found to have approved all changes. If a change is found without any approvals from this list, the check fails.

### Policy schema

Policies follow the following schema:

```yaml
---
type: "//rec"
optional:
    preventBinaryArtifacts: "//bool"
    allowedBinaryArtifacts:
        type: "//arr"
        contents: "//str" # Accepts glob-based filepaths as strings here
    ensureNoVulnerabilities: "//bool"
    ensureDependenciesPinned: "//bool"
    allowedUnpinnedDependencies:
        type: "//arr"
        contents:
            type: "//rec"
            optional:
                packagename: "//str"
                filepath: "//str"
                version: "//str"
    ensureCodeReviewed: "//bool"
    codeReviewRequirements:
        type: "//rec"
        optional:
            requiredApprovers:
                type: "//arr"
                contents: "//str"
            minReviewers: "//int"
```

## Sample

Examples of how to use scorecard-attestor with binary authorization in your project can be found in these two repos:

* [scorecard-binauthz-test-good](https://github.com/ossf-tests/scorecard-binauthz-test-good)
* [scorecard-binauthz-test-bad](https://github.com/ossf-tests/scorecard-binauthz-test-bad)

Sample code comes with:

* `cloudbuild.yaml` to build the application and run scorecard-attestor
* Terraform files to set up the binary authorization environment, including KMS and IAM.
