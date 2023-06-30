// Copyright 2021 OpenSSF Scorecard Authors
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
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
)

// Fuzzing applies the score policy for the Fuzzing check.
func Fuzzing(name string,
	findings []finding.Finding,
) checker.CheckResult {
	// The probes should always contain at least on finding.
	if len(findings) == 0 {
		e := sce.WithMessage(sce.ErrScorecardInternal, "no findings")
		return checker.CreateRuntimeErrorResult(name, e)
	}
	for i := range findings {
		f := &findings[i]
		if f.Outcome == finding.OutcomePositive {
			return checker.CreateMaxScoreResult(name, "project is fuzzed")
		}
	}
	return checker.CreateMinScoreResult(name, "project is not fuzzed")
}
