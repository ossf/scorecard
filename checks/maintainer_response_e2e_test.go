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

package checks

import (
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
)

// Helper functions for creating test data.

func timePtr(year, month, day int) *time.Time {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return &t
}

func timePtrDaysAgo(daysAgo int) *time.Time {
	t := time.Now().AddDate(0, 0, -daysAgo)
	return &t
}

func strPtr(s string) *string {
	return &s
}

func makeUser(login string) *clients.User {
	return &clients.User{
		Login: login,
	}
}

func assocPtr(assoc clients.RepoAssociation) *clients.RepoAssociation {
	return &assoc
}

// makeIssue creates a test issue with common fields populated.
func makeIssue(issueNum, createdDaysAgo, closedDaysAgo int, labelEvents []clients.LabelEvent, comments []clients.IssueComment) clients.Issue {
	issue := clients.Issue{
		URI:         strPtr("https://example.com/issues/" + string(rune(issueNum+'0'))),
		IssueNumber: issueNum,
		CreatedAt:   timePtrDaysAgo(createdDaysAgo),
		Author: &clients.User{
			Login: "external-user",
		},
		AuthorAssociation: assocPtr(clients.RepoAssociationNone),
		LabelEvents:       labelEvents,
		AllLabelEvents:    labelEvents, // Populate AllLabelEvents for maintainer activity tracking
		Comments:          comments,
	}

	if closedDaysAgo >= 0 {
		issue.ClosedAt = timePtrDaysAgo(closedDaysAgo)
		// Add a CLOSED state change event when the issue is closed
		// Assume it's closed by a maintainer (not the issue author)
		issue.StateChangeEvents = []clients.StateChangeEvent{
			{
				CreatedAt:    time.Now().AddDate(0, 0, -closedDaysAgo),
				Closed:       true, // This is a close event
				Actor:        "maintainer",
				IsMaintainer: true,
			},
		}
	}

	return issue
}

// makeLabelEvent creates a label add/remove event.
func makeLabelEvent(label string, added bool, daysAgo int) clients.LabelEvent {
	return clients.LabelEvent{
		Label:     label,
		Added:     added,
		CreatedAt: time.Now().AddDate(0, 0, -daysAgo),
		Actor:     "some-actor",
	}
}

// makeMaintainerComment creates a comment from a maintainer.
func makeMaintainerComment(daysAgo int, association clients.RepoAssociation) clients.IssueComment {
	return clients.IssueComment{
		CreatedAt:         timePtrDaysAgo(daysAgo),
		Author:            makeUser("maintainer"),
		AuthorAssociation: assocPtr(association),
		IsMaintainer:      true,
	}
}

// E2E Tests.

func TestMaintainerResponse_E2E_PerfectScore(t *testing.T) {
	t.Parallel()
	// Setup: 10 bug/security issues, all with maintainer responses
	ctrl := gomock.NewController(t)
	mockRepo := mockrepo.NewMockRepoClient(ctrl)

	issues := []clients.Issue{
		// Issues labeled 'bug' with responses
		makeIssue(1, 30, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 30),
		}, []clients.IssueComment{
			makeMaintainerComment(29, clients.RepoAssociationMember),
		}),
		makeIssue(2, 25, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 25),
		}, []clients.IssueComment{
			makeMaintainerComment(24, clients.RepoAssociationOwner),
		}),
		makeIssue(3, 20, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 20),
		}, []clients.IssueComment{
			makeMaintainerComment(19, clients.RepoAssociationCollaborator),
		}),
		// Security issues with responses
		makeIssue(4, 15, -1, []clients.LabelEvent{
			makeLabelEvent("security", true, 15),
		}, []clients.IssueComment{
			makeMaintainerComment(14, clients.RepoAssociationMember),
		}),
		makeIssue(5, 10, -1, []clients.LabelEvent{
			makeLabelEvent("security", true, 10),
		}, []clients.IssueComment{
			makeMaintainerComment(9, clients.RepoAssociationOwner),
		}),
		// Issues with both labels
		makeIssue(6, 8, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 8),
			makeLabelEvent("security", true, 8),
		}, []clients.IssueComment{
			makeMaintainerComment(7, clients.RepoAssociationMember),
		}),
		makeIssue(7, 7, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 7),
			makeLabelEvent("security", true, 7),
		}, []clients.IssueComment{
			makeMaintainerComment(6, clients.RepoAssociationOwner),
		}),
		makeIssue(8, 6, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 6),
		}, []clients.IssueComment{
			makeMaintainerComment(5, clients.RepoAssociationMember),
		}),
		makeIssue(9, 5, -1, []clients.LabelEvent{
			makeLabelEvent("security", true, 5),
		}, []clients.IssueComment{
			makeMaintainerComment(4, clients.RepoAssociationOwner),
		}),
		makeIssue(10, 4, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 4),
		}, []clients.IssueComment{
			makeMaintainerComment(3, clients.RepoAssociationMember),
		}),
	}

	mockRepo.EXPECT().ListIssuesWithHistory().Return(issues, nil)

	req := &checker.CheckRequest{
		Ctx:        t.Context(),
		RepoClient: mockRepo,
		Dlogger:    &testDetailLogger{},
	}

	result := MaintainerResponse(req)

	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	if result.Score != 10 {
		t.Errorf("Expected score 10, got %d", result.Score)
	}
	// With 0 violations, the reason should say "All X had timely maintainer activity"
	if !strings.Contains(result.Reason, "All 10 had timely maintainer activity") {
		t.Errorf("Expected reason to contain 'All 10 had timely maintainer activity', got: %s", result.Reason)
	}
}

