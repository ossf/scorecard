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
	"fmt"
	"math"
	"strings"

	scorecarderrors "github.com/ossf/scorecard/v2/errors"
)

// UPGRADEv2: to remove.
const (
	MaxResultConfidence  = 10
	HalfResultConfidence = 5
	MinResultConfidence  = 0
)

// UPGRADEv2: to remove.
const migrationThresholdPassValue = 8

// ErrorDemoninatorZero indicates the denominator for a proportional result is 0.
// UPGRADEv2: to remove.
var ErrorDemoninatorZero = errors.New("internal error: denominator is 0")

// Types of details.
type DetailType int

const (
	DetailInfo DetailType = iota
	DetailWarn
	DetailDebug
)

// CheckDetail contains information for each detail.
//nolint:govet
type CheckDetail struct {
	Type DetailType // Any of DetailWarn, DetailInfo, DetailDebug.
	Msg  string     // A short string explaining why the details was recorded/logged..
}

type DetailLogger interface {
	Info(desc string, args ...interface{})
	Warn(desc string, args ...interface{})
	Debug(desc string, args ...interface{})
}

//nolint
const (
	MaxResultScore          = 10
	MinResultScore          = 0
	InconclusiveResultScore = -1
)

//nolint
type CheckResult struct {
	// Old structure
	Error       error `json:"-"`
	Name        string
	Details     []string
	Confidence  int
	Pass        bool
	ShouldRetry bool `json:"-"`

	// UPGRADEv2: New structure. Omitting unchanged Name field
	// for simplicity.
	Version  int           `json:"-"` // Default value of 0 indicates old structure.
	Error2   error         `json:"-"` // Runtime error indicate a filure to run the check.
	Details2 []CheckDetail `json:"-"` // Details of tests and sub-checks
	Score    int           `json:"-"` // {[-1,0...10], -1 = Inconclusive}
	Reason   string        `json:"-"` // A sentence describing the check result (score, etc)
}

// CreateProportionalScore() creates a proportional score.
func CreateProportionalScore(b, t int) int {
	if t == 0 {
		return 0
	}

	return int(math.Min(float64(MaxResultScore*b/t), float64(MaxResultScore)))
}

// AggregateScores adds up all scores
// and normalizes the result.
// Each score contributes equally.
func AggregateScores(scores ...int) int {
	n := float64(len(scores))
	r := 0
	for _, s := range scores {
		r += s
	}
	return int(math.Floor(float64(r) / n))
}

// AggregateScoresWithWeight adds up all scores
// and normalizes the result.
// The caller is responsible for ensuring the sum of
// weights is 10.
func AggregateScoresWithWeight(scores map[int]int) int {
	r := 0
	for s, w := range scores {
		r += s * w
	}
	return int(math.Floor(float64(r) / float64(MaxResultScore)))
}

func NormalizeReason(reason string, score int) string {
	return fmt.Sprintf("%v -- score normalized to %d", reason, score)
}

// CreateResultWithScore is used when
// the check runs without runtime errors and we want to assign a
// specific score.
func CreateResultWithScore(name, reason string, score int) CheckResult {
	pass := true
	if score < migrationThresholdPassValue {
		pass = false
	}
	return CheckResult{
		Name: name,
		// Old structure.
		Error:       nil,
		Confidence:  MaxResultScore,
		Pass:        pass,
		ShouldRetry: false,
		// New structure.
		//nolint
		Version: 2,
		Error2:  nil,
		Score:   score,
		Reason:  reason,
	}
}

// CreateProportionalScoreResult is used when
// the check runs without runtime errors and we assign a
// proportional score. This may be used if a check contains
// multiple tests and we want to assign a score proportional
// the the number of tests that succeeded.
func CreateProportionalScoreResult(name, reason string, b, t int) CheckResult {
	pass := true
	score := CreateProportionalScore(b, t)
	if score < migrationThresholdPassValue {
		pass = false
	}
	return CheckResult{
		Name: name,
		// Old structure.
		Error:       nil,
		Confidence:  MaxResultConfidence,
		Pass:        pass,
		ShouldRetry: false,
		// New structure.
		//nolint
		Version: 2,
		Error2:  nil,
		Score:   score,
		Reason:  NormalizeReason(reason, score),
	}
}

// CreateMaxScoreResult is used when
// the check runs without runtime errors and we can assign a
// maximum score to the result.
func CreateMaxScoreResult(name, reason string) CheckResult {
	return CreateResultWithScore(name, reason, MaxResultScore)
}

