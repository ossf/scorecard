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

// DangerousWorkflow applies the score policy for the DangerousWorkflow check.
func DangerousWorkflow(name string, dl checker.DetailLogger,
	r *checker.DangerousWorkflowData,
) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	for _, e := range r.Workflows {
		var text string
		switch e.Type {
		case checker.DangerousWorkflowUntrustedCheckout:
			text = fmt.Sprintf("untrusted code checkout '%v'", e.File.Snippet)
		case checker.DangerousWorkflowScriptInjection:
			text = fmt.Sprintf("script injection with untrusted input '%v'", e.File.Snippet)
		default:
			err := sce.WithMessage(sce.ErrScorecardInternal, "invalid type")
			return checker.CreateRuntimeErrorResult(name, err)
		}

		dl.Warn(&checker.LogMessage{
			Path:    e.File.Path,
			Type:    e.File.Type,
			Offset:  e.File.Offset,
			Text:    text,
			Snippet: e.File.Snippet,
		})
	}

	if len(r.Workflows) > 0 {
		return createResult(name, checker.MinResultScore)
	}
	return createResult(name, checker.MaxResultScore)
}

// Create the result.
func createResult(name string, score int) checker.CheckResult {
	if score != checker.MaxResultScore {
		return checker.CreateResultWithScore(name,
			"dangerous workflow patterns detected", score)
	}

	return checker.CreateMaxScoreResult(name,
		"no dangerous workflow patterns detected")
}
