# How to handle errors

```golang
import sce "github.com/ossf/scorecard/errors"

// Public errors are defined in errors/public.go and are exposed to callers.
// Internal errors are defined in errors/innternal.go. Their names start with ErrInternalXXX

// Examples:

// Return a standard check run failure, with an error message from an internal error.
// We only create internal errors for errors that may happen in several places in the code: this provides
// consistent erorr messages to the caller. 
return sce.Create(sce.ErrRunFailure, ErrInternalInvalidYamlFile.Error())

// Return a standard check run failure, with an error message from an internal error and an API call error.
err := dependency.apiCall()
if err != nil {
    return sce.Create(sce.ErrRunFailure, fmt.Sprintf("%v: %v", sce.ErrInternalSomething, err))
}

// Return a standard check run failure, only with API call error. Use this format when there is no internal error associated
// to the failure. In many cases, we don't need internal errors.
err := dependency.apiCall()
if err != nil {
    return sce.Create(sce.ErrRunFailure, fmt.Sprintf("dependency.apiCall: %v", err))
}
```