func TestMaintainerResponse_E2E_Score10Boundary(t *testing.T) {
	t.Parallel()
	// Setup: 10 issues, exactly 2 unresponsive and >= 180 days old (20% - boundary for score 10)
	// Key: Issues must be labeled >= 180 days ago without response to count as violations
	ctrl := gomock.NewController(t)
	mockRepo := mockrepo.NewMockRepoClient(ctrl)

	issues := []clients.Issue{
		// 8 issues with responses (all recent, won't be violations)
		makeIssue(1, 30, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 30),
		}, []clients.IssueComment{
			makeMaintainerComment(29, clients.RepoAssociationMember),
		}),
		makeIssue(2, 25, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 25),
		}, []clients.IssueComment{
			makeMaintainerComment(24, clients.RepoAssociationOwner),
		}),
		makeIssue(3, 20, -1, []clients.LabelEvent{
			makeLabelEvent("security", true, 20),
		}, []clients.IssueComment{
			makeMaintainerComment(19, clients.RepoAssociationMember),
		}),
		makeIssue(4, 15, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 15),
		}, []clients.IssueComment{
			makeMaintainerComment(14, clients.RepoAssociationOwner),
		}),
		makeIssue(5, 10, -1, []clients.LabelEvent{
			makeLabelEvent("security", true, 10),
		}, []clients.IssueComment{
			makeMaintainerComment(9, clients.RepoAssociationMember),
		}),
		makeIssue(6, 8, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 8),
		}, []clients.IssueComment{
			makeMaintainerComment(7, clients.RepoAssociationOwner),
		}),
		makeIssue(7, 7, -1, []clients.LabelEvent{
			makeLabelEvent("security", true, 7),
		}, []clients.IssueComment{
			makeMaintainerComment(6, clients.RepoAssociationMember),
		}),
		makeIssue(8, 6, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 6),
		}, []clients.IssueComment{
			makeMaintainerComment(5, clients.RepoAssociationOwner),
		}),
		// 2 issues WITHOUT responses AND >= 180 days old (these are violations)
		// With the fix to createInterval, open issues can now be violations
		makeIssue(9, 200, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 200), // Label added 200 days ago, still present
		}, []clients.IssueComment{
			// NO COMMENTS - any comment counts as reaction
		}),
		makeIssue(10, 190, -1, []clients.LabelEvent{
			makeLabelEvent("security", true, 190), // Label added 190 days ago, still present
		}, []clients.IssueComment{
			// NO COMMENTS
		}),
	}

	mockRepo.EXPECT().ListIssuesWithHistory().Return(issues, nil)

	req := &checker.CheckRequest{
		Ctx:        t.Context(),
		RepoClient: mockRepo,
		Dlogger:    &testDetailLogger{},
	}

	result := MaintainerResponse(req)

	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	if result.Score != 10 {
		t.Errorf("Expected score 10 (exactly 20%% boundary), got %d", result.Score)
	}
	if !strings.Contains(result.Reason, "20") {
		t.Errorf("Expected reason to contain 20%% violations, got: %s", result.Reason)
	}
}

