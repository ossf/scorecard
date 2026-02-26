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

//nolint:gocognit,paralleltest // Test mocks for GitLab maintainer activity
package gitlabrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

//
// High-level, readable “story” tests for maintainer activity on GitLab
// -------------------------------------------------------------------
// Each scenario is a sequence of human-readable steps (≥5, some up to ~20)
// that we both display and apply to a fake GitLab server. We then assert the
// expected active/inactive status for maintainers within the last 6 months.
//

//nolint:paralleltest // integration test
func TestInactiveMaintainers_GitLab_Storybook(t *testing.T) {
	now := time.Date(2025, 8, 1, 12, 0, 0, 0, time.UTC)
	cutoff := now.AddDate(0, -6, 0) // “last 6 months”

	type actors struct{ MaintA, MaintB, ContribA string }
	A := actors{MaintA: "maint-a", MaintB: "maint-b", ContribA: "contrib-a"}

	makeClient := func(t *testing.T, story *GLStory) *MaintainerActivityHandler {
		t.Helper()
		srv := httptest.NewServer(story.Handler())
		t.Cleanup(srv.Close)
		gl, err := gitlab.NewClient("x", gitlab.WithBaseURL(srv.URL+"/api/v4"))
		if err != nil {
			t.Fatalf("gitlab client: %v", err)
		}
		h := &MaintainerActivityHandler{}
		h.init(context.Background(), gl, strconv.Itoa(story.ProjectID))
		h.setCutoff(cutoff)
		return h
	}

	scenarios := []struct {
		expect      map[string]bool
		name        string
		description string
		steps       []StoryStep
	}{
		// ---------------------------------------------------------------------
		// EXISTING 4 STORY GROUPS (already have rich documentation)
		// ---------------------------------------------------------------------

		{
			name: "Approvals-old__merge-recent__issue-close-recent__MaintB-inactive",
			description: `
A multi-step timeline where approvals are old, but merge and an issue close are recent:

1) Contributor A creates a merge request 8 months ago.
2) Maintainer A approves it 7.5 months ago.
3) Maintainer B approves it 7 months ago.
4) Maintainer A merges the MR 5 months ago.
5) Maintainer A closes an issue 4 months ago.

Expected: Maintainer A active, Maintainer B inactive (no activity in last 6 months).
`,
			steps: []StoryStep{
				Seq(
					ProjectSetup("group/p", 10101),
					Maintainers(A.MaintA, A.MaintB),
					Contributors(A.ContribA),
				),
				MRCreated(A.ContribA, 1, monthsAgo(now, 8)),
				MRApproved(A.MaintA, 1, monthsAgo(now, 7.5)),
				MRApproved(A.MaintB, 1, monthsAgo(now, 7)),
				MRMerged(A.MaintA, 1, monthsAgo(now, 5)),
				IssueState(A.MaintA, 11, "closed", monthsAgo(now, 4)),
			},
			expect: map[string]bool{
				A.MaintA: true,  // merge + issue close are recent
				A.MaintB: false, // only old approval
			},
		},
		{
			name: "Deep-activity-mix__only-MaintB-recent",
			description: `
A long, readable story (≈15 steps). MaintA is busy earlier, MaintB does one recent action.

1) Contributor A opens MR#2 10 months ago.
2) MaintA comments on the MR 9.5 months ago (counts via user events).
3) MaintA adds label "triage" to Issue#22 9 months ago.
4) MaintA sets milestone on Issue#22 8.5 months ago.
5) MaintA approves MR#2 8 months ago.
6) MaintA pushes to default branch 7.5 months ago (user event).
7) MaintB sets pipeline schedule 7 months ago (createdAt 7m).
8) MaintA reacts 👍 on Issue#22 6.5 months ago (just outside window).
9) MaintA updates project wiki 6.2 months ago (outside window).
10) MaintA releases v1.2.3 6.1 months ago (outside window).
11) MaintB plays a manual job 5.9 months ago (inside window).
12) MaintB adds label "ready" to MR#2 5.8 months ago.
13) MaintB reacts 👍 on MR#2 5.7 months ago.
14) MaintB closes Issue#22 5.6 months ago.
15) MaintB adds milestone to MR#2 5.5 months ago.

Expected: MaintB active, MaintA inactive (all of MaintA’s activity is older than 6 months).
`,
			steps: []StoryStep{
				Seq(
					ProjectSetup("group/p", 20202),
					Maintainers(A.MaintA, A.MaintB),
					Contributors(A.ContribA),
				),
				MRCreated(A.ContribA, 2, monthsAgo(now, 10)),
				UserEvent(A.MaintA, "comment-on-MR#2", monthsAgo(now, 9.5)),
				IssueLabeled(A.MaintA, 22, "triage", monthsAgo(now, 9)),
				IssueMilestoned(A.MaintA, 22, "M-1", monthsAgo(now, 8.5)),
				MRApproved(A.MaintA, 2, monthsAgo(now, 8)),
				UserEvent(A.MaintA, "push", monthsAgo(now, 7.5)),
				PipelineScheduleSet(A.MaintB, monthsAgo(now, 7)),
				IssueAward(A.MaintA, 22, monthsAgo(now, 6.5)),
				ProjectEvent(A.MaintA, "wiki_edit", monthsAgo(now, 6.2)),
				ReleaseAuthored(A.MaintA, "v1.2.3", monthsAgo(now, 6.1)),
				ManualJob(A.MaintB, monthsAgo(now, 5.9)),
				MRLabel(A.MaintB, 2, "ready", monthsAgo(now, 5.8)),
				MRAward(A.MaintB, 2, monthsAgo(now, 5.7)),
				IssueState(A.MaintB, 22, "closed", monthsAgo(now, 5.6)),
				MRMilestoned(A.MaintB, 2, "M-2", monthsAgo(now, 5.5)),
			},
			expect: map[string]bool{
				A.MaintA: false,
				A.MaintB: true,
			},
		},
		{
			name: "Both-active__rich-sequence__15-steps",
			description: `
Both maintainers perform clear in-window actions among many older ones.

1) Contributor A opens MR#3 12 months ago.
2) MaintA approves MR#3 11 months ago.
3) MaintB approves MR#3 10 months ago.
4) MaintA pushes 7 months ago (old).
5) MaintB wiki edit 6.2 months ago (old).
6) MaintA merges MR#3 5.9 months ago (recent).
7) MaintB labels Issue#33 5.8 months ago (recent).
8) MaintA milestones MR#3 5.7 months ago (recent).
9) MaintB plays manual job 5.6 months ago (recent).
10) MaintA reacts on MR#3 5.5 months ago (recent).
11) MaintB reacts on Issue#33 5.4 months ago (recent).
12) MaintA closes Issue#33 5.3 months ago (recent).
13) MaintB releases v2.0.0 5.2 months ago (recent).
14) MaintA milestones Issue#33 5.1 months ago (recent).
15) MaintB user event: comment on MR#3 5.0 months ago (recent).

Expected: both active.
`,
			steps: []StoryStep{
				Seq(
					ProjectSetup("group/p", 30303),
					Maintainers(A.MaintA, A.MaintB),
					Contributors(A.ContribA),
				),
				MRCreated(A.ContribA, 3, monthsAgo(now, 12)),
				MRApproved(A.MaintA, 3, monthsAgo(now, 11)),
				MRApproved(A.MaintB, 3, monthsAgo(now, 10)),
				UserEvent(A.MaintA, "push", monthsAgo(now, 7)),
				ProjectEvent(A.MaintB, "wiki_edit", monthsAgo(now, 6.2)),
				MRMerged(A.MaintA, 3, monthsAgo(now, 5.9)),
				IssueLabeled(A.MaintB, 33, "help wanted", monthsAgo(now, 5.8)),
				MRMilestoned(A.MaintA, 3, "M-3", monthsAgo(now, 5.7)),
				ManualJob(A.MaintB, monthsAgo(now, 5.6)),
				MRAward(A.MaintA, 3, monthsAgo(now, 5.5)),
				IssueAward(A.MaintB, 33, monthsAgo(now, 5.4)),
				IssueState(A.MaintA, 33, "closed", monthsAgo(now, 5.3)),
				ReleaseAuthored(A.MaintB, "v2.0.0", monthsAgo(now, 5.2)),
				IssueMilestoned(A.MaintA, 33, "M-33", monthsAgo(now, 5.1)),
				UserEvent(A.MaintB, "comment-on-MR#3", monthsAgo(now, 5.0)),
			},
			expect: map[string]bool{
				A.MaintA: true,
				A.MaintB: true,
			},
		},
		{
			name: "Lots-of-older-work__no-recent__both-inactive",
			description: `
A dense 20-step history, but everything is older than 6 months.

1) MR#4 is opened 18 months ago.
2) MaintA approves 17.5 months ago.
3) MaintB approves 17.4 months ago.
4) MR#4 triaged 17.3 months ago.
5) MR#4 milestone set 17.2 months ago.
6) MR#4 merged 17.0 months ago.
7) Issue#44 labeled 16.9 months ago.
8) Issue#44 closed 16.8 months ago.
9) Issue#44 milestone set 16.7 months ago.
10) Issue#44 reaction 16.6 months ago.
11) Project wiki edit 16.5 months ago.
12) Release v0.9.0 16.4 months ago.
13) Manual job played 16.3 months ago.
14) Push event 16.2 months ago.
15) Comment event 16.1 months ago.
16) Pipeline schedule updated 16.0 months ago.
17) Snippet created/updated 15.9 months ago.
18) Snippet reaction 15.8 months ago.
19) Commit reaction A 15.7 months ago.
20) Commit reaction B 15.6 months ago.

Expected: both inactive.
`,
			steps: []StoryStep{
				Seq(
					ProjectSetup("group/p", 40404),
					Maintainers(A.MaintA, A.MaintB),
					Contributors(A.ContribA),
				),
				MRCreated(A.ContribA, 4, monthsAgo(now, 18)),
				MRApproved(A.MaintA, 4, monthsAgo(now, 17.5)),
				MRApproved(A.MaintB, 4, monthsAgo(now, 17.4)),
				MRLabel(A.MaintA, 4, "triage", monthsAgo(now, 17.3)),
				MRMilestoned(A.MaintB, 4, "M-4", monthsAgo(now, 17.2)),
				MRMerged(A.MaintA, 4, monthsAgo(now, 17.0)),
				IssueLabeled(A.MaintB, 44, "bug", monthsAgo(now, 16.9)),
				IssueState(A.MaintA, 44, "closed", monthsAgo(now, 16.8)),
				IssueMilestoned(A.MaintB, 44, "M-44", monthsAgo(now, 16.7)),
				IssueAward(A.MaintA, 44, monthsAgo(now, 16.6)),
				ProjectEvent(A.MaintB, "wiki_edit", monthsAgo(now, 16.5)),
				ReleaseAuthored(A.MaintA, "v0.9.0", monthsAgo(now, 16.4)),
				ManualJob(A.MaintB, monthsAgo(now, 16.3)),
				UserEvent(A.MaintA, "push", monthsAgo(now, 16.2)),
				UserEvent(A.MaintB, "comment", monthsAgo(now, 16.1)),
				PipelineScheduleSet(A.MaintA, monthsAgo(now, 16.0)),
				Snippet(A.MaintB, 7001, monthsAgo(now, 15.9)),
				SnippetAward(A.MaintA, 7001, monthsAgo(now, 15.8)),
				CommitAward(A.MaintB, "deadbeef4444", monthsAgo(now, 15.7)),
				CommitAward(A.MaintA, "deadbeef5555", monthsAgo(now, 15.6)),
			},
			expect: map[string]bool{
				A.MaintA: false,
				A.MaintB: false,
			},
		},

		// ---------------------------------------------------------------------
		// NEW GROUPS: each now includes explicit, numbered documentation
		// ---------------------------------------------------------------------

		// 1) AUDIT EVENTS
		{
			name: "audit_events__inactive",
			description: `
A short history with only pre-window activity; no audit entries inside the 6-month window.

1) MR#501 is created 8.2 months ago by a contributor.
2) MaintB labels Issue#77 as "triage" 7.8 months ago.
3) MaintA performs a project audit activity (project_setting_change) 7.0 months ago.
4) MaintB edits the project wiki 6.3 months ago (outside the window).
5) MaintA closes Issue#77 6.1 months ago (outside the window).

Expected: both MaintA and MaintB are inactive (no in-window maintainer activity).
`,
			steps: []StoryStep{
				Seq(ProjectSetup("group/p", 51001), Maintainers(A.MaintA, A.MaintB), Contributors(A.ContribA)),
				MRCreated(A.ContribA, 501, monthsAgo(now, 8.2)),
				IssueLabeled(A.MaintB, 77, "triage", monthsAgo(now, 7.8)),
				AuditEvent(A.MaintA, "project_setting_change", monthsAgo(now, 7.0)),
				ProjectEvent(A.MaintB, "wiki_edit", monthsAgo(now, 6.3)),
				IssueState(A.MaintA, 77, "closed", monthsAgo(now, 6.1)),
			},
			expect: map[string]bool{A.MaintA: false, A.MaintB: false},
		},
		{
			name: "audit_events__active",
			description: `
MaintA performs an audit event inside the window while other noise is outside.

1) MR#502 is created 9.0 months ago (noise).
2) MaintB labels Issue#79 as "bug" 8.1 months ago (noise).
3) MaintB edits wiki 7.3 months ago (noise).
4) MaintA updates branch protection via audit event 5.9 months ago (in-window).
5) MaintB sets milestone on Issue#79 6.1 months ago (outside the window).

Expected: MaintA active, MaintB inactive.
`,
			steps: []StoryStep{
				Seq(ProjectSetup("group/p", 51002), Maintainers(A.MaintA, A.MaintB), Contributors(A.ContribA)),
				MRCreated(A.ContribA, 502, monthsAgo(now, 9.0)),
				IssueLabeled(A.MaintB, 79, "bug", monthsAgo(now, 8.1)),
				ProjectEvent(A.MaintB, "wiki_edit", monthsAgo(now, 7.3)),
				AuditEvent(A.MaintA, "branch_protection_updated", monthsAgo(now, 5.9)),
				IssueMilestoned(A.MaintB, 79, "M-79", monthsAgo(now, 6.1)),
			},
			expect: map[string]bool{A.MaintA: true, A.MaintB: false},
		},
		{
			name: "audit_events__mixed",
			description: `
MaintA has an in-window audit event; MaintB only has older activity.

1) MaintB edits the wiki 7.2 months ago (old).
2) MR#503 is created 8.0 months ago (noise).
3) MaintA adds a project member (audit event) 2.0 months ago (in-window).
4) MaintB labels Issue#81 "help wanted" 6.2 months ago (old).
5) MaintA user-event comment 6.1 months ago (old).

Expected: MaintA active, MaintB inactive.
`,
			steps: []StoryStep{
				Seq(ProjectSetup("group/p", 51003), Maintainers(A.MaintA, A.MaintB), Contributors(A.ContribA)),
				ProjectEvent(A.MaintB, "wiki_edit", monthsAgo(now, 7.2)),
				MRCreated(A.ContribA, 503, monthsAgo(now, 8.0)),
				AuditEvent(A.MaintA, "member_added", monthsAgo(now, 2.0)),
				IssueLabeled(A.MaintB, 81, "help wanted", monthsAgo(now, 6.2)),
				UserEvent(A.MaintA, "comment", monthsAgo(now, 6.1)),
			},
			expect: map[string]bool{A.MaintA: true, A.MaintB: false},
		},
		{
			name: "audit_events__pre_window_only",
			description: `
All audit and project-related activity occurs before the 6-month window.

1) MR#504 is created 12.0 months ago by a contributor.
2) MaintB assigns a project runner (audit event) 9.0 months ago.
3) MaintA closes Issue#84 7.1 months ago (old).
4) MaintA edits the wiki 6.2 months ago (old).
5) MaintB labels MR#504 "triage" 6.15 months ago (old).

Expected: both inactive.
`,
			steps: []StoryStep{
				Seq(ProjectSetup("group/p", 51004), Maintainers(A.MaintA, A.MaintB), Contributors(A.ContribA)),
				MRCreated(A.ContribA, 504, monthsAgo(now, 12.0)),
				AuditEvent(A.MaintB, "runner_assigned", monthsAgo(now, 9.0)),
				IssueState(A.MaintA, 84, "closed", monthsAgo(now, 7.1)),
				ProjectEvent(A.MaintA, "wiki_edit", monthsAgo(now, 6.2)),
				MRLabel(A.MaintB, 504, "triage", monthsAgo(now, 6.15)),
			},
			expect: map[string]bool{A.MaintA: false, A.MaintB: false},
		},

		// 2) MR STATE EVENTS (close/reopen etc., not merges)
		{
			name: "mr_state_events__inactive",
			description: `
MR state changes exist, but they’re all outside the 6-month window.

1) MR#521 is created 9.0 months ago.
2) MaintA closes MR#521 7.0 months ago (old).
3) MaintB labels Issue#90 "bug" 6.4 months ago (old).
4) MaintA labels MR#521 "needs review" 6.3 months ago (old).
5) MaintB sets milestone "M-90" on Issue#90 6.1 months ago (old).

Expected: both inactive.
`,
			steps: []StoryStep{
				Seq(ProjectSetup("group/p", 52001), Maintainers(A.MaintA, A.MaintB), Contributors(A.ContribA)),
				MRCreated(A.ContribA, 521, monthsAgo(now, 9.0)),
				MRState(A.MaintA, 521, "closed", monthsAgo(now, 7.0)),
				IssueLabeled(A.MaintB, 90, "bug", monthsAgo(now, 6.4)),
				MRLabel(A.MaintA, 521, "needs review", monthsAgo(now, 6.3)),
				IssueMilestoned(A.MaintB, 90, "M-90", monthsAgo(now, 6.1)),
			},
			expect: map[string]bool{A.MaintA: false, A.MaintB: false},
		},
		{
			name: "mr_state_events__active",
			description: `
MaintA performs an in-window MR state change; other events are old/noise.

1) MR#522 is created 7.5 months ago.
2) MaintB labels MR#522 "triage" 7.0 months ago (old).
3) MaintA closes MR#522 5.5 months ago (in-window).
4) MaintB labels Issue#91 "bug" 7.2 months ago (old).
5) MaintB closes Issue#91 6.2 months ago (old).

Expected: MaintA active, MaintB inactive.
`,
			steps: []StoryStep{
				Seq(ProjectSetup("group/p", 52002), Maintainers(A.MaintA, A.MaintB), Contributors(A.ContribA)),
				MRCreated(A.ContribA, 522, monthsAgo(now, 7.5)),
				MRLabel(A.MaintB, 522, "triage", monthsAgo(now, 7.0)),
				MRState(A.MaintA, 522, "closed", monthsAgo(now, 5.5)),
				IssueLabeled(A.MaintB, 91, "bug", monthsAgo(now, 7.2)),
				IssueState(A.MaintB, 91, "closed", monthsAgo(now, 6.2)),
			},
			expect: map[string]bool{A.MaintA: true, A.MaintB: false},
		},
		{
			name: "mr_state_events__mixed",
			description: `
Only MaintA has an in-window MR state event; MaintB’s actions are older.

1) MR#523 is created 8.0 months ago.
2) MaintB labels MR#523 "ready" 7.9 months ago (old).
3) MaintA reopens MR#523 5.8 months ago (in-window).
4) MaintB labels Issue#92 "help wanted" 6.3 months ago (old).
5) MaintB edits the wiki 6.2 months ago (old).

Expected: MaintA active, MaintB inactive.
`,
			steps: []StoryStep{
				Seq(ProjectSetup("group/p", 52003), Maintainers(A.MaintA, A.MaintB), Contributors(A.ContribA)),
				MRCreated(A.ContribA, 523, monthsAgo(now, 8.0)),
				MRLabel(A.MaintB, 523, "ready", monthsAgo(now, 7.9)),
				MRState(A.MaintA, 523, "reopened", monthsAgo(now, 5.8)),
				IssueLabeled(A.MaintB, 92, "help wanted", monthsAgo(now, 6.3)),
				ProjectEvent(A.MaintB, "wiki_edit", monthsAgo(now, 6.2)),
			},
			expect: map[string]bool{A.MaintA: true, A.MaintB: false},
		},
		{
			name: "mr_state_events__pre_window_only",
			description: `
There are MR state changes, but all are outside the window.

1) MR#524 is created 12.0 months ago.
2) MaintB closes MR#524 6.1 months ago (just outside).
3) MaintA labels MR#524 "needs review" 7.8 months ago.
4) MaintA labels Issue#93 "bug" 7.5 months ago.
5) MaintB closes Issue#93 7.2 months ago.

Expected: both inactive.
`,
			steps: []StoryStep{
				Seq(ProjectSetup("group/p", 52004), Maintainers(A.MaintA, A.MaintB), Contributors(A.ContribA)),
				MRCreated(A.ContribA, 524, monthsAgo(now, 12.0)),
				MRState(A.MaintB, 524, "closed", monthsAgo(now, 6.1)),
				MRLabel(A.MaintA, 524, "needs review", monthsAgo(now, 7.8)),
				IssueLabeled(A.MaintA, 93, "bug", monthsAgo(now, 7.5)),
				IssueState(A.MaintB, 93, "closed", monthsAgo(now, 7.2)),
			},
			expect: map[string]bool{A.MaintA: false, A.MaintB: false},
		},

		// 3) PROJECT-LEVEL EVENTS (fallback) IN-WINDOW
		{
			name: "project_events__inactive",
			description: `
Project-level events occur, but none inside the window.

1) MaintA edits the wiki 6.2 months ago (old).
2) MaintB labels Issue#110 "bug" 7.5 months ago (old).
3) MR#531 is created 8.4 months ago (noise).
4) MaintA labels MR#531 "triage" 7.9 months ago (old).
5) MaintB sets milestone M-110 on Issue#110 6.3 months ago (old).

Expected: both inactive.
`,
			steps: []StoryStep{
				Seq(ProjectSetup("group/p", 53001), Maintainers(A.MaintA, A.MaintB), Contributors(A.ContribA)),
				ProjectEvent(A.MaintA, "wiki_edit", monthsAgo(now, 6.2)),
				IssueLabeled(A.MaintB, 110, "bug", monthsAgo(now, 7.5)),
				MRCreated(A.ContribA, 531, monthsAgo(now, 8.4)),
				MRLabel(A.MaintA, 531, "triage", monthsAgo(now, 7.9)),
				IssueMilestoned(A.MaintB, 110, "M-110", monthsAgo(now, 6.3)),
			},
			expect: map[string]bool{A.MaintA: false, A.MaintB: false},
		},
		{
			name: "project_events__active",
			description: `
MaintA performs a project-level action (wiki edit) inside the window; MaintB is old/noise.

1) MR#532 is created 9.0 months ago (noise).
2) MaintB labels Issue#111 "help wanted" 7.2 months ago (old).
3) MaintA performs wiki edit 5.9 months ago (in-window).
4) MaintB labels MR#532 "needs work" 7.0 months ago (old).
5) MaintB closes Issue#111 6.1 months ago (old).

Expected: MaintA active, MaintB inactive.
`,
			steps: []StoryStep{
				Seq(ProjectSetup("group/p", 53002), Maintainers(A.MaintA, A.MaintB), Contributors(A.ContribA)),
				MRCreated(A.ContribA, 532, monthsAgo(now, 9.0)),
				IssueLabeled(A.MaintB, 111, "help wanted", monthsAgo(now, 7.2)),
				ProjectEvent(A.MaintA, "wiki_edit", monthsAgo(now, 5.9)),
				MRLabel(A.MaintB, 532, "needs work", monthsAgo(now, 7.0)),
				IssueState(A.MaintB, 111, "closed", monthsAgo(now, 6.1)),
			},
			expect: map[string]bool{A.MaintA: true, A.MaintB: false},
		},
		{
			name: "project_events__mixed",
			description: `
MaintA has one in-window project-level event; MaintB has none.

1) MaintA renames the project 4.0 months ago (in-window).
2) MR#533 is created 10.0 months ago.
3) MaintB labels MR#533 "triage" 7.0 months ago (old).
4) MaintB labels Issue#112 "bug" 6.2 months ago (old).
5) MaintA user-event comment 6.1 months ago (old).

Expected: MaintA active, MaintB inactive.
`,
			steps: []StoryStep{
				Seq(ProjectSetup("group/p", 53003), Maintainers(A.MaintA, A.MaintB), Contributors(A.ContribA)),
				ProjectEvent(A.MaintA, "project_renamed", monthsAgo(now, 4.0)),
				MRCreated(A.ContribA, 533, monthsAgo(now, 10.0)),
				MRLabel(A.MaintB, 533, "triage", monthsAgo(now, 7.0)),
				IssueLabeled(A.MaintB, 112, "bug", monthsAgo(now, 6.2)),
				UserEvent(A.MaintA, "comment", monthsAgo(now, 6.1)),
			},
			expect: map[string]bool{A.MaintA: true, A.MaintB: false},
		},
		{
			name: "project_events__pre_window_only",
			description: `
Project-level activity exists only before the window; rest is noise.

1) MR#534 created 12.0 months ago.
2) MaintB edits wiki 9.0 months ago (old).
3) MaintA labels MR#534 "needs review" 7.5 months ago (old).
4) MaintA labels Issue#113 "help wanted" 7.1 months ago (old).
5) MaintB closes Issue#113 6.2 months ago (old).

Expected: both inactive.
`,
			steps: []StoryStep{
				Seq(ProjectSetup("group/p", 53004), Maintainers(A.MaintA, A.MaintB), Contributors(A.ContribA)),
				MRCreated(A.ContribA, 534, monthsAgo(now, 12.0)),
				ProjectEvent(A.MaintB, "wiki_edit", monthsAgo(now, 9.0)),
				MRLabel(A.MaintA, 534, "needs review", monthsAgo(now, 7.5)),
				IssueLabeled(A.MaintA, 113, "help wanted", monthsAgo(now, 7.1)),
				IssueState(A.MaintB, 113, "closed", monthsAgo(now, 6.2)),
			},
			expect: map[string]bool{A.MaintA: false, A.MaintB: false},
		},

		// 4) PIPELINE SCHEDULE UPDATES IN-WINDOW
		{
			name: "pipeline_schedule__inactive",
			description: `
There are schedule updates, but only before the window.

1) MaintA updates a pipeline schedule 7.0 months ago (old).
2) MR#541 is created 8.1 months ago (noise).
3) MaintB labels Issue#120 "bug" 6.4 months ago (old).
4) MaintB edits wiki 6.3 months ago (old).
5) MaintA sets milestone M-120 on Issue#120 6.2 months ago (old).

Expected: both inactive.
`,
			steps: []StoryStep{
				Seq(ProjectSetup("group/p", 54001), Maintainers(A.MaintA, A.MaintB), Contributors(A.ContribA)),
				PipelineScheduleSet(A.MaintA, monthsAgo(now, 7.0)),
				MRCreated(A.ContribA, 541, monthsAgo(now, 8.1)),
				IssueLabeled(A.MaintB, 120, "bug", monthsAgo(now, 6.4)),
				ProjectEvent(A.MaintB, "wiki_edit", monthsAgo(now, 6.3)),
				IssueMilestoned(A.MaintA, 120, "M-120", monthsAgo(now, 6.2)),
			},
			expect: map[string]bool{A.MaintA: false, A.MaintB: false},
		},
		{
			name: "pipeline_schedule__active",
			description: `
MaintA updates a pipeline schedule inside the window; MaintB only has older activity.

1) MR#542 created 9.0 months ago (noise).
2) MaintB labels Issue#121 "triage" 7.1 months ago (old).
3) MaintA updates a pipeline schedule 5.8 months ago (in-window).
4) MaintB labels MR#542 "ready" 7.0 months ago (old).
5) MaintB closes Issue#121 6.2 months ago (old).

Expected: MaintA active, MaintB inactive.
`,
			steps: []StoryStep{
				Seq(ProjectSetup("group/p", 54002), Maintainers(A.MaintA, A.MaintB), Contributors(A.ContribA)),
				MRCreated(A.ContribA, 542, monthsAgo(now, 9.0)),
				IssueLabeled(A.MaintB, 121, "triage", monthsAgo(now, 7.1)),
				PipelineScheduleSet(A.MaintA, monthsAgo(now, 5.8)),
				MRLabel(A.MaintB, 542, "ready", monthsAgo(now, 7.0)),
				IssueState(A.MaintB, 121, "closed", monthsAgo(now, 6.2)),
			},
			expect: map[string]bool{A.MaintA: true, A.MaintB: false},
		},
		{
			name: "pipeline_schedule__mixed",
			description: `
Only MaintA updates a schedule inside the window.

1) MaintA updates a pipeline schedule 5.2 months ago (in-window).
2) MR#543 created 10.0 months ago.
3) MaintB labels MR#543 "triage" 7.3 months ago (old).
4) MaintB labels Issue#122 "help wanted" 6.2 months ago (old).
5) MaintB edits wiki 6.1 months ago (old).

Expected: MaintA active, MaintB inactive.
`,
			steps: []StoryStep{
				Seq(ProjectSetup("group/p", 54003), Maintainers(A.MaintA, A.MaintB), Contributors(A.ContribA)),
				PipelineScheduleSet(A.MaintA, monthsAgo(now, 5.2)),
				MRCreated(A.ContribA, 543, monthsAgo(now, 10.0)),
				MRLabel(A.MaintB, 543, "triage", monthsAgo(now, 7.3)),
				IssueLabeled(A.MaintB, 122, "help wanted", monthsAgo(now, 6.2)),
				ProjectEvent(A.MaintB, "wiki_edit", monthsAgo(now, 6.1)),
			},
			expect: map[string]bool{A.MaintA: true, A.MaintB: false},
		},
		{
			name: "pipeline_schedule__pre_window_only",
			description: `
Schedule updates happen, but only outside the window (plus some old noise).

1) MaintB updates a pipeline schedule 7.1 months ago (old).
2) MR#544 created 11.0 months ago.
3) MaintA labels Issue#123 "bug" 7.2 months ago (old).
4) MaintA labels MR#544 "needs work" 7.0 months ago (old).
5) MaintB closes Issue#123 6.2 months ago (old).

Expected: both inactive.
`,
			steps: []StoryStep{
				Seq(ProjectSetup("group/p", 54004), Maintainers(A.MaintA, A.MaintB), Contributors(A.ContribA)),
				PipelineScheduleSet(A.MaintB, monthsAgo(now, 7.1)),
				MRCreated(A.ContribA, 544, monthsAgo(now, 11.0)),
				IssueLabeled(A.MaintA, 123, "bug", monthsAgo(now, 7.2)),
				MRLabel(A.MaintA, 544, "needs work", monthsAgo(now, 7.0)),
				IssueState(A.MaintB, 123, "closed", monthsAgo(now, 6.2)),
			},
			expect: map[string]bool{A.MaintA: false, A.MaintB: false},
		},

		// 5) COMMIT AWARD EMOJI IN-WINDOW (via MR sha)
		{
			name: "commit_award__inactive",
			description: `
Commit award emojis happen, but only before the window.

1) MR#551 is created 8.0 months ago (provides commit SHA deadbeef551).
2) MaintA reacts on commit deadbeef551 6.5 months ago (old).
3) MaintB labels MR#551 "triage" 7.1 months ago (old).
4) MaintB labels Issue#130 "bug" 6.3 months ago (old).
5) MaintA sets milestone M-130 6.2 months ago (old).

Expected: both inactive.
`,
			steps: []StoryStep{
				Seq(ProjectSetup("group/p", 55001), Maintainers(A.MaintA, A.MaintB), Contributors(A.ContribA)),
				MRCreated(A.ContribA, 551, monthsAgo(now, 8.0)),           // MR for SHA deadbeef551
				CommitAward(A.MaintA, "deadbeef551", monthsAgo(now, 6.5)), // old
				MRLabel(A.MaintB, 551, "triage", monthsAgo(now, 7.1)),
				IssueLabeled(A.MaintB, 130, "bug", monthsAgo(now, 6.3)),
				IssueMilestoned(A.MaintA, 130, "M-130", monthsAgo(now, 6.2)),
			},
			expect: map[string]bool{A.MaintA: false, A.MaintB: false},
		},
		{
			name: "commit_award__active",
			description: `
MaintA reacts on a commit inside the window.

1) MR#552 is created 8.0 months ago (commit SHA deadbeef552).
2) MaintB labels Issue#131 "help wanted" 7.2 months ago (old).
3) MaintA reacts on commit deadbeef552 5.8 months ago (in-window).
4) MaintB labels MR#552 "ready" 7.0 months ago (old).
5) MaintB closes Issue#131 6.2 months ago (old).

Expected: MaintA active, MaintB inactive.
`,
			steps: []StoryStep{
				Seq(ProjectSetup("group/p", 55002), Maintainers(A.MaintA, A.MaintB), Contributors(A.ContribA)),
				MRCreated(A.ContribA, 552, monthsAgo(now, 8.0)), // MR for SHA deadbeef552
				IssueLabeled(A.MaintB, 131, "help wanted", monthsAgo(now, 7.2)),
				CommitAward(A.MaintA, "deadbeef552", monthsAgo(now, 5.8)),
				MRLabel(A.MaintB, 552, "ready", monthsAgo(now, 7.0)),
				IssueState(A.MaintB, 131, "closed", monthsAgo(now, 6.2)),
			},
			expect: map[string]bool{A.MaintA: true, A.MaintB: false},
		},
		{
			name: "commit_award__mixed",
			description: `
Only MaintA reacts on a commit within the window; MaintB has older actions.

1) MR#553 is created 8.0 months ago (commit SHA deadbeef553).
2) MaintB labels MR#553 "triage" 7.5 months ago (old).
3) MaintA reacts on commit deadbeef553 5.4 months ago (in-window).
4) MaintB labels Issue#132 "bug" 6.3 months ago (old).
5) MaintB edits wiki 6.1 months ago (old).

Expected: MaintA active, MaintB inactive.
`,
			steps: []StoryStep{
				Seq(ProjectSetup("group/p", 55003), Maintainers(A.MaintA, A.MaintB), Contributors(A.ContribA)),
				MRCreated(A.ContribA, 553, monthsAgo(now, 8.0)), // MR for SHA deadbeef553
				MRLabel(A.MaintB, 553, "triage", monthsAgo(now, 7.5)),
				CommitAward(A.MaintA, "deadbeef553", monthsAgo(now, 5.4)),
				IssueLabeled(A.MaintB, 132, "bug", monthsAgo(now, 6.3)),
				ProjectEvent(A.MaintB, "wiki_edit", monthsAgo(now, 6.1)),
			},
			expect: map[string]bool{A.MaintA: true, A.MaintB: false},
		},
		{
			name: "commit_award__pre_window_only",
			description: `
Commit reactions exist, but occur only before the window.

1) MR#554 is created 10.0 months ago (commit SHA deadbeef554).
2) MaintB reacts on commit deadbeef554 6.1 months ago (old).
3) MaintA labels MR#554 "needs work" 7.8 months ago (old).
4) MaintA labels Issue#133 "bug" 7.2 months ago (old).
5) MaintB closes Issue#133 6.2 months ago (old).

Expected: both inactive.
`,
			steps: []StoryStep{
				Seq(ProjectSetup("group/p", 55004), Maintainers(A.MaintA, A.MaintB), Contributors(A.ContribA)),
				MRCreated(A.ContribA, 554, monthsAgo(now, 10.0)), // MR for SHA deadbeef554
				CommitAward(A.MaintB, "deadbeef554", monthsAgo(now, 6.1)),
				MRLabel(A.MaintA, 554, "needs work", monthsAgo(now, 7.8)),
				IssueLabeled(A.MaintA, 133, "bug", monthsAgo(now, 7.2)),
				IssueState(A.MaintB, 133, "closed", monthsAgo(now, 6.2)),
			},
			expect: map[string]bool{A.MaintA: false, A.MaintB: false},
		},

		// 6) SNIPPET AWARD EMOJI IN-WINDOW
		{
			name: "snippet_award__inactive",
			description: `
Snippet reactions appear, but only before the window.

1) Snippet#7601 updated 7.0 months ago.
2) MaintA reacts on snippet#7601 6.2 months ago (old).
3) MaintB labels Issue#140 "bug" 7.4 months ago (old).
4) MR#561 created 8.3 months ago (noise).
5) MaintB labels MR#561 "triage" 6.3 months ago (old).

Expected: both inactive.
`,
			steps: []StoryStep{
				Seq(ProjectSetup("group/p", 56001), Maintainers(A.MaintA, A.MaintB), Contributors(A.ContribA)),
				Snippet(A.ContribA, 7601, monthsAgo(now, 7.0)),
				SnippetAward(A.MaintA, 7601, monthsAgo(now, 6.2)), // old
				IssueLabeled(A.MaintB, 140, "bug", monthsAgo(now, 7.4)),
				MRCreated(A.ContribA, 561, monthsAgo(now, 8.3)),
				MRLabel(A.MaintB, 561, "triage", monthsAgo(now, 6.3)),
			},
			expect: map[string]bool{A.MaintA: false, A.MaintB: false},
		},
		{
			name: "snippet_award__active",
			description: `
MaintA reacts on a snippet within the window; MaintB only has older actions.

1) Snippet#7602 updated 7.0 months ago.
2) MaintB labels Issue#141 "help wanted" 7.2 months ago (old).
3) MaintA reacts on snippet#7602 5.9 months ago (in-window).
4) MR#562 created 8.1 months ago (noise).
5) MaintB closes Issue#141 6.2 months ago (old).

Expected: MaintA active, MaintB inactive.
`,
			steps: []StoryStep{
				Seq(ProjectSetup("group/p", 56002), Maintainers(A.MaintA, A.MaintB), Contributors(A.ContribA)),
				Snippet(A.ContribA, 7602, monthsAgo(now, 7.0)),
				IssueLabeled(A.MaintB, 141, "help wanted", monthsAgo(now, 7.2)),
				SnippetAward(A.MaintA, 7602, monthsAgo(now, 5.9)),
				MRCreated(A.ContribA, 562, monthsAgo(now, 8.1)),
				IssueState(A.MaintB, 141, "closed", monthsAgo(now, 6.2)),
			},
			expect: map[string]bool{A.MaintA: true, A.MaintB: false},
		},
		{
			name: "snippet_award__mixed",
			description: `
Only MaintA reacts on a snippet inside the window.

1) Snippet#7603 updated 7.0 months ago.
2) MR#563 created 9.0 months ago.
3) MaintB labels MR#563 "triage" 7.5 months ago (old).
4) MaintA reacts on snippet#7603 5.0 months ago (in-window).
5) MaintB edits wiki 6.2 months ago (old).

Expected: MaintA active, MaintB inactive.
`,
			steps: []StoryStep{
				Seq(ProjectSetup("group/p", 56003), Maintainers(A.MaintA, A.MaintB), Contributors(A.ContribA)),
				Snippet(A.ContribA, 7603, monthsAgo(now, 7.0)),
				MRCreated(A.ContribA, 563, monthsAgo(now, 9.0)),
				MRLabel(A.MaintB, 563, "triage", monthsAgo(now, 7.5)),
				SnippetAward(A.MaintA, 7603, monthsAgo(now, 5.0)),
				ProjectEvent(A.MaintB, "wiki_edit", monthsAgo(now, 6.2)),
			},
			expect: map[string]bool{A.MaintA: true, A.MaintB: false},
		},
		{
			name: "snippet_award__pre_window_only",
			description: `
Snippet reactions exist but only before the window.

1) Snippet#7604 updated 10.0 months ago.
2) MaintB reacts on snippet#7604 6.2 months ago (old).
3) MR#564 created 8.8 months ago (noise).
4) MaintA labels MR#564 "needs work" 7.6 months ago (old).
5) MaintA labels Issue#142 "bug" 7.3 months ago (old).

Expected: both inactive.
`,
			steps: []StoryStep{
				Seq(ProjectSetup("group/p", 56004), Maintainers(A.MaintA, A.MaintB), Contributors(A.ContribA)),
				Snippet(A.ContribA, 7604, monthsAgo(now, 10.0)),
				SnippetAward(A.MaintB, 7604, monthsAgo(now, 6.2)),
				MRCreated(A.ContribA, 564, monthsAgo(now, 8.8)),
				MRLabel(A.MaintA, 564, "needs work", monthsAgo(now, 7.6)),
				IssueLabeled(A.MaintA, 142, "bug", monthsAgo(now, 7.3)),
			},
			expect: map[string]bool{A.MaintA: false, A.MaintB: false},
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			// Build story by applying all DSL steps
			builder := NewStoryBuilder(now)
			for _, step := range sc.steps {
				builder = step.apply(builder)
			}

			// Log human-readable story description
			t.Log("\nSTORY:", sc.description)
			t.Log(builder.Pretty())

			// Launch mock server and execute maintainer activity check
			h := makeClient(t, builder.story)
			got, err := h.GetMaintainerActivity()
			if err != nil {
				t.Fatalf("GetMaintainerActivity: %v", err)
			}

			// Verify expected active/inactive status for each maintainer
			if len(got) != len(sc.expect) {
				t.Fatalf("unexpected user map size: got=%v want=%v", got, sc.expect)
			}
			for user, want := range sc.expect {
				if got[user] != want {
					t.Errorf("user %q active=%v, want %v; full=%v", user, got[user], want, got)
				}
			}
		})
	}
}

