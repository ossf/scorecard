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
	"path"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/raw"
	sce "github.com/ossf/scorecard/v4/errors"
)

// Fuzzing applies the score policy for the Fuzzing check.
func Fuzzing(name string, dl checker.DetailLogger,
	r *checker.FuzzingData,
) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	for i := range r.Fuzzers {
		fuzzer := r.Fuzzers[i]
		if fuzzer.Name == raw.FuzzNameUserDefinedFunc {
			for _, f := range fuzzer.File {
				msg := checker.LogMessage{
					Path:   path.Join(f.Path, f.Snippet),
					Type:   f.Type,
					Offset: f.Offset,
				}
				dl.Info(&msg)
			}
			score := int(checker.MaxResultScore * fuzzer.LanguageCoverage)
			return checker.CreateResultWithScore(
				name,
				fmt.Sprintf("project is fuzzed by %s, with a language coverage of %.2f", fuzzer.Name, fuzzer.LanguageCoverage),
				score,
			)
		} else {
			// Otherwise, the fuzzer is either OSS-Fuzz or CFL
			return checker.CreateMaxScoreResult(name,
				fmt.Sprintf("project is fuzzed with %s", fuzzer.Name))
		}
	}

	return checker.CreateMinScoreResult(name, "project is not fuzzed")
}