func TestMaintainerResponse_E2E_Score5(t *testing.T) {
	t.Parallel()
	// Setup: 8 issues, 2 unresponsive and >= 180 days old (25% violations -> score 5)
	ctrl := gomock.NewController(t)
	mockRepo := mockrepo.NewMockRepoClient(ctrl)

	issues := []clients.Issue{
		// 6 issues with responses (won't be violations)
		makeIssue(1, 30, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 30),
		}, []clients.IssueComment{
			makeMaintainerComment(29, clients.RepoAssociationMember),
		}),
		makeIssue(2, 25, -1, []clients.LabelEvent{
			makeLabelEvent("security", true, 25),
		}, []clients.IssueComment{
			makeMaintainerComment(24, clients.RepoAssociationOwner),
		}),
		makeIssue(3, 20, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 20),
		}, []clients.IssueComment{
			makeMaintainerComment(19, clients.RepoAssociationCollaborator),
		}),
		makeIssue(4, 15, -1, []clients.LabelEvent{
			makeLabelEvent("security", true, 15),
		}, []clients.IssueComment{
			makeMaintainerComment(14, clients.RepoAssociationMember),
		}),
		makeIssue(5, 10, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 10),
		}, []clients.IssueComment{
			makeMaintainerComment(9, clients.RepoAssociationOwner),
		}),
		makeIssue(6, 8, -1, []clients.LabelEvent{
			makeLabelEvent("security", true, 8),
		}, []clients.IssueComment{
			makeMaintainerComment(7, clients.RepoAssociationMember),
		}),
		// 2 issues WITHOUT responses AND >= 180 days old (violations)
		makeIssue(7, 200, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 200),
		}, []clients.IssueComment{}),
		makeIssue(8, 185, -1, []clients.LabelEvent{
			makeLabelEvent("security", true, 185),
		}, []clients.IssueComment{}),
	}

	mockRepo.EXPECT().ListIssuesWithHistory().Return(issues, nil)

	req := &checker.CheckRequest{
		Ctx:        t.Context(),
		RepoClient: mockRepo,
		Dlogger:    &testDetailLogger{},
	}

	result := MaintainerResponse(req)

	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	if result.Score != 5 {
		t.Errorf("Expected score 5, got %d", result.Score)
	}
	if !strings.Contains(result.Reason, "25") {
		t.Errorf("Expected reason to contain 25%% violations, got: %s", result.Reason)
	}
}

func TestMaintainerResponse_E2E_Score5Boundary(t *testing.T) {
	t.Parallel()
	// Setup: 10 issues, exactly 4 unresponsive and >= 180 days old (40% - boundary for score 5)
	ctrl := gomock.NewController(t)
	mockRepo := mockrepo.NewMockRepoClient(ctrl)

	issues := []clients.Issue{
		// 6 issues with responses
		makeIssue(1, 30, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 30),
		}, []clients.IssueComment{
			makeMaintainerComment(29, clients.RepoAssociationMember),
		}),
		makeIssue(2, 25, -1, []clients.LabelEvent{
			makeLabelEvent("security", true, 25),
		}, []clients.IssueComment{
			makeMaintainerComment(24, clients.RepoAssociationOwner),
		}),
		makeIssue(3, 20, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 20),
		}, []clients.IssueComment{
			makeMaintainerComment(19, clients.RepoAssociationMember),
		}),
		makeIssue(4, 15, -1, []clients.LabelEvent{
			makeLabelEvent("security", true, 15),
		}, []clients.IssueComment{
			makeMaintainerComment(14, clients.RepoAssociationOwner),
		}),
		makeIssue(5, 10, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 10),
		}, []clients.IssueComment{
			makeMaintainerComment(9, clients.RepoAssociationMember),
		}),
		makeIssue(6, 8, -1, []clients.LabelEvent{
			makeLabelEvent("security", true, 8),
		}, []clients.IssueComment{
			makeMaintainerComment(7, clients.RepoAssociationOwner),
		}),
		// 4 issues WITHOUT responses AND >= 180 days old (40% violations)
		makeIssue(7, 200, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 200),
		}, []clients.IssueComment{}),
		makeIssue(8, 195, -1, []clients.LabelEvent{
			makeLabelEvent("security", true, 195),
		}, []clients.IssueComment{}),
		makeIssue(9, 185, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 185),
		}, []clients.IssueComment{}),
		makeIssue(10, 180, -1, []clients.LabelEvent{
			makeLabelEvent("security", true, 180),
		}, []clients.IssueComment{}),
	}

	mockRepo.EXPECT().ListIssuesWithHistory().Return(issues, nil)

	req := &checker.CheckRequest{
		Ctx:        t.Context(),
		RepoClient: mockRepo,
		Dlogger:    &testDetailLogger{},
	}

	result := MaintainerResponse(req)

	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	if result.Score != 5 {
		t.Errorf("Expected score 5 (exactly 40%% boundary), got %d", result.Score)
	}
	if !strings.Contains(result.Reason, "40") {
		t.Errorf("Expected reason to contain 40%% violations, got: %s", result.Reason)
	}
}

