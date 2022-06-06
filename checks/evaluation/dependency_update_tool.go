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

package evaluation

import (
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
)

// DependencyUpdateTool applies the score policy for the Dependency-Update-Tool check.
func DependencyUpdateTool(name string, dl checker.DetailLogger,
	r *checker.DependencyUpdateToolData,
) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	// Apply the policy evaluation.
	if r.Tools == nil || len(r.Tools) == 0 {
		dl.Warn(&checker.LogMessage{
			Text: `dependabot config file not detected in source location.
			We recommend setting this configuration in code so it can be easily verified by others.`,
		})
		dl.Warn(&checker.LogMessage{
			Text: `renovatebot config file not detected in source location.
			We recommend setting this configuration in code so it can be easily verified by others.`,
		})
		return checker.CreateMinScoreResult(name, "no update tool detected")
	}

	// Validate the input.
	if len(r.Tools) != 1 {
		e := sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("found %d tools, expected 1", len(r.Tools)))
		return checker.CreateRuntimeErrorResult(name, e)
	}

	if r.Tools[0].Files == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "Files are nil")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	// Iterate over all the files, since a Tool can contain multiple files.
	for _, file := range r.Tools[0].Files {
		dl.Info(&checker.LogMessage{
			Path:   file.Path,
			Type:   file.Type,
			Offset: file.Offset,
			Text:   fmt.Sprintf("%s detected", r.Tools[0].Name),
		})
	}

	// High score result.
	return checker.CreateMaxScoreResult(name, "update tool detected")
}
