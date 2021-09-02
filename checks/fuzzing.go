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
	"sync"

	"go.uber.org/zap"

	"github.com/ossf/scorecard/v2/checker"
	"github.com/ossf/scorecard/v2/clients"
	"github.com/ossf/scorecard/v2/clients/githubrepo"
	sce "github.com/ossf/scorecard/v2/errors"
)

// CheckFuzzing is the registered name for Fuzzing.
const CheckFuzzing = "Fuzzing"

var (
	ossFuzzRepo    clients.RepoClient
	errOssFuzzRepo error
	logger         *zap.Logger
	once           sync.Once
)

//nolint:gochecknoinits
func init() {
	registerCheck(CheckFuzzing, Fuzzing)
}

// Fuzzing runs Fuzzing check.
func Fuzzing(c *checker.CheckRequest) checker.CheckResult {
	once.Do(func() {
		logger, errOssFuzzRepo = githubrepo.NewLogger(zap.InfoLevel)
		if errOssFuzzRepo != nil {
			return
		}
		ossFuzzRepo = githubrepo.CreateGithubRepoClient(c.Ctx, logger)
		errOssFuzzRepo = ossFuzzRepo.InitRepo("google", "oss-fuzz")
	})
	if errOssFuzzRepo != nil {
		e := sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("InitRepo: %v", errOssFuzzRepo))
		return checker.CreateRuntimeErrorResult(CheckFuzzing, e)
	}

	req := clients.SearchRequest{
		Query:    c.RepoClient.URL(),
		Filename: "project.yaml",
	}
	result, err := ossFuzzRepo.Search(req)
	if err != nil {
		e := sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("Client.Search.Code: %v", err))
		return checker.CreateRuntimeErrorResult(CheckFuzzing, e)
	}

	if result.Hits > 0 {
		return checker.CreateMaxScoreResult(CheckFuzzing,
			"project is fuzzed in OSS-Fuzz")
	}

	return checker.CreateMinScoreResult(CheckFuzzing, "project is not fuzzed in OSS-Fuzz")
}
