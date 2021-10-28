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

	"github.com/ossf/scorecard/v3/checker"
	"github.com/ossf/scorecard/v3/clients"
	"github.com/ossf/scorecard/v3/clients/githubrepo"
	sce "github.com/ossf/scorecard/v3/errors"
)

// CheckFuzzing is the registered name for Fuzzing.
const CheckFuzzing = "Fuzzing"

var (
	ossFuzzRepo       clients.Repo
	ossFuzzRepoClient clients.RepoClient
	errOssFuzzRepo    error
	logger            *zap.Logger
	once              sync.Once
)

//nolint:gochecknoinits
func init() {
	registerCheck(CheckFuzzing, Fuzzing)
}

func checkCFLite(c *checker.CheckRequest) (bool, error) {
	result := false
	e := CheckFilesContent(".clusterfuzzlite/Dockerfile", true, c,
		func(path string, content []byte, dl checker.DetailLogger, data FileCbData) (bool, error) {
			result = CheckFileContainsCommands(content, "#")
			return false, nil
		}, nil)

	return result, e
}

func checkOSSFuzz(c *checker.CheckRequest) (bool, error) {
	once.Do(func() {
		logger, errOssFuzzRepo = githubrepo.NewLogger(zap.InfoLevel)
		if errOssFuzzRepo != nil {
			return
		}
		ossFuzzRepo, errOssFuzzRepo = githubrepo.MakeGithubRepo("google/oss-fuzz")
		if errOssFuzzRepo != nil {
			return
		}

		ossFuzzRepoClient = githubrepo.CreateGithubRepoClient(c.Ctx, logger)
		errOssFuzzRepo = ossFuzzRepoClient.InitRepo(ossFuzzRepo)
	})
	if errOssFuzzRepo != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("InitRepo: %v", errOssFuzzRepo))
		return false, e
	}

	req := clients.SearchRequest{
		Query:    c.RepoClient.URI(),
		Filename: "project.yaml",
	}
	result, err := ossFuzzRepoClient.Search(req)
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
