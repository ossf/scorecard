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

	"github.com/google/go-github/v32/github"
	"github.com/ossf/scorecard/checker"
	"github.com/ossf/scorecard/repos"
	"github.com/ossf/scorecard/roundtripper"
	"github.com/shurcooL/githubv4"
	"go.uber.org/zap"
)

func runEnabledChecks(ctx context.Context, logger *zap.SugaredLogger, repo repos.RepoURL,
	enabledChecks checker.CheckNameToFnMap) <-chan checker.CheckResult {
	resultsCh := make(chan checker.CheckResult)
	wg := sync.WaitGroup{}
	for _, checkFn := range enabledChecks {
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
	go func() {
		wg.Wait()
		close(resultsCh)
	}()
	return resultsCh
}

func RunScorecards(ctx context.Context, logger *zap.SugaredLogger,
	repoRequest repos.RepoRequest) repos.RepoResult {
	ret := repos.RepoResult{
		Repo: repoRequest.Repo.Url(),
		Date: "",
	}
	resultsCh := runEnabledChecks(ctx, logger, repoRequest.Repo, repoRequest.EnabledChecks)
	for result := range resultsCh {
		ret.CheckResults = append(ret.CheckResults, result)
	}
	return ret
}
