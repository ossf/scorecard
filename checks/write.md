# How to write a check
The steps to writting a check are as follow:

1. Create a file under checks, say `checks/mycheck.go`
2. Decide on a name, register the check:
```
// Note: do not export the name: start its name with a lower-case letter.
const checkMyChech string = "My-Check"

func init() {
	registerCheck(checkBinaryArtifacts, BinaryArtifacts)
}
```
3. Log information that is benfical to the user using `checker.DetailLogger`:
    * Use `checker.DetailLogger.Warn()` to provide detail on low-score results. This is showed when the user supplies the `show-results` option.
    * Use `checker.DetailLogger.Info()` to provide detail on high-score results. This is showed when the user supplies the `show-results` option.
    * Use `checker.DetailLogger.Debug()` to provide detail on in verbose mode: this is showed only when the user supplies the `--verbosity Debug` option.
4. If the checks fails to run in a way that is irrecoverable, use `checker.CreateRuntimeErrorResult()` function. An exmple of this is if an error is returned from an API you call.
5. Create the result of the check as follow:
    * Always provide a high-level sentence explaining the result/score of the check.
    * If the check runs properly but is unable to conclude babout the score, use `checker.CreateInconclusiveResult()` function.
    * For propertional results, use `checker.CreateProportionalScoreResult()`.
    * For maximum score, use `checker.CreateMaxScoreResult()`; for min score use `checker.CreateMinScoreResult()`
    * If you need more flexibility and need to set a specific score, use `checker.CreateResultWithScore()` with one of the constants declared, such as `checker.HalfResultScore`.
    --
6. Dealing with errors: see [../errors/errors.md](errors/errors/md).
7. Create unit tests for both low, high and inconclusive score. Put them in a file `checks/mycheck_test.go`
8. Create e2e tests in `e2e/mycheck_test.go`. Use a dedicated repo whereata will not change over time, so that it's reliable for the tests.
9. Update the `checks/checks.yaml` with the description of your check. 
10. Gerenate the `checks/check.md` using `go build && cd checks/main && ./main`. Verify `checks/check.md` was updated.
10. Update the [README.md](https://github.com/ossf/scorecard#scorecard-checks) with a short description of your check.

For actual examples, look at [checks/binary_artifact.go](binary_artifact.go), [checks/code_review.go](code_review.go) and [checks/frozen_deps.go](frozen_deps.go).