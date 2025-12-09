// Copyright 2025 OpenSSF Scorecard Authors
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

	"github.com/google/go-github/v53/github"
	githubv4 "github.com/shurcooL/githubv4"

	"github.com/ossf/scorecard/v5/clients"
	sce "github.com/ossf/scorecard/v5/errors"
)

// isTrackedLabel checks if a label name is in the TrackedIssueLabels list.
func isTrackedLabel(name string) bool {
	normalized := strings.ToLower(strings.TrimSpace(name))
	for _, label := range clients.TrackedIssueLabels {
		if normalized == label {
			return true
		}
	}
	return false
}

// issuesHandler fetches issues and their timeline once (GraphQL) and caches them.
//
//nolint:govet // fieldalignment: function pointers first for alignment
type issuesHandler struct {
	// Legacy hooks (kept for tests/back-compat but unused now).
	listIssuesFn func(
		context.Context, string, string, *github.IssueListByRepoOptions,
	) ([]*github.Issue, *github.Response, error)
	listRepoEventsFn func(
		context.Context, string, string, *github.ListOptions,
	) ([]*github.IssueEvent, *github.Response, error)
	listRepoCommentsFn func(
		context.Context, string, string, int, *github.IssueListCommentsOptions,
	) ([]*github.IssueComment, *github.Response, error)
	listCollaboratorsFn func(
		context.Context, string, string, *github.ListCollaboratorsOptions,
	) ([]*github.User, *github.Response, error)

	ghClient    *github.Client
	graphClient *githubv4.Client
	repourl     *Repo
	once        *sync.Once
	ctx         context.Context

	// Cached issues with comments + label events.
	issues []clients.Issue

	errSetup error
}

// init wires ctx/repo and sets default hooks (unused in GraphQL path).
func (h *issuesHandler) init(ctx context.Context, repourl *Repo) {
	h.ctx = ctx
	h.repourl = repourl
	h.once = new(sync.Once)

	if h.listIssuesFn == nil {
		h.listIssuesFn = func(
			ctx context.Context, owner, repo string, opt *github.IssueListByRepoOptions,
		) ([]*github.Issue, *github.Response, error) {
			return h.ghClient.Issues.ListByRepo(ctx, owner, repo, opt)
		}
	}
	if h.listRepoEventsFn == nil {
		h.listRepoEventsFn = func(
			ctx context.Context, owner, repo string, opt *github.ListOptions,
		) ([]*github.IssueEvent, *github.Response, error) {
			return h.ghClient.Issues.ListRepositoryEvents(ctx, owner, repo, opt)
		}
	}
	if h.listRepoCommentsFn == nil {
		h.listRepoCommentsFn = func(
			ctx context.Context, owner, repo string, number int, opt *github.IssueListCommentsOptions,
		) ([]*github.IssueComment, *github.Response, error) {
			return h.ghClient.Issues.ListComments(ctx, owner, repo, number, opt)
		}
	}
	if h.listCollaboratorsFn == nil {
		h.listCollaboratorsFn = func(
			ctx context.Context, owner, repo string, opt *github.ListCollaboratorsOptions,
		) ([]*github.User, *github.Response, error) {
			return h.ghClient.Repositories.ListCollaborators(ctx, owner, repo, opt)
		}
	}
}