////////////////////////////////////////////////////////////////////////////////
// SECTION 2: STORY DSL (Domain-Specific Language)
////////////////////////////////////////////////////////////////////////////////
// A fluent API for expressing GitLab project activity in test scenarios.
// Each DSL function returns a StoryStep that can be composed with others.
//
// DESIGN PATTERNS
// ---------------
// - Fluent Interface: Methods return StoryStep for chaining
// - Builder Pattern: Steps accumulate to build complete story
// - Human-Readable: Function names match natural language descriptions
//
// DSL CATEGORIES
// --------------
// Setup:
//   - ProjectSetup(): Initialize project metadata (path, ID)
//   - Maintainers(): Define users with elevated permissions
//   - Contributors(): Define regular contributors
//   - Seq(): Compose multiple steps into one logical unit
//
// Merge Requests:
//   - MRCreated(), MRMerged(), MRApproved(): MR lifecycle
//   - MRLabel(), MRMilestoned(), MRAward(): MR management
//   - MRState(): MR state transitions
//
// Issues:
//   - IssueLabeled(), IssueState(): Issue triaging
//   - IssueMilestoned(), IssueAward(): Issue management
//
// Project Events:
//   - ReleaseAuthored(): Version releases
//   - ManualJob(), PipelineScheduleSet(): CI/CD activity
//   - AuditEvent(), ProjectEvent(), UserEvent(): Admin/system events
//   - Snippet(), SnippetAward(), CommitAward(): Misc. activity
//
// TIME HELPERS
// ------------
// monthsAgo(): Calculate relative timestamps (e.g., "3.5 months ago")
//              Uses approximate 30.4375 days per month for stability

