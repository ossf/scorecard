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

package checks

import (
	"fmt"

	"github.com/google/go-github/v38/github"

	"github.com/ossf/scorecard/v2/checker"
	sce "github.com/ossf/scorecard/v2/errors"
)

// CheckFuzzing is the registered name for Fuzzing.
const CheckFuzzing = "Fuzzing"

//nolint:gochecknoinits
func init() {
	registerCheck(CheckFuzzing, Fuzzing)
}

// Fuzzing runs Fuzzing check.
func Fuzzing(c *checker.CheckRequest) checker.CheckResult {
	url := fmt.Sprintf("github.com/%s/%s", c.Owner, c.Repo)
	searchString := url + " repo:google/oss-fuzz in:file filename:project.yaml"
	results, _, err := c.Client.Search.Code(c.Ctx, searchString, &github.SearchOptions{})
	if err != nil {
		e := sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("Client.Search.Code: %v", err))
		return checker.CreateRuntimeErrorResult(CheckFuzzing, e)
	}

	if *results.Total > 0 {
		return checker.CreateMaxScoreResult(CheckFuzzing,
			"project is fuzzed in OSS-Fuzz")
	}

	return checker.CreateMinScoreResult(CheckFuzzing, "project is not fuzzed in OSS-Fuzz")
}
