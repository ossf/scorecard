// Copyright 2026 OpenSSF Scorecard Authors
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
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v82/github"

	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/log"
)

// maintainerHandler tracks maintainer activity.
type maintainerHandler struct {
	ctx       context.Context
	repo      *Repo
	ghClient  *github.Client
	logger    *log.Logger
	setupOnce *sync.Once
	errSetup  error

	cutoff   time.Time // lower bound for "recent" activity
	elevated map[string]struct{}
	active   map[string]bool // username -> has recent activity
}

// init wires handler with context.
// Call during InitRepo.
func (h *maintainerHandler) init(
	ctx context.Context,
	ghClient *github.Client,
	repo *Repo,
) {
	h.ctx = ctx
	h.ghClient = ghClient
	h.repo = repo
	h.logger = log.NewLogger(log.DefaultLevel)
	h.setupOnce = new(sync.Once)
	h.elevated = make(map[string]struct{})
	h.active = make(map[string]bool)
}

// setCutoff updates cutoff time.
func (h *maintainerHandler) setCutoff(cutoff time.Time) {
	h.cutoff = cutoff.UTC()
}

// setup runs once to collect activity.
func (h *maintainerHandler) setup() error {
	h.setupOnce.Do(func() {
		owner, name := h.repo.owner, h.repo.repo

		// 1) Discover elevated users
		elevated, err := h.fetchElevatedUsers(owner, name)
		if err != nil {
			h.errSetup = sce.WithMessage(
				sce.ErrScorecardInternal,
				fmt.Sprintf("fetchElevatedUsers: %v", err),
			)
			return
		}
		h.elevated = elevated

		if len(h.elevated) == 0 {
			return // No maintainers found
		}

		// 2) Collect activity signals for elevated users
		if err := h.collectActivity(owner, name); err != nil {
			h.errSetup = err
			return
		}
	})
	return h.errSetup
}

// query returns the activity map after setup.
func (h *maintainerHandler) query() (map[string]bool, error) {
	if err := h.setup(); err != nil {
		return nil, err
	}

	// Ensure all users appear in result
	for user := range h.elevated {
		if _, exists := h.active[user]; !exists {
			h.active[user] = false
		}
	}

	// Debug logging: log maintainer activity details when enabled
	if os.Getenv("SCORECARD_DEBUG_MAINTAINERS") == "1" {
		var activeList, inactiveList []string
		for user, isActive := range h.active {
			if isActive {
				activeList = append(activeList, user)
			} else {
				inactiveList = append(inactiveList, user)
			}
		}

		h.logger.Info(fmt.Sprintf("Maintainer activity: %d total, %d active, %d inactive (cutoff: %s)",
			len(h.active), len(activeList), len(inactiveList), h.cutoff.Format("2006-01-02")))

		for _, user := range activeList {
			h.logger.Info(fmt.Sprintf("  ✓ Active: %s", user))
		}
		for _, user := range inactiveList {
			h.logger.Info(fmt.Sprintf("  ✗ Inactive: %s", user))
		}
	}

	return h.active, nil
}