//
// Story DSL - Human-readable activity expressions
//

type StoryStep struct {
	apply func(*StoryBuilder) *StoryBuilder
}

// Seq composes multiple steps into a single logical unit for better readability.
// Example: Seq(ProjectSetup(...), Maintainers(...), Contributors(...)).
func Seq(steps ...StoryStep) StoryStep {
	return StoryStep{apply: func(b *StoryBuilder) *StoryBuilder {
		for _, s := range steps {
			b = s.apply(b)
		}
		return b
	}}
}

// helpers to express “N months ago” with fractional months (approximate).
func monthsAgo(base time.Time, m float64) time.Time {
	// Approximate 30.4375 days per month
	d := time.Duration(m * 30.4375 * 24 * float64(time.Hour))
	return base.Add(-d)
}

// ProjectSetup initializes a new GitLab project story.
// Must be called before other DSL functions that reference the project.
//
// Parameters:
//
//	path: Project path in "group/project" format
//	pid: Numeric project ID (used in API responses)
func ProjectSetup(path string, pid int) StoryStep {
	return StoryStep{apply: func(b *StoryBuilder) *StoryBuilder {
		group, proj := splitPath(path)
		b.story = NewGLStory(group, proj, pid, b.now)
		return b
	}}
}

// Maintainers defines users with elevated permissions (Developer+).
// These users are the subjects of activity tracking.
func Maintainers(users ...string) StoryStep {
	return StoryStep{apply: func(b *StoryBuilder) *StoryBuilder {
		b.story.WithMaintainers(users...)
		b.maintainers = append(b.maintainers, users...)
		for _, u := range users {
			b.logf("Maintainer: %s", u)
		}
		return b
	}}
}