func TestMaintainerResponse_E2E_Score0(t *testing.T) {
	t.Parallel()
	// Setup: 10 issues, 5 unresponsive and >= 180 days old (50% violations -> score 0)
	ctrl := gomock.NewController(t)
	mockRepo := mockrepo.NewMockRepoClient(ctrl)

	issues := []clients.Issue{
		// 5 issues with responses
		makeIssue(1, 30, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 30),
		}, []clients.IssueComment{
			makeMaintainerComment(29, clients.RepoAssociationMember),
		}),
		makeIssue(2, 25, -1, []clients.LabelEvent{
			makeLabelEvent("security", true, 25),
		}, []clients.IssueComment{
			makeMaintainerComment(24, clients.RepoAssociationOwner),
		}),
		makeIssue(3, 20, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 20),
		}, []clients.IssueComment{
			makeMaintainerComment(19, clients.RepoAssociationMember),
		}),
		makeIssue(4, 15, -1, []clients.LabelEvent{
			makeLabelEvent("security", true, 15),
		}, []clients.IssueComment{
			makeMaintainerComment(14, clients.RepoAssociationOwner),
		}),
		makeIssue(5, 10, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 10),
		}, []clients.IssueComment{
			makeMaintainerComment(9, clients.RepoAssociationMember),
		}),
		// 5 issues WITHOUT responses AND >= 180 days old (50% violations)
		makeIssue(6, 250, -1, []clients.LabelEvent{
			makeLabelEvent("security", true, 250),
		}, []clients.IssueComment{}),
		makeIssue(7, 220, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 220),
		}, []clients.IssueComment{}),
		makeIssue(8, 200, -1, []clients.LabelEvent{
			makeLabelEvent("security", true, 200),
		}, []clients.IssueComment{}),
		makeIssue(9, 190, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 190),
		}, []clients.IssueComment{}),
		makeIssue(10, 180, -1, []clients.LabelEvent{
			makeLabelEvent("security", true, 180),
		}, []clients.IssueComment{}),
	}

	mockRepo.EXPECT().ListIssuesWithHistory().Return(issues, nil)

	req := &checker.CheckRequest{
		Ctx:        t.Context(),
		RepoClient: mockRepo,
		Dlogger:    &testDetailLogger{},
	}

	result := MaintainerResponse(req)

	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	if result.Score != 0 {
		t.Errorf("Expected score 0, got %d", result.Score)
	}
	if !strings.Contains(result.Reason, "50") {
		t.Errorf("Expected reason to contain 50%% violations, got: %s", result.Reason)
	}
}

func TestMaintainerResponse_E2E_NoIssues(t *testing.T) {
	t.Parallel()
	// Setup: No issues at all
	ctrl := gomock.NewController(t)
	mockRepo := mockrepo.NewMockRepoClient(ctrl)

	issues := []clients.Issue{} // Empty

	mockRepo.EXPECT().ListIssuesWithHistory().Return(issues, nil)

	req := &checker.CheckRequest{
		Ctx:        t.Context(),
		RepoClient: mockRepo,
		Dlogger:    &testDetailLogger{},
	}

	result := MaintainerResponse(req)

	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	// When no issues, evaluation gives max score
	if result.Score != 10 {
		t.Errorf("Expected score 10 (max) when no issues, got %d", result.Score)
	}
	if !strings.Contains(strings.ToLower(result.Reason), "no issues") {
		t.Errorf("Expected reason to mention no issues, got: %s", result.Reason)
	}
}

func TestMaintainerResponse_E2E_NoRelevantLabels(t *testing.T) {
	t.Parallel()
	// Setup: Issues exist but none have bug/security labels
	ctrl := gomock.NewController(t)
	mockRepo := mockrepo.NewMockRepoClient(ctrl)

	issues := []clients.Issue{
		makeIssue(1, 30, -1, []clients.LabelEvent{
			makeLabelEvent("enhancement", true, 30), // Not bug/security
		}, []clients.IssueComment{
			makeMaintainerComment(29, clients.RepoAssociationMember),
		}),
		makeIssue(2, 25, -1, []clients.LabelEvent{
			makeLabelEvent("documentation", true, 25), // Not bug/security
		}, []clients.IssueComment{
			makeMaintainerComment(24, clients.RepoAssociationOwner),
		}),
		makeIssue(3, 20, -1, []clients.LabelEvent{
			makeLabelEvent("question", true, 20), // Not bug/security
		}, []clients.IssueComment{}),
	}

	mockRepo.EXPECT().ListIssuesWithHistory().Return(issues, nil)

	req := &checker.CheckRequest{
		Ctx:        t.Context(),
		RepoClient: mockRepo,
		Dlogger:    &testDetailLogger{},
	}

	result := MaintainerResponse(req)

	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	// When no bug/security labels, evaluation gives max score with explanatory reason
	if result.Score != 10 {
		t.Errorf("Expected score 10 (max) when no bug/security labels, got %d", result.Score)
	}
	if !strings.Contains(strings.ToLower(result.Reason), "no issues") {
		t.Errorf("Expected reason to mention no issues, got: %s", result.Reason)
	}
}

