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

// Fuzzing applies the score policy for the Fuzzing check.
func Fuzzing(name string, dl checker.DetailLogger,
	r *checker.FuzzingData,
) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	// Keep track of reported results.
	reportedRuleResults := map[string]bool{
		"FuzzingWithGoNative": false,
	}

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
		for j := range fuzzer.Files {
			f := fuzzer.Files[j]
			loc := checker.LocationFromPath(&f)
			switch fuzzer.Name {
			case "GoBuiltInFuzzer":
				if err := checker.LogFinding(rules, "FuzzingWithGoNative",
					"project fuzzed with Go native framework",
					loc, finding.OutcomePositive, dl); err != nil {
					return checker.CreateRuntimeErrorResult(name, err)
				}
				reportedRuleResults["FuzzingWithGoNative"] = true
			default:
				return checker.CreateRuntimeErrorResult(name, fmt.Errorf("unsupported fuzzer: %v", fuzzer))
			}
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
	if !r["FuzzingWithGoNative"] {
		if err := checker.LogFinding(rules, "FuzzingWithGoNative",
			"no fuzzing using Go native framework",
			nil, finding.OutcomeNegative, dl); err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
	}

	return nil
}