// Contributors defines regular contributors (non-maintainers).
// Used to author MRs/issues that maintainers then interact with.
func Contributors(users ...string) StoryStep {
	return StoryStep{apply: func(b *StoryBuilder) *StoryBuilder {
		b.story.WithContributors(users...)
		for _, u := range users {
			b.logf("Contributor: %s", u)
		}
		return b
	}}
}

// (We keep the method forms too, in case other tests import them.)

func (s StoryStep) Maintainers(users ...string) StoryStep  { return Maintainers(users...) }
func (s StoryStep) Contributors(users ...string) StoryStep { return Contributors(users...) }

func MRCreated(by string, iid int, at time.Time) StoryStep {
	return StoryStep{apply: func(b *StoryBuilder) *StoryBuilder {
		b.story.MRCreated(iid, by, at)
		b.logf("%s created MR#%d at %s", by, iid, at.Format(time.RFC3339))
		return b
	}}
}

// MRApproved records a merge request approval.
// Note: Approvals alone don't count as activity in production code due to
// GitLab API limitations (approval timestamps not reliably available).
func MRApproved(by string, iid int, at time.Time) StoryStep {
	return StoryStep{apply: func(b *StoryBuilder) *StoryBuilder {
		b.story.MRApproved(iid, by, at)
		b.logf("%s approved MR#%d at %s", by, iid, at.Format(time.RFC3339))
		return b
	}}
}

