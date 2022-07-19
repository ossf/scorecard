# Requirements for a check

If you'd like to add a check, make sure it meets the following criteria and then
create a new GitHub Issue to discuss with the team:

-   The scorecard must only be composed of automate-able, objective data. For
    example, a project having 10 contributors doesn’t necessarily mean it’s more
    secure than a project with 50 contributors. But, having two maintainers
    might be preferable to only having one - the larger bus factor and ability
    to provide code reviews is objectively better.
-   The scorecard criteria can be as specific as possible and are not limited to
    general recommendations. For example, for Go, we can recommend/require
    specific linters and analyzers to be run on the codebase.
-   The scorecard can be populated for any open source project without any work
    or interaction from maintainers.
-   Maintainers must be provided with a mechanism to correct any automated
    scorecard findings they feel were made in error, provide "hints" for
    anything we can't detect automatically, and even dispute the applicability
    of a given scorecard finding for that repository.
-   Any criteria in the scorecard must be actionable. It should be possible,
    with help, for any project to "check all the boxes".
-   Any solution to compile a scorecard should be usable by the greater open
    source community to monitor upstream security.

# How to write a check

The steps to writing a check are as follows:

1.  Create a file under the `checks/` folder, say `checks/mycheck.go`
2.  Give the check a name and register the check:

    ```go
    // Note: export the name by starting it with an upper-case letter.
    const CheckMyCheckName string = "My-Check"

    func init() {
        registerCheck(CheckMyCheckName, EntryPointMyCheck)
    }
    ```

3.  Log useful information with `checker.DetailLogger`:

    *   Use `checker.DetailLogger.Warn()` to provide detail on low-score
        results. This is shown when the user supplies the `show-results` option.
    *   Use `checker.DetailLogger.Info()` to provide detail on high-score
        results. This is shown when the user supplies the `show-results` option.
    *   Use `checker.DetailLogger.Debug()` to provide detail in verbose mode:
        this is shown only when the user supplies the `--verbosity Debug`
        option.
    *   If your message relates to a file, try to provide information such as
        the `Path`, line number `Offset` and `Snippet`.

4.  If the check fails in a way that is irrecoverable, return a result with the
    `checker.CreateRuntimeErrorResult()` function. For example, if an error is
    returned from an API you call, use this function.

5.  Create the result of the check as follows:

    *   Always provide a high-level sentence explaining the check's
        result/score.
    *   If the check runs properly but is unable to determine a score, use
        `checker.CreateInconclusiveResult()`.
    *   For proportional results, use `checker.CreateProportionalScoreResult()`.
    *   For maximum and minimum scores, use `checker.CreateMaxScoreResult()` and
        `checker.CreateMinScoreResult()`, respectively.
    *   If you need more flexibility and need to set a specific score, use
        `checker.CreateResultWithScore()` with one of the constants declared,
        such as `checker.HalfResultScore`.

6.  Dealing with errors: see [errors/errors.md](/errors/errors.md).

7.  Create unit tests for low, high and inconclusive scores. Put them in a
    file `checks/mycheck_test.go`.

8.  Create e2e tests in `e2e/mycheck_test.go`. Use a dedicated repo that will
    not change over time, so that it's reliable for the tests.

9.  Update the `checks/checks.yaml` with a description of your check.

10. Generate `docs/check.md` using `make generate-docs`. This will validate and
    generate `docs/check.md`.

11. Update the [README.md](https://github.com/ossf/scorecard#scorecard-checks)
    with a short description of your check.

For actual examples, look at [checks/binary_artifact.go](binary_artifact.go),
[checks/code_review.go](code_review.go) and
[checks/pinned_dependencies.go](pinned_dependencies.go).
