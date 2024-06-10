# Maintainer Annotations

*Maintainer Annotations* are an *experimental feature* that let maintainers *add context* to Scorecard check results in `scorecard.yml`. Annotations can be useful when Scorecard incorrectly assesses a project's security practices.

## Annotating Your Project

To annotate your repository, create a `scorecard.yml` file in the root of your repository.
> You can also place your annotations in `.scorecard.yml` or `.github/scorecard.yml`.

The file `scorecard.yml` looks like so:

```yml
exemptions:
  - checks:
      - binary-artifacts
    annotations:
      - annotation: test-data # the binary files are only used for testing
  - checks:
      - dangerous-workflow
    annotations:
      - annotation: remediated # the workflow is dangerous but only run under maintainers verification and approval
      -
```

You can annotate multiple checks at a time:

```yml
exemptions:
  - checks:
      - binary-artifacts
      - pinned-dependencies
    annotations:
      - annotation: test-data # the binary files and files with unpinned dependencies are only used for testing
 
```

And also provide multiple annotations for checks:

```yml
exemptions:
  - checks:
      - binary-artifacts
    annotations:
      - annotation: test-data # test.exe is only used for testing 
      - annotation: remediated # dependency.exe is needed and it's used but the binary signature is verified
  
```

The available checks are the Scorecard checks in lower case e.g. Binary-Artifacts is `binary-artifacts`.

## Types of Annotations

The annotations are predefined as shown in the table below:

| Annotation | Description | Example |
|------------|-------------|---------|
| test-data | A check or probe has found an issue in files or code snippets only used for test or example purposes. | The binary files are only used for testing. |
| remediated | To annotate when a check or probe has found a security issue to which a remediation was already applied. | A workflow is dangerous but only run under maintainers verification and approval, or a binary is needed but it is signed or has provenance. |
| not-applicable | To annotate when a check or probe is not applicable for the case. | The dependencies should not be pinned because the project is a library. |
| not-supported | To annotate when the maintainer fulfills a check or probe in a way that is not supported by Scorecard. | Clang-Tidy is used as SAST tool but not identified because its not supported. |
| not-detected | To annotate when the maintainer fulfills a check or probe in a way that is supported by Scorecard but not identified. | Dependabot is configured in the repository settings and not in a file. |

## Viewing Maintainer Annotations

To see the maintainers annotations for each check on Scorecard results, use the `--show-annotations` option.
