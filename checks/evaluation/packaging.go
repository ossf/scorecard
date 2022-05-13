// Copyright 2021 Security Scorecard Authors
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

package evaluation

import (
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
)

// Packaging applies the score policy for the Packaging check.
func Packaging(name string, dl checker.DetailLogger, r *checker.PackagingData) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	passed := false
	for _, p := range r.Packages {
		switch p.Outcome {
		case checker.OutcomeTypeDebug:
			msg, err := createLogMessage(p)
			if err != nil {
				return checker.CreateRuntimeErrorResult(name, err)
			}
			dl.Debug(&msg)

		case checker.OutcomeTypeNegativeLow, checker.OutcomeTypeNegativeMedium,
			checker.OutcomeTypeNegativeHigh, checker.OutcomeTypeNegativeCritical:

			msg, err := createLogMessage(p)
			if err != nil {
				return checker.CreateRuntimeErrorResult(name, err)
			}
			dl.Warn(&msg)

		case checker.OutcomeTypePositiveLow, checker.OutcomeTypePositiveMedium,
			checker.OutcomeTypePositiveHigh, checker.OutcomeTypePositiveCritical:

			msg, err := createLogMessage(p)
			if err != nil {
				return checker.CreateRuntimeErrorResult(name, err)
			}

			passed = true
			dl.Info(&msg)
		default:
			e := sce.WithMessage(sce.ErrScorecardInternal,
				fmt.Sprintf("outcome type not expected: %v", p.Outcome))
			return checker.CreateRuntimeErrorResult(name, e)
		}
	}

	if passed {
		return checker.CreateMaxScoreResult(name,
			"publishing workflow detected")
	}

	return checker.CreateInconclusiveResult(name,
		"no published package detected")
}

func createLogMessage(p checker.Package) (checker.LogMessage, error) {
	var msg checker.LogMessage

	if p.File == nil && p.Msg == nil {
		return msg, sce.WithMessage(sce.ErrScorecardInternal, "File and Msg fields are nil")
	}

	if p.File != nil {
		msg.Path = p.File.Path
		msg.Type = p.File.Type
		msg.Offset = p.File.Offset
	}

	if p.Msg != nil {
		msg.Text = *p.Msg
	}

	return msg, nil
}