// CreateMinScoreResult is used when
// the check runs without runtime errors and we can assign a
// minimum score to the result.
func CreateMinScoreResult(name, reason string) CheckResult {
	return CreateResultWithScore(name, reason, MinResultScore)
}

// CreateInconclusiveResult is used when
// the check runs without runtime errors, but we don't
// have enough evidence to set a score.
func CreateInconclusiveResult(name, reason string) CheckResult {
	return CheckResult{
		Name: name,
		// Old structure.
		Confidence:  0,
		Pass:        false,
		ShouldRetry: false,
		// New structure.
		//nolint
		Version: 2,
		Score:   InconclusiveResultScore,
		Reason:  reason,
	}
}

// CreateRuntimeErrorResult is used when the check fails to run because of a runtime error.
func CreateRuntimeErrorResult(name string, e error) CheckResult {
	return CheckResult{
		Name: name,
		// Old structure.
		Error:       e,
		Confidence:  0,
		Pass:        false,
		ShouldRetry: false,
		// New structure.
		//nolint
		Version: 2,
		Error2:  e,
		Score:   InconclusiveResultScore,
		Reason:  e.Error(), // Note: message already accessible by caller thru `Error`.
	}
}

// UPGRADEv2: functions below will be renamed.
func MakeAndResult2(checks ...CheckResult) CheckResult {
	if len(checks) == 0 {
		// That should never happen.
		panic("MakeResult called with no checks")
	}

	worseResult := checks[0]
	// UPGRADEv2: will go away after old struct is removed.
	//nolint
	for _, result := range checks[1:] {
		if result.Score < worseResult.Score {
			worseResult = result
		}
	}
	return worseResult
}

func MakeOrResult(c *CheckRequest, checks ...CheckResult) CheckResult {
	if len(checks) == 0 {
		// That should never happen.
		panic("MakeResult called with no checks")
	}

	bestResult := checks[0]
	//nolint
	for _, result := range checks[1:] {
		if result.Score >= bestResult.Score {
			i := strings.Index(bestResult.Reason, "-- score normalized")
			if i < 0 {
				i = len(bestResult.Reason)
			}
			c.Dlogger.Info(bestResult.Reason[:i])
			bestResult = result
		} else {
			i := strings.Index(result.Reason, "-- score normalized")
			if i < 0 {
				i = len(result.Reason)
			}
			c.Dlogger.Info(result.Reason[:i])
		}

		// Do not exit early so we can show all the details
		// to the user.
	}

	return bestResult
}

func MakeInconclusiveResult(name string, err error) CheckResult {
	return CheckResult{
		Name:       name,
		Pass:       false,
		Confidence: 0,
		Error:      scorecarderrors.MakeLowConfidenceError(err),
	}
}

func MakePassResult(name string) CheckResult {
	return CheckResult{
		Name:       name,
		Pass:       true,
		Confidence: MaxResultConfidence,
		Error:      nil,
	}
}

func MakeFailResult(name string, err error) CheckResult {
	return CheckResult{
		Name:       name,
		Pass:       false,
		Confidence: MaxResultConfidence,
		Error:      err,
	}
}

func MakeRetryResult(name string, err error) CheckResult {
	return CheckResult{
		Name:        name,
		Pass:        false,
		ShouldRetry: true,
		Error:       scorecarderrors.MakeRetryError(err),
	}
}

func MakeProportionalResult(name string, numerator int, denominator int,
	threshold float32) CheckResult {
	if denominator == 0 {
		return MakeInconclusiveResult(name, ErrorDemoninatorZero)
	}
	if numerator == 0 {
		return CheckResult{
			Name:       name,
			Pass:       false,
			Confidence: MaxResultConfidence,
		}
	}
	actual := float32(numerator) / float32(denominator)
	if actual >= threshold {
		return CheckResult{
			Name:       name,
			Pass:       true,
			Confidence: int(actual * MaxResultConfidence),
		}
	}

	return CheckResult{
		Name:       name,
		Pass:       false,
		Confidence: MaxResultConfidence - int(actual*MaxResultConfidence),
	}
}

// Given a min result, check if another result is worse.
//nolint
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
		Confidence: MaxResultConfidence,
	}
	// UPGRADEv2: will go away after old struct is removed.
	//nolint
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
