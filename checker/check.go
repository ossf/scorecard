package checker

type CheckResult struct {
	Pass        bool
	Details     string
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

type CheckFn func(Checker) CheckResult

func MultiCheck(fns ...CheckFn) CheckFn {
	threshold := 7

	return func(c Checker) CheckResult {
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

func ProportionalResult(numerator, denominator int, threshold float32) CheckResult {
	actual := float32(numerator) / float32(denominator)
	if actual >= threshold {
		return CheckResult{
			Pass:       true,
			Confidence: int(actual * 10),
		}
	}
	return CheckResult{
		Pass:       false,
		Confidence: int(10 - int(actual*10)),
	}
}

type NamedCheck struct {
	Name string
	Fn   CheckFn
}
