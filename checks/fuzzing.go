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

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

// CheckFuzzing is the registered name for Fuzzing.
const CheckFuzzing = "Fuzzing"

//nolint:gochecknoinits
func init() {
	registerCheck(CheckFuzzing, Fuzzing)
}

func checkCFLite(c *checker.CheckRequest) (bool, error) {
	result := false
	e := fileparser.CheckFilesContent(".clusterfuzzlite/Dockerfile", true, c,
		func(path string, content []byte, dl checker.DetailLogger, data fileparser.FileCbData) (bool, error) {
			result = fileparser.CheckFileContainsCommands(content, "#")
			return false, nil
		}, nil)
	if e != nil {
		return result, fmt.Errorf("%w", e)
	}
	return result, nil
}

func checkOSSFuzz(c *checker.CheckRequest) (bool, error) {
	if c.OssFuzzRepo == nil {
		return false, nil
	}

	req := clients.SearchRequest{
		Query:    c.RepoClient.URI(),
		Filename: "project.yaml",
	}
	result, err := c.OssFuzzRepo.Search(req)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("Client.Search.Code: %v", err))
		return false, e
	}
	return result.Hits > 0, nil
}

// Fuzzing runs Fuzzing check.
func Fuzzing(c *checker.CheckRequest) checker.CheckResult {
	usingCFLite, e := checkCFLite(c)
	if e != nil {
		return checker.CreateRuntimeErrorResult(CheckFuzzing, e)
	}
	if usingCFLite {
		return checker.CreateMaxScoreResult(CheckFuzzing,
			"project uses ClusterFuzzLite")
	}

	usingOSSFuzz, e := checkOSSFuzz(c)
	if e != nil {
		return checker.CreateRuntimeErrorResult(CheckFuzzing, e)
	}
	if usingOSSFuzz {
		return checker.CreateMaxScoreResult(CheckFuzzing,
			"project is fuzzed in OSS-Fuzz")
	}

	return checker.CreateMinScoreResult(CheckFuzzing, "project is not fuzzed")
}
