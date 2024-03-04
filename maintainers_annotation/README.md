# Maintainers Annotation

Maintainers Annotation is an experimental feature to let maintainers add annotations to Scorecard checks.

## Showing Maintainers Annotation

To see maintainers annotations for each check on Scorecard results, use the `--show-ma` option.

## Adding Annotations

To add annotations to your repository, create a `scorecard.yml` file in the root of your repository.

The file structure is as follows:
```yml
exemptions:
  - checks:
      - binary-artifacts
    annotations:
      # the binary files are only used for testing
      - annotation: test-data
  - checks:
      - dangerous-workflow
    annotations:
      # the workflow is dangerous but only run under maintainers verification and approval
      - annotation: remediated
```

You can annotate multiple checks at a time:
```yml
exemptions:
  - checks:
      - binary-artifacts
      - pinned-dependencies
    annotations:
      # the binary files and files with unpinned dependencies are only used for testing
      - annotation: test-data
```

And also provide multiple annotations for checks:
```yml
exemptions:
  - checks:
      - binary-artifacts
    annotations:
      # test.exe is only used for testing
      - annotation: test-data
      # dependency.exe is needed and it's used but the binary signature is verified
      - annotation: remediated
```

The available checks are the Scorecard checks in lower case e.g. Binary-Artifacts is `binary-artifacts`.

The annotations are predefined as shown in the table below:

| Annotation | Description | Example |
|------------|-------------|---------|
| test-data | To annotate when a check or probe is targeting a danger in files or code snippets only used for test or example purposes. | The binary files are only used for testing. |
| remediated | To annotate when a check or probe correctly identified a danger and, even though the danger is necessary, a remediation was already applied. | A workflow is dangerous but only run under maintainers verification and approval, or a binary is needed but it is signed or has provenance. |
| not-applicable | To annotate when a check or probe is not applicable for the case. | The dependencies should not be pinned because the project is a library. |
| not-supported | To annotate when the maintainer fulfills a check or probe in a way that is not supported by Scorecard. | Clang-Tidy is used as SAST tool but not identified because its not supported. |
| not-detected | To annotate when the maintainer fulfills a check or probe in a way that is supported by Scorecard but not identified. | Dependabot is configured in the repository settings and not in a file. |

These annotations, when displayed in Scorecard results are parsed to a human-readable format that is similar to the annotation description described in the table above.