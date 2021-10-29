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
	"fmt"
	"sync"
	"time"

	"github.com/ossf/scorecard/v3/checker"
	"github.com/ossf/scorecard/v3/clients"
	"github.com/ossf/scorecard/v3/clients/githubrepo"
	"github.com/ossf/scorecard/v3/clients/localdir"
	sce "github.com/ossf/scorecard/v3/errors"
	"github.com/ossf/scorecard/v3/repos"
)

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
				Repo:         repo.URI(),
				CheckName:    checkName,
				CheckRequest: request,
			}
			resultsCh <- runner.Run(ctx, checkFn)
		}()
	}
	wg.Wait()
	close(resultsCh)
}

func createRepo(uri *repos.RepoURI) (clients.Repo, error) {
	var c clients.Repo
	var e error
	switch uri.RepoType() {
	// URL.
	case repos.RepoTypeURL:
		c, e = githubrepo.MakeGithubRepo(uri.URL())
	// LocalDir.
	case repos.RepoTypeLocalDir:
		c, e = localdir.MakeLocalDirRepo(uri.Path())
	default:
		return nil,
			sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("unsupported URI type:%v", uri.RepoType()))
	}

	if e != nil {
		return c, sce.WithMessage(sce.ErrScorecardInternal, e.Error())
	}

	return c, nil
}

func getRepoCommitHash(r clients.RepoClient, uri *repos.RepoURI) (string, error) {
	switch uri.RepoType() {
	// URL.
	case repos.RepoTypeURL:
		commits, err := r.ListCommits()
		if err != nil {
			return "", sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("ListCommits:%v", err.Error()))
		}

		if len(commits) > 0 {
			return commits[0].SHA, nil
		}

		return "no commits found", nil

	// LocalDir.
	case repos.RepoTypeLocalDir:
		return "no commits for directory repo", nil

	default:
		return "",
			sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("unsupported URI type:%v", uri.RepoType()))
	}
}

// RunScorecards runs enabled Scorecard checks on a Repo.
func RunScorecards(ctx context.Context,
	repoURI *repos.RepoURI,
	checksToRun checker.CheckNameToFnMap,
	repoClient clients.RepoClient) (ScorecardResult, error) {
	repo, err := createRepo(repoURI)
	if err != nil {
		return ScorecardResult{}, sce.WithMessage(err, "")
	}
	if err := repoClient.InitRepo(repo); err != nil {
		// No need to call sce.WithMessage() since InitRepo will do that for us.
		//nolint:wrapcheck
		return ScorecardResult{}, err
	}
	defer repoClient.Close()

	commitSHA, err := getRepoCommitHash(repoClient, repoURI)
	if err != nil {
		return ScorecardResult{}, err
	}

	ret := ScorecardResult{
		Repo: RepoInfo{
			Name:      repo.URI(),
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
