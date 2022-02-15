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
	"strings"
	"sync"
	"time"

	"github.com/shurcooL/githubv4"

	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

const (
	pullRequestsToAnalyze  = 1
	issuesToAnalyze        = 30
	issueCommentsToAnalyze = 30
	reviewsToAnalyze       = 30
	labelsToAnalyze        = 30
	commitsToAnalyze       = 30
)

// nolint: govet
type graphqlData struct {
	Repository struct {
		IsArchived githubv4.Boolean
		Object     struct {
			Commit struct {
				History struct {
					Nodes []struct {
						CommittedDate githubv4.DateTime
						Message       githubv4.String
						Oid           githubv4.GitObjectID
						Author        struct {
							User struct {
								Login githubv4.String
							}
						}
						Committer struct {
							Name *string
							User struct {
								Login *string
							}
						}
						AssociatedPullRequests struct {
							Nodes []struct {
								Repository struct {
									Name  githubv4.String
									Owner struct {
										Login githubv4.String
									}
								}
								Author struct {
									Login githubv4.String
								}
								Number     githubv4.Int
								HeadRefOid githubv4.String
								MergedAt   githubv4.DateTime
								Labels     struct {
									Nodes []struct {
										Name githubv4.String
									}
								} `graphql:"labels(last: $labelsToAnalyze)"`
								Reviews struct {
									Nodes []struct {
										State  githubv4.String
										Author struct {
											Login githubv4.String
										}
									}
								} `graphql:"reviews(last: $reviewsToAnalyze)"`
							}
						} `graphql:"associatedPullRequests(first: $pullRequestsToAnalyze)"`
					}
				} `graphql:"history(first: $commitsToAnalyze)"`
			} `graphql:"... on Commit"`
		} `graphql:"object(expression: $commitExpression)"`
		Issues struct {
			Nodes []struct {
				// nolint: revive,stylecheck // naming according to githubv4 convention.
				Url               *string
				AuthorAssociation *string
				CreatedAt         *time.Time
				Comments          struct {
					Nodes []struct {
						AuthorAssociation *string
						CreatedAt         *time.Time
					}
				} `graphql:"comments(last: $issueCommentsToAnalyze)"`
			}
		} `graphql:"issues(first: $issuesToAnalyze, orderBy:{field:UPDATED_AT, direction:DESC})"`
	} `graphql:"repository(owner: $owner, name: $name)"`
	RateLimit struct {
		Cost *int
	}
}

type graphqlHandler struct {
	client   *githubv4.Client
	data     *graphqlData
	once     *sync.Once
	ctx      context.Context
	errSetup error
	repourl  *repoURL
	commits  []clients.Commit
	issues   []clients.Issue
	archived bool
}

func (handler *graphqlHandler) init(ctx context.Context, repourl *repoURL) {
	handler.ctx = ctx
	handler.repourl = repourl
	handler.data = new(graphqlData)
	handler.errSetup = nil
	handler.once = new(sync.Once)
}

func (handler *graphqlHandler) setup() error {
	handler.once.Do(func() {
		commitExpression := handler.repourl.commitSHA
		if strings.EqualFold(handler.repourl.commitSHA, clients.HeadSHA) {
			// TODO(#575): Confirm that this works as expected.
			commitExpression = fmt.Sprintf("heads/%s", handler.repourl.defaultBranch)
		}

		vars := map[string]interface{}{
			"owner":                  githubv4.String(handler.repourl.owner),
			"name":                   githubv4.String(handler.repourl.repo),
			"pullRequestsToAnalyze":  githubv4.Int(pullRequestsToAnalyze),
			"issuesToAnalyze":        githubv4.Int(issuesToAnalyze),
			"issueCommentsToAnalyze": githubv4.Int(issueCommentsToAnalyze),
			"reviewsToAnalyze":       githubv4.Int(reviewsToAnalyze),
			"labelsToAnalyze":        githubv4.Int(labelsToAnalyze),
			"commitsToAnalyze":       githubv4.Int(commitsToAnalyze),
			"commitExpression":       githubv4.String(commitExpression),
		}
		if err := handler.client.Query(handler.ctx, handler.data, vars); err != nil {
			handler.errSetup = sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("githubv4.Query: %v", err))
			return
		}
		handler.archived = bool(handler.data.Repository.IsArchived)
		handler.commits, handler.errSetup = commitsFrom(handler.data, handler.repourl.owner, handler.repourl.repo)
		if handler.errSetup != nil {
			return
		}
		handler.issues = issuesFrom(handler.data)
	})
	return handler.errSetup
}

func (handler *graphqlHandler) getCommits() ([]clients.Commit, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during graphqlHandler.setup: %w", err)
	}
	return handler.commits, nil
}

