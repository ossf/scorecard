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

package gitlabrepo

// FILE OVERVIEW
// =============
// This file implements maintainer activity tracking for GitLab repositories.
// It identifies users with elevated permissions (Developer+) and checks if they
// have shown activity within a configurable time window (typically 6 months).
//
// ARCHITECTURE
// ------------
// The implementation uses GitLab's REST API to collect activity signals from
// multiple sources. Activity is detected through:
//   - Core signals: MR merges, releases, audit events, CI/CD jobs, pipeline schedules
//   - Extended signals: User events, resource labels/state/milestones, award emojis
//
// The handler implements early termination: once all elevated maintainers are
// confirmed active, it stops checking additional signal sources for efficiency.
//
// SIGNAL SOURCES
// --------------
// Primary (always checked):
//   1. Merge Request merges (B1)
//   2. Release authorship (B2)
// Secondary (best-effort, may have limited permissions):
//   3. Audit events (B3) - admin/settings changes
//   4. Manual CI/CD jobs (B4) - pipeline interactions
//   5. Pipeline schedules (B5) - automation management
// Extended (comprehensive coverage via raw REST):
//   6. User contribution events (B6) - pushes, comments, etc.
//   7. Resource label events (B7) - issue/MR triaging
//   8. Resource state/milestone events (B8) - issue/MR management
//   9. Award emoji (B9-B11) - reactions on issues/MRs/commits/snippets
//  10. Project events fallback (B12) - wiki edits, misc.

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	sce "github.com/ossf/scorecard/v5/errors"
)

var errUserNotFound = errors.New("user not found")

// API path format constants.
const (
	sortOrderUpdatedAt                  = "updated_at"
	apiPathProjectIssues                = "projects/%s/issues"
	apiPathIssueResourceLabelEvents     = "projects/%s/issues/%d/resource_label_events"
	apiPathIssueResourceStateEvents     = "projects/%s/issues/%d/resource_state_events"
	apiPathIssueResourceMilestoneEvents = "projects/%s/issues/%d/resource_milestone_events"
	apiPathIssueAwardEmoji              = "projects/%s/issues/%d/award_emoji"
	apiPathMRResourceLabelEvents        = "projects/%s/merge_requests/%d/resource_label_events"
	apiPathMRResourceStateEvents        = "projects/%s/merge_requests/%d/resource_state_events"
	apiPathMRResourceMilestoneEvents    = "projects/%s/merge_requests/%d/resource_milestone_events"
	apiPathMRAwardEmoji                 = "projects/%s/merge_requests/%d/award_emoji"
	apiPathCommitAwardEmoji             = "projects/%s/repository/commits/%s/award_emoji"
	apiPathSnippets                     = "projects/%s/snippets"
	apiPathSnippetAwardEmoji            = "projects/%s/snippets/%d/award_emoji"
	apiPathEvents                       = "projects/%s/events"
)

// MaintainerActivityHandler tracks activity.
// Lazy init via sync.Once; early termination.
type MaintainerActivityHandler struct {
	cutoff           time.Time
	ctx              context.Context
	setupErr         error
	gl               *gitlab.Client
	elevated         map[string]struct{}
	active           map[string]bool
	projectID        string
	minAccessLevel   gitlab.AccessLevelValue
	pageSize         int
	projectNumericID int64
	setupOnce        sync.Once
}

// init prepares the handler with GitLab client, context, and project ID.
// This should be called during InitRepo. The cutoff time is set later via setCutoff.
// Access levels: 30=Developer, 40=Maintainer, 50=Owner.
// By default, tracks Developer+ permissions (can be changed via SetMinAccessLevel).
func (h *MaintainerActivityHandler) init(
	ctx context.Context,
	gl *gitlab.Client,
	projectID string,
) {
	h.ctx = ctx
	h.gl = gl
	h.projectID = projectID

	// Default to Developer+ (level 30)
	h.minAccessLevel = gitlab.DeveloperPermissions
	h.pageSize = 100

	h.elevated = make(map[string]struct{})
	h.active = make(map[string]bool)
	h.projectNumericID = 0
}

// setCutoff updates the cutoff time for activity queries.
// Must be called before GetMaintainerActivity().
func (h *MaintainerActivityHandler) setCutoff(cutoff time.Time) {
	h.cutoff = cutoff.UTC()
}

// Allow callers to tighten the minimum access level (e.g., Maintainer+).
func (h *MaintainerActivityHandler) SetMinAccessLevel(level gitlab.AccessLevelValue) {
	h.minAccessLevel = level
}

// setup runs once to init handler.
// Resolves ID, loads users, collects activity.
// Early exit if no elevated users.
func (h *MaintainerActivityHandler) setup() error {
	h.setupOnce.Do(func() {
		if err := h.resolveProjectNumericID(); err != nil {
			h.setSetupError("resolve project id", err)
			return
		}
		if err := h.loadElevatedMembers(); err != nil {
			h.setSetupError("load members", err)
			return
		}
		// Start pessimistic (all inactive)
		for u := range h.elevated {
			h.active[u] = false
		}
		if len(h.elevated) == 0 {
			return // No maintainers to track
		}
		if err := h.collectActivity(); err != nil {
			h.setSetupError("collect activity", err)
			return
		}
	})
	return h.setupErr
}

