// Copyright 2021 Security Scorecard Authors
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

	"github.com/shurcooL/githubv4"

	"github.com/ossf/scorecard/v2/clients"
	sce "github.com/ossf/scorecard/v2/errors"
)

const (
	pullRequestsToAnalyze = 30
	reviewsToAnalyze      = 30
	labelsToAnalyze       = 30
	commitsToAnalyze      = 30
)

// nolint: govet
type graphqlData struct {
	Repository struct {
		IsArchived       githubv4.Boolean
		DefaultBranchRef struct {
			Name                 githubv4.String
			BranchProtectionRule struct {
				RequiredApprovingReviewCount githubv4.Int
			}
			Target struct {
				Commit struct {
					History struct {
						Nodes []struct {
							CommittedDate githubv4.DateTime
							Message       githubv4.String
							Oid           githubv4.GitObjectID
							Committer     struct {
								User struct {
									Login githubv4.String
								}
							}
						}
					} `graphql:"history(first: $commitsToAnalyze)"`
				} `graphql:"... on Commit"`
			}
		}
		PullRequests struct {
			Nodes []struct {
				Number      githubv4.Int
				MergeCommit struct {
					AuthoredByCommitter githubv4.Boolean
				}
				MergedAt githubv4.DateTime
				Labels   struct {
					Nodes []struct {
						Name githubv4.String
					}
				} `graphql:"labels(last: $labelsToAnalyze)"`
				LatestReviews struct {
					Nodes []struct {
						State githubv4.String
					}
				} `graphql:"latestReviews(last: $reviewsToAnalyze)"`
			}
		} `graphql:"pullRequests(last: $pullRequestsToAnalyze, states: MERGED)"`
	} `graphql:"repository(owner: $owner, name: $name)"`
}

type graphqlHandler struct {
	client           *githubv4.Client
	data             *graphqlData
	prs              []clients.PullRequest
	commits          []clients.Commit
	defaultBranchRef clients.BranchRef
	archived         bool
}

func (handler *graphqlHandler) init(ctx context.Context, owner, repo string) error {
	vars := map[string]interface{}{
		"owner":                 githubv4.String(owner),
		"name":                  githubv4.String(repo),
		"pullRequestsToAnalyze": githubv4.Int(pullRequestsToAnalyze),
		"reviewsToAnalyze":      githubv4.Int(reviewsToAnalyze),
		"labelsToAnalyze":       githubv4.Int(labelsToAnalyze),
		"commitsToAnalyze":      githubv4.Int(commitsToAnalyze),
	}
	handler.data = new(graphqlData)
	if err := handler.client.Query(ctx, handler.data, vars); err != nil {
		// nolint: wrapcheck
		return sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("githubv4.Query: %v", err))
	}
	handler.archived = bool(handler.data.Repository.IsArchived)
	handler.prs = pullRequestFrom(handler.data)
	handler.defaultBranchRef = defaultBranchRefFrom(handler.data)
	handler.commits = commitsFrom(handler.data)
	return nil
}

func (handler *graphqlHandler) getMergedPRs() ([]clients.PullRequest, error) {
	return handler.prs, nil
}

func (handler *graphqlHandler) getDefaultBranch() (clients.BranchRef, error) {
	return handler.defaultBranchRef, nil
}

func (handler *graphqlHandler) getCommits() ([]clients.Commit, error) {
	return handler.commits, nil
}

func (handler *graphqlHandler) isArchived() (bool, error) {
	return handler.archived, nil
}

func pullRequestFrom(data *graphqlData) []clients.PullRequest {
	ret := make([]clients.PullRequest, len(data.Repository.PullRequests.Nodes))
	for i, pr := range data.Repository.PullRequests.Nodes {
		toAppend := clients.PullRequest{
			Number:   int(pr.Number),
			MergedAt: pr.MergedAt.Time,
			MergeCommit: clients.Commit{
				AuthoredByCommitter: bool(pr.MergeCommit.AuthoredByCommitter),
			},
		}
		for _, label := range pr.Labels.Nodes {
			toAppend.Labels = append(toAppend.Labels, clients.Label{
				Name: string(label.Name),
			})
		}
		for _, review := range pr.LatestReviews.Nodes {
			toAppend.Reviews = append(toAppend.Reviews, clients.Review{
				State: string(review.State),
			})
		}
		ret[i] = toAppend
	}
	return ret
}

func defaultBranchRefFrom(data *graphqlData) clients.BranchRef {
	return clients.BranchRef{
		Name: string(data.Repository.DefaultBranchRef.Name),
		BranchProtectionRule: clients.BranchProtectionRule{
			RequiredApprovingReviewCount: int(
				data.Repository.DefaultBranchRef.BranchProtectionRule.RequiredApprovingReviewCount),
		},
	}
}

func commitsFrom(data *graphqlData) []clients.Commit {
	ret := make([]clients.Commit, 0)
	for _, commit := range data.Repository.DefaultBranchRef.Target.Commit.History.Nodes {
		ret = append(ret, clients.Commit{
			CommittedDate: commit.CommittedDate.Time,
			Message:       string(commit.Message),
			SHA:           string(commit.Oid),
			Committer: clients.User{
				Login: string(commit.Committer.User.Login),
			},
		})
	}
	return ret
}
