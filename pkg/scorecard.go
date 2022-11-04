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

// Package pkg defines fns for running Scorecard checks on a Repo.
package pkg

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"sigs.k8s.io/release-utils/version"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

func runEnabledChecks(ctx context.Context,
	repo clients.Repo, raw *checker.RawResults, checksToRun checker.CheckNameToFnMap,
	repoClient clients.RepoClient, ossFuzzRepoClient clients.RepoClient, ciiClient clients.CIIBestPracticesClient,
	vulnsClient clients.VulnerabilitiesClient,
	resultsCh chan checker.CheckResult,
) {
	request := checker.CheckRequest{
		Ctx:                   ctx,
		RepoClient:            repoClient,
		OssFuzzRepo:           ossFuzzRepoClient,
		CIIClient:             ciiClient,
		VulnerabilitiesClient: vulnsClient,
		Repo:                  repo,
		RawResults:            raw,
	}
	wg := sync.WaitGroup{}
	for checkName, checkFn := range checksToRun {
		checkName := checkName
		checkFn := checkFn
		wg.Add(1)
		go func() {
			defer wg.Done()
			runner := checker.NewRunner(
				checkName,
				repo.URI(),
				&request,
			)

			resultsCh <- runner.Run(ctx, checkFn)
		}()
	}
	wg.Wait()
	close(resultsCh)
}

func getRepoCommitHash(r clients.RepoClient) (string, error) {
	commits, err := r.ListCommits()
	if err != nil {
		// allow --local repos to still process
		if errors.Is(err, clients.ErrUnsupportedFeature) {
			return "unknown", nil
		}
		return "", sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("ListCommits:%v", err.Error()))
	}

	if len(commits) > 0 {
		return commits[0].SHA, nil
	}
	return "", nil
}

// RunScorecard runs enabled Scorecard checks on a Repo.
func RunScorecard(ctx context.Context,
	repo clients.Repo,
	commitSHA string,
	commitDepth int,
	checksToRun checker.CheckNameToFnMap,
	repoClient clients.RepoClient,
	ossFuzzRepoClient clients.RepoClient,
	ciiClient clients.CIIBestPracticesClient,
	vulnsClient clients.VulnerabilitiesClient,
) (ScorecardResult, error) {
	if err := repoClient.InitRepo(repo, commitSHA, commitDepth); err != nil {
		// No need to call sce.WithMessage() since InitRepo will do that for us.
		//nolint:wrapcheck
		return ScorecardResult{}, err
	}
	defer repoClient.Close()

	commitSHA, err := getRepoCommitHash(repoClient)
	if err != nil || commitSHA == "" {
		return ScorecardResult{}, err
	}
	versionInfo := version.GetVersionInfo()
	ret := ScorecardResult{
		Repo: RepoInfo{
			Name:      repo.URI(),
			CommitSHA: commitSHA,
		},
		Scorecard: ScorecardInfo{
			Version:   versionInfo.GitVersion,
			CommitSHA: versionInfo.GitCommit,
		},
		Date: time.Now(),
	}
	resultsCh := make(chan checker.CheckResult)
	go runEnabledChecks(ctx, repo, &ret.RawResults, checksToRun, repoClient, ossFuzzRepoClient,
		ciiClient, vulnsClient, resultsCh)

	for result := range resultsCh {
		ret.Checks = append(ret.Checks, result)
	}
	return ret, nil
}
