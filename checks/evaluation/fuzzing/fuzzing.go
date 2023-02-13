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
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
)

//go:embed *.yml
var rules embed.FS

var (
	fuzzingWithClustterFuzzLite = "FuzzingWithClusterFuzzLite"
	fuzzingWithOneFuzz          = "FuzzingWithOneFuzz"
	fuzzingWithOSSFuzz          = "FuzzingWithOSSFuzz"
	fuzzingWithGoNative         = "FuzzingWithGoNative"
)

// Fuzzing applies the score policy for the Fuzzing check.
func Fuzzing(name string, dl checker.DetailLogger,
	r *checker.FuzzingData,
) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	// Keep track of reported results.
	reportedRuleResults := make(map[string]bool)

	if len(r.Fuzzers) == 0 {
		// Report findings for all rules.
		if err := logDefaultFindings(dl, reportedRuleResults); err != nil {
			return checker.CreateRuntimeErrorResult(name, err)
		}
		return checker.CreateMinScoreResult(name, "project is not fuzzed")
	}
	fuzzers := []string{}
	for i := range r.Fuzzers {
		fuzzer := r.Fuzzers[i]
		// NOTE: files are not populated for all fuzzers yet.
		// To simplify the code, we currently do not report the locations.
		switch fuzzer.Name {
		case "GoNativeFuzzer":
			if err := checker.LogFinding(rules, fuzzingWithGoNative,
				"project fuzzed with Go native framework",
				nil, finding.OutcomePositive, dl); err != nil {
				return checker.CreateRuntimeErrorResult(name, err)
			}
			reportedRuleResults[fuzzingWithGoNative] = true
		case "OSSFuzz":
			if err := checker.LogFinding(rules, fuzzingWithOSSFuzz,
				"project fuzzed with OSS-Fuzz",
				nil, finding.OutcomePositive, dl); err != nil {
				return checker.CreateRuntimeErrorResult(name, err)
			}
			reportedRuleResults[fuzzingWithOSSFuzz] = true
		case "ClusterFuzzLite":
			if err := checker.LogFinding(rules, fuzzingWithClustterFuzzLite,
				"project fuzzed with ClusterFuzzLite",
				nil, finding.OutcomePositive, dl); err != nil {
				return checker.CreateRuntimeErrorResult(name, err)
			}
			reportedRuleResults[fuzzingWithClustterFuzzLite] = true
		case "OneFuzz":
			if err := checker.LogFinding(rules, fuzzingWithOneFuzz,
				"project fuzzed with OneFuzz",
				nil, finding.OutcomePositive, dl); err != nil {
				return checker.CreateRuntimeErrorResult(name, err)
			}
			reportedRuleResults[fuzzingWithOneFuzz] = true
		default:
			return checker.CreateRuntimeErrorResult(name, fmt.Errorf("unsupported fuzzer: %v", fuzzer))
		}
		fuzzers = append(fuzzers, fuzzer.Name)
	}
	// Report findings for all rules.
	if err := logDefaultFindings(dl, reportedRuleResults); err != nil {
		return checker.CreateRuntimeErrorResult(name, err)
	}

	return checker.CreateMaxScoreResult(name,
		fmt.Sprintf("project is fuzzed with %v", fuzzers))
}

func logDefaultFindings(dl checker.DetailLogger, r map[string]bool) error {
	// We always report at least one finding for each rule.
	if !r[fuzzingWithGoNative] {
		if err := checker.LogFinding(rules, fuzzingWithGoNative,
			"no fuzzing using Go native framework",
			nil, finding.OutcomeNegative, dl); err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
	}

	if !r[fuzzingWithOSSFuzz] {
		if err := checker.LogFinding(rules, fuzzingWithOSSFuzz,
			"no fuzzing using OSS-Fuzz",
			nil, finding.OutcomeNegative, dl); err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
	}

	if !r[fuzzingWithOneFuzz] {
		if err := checker.LogFinding(rules, fuzzingWithOneFuzz,
			"no fuzzing using OneFuzz",
			nil, finding.OutcomeNegative, dl); err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
	}

	if !r[fuzzingWithClustterFuzzLite] {
		if err := checker.LogFinding(rules, fuzzingWithClustterFuzzLite,
			"no fuzzing using ClusterFuzzLite",
			nil, finding.OutcomeNegative, dl); err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
	}

	return nil
}