func TestMaintainerResponse_E2E_ComplexLabelTimeline(t *testing.T) {
	t.Parallel()
	// Setup: Issue where bug label is removed then re-added (multiple intervals)
	ctrl := gomock.NewController(t)
	mockRepo := mockrepo.NewMockRepoClient(ctrl)

	issues := []clients.Issue{
		makeIssue(1, 30, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 30),  // Added
			makeLabelEvent("bug", false, 20), // Removed
			makeLabelEvent("bug", true, 10),  // Re-added
		}, []clients.IssueComment{
			makeMaintainerComment(28, clients.RepoAssociationMember), // Response during first interval
			makeMaintainerComment(8, clients.RepoAssociationOwner),   // Response during second interval
		}),
		// Add another simple issue to ensure we have valid data
		makeIssue(2, 25, -1, []clients.LabelEvent{
			makeLabelEvent("security", true, 25),
		}, []clients.IssueComment{
			makeMaintainerComment(24, clients.RepoAssociationMember),
		}),
	}

	mockRepo.EXPECT().ListIssuesWithHistory().Return(issues, nil)

	req := &checker.CheckRequest{
		Ctx:        t.Context(),
		RepoClient: mockRepo,
		Dlogger:    &testDetailLogger{},
	}

	result := MaintainerResponse(req)

	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	// Both intervals should have responses, so should get good score
	if result.Score < 5 {
		t.Errorf("Expected decent score with responses in both intervals, got %d", result.Score)
	}
}

func TestMaintainerResponse_E2E_MixedGitHubAssociations(t *testing.T) {
	t.Parallel()
	// Setup: Test different GitHub association types (MEMBER, OWNER, COLLABORATOR)
	// Only issue without maintainer response AND >= 180 days is a violation
	ctrl := gomock.NewController(t)
	mockRepo := mockrepo.NewMockRepoClient(ctrl)

	issues := []clients.Issue{
		// Issue with MEMBER response
		makeIssue(1, 30, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 30),
		}, []clients.IssueComment{
			makeMaintainerComment(29, clients.RepoAssociationMember),
		}),
		// Issue with OWNER response
		makeIssue(2, 25, -1, []clients.LabelEvent{
			makeLabelEvent("security", true, 25),
		}, []clients.IssueComment{
			makeMaintainerComment(24, clients.RepoAssociationOwner),
		}),
		// Issue with COLLABORATOR response
		makeIssue(3, 20, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 20),
		}, []clients.IssueComment{
			makeMaintainerComment(19, clients.RepoAssociationCollaborator),
		}),
		// Issue with no maintainer response AND >= 180 days old (VIOLATION)
		makeIssue(4, 200, -1, []clients.LabelEvent{
			makeLabelEvent("security", true, 200),
		}, []clients.IssueComment{
			// NO COMMENTS AT ALL
		}),
	}

	mockRepo.EXPECT().ListIssuesWithHistory().Return(issues, nil)

	req := &checker.CheckRequest{
		Ctx:        t.Context(),
		RepoClient: mockRepo,
		Dlogger:    &testDetailLogger{},
	}

	result := MaintainerResponse(req)

	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	// 1 out of 4 is a violation (25%) -> score 5
	if result.Score != 5 {
		t.Errorf("Expected score 5 (25%% violations), got %d", result.Score)
	}
	if !strings.Contains(result.Reason, "25") {
		t.Errorf("Expected reason to contain 25%% violations, got: %s", result.Reason)
	}
}

func TestMaintainerResponse_E2E_ClosedIssues(t *testing.T) {
	t.Parallel()
	// Setup: Mix of open and closed issues with/without responses
	// Closing counts as a reaction, so closed issues without comments before closing are NOT violations
	ctrl := gomock.NewController(t)
	mockRepo := mockrepo.NewMockRepoClient(ctrl)

	issues := []clients.Issue{
		// Closed issue >= 180 days with response before closing (not a violation - has response)
		makeIssue(1, 200, 100, []clients.LabelEvent{
			makeLabelEvent("bug", true, 200),
		}, []clients.IssueComment{
			makeMaintainerComment(195, clients.RepoAssociationMember),
		}),
		// Closed issue >= 180 days without comments (not a violation - closing counts as reaction)
		makeIssue(2, 220, 150, []clients.LabelEvent{
			makeLabelEvent("security", true, 220),
		}, []clients.IssueComment{}),
		// Open issue with response (not a violation - has response)
		makeIssue(3, 30, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 30),
		}, []clients.IssueComment{
			makeMaintainerComment(29, clients.RepoAssociationOwner),
		}),
		// Open issue >= 180 days without response (VIOLATION - the fix enables this)
		makeIssue(4, 190, -1, []clients.LabelEvent{
			makeLabelEvent("security", true, 190),
		}, []clients.IssueComment{}),
	}

	mockRepo.EXPECT().ListIssuesWithHistory().Return(issues, nil)

	req := &checker.CheckRequest{
		Ctx:        t.Context(),
		RepoClient: mockRepo,
		Dlogger:    &testDetailLogger{},
	}

	result := MaintainerResponse(req)

	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	// 1 out of 4 is a violation (25%) -> score 5
	if result.Score != 5 {
		t.Errorf("Expected score 5 (25%% violations), got %d", result.Score)
	}
	if !strings.Contains(result.Reason, "25") {
		t.Errorf("Expected reason to contain 25%% violations, got: %s", result.Reason)
	}
}

