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
	"time"

	"github.com/shurcooL/githubv4"

	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

const (
	pullRequestsToAnalyze  = 1
	checksToAnalyze        = 30
	issuesToAnalyze        = 30
	issueCommentsToAnalyze = 30
	reviewsToAnalyze       = 30
	labelsToAnalyze        = 30
)

//nolint:govet
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
						Signature struct {
							IsValid           bool
							WasSignedByGitHub bool
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
									Login        githubv4.String
									ResourcePath githubv4.String
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
								MergedBy struct {
									Login githubv4.String
								}
							}
						} `graphql:"associatedPullRequests(first: $pullRequestsToAnalyze)"`
					}
					PageInfo struct {
						StartCursor githubv4.String
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"history(first: $commitsToAnalyze, after: $historyCursor)"`
			} `graphql:"... on Commit"`
		} `graphql:"object(expression: $commitExpression)"`
		Issues struct {
			Nodes []struct {
				//nolint: revive,stylecheck // naming according to githubv4 convention.
				Url               *string
				AuthorAssociation *string
				Author            struct {
					Login githubv4.String
				}
				CreatedAt *time.Time
				Comments  struct {
					Nodes []struct {
						AuthorAssociation *string
						CreatedAt         *time.Time
						Author            struct {
							Login githubv4.String
						}
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
	client      *githubv4.Client
	data        *graphqlData
	setupOnce   *sync.Once
	ctx         context.Context
	errSetup    error
	repourl     *repoURL
	commits     []clients.Commit
	issues      []clients.Issue
	archived    bool
	commitDepth int
}

func (handler *graphqlHandler) init(ctx context.Context, repourl *repoURL, commitDepth int) {
	handler.ctx = ctx
	handler.repourl = repourl
	handler.data = new(graphqlData)
	handler.errSetup = nil
	handler.setupOnce = new(sync.Once)
	handler.commitDepth = commitDepth
	handler.commits = nil
	handler.issues = nil
}

func populateCommits(handler *graphqlHandler, vars map[string]interface{}) ([]clients.Commit, error) {
	var allCommits []clients.Commit
	var commitsLeft githubv4.Int
	commitsLeft, ok := vars["commitsToAnalyze"].(githubv4.Int)
	if !ok {
		return nil, nil
	}
	for vars["commitsToAnalyze"] = githubv4.Int(100); commitsLeft > 0; commitsLeft = commitsLeft - 100 {
		if commitsLeft < 100 {
			vars["commitsToAnalyze"] = commitsLeft
		}
		err := handler.client.Query(handler.ctx, handler.data, vars)
		if err != nil {
			return nil, fmt.Errorf("failed to populate commits: %w", err)
		}
		vars["historyCursor"] = handler.data.Repository.Object.Commit.History.PageInfo.EndCursor
		tmp, err := commitsFrom(handler.data, handler.repourl.owner, handler.repourl.repo)
		if err != nil {
			return nil, fmt.Errorf("failed to populate commits: %w", err)
		}
		allCommits = append(allCommits, tmp...)
	}
	return allCommits, nil
}

func (handler *graphqlHandler) setup() error {
	handler.setupOnce.Do(func() {
		commitExpression := handler.repourl.commitExpression()
		vars := map[string]interface{}{
			"owner":                  githubv4.String(handler.repourl.owner),
			"name":                   githubv4.String(handler.repourl.repo),
			"pullRequestsToAnalyze":  githubv4.Int(pullRequestsToAnalyze),
			"issuesToAnalyze":        githubv4.Int(issuesToAnalyze),
			"issueCommentsToAnalyze": githubv4.Int(issueCommentsToAnalyze),
			"reviewsToAnalyze":       githubv4.Int(reviewsToAnalyze),
			"labelsToAnalyze":        githubv4.Int(labelsToAnalyze),
			"commitsToAnalyze":       githubv4.Int(handler.commitDepth),
			"commitExpression":       githubv4.String(commitExpression),
			"historyCursor":          (*githubv4.String)(nil),
		}
		// if NumberOfCommits set to < 99 we are required by the graphql to page by 100 commits.
		if handler.commitDepth > 99 {
			handler.commits, handler.errSetup = populateCommits(handler, vars)
			handler.issues = issuesFrom(handler.data)
			handler.archived = bool(handler.data.Repository.IsArchived)
			return
		}
		if err := handler.client.Query(handler.ctx, handler.data, vars); err != nil {
			handler.errSetup = sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("githubv4.Query: %v", err))
			return
		}
		handler.commits, handler.errSetup = commitsFrom(handler.data, handler.repourl.owner, handler.repourl.repo)
		handler.issues = issuesFrom(handler.data)
		handler.archived = bool(handler.data.Repository.IsArchived)
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

// nolint
func commitsFrom(data *graphqlData, repoOwner, repoName string) ([]clients.Commit, error) {
	ret := make([]clients.Commit, 0)
	for _, commit := range data.Repository.Object.Commit.History.Nodes {
		var committer string
		// Find the commit's committer.
		if commit.Committer.User.Login != nil && *commit.Committer.User.Login != "" {
			committer = *commit.Committer.User.Login
		} else if commit.Committer.Name != nil &&
			// Username "GitHub" may indicate the commit was committed by GitHub.
			// We verify that the commit is signed by GitHub, because the name can be spoofed.
			*commit.Committer.Name == "GitHub" &&
			commit.Signature.IsValid &&
			commit.Signature.WasSignedByGitHub {
			committer = "github"
		}

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
			// ResourcePath: e.g., for dependabot, "/apps/dependabot", or "/apps/renovate"
			// Path that can be appended to "https://github.com" for a Github resource
			openedByBot := strings.HasPrefix(string(pr.Author.ResourcePath), "/apps/")
			associatedPR = clients.PullRequest{
				Number:   int(pr.Number),
				HeadSHA:  string(pr.HeadRefOid),
				MergedAt: pr.MergedAt.Time,
				Author: clients.User{
					Login: string(pr.Author.Login),
					IsBot: openedByBot,
				},
				MergedBy: clients.User{
					Login: string(pr.MergedBy.Login),
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
					Author: &clients.User{
						Login: string(review.Author.Login),
					},
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
		if issue.Author.Login != "" {
			tmpIssue.Author = &clients.User{
				Login: string(issue.Author.Login),
			}
		}
		for _, comment := range issue.Comments.Nodes {
			var tmpComment clients.IssueComment
			copyRepoAssociationPtr(getRepoAssociation(comment.AuthorAssociation), &tmpComment.AuthorAssociation)
			copyTimePtr(comment.CreatedAt, &tmpComment.CreatedAt)
			if comment.Author.Login != "" {
				tmpComment.Author = &clients.User{
					Login: string(comment.Author.Login),
				}
			}
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
