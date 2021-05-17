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

package pkg

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/ossf/scorecard/checker"
	"github.com/ossf/scorecard/repos"
	"github.com/shurcooL/githubv4"
)

func runEnabledChecks(ctx context.Context,
	repo repos.RepoURL, checksToRun checker.CheckNameToFnMap,
	httpClient *http.Client, githubClient *github.Client, graphClient *githubv4.Client,
	resultsCh chan checker.CheckResult) {
	request := checker.CheckRequest{
		Ctx:         ctx,
		Client:      githubClient,
		HTTPClient:  httpClient,
		Owner:       repo.Owner,
		Repo:        repo.Repo,
		GraphClient: graphClient,
	}
	wg := sync.WaitGroup{}
	for _, checkFn := range checksToRun {
		checkFn := checkFn
		wg.Add(1)
		go func() {
			defer wg.Done()
			runner := checker.Runner{CheckRequest: request}
			resultsCh <- runner.Run(checkFn)
		}()
	}
	wg.Wait()
	close(resultsCh)
}

func RunScorecards(ctx context.Context,
	repo repos.RepoURL,
	checksToRun checker.CheckNameToFnMap,
	httpClient *http.Client,
	githubClient *github.Client,
	graphClient *githubv4.Client) repos.RepoResult {
	ret := repos.RepoResult{
		Repo: repo.URL(),
		Date: time.Now().Format("2006-01-02"),
	}
	resultsCh := make(chan checker.CheckResult)
	go runEnabledChecks(ctx, repo, checksToRun,
		httpClient, githubClient, graphClient,
		resultsCh)
	for result := range resultsCh {
		ret.Checks = append(ret.Checks, result)
	}
	return ret
}