func TestMaintainerResponse_E2E_ResponseTimingEdgeCases(t *testing.T) {
	t.Parallel()
	// Setup: Test timing edge cases - issues must be >= 180 days old to be violations
	ctrl := gomock.NewController(t)
	mockRepo := mockrepo.NewMockRepoClient(ctrl)

	now := time.Now()

	issues := []clients.Issue{
		// Response same day as label, but recent (not a violation - not old enough)
		{
			URI:         strPtr("https://example.com/issues/1"),
			IssueNumber: 1,
			CreatedAt:   timePtrDaysAgo(30),
			LabelEvents: []clients.LabelEvent{
				{
					Label:     "bug",
					Added:     true,
					CreatedAt: now.AddDate(0, 0, -10),
					Actor:     "someone",
				},
			},
			Comments: []clients.IssueComment{
				{
					CreatedAt:         timePtr(now.AddDate(0, 0, -10).Year(), int(now.AddDate(0, 0, -10).Month()), now.AddDate(0, 0, -10).Day()),
					Author:            makeUser("maintainer"),
					AuthorAssociation: assocPtr(clients.RepoAssociationMember),
					IsMaintainer:      true,
				},
			},
		},
		// Response before label, >= 180 days old (VIOLATION - response doesn't count)
		{
			URI:         strPtr("https://example.com/issues/2"),
			IssueNumber: 2,
			CreatedAt:   timePtrDaysAgo(300),
			LabelEvents: []clients.LabelEvent{
				{
					Label:     "security",
					Added:     true,
					CreatedAt: now.AddDate(0, 0, -200),
					Actor:     "someone",
				},
			},
			Comments: []clients.IssueComment{
				{
					CreatedAt:         timePtrDaysAgo(210), // Before label was added
					Author:            makeUser("maintainer"),
					AuthorAssociation: assocPtr(clients.RepoAssociationMember),
					IsMaintainer:      true,
				},
			},
		},
		// Response well after label, but issue is recent (not a violation - not old enough)
		makeIssue(3, 30, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 30),
		}, []clients.IssueComment{
			makeMaintainerComment(5, clients.RepoAssociationMember), // Much later but issue is recent
		}),
	}

	mockRepo.EXPECT().ListIssuesWithHistory().Return(issues, nil)

	req := &checker.CheckRequest{
		Ctx:        t.Context(),
		RepoClient: mockRepo,
		Dlogger:    &testDetailLogger{},
	}

	result := MaintainerResponse(req)

	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	// Issue 1: not old enough (no violation)
	// Issue 2: >= 180 days old, response before label (VIOLATION)
	// Issue 3: not old enough (no violation)
	// Total: 3 evaluated, 1 violation = 33% -> score 5
	if result.Score < 5 || result.Score > 10 {
		t.Errorf("Expected score between 5-10 (33%% violations), got %d", result.Score)
	}
}

