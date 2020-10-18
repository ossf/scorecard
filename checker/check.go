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

var maxConfidence int = 10

func RetryResult(err error) CheckResult {
	r := retryResult
	r.Error = err
	return r
}

type CheckFn func(Checker) CheckResult

func Bool2int(b bool) int {
	if b {
		return 1
	}
	return 0
}

func MultiCheck(fns ...CheckFn) CheckFn {
	return func(c Checker) CheckResult {
		var maxResult CheckResult

		for _, fn := range fns {
			result := fn(c)
			if Bool2int(result.Pass) < Bool2int(maxResult.Pass) {
				continue
			}
			if result.Pass && result.Confidence >= maxConfidence {
				return result
			}
			if result.Confidence >= maxResult.Confidence {
				maxResult = result
			}
		}
		return maxResult
	}
}

func ProportionalResult(numerator int, denominator int, threshold float32) CheckResult {
	if numerator == 0 {
		return CheckResult{
			Pass:       false,
			Confidence: maxConfidence,
		}
	}

	actual := float32(numerator) / float32(denominator)
	if actual >= threshold {
		return CheckResult{
			Pass:       true,
			Confidence: int(actual * 10),
		}
	}

	return CheckResult{
		Pass:       false,
		Confidence: int(maxConfidence - int(actual*10)),
	}
}

type NamedCheck struct {
	Name string
	Fn   CheckFn
}
