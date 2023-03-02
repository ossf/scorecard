// Copyright 2021 OpenSSF Scorecard Authors
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

package githubrepo

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/google/go-github/v38/github"
	"github.com/shurcooL/githubv4"

	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/log"
)

//nolint:govet
type checkRunsGraphqlData struct {
	Repository struct {
		Object struct {
			Commit struct {
				History struct {
					Nodes []struct {
						AssociatedPullRequests struct {
							Nodes []struct {
								HeadRefOid githubv4.String
								Commits    struct {
									Nodes []struct {
										Commit struct {
											CheckSuites struct {
												Nodes []struct {
													App struct {
														Slug githubv4.String
													}
													Conclusion githubv4.CheckConclusionState
													Status     githubv4.CheckStatusState
												}
											} `graphql:"checkSuites(first: $checksToAnalyze)"`
										}
									}
								} `graphql:"commits(last:1)"`
							}
						} `graphql:"associatedPullRequests(first: $pullRequestsToAnalyze)"`
					}
				} `graphql:"history(first: $commitsToAnalyze)"`
			} `graphql:"... on Commit"`
		} `graphql:"object(expression: $commitExpression)"`
	} `graphql:"repository(owner: $owner, name: $name)"`
	RateLimit struct {
		Cost *int
	}
}

type checkRunsByRef = map[string][]clients.CheckRun

// nolint: govet
type checkrunsHandler struct {
	client         *github.Client
	graphClient    *githubv4.Client
	repourl        *repoURL
	logger         *log.Logger
	checkData      *checkRunsGraphqlData
	setupOnce      *sync.Once
	ctx            context.Context
	commitDepth    int
	checkRunsByRef checkRunsByRef
	errSetup       error
}

func (handler *checkrunsHandler) init(ctx context.Context, repourl *repoURL, commitDepth int) {
	handler.ctx = ctx
	handler.repourl = repourl
	handler.commitDepth = commitDepth
	handler.logger = log.NewLogger(log.DefaultLevel)
	handler.checkData = new(checkRunsGraphqlData)
	handler.setupOnce = new(sync.Once)
	handler.checkRunsByRef = checkRunsByRef{}
}

func (handler *checkrunsHandler) setup() error {
	handler.setupOnce.Do(func() {
		commitExpression := handler.repourl.commitExpression()
		vars := map[string]interface{}{
			"owner":                 githubv4.String(handler.repourl.owner),
			"name":                  githubv4.String(handler.repourl.repo),
			"pullRequestsToAnalyze": githubv4.Int(pullRequestsToAnalyze),
			"commitsToAnalyze":      githubv4.Int(handler.commitDepth),
			"commitExpression":      githubv4.String(commitExpression),
			"checksToAnalyze":       githubv4.Int(checksToAnalyze),
		}
		// TODO(#2224):
		// sast and ci checks causes cache miss if commits dont match number of check runs.
		// paging for this needs to be implemented if using higher than 100 --number-of-commits
		if handler.commitDepth > 99 {
			vars["commitsToAnalyze"] = githubv4.Int(99)
		}
		if err := handler.graphClient.Query(handler.ctx, handler.checkData, vars); err != nil {
			// quit early without setting crsErrSetup for "Resource not accessible by integration" error
			// for whatever reason, this check doesn't work with a GITHUB_TOKEN, only a PAT
			if strings.Contains(err.Error(), "Resource not accessible by integration") {
				return
			}
			handler.errSetup = err
			return
		}
		handler.checkRunsByRef = parseCheckRuns(handler.checkData)
	})
	return handler.errSetup
}

func (handler *checkrunsHandler) listCheckRunsForRef(ref string) ([]clients.CheckRun, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during graphqlHandler.setupCheckRuns: %w", err)
	}
	if crs, ok := handler.checkRunsByRef[ref]; ok {
		return crs, nil
	}
	msg := fmt.Sprintf("listCheckRunsForRef cache miss: %s/%s:%s", handler.repourl.owner, handler.repourl.repo, ref)
	handler.logger.Info(msg)

	checkRuns, _, err := handler.client.Checks.ListCheckRunsForRef(
		handler.ctx, handler.repourl.owner, handler.repourl.repo, ref, &github.ListCheckRunsOptions{})
	if err != nil {
		return nil, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("ListCheckRunsForRef: %v", err))
	}
	handler.checkRunsByRef[ref] = checkRunsFrom(checkRuns)
	return handler.checkRunsByRef[ref], nil
}

func parseCheckRuns(data *checkRunsGraphqlData) checkRunsByRef {
	checkCache := checkRunsByRef{}
	for _, commit := range data.Repository.Object.Commit.History.Nodes {
		for _, pr := range commit.AssociatedPullRequests.Nodes {
			var crs []clients.CheckRun
			for _, c := range pr.Commits.Nodes {
				for _, checkRun := range c.Commit.CheckSuites.Nodes {
					crs = append(crs, clients.CheckRun{
						// the REST API returns lowercase. the graphQL API returns upper
						Status:     strings.ToLower(string(checkRun.Status)),
						Conclusion: strings.ToLower(string(checkRun.Conclusion)),
						App: clients.CheckRunApp{
							Slug: string(checkRun.App.Slug),
						},
					})
				}
			}
			headRef := string(pr.HeadRefOid)
			checkCache[headRef] = crs
		}
	}
	return checkCache
}

func checkRunsFrom(data *github.ListCheckRunsResults) []clients.CheckRun {
	var checkRuns []clients.CheckRun
	for _, checkRun := range data.CheckRuns {
		checkRuns = append(checkRuns, clients.CheckRun{
			Status:     checkRun.GetStatus(),
			Conclusion: checkRun.GetConclusion(),
			URL:        checkRun.GetURL(),
			App: clients.CheckRunApp{
				Slug: checkRun.GetApp().GetSlug(),
			},
		})
	}
	return checkRuns
}
