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
	"github.com/ossf/scorecard/v3/checker"
	sce "github.com/ossf/scorecard/v3/errors"
)

// DependencyUpdateTool applies the score policy for the Dependency-Update-Tool check.
func DependencyUpdateTool(name string, dl checker.DetailLogger,
	r *checker.DependencyUpdateToolData) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	// Apply the policy evaluation.
	if r.ConfigFiles == nil || len(r.ConfigFiles) == 0 {
		dl.Warn3(&checker.LogMessage{
			Text: `dependabot config file not detected in source location.
			We recommend setting this configuration in code so it can be easily verified by others.`,
		})
		dl.Warn3(&checker.LogMessage{
			Text: `renovatebot config file not detected in source location.
			We recommend setting this configuration in code so it can be easily verified by others.`,
		})
		return checker.CreateMinScoreResult(name, "no update tool detected")
	}

	for _, f := range r.ConfigFiles {
		dl.Info3(&checker.LogMessage{
			Path:   f.File.Path,
			Type:   f.File.Type,
			Offset: f.File.Offset,
		})
	}

	// High score result.
	return checker.CreateMaxScoreResult(name, "update tool detected")
}
