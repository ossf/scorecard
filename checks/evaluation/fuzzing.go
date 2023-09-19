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
	"github.com/ossf/scorecard/v4/probes/fuzzedWithCLibFuzzer"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithClusterFuzzLite"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithCppLibFuzzer"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithGoNative"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithJavaJazzerFuzzer"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithOSSFuzz"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithOneFuzz"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithPropertyBasedHaskell"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithPropertyBasedJavascript"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithPropertyBasedTypescript"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithPythonAtheris"
	"github.com/ossf/scorecard/v4/probes/fuzzedWithRustCargofuzz"
)

// Fuzzing applies the score policy for the Fuzzing check.
func Fuzzing(name string,
	findings []finding.Finding,
) checker.CheckResult {
	// We have 7 unique probes, each should have a finding.
	expectedProbes := []string{
		fuzzedWithClusterFuzzLite.Probe,
		fuzzedWithGoNative.Probe,
		fuzzedWithPythonAtheris.Probe,
		fuzzedWithCLibFuzzer.Probe,
		fuzzedWithCppLibFuzzer.Probe,
		fuzzedWithRustCargofuzz.Probe,
		fuzzedWithJavaJazzerFuzzer.Probe,
		fuzzedWithOneFuzz.Probe,
		fuzzedWithOSSFuzz.Probe,
		fuzzedWithPropertyBasedHaskell.Probe,
		fuzzedWithPropertyBasedJavascript.Probe,
		fuzzedWithPropertyBasedTypescript.Probe,
	}
	// TODO: other packages to consider:
	// - github.com/google/fuzztest

	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	// Compute the score.
	for i := range findings {
		f := &findings[i]
		if f.Outcome == finding.OutcomePositive {
			return checker.CreateMaxScoreResult(name, "project is fuzzed")
		}
	}
	return checker.CreateMinScoreResult(name, "project is not fuzzed")
}
