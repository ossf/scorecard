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

package evaluation

import (

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
)

// DependencyUpdateTool applies the score policy for the Dependency-Update-Tool check.
func DependencyUpdateTool(name string, dl checker.DetailLogger,
	findings []finding.Finding,
) checker.CheckResult {
	
	for i := range findings {
		f := findings[i]
		if f.Outcome == finding.OutcomePositive {
			return checker.CreateMaxScoreResult(name, "update tool detected")
		}
	}

	return checker.CreateMinScoreResult(name, "no update tool detected")
}