func TestMaintainerResponse_E2E_BothLabelsOnSameIssue(t *testing.T) {
	t.Parallel()
	// Setup: Issues with both "bug" and "security" labels
	// Note: The probe emits ONE finding per ISSUE, not per label interval.
	// If an issue has both labels and either violates, the whole issue is a violation.
	ctrl := gomock.NewController(t)
	mockRepo := mockrepo.NewMockRepoClient(ctrl)

	issues := []clients.Issue{
		// Issue with both labels, both get responses (1 issue, not a violation)
		makeIssue(1, 30, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 30),
			makeLabelEvent("security", true, 30),
		}, []clients.IssueComment{
			makeMaintainerComment(29, clients.RepoAssociationMember),
		}),
		// Issue with both labels >= 180 days, no response (1 issue, 1 violation)
		makeIssue(2, 200, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 200),
			makeLabelEvent("security", true, 200),
		}, []clients.IssueComment{}),
		// Regular bug issue with response (1 issue, not a violation)
		makeIssue(3, 20, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 20),
		}, []clients.IssueComment{
			makeMaintainerComment(19, clients.RepoAssociationOwner),
		}),
	}

	mockRepo.EXPECT().ListIssuesWithHistory().Return(issues, nil)

	req := &checker.CheckRequest{
		Ctx:        t.Context(),
		RepoClient: mockRepo,
		Dlogger:    &testDetailLogger{},
	}

	result := MaintainerResponse(req)

	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	// 3 issues evaluated, 1 violation = 33.3% -> score 5
	// (Probe emits one finding per issue, not per label interval)
	if result.Score != 5 {
		t.Errorf("Expected score 5 (33.3%% violations), got %d", result.Score)
	}
	if !strings.Contains(result.Reason, "33") {
		t.Errorf("Expected reason to contain 33%% violations, got: %s", result.Reason)
	}
}

func TestMaintainerResponse_E2E_RepoClientError(t *testing.T) {
	t.Parallel()
	// Setup: RepoClient returns an error
	ctrl := gomock.NewController(t)
	mockRepo := mockrepo.NewMockRepoClient(ctrl)

	mockRepo.EXPECT().ListIssuesWithHistory().Return(nil, clients.ErrUnsupportedFeature)

	req := &checker.CheckRequest{
		Ctx:        t.Context(),
		RepoClient: mockRepo,
		Dlogger:    &testDetailLogger{},
	}

	result := MaintainerResponse(req)

	if result.Error == nil {
		t.Errorf("Expected error when RepoClient fails, got nil")
	}
	if result.Score != -1 {
		t.Errorf("Expected inconclusive score (-1) on error, got %d", result.Score)
	}
}

func TestMaintainerResponse_E2E_NewLabels_KindBug(t *testing.T) {
	t.Parallel()
	// Test the new "kind/bug" label
	ctrl := gomock.NewController(t)
	mockRepo := mockrepo.NewMockRepoClient(ctrl)

	issues := []clients.Issue{
		// Issue with kind/bug label and response
		makeIssue(1, 30, -1, []clients.LabelEvent{
			makeLabelEvent("kind/bug", true, 30),
		}, []clients.IssueComment{
			makeMaintainerComment(29, clients.RepoAssociationMember),
		}),
		// Issue with kind/bug label >= 180 days without response (violation)
		makeIssue(2, 200, -1, []clients.LabelEvent{
			makeLabelEvent("kind/bug", true, 200),
		}, []clients.IssueComment{}),
	}

	mockRepo.EXPECT().ListIssuesWithHistory().Return(issues, nil)

	req := &checker.CheckRequest{
		Ctx:        t.Context(),
		RepoClient: mockRepo,
		Dlogger:    &testDetailLogger{},
	}

	result := MaintainerResponse(req)

	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	// 2 issues evaluated, 1 violation = 50% -> score 0
	if result.Score != 0 {
		t.Errorf("Expected score 0 (50%% violations), got %d", result.Score)
	}
	if !strings.Contains(result.Reason, "50") {
		t.Errorf("Expected reason to contain 50%% violations, got: %s", result.Reason)
	}
}

func TestMaintainerResponse_E2E_NewLabels_AreaSecurity(t *testing.T) {
	t.Parallel()
	// Test the new "area/security" label
	ctrl := gomock.NewController(t)
	mockRepo := mockrepo.NewMockRepoClient(ctrl)

	issues := []clients.Issue{
		// Issue with area/security label and response
		makeIssue(1, 30, -1, []clients.LabelEvent{
			makeLabelEvent("area/security", true, 30),
		}, []clients.IssueComment{
			makeMaintainerComment(29, clients.RepoAssociationOwner),
		}),
		// Issue with area/security label >= 180 days without response (violation)
		makeIssue(2, 190, -1, []clients.LabelEvent{
			makeLabelEvent("area/security", true, 190),
		}, []clients.IssueComment{}),
		// Issue with response
		makeIssue(3, 50, -1, []clients.LabelEvent{
			makeLabelEvent("area/security", true, 50),
		}, []clients.IssueComment{
			makeMaintainerComment(45, clients.RepoAssociationCollaborator),
		}),
	}

	mockRepo.EXPECT().ListIssuesWithHistory().Return(issues, nil)

	req := &checker.CheckRequest{
		Ctx:        t.Context(),
		RepoClient: mockRepo,
		Dlogger:    &testDetailLogger{},
	}

	result := MaintainerResponse(req)

	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	// 3 issues evaluated, 1 violation = 33.3% -> score 5
	if result.Score != 5 {
		t.Errorf("Expected score 5 (33.3%% violations), got %d", result.Score)
	}
	if !strings.Contains(result.Reason, "33") {
		t.Errorf("Expected reason to contain 33%% violations, got: %s", result.Reason)
	}
}

