package checks

import "github.com/dlorenc/scorecard/checker"

type CheckResult struct {
	Pass        bool
	Message     string
	Confidence  int
	ShouldRetry bool
	Error       error
}

var InconclusiveResult = CheckResult{
	Pass:       false,
	Confidence: 0,
}

var retryResult = CheckResult{
	Pass:        false,
	ShouldRetry: true,
}

func RetryResult(err error) CheckResult {
	r := retryResult
	r.Error = err
	return r
}

type CheckFn func(*checker.Checker) CheckResult

func MultiCheck(fns ...CheckFn) CheckFn {
	threshold := 7

	return func(c *checker.Checker) CheckResult {
		var maxResult CheckResult

		for _, fn := range fns {
			result := fn(c)
			if result.Confidence > threshold {
				return result
			}
			if result.Confidence >= maxResult.Confidence {
				maxResult = result
			}
		}
		return maxResult
	}
}

type NamedCheck struct {
	Name string
	Fn   CheckFn
}

var AllChecks = []NamedCheck{}
