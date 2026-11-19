# How to handle errors

Public errors are defined in [errors/public.go](public.go) and are exposed to
callers, who test for them with `errors.Is`. `sce.ErrScorecardInternal` marks
a failure inside Scorecard; others, such as `sce.ErrRepoUnreachable`, describe
a problem with the request or the repository.

Errors used only inside a package are defined in that package, for example
[checks/raw/errors.go](../checks/raw/errors.go). We only create these for
errors that may happen in several places in the code: this gives the caller
consistent error messages.

```golang
import sce "github.com/ossf/scorecard/v5/errors"

// Examples:

// In a check's collector, return the error with context. The check's
// definition reports any collector error as an internal error.
files, err := c.RepoClient.ListFiles(predicate)
if err != nil {
    return checker.MyCheckData{}, fmt.Errorf("RepoClient.ListFiles: %w", err)
}

// Return a package's internal error, so the message is the same wherever
// the failure happens.
return checker.MyCheckData{}, errInternalInvalidDockerFile

// Wrap a public error with a message.
return checker.MyCheckData{}, sce.WithMessage(sce.ErrScorecardInternal, "bad file or information reference")

// In a check's evaluation, return a runtime error result.
e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
return checker.CreateRuntimeErrorResult(name, e)
```

See [checks/write.md](../checks/write.md#errors) for where each part of a
check reports errors.
