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

	opencensusstats "go.opencensus.io/stats"
	"go.opencensus.io/tag"

<<<<<<< HEAD
	"github.com/ossf/scorecard/v3/checker"
	"github.com/ossf/scorecard/v3/clients"
	"github.com/ossf/scorecard/v3/clients/githubrepo"
	"github.com/ossf/scorecard/v3/clients/localdir"
	sce "github.com/ossf/scorecard/v3/errors"
	"github.com/ossf/scorecard/v3/repos"
	"github.com/ossf/scorecard/v3/stats"
=======
	"github.com/ossf/scorecard/v2/checker"
	"github.com/ossf/scorecard/v2/clients"
	"github.com/ossf/scorecard/v2/clients/githubrepo"
	"github.com/ossf/scorecard/v2/clients/localdir"
	sce "github.com/ossf/scorecard/v2/errors"
	"github.com/ossf/scorecard/v2/repos"
	"github.com/ossf/scorecard/v2/stats"
>>>>>>> 22b8b74 (draft)
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

<<<<<<< HEAD
<<<<<<< HEAD
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
		//nolint:unwrapped
		commits, err := r.ListCommits()
		if err != nil {
			// nolint:wrapcheck
			return "", err
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

=======
>>>>>>> 6c86056 (draft)
=======
func createRepo(uri *repos.RepoURI) (clients.Repo, error) {
	switch uri.GetType() {
	// URL.
	case repos.RepoTypeURL:
		//nolint:wrapcheck
		return githubrepo.MakeGithubRepo(uri.GetURL())
	// LocalDir.
	case repos.RepoTypeLocalDir:
		//nolint:wrapcheck
		return localdir.MakeLocalDirRepo(uri.GetPath())
	default:
		return nil,
			sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("unsupported URI type:%v", uri.GetType()))
	}
}

<<<<<<< HEAD
>>>>>>> 22b8b74 (draft)
=======
func getRepoCommitHash(r clients.RepoClient, uri *repos.RepoURI) (string, error) {
	switch uri.GetType() {
	// URL.
	case repos.RepoTypeURL:
		//nolint:unwrapped
		commits, err := r.ListCommits()
		if err != nil {
			// nolint:wrapcheck
			return "", err
		}

		if len(commits) > 0 {
			return commits[0].SHA, nil
		} else {
			return "no commits found", nil
		}

	// LocalDir.
	case repos.RepoTypeLocalDir:
		return "no commits for directory repo", nil

	default:
		return "",
			sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("unsupported URI type:%v", uri.GetType()))
	}
}

>>>>>>> 376995a (docker file)
// RunScorecards runs enabled Scorecard checks on a Repo.
func RunScorecards(ctx context.Context,
	repoURI *repos.RepoURI,
	checksToRun checker.CheckNameToFnMap,
	repoClient clients.RepoClient) (ScorecardResult, error) {
<<<<<<< HEAD
	ctx, err := tag.New(ctx, tag.Upsert(stats.Repo, repoURI.URL()))
=======
	ctx, err := tag.New(ctx, tag.Upsert(stats.Repo, repoURI.GetURL()))
>>>>>>> 6c86056 (draft)
	if err != nil {
		return ScorecardResult{}, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("tag.New: %v", err))
	}
	defer logStats(ctx, time.Now())

<<<<<<< HEAD
<<<<<<< HEAD
	repo, err := createRepo(repoURI)
=======
	// TODO: get type.
	repo, err := githubrepo.MakeGithubRepo(repoURI.GetURL())
>>>>>>> 6c86056 (draft)
=======
	repo, err := createRepo(repoURI)
>>>>>>> 22b8b74 (draft)
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
