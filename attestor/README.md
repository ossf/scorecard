# Scorecard Attestor

## What is scorecard-attestor?

scorecard-attestor is a tool that runs scorecard on a software source repo, and based on certain policies about those results, produces a Google Cloud binary authorization attestation.

scorecard-attestor helps users secure their software deployment systems by ensuring the code that they deploy passes certain criteria.

## Building and using scorecard-attestor

scorecard-attestor can be built as a standalone binary from source using `make build-attestor`, or with Docker, using `make build-attestor-docker`. scorecard-attestor is intended to be used as part of a Google Cloud Build pipeline, and inherits environment variables based on [build substitutions](https://cloud.google.com/build/docs/configuring-builds/substitute-variable-values).

Unless there's an internal error, scorecard-attestor will always return a successful status code, but will only produce a binary authorization attestation if the policy check passes.

## Configuring policies for scorecard-attestor

Policies for scorecard attestor can be passed through the CLI using the `--policy` flag. Examples of policies can be seen in [attestor/policy/testdata](/attestor/policy/testdata).

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

### Missing parameters

Policies that are left blank will be ignored. Policies that allow users additional configuration options will be given default parameters as listed below.

* `PreventBinaryArtifacts`: If not specified, `AllowedBinaryArtifacts` will be empty, i.e. no binary artifacts will be allowed
* `PreventUnpinnedDependencies`: If not specified, `AllowedUnpinnedDependencies` will be empty, i.e. no unpinned dependencies will be allowed
* `RequireCodeReviewed`: If not specified, `CodeReviewRequirements` will require at least one reviewer on all changesets.

## Sample

Examples of how to use scorecard-attestor with binary authorization in your project can be found in these two repos:

* [scorecard-binauthz-test-good](https://github.com/ossf-tests/scorecard-binauthz-test-good)
* [scorecard-binauthz-test-bad](https://github.com/ossf-tests/scorecard-binauthz-test-bad)

Sample code comes with:

* `cloudbuild.yaml` to build the application and run scorecard-attestor
* Terraform files to set up the binary authorization environment, including KMS and IAM.