func TestMaintainerResponse_E2E_NewLabels_AreaProductSecurity(t *testing.T) {
	t.Parallel()
	// Test the new "area/product security" label
	ctrl := gomock.NewController(t)
	mockRepo := mockrepo.NewMockRepoClient(ctrl)

	issues := []clients.Issue{
		// Issue with area/product security label >= 180 days without response (violation)
		makeIssue(1, 220, -1, []clients.LabelEvent{
			makeLabelEvent("area/product security", true, 220),
		}, []clients.IssueComment{}),
		// Issue with area/product security label and response
		makeIssue(2, 40, -1, []clients.LabelEvent{
			makeLabelEvent("area/product security", true, 40),
		}, []clients.IssueComment{
			makeMaintainerComment(39, clients.RepoAssociationMember),
		}),
	}

	mockRepo.EXPECT().ListIssuesWithHistory().Return(issues, nil)

	req := &checker.CheckRequest{
		Ctx:        t.Context(),
		RepoClient: mockRepo,
		Dlogger:    &testDetailLogger{},
	}

	result := MaintainerResponse(req)

	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	// 2 issues evaluated, 1 violation = 50% -> score 0
	if result.Score != 0 {
		t.Errorf("Expected score 0 (50%% violations), got %d", result.Score)
	}
	if !strings.Contains(result.Reason, "50") {
		t.Errorf("Expected reason to contain 50%% violations, got: %s", result.Reason)
	}
}

func TestMaintainerResponse_E2E_NewLabels_Mixed(t *testing.T) {
	t.Parallel()
	// Test mixing old and new labels
	ctrl := gomock.NewController(t)
	mockRepo := mockrepo.NewMockRepoClient(ctrl)

	issues := []clients.Issue{
		// Old bug label with response
		makeIssue(1, 30, -1, []clients.LabelEvent{
			makeLabelEvent("bug", true, 30),
		}, []clients.IssueComment{
			makeMaintainerComment(29, clients.RepoAssociationMember),
		}),
		// New kind/bug label with violation
		makeIssue(2, 200, -1, []clients.LabelEvent{
			makeLabelEvent("kind/bug", true, 200),
		}, []clients.IssueComment{}),
		// Old security label with response
		makeIssue(3, 50, -1, []clients.LabelEvent{
			makeLabelEvent("security", true, 50),
		}, []clients.IssueComment{
			makeMaintainerComment(48, clients.RepoAssociationOwner),
		}),
		// New area/security label with violation
		makeIssue(4, 195, -1, []clients.LabelEvent{
			makeLabelEvent("area/security", true, 195),
		}, []clients.IssueComment{}),
		// New area/product security with response
		makeIssue(5, 60, -1, []clients.LabelEvent{
			makeLabelEvent("area/product security", true, 60),
		}, []clients.IssueComment{
			makeMaintainerComment(55, clients.RepoAssociationCollaborator),
		}),
	}

	mockRepo.EXPECT().ListIssuesWithHistory().Return(issues, nil)

	req := &checker.CheckRequest{
		Ctx:        t.Context(),
		RepoClient: mockRepo,
		Dlogger:    &testDetailLogger{},
	}

	result := MaintainerResponse(req)

	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	// 5 issues evaluated, 2 violations = 40% -> score 5
	if result.Score != 5 {
		t.Errorf("Expected score 5 (40%% violations), got %d", result.Score)
	}
	if !strings.Contains(result.Reason, "40") {
		t.Errorf("Expected reason to contain 40%% violations, got: %s", result.Reason)
	}
}

// Simple test logger implementation.
type testDetailLogger struct {
	details []checker.CheckDetail
}

func (l *testDetailLogger) Info(msg *checker.LogMessage) {
	l.details = append(l.details, checker.CheckDetail{
		Type: checker.DetailInfo,
		Msg:  *msg,
	})
}

func (l *testDetailLogger) Warn(msg *checker.LogMessage) {
	l.details = append(l.details, checker.CheckDetail{
		Type: checker.DetailWarn,
		Msg:  *msg,
	})
}

func (l *testDetailLogger) Debug(msg *checker.LogMessage) {
	l.details = append(l.details, checker.CheckDetail{
		Type: checker.DetailDebug,
		Msg:  *msg,
	})
}

func (l *testDetailLogger) Flush() []checker.CheckDetail {
	ret := l.details
	l.details = nil
	return ret
}