// MRMerged records a merge request merge event (Primary Signal).
// This is one of the strongest indicators of maintainer activity.
func MRMerged(by string, iid int, at time.Time) StoryStep {
	return StoryStep{apply: func(b *StoryBuilder) *StoryBuilder {
		b.story.MRMerged(iid, by, at)
		b.logf("%s merged MR#%d at %s", by, iid, at.Format(time.RFC3339))
		return b
	}}
}

// MRLabel records adding a label to a merge request (Extended Signal).
// Demonstrates triage activity by maintainers.
func MRLabel(by string, iid int, label string, at time.Time) StoryStep {
	return StoryStep{apply: func(b *StoryBuilder) *StoryBuilder {
		b.story.mrLabelEvents[iid] = append(
			b.story.mrLabelEvents[iid],
			glLabelEvent{User: by, Label: label, At: at},
		)
		b.logf(
			"%s added label %q on MR#%d at %s",
			by, label, iid, at.Format(time.RFC3339),
		)
		return b
	}}
}

// MRMilestoned records setting a milestone on an MR (Extended Signal).
func MRMilestoned(by string, iid int, ms string, at time.Time) StoryStep {
	return StoryStep{apply: func(b *StoryBuilder) *StoryBuilder {
		b.story.mrMilEvents[iid] = append(b.story.mrMilEvents[iid],
			glMilEvent{User: by, Milestone: ms, At: at},
		)
		b.logf(
			"%s set milestone %q on MR#%d at %s",
			by, ms, iid, at.Format(time.RFC3339),
		)
		return b
	}}
}

func MRAward(by string, iid int, at time.Time) StoryStep {
	return StoryStep{apply: func(b *StoryBuilder) *StoryBuilder {
		b.story.MRAwarded(iid, by, at)
		b.logf("%s reacted on MR#%d at %s", by, iid, at.Format(time.RFC3339))
		return b
	}}
}

// NEW DSL helpers.
func MRState(by string, iid int, state string, at time.Time) StoryStep {
	return StoryStep{apply: func(b *StoryBuilder) *StoryBuilder {
		b.story.mrStateEvents[iid] = append(
			b.story.mrStateEvents[iid],
			glStateEvent{User: by, State: state, At: at},
		)
		b.logf(
			"%s changed MR#%d state to %q at %s",
			by, iid, state, at.Format(time.RFC3339),
		)
		return b
	}}
}

func AuditEvent(by, kind string, at time.Time) StoryStep {
	return StoryStep{apply: func(b *StoryBuilder) *StoryBuilder {
		b.story.Audit(by, kind, at)
		b.logf("%s audit event %q at %s", by, kind, at.Format(time.RFC3339))
		return b
	}}
}

func IssueLabeled(by string, iid int, label string, at time.Time) StoryStep {
	return StoryStep{apply: func(b *StoryBuilder) *StoryBuilder {
		b.story.IssueLabeled(iid, by, label, at)
		b.logf(
			"%s added label %q on Issue#%d at %s",
			by, label, iid, at.Format(time.RFC3339),
		)
		return b
	}}
}

func IssueState(by string, iid int, state string, at time.Time) StoryStep {
	return StoryStep{apply: func(b *StoryBuilder) *StoryBuilder {
		b.story.IssueStateChanged(iid, by, state, at)
		b.logf(
			"%s changed Issue#%d state to %q at %s",
			by, iid, state, at.Format(time.RFC3339),
		)
		return b
	}}
}

func IssueMilestoned(by string, iid int, ms string, at time.Time) StoryStep {
	return StoryStep{apply: func(b *StoryBuilder) *StoryBuilder {
		b.story.IssueMilestoneChanged(iid, by, ms, at)
		b.logf(
			"%s set milestone %q on Issue#%d at %s",
			by, ms, iid, at.Format(time.RFC3339),
		)
		return b
	}}
}

func IssueAward(by string, iid int, at time.Time) StoryStep {
	return StoryStep{apply: func(b *StoryBuilder) *StoryBuilder {
		b.story.IssueAwarded(iid, by, at)
		b.logf("%s reacted on Issue#%d at %s", by, iid, at.Format(time.RFC3339))
		return b
	}}
}

func ManualJob(by string, at time.Time) StoryStep {
	return StoryStep{apply: func(b *StoryBuilder) *StoryBuilder {
		b.story.ManualJobPlayed(by, at)
		b.logf("%s played a manual CI job at %s", by, at.Format(time.RFC3339))
		return b
	}}
}

func PipelineScheduleSet(by string, at time.Time) StoryStep {
	return StoryStep{apply: func(b *StoryBuilder) *StoryBuilder {
		b.story.PipelineScheduleUpdated(by, at)
		b.logf("%s updated a pipeline schedule at %s", by, at.Format(time.RFC3339))
		return b
	}}
}

func ReleaseAuthored(by, tag string, at time.Time) StoryStep {
	return StoryStep{apply: func(b *StoryBuilder) *StoryBuilder {
		b.story.ReleaseAuthored(tag, by, at)
		b.logf("%s authored release %s at %s", by, tag, at.Format(time.RFC3339))
		return b
	}}
}

func ProjectEvent(by, kind string, at time.Time) StoryStep {
	return StoryStep{apply: func(b *StoryBuilder) *StoryBuilder {
		b.story.ProjectEvent(by, kind, at)
		b.logf("%s performed project event %q at %s", by, kind, at.Format(time.RFC3339))
		return b
	}}
}

func UserEvent(by, kind string, at time.Time) StoryStep {
	return StoryStep{apply: func(b *StoryBuilder) *StoryBuilder {
		// Target project event only
		b.story.UserEvent(by, b.story.ProjectID, at, kind)
		b.logf("%s user event %q at %s", by, kind, at.Format(time.RFC3339))
		return b
	}}
}

func Snippet(by string, id int, updatedAt time.Time) StoryStep {
	return StoryStep{apply: func(b *StoryBuilder) *StoryBuilder {
		b.story.Snippet(id, updatedAt)
		b.logf("%s created/updated snippet %d at %s", by, id, updatedAt.Format(time.RFC3339))
		return b
	}}
}

func SnippetAward(by string, id int, at time.Time) StoryStep {
	return StoryStep{apply: func(b *StoryBuilder) *StoryBuilder {
		b.story.SnippetAwarded(id, by, at)
		b.logf("%s reacted on snippet %d at %s", by, id, at.Format(time.RFC3339))
		return b
	}}
}

func CommitAward(by, sha string, at time.Time) StoryStep {
	return StoryStep{apply: func(b *StoryBuilder) *StoryBuilder {
		b.story.CommitAwarded(sha, by, at)
		b.logf("%s reacted on commit %s at %s", by, sha, at.Format(time.RFC3339))
		return b
	}}
}

////////////////////////////////////////////////////////////////////////////////
// SECTION 3: STORY BUILDER
////////////////////////////////////////////////////////////////////////////////
// StoryBuilder translates DSL steps into GLStory data and provides
// human-readable logging of each step for test output.
//
// RESPONSIBILITIES
// ----------------
// 1. Execute DSL steps in sequence
// 2. Build GLStory data structure for mock server
// 3. Generate human-readable descriptions for debugging
// 4. Maintain test actors (maintainers, contributors)

// StoryBuilder accumulates DSL steps and builds a GLStory.
type StoryBuilder struct {
	now         time.Time // Reference time for relative calculations
	story       *GLStory  // Accumulates activity data for mock server
	maintainers []string  // List of maintainer usernames
	lines       []string  // Human-readable log of steps (for debugging)
}

func NewStoryBuilder(now time.Time) *StoryBuilder { return &StoryBuilder{now: now} }

// Pretty returns a human-readable summary of all story steps.
func (b *StoryBuilder) Pretty() string { return strings.Join(b.lines, "\n") }

// logf adds a formatted log line describing a story step.
func (b *StoryBuilder) logf(f string, a ...any) {
	b.lines = append(b.lines, " - "+fmt.Sprintf(f, a...))
}

