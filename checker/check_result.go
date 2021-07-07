// Copyright 2020 Security Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package checker

import (
	"errors"

	scorecarderrors "github.com/ossf/scorecard/errors"
)

const (
	MaxResultConfidence  = 10
	HalfResultConfidence = 5
	MinResultConfidence  = 0
)

// ErrorDemoninatorZero indicates the denominator for a proportional result is 0.
var ErrorDemoninatorZero = errors.New("internal error: denominator is 0")

// Types of details.
type DetailType int

const (
	DetailFail DetailType = iota
	DetailPass
	DetailInfo
	DetailWarn
	DetailDebug
)

// CheckDetail contains information for each detail.
//nolint:govet
type CheckDetail struct {
	Type DetailType // Any of DetailFail, DetailPass, DetailInfo.
	Code string     // A string identifying the sub-check, e.g. to lookup remediation info.
	Desc string     // A short string representation of the information.
}

// Types of results.
const (
	ResultPass     = 0
	ResultFail     = 1
	ResultDontKnow = 2
)

type CheckResult struct {
	Error      error `json:"-"`
	Name       string
	Details    []string
	Details2   []CheckDetail
	Confidence int
	// Note: Pass2 will ultimately be renamed
	// as Pass.
	Pass        bool
	Pass2       int
	ShouldRetry bool `json:"-"`
}

// Will be removed.
func MakeInconclusiveResult(name string, err error) CheckResult {
	return CheckResult{
		Name:       name,
		Pass:       false,
		Confidence: 0,
		Pass2:      ResultDontKnow,
		Error:      scorecarderrors.MakeLowConfidenceError(err),
	}
}

// TODO: these functions should set the details as well.
func MakeInternalErrorResult(name string, err error) CheckResult {
	return CheckResult{
		Name:       name,
		Pass:       false,
		Confidence: 0,
		Pass2:      ResultDontKnow,
		Error:      scorecarderrors.MakeLowConfidenceError(err),
	}
}

func MakeInconclusiveResult2(name string, c *CheckRequest, reason string) CheckResult {
	c.CLogger.Warn("lowering result confidence to %d because %s", 0, reason)
	return CheckResult{
		Name:       name,
		Pass:       false,
		Confidence: 0,
		Pass2:      ResultDontKnow,
		Error:      nil,
	}
}

func MakePassResult(name string) CheckResult {
	return CheckResult{
		Name:       name,
		Pass:       true,
		Pass2:      ResultPass,
		Confidence: MaxResultConfidence,
		Error:      nil,
	}
}

func MakePassResultWithHighConfidence(name string) CheckResult {
	return CheckResult{
		Name:       name,
		Pass:       true,
		Pass2:      ResultPass,
		Confidence: MaxResultConfidence,
		Error:      nil,
	}
}

func MakePassResultWithHighConfidenceAndReason(name string, c *CheckRequest, reason string) CheckResult {
	c.CLogger.Pass("%s", reason)
	return CheckResult{
		Name:       name,
		Pass:       true,
		Pass2:      ResultPass,
		Confidence: MaxResultConfidence,
		Error:      nil,
	}
}

func MakePassResultWithHighConfidenceAndReasonAndCode(name string, c *CheckRequest, code, reason string) CheckResult {
	c.CLogger.PassWithCode(code, reason)
	return CheckResult{
		Name:       name,
		Pass:       true,
		Pass2:      ResultPass,
		Confidence: MaxResultConfidence,
		Error:      nil,
	}
}

func MakePassResultWithLowConfidenceAndReason(name string, c *CheckRequest, conf int, reason string) CheckResult {
	c.CLogger.Warn("%s (lowering confidence to %d)", reason, conf)
	return CheckResult{
		Name:       name,
		Pass:       true,
		Pass2:      ResultPass,
		Confidence: conf,
		Error:      nil,
	}
}

func MakeFailResult(name string, err error) CheckResult {
	return CheckResult{
		Name:       name,
		Pass:       false,
		Pass2:      ResultFail,
		Confidence: MaxResultConfidence,
		Error:      err,
	}
}

func MakeFailResultWithHighConfidence(name string) CheckResult {
	return CheckResult{
		Name:       name,
		Pass:       false,
		Pass2:      ResultFail,
		Confidence: MaxResultConfidence,
		Error:      nil,
	}
}

func MakeFailResultWithHighConfidenceAndReason(name string, c *CheckRequest, reason string) CheckResult {
	c.CLogger.Info("%s", reason)
	return CheckResult{
		Name:       name,
		Pass:       false,
		Pass2:      ResultFail,
		Confidence: MaxResultConfidence,
		Error:      nil,
	}
}

func MakeFailResultWithHighConfidenceAndReasonAndCode(name string, c *CheckRequest, code, reason string) CheckResult {
	c.CLogger.FailWithCode(code, reason)
	return CheckResult{
		Name:       name,
		Pass:       false,
		Pass2:      ResultFail,
		Confidence: MaxResultConfidence,
		Error:      nil,
	}
}

func MakeFailResultLowConfidenceAndReason(name string, c *CheckRequest, conf int, reason string) CheckResult {
	c.CLogger.Fail("%s (lowering confidence to %d)", reason, conf)
	return CheckResult{
		Name:       name,
		Pass:       false,
		Pass2:      ResultFail,
		Confidence: conf,
		Error:      nil,
	}
}

func MakeRetryResult(name string, err error) CheckResult {
	return CheckResult{
		Name:        name,
		Pass:        false,
		Pass2:       ResultDontKnow,
		ShouldRetry: true,
		Error:       scorecarderrors.MakeRetryError(err),
	}
}

// TODO: update this function to return a ResultDontKnow
// if the confidence is low?
func MakeProportionalResult(name string, numerator int, denominator int,
	threshold float32) CheckResult {
	if denominator == 0 {
		return MakeInconclusiveResult(name, ErrorDemoninatorZero)
	}
	if numerator == 0 {
		return CheckResult{
			Name:       name,
			Pass:       false,
			Pass2:      ResultFail,
			Confidence: MaxResultConfidence,
		}
	}
	actual := float32(numerator) / float32(denominator)
	if actual >= threshold {
		return CheckResult{
			Name:       name,
			Pass:       true,
			Pass2:      ResultPass,
			Confidence: int(actual * MaxResultConfidence),
		}
	}

	return CheckResult{
		Name:       name,
		Pass:       false,
		Pass2:      ResultFail,
		Confidence: MaxResultConfidence - int(actual*MaxResultConfidence),
	}
}

// Given a min result, check if another result is worse.
func isMinResult(result, min CheckResult) bool {
	if Bool2int(result.Pass) < Bool2int(min.Pass) {
		return true
	}
	if result.Pass && result.Confidence < min.Confidence {
		return true
	} else if !result.Pass && result.Confidence > min.Confidence {
		return true
	}
	return false
}

// MakeAndResult means all checks must succeed. This returns a conservative result
// where the worst result is returned.
func MakeAndResult(checks ...CheckResult) CheckResult {
	minResult := CheckResult{
		Pass:       true,
		Pass2:      ResultPass,
		Confidence: MaxResultConfidence,
	}

	for _, result := range checks {
		if minResult.Name == "" {
			minResult.Name = result.Name
		}
		if isMinResult(result, minResult) {
			minResult = result
		}
	}
	return minResult
}
