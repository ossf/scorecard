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

// Package pkg defines fns for running Scorecard checks on a RepoURL.
package pkg

import (
	"context"
	"fmt"
	"sync"
	"time"

	opencensusstats "go.opencensus.io/stats"
	"go.opencensus.io/tag"

	"github.com/ossf/scorecard/v3/checker"
	"github.com/ossf/scorecard/v3/clients"
	"github.com/ossf/scorecard/v3/clients/githubrepo"
	sce "github.com/ossf/scorecard/v3/errors"
	"github.com/ossf/scorecard/v3/repos"
	"github.com/ossf/scorecard/v3/stats"
)

func logStats(ctx context.Context, startTime time.Time) {
	runTimeInSecs := time.Now().Unix() - startTime.Unix()
	opencensusstats.Record(ctx, stats.RepoRuntimeInSec.M(runTimeInSecs))
}

func runEnabledChecks(ctx context.Context,
	repo clients.Repo, checksToRun checker.CheckNameToFnMap, repoClient clients.RepoClient,
	resultsCh chan checker.CheckResult) {
	request := checker.CheckRequest{
		Ctx:        ctx,
		RepoClient: repoClient,
		Repo:       repo,
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

// RunScorecards runs enabled Scorecard checks on a RepoURL.
func RunScorecards(ctx context.Context,
	repoURL repos.RepoURL,
	checksToRun checker.CheckNameToFnMap,
	repoClient clients.RepoClient) (ScorecardResult, error) {
	ctx, err := tag.New(ctx, tag.Upsert(stats.Repo, repoURL.URL()))
	if err != nil {
		return ScorecardResult{}, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("tag.New: %v", err))
	}
	defer logStats(ctx, time.Now())

	repo, err := githubrepo.MakeGithubRepo(repoURL.URL())
	if err != nil {
		return ScorecardResult{}, sce.WithMessage(err, "")
	}
	if err := repoClient.InitRepo(repo); err != nil {
		// No need to call sce.WithMessage() since InitRepo will do that for us.
		//nolint:wrapcheck
		return ScorecardResult{}, err
	}
	defer repoClient.Close()

	commits, err := repoClient.ListCommits()
	if err != nil {
		// nolint:wrapcheck
		return ScorecardResult{}, err
	}
	var commitSHA string
	if len(commits) > 0 {
		commitSHA = commits[0].SHA
	} else {
		commitSHA = "no commits found"
	}

	ret := ScorecardResult{
		Repo: RepoInfo{
			Name:      repo.URL(),
			CommitSHA: commitSHA,
		},
		Scorecard: ScorecardInfo{
			Version:   GetSemanticVersion(),
			CommitSHA: GetCommit(),
		},
		Date: time.Now(),
	}
	resultsCh := make(chan checker.CheckResult)
	go runEnabledChecks(ctx, repo, checksToRun, repoClient, resultsCh)
	for result := range resultsCh {
		ret.Checks = append(ret.Checks, result)
	}
	return ret, nil
}