// splitPath parses "group/project" path into components.
// Defaults to ("group", path) if no separator found.
func splitPath(p string) (string, string) {
	parts := strings.SplitN(p, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "group", p
}

////////////////////////////////////////////////////////////////////////////////
// SECTION 4: MOCK GITLAB SERVER
////////////////////////////////////////////////////////////////////////////////
// GLStory stores project activity data and provides HTTP handlers that
// respond to GitLab API calls with the stored data.
//
// ARCHITECTURE
// ------------
// 1. GLStory: Data model storing all project activity (MRs, issues, etc.)
// 2. Handler(): Returns http.Handler that serves GitLab API endpoints
// 3. Helper methods: Build API responses from stored data
//
// SUPPORTED ENDPOINTS
// -------------------
// Projects:     /api/v4/projects/:id
// Members:      /api/v4/projects/:id/members/all
// Users:        /api/v4/users?username=...
// User Events:  /api/v4/users/:id/events
// MRs:          /api/v4/projects/:id/merge_requests
// MR Details:   /api/v4/projects/:id/merge_requests/:iid/(approvals|resource_label_events|...)
// Issues:       /api/v4/projects/:id/issues
// Issue Events: /api/v4/projects/:id/issues/:iid/(resource_label_events|...)
// Commits:      /api/v4/projects/:id/repository/commits/:sha/award_emoji
// Snippets:     /api/v4/projects/:id/snippets
// Releases:     /api/v4/projects/:id/releases
// Pipelines:    /api/v4/projects/:id/pipelines
// Jobs:         /api/v4/projects/:id/pipelines/:id/jobs
// Schedules:    /api/v4/projects/:id/pipeline_schedules
// Events:       /api/v4/projects/:id/events
// Audit:        /api/v4/projects/:id/audit_events
//
// DATA MODEL
// ----------
// GLStory stores activity in categorized collections:
//   - User management: elevated (maintainers), contributors, userIDs
//   - MR activity: mrs, mrApprovals, mrLabelEvents, mrStateEvents, etc.
//   - Issue activity: issues, issueLabelEvents, issueStateEvents, etc.
//   - Project events: releases, pipelines, schedules, auditEvents, etc.
//   - Awards: mrAward, issueAward, commitAward, snippetAward

// GLStory represents a GitLab project's complete activity history for testing.
type GLStory struct {
	Now              time.Time
	mrLabelEvents    map[int][]glLabelEvent
	issueLabelEvents map[int][]glLabelEvent
	issueMilEvents   map[int][]glMilEvent
	snippetAward     map[int][]glAward
	issueStateEvents map[int][]glStateEvent
	mrAward          map[int][]glAward
	userIDs          map[string]int
	mrStateEvents    map[int][]glStateEvent
	issueAward       map[int][]glAward
	mrApprovals      map[int][]glApproval
	commitAward      map[string][]glAward
	jobs             map[int][]glJob
	mrMilEvents      map[int][]glMilEvent
	Group            string
	Project          string
	events           []glUserEvent
	contributors     []string
	auditEvents      []glAuditEvent
	pipelines        []glPipeline
	releases         []glRelease
	schedules        []glSchedule
	issues           []glIssue
	mrs              []glMR
	snippets         []glSnippet
	elevated         []string
	projectEvents    []glProjectEvent
	ProjectID        int
	nextUserID       int
}

func NewGLStory(group, project string, projectID int, now time.Time) *GLStory {
	return &GLStory{
		Group:            group,
		Project:          project,
		ProjectID:        projectID,
		Now:              now,
		nextUserID:       1000,
		userIDs:          map[string]int{},
		mrApprovals:      map[int][]glApproval{},
		issueLabelEvents: map[int][]glLabelEvent{},
		mrLabelEvents:    map[int][]glLabelEvent{},
		issueStateEvents: map[int][]glStateEvent{},
		mrStateEvents:    map[int][]glStateEvent{},
		issueMilEvents:   map[int][]glMilEvent{},
		mrMilEvents:      map[int][]glMilEvent{},
		jobs:             map[int][]glJob{},
		commitAward:      map[string][]glAward{},
		issueAward:       map[int][]glAward{},
		mrAward:          map[int][]glAward{},
		snippetAward:     map[int][]glAward{},
	}
}

// WithMaintainers adds elevated users to the project.
// These users will be returned in the /members/all API endpoint.
func (s *GLStory) WithMaintainers(users ...string) *GLStory {
	s.elevated = append(s.elevated, users...)
	return s
}

// WithContributors adds regular contributors to the project.
func (s *GLStory) WithContributors(users ...string) *GLStory {
	s.contributors = append(s.contributors, users...)
	return s
}

////////////////////////////////////////////////////////////////////////////////
// GLStory Builder Methods
////////////////////////////////////////////////////////////////////////////////
// These methods populate the mock GitLab server's data store.
// Each method adds activity data that will be returned by API endpoints.

// MRCreated adds a merge request creation event.
func (s *GLStory) MRCreated(iid int, author string, at time.Time) *GLStory {
	mr := glMR{IID: iid, Author: author, CreatedAt: at}
	s.mrs = append(s.mrs, mr)
	return s
}

// MRMerged adds a merge request merge event.
func (s *GLStory) MRMerged(iid int, merger string, mergedAt time.Time) *GLStory {
	mr := glMR{IID: iid, MergedBy: merger, MergedAt: &mergedAt}
	s.mrs = append(s.mrs, mr)
	return s
}

func (s *GLStory) MRApproved(iid int, approver string, approvedAt time.Time) *GLStory {
	s.mrApprovals[iid] = append(s.mrApprovals[iid], glApproval{User: approver, At: approvedAt})
	return s
}

func (s *GLStory) MRAwarded(iid int, by string, at time.Time) *GLStory {
	s.mrAward[iid] = append(s.mrAward[iid], glAward{User: by, At: at})
	return s
}

// IssueLabeled adds a label event to an issue.
// Automatically ensures the issue exists in the story.
func (s *GLStory) IssueLabeled(iid int, user, label string, at time.Time) *GLStory {
	s.ensureIssue(iid)
	evt := glLabelEvent{User: user, Label: label, At: at}
	s.issueLabelEvents[iid] = append(s.issueLabelEvents[iid], evt)
	return s
}

// IssueStateChanged adds a state transition event (opened, closed, reopened).
func (s *GLStory) IssueStateChanged(iid int, user, state string, at time.Time) *GLStory {
	s.ensureIssue(iid)
	evt := glStateEvent{User: user, State: state, At: at}
	s.issueStateEvents[iid] = append(s.issueStateEvents[iid], evt)
	return s
}

// IssueMilestoneChanged adds a milestone change event.
func (s *GLStory) IssueMilestoneChanged(iid int, user, ms string, at time.Time) *GLStory {
	s.ensureIssue(iid)
	evt := glMilEvent{User: user, Milestone: ms, At: at}
	s.issueMilEvents[iid] = append(s.issueMilEvents[iid], evt)
	return s
}

// IssueAwarded adds an award emoji reaction to an issue.
func (s *GLStory) IssueAwarded(iid int, by string, at time.Time) *GLStory {
	s.ensureIssue(iid)
	s.issueAward[iid] = append(s.issueAward[iid], glAward{User: by, At: at})
	return s
}

// ReleaseAuthored adds a release creation event (Primary Signal).
func (s *GLStory) ReleaseAuthored(tag, author string, at time.Time) *GLStory {
	rel := glRelease{Tag: tag, Author: author, At: at}
	s.releases = append(s.releases, rel)
	return s
}

func (s *GLStory) ManualJobPlayed(user string, at time.Time) *GLStory {
	pipeID := len(s.pipelines) + 1
	s.pipelines = append(s.pipelines, glPipeline{ID: pipeID, UpdatedAt: at})
	job := glJob{User: user, When: "manual", StartedAt: &at}
	s.jobs[pipeID] = append(s.jobs[pipeID], job)
	return s
}

func (s *GLStory) PipelineScheduleUpdated(owner string, at time.Time) *GLStory {
	s.schedules = append(s.schedules, glSchedule{Owner: owner, UpdatedAt: &at})
	return s
}

func (s *GLStory) ProjectEvent(user, kind string, at time.Time) *GLStory {
	evt := glProjectEvent{Author: user, Kind: kind, At: at}
	s.projectEvents = append(s.projectEvents, evt)
	return s
}

func (s *GLStory) UserEvent(user string, projectID int, at time.Time, kind string) *GLStory {
	evt := glUserEvent{User: user, ProjectID: projectID, At: at, Kind: kind}
	s.events = append(s.events, evt)
	return s
}

func (s *GLStory) CommitAwarded(sha, by string, at time.Time) *GLStory {
	s.commitAward[sha] = append(s.commitAward[sha], glAward{User: by, At: at})
	return s
}

func (s *GLStory) Snippet(id int, updatedAt time.Time) *GLStory {
	snip := glSnippet{ID: id, Title: "snip", UpdatedAt: updatedAt}
	s.snippets = append(s.snippets, snip)
	return s
}

func (s *GLStory) SnippetAwarded(id int, by string, at time.Time) *GLStory {
	s.snippetAward[id] = append(s.snippetAward[id], glAward{User: by, At: at})
	return s
}

func (s *GLStory) Audit(author, kind string, at time.Time) *GLStory {
	evt := glAuditEvent{Author: author, Kind: kind, At: at}
	s.auditEvents = append(s.auditEvents, evt)
	return s
}

func (s *GLStory) ensureIssue(iid int) {
	for _, is := range s.issues {
		if is.IID == iid {
			return
		}
	}
	s.issues = append(s.issues, glIssue{IID: iid, UpdatedAt: s.Now})
}

////////////////////////////////////////////////////////////////////////////////
// HTTP Handler - Mock GitLab API Server
////////////////////////////////////////////////////////////////////////////////
// Handler returns an HTTP handler that responds to GitLab API requests with
// data from this story. The handler implements 30+ API endpoints covering:
//   - Project metadata and members
//   - Merge requests and their events
//   - Issues and their events
//   - Releases, pipelines, jobs, schedules
//   - User events and audit logs
//
// Request handling:
//   - Supports pagination via page/per_page query params
//   - Returns JSON responses matching GitLab API format
//   - Filters by time ranges where applicable
//   - Maps usernames to numeric IDs consistently

//nolint:gocognit // test mock
func (s *GLStory) Handler() http.Handler {
	mux := http.NewServeMux()
	pid := strconv.Itoa(s.ProjectID)

	// Helper to build API paths
	apiPath := func(endpoint string) string {
		return "/api/v4/projects/" + pid + endpoint
	}

	mux.HandleFunc(apiPath(""), func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"id": s.ProjectID, "path_with_namespace": s.Group + "/" + s.Project}) //nolint:errcheck // test helper
	})
	mux.HandleFunc(
		apiPath("/members/all"),
		func(w http.ResponseWriter, r *http.Request) {
			var out []map[string]any
			for _, u := range s.elevated {
				out = append(out, map[string]any{
					"username":     u,
					"access_level": 40,
				})
			}
			for _, u := range s.contributors {
				out = append(out, map[string]any{
					"username":     u,
					"access_level": 20,
				})
			}
			_ = json.NewEncoder(w).Encode(out) //nolint:errcheck // test helper
		})
	mux.HandleFunc("/api/v4/users", func(w http.ResponseWriter, r *http.Request) {
		u := r.URL.Query().Get("username")
		if u == "" {
			_ = json.NewEncoder(w).Encode([]any{}) //nolint:errcheck // test helper
			return
		}
		user := map[string]any{"id": s.uid(u), "username": u}
		_ = json.NewEncoder(w).Encode([]map[string]any{user}) //nolint:errcheck // test helper
	})
	mux.HandleFunc("/api/v4/users/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v4/users/")
		parts := strings.Split(path, "/")
		if len(parts) >= 2 && parts[1] == "events" {
			uid, _ := strconv.Atoi(parts[0]) //nolint:errcheck // test helper
			user := s.uname(uid)
			var out []map[string]any
			after := r.URL.Query().Get("after")
			var afterT time.Time
			if after != "" {
				afterT, _ = time.Parse(time.RFC3339, after) //nolint:errcheck // test helper
			}
			for _, e := range s.events {
				if e.User == user && e.ProjectID == s.ProjectID && (afterT.IsZero() || e.At.After(afterT)) {
					out = append(out, map[string]any{
						"project_id":  s.ProjectID,
						"created_at":  e.At,
						"action_name": e.Kind,
					})
				}
			}
			_ = json.NewEncoder(w).Encode(out) //nolint:errcheck // test helper
			return
		}
	})

	mux.HandleFunc(
		apiPath("/merge_requests"),
		func(w http.ResponseWriter, r *http.Request) {
			ua := r.URL.Query().Get("updated_after")
			var uaT time.Time
			if ua != "" {
				uaT, _ = time.Parse(time.RFC3339, ua) //nolint:errcheck // test helper
			}
			var out []map[string]any
			seen := map[int]bool{}
			for _, mr := range s.mrs {
				if seen[mr.IID] {
					continue
				}
				seen[mr.IID] = true
				if mr.MergedAt != nil && !uaT.IsZero() && mr.MergedAt.Before(uaT) {
					continue
				}
				obj := map[string]any{
					"iid": mr.IID,
					"sha": "deadbeef" + strconv.Itoa(mr.IID),
					"author": map[string]any{
						"username": mr.Author,
					},
				}
				if mr.MergedAt != nil {
					obj["merged_at"] = mr.MergedAt
					obj["merged_by"] = map[string]any{
						"username": mr.MergedBy,
					}
				}
				out = append(out, obj)
			}
			_ = json.NewEncoder(w).Encode(out) //nolint:errcheck // test helper
		})
	mux.HandleFunc(
		apiPath("/merge_requests/"),
		func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, apiPath("/merge_requests/"))
			parts := strings.Split(path, "/")
			if len(parts) < 2 {
				http.NotFound(w, r)
				return
			}
			iid, _ := strconv.Atoi(parts[0]) //nolint:errcheck // test helper
			switch parts[1] {
			case "approvals":
				var approvedBy []map[string]any
				for _, ap := range s.mrApprovals[iid] {
					approvedBy = append(approvedBy, map[string]any{
						"user": map[string]any{
							"username": ap.User,
						},
						"approved_at": ap.At,
					})
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"approved_by": approvedBy}) //nolint:errcheck // test helper
			case "resource_label_events":
				var out []map[string]any
				for _, e := range s.mrLabelEvents[iid] {
					out = append(out, map[string]any{
						"user": map[string]any{
							"username": e.User,
						},
						"created_at": e.At,
						"label": map[string]any{
							"name": e.Label,
						},
					})
				}
				_ = json.NewEncoder(w).Encode(out) //nolint:errcheck // test helper
			case "resource_state_events":
				var out []map[string]any
				for _, e := range s.mrStateEvents[iid] {
					out = append(out, map[string]any{
						"user": map[string]any{
							"username": e.User,
						},
						"created_at": e.At,
						"state":      e.State,
					})
				}
				_ = json.NewEncoder(w).Encode(out) //nolint:errcheck // test helper
			case "resource_milestone_events":
				var out []map[string]any
				for _, e := range s.mrMilEvents[iid] {
					out = append(out, map[string]any{
						"user": map[string]any{
							"username": e.User,
						},
						"created_at": e.At,
						"milestone": map[string]any{
							"title": e.Milestone,
						},
					})
				}
				_ = json.NewEncoder(w).Encode(out) //nolint:errcheck // test helper
			case "award_emoji":
				var out []map[string]any
				for _, a := range s.mrAward[iid] {
					out = append(out, map[string]any{
						"user": map[string]any{
							"username": a.User,
						},
						"created_at": a.At,
						"name":       "thumbsup",
					})
				}
				_ = json.NewEncoder(w).Encode(out) //nolint:errcheck // test helper
			default:
				http.NotFound(w, r)
			}
		})

	mux.HandleFunc(
		apiPath("/issues"),
		func(w http.ResponseWriter, r *http.Request) {
			var out []map[string]any
			for _, is := range s.issues {
				out = append(out, map[string]any{
					"iid":        is.IID,
					"updated_at": is.UpdatedAt,
				})
			}
			_ = json.NewEncoder(w).Encode(out) //nolint:errcheck // test helper
		})
	mux.HandleFunc(
		apiPath("/issues/"),
		func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, apiPath("/issues/"))
			parts := strings.Split(path, "/")
			if len(parts) < 2 {
				http.NotFound(w, r)
				return
			}
			iid, _ := strconv.Atoi(parts[0]) //nolint:errcheck // test helper
			switch parts[1] {
			case "resource_label_events":
				var out []map[string]any
				for _, e := range s.issueLabelEvents[iid] {
					out = append(out, map[string]any{
						"user": map[string]any{
							"username": e.User,
						},
						"created_at": e.At,
						"label": map[string]any{
							"name": e.Label,
						},
					})
				}
				_ = json.NewEncoder(w).Encode(out) //nolint:errcheck // test helper
			case "resource_state_events":
				var out []map[string]any
				for _, e := range s.issueStateEvents[iid] {
					out = append(out, map[string]any{
						"user": map[string]any{
							"username": e.User,
						},
						"created_at": e.At,
						"state":      e.State,
					})
				}
				_ = json.NewEncoder(w).Encode(out) //nolint:errcheck // test helper
			case "resource_milestone_events":
				var out []map[string]any
				for _, e := range s.issueMilEvents[iid] {
					out = append(out, map[string]any{
						"user": map[string]any{
							"username": e.User,
						},
						"created_at": e.At,
						"milestone": map[string]any{
							"title": e.Milestone,
						},
					})
				}
				_ = json.NewEncoder(w).Encode(out) //nolint:errcheck // test helper
			case "award_emoji":
				var out []map[string]any
				for _, a := range s.issueAward[iid] {
					out = append(out, map[string]any{
						"user": map[string]any{
							"username": a.User,
						},
						"created_at": a.At,
						"name":       "thumbsup",
					})
				}
				_ = json.NewEncoder(w).Encode(out) //nolint:errcheck // test helper
			default:
				http.NotFound(w, r)
			}
		})

	mux.HandleFunc(
		apiPath("/repository/commits/"),
		func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, apiPath("/repository/commits/"))
			parts := strings.Split(path, "/")
			if len(parts) >= 2 && parts[1] == "award_emoji" {
				sha := parts[0]
				var out []map[string]any
				for _, a := range s.commitAward[sha] {
					out = append(out, map[string]any{
						"user": map[string]any{
							"username": a.User,
						},
						"created_at": a.At,
						"name":       "thumbsup",
					})
				}
				_ = json.NewEncoder(w).Encode(out) //nolint:errcheck // test helper
				return
			}
			http.NotFound(w, r)
		})

	mux.HandleFunc(
		apiPath("/snippets"),
		func(w http.ResponseWriter, r *http.Request) {
			var out []map[string]any
			for _, s2 := range s.snippets {
				out = append(out, map[string]any{
					"id":         s2.ID,
					"title":      s2.Title,
					"updated_at": s2.UpdatedAt,
				})
			}
			_ = json.NewEncoder(w).Encode(out) //nolint:errcheck // test helper
		})
	mux.HandleFunc(
		apiPath("/snippets/"),
		func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, apiPath("/snippets/"))
			parts := strings.Split(path, "/")
			if len(parts) >= 2 && parts[1] == "award_emoji" {
				sid, _ := strconv.Atoi(parts[0]) //nolint:errcheck // test helper
				var out []map[string]any
				for _, a := range s.snippetAward[sid] {
					out = append(out, map[string]any{
						"user": map[string]any{
							"username": a.User,
						},
						"created_at": a.At,
						"name":       "thumbsup",
					})
				}
				_ = json.NewEncoder(w).Encode(out) //nolint:errcheck // test helper
				return
			}
			http.NotFound(w, r)
		})

	mux.HandleFunc(
		apiPath("/releases"),
		func(w http.ResponseWriter, r *http.Request) {
			sort.SliceStable(s.releases, func(i, j int) bool {
				return s.releases[i].At.After(s.releases[j].At)
			})
			var out []map[string]any
			for _, rel := range s.releases {
				out = append(out, map[string]any{
					"tag_name": rel.Tag,
					"author": map[string]any{
						"username": rel.Author,
					},
					"released_at": rel.At,
				})
			}
			_ = json.NewEncoder(w).Encode(out) //nolint:errcheck // test helper
		})

	mux.HandleFunc(
		apiPath("/pipelines"),
		func(w http.ResponseWriter, r *http.Request) {
			var out []map[string]any
			for _, p := range s.pipelines {
				out = append(out, map[string]any{
					"id":         p.ID,
					"updated_at": p.UpdatedAt,
				})
			}
			_ = json.NewEncoder(w).Encode(out) //nolint:errcheck // test helper
		})
	mux.HandleFunc(
		apiPath("/pipelines/"),
		func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, apiPath("/pipelines/"))
			parts := strings.Split(path, "/")
			if len(parts) >= 2 && parts[1] == "jobs" {
				id, _ := strconv.Atoi(parts[0]) //nolint:errcheck // test helper
				var out []map[string]any
				for _, j := range s.jobs[id] {
					m := map[string]any{
						"id":     123,
						"name":   "deploy",
						"status": "success",
						"when":   j.When,
						"user": map[string]any{
							"username": j.User,
						},
					}
					if j.StartedAt != nil {
						m["started_at"] = *j.StartedAt
					}
					if j.FinishedAt != nil {
						m["finished_at"] = *j.FinishedAt
					}
					if j.CreatedAt != nil {
						m["created_at"] = *j.CreatedAt
					}
					out = append(out, m)
				}
				_ = json.NewEncoder(w).Encode(out) //nolint:errcheck // test helper
				return
			}
			http.NotFound(w, r)
		})

	mux.HandleFunc(
		apiPath("/pipeline_schedules"),
		func(w http.ResponseWriter, r *http.Request) {
			var out []map[string]any
			for _, s2 := range s.schedules {
				obj := map[string]any{
					"id": s2.ID,
					"owner": map[string]any{
						"username": s2.Owner,
					},
				}
				if s2.UpdatedAt != nil {
					obj["updated_at"] = *s2.UpdatedAt
				}
				if s2.CreatedAt != nil {
					obj["created_at"] = *s2.CreatedAt
				}
				out = append(out, obj)
			}
			_ = json.NewEncoder(w).Encode(out) //nolint:errcheck // test helper
		})

	mux.HandleFunc(
		apiPath("/events"),
		func(w http.ResponseWriter, r *http.Request) {
			var out []map[string]any
			after := r.URL.Query().Get("after")
			var afterT time.Time
			if after != "" {
				afterT, _ = time.Parse(time.RFC3339, after) //nolint:errcheck // test helper
			}
			for _, ev := range s.projectEvents {
				if afterT.IsZero() || ev.At.After(afterT) {
					out = append(out, map[string]any{
						"author_username": ev.Author,
						"created_at":      ev.At,
						"action_name":     ev.Kind,
					})
				}
			}
			_ = json.NewEncoder(w).Encode(out) //nolint:errcheck // test helper
		})

	mux.HandleFunc(
		apiPath("/audit_events"),
		func(w http.ResponseWriter, r *http.Request) {
			var out []map[string]any
			createdAfter := r.URL.Query().Get("created_after")
			var aft time.Time
			if createdAfter != "" {
				aft, _ = time.Parse(time.RFC3339, createdAfter) //nolint:errcheck // test helper
			}
			for _, a := range s.auditEvents {
				if aft.IsZero() || a.At.After(aft) {
					out = append(out, map[string]any{
						"created_at": a.At,
						"details": map[string]any{
							"author_name": a.Author,
							"target":      "project",
							"with":        a.Kind,
						},
					})
				}
			}
			_ = json.NewEncoder(w).Encode(out) //nolint:errcheck // test helper
		})

	return mux
}

