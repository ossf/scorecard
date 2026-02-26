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
	"github.com/ossf/scorecard/v5/checker"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/fuzzed"
)

// Fuzzing applies the score policy for the Fuzzing check.
func Fuzzing(name string,
	findings []finding.Finding, dl checker.DetailLogger,
	usesMemoryUnsafeLanguage bool,
) checker.CheckResult {
	expectedProbes := []string{
		fuzzed.Probe,
	}
	// TODO: other packages to consider:
	// - github.com/google/fuzztest

	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	var fuzzerDetected bool
	// Compute the score.
	for i := range findings {
		f := &findings[i]
		var logLevel checker.DetailType
		switch f.Outcome {
		case finding.OutcomeFalse:
			logLevel = checker.DetailWarn
		case finding.OutcomeTrue:
			fuzzerDetected = true
			logLevel = checker.DetailInfo
		default:
			logLevel = checker.DetailDebug
		}
		checker.LogFinding(dl, f, logLevel)
	}

	if fuzzerDetected {
		return checker.CreateMaxScoreResult(name, "project is fuzzed")
	}

	if usesMemoryUnsafeLanguage {
		return checker.CreateMinScoreResult(name, "project is not fuzzed")
	}

	return checker.CreateInconclusiveResult(name, "fuzzing not required for memory-safe languages")
}
