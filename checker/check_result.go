// Copyright 2020 OpenSSF Scorecard Authors
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

// Package checker includes structs and functions used for running a check.
package checker

import (
	"errors"
	"fmt"
	"math"

	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/rule"
)

type (
	// DetailType is the type of details.
	DetailType int
)

const (
	// MaxResultScore is the best score that can be given by a check.
	MaxResultScore = 10
	// MinResultScore is the worst score that can be given by a check.
	MinResultScore = 0
	// InconclusiveResultScore is returned when no reliable information can be retrieved by a check.
	InconclusiveResultScore = -1

	// OffsetDefault is used if we can't determine the offset, for example when referencing a file but not a
	// specific location in the file.
	OffsetDefault = uint(1)
)

const (
	// DetailInfo is info-level log.
	DetailInfo DetailType = iota
	// DetailWarn is warned log.
	DetailWarn
	// DetailDebug is debug log.
	DetailDebug
)

// errSuccessTotal indicates a runtime error because number of success cases should
// be smaller than the total cases to create a proportional score.
var errSuccessTotal = errors.New("unexpected number of success is higher than total")

// CheckResult captures result from a check run.
//
//nolint:govet
type CheckResult struct {
	Name    string
	Version int
	Error   error
	Score   int
	Reason  string
	Details []CheckDetail
	// Structured results.
	Rules []string // TODO(X): add support.
}

// CheckDetail contains information for each detail.
type CheckDetail struct {
	Msg  LogMessage
	Type DetailType // Any of DetailWarn, DetailInfo, DetailDebug.
}

// LogMessage is a structure that encapsulates detail's information.
// This allows updating the definition easily.
//
//nolint:govet
type LogMessage struct {
	// Structured results.
	Finding *finding.Finding

	// Non-structured results.
	Text        string            // A short string explaining why the detail was recorded/logged.
	Path        string            // Fullpath to the file.
	Type        finding.FileType  // Type of file.
	Offset      uint              // Offset in the file of Path (line for source/text files).
	EndOffset   uint              // End of offset in the file, e.g. if the command spans multiple lines.
	Snippet     string            // Snippet of code
	Remediation *rule.Remediation // Remediation information, if any.
}

// ProportionalScoreWeighted is a structure that contains
// the fields to calculate weighted proportional scores.
type ProportionalScoreWeighted struct {
	Success int
	Total   int
	Weight  int
}

// CreateProportionalScore creates a proportional score.
func CreateProportionalScore(success, total int) int {
	if total == 0 {
		return 0
	}

	return int(math.Min(float64(MaxResultScore*success/total), float64(MaxResultScore)))
}

// CreateProportionalScoreWeighted creates the proportional score
// between multiple successes over the total, but some proportions
// are worth more.
func CreateProportionalScoreWeighted(scores ...ProportionalScoreWeighted) (int, error) {
	var ws, wt int
	allWeightsZero := true
	noScoreGroups := true
	for _, score := range scores {
		if score.Success > score.Total {
			return InconclusiveResultScore, fmt.Errorf("%w: %d, %d", errSuccessTotal, score.Success, score.Total)
		}
		if score.Total == 0 {
			continue // Group with 0 total, does not count for score
		}
		noScoreGroups = false
		if score.Weight != 0 {
			allWeightsZero = false
		}
		// Group with zero weight, adds nothing to the score

		ws += score.Success * score.Weight
		wt += score.Total * score.Weight
	}
	if noScoreGroups {
		return InconclusiveResultScore, nil
	}
	// If has score groups but no groups matter to the score, result in max score
	if allWeightsZero {
		return MaxResultScore, nil
	}

	return int(math.Min(float64(MaxResultScore*ws/wt), float64(MaxResultScore))), nil
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
func AggregateScoresWithWeight(scores map[int]int) int {
	r := 0
	ws := 0
	for s, w := range scores {
		r += s * w
		ws += w
	}
	return int(math.Floor(float64(r) / float64(ws)))
}

// NormalizeReason - placeholder function if we want to update range of scores.
func NormalizeReason(reason string, score int) string {
	return fmt.Sprintf("%v -- score normalized to %d", reason, score)
}

// CreateResultWithScore is used when
// the check runs without runtime errors, and we want to assign a
// specific score.
func CreateResultWithScore(name, reason string, score int) CheckResult {
	return CheckResult{
		Name:    name,
		Version: 2,
		Error:   nil,
		Score:   score,
		Reason:  reason,
	}
}

// CreateProportionalScoreResult is used when
// the check runs without runtime errors and we assign a
// proportional score. This may be used if a check contains
// multiple tests, and we want to assign a score proportional
// the number of tests that succeeded.
func CreateProportionalScoreResult(name, reason string, b, t int) CheckResult {
	score := CreateProportionalScore(b, t)
	return CheckResult{
		Name: name,
		// Old structure.
		// New structure.
		Version: 2,
		Error:   nil,
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
		Name:    name,
		Version: 2,
		Score:   InconclusiveResultScore,
		Reason:  reason,
	}
}

// CreateRuntimeErrorResult is used when the check fails to run because of a runtime error.
func CreateRuntimeErrorResult(name string, e error) CheckResult {
	return CheckResult{
		Name:    name,
		Version: 2,
		Error:   e,
		Score:   InconclusiveResultScore,
		Reason:  e.Error(), // Note: message already accessible by caller through `Error`.
	}
}

// LogFindings logs the list of findings.
func LogFindings(findings []finding.Finding, dl DetailLogger) {
	for i := range findings {
		f := &findings[i]
		switch f.Outcome {
		case finding.OutcomeNegative:
			dl.Warn(&LogMessage{
				Finding: f,
			})
		case finding.OutcomePositive:
			dl.Info(&LogMessage{
				Finding: f,
			})
		default:
			dl.Debug(&LogMessage{
				Finding: f,
			})
		}
	}
}