func (s *GLStory) uid(username string) int {
	if id, ok := s.userIDs[username]; ok {
		return id
	}
	s.nextUserID++
	s.userIDs[username] = s.nextUserID
	return s.nextUserID
}

func (s *GLStory) uname(id int) string {
	for name, uid := range s.userIDs {
		if uid == id {
			return name
		}
	}
	return ""
}

// tiny types.
type glUserEvent struct {
	At        time.Time
	User      string
	Kind      string
	ProjectID int
}
type glMR struct {
	CreatedAt time.Time
	MergedAt  *time.Time
	Author    string
	MergedBy  string
	IID       int
}
type glApproval struct {
	At   time.Time
	User string
}
type glIssue struct {
	UpdatedAt time.Time
	IID       int
}
type glLabelEvent struct {
	At    time.Time
	User  string
	Label string
}
type glStateEvent struct {
	At    time.Time
	User  string
	State string
}
type glMilEvent struct {
	At        time.Time
	User      string
	Milestone string
}
type glRelease struct {
	At     time.Time
	Tag    string
	Author string
}
type glAuditEvent struct {
	At     time.Time
	Author string
	Kind   string
}
type glPipeline struct {
	UpdatedAt time.Time
	ID        int
}
type glJob struct {
	CreatedAt  *time.Time
	StartedAt  *time.Time
	FinishedAt *time.Time
	User       string
	When       string
}
type glSchedule struct {
	CreatedAt *time.Time
	UpdatedAt *time.Time
	Owner     string
	ID        int
}
type glSnippet struct {
	UpdatedAt time.Time
	Title     string
	ID        int
}
type glAward struct {
	At   time.Time
	User string
}
type glProjectEvent struct {
	Author string
	At     time.Time
	Kind   string
}

