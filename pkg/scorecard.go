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
	"github.com/ossf/scorecard/roundtripper"
	"github.com/shurcooL/githubv4"
	"go.uber.org/zap"
)

func runEnabledChecks(ctx context.Context, logger *zap.SugaredLogger, repo repos.RepoURL,
	checksToRun checker.CheckNameToFnMap, resultsCh chan checker.CheckResult) {
	wg := sync.WaitGroup{}
	for _, checkFn := range checksToRun {
		checkFn := checkFn
		wg.Add(1)
		go func() {
			// Use our custom roundtripper
			rt := roundtripper.NewTransport(ctx, logger)

			client := &http.Client{
				Transport: rt,
			}
			ghClient := github.NewClient(client)
			graphClient := githubv4.NewClient(client)

			c := checker.CheckRequest{
				Ctx:         ctx,
				Client:      ghClient,
				HttpClient:  client,
				Owner:       repo.Owner,
				Repo:        repo.Repo,
				GraphClient: graphClient,
			}
			defer wg.Done()
			runner := checker.Runner{CheckRequest: c}
			resultsCh <- runner.Run(checkFn)
		}()
	}
	wg.Wait()
	close(resultsCh)
}

func RunScorecards(ctx context.Context, logger *zap.SugaredLogger,
	repo repos.RepoURL, checksToRun checker.CheckNameToFnMap) repos.RepoResult {
	ret := repos.RepoResult{
		Repo: repo.Url(),
		Date: time.Now().Format("2006-01-02"),
	}
	resultsCh := make(chan checker.CheckResult)
	go runEnabledChecks(ctx, logger, repo, checksToRun, resultsCh)
	for result := range resultsCh {
		ret.Checks = append(ret.Checks, result)
	}
	return ret
}