// GetMaintainerActivity returns activity map.
func (h *MaintainerActivityHandler) GetMaintainerActivity() (map[string]bool, error) {
	if err := h.setup(); err != nil {
		return nil, err
	}
	out := make(map[string]bool, len(h.active))
	for u, v := range h.active {
		out[u] = v
	}
	return out, nil
}

////////////////////////////////////////////////////////////////////////////////
// HELPER METHODS
////////////////////////////////////////////////////////////////////////////////
// These utilities support the main activity collection logic:
//   - resolveProjectNumericID: Converts project path to numeric ID
//   - markActive: Records activity for an elevated user
//   - allActive: Checks if all maintainers are active (enables early termination)

// resolveProjectNumericID gets numeric ID.
func (h *MaintainerActivityHandler) resolveProjectNumericID() error {
	prj, _, err := h.gl.Projects.GetProject(h.projectID, &gitlab.GetProjectOptions{})
	if err != nil || prj == nil {
		return fmt.Errorf("get project: %w", err)
	}
	h.projectNumericID = prj.ID
	return nil
}

// markActive marks user active if elevated.
// Returns true if newly marked, false otherwise.
func (h *MaintainerActivityHandler) markActive(user string) bool {
	u := strings.ToLower(strings.TrimSpace(user))
	if u == "" {
		return false
	}
	if _, elevated := h.elevated[u]; elevated {
		if h.active[u] {
			return false // Already active
		}
		h.active[u] = true
		return true // Newly marked active
	}
	return false
}

// allActive checks if all elevated maintainers have been confirmed active.
// Used for early termination to avoid unnecessary API calls once all users are active.
func (h *MaintainerActivityHandler) allActive() bool {
	for u := range h.elevated {
		if !h.active[u] {
			return false
		}
	}
	return true
}

// Helper methods for common validation patterns.

// setSetupError creates a setup error with the given message.
func (h *MaintainerActivityHandler) setSetupError(msg string, err error) {
	h.setupErr = sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("%s: %v", msg, err))
}