////////////////////////////////////////////////////////////////////////////////
// UNIT TESTS FOR EARLY TERMINATION OPTIMIZATION
////////////////////////////////////////////////////////////////////////////////

// TestMarkActive verifies markActive returns correct boolean values.
func TestMarkActive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		elevated map[string]struct{}
		user     string
		want     bool
	}{
		{
			name:     "elevated user marked active",
			elevated: map[string]struct{}{"alice": {}},
			user:     "alice",
			want:     true,
		},
		{
			name:     "non-elevated user not marked",
			elevated: map[string]struct{}{"alice": {}},
			user:     "bob",
			want:     false,
		},
		{
			name:     "empty username",
			elevated: map[string]struct{}{"alice": {}},
			user:     "",
			want:     false,
		},
		{
			name:     "case insensitive match",
			elevated: map[string]struct{}{"alice": {}},
			user:     "ALICE",
			want:     true,
		},
		{
			name:     "whitespace trimmed",
			elevated: map[string]struct{}{"alice": {}},
			user:     "  alice  ",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := &MaintainerActivityHandler{
				elevated: tt.elevated,
				active:   make(map[string]bool),
			}
			got := h.markActive(tt.user)
			if got != tt.want {
				t.Errorf("markActive(%q) = %v, want %v", tt.user, got, tt.want)
			}
		})
	}
}

// TestAllActive verifies allActive
// correctly checks if all elevated
// users are active.
func TestAllActive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		elevated map[string]struct{}
		active   map[string]bool
		name     string
		want     bool
	}{
		{
			name:     "no elevated users",
			elevated: map[string]struct{}{},
			active:   map[string]bool{},
			want:     true,
		},
		{
			name:     "all elevated users active",
			elevated: map[string]struct{}{"alice": {}, "bob": {}},
			active:   map[string]bool{"alice": true, "bob": true},
			want:     true,
		},
		{
			name:     "some elevated users active",
			elevated: map[string]struct{}{"alice": {}, "bob": {}},
			active:   map[string]bool{"alice": true},
			want:     false,
		},
		{
			name:     "no elevated users active",
			elevated: map[string]struct{}{"alice": {}, "bob": {}},
			active:   map[string]bool{},
			want:     false,
		},
		{
			name:     "active contains non-elevated users",
			elevated: map[string]struct{}{"alice": {}},
			active:   map[string]bool{"alice": true, "charlie": true},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := &MaintainerActivityHandler{
				elevated: tt.elevated,
				active:   tt.active,
			}
			got := h.allActive()
			if got != tt.want {
				t.Errorf("allActive() = %v, want %v (elevated=%v, active=%v)",
					got, tt.want, tt.elevated, tt.active)
			}
		})
	}
}

// TestMarkActiveEarlyTermination verifies sequential
// marking with allActive checks.
func TestMarkActiveEarlyTermination(t *testing.T) {
	t.Parallel()

	h := &MaintainerActivityHandler{
		elevated: map[string]struct{}{"alice": {}, "bob": {}, "charlie": {}},
		active:   make(map[string]bool),
	}

	// Initially no one is active
	if h.allActive() {
		t.Error("allActive() should be false initially")
	}

	// Mark alice active
	if !h.markActive("alice") {
		t.Error("markActive(alice) should return true")
	}
	if h.allActive() {
		t.Error("allActive() should be false with only alice active")
	}

	// Mark bob active
	if !h.markActive("bob") {
		t.Error("markActive(bob) should return true")
	}
	if h.allActive() {
		t.Error("allActive() should be false with alice and bob active")
	}

	// Mark charlie active - now all active
	if !h.markActive("charlie") {
		t.Error("markActive(charlie) should return true")
	}
	if !h.allActive() {
		t.Error("allActive() should be true when all users are active")
	}
}

// TestMarkActiveReturnValueOptimization tests duplicatedetection.
func TestMarkActiveReturnValueOptimization(t *testing.T) {
	t.Parallel()

	h := &MaintainerActivityHandler{
		elevated: map[string]struct{}{"alice": {}, "bob": {}},
		active:   make(map[string]bool),
	}

	// Simulate processing 5 events but only 2 unique users
	events := []string{"alice", "bob", "alice", "alice", "bob"}
	newActivations := 0

	for _, user := range events {
		if h.markActive(user) {
			newActivations++
		}
	}

	if newActivations != 2 {
		t.Errorf("Expected 2 new activations, got %d", newActivations)
	}

	if !h.allActive() {
		t.Error("allActive() should be true after processing all events")
	}
}

// TestEarlyTerminationBetweenSignals simulates stopping
// after 3 of 5 signals.
func TestEarlyTerminationBetweenSignals(t *testing.T) {
	t.Parallel()

	h := &MaintainerActivityHandler{
		elevated: map[string]struct{}{"alice": {}, "bob": {}},
		active:   make(map[string]bool),
	}

	signalsChecked := 0

	// Simulate 5 signals, each marking one user active
	signals := []string{"alice", "bob", "charlie", "david", "eve"}

	for _, user := range signals {
		signalsChecked++
		h.markActive(user)

		if h.allActive() {
			// Early termination - stop checking remaining signals
			break
		}
	}

	// Should stop after 2 signals
	// (alice, bob) since both
	// elevated users are active
	if signalsChecked != 2 {
		t.Errorf("Expected to check 2 signals, but checked %d", signalsChecked)
	}

	if !h.allActive() {
		t.Error("allActive() should be true after early termination")
	}
}

// TestEarlyTerminationWithinNestedCalls tests nested loop termination.
func TestEarlyTerminationWithinNestedCalls(t *testing.T) {
	t.Parallel()

	h := &MaintainerActivityHandler{
		elevated: map[string]struct{}{"alice": {}, "bob": {}},
		active:   make(map[string]bool),
	}

	// Simulate processing items that
	// could each mark users active
	items := [][]string{
		{"alice"},
		{"bob"},
		{"charlie"},
		{"david"},
		{"eve"},
	}

	itemsProcessed := 0

	for _, users := range items {
		if h.allActive() {
			// Early termination - skip remaining items
			break
		}

		itemsProcessed++
		for _, user := range users {
			h.markActive(user)
		}
	}

	// Should process only 2 items
	// before all elevated users
	// are active
	if itemsProcessed != 2 {
		t.Errorf("Expected to process 2 items, but processed %d", itemsProcessed)
	}

	if !h.allActive() {
		t.Error("allActive() should be true after processing")
	}
}

// TestEarlyTerminationWithMarkActiveReturn tests combined pattern.
func TestEarlyTerminationWithMarkActiveReturn(t *testing.T) {
	t.Parallel()

	h := &MaintainerActivityHandler{
		elevated: map[string]struct{}{"alice": {}, "bob": {}},
		active:   make(map[string]bool),
	}

	// Simulate processing events with duplicate detection and early termination
	events := []string{"alice", "alice", "charlie", "bob", "david", "eve"}
	eventsProcessed := 0
	newActivations := 0

	for _, user := range events {
		if h.allActive() {
			break
		}

		eventsProcessed++
		if h.markActive(user) {
			newActivations++
		}
	}

	// Should process 4 events: alice, alice (dup),
	// charlie (not elevated), bob.
	// Then stop because allActive() is true
	if eventsProcessed != 4 {
		t.Errorf("Expected to process 4 events, but processed %d", eventsProcessed)
	}

	// Should have 2 new activations (alice, bob)
	if newActivations != 2 {
		t.Errorf("Expected 2 new activations, got %d", newActivations)
	}

	if !h.allActive() {
		t.Error("allActive() should be true after early termination")
	}
}