// setup uses GraphQL to fetch issues (not PRs) ordered by UPDATED_AT DESC,
// and for each issue grabs timeline items: Labeled/Unlabeled events and comments.
//
//nolint:gocognit // Complex GraphQL query setup with multiple nested iterations
func (h *issuesHandler) setup() error {
	h.once.Do(func() {
		const maxIssues = 2000

		if h.graphClient == nil {
			h.errSetup = sce.WithMessage(sce.ErrScorecardInternal, "github GraphQL client not initialized")
			return
		}

		type timelineNode struct {
			Typename     githubv4.String `graphql:"__typename"`
			LabeledEvent struct {
				CreatedAt time.Time
				Label     struct{ Name string }
				Actor     struct {
					Login githubv4.String
				}
			} `graphql:"... on LabeledEvent"`
			UnlabeledEvent struct {
				CreatedAt time.Time
				Label     struct{ Name string }
				Actor     struct {
					Login githubv4.String
				}
			} `graphql:"... on UnlabeledEvent"`
			IssueComment struct {
				CreatedAt         time.Time
				URL               githubv4.URI
				AuthorAssociation githubv4.String
				Author            struct{ Login githubv4.String }
			} `graphql:"... on IssueComment"`
			ClosedEvent struct {
				CreatedAt time.Time
				Actor     struct{ Login githubv4.String }
			} `graphql:"... on ClosedEvent"`
			ReopenedEvent struct {
				CreatedAt time.Time
				Actor     struct{ Login githubv4.String }
			} `graphql:"... on ReopenedEvent"`
		}

		type issuesQuery struct {
			Repository struct {
				Issues struct {
					Nodes []struct { //nolint:govet // fieldalignment: GraphQL query struct field order matches API schema
						URL               githubv4.URI
						ClosedAt          *githubv4.DateTime
						CreatedAt         time.Time
						Number            int
						Author            struct{ Login githubv4.String }
						AuthorAssociation githubv4.String
						Timeline          struct {
							Nodes    []timelineNode
							PageInfo struct {
								EndCursor   githubv4.String
								HasNextPage bool
							}
						} `graphql:"timelineItems(first: 100, itemTypes: [LABELED_EVENT, UNLABELED_EVENT, ISSUE_COMMENT, CLOSED_EVENT, REOPENED_EVENT])"` //nolint:lll
					}
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"issues(first: 100, after: $afterIssues, states: [OPEN, CLOSED], orderBy: {field: UPDATED_AT, direction: DESC})"` //nolint:lll
			} `graphql:"repository(owner: $owner, name: $name)"`
		}

		type timelineMoreQuery struct {
			Repository struct {
				Issue struct {
					Timeline struct {
						Nodes    []timelineNode
						PageInfo struct {
							EndCursor   githubv4.String
							HasNextPage bool
						}
					} `graphql:"timelineItems(first: $first, after: $after, itemTypes: [LABELED_EVENT, UNLABELED_EVENT, ISSUE_COMMENT, CLOSED_EVENT, REOPENED_EVENT])"` //nolint:lll
				} `graphql:"issue(number: $number)"`
			} `graphql:"repository(owner: $owner, name: $name)"`
		}

		owner, repo := h.repourl.owner, h.repourl.repo
		vars := map[string]interface{}{
			"owner":       githubv4.String(owner),
			"name":        githubv4.String(repo),
			"afterIssues": (*githubv4.String)(nil),
		}

		var out []clients.Issue
	issuePages:
		for {
			// Context cancellation between pages.
			select {
			case <-h.ctx.Done():
				h.errSetup = h.ctx.Err()
				return
			default:
			}

			var q issuesQuery
			if err := h.graphClient.Query(h.ctx, &q, vars); err != nil {
				h.errSetup = sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("githubv4.Query issues: %v", err))
				return
			}

			for i := range q.Repository.Issues.Nodes {
				n := &q.Repository.Issues.Nodes[i]
				u := n.URL.String()
				created := n.CreatedAt
				var closedPtr *time.Time
				if n.ClosedAt != nil {
					ct := n.ClosedAt.Time
					closedPtr = &ct
				}

				issueAuthor := strings.ToLower(string(n.Author.Login))
				authorAssocStr := strings.ToUpper(string(n.AuthorAssociation))

				ci := clients.Issue{
					URI:               &u,
					IssueNumber:       n.Number,
					CreatedAt:         &created,
					ClosedAt:          closedPtr,
					Author:            &clients.User{Login: issueAuthor},
					AuthorAssociation: getRepoAssociation(&authorAssocStr),
				}

				// Attach first page of timeline.
				appendTimeline := func(nodes []timelineNode) {
					for i := range nodes {
						tn := &nodes[i]
						switch string(tn.Typename) {
						case "LabeledEvent":
							name := strings.ToLower(strings.TrimSpace(tn.LabeledEvent.Label.Name))
							actor := strings.ToLower(string(tn.LabeledEvent.Actor.Login))
							// Consider actor a maintainer if they're the repo owner or not the issue author
							isMaint := actor == owner || actor != issueAuthor

							labelEvent := clients.LabelEvent{
								Label:        name,
								Added:        true,
								Actor:        actor,
								CreatedAt:    tn.LabeledEvent.CreatedAt,
								IsMaintainer: isMaint,
							}

							// Add to AllLabelEvents for maintainer activity tracking
							ci.AllLabelEvents = append(ci.AllLabelEvents, labelEvent)

							// Also add to LabelEvents if it's a tracked label
							if isTrackedLabel(name) {
								ci.LabelEvents = append(ci.LabelEvents, labelEvent)
							}
						case "UnlabeledEvent":
							name := strings.ToLower(strings.TrimSpace(tn.UnlabeledEvent.Label.Name))
							actor := strings.ToLower(string(tn.UnlabeledEvent.Actor.Login))
							// Consider actor a maintainer if they're the repo owner or not the issue author
							isMaint := actor == owner || actor != issueAuthor

							labelEvent := clients.LabelEvent{
								Label:        name,
								Added:        false,
								Actor:        actor,
								CreatedAt:    tn.UnlabeledEvent.CreatedAt,
								IsMaintainer: isMaint,
							}

							// Add to AllLabelEvents for maintainer activity tracking
							ci.AllLabelEvents = append(ci.AllLabelEvents, labelEvent)

							// Also add to LabelEvents if it's a tracked label
							if isTrackedLabel(name) {
								ci.LabelEvents = append(ci.LabelEvents, labelEvent)
							}
						case "IssueComment":
							aa := strings.ToUpper(string(tn.IssueComment.AuthorAssociation))
							isMaint := (aa == "OWNER" || aa == "MEMBER" || aa == "COLLABORATOR")
							t := tn.IssueComment.CreatedAt
							ci.Comments = append(ci.Comments, clients.IssueComment{
								Author:       &clients.User{Login: strings.ToLower(string(tn.IssueComment.Author.Login))},
								CreatedAt:    &t,
								IsMaintainer: isMaint,
								URL:          tn.IssueComment.URL.String(),
							})
						case "ClosedEvent":
							actor := strings.ToLower(string(tn.ClosedEvent.Actor.Login))
							isMaint := actor != issueAuthor
							ci.StateChangeEvents = append(ci.StateChangeEvents, clients.StateChangeEvent{
								CreatedAt:    tn.ClosedEvent.CreatedAt,
								Actor:        actor,
								IsMaintainer: isMaint,
								Closed:       true,
							})
						case "ReopenedEvent":
							actor := strings.ToLower(string(tn.ReopenedEvent.Actor.Login))
							isMaint := actor != issueAuthor
							ci.StateChangeEvents = append(ci.StateChangeEvents, clients.StateChangeEvent{
								CreatedAt:    tn.ReopenedEvent.CreatedAt,
								Actor:        actor,
								IsMaintainer: isMaint,
								Closed:       false,
							})
						}
					}
				}
				appendTimeline(n.Timeline.Nodes)

				// If >100 timeline items, page the issue timeline.
				if n.Timeline.PageInfo.HasNextPage {
					tvars := map[string]interface{}{
						"owner":  githubv4.String(owner),
						"name":   githubv4.String(repo),
						"number": githubv4.Int(n.Number),
						"first":  githubv4.Int(100),
						"after":  n.Timeline.PageInfo.EndCursor,
					}
					for {
						select {
						case <-h.ctx.Done():
							h.errSetup = h.ctx.Err()
							return
						default:
						}
						var tq timelineMoreQuery
						if err := h.graphClient.Query(h.ctx, &tq, tvars); err != nil {
							h.errSetup = sce.WithMessage(sce.ErrScorecardInternal,
								fmt.Sprintf("githubv4.Query timeline issue#%d: %v", n.Number, err))
							return
						}
						appendTimeline(tq.Repository.Issue.Timeline.Nodes)
						if !tq.Repository.Issue.Timeline.PageInfo.HasNextPage {
							break
						}
						tvars["after"] = tq.Repository.Issue.Timeline.PageInfo.EndCursor
					}
				}

				out = append(out, ci)
				if len(out) >= maxIssues {
					h.issues = out
					break issuePages
				}
			}

			if !q.Repository.Issues.PageInfo.HasNextPage {
				break
			}
			vars["afterIssues"] = q.Repository.Issues.PageInfo.EndCursor
		}

		h.issues = out
	})
	return h.errSetup
}

// getIssues returns the cached issues without timeline history.
func (h *issuesHandler) getIssues() ([]clients.Issue, error) {
	if err := h.setup(); err != nil {
		return nil, fmt.Errorf("issuesHandler.setup: %w", err)
	}
	return h.issues, nil
}

// Since setup() already fetched timeline data, just return a shallow copy.
func (h *issuesHandler) listIssuesWithHistory() ([]clients.Issue, error) {
	if err := h.setup(); err != nil {
		return nil, fmt.Errorf("issuesHandler.setup: %w", err)
	}
	out := make([]clients.Issue, len(h.issues))
	copy(out, h.issues)
	return out, nil
}
