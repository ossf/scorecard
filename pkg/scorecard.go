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
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/shurcooL/githubv4"
	opencensusstats "go.opencensus.io/stats"
	"go.opencensus.io/tag"

	"github.com/ossf/scorecard/checker"
	"github.com/ossf/scorecard/clients"
	"github.com/ossf/scorecard/repos"
	"github.com/ossf/scorecard/stats"
)

func logStats(ctx context.Context, startTime time.Time) {
	runTimeInSecs := time.Now().Unix() - startTime.Unix()
	opencensusstats.Record(ctx, stats.RepoRuntimeInSec.M(runTimeInSecs))
}

func runEnabledChecks(ctx context.Context,
	repo repos.RepoURL, checksToRun checker.CheckNameToFnMap, repoClient clients.RepoClient,
	httpClient *http.Client, githubClient *github.Client, graphClient *githubv4.Client,
	resultsCh chan checker.CheckResult) {
	request := checker.CheckRequest{
		Ctx:         ctx,
		Client:      githubClient,
		RepoClient:  repoClient,
		HTTPClient:  httpClient,
		Owner:       repo.Owner,
		Repo:        repo.Repo,
		GraphClient: graphClient,
	}
	wg := sync.WaitGroup{}
	for checkName, checkFn := range checksToRun {
		checkName := checkName
		checkFn := checkFn
		wg.Add(1)
		go func() {
			defer wg.Done()
			runner := checker.Runner{
				Repo:         repo.URL(),
				CheckName:    checkName,
				CheckRequest: request,
			}
			resultsCh <- runner.Run(ctx, checkFn)
		}()
	}
	wg.Wait()
	close(resultsCh)
}

func RunScorecards(ctx context.Context,
	repo repos.RepoURL,
	checksToRun checker.CheckNameToFnMap,
	repoClient clients.RepoClient,
	httpClient *http.Client,
	githubClient *github.Client,
	graphClient *githubv4.Client) (repos.RepoResult, error) {
	ctx, err := tag.New(ctx, tag.Upsert(stats.Repo, repo.URL()))
	if err != nil {
		return repos.RepoResult{}, fmt.Errorf("error during tag.New: %w", err)
	}
	defer logStats(ctx, time.Now())

	if err := repoClient.InitRepo(repo.Owner, repo.Repo); err != nil {
		return repos.RepoResult{}, fmt.Errorf("error during InitRepo for %s: %w", repo.URL(), err)
	}

	ret := repos.RepoResult{
		Repo: repo.URL(),
		Date: time.Now().Format("2006-01-02"),
	}
	resultsCh := make(chan checker.CheckResult)
	go runEnabledChecks(ctx, repo, checksToRun, repoClient,
		httpClient, githubClient, graphClient,
		resultsCh)
	for result := range resultsCh {
		ret.Checks = append(ret.Checks, result)
	}
	return ret, nil
}