// fetchElevatedUsers gets elevated users.
func (h *maintainerHandler) fetchElevatedUsers(
	owner, repo string,
) (map[string]struct{}, error) {
	elevated := make(map[string]struct{})

	// Fetch collaborators with elevated perms
	opts := &github.ListCollaboratorsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		collaborators, resp, err := h.ghClient.Repositories.
			ListCollaborators(h.ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("ListCollaborators: %w", err)
		}

		for _, collab := range collaborators {
			if collab.Permissions == nil || collab.Login == nil {
				continue
			}

			// Check for elevated permissions
			if collab.Permissions["admin"] || collab.Permissions["maintain"] ||
				collab.Permissions["push"] || collab.Permissions["triage"] {
				login := strings.ToLower(*collab.Login)
				elevated[login] = struct{}{}
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return elevated, nil
}

// collectActivity gathers activity signals.
// Early termination when all users active.
//
//nolint:gocognit,gocyclo
func (h *maintainerHandler) collectActivity(owner, repo string) error {
	// Primary signals - strong maintainer activity
	if err := h.checkCommits(owner, repo); err != nil {
		return fmt.Errorf("checkCommits: %w", err)
	}
	if h.allActive() {
		return nil
	}

	if err := h.checkMergedPRs(owner, repo); err != nil {
		return fmt.Errorf("checkMergedPRs: %w", err)
	}
	if h.allActive() {
		return nil
	}

	if err := h.checkReleases(owner, repo); err != nil {
		return fmt.Errorf("checkReleases: %w", err)
	}
	if h.allActive() {
		return nil
	}

	// Secondary signals - code review and triage
	if err := h.checkReviews(owner, repo); err != nil {
		return fmt.Errorf("checkReviews: %w", err)
	}
	if h.allActive() {
		return nil
	}

	if err := h.checkIssueComments(owner, repo); err != nil {
		return fmt.Errorf("checkIssueComments: %w", err)
	}
	if h.allActive() {
		return nil
	}

	if err := h.checkCommitComments(owner, repo); err != nil {
		return fmt.Errorf("checkCommitComments: %w", err)
	}
	if h.allActive() {
		return nil
	}

	// Extended signals - issue/PR management
	if err := h.checkIssueActivity(owner, repo); err != nil {
		return fmt.Errorf("checkIssueActivity: %w", err)
	}
	if h.allActive() {
		return nil
	}

	if err := h.checkPRActivity(owner, repo); err != nil {
		return fmt.Errorf("checkPRActivity: %w", err)
	}
	if h.allActive() {
		return nil
	}

	// Triage signals - labels and milestones
	if err := h.checkLabelActivity(owner, repo); err != nil {
		return fmt.Errorf("checkLabelActivity: %w", err)
	}
	if h.allActive() {
		return nil
	}

	if err := h.checkMilestoneActivity(owner, repo); err != nil {
		return fmt.Errorf("checkMilestoneActivity: %w", err)
	}
	if h.allActive() {
		return nil
	}

	// Discussion signals
	if err := h.checkDiscussions(owner, repo); err != nil {
		return fmt.Errorf("checkDiscussions: %w", err)
	}
	if h.allActive() {
		return nil
	}

	// Project board signals
	if err := h.checkProjectActivity(owner, repo); err != nil {
		return fmt.Errorf("checkProjectActivity: %w", err)
	}
	if h.allActive() {
		return nil
	}

	// Security signals
	if err := h.checkSecurityActivity(owner, repo); err != nil {
		return fmt.Errorf("checkSecurityActivity: %w", err)
	}
	if h.allActive() {
		return nil
	}

	// Workflow/Actions signals
	if err := h.checkWorkflowActivity(owner, repo); err != nil {
		return fmt.Errorf("checkWorkflowActivity: %w", err)
	}
	if h.allActive() {
		return nil
	}

	// Reaction signals - lightweight engagement indicators
	if err := h.checkIssueReactions(owner, repo); err != nil {
		return fmt.Errorf("checkIssueReactions: %w", err)
	}
	if h.allActive() {
		return nil
	}

	if err := h.checkPRReactions(owner, repo); err != nil {
		return fmt.Errorf("checkPRReactions: %w", err)
	}
	if h.allActive() {
		return nil
	}

	if err := h.checkCommentReactions(owner, repo); err != nil {
		return fmt.Errorf("checkCommentReactions: %w", err)
	}
	if h.allActive() {
		return nil
	}

	if err := h.checkCommitCommentReactions(owner, repo); err != nil {
		return fmt.Errorf("checkCommitCommentReactions: %w", err)
	}
	if h.allActive() {
		return nil
	}

	// Repository management signals (settings, etc.)
	if err := h.checkRepoEvents(owner, repo); err != nil {
		return fmt.Errorf("checkRepoEvents: %w", err)
	}

	return nil
}

// markActive marks a user as active if they're in the elevated set
// and not already marked.
// Returns true if the user was newly marked active, false if already
// active or not elevated.
func (h *maintainerHandler) markActive(login string) bool {
	if login == "" {
		return false
	}
	login = strings.ToLower(login)
	if _, isElevated := h.elevated[login]; isElevated {
		if !h.active[login] {
			h.active[login] = true
			return true
		}
	}
	return false
}

// allActive checks all users active.
// Used for early termination.
func (h *maintainerHandler) allActive() bool {
	for user := range h.elevated {
		if !h.active[user] {
			return false
		}
	}
	return len(h.elevated) > 0
}

// isLabelEvent checks for label actions.
func isLabelEvent(event *github.Timeline) bool {
	return event.Event != nil && (*event.Event == "labeled" || *event.Event == "unlabeled")
}

// isMilestoneEvent checks for milestone actions.
func isMilestoneEvent(event *github.Timeline) bool {
	return event.Event != nil && (*event.Event == "milestoned" || *event.Event == "demilestoned")
}

// hasValidTimelineActor checks for non-nil actor.
func hasValidTimelineActor(event *github.Timeline) bool {
	return event.Actor != nil && event.Actor.Login != nil
}

// hasValidEventActor checks for non-nil actor.
func hasValidEventActor(event *github.Event) bool {
	return event.Actor != nil && event.Actor.Login != nil
}

// isMilestoneRecentlyCreated checks creation after cutoff.
func (h *maintainerHandler) isMilestoneRecentlyCreated(milestone *github.Milestone) bool {
	return milestone.CreatedAt != nil && !milestone.CreatedAt.Before(h.cutoff)
}

// isMilestoneRecentlyUpdated checks update after cutoff.
func (h *maintainerHandler) isMilestoneRecentlyUpdated(milestone *github.Milestone) bool {
	return milestone.UpdatedAt != nil && !milestone.UpdatedAt.Before(h.cutoff)
}

// hasValidCreator checks for non-nil creator.
func hasValidCreator(milestone *github.Milestone) bool {
	return milestone.Creator != nil && milestone.Creator.Login != nil
}

// hasValidWorkflowActor checks for non-nil actor.
func hasValidWorkflowActor(run *github.WorkflowRun) bool {
	return run.Actor != nil && run.Actor.Login != nil
}

// checkCommits scans recent commits on the default branch.
func (h *maintainerHandler) checkCommits(owner, repo string) error {
	opts := &github.CommitsListOptions{
		Since:       h.cutoff,
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		commits, resp, err := h.ghClient.Repositories.ListCommits(
			h.ctx,
			owner,
			repo,
			opts,
		)
		if err != nil {
			return fmt.Errorf("ListCommits: %w", err)
		}

		for _, commit := range commits {
			if commit.Author != nil && commit.Author.Login != nil {
				h.markActive(*commit.Author.Login)
			}
			if commit.Committer != nil && commit.Committer.Login != nil {
				h.markActive(*commit.Committer.Login)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return nil
}

// checkMergedPRs scans recently merged pull requests.
func (h *maintainerHandler) checkMergedPRs(owner, repo string) error {
	opts := &github.PullRequestListOptions{
		State:       "closed",
		Sort:        "updated",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		prs, resp, err := h.ghClient.PullRequests.List(
			h.ctx,
			owner,
			repo,
			opts,
		)
		if err != nil {
			return fmt.Errorf("list PRs: %w", err)
		}

		foundOld := false
		for _, pr := range prs {
			if pr.MergedAt == nil {
				continue
			}
			if pr.MergedAt.Before(h.cutoff) {
				foundOld = true
				break
			}

			// PR was merged recently
			if pr.User != nil && pr.User.Login != nil {
				h.markActive(*pr.User.Login)
			}
			if pr.MergedBy != nil && pr.MergedBy.Login != nil {
				h.markActive(*pr.MergedBy.Login)
			}
		}

		if foundOld || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return nil
}

// checkReviews scans PR reviews.
func (h *maintainerHandler) checkReviews(owner, repo string) error {
	// Get recent PRs
	opts := &github.PullRequestListOptions{
		State:       "all",
		Sort:        "updated",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 50},
	}

	prs, _, err := h.ghClient.PullRequests.List(
		h.ctx,
		owner,
		repo,
		opts,
	)
	if err != nil {
		return fmt.Errorf("list PRs for reviews: %w", err)
	}

	for _, pr := range prs {
		if pr.UpdatedAt != nil && pr.UpdatedAt.Before(h.cutoff) {
			break
		}

		if h.allActive() {
			break // All maintainers already active
		}

		// Get reviews for this PR
		reviewOpts := &github.ListOptions{PerPage: 100}
		reviews, _, err := h.ghClient.PullRequests.ListReviews(
			h.ctx,
			owner,
			repo,
			*pr.Number,
			reviewOpts,
		)
		if err != nil {
			continue // Skip on error
		}

		for _, review := range reviews {
			if review.SubmittedAt != nil && !review.SubmittedAt.Before(h.cutoff) {
				if review.User != nil && review.User.Login != nil {
					if h.markActive(*review.User.Login) && h.allActive() {
						return nil // All maintainers now active
					}
				}
			}
		}
	}

	return nil
}

// checkIssueComments scans recent issue and PR comments.
func (h *maintainerHandler) checkIssueComments(owner, repo string) error {
	opts := &github.IssueListCommentsOptions{
		Since:       &h.cutoff,
		Sort:        github.Ptr("updated"),
		Direction:   github.Ptr("desc"),
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		comments, resp, err := h.ghClient.Issues.ListComments(
			h.ctx,
			owner,
			repo,
			0,
			opts,
		)
		if err != nil {
			return fmt.Errorf("ListComments: %w", err)
		}

		for _, comment := range comments {
			if comment.User != nil && comment.User.Login != nil {
				h.markActive(*comment.User.Login)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return nil
}

// checkReleases scans recent releases.
func (h *maintainerHandler) checkReleases(owner, repo string) error {
	opts := &github.ListOptions{PerPage: 100}

	for {
		releases, resp, err := h.ghClient.Repositories.ListReleases(
			h.ctx,
			owner,
			repo,
			opts,
		)
		if err != nil {
			return fmt.Errorf("ListReleases: %w", err)
		}

		for _, release := range releases {
			if release.CreatedAt != nil && !release.CreatedAt.Before(h.cutoff) {
				if release.Author != nil && release.Author.Login != nil {
					h.markActive(*release.Author.Login)
				}
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return nil
}

// checkIssueActivity scans issues.
//
//nolint:gocognit
func (h *maintainerHandler) checkIssueActivity(owner, repo string) error {
	opts := &github.IssueListByRepoOptions{
		State:       "all",
		Sort:        "updated",
		Direction:   "desc",
		Since:       h.cutoff,
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		issues, resp, err := h.ghClient.Issues.ListByRepo(
			h.ctx,
			owner,
			repo,
			opts,
		)
		if err != nil {
			return fmt.Errorf("ListByRepo: %w", err)
		}

		for _, issue := range issues {
			if issue.IsPullRequest() {
				continue // Skip PRs, handled separately
			}

			// Creator activity
			if issue.CreatedAt != nil && !issue.CreatedAt.Before(h.cutoff) {
				if issue.User != nil && issue.User.Login != nil {
					h.markActive(*issue.User.Login)
				}
			}

			// Closer activity
			if issue.ClosedAt != nil && !issue.ClosedAt.Before(h.cutoff) {
				if issue.ClosedBy != nil && issue.ClosedBy.Login != nil {
					h.markActive(*issue.ClosedBy.Login)
				}
			}

			// Assignee activity (being assigned)
			if issue.UpdatedAt != nil && !issue.UpdatedAt.Before(h.cutoff) {
				if issue.Assignee != nil && issue.Assignee.Login != nil {
					h.markActive(*issue.Assignee.Login)
				}
				for _, assignee := range issue.Assignees {
					if assignee.Login != nil {
						h.markActive(*assignee.Login)
					}
				}
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.ListOptions.Page = resp.NextPage
	}

	return nil
}

// checkPRActivity scans PRs.
//
//nolint:gocognit
func (h *maintainerHandler) checkPRActivity(owner, repo string) error {
	opts := &github.PullRequestListOptions{
		State:       "all",
		Sort:        "updated",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		prs, resp, err := h.ghClient.PullRequests.List(
			h.ctx,
			owner,
			repo,
			opts,
		)
		if err != nil {
			return fmt.Errorf("list PRs: %w", err)
		}

		foundOld := false
		for _, pr := range prs {
			if pr.UpdatedAt != nil && pr.UpdatedAt.Before(h.cutoff) {
				foundOld = true
				break
			}

			// PR creation
			if pr.CreatedAt != nil && !pr.CreatedAt.Before(h.cutoff) {
				if pr.User != nil && pr.User.Login != nil {
					h.markActive(*pr.User.Login)
				}
			}

			// Review requesters indicate triage
			for _, reviewer := range pr.RequestedReviewers {
				if reviewer.Login != nil {
					h.markActive(*reviewer.Login)
				}
			}

			// Assignee activity
			if pr.Assignee != nil && pr.Assignee.Login != nil {
				h.markActive(*pr.Assignee.Login)
			}
			for _, assignee := range pr.Assignees {
				if assignee.Login != nil {
					h.markActive(*assignee.Login)
				}
			}
		}

		if foundOld || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return nil
}

// checkIssueReactions scans reactions on issues.
// Note: Reactions lack timestamps.
func (h *maintainerHandler) checkIssueReactions(owner, repo string) error {
	opts := &github.IssueListByRepoOptions{
		State:       "all",
		Sort:        "updated",
		Direction:   "desc",
		Since:       h.cutoff,
		ListOptions: github.ListOptions{PerPage: 50},
	}

	issues, _, err := h.ghClient.Issues.ListByRepo(
		h.ctx,
		owner,
		repo,
		opts,
	)
	if err != nil {
		return fmt.Errorf("ListByRepo: %w", err)
	}

	for _, issue := range issues {
		if issue.IsPullRequest() {
			continue
		}

		// Get reactions for this issue
		reactions, _, err := h.ghClient.Reactions.ListIssueReactions(
			h.ctx,
			owner,
			repo,
			*issue.Number,
			&github.ListReactionOptions{ListOptions: github.ListOptions{PerPage: 100}},
		)
		if err != nil {
			continue // Skip on error
		}

		// Reactions indicate activity
		for _, reaction := range reactions {
			if reaction.User != nil && reaction.User.Login != nil {
				h.markActive(*reaction.User.Login)
			}
		}
	}

	return nil
}

// checkPRReactions scans reactions on PRs.
// Note: Reactions lack timestamps.
func (h *maintainerHandler) checkPRReactions(owner, repo string) error {
	opts := &github.PullRequestListOptions{
		State:       "all",
		Sort:        "updated",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 50},
	}

	prs, _, err := h.ghClient.PullRequests.List(
		h.ctx,
		owner,
		repo,
		opts,
	)
	if err != nil {
		return fmt.Errorf("list PRs: %w", err)
	}

	for _, pr := range prs {
		if pr.UpdatedAt != nil && pr.UpdatedAt.Before(h.cutoff) {
			break
		}

		// Get reactions for this PR
		reactions, _, err := h.ghClient.Reactions.ListIssueReactions(
			h.ctx,
			owner,
			repo,
			*pr.Number,
			&github.ListReactionOptions{ListOptions: github.ListOptions{PerPage: 100}},
		)
		if err != nil {
			continue // Skip on error
		}

		// Reactions indicate activity
		for _, reaction := range reactions {
			if reaction.User != nil && reaction.User.Login != nil {
				h.markActive(*reaction.User.Login)
			}
		}
	}

	return nil
}

// checkCommentReactions scans reactions.
// Note: Reactions lack timestamps.
func (h *maintainerHandler) checkCommentReactions(owner, repo string) error {
	opts := &github.IssueListCommentsOptions{
		Since:       &h.cutoff,
		Sort:        github.Ptr("updated"),
		Direction:   github.Ptr("desc"),
		ListOptions: github.ListOptions{PerPage: 100},
	}

	comments, _, err := h.ghClient.Issues.ListComments(
		h.ctx,
		owner,
		repo,
		0,
		opts,
	)
	if err != nil {
		return fmt.Errorf("ListComments: %w", err)
	}

	for _, comment := range comments {
		// Get reactions for this comment
		reactions, _, err := h.ghClient.Reactions.ListIssueCommentReactions(
			h.ctx,
			owner,
			repo,
			*comment.ID,
			&github.ListReactionOptions{ListOptions: github.ListOptions{PerPage: 100}},
		)
		if err != nil {
			continue // Skip on error
		}

		// Reactions indicate activity
		for _, reaction := range reactions {
			if reaction.User != nil && reaction.User.Login != nil {
				h.markActive(*reaction.User.Login)
			}
		}
	}

	return nil
}

// checkCommitCommentReactions scans commit comments.
// Note: Reactions lack timestamps.
func (h *maintainerHandler) checkCommitCommentReactions(owner, repo string) error {
	// Get recent commits
	commitOpts := &github.CommitsListOptions{
		Since:       h.cutoff,
		ListOptions: github.ListOptions{PerPage: 50},
	}

	commits, _, err := h.ghClient.Repositories.ListCommits(
		h.ctx,
		owner,
		repo,
		commitOpts,
	)
	if err != nil {
		return fmt.Errorf("ListCommits: %w", err)
	}

	for _, commit := range commits {
		if commit.SHA == nil {
			continue
		}

		// Get comments on this commit
		commentOpts := &github.ListOptions{PerPage: 100}
		comments, _, err := h.ghClient.Repositories.ListComments(
			h.ctx,
			owner,
			repo,
			commentOpts,
		)
		if err != nil {
			continue
		}

		for _, comment := range comments {
			if comment.CommitID == nil || *comment.CommitID != *commit.SHA {
				continue
			}

			// Get reactions for this commit comment
			reactions, _, err := h.ghClient.Reactions.ListCommentReactions(
				h.ctx,
				owner,
				repo,
				*comment.ID,
				&github.ListReactionOptions{ListOptions: github.ListOptions{PerPage: 100}},
			)
			if err != nil {
				continue
			}

			// Reactions indicate activity
			for _, reaction := range reactions {
				if reaction.User != nil && reaction.User.Login != nil {
					h.markActive(*reaction.User.Login)
				}
			}
		}
	}

	return nil
}

// checkRepoEvents checks repo events.
func (h *maintainerHandler) checkRepoEvents(owner, repo string) error {
	// Events API for repository activities
	opts := &github.ListOptions{PerPage: 100}

	for {
		events, resp, err := h.ghClient.Activity.ListRepositoryEvents(
			h.ctx,
			owner,
			repo,
			opts,
		)
		if err != nil {
			return fmt.Errorf("ListRepositoryEvents: %w", err)
		}

		foundOld := false
		for _, event := range events {
			if event.CreatedAt != nil && event.CreatedAt.Before(h.cutoff) {
				foundOld = true
				break
			}

			// Track various event types that indicate maintainer activity
			if !hasValidEventActor(event) {
				continue
			}
			switch *event.Type {
			case "PushEvent", "CreateEvent", "DeleteEvent",
				"ReleaseEvent", "PublicEvent", "MemberEvent",
				"GollumEvent",                   // Wiki edits
				"IssuesEvent",                   // Issue lifecycle
				"IssueCommentEvent",             // Issue comments
				"PullRequestEvent",              // PR lifecycle
				"PullRequestReviewEvent",        // Review submissions
				"PullRequestReviewCommentEvent", // Review comments
				"CommitCommentEvent",            // Commit comments
				"RepositoryEvent":               // Repository settings changes
				h.markActive(*event.Actor.Login)
			}
		}

		if foundOld || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return nil
}

// checkCommitComments scans direct comments on commits (not reactions).
// Note: This requires iterating through commits and checking for comments on each.
func (h *maintainerHandler) checkCommitComments(owner, repo string) error {
	// Get recent commits
	commitOpts := &github.CommitsListOptions{
		Since:       h.cutoff,
		ListOptions: github.ListOptions{PerPage: 100},
	}

	commits, _, err := h.ghClient.Repositories.ListCommits(
		h.ctx,
		owner,
		repo,
		commitOpts,
	)
	if err != nil {
		return fmt.Errorf("ListCommits: %w", err)
	}

	// Check for comments on each recent commit
	for _, commit := range commits {
		if commit.SHA == nil {
			continue
		}

		if h.allActive() {
			break // All maintainers already active
		}

		commentOpts := &github.ListOptions{PerPage: 100}
		comments, _, err := h.ghClient.Repositories.ListCommitComments(
			h.ctx,
			owner,
			repo,
			*commit.SHA,
			commentOpts,
		)
		if err != nil {
			continue // Skip on error
		}

		for _, comment := range comments {
			if comment.CreatedAt != nil && !comment.CreatedAt.Before(h.cutoff) {
				if comment.User != nil && comment.User.Login != nil {
					if h.markActive(*comment.User.Login) && h.allActive() {
						return nil // All maintainers now active
					}
				}
			}
		}
	}

	return nil
}

// checkLabelActivity scans label management actions.
func (h *maintainerHandler) checkLabelActivity(owner, repo string) error {
	// Get repository label events via issue timeline
	opts := &github.IssueListByRepoOptions{
		State:       "all",
		Sort:        "updated",
		Direction:   "desc",
		Since:       h.cutoff,
		ListOptions: github.ListOptions{PerPage: 50},
	}

	issues, _, err := h.ghClient.Issues.ListByRepo(
		h.ctx,
		owner,
		repo,
		opts,
	)
	if err != nil {
		return fmt.Errorf("ListByRepo: %w", err)
	}

	for _, issue := range issues {
		if h.allActive() {
			break // All maintainers already active
		}

		// Get timeline events for this issue
		timelineOpts := &github.ListOptions{PerPage: 100}
		timeline, _, err := h.ghClient.Issues.ListIssueTimeline(
			h.ctx,
			owner,
			repo,
			*issue.Number,
			timelineOpts,
		)
		if err != nil {
			continue // Skip on error
		}

		for _, event := range timeline {
			if event.CreatedAt == nil || event.CreatedAt.Before(h.cutoff) {
				continue
			}

			// Check for label events
			if !isLabelEvent(event) {
				continue
			}
			if !hasValidTimelineActor(event) {
				continue
			}
			if h.markActive(*event.Actor.Login) && h.allActive() {
				return nil // All maintainers now active
			}
		}
	}

	return nil
}

// checkMilestoneActivity scans milestone management actions.
//
//nolint:gocognit // Handles creation, updates, and timeline events with pagination
func (h *maintainerHandler) checkMilestoneActivity(owner, repo string) error {
	// Check milestone creation/editing
	opts := &github.MilestoneListOptions{
		State:       "all",
		Sort:        "updated",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		milestones, resp, err := h.ghClient.Issues.ListMilestones(
			h.ctx,
			owner,
			repo,
			opts,
		)
		if err != nil {
			return fmt.Errorf("ListMilestones: %w", err)
		}

		for _, milestone := range milestones {
			// Check milestone creation
			if h.isMilestoneRecentlyCreated(milestone) && hasValidCreator(milestone) {
				h.markActive(*milestone.Creator.Login)
			}

			// Check milestone updates
			if h.isMilestoneRecentlyUpdated(milestone) && hasValidCreator(milestone) {
				h.markActive(*milestone.Creator.Login)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	// Check milestone assignment events via issue timeline
	opts2 := &github.IssueListByRepoOptions{
		State:       "all",
		Sort:        "updated",
		Direction:   "desc",
		Since:       h.cutoff,
		ListOptions: github.ListOptions{PerPage: 50},
	}

	issues, _, err := h.ghClient.Issues.ListByRepo(
		h.ctx,
		owner,
		repo,
		opts2,
	)
	if err != nil {
		return fmt.Errorf("ListByRepo: %w", err)
	}

	for _, issue := range issues {
		if h.allActive() {
			break // All maintainers already active
		}

		timelineOpts := &github.ListOptions{PerPage: 100}
		timeline, _, err := h.ghClient.Issues.ListIssueTimeline(
			h.ctx,
			owner,
			repo,
			*issue.Number,
			timelineOpts,
		)
		if err != nil {
			continue
		}

		for _, event := range timeline {
			if event.CreatedAt == nil || event.CreatedAt.Before(h.cutoff) {
				continue
			}

			if !isMilestoneEvent(event) {
				continue
			}
			if !hasValidTimelineActor(event) {
				continue
			}
			if h.markActive(*event.Actor.Login) && h.allActive() {
				return nil // All maintainers now active
			}
		}
	}

	return nil
}

// checkDiscussions scans GitHub Discussions activity.
func (h *maintainerHandler) checkDiscussions(owner, repo string) error {
	// Note: GitHub Discussions require GraphQL API or special REST endpoints
	// The go-github v53 library has limited discussions support
	// We'll use the repository discussions events if available
	opts := &github.ListOptions{PerPage: 100}

	for {
		events, resp, err := h.ghClient.Activity.ListRepositoryEvents(
			h.ctx,
			owner,
			repo,
			opts,
		)
		if err != nil {
			return fmt.Errorf("ListRepositoryEvents: %w", err)
		}

		foundOld := false
		for _, event := range events {
			if event.CreatedAt != nil && event.CreatedAt.Before(h.cutoff) {
				foundOld = true
				break
			}

			if !hasValidEventActor(event) || event.Type == nil {
				continue
			}
			if *event.Type == "DiscussionEvent" || *event.Type == "DiscussionCommentEvent" {
				h.markActive(*event.Actor.Login)
			}
		}

		if foundOld || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return nil
}

// checkProjectActivity scans GitHub Projects (board) activity.
func (h *maintainerHandler) checkProjectActivity(owner, repo string) error {
	// GitHub Projects (Classic) API was deprecated and removed in v82.
	// This check is no longer functional and returns early.
	// Project activity tracking would need to be reimplemented using the new Projects v2 GraphQL API.
	return nil
}

// checkSecurityActivity scans security.
func (h *maintainerHandler) checkSecurityActivity(owner, repo string) error {
	// Check for Dependabot alerts
	// Note: This requires special permissions and may not be available
	vulnOpts := &github.ListAlertsOptions{
		State: github.Ptr("dismissed"),
	}

	alerts, _, err := h.ghClient.Dependabot.ListRepoAlerts(
		h.ctx,
		owner,
		repo,
		vulnOpts,
	)
	if err != nil {
		// Dependabot might not be enabled or accessible
		return fmt.Errorf("ListRepoAlerts: %w", err)
	}

	for _, alert := range alerts {
		// Check who dismissed the alert
		if alert.DismissedAt != nil && !alert.DismissedAt.Before(h.cutoff) {
			if alert.DismissedBy != nil && alert.DismissedBy.Login != nil {
				h.markActive(*alert.DismissedBy.Login)
			}
		}
	}

	return nil
}

// checkWorkflowActivity scans GitHub Actions/workflow management.
func (h *maintainerHandler) checkWorkflowActivity(owner, repo string) error {
	// Check workflow runs (manual triggers)
	opts := &github.ListWorkflowRunsOptions{
		Created:     fmt.Sprintf(">=%s", h.cutoff.Format("2006-01-02")),
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		runs, resp, err := h.ghClient.Actions.ListRepositoryWorkflowRuns(
			h.ctx,
			owner,
			repo,
			opts,
		)
		if err != nil {
			// Actions might not be enabled
			return fmt.Errorf("ListRepositoryWorkflowRuns: %w", err)
		}

		for _, run := range runs.WorkflowRuns {
			// Check who triggered the run
			if run.CreatedAt == nil || run.CreatedAt.Before(h.cutoff) {
				continue
			}
			if hasValidWorkflowActor(run) {
				h.markActive(*run.Actor.Login)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return nil
}
