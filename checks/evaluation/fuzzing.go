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
	"github.com/ossf/scorecard/v4/checker"
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

	for _, fuzzer := range r.Fuzzers {
		switch fuzzer.Name {
		case checker.FuzzerNameCIFuzz:
			return checker.CreateMaxScoreResult(name, "project uses ClusterFuzzLite")
		case checker.FuzzerNameOSSFuzz:
			return checker.CreateMaxScoreResult(name, "project is fuzzed in OSS-Fuzz")
		case checker.FuzzerNameGoBuiltin:
			return checker.CreateMaxScoreResult(name, "project is fuzzed using Golang's fuzzing")
		}
	}

	return checker.CreateMinScoreResult(name, "project is not fuzzed")
}