// buildAPIPath creates an API path from a format string and arguments.
func (h *MaintainerActivityHandler) buildAPIPath(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

// isMergeRequestMergedRecently checks if a merge request was merged after the cutoff time.
// Accepts BasicMergeRequest which has the MergedAt field.
func (h *MaintainerActivityHandler) isMergeRequestMergedRecently(mr *gitlab.BasicMergeRequest) bool {
	return mr != nil && mr.MergedAt != nil && !mr.MergedAt.IsZero() && mr.MergedAt.After(h.cutoff)
}

// isManualJobValid checks if a manual job has valid user information.
func isManualJobValid(j *gitlab.Job) bool {
	return j != nil && j.User != nil && j.User.Username != ""
}

// isUserEventRecentForProject checks if a user event is recent and for the current project.
func (h *MaintainerActivityHandler) isUserEventRecentForProject(e *glUserEventJSON) bool {
	return e.ProjectID == int(h.projectNumericID) && e.CreatedAt.After(h.cutoff)
}

// isLabelEventRecentWithUser checks if a label event is recent and has a valid user.
func (h *MaintainerActivityHandler) isLabelEventRecentWithUser(e *glLabelEventJSON) bool {
	return e.CreatedAt != nil && e.CreatedAt.After(h.cutoff) && e.User.Username != ""
}

// isStateEventRecentWithUser checks if a state event is recent and has a valid user.
func (h *MaintainerActivityHandler) isStateEventRecentWithUser(e *glStateEventJSON) bool {
	return e.CreatedAt != nil && e.CreatedAt.After(h.cutoff) && e.User.Username != ""
}

// isMilestoneEventRecentWithUser checks if a milestone event is recent and has a valid user.
func (h *MaintainerActivityHandler) isMilestoneEventRecentWithUser(e *glMilestoneEventJSON) bool {
	return e.CreatedAt != nil && e.CreatedAt.After(h.cutoff) && e.User.Username != ""
}

// isAwardEmojiRecentWithUser checks if an award emoji is recent and has a valid user.
func (h *MaintainerActivityHandler) isAwardEmojiRecentWithUser(a *glAwardEmojiJSON) bool {
	return a.CreatedAt != nil && a.CreatedAt.After(h.cutoff) && a.User.Username != ""
}

// isProjectEventRecentWithAuthor checks if a project event is recent and has a valid author.
func (h *MaintainerActivityHandler) isProjectEventRecentWithAuthor(e *glProjectEventJSON) bool {
	return e.CreatedAt != nil && e.CreatedAt.After(h.cutoff) && e.AuthorUsername != ""
}

////////////////////////////////////////////////////////////////////////////////
// SECTION A: ELEVATED USER DISCOVERY
////////////////////////////////////////////////////////////////////////////////
// Identifies users with elevated repository permissions (Developer, Maintainer, Owner).
// Uses ListAllProjectMembers which includes both direct members and inherited
// permissions from parent groups.

// loadElevatedMembers fetches all project members and filters those meeting the
// minimum access level threshold (Developer+ by default).
func (h *MaintainerActivityHandler) loadElevatedMembers() error {
	opt := &gitlab.ListProjectMembersOptions{
		ListOptions: gitlab.ListOptions{PerPage: int64(h.pageSize), Page: 1},
	}
	for {
		members, resp, err := h.gl.ProjectMembers.ListAllProjectMembers(h.projectID, opt)
		if err != nil {
			return fmt.Errorf("list project members: %w", err)
		}
		for _, m := range members {
			if m == nil || strings.TrimSpace(m.Username) == "" {
				continue
			}
			if m.AccessLevel >= h.minAccessLevel {
				h.elevated[strings.ToLower(m.Username)] = struct{}{}
			}
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// SECTION B: ACTIVITY COLLECTION
////////////////////////////////////////////////////////////////////////////////
// Collects maintainer activity signals from multiple GitLab API sources.
// Implements early termination: stops checking once all maintainers are confirmed active.
//
// SIGNAL PRIORITY
// ---------------
// Primary signals (critical errors halt execution):
//   B1: Merge Request merges - direct code contributions
//   B2: Release authorship - version management
//
// Secondary signals (best-effort, permissions may limit access):
//   B3: Audit events - admin/settings changes
//   B4: Manual CI/CD jobs - pipeline interactions
//   B5: Pipeline schedules - automation management
//
// Extended signals (comprehensive coverage, non-critical):
//   B6: User contribution events - pushes, comments
//   B7: Resource label events - issue/MR triaging
//   B8: Resource state/milestone events - project management
//   B9-B11: Award emoji - engagement signals
//   B12: Project events - wiki edits and misc. activities

// collectActivity orchestrates signal collection.
func (h *MaintainerActivityHandler) collectActivity() error {
	// === PRIMARY SIGNALS (critical) ===

	// B1: MR merges - direct code contribution
	if err := h.markActiveFromMRsAndApprovals(); err != nil {
		return err
	}
	if h.allActive() {
		return nil // All active - skip remaining checks
	}

	// B2: Release authorship - version mgmt
	if err := h.markActiveFromReleases(); err != nil {
		return err
	}
	if h.allActive() {
		return nil
	}

	if err := h.markActiveFromSnippetAwards(context.Background()); err != nil {
		return err
	}
	if h.allActive() {
		return nil
	}

	// === SECONDARY SIGNALS (best-effort, may require special permissions) ===

	// B3: Audit events - admin changes
	h.markActiveFromAuditEvents()
	if h.allActive() {
		return nil
	}

	// B4: Manual CI/CD jobs
	h.markActiveFromManualJobs()
	if h.allActive() {
		return nil
	}

	// B5: Pipeline schedules - tracks users managing automated pipelines
	h.markActiveFromPipelineSchedules()
	if h.allActive() {
		return nil
	}

	// === EXTENDED SIGNALS (comprehensive coverage via raw REST API) ===

	// B6: User contribution events - pushes, comments scoped to this project
	h.markActiveFromUserContributionEvents()
	if h.allActive() {
		return nil
	}
	// B7: Resource label events - triaging via labels on issues/MRs
	h.markActiveFromResourceLabelEvents()
	if h.allActive() {
		return nil
	}
	// B8: Resource state/milestone events - issue/MR state management
	h.markActiveFromResourceStateMilestone()
	if h.allActive() {
		return nil
	}
	// B9: Award emoji on issues/MRs - engagement through reactions
	h.markActiveFromAwardEmojiIssuesMRs()
	if h.allActive() {
		return nil
	}
	// B10: Award emoji on commits - code review engagement
	h.markActiveFromAwardEmojiCommits()
	if h.allActive() {
		return nil
	}
	// B11: Award emoji on snippets - snippet engagement
	h.markActiveFromAwardEmojiSnippets()
	if h.allActive() {
		return nil
	}
	// B12: Project events fallback - wiki edits and misc. activities
	h.markActiveFromProjectEventsFallback()

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// SECTION B1: MERGE REQUEST ACTIVITY
////////////////////////////////////////////////////////////////////////////////
// Tracks users who merged MRs within the activity window.
// Note: Approvals are NOT counted due to missing timestamps in the API response,
// which would cause false positives (can't verify approval happened after cutoff).

// markActiveFromMRsAndApprovals marks from MRs.
// Only considers MRs updated after cutoff.
func (h *MaintainerActivityHandler) markActiveFromMRsAndApprovals() error {
	// List MRs updated after the cutoff to constrain the result set.
	opts := &gitlab.ListProjectMergeRequestsOptions{
		UpdatedAfter: &h.cutoff,
		// If you want to include all states explicitly, you can set:
		// State: gitlab.String("all"),
		// but it's optional for our purposes here.
	}

	// NOTE: h.projectID is the string project identifier set in init(...).
	mrs, _, err := h.gl.MergeRequests.ListProjectMergeRequests(h.projectID, opts)
	if err != nil {
		return fmt.Errorf("list merge requests: %w", err)
	}

	for _, mr := range mrs {
		// Mark merger active only if merged_at is after cutoff.
		if h.isMergeRequestMergedRecently(mr) {
			if mr.MergeUser != nil && mr.MergeUser.Username != "" {
				h.markActive(mr.MergeUser.Username)
			}
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// SECTION B2: RELEASE AUTHORSHIP
////////////////////////////////////////////////////////////////////////////////
// Tracks users who authored releases (indicates version management activity).
// Releases are ordered by release date (descending) for efficient processing.

// markActiveFromReleases marks from releases.
func (h *MaintainerActivityHandler) markActiveFromReleases() error {
	order := "released_at"
	sort := "desc"
	opt := &gitlab.ListReleasesOptions{
		ListOptions: gitlab.ListOptions{PerPage: int64(h.pageSize), Page: 1},
		OrderBy:     &order,
		Sort:        &sort,
	}
	for {
		rels, resp, err := h.gl.Releases.ListReleases(h.projectID, opt)
		if err != nil {
			return fmt.Errorf("list releases: %w", err)
		}
		for _, r := range rels {
			if r == nil || r.Author.Username == "" {
				continue
			}
			var t time.Time
			if r.ReleasedAt != nil {
				t = *r.ReleasedAt
			} else if r.CreatedAt != nil {
				t = *r.CreatedAt
			}
			if !t.IsZero() && t.After(h.cutoff) {
				h.markActive(r.Author.Username)
			}
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// B3) Project Audit Events (admin/permissions/settings) – best-effort
////////////////////////////////////////////////////////////////////////////////

func (h *MaintainerActivityHandler) markActiveFromAuditEvents() {
	// NOTE: the SDK exposes ListAuditEventsOptions and AuditEvent.Details.AuthorName.
	opt := &gitlab.ListAuditEventsOptions{
		ListOptions:  gitlab.ListOptions{PerPage: int64(h.pageSize), Page: 1},
		CreatedAfter: &h.cutoff,
	}
	for {
		evs, resp, err := h.gl.AuditEvents.ListProjectAuditEvents(h.projectID, opt)
		if err != nil {
			break // unavailable or insufficient perms -> non-fatal
		}
		for _, e := range evs {
			if e == nil || e.CreatedAt == nil {
				continue
			}
			if e.CreatedAt.After(h.cutoff) && e.Details.AuthorName != "" {
				h.markActive(e.Details.AuthorName)
			}
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
}

////////////////////////////////////////////////////////////////////////////////
// B4) CI/CD manual jobs (jobs "played" by a human) – best-effort
////////////////////////////////////////////////////////////////////////////////

//nolint:gocognit // complex API logic
func (h *MaintainerActivityHandler) markActiveFromManualJobs() {
	order := "updated_at //nolint:goconst // API field name"
	popt := &gitlab.ListProjectPipelinesOptions{
		ListOptions:  gitlab.ListOptions{PerPage: 50, Page: 1},
		UpdatedAfter: &h.cutoff,
		OrderBy:      &order,
	}
	for {
		if h.allActive() {
			return // All maintainers active
		}
		pipes, resp, err := h.gl.Pipelines.ListProjectPipelines(h.projectID, popt)
		if err != nil {
			break
		}
		for _, p := range pipes {
			if p == nil {
				continue
			}
			// Scope uses []BuildStateValue (e.g., manual,running,success,...)
			scope := []gitlab.BuildStateValue{
				gitlab.Manual, gitlab.Running, gitlab.Success, gitlab.Failed,
				gitlab.Canceled, gitlab.Skipped, gitlab.Pending,
			}
			includeRetried := true
			jopt := &gitlab.ListJobsOptions{
				ListOptions:    gitlab.ListOptions{PerPage: 100, Page: 1},
				Scope:          &scope,
				IncludeRetried: &includeRetried,
			}
			for {
				jobs, jresp, err := h.gl.Jobs.ListPipelineJobs(h.projectID, p.ID, jopt)
				if err != nil {
					break
				}
				for _, j := range jobs {
					if !isManualJobValid(j) {
						continue
					}
					// Human activity if user and time within window.
					var t time.Time
					switch {
					case j.StartedAt != nil:
						t = *j.StartedAt
					case j.FinishedAt != nil:
						t = *j.FinishedAt
					case j.CreatedAt != nil:
						t = *j.CreatedAt
					}
					if !t.IsZero() && t.After(h.cutoff) {
						h.markActive(j.User.Username)
						if h.allActive() {
							return // All maintainers active
						}
					}
				}
				if jresp == nil || jresp.NextPage == 0 {
					break
				}
				jopt.Page = jresp.NextPage
			}
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		popt.Page = resp.NextPage
	}
}

////////////////////////////////////////////////////////////////////////////////
// B5) Pipeline schedules admin (owner/timestamps) – best-effort
////////////////////////////////////////////////////////////////////////////////

func (h *MaintainerActivityHandler) markActiveFromPipelineSchedules() {
	opt := &gitlab.ListPipelineSchedulesOptions{
		ListOptions: gitlab.ListOptions{PerPage: 50, Page: 1},
	}
	for {
		scheds, resp, err := h.gl.PipelineSchedules.ListPipelineSchedules(h.projectID, opt)
		if err != nil {
			break
		}
		for _, s := range scheds {
			if s == nil {
				continue
			}
			if s.Owner != nil && s.Owner.Username != "" {
				updatedRecent := s.UpdatedAt != nil && s.UpdatedAt.After(h.cutoff)
				createdRecent := s.CreatedAt != nil && s.CreatedAt.After(h.cutoff)
				if updatedRecent || createdRecent {
					h.markActive(s.Owner.Username)
				}
			}
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
}

////////////////////////////////////////////////////////////////////////////////
// RAW REST below (via client.NewRequest/Do) to keep full coverage independent
// of exact client-go types. We define local thin types per endpoint.
////////////////////////////////////////////////////////////////////////////////

// ---------- common paging options ----------

type qPage struct {
	Page    int `url:"page,omitempty"`
	PerPage int `url:"per_page,omitempty"`
}
type qAfter struct {
	After string `url:"after,omitempty"`
	qPage
}

// ---------- minimal issue struct + lister (avoid SDK unmarshaler) ----------

type issueLite struct {
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	IID       int        `json:"iid"`
}

type qListIssues struct {
	UpdatedAfter string `url:"updated_after,omitempty"`
	Scope        string `url:"scope,omitempty"`
	OrderBy      string `url:"order_by,omitempty"`
	qPage
}

func (h *MaintainerActivityHandler) listProjectIssuesUpdated(
	after time.Time,
	page, perPage int,
) ([]issueLite, *gitlab.Response, error) {
	opt := &qListIssues{
		qPage:        qPage{Page: page, PerPage: perPage},
		UpdatedAfter: after.Format(time.RFC3339),
		Scope:        "all",
		OrderBy:      "updated_at",
	}
	var items []issueLite
	path := fmt.Sprintf(apiPathProjectIssues, h.projectID)
	req, err := h.gl.NewRequest(
		http.MethodGet,
		path,
		opt,
		nil,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("create request: %w", err)
	}
	resp, err := h.gl.Do(req, &items)
	if err != nil {
		return nil, resp, fmt.Errorf("execute request: %w", err)
	}
	return items, resp, nil
}

// ---------- B6) user contribution events (pushes/comments/etc) ----------

type glUserEventJSON struct {
	CreatedAt *time.Time `json:"created_at"`
	ProjectID int        `json:"project_id"`
}

func (h *MaintainerActivityHandler) listUserByUsername(username string) (int, error) {
	type qUsers struct {
		Username string `url:"username,omitempty"`
	}
	var users []struct {
		Username string `json:"username"`
		ID       int    `json:"id"`
	}
	opt := &qUsers{Username: username}
	req, err := h.gl.NewRequest(
		http.MethodGet,
		"users",
		opt,
		nil,
	)
	if err != nil {
		return 0, fmt.Errorf("create user search request: %w", err)
	}
	resp, err := h.gl.Do(req, &users)
	if err != nil {
		return 0, fmt.Errorf("execute user search: %w", err)
	}
	if resp == nil || len(users) == 0 {
		return 0, fmt.Errorf("%w: %s", errUserNotFound, username)
	}
	return users[0].ID, nil
}

//nolint:gocognit // Nested pagination loops for GitLab API integration
func (h *MaintainerActivityHandler) markActiveFromUserContributionEvents() {
	after := h.cutoff.Format(time.RFC3339)
	for u := range h.elevated {
		if h.allActive() {
			return // All maintainers active - skip remaining users
		}
		uid, err := h.listUserByUsername(u)
		if err != nil {
			continue
		}
		opt := &qAfter{qPage: qPage{PerPage: 100, Page: 1}, After: after}
		for {
			var evs []glUserEventJSON
			path := fmt.Sprintf("users/%d/events", uid)
			req, err := h.gl.NewRequest(
				http.MethodGet,
				path,
				opt,
				nil,
			)
			if err != nil {
				break
			}
			resp, err := h.gl.Do(req, &evs)
			if err != nil {
				break
			}
			for _, e := range evs {
				if e.CreatedAt == nil {
					continue
				}
				if h.isUserEventRecentForProject(&e) {
					h.markActive(u)
					if h.allActive() {
						return // All maintainers active
					}
					break
				}
			}
			if resp == nil || resp.NextPage == 0 {
				break
			}
			opt.Page = int(resp.NextPage)
		}
	}
}

// ---------- B7) resource label events (issues & MRs) ----------

type glLabelEventJSON struct {
	CreatedAt *time.Time `json:"created_at"`
	User      struct {
		Username string `json:"username"`
	} `json:"user"`
}

//nolint:gocognit // complex API logic
func (h *MaintainerActivityHandler) markActiveFromResourceLabelEvents() {
	// issues: use our raw issue lister to avoid SDK unmarshal
	page := 1
	for {
		if h.allActive() {
			return // All maintainers active - skip remaining issues
		}
		issues, resp, err := h.listProjectIssuesUpdated(h.cutoff, page, 50)
		if err != nil {
			break
		}
		for _, is := range issues {
			opt := &qPage{PerPage: 100, Page: 1}
			for {
				var evs []glLabelEventJSON
				path := fmt.Sprintf(
					apiPathIssueResourceLabelEvents,
					h.projectID,
					is.IID,
				)
				req, err := h.gl.NewRequest(
					http.MethodGet,
					path,
					opt,
					nil,
				)
				if err != nil {
					break
				}
				resp2, err := h.gl.Do(req, &evs)
				if err != nil {
					break
				}
				for _, e := range evs {
					if h.isLabelEventRecentWithUser(&e) {
						h.markActive(e.User.Username)
					}
				}
				if resp2 == nil || resp2.NextPage == 0 {
					break
				}
				opt.Page = int(resp2.NextPage)
			}
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		page = int(resp.NextPage)
	}

	// merge requests: typed lister is OK
	scope := "all" //nolint:goconst // API parameter
	order := sortOrderUpdatedAt
	mrOpt := &gitlab.ListProjectMergeRequestsOptions{
		ListOptions:  gitlab.ListOptions{PerPage: 50, Page: 1},
		Scope:        &scope,
		UpdatedAfter: &h.cutoff,
		OrderBy:      &order,
	}
	for {
		if h.allActive() {
			return // All maintainers active - skip remaining MRs
		}
		mrs, resp, err := h.gl.MergeRequests.ListProjectMergeRequests(h.projectID, mrOpt)
		if err != nil {
			break
		}
		for _, mr := range mrs {
			opt := &qPage{PerPage: 100, Page: 1}
			for {
				var evs []glLabelEventJSON
				path := h.buildAPIPath(
					apiPathMRResourceLabelEvents,
					h.projectID,
					mr.IID,
				)
				req, err := h.gl.NewRequest(
					http.MethodGet,
					path,
					opt,
					nil,
				)
				if err != nil {
					break
				}
				resp2, err := h.gl.Do(req, &evs)
				if err != nil {
					break
				}
				for _, e := range evs {
					if h.isLabelEventRecentWithUser(&e) {
						h.markActive(e.User.Username)
						if h.allActive() {
							return // All maintainers active
						}
					}
				}
				if resp2 == nil || resp2.NextPage == 0 {
					break
				}
				opt.Page = int(resp2.NextPage)
			}
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		mrOpt.Page = resp.NextPage
	}
}

// ---------- B8) resource state & milestone events (issues & MRs) ----------

type glStateEventJSON struct {
	User struct {
		Username string `json:"username"`
	} `json:"user"`
	CreatedAt *time.Time `json:"created_at"`
	State     string     `json:"state"`
}
type glMilestoneEventJSON struct {
	CreatedAt *time.Time `json:"created_at"`
	User      struct {
		Username string `json:"username"`
	} `json:"user"`
}

//nolint:gocognit,gocyclo // Nested pagination loops for GitLab API integration
func (h *MaintainerActivityHandler) markActiveFromResourceStateMilestone() {
	// issues: use raw lister
	page := 1
	for {
		if h.allActive() {
			return // All maintainers active - skip remaining issues
		}
		issues, resp, err := h.listProjectIssuesUpdated(h.cutoff, page, 50)
		if err != nil {
			break
		}
		for _, is := range issues {
			// state events
			opt := &qPage{PerPage: 100, Page: 1}
			for {
				var evs []glStateEventJSON
				path := h.buildAPIPath(
					apiPathIssueResourceStateEvents,
					h.projectID,
					is.IID,
				)
				req, err := h.gl.NewRequest(
					http.MethodGet,
					path,
					opt,
					nil,
				)
				if err != nil {
					break
				}
				resp2, err := h.gl.Do(req, &evs)
				if err != nil {
					break
				}
				for _, e := range evs {
					if h.isStateEventRecentWithUser(&e) {
						h.markActive(e.User.Username)
						if h.allActive() {
							return // All maintainers active
						}
					}
				}
				if resp2 == nil || resp2.NextPage == 0 {
					break
				}
				opt.Page = int(resp2.NextPage)
			}
			// milestone events
			opt = &qPage{PerPage: 100, Page: 1}
			for {
				var evs []glMilestoneEventJSON
				path := fmt.Sprintf(
					apiPathIssueResourceMilestoneEvents,
					h.projectID,
					is.IID,
				)
				req, err := h.gl.NewRequest(
					http.MethodGet,
					path,
					opt,
					nil,
				)
				if err != nil {
					break
				}
				resp2, err := h.gl.Do(req, &evs)
				if err != nil {
					break
				}
				for _, e := range evs {
					if h.isMilestoneEventRecentWithUser(&e) {
						h.markActive(e.User.Username)
						if h.allActive() {
							return // All maintainers active
						}
					}
				}
				if resp2 == nil || resp2.NextPage == 0 {
					break
				}
				opt.Page = int(resp2.NextPage)
			}
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		page = int(resp.NextPage)
	}

	// merge requests (typed lister is OK)
	scope := "all"
	order := sortOrderUpdatedAt
	mrOpt := &gitlab.ListProjectMergeRequestsOptions{
		ListOptions:  gitlab.ListOptions{PerPage: 50, Page: 1},
		Scope:        &scope,
		UpdatedAfter: &h.cutoff,
		OrderBy:      &order,
	}
	for {
		if h.allActive() {
			return // All maintainers active - skip remaining MRs
		}
		mrs, resp, err := h.gl.MergeRequests.ListProjectMergeRequests(h.projectID, mrOpt)
		if err != nil {
			break
		}
		for _, mr := range mrs {
			// state
			opt := &qPage{PerPage: 100, Page: 1}
			for {
				var evs []glStateEventJSON
				path := h.buildAPIPath(
					apiPathMRResourceStateEvents,
					h.projectID,
					mr.IID,
				)
				req, err := h.gl.NewRequest(
					http.MethodGet,
					path,
					opt,
					nil,
				)
				if err != nil {
					break
				}
				resp2, err := h.gl.Do(req, &evs)
				if err != nil {
					break
				}
				for _, e := range evs {
					if h.isStateEventRecentWithUser(&e) {
						h.markActive(e.User.Username)
						if h.allActive() {
							return // All maintainers active
						}
					}
				}
				if resp2 == nil || resp2.NextPage == 0 {
					break
				}
				opt.Page = int(resp2.NextPage)
			}
			// milestone
			opt = &qPage{PerPage: 100, Page: 1}
			for {
				var evs []glMilestoneEventJSON
				path := h.buildAPIPath(
					apiPathMRResourceMilestoneEvents,
					h.projectID,
					mr.IID,
				)
				req, err := h.gl.NewRequest(
					http.MethodGet,
					path,
					opt,
					nil,
				)
				if err != nil {
					break
				}
				resp2, err := h.gl.Do(req, &evs)
				if err != nil {
					break
				}
				for _, e := range evs {
					if h.isMilestoneEventRecentWithUser(&e) {
						h.markActive(e.User.Username)
						if h.allActive() {
							return // All maintainers active
						}
					}
				}
				if resp2 == nil || resp2.NextPage == 0 {
					break
				}
				opt.Page = int(resp2.NextPage)
			}
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		mrOpt.Page = resp.NextPage
	}
}

// ---------- B9) award emoji (issues/MRs/commits/snippets) ----------

type glAwardEmojiJSON struct {
	CreatedAt *time.Time `json:"created_at"`
	User      struct {
		Username string `json:"username"`
	} `json:"user"`
}

//nolint:gocognit // Nested pagination loops for GitLab API integration
func (h *MaintainerActivityHandler) markActiveFromAwardEmojiIssuesMRs() {
	// issues (use raw issue lister for pagination)
	page := 1
	for {
		issues, resp, err := h.listProjectIssuesUpdated(h.cutoff, page, 50)
		if err != nil {
			break
		}
		for _, is := range issues {
			opt := &qPage{PerPage: 100, Page: 1}
			for {
				var emj []glAwardEmojiJSON
				path := fmt.Sprintf(
					apiPathIssueAwardEmoji,
					h.projectID,
					is.IID,
				)
				req, err := h.gl.NewRequest(
					http.MethodGet,
					path,
					opt,
					nil,
				)
				if err != nil {
					break
				}
				resp2, err := h.gl.Do(req, &emj)
				if err != nil {
					break
				}
				for _, a := range emj {
					if h.isAwardEmojiRecentWithUser(&a) {
						h.markActive(a.User.Username)
					}
				}
				if resp2 == nil || resp2.NextPage == 0 {
					break
				}
				opt.Page = int(resp2.NextPage)
			}
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		page = int(resp.NextPage)
	}

	// MRs
	scope := "all"
	order := sortOrderUpdatedAt
	mrOpt := &gitlab.ListProjectMergeRequestsOptions{
		ListOptions:  gitlab.ListOptions{PerPage: 50, Page: 1},
		Scope:        &scope,
		UpdatedAfter: &h.cutoff,
		OrderBy:      &order,
	}
	for {
		mrs, resp, err := h.gl.MergeRequests.ListProjectMergeRequests(h.projectID, mrOpt)
		if err != nil {
			break
		}
		for _, mr := range mrs {
			opt := &qPage{PerPage: 100, Page: 1}
			for {
				var emj []glAwardEmojiJSON
				path := fmt.Sprintf(
					apiPathMRAwardEmoji,
					h.projectID,
					mr.IID,
				)
				req, err := h.gl.NewRequest(
					http.MethodGet,
					path,
					opt,
					nil,
				)
				if err != nil {
					break
				}
				resp2, err := h.gl.Do(req, &emj)
				if err != nil {
					break
				}
				for _, a := range emj {
					if h.isAwardEmojiRecentWithUser(&a) {
						h.markActive(a.User.Username)
					}
				}
				if resp2 == nil || resp2.NextPage == 0 {
					break
				}
				opt.Page = int(resp2.NextPage)
			}
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		mrOpt.Page = resp.NextPage
	}
}

//nolint:gocognit // Nested pagination loops for GitLab API integration
func (h *MaintainerActivityHandler) markActiveFromAwardEmojiCommits() {
	scope := "all"
	order := sortOrderUpdatedAt
	mrOpt := &gitlab.ListProjectMergeRequestsOptions{
		ListOptions:  gitlab.ListOptions{PerPage: 50, Page: 1},
		Scope:        &scope,
		UpdatedAfter: &h.cutoff,
		OrderBy:      &order,
	}
	for {
		mrs, resp, err := h.gl.MergeRequests.ListProjectMergeRequests(h.projectID, mrOpt)
		if err != nil {
			break
		}
		for _, mr := range mrs {
			sha := ""
			if mr != nil && mr.SHA != "" {
				sha = mr.SHA
			} else if mr != nil && mr.MergeCommitSHA != "" {
				sha = mr.MergeCommitSHA
			}
			if sha == "" {
				continue
			}
			opt := &qPage{PerPage: 100, Page: 1}
			for {
				var emj []glAwardEmojiJSON
				path := fmt.Sprintf(
					apiPathCommitAwardEmoji,
					h.projectID,
					sha,
				)
				req, err := h.gl.NewRequest(
					http.MethodGet,
					path,
					opt,
					nil,
				)
				if err != nil {
					break
				}
				resp2, err := h.gl.Do(req, &emj)
				if err != nil {
					break
				}
				for _, a := range emj {
					if h.isAwardEmojiRecentWithUser(&a) {
						h.markActive(a.User.Username)
					}
				}
				if resp2 == nil || resp2.NextPage == 0 {
					break
				}
				opt.Page = int(resp2.NextPage)
			}
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		mrOpt.Page = resp.NextPage
	}
}

//nolint:gocognit // Nested pagination loops for GitLab API integration
func (h *MaintainerActivityHandler) markActiveFromAwardEmojiSnippets() {
	// Use raw REST for listing project snippets, then fetch award emoji per snippet.
	opt := &qPage{PerPage: 50, Page: 1}
	type snippetJSON struct {
		UpdatedAt *time.Time `json:"updated_at"`
		ID        int        `json:"id"`
	}
	for {
		var snips []snippetJSON
		path := fmt.Sprintf(apiPathSnippets, h.projectID)
		req, err := h.gl.NewRequest(
			http.MethodGet,
			path,
			opt,
			nil,
		)
		if err != nil {
			break
		}
		resp, err := h.gl.Do(req, &snips)
		if err != nil {
			break
		}
		for _, s := range snips {
			// only bother if updated in/after window to minimize calls
			if s.UpdatedAt != nil && s.UpdatedAt.Before(h.cutoff) {
				continue
			}
			p := &qPage{PerPage: 100, Page: 1}
			for {
				var emj []glAwardEmojiJSON
				ep := fmt.Sprintf(
					apiPathSnippetAwardEmoji,
					h.projectID,
					s.ID,
				)
				req2, err := h.gl.NewRequest(
					http.MethodGet,
					ep,
					p,
					nil,
				)
				if err != nil {
					break
				}
				resp2, err := h.gl.Do(req2, &emj)
				if err != nil {
					break
				}
				for _, a := range emj {
					if h.isAwardEmojiRecentWithUser(&a) {
						h.markActive(a.User.Username)
					}
				}
				if resp2 == nil || resp2.NextPage == 0 {
					break
				}
				p.Page = int(resp2.NextPage)
			}
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		opt.Page = int(resp.NextPage)
	}
}

// ---------- B10) project events fallback (wiki edits etc.) ----------

type glProjectEventJSON struct {
	AuthorUsername string     `json:"author_username"`
	CreatedAt      *time.Time `json:"created_at"`
	ActionName     string     `json:"action_name"`
}

func (h *MaintainerActivityHandler) markActiveFromProjectEventsFallback() {
	after := h.cutoff.Format(time.RFC3339)
	opt := &qAfter{qPage: qPage{PerPage: h.pageSize, Page: 1}, After: after}
	for {
		var evs []glProjectEventJSON
		path := fmt.Sprintf(apiPathEvents, h.projectID)
		req, err := h.gl.NewRequest(
			http.MethodGet,
			path,
			opt,
			nil,
		)
		if err != nil {
			break
		}
		resp, err := h.gl.Do(req, &evs)
		if err != nil {
			break
		}
		for _, e := range evs {
			if h.isProjectEventRecentWithAuthor(&e) {
				h.markActive(e.AuthorUsername)
			}
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		opt.Page = int(resp.NextPage)
	}
}

// markActiveFromSnippetAwards checks award emoji on project snippets and marks
// the awarding maintainer as active if any award is within the cutoff window.
func (h *MaintainerActivityHandler) markActiveFromSnippetAwards(ctx context.Context) error {
	// Correct service: ProjectSnippets (not Projects).
	// Options are optional; pass nil to fetch with defaults.
	snippets, _, err := h.gl.ProjectSnippets.ListSnippets(h.projectID, nil)
	if err != nil {
		return fmt.Errorf("list snippets: %w", err)
	}

	for _, sn := range snippets {
		awards, _, err := h.gl.AwardEmoji.ListSnippetAwardEmoji(h.projectID, sn.ID, nil)
		if err != nil {
			// Be resilient: skip this snippet on error.
			continue
		}
		for _, aw := range awards {
			if aw == nil || aw.CreatedAt == nil {
				continue
			}
			// In this client version, User is a value struct, not a pointer.
			u := aw.User.Username
			if u == "" {
				continue
			}
			// Only flip to true for users we’re already tracking as elevated.
			if _, ok := h.active[u]; ok && aw.CreatedAt.After(h.cutoff) {
				h.active[u] = true
			}
		}
	}
	return nil
}