func (handler *graphqlHandler) getIssues() ([]clients.Issue, error) {
	if !strings.EqualFold(handler.repourl.commitSHA, clients.HeadSHA) {
		return nil, fmt.Errorf("%w: ListIssues only supported for HEAD queries", clients.ErrUnsupportedFeature)
	}
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during graphqlHandler.setup: %w", err)
	}
	return handler.issues, nil
}

func (handler *graphqlHandler) isArchived() (bool, error) {
	if !strings.EqualFold(handler.repourl.commitSHA, clients.HeadSHA) {
		return false, fmt.Errorf("%w: IsArchived only supported for HEAD queries", clients.ErrUnsupportedFeature)
	}
	if err := handler.setup(); err != nil {
		return false, fmt.Errorf("error during graphqlHandler.setup: %w", err)
	}
	return handler.archived, nil
}

// nolint: unparam
func commitsFrom(data *graphqlData, repoOwner, repoName string) ([]clients.Commit, error) {
	ret := make([]clients.Commit, 0)
	for _, commit := range data.Repository.Object.Commit.History.Nodes {
		var committer string
		if commit.Committer.User.Login != nil {
			committer = *commit.Committer.User.Login
		}
		// TODO(#1543): Figure out a way to safely get committer if `User.Login` is `nil`.
		var associatedPR clients.PullRequest
		for i := range commit.AssociatedPullRequests.Nodes {
			pr := commit.AssociatedPullRequests.Nodes[i]
			// NOTE: PR mergeCommit may not match commit.SHA in case repositories
			// have `enableSquashCommit` disabled. So we accept any associatedPR
			// to handle this case.
			if string(pr.Repository.Owner.Login) != repoOwner ||
				string(pr.Repository.Name) != repoName {
				continue
			}
			associatedPR = clients.PullRequest{
				Number:   int(pr.Number),
				HeadSHA:  string(pr.HeadRefOid),
				MergedAt: pr.MergedAt.Time,
				Author: clients.User{
					Login: string(pr.Author.Login),
				},
			}
			for _, label := range pr.Labels.Nodes {
				associatedPR.Labels = append(associatedPR.Labels, clients.Label{
					Name: string(label.Name),
				})
			}
			for _, review := range pr.Reviews.Nodes {
				associatedPR.Reviews = append(associatedPR.Reviews, clients.Review{
					State: string(review.State),
				})
			}
			break
		}
		ret = append(ret, clients.Commit{
			CommittedDate: commit.CommittedDate.Time,
			Message:       string(commit.Message),
			SHA:           string(commit.Oid),
			Committer: clients.User{
				Login: committer,
			},
			AssociatedMergeRequest: associatedPR,
		})
	}
	return ret, nil
}

func issuesFrom(data *graphqlData) []clients.Issue {
	var ret []clients.Issue
	for _, issue := range data.Repository.Issues.Nodes {
		var tmpIssue clients.Issue
		copyStringPtr(issue.Url, &tmpIssue.URI)
		copyRepoAssociationPtr(getRepoAssociation(issue.AuthorAssociation), &tmpIssue.AuthorAssociation)
		copyTimePtr(issue.CreatedAt, &tmpIssue.CreatedAt)
		for _, comment := range issue.Comments.Nodes {
			var tmpComment clients.IssueComment
			copyRepoAssociationPtr(getRepoAssociation(comment.AuthorAssociation), &tmpComment.AuthorAssociation)
			copyTimePtr(comment.CreatedAt, &tmpComment.CreatedAt)
			tmpIssue.Comments = append(tmpIssue.Comments, tmpComment)
		}
		ret = append(ret, tmpIssue)
	}
	return ret
}

// getRepoAssociation returns the association of the user with the repository.
func getRepoAssociation(association *string) *clients.RepoAssociation {
	if association == nil {
		return nil
	}
	var repoAssociaton clients.RepoAssociation
	switch *association {
	case "COLLABORATOR":
		repoAssociaton = clients.RepoAssociationCollaborator
	case "CONTRIBUTOR":
		repoAssociaton = clients.RepoAssociationContributor
	case "FIRST_TIMER":
		repoAssociaton = clients.RepoAssociationFirstTimer
	case "FIRST_TIME_CONTRIBUTOR":
		repoAssociaton = clients.RepoAssociationFirstTimeContributor
	case "MANNEQUIN":
		repoAssociaton = clients.RepoAssociationMannequin
	case "MEMBER":
		repoAssociaton = clients.RepoAssociationMember
	case "NONE":
		repoAssociaton = clients.RepoAssociationNone
	case "OWNER":
		repoAssociaton = clients.RepoAssociationOwner
	default:
		return nil
	}
	return &repoAssociaton
}
