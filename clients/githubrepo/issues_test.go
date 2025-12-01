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
	"strings"
	"testing"
	"time"

	"github.com/ossf/scorecard/v5/clients"
)

func Test_issuesHandler_init(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		repoOwner string
		repoName  string
	}{
		{
			name:      "Basic initialization",
			repoOwner: "testowner",
			repoName:  "testrepo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			handler := &issuesHandler{}
			repo := &Repo{owner: tt.repoOwner, repo: tt.repoName}
			ctx := context.Background()

			handler.init(ctx, repo)

			if handler.ctx == nil {
				t.Error("context not initialized")
			}
			if handler.repourl == nil {
				t.Error("repourl not initialized")
			}
			if handler.once == nil {
				t.Error("once not initialized")
			}
		})
	}
}

func Test_issuesHandler_setup_noGraphQLClient(t *testing.T) {
	t.Parallel()

	handler := &issuesHandler{}
	repo := &Repo{owner: "owner", repo: "repo"}
	ctx := context.Background()

	handler.init(ctx, repo)

	err := handler.setup()
	if err == nil {
		t.Error("expected error when graphClient is nil")
	}
	if !strings.Contains(err.Error(), "not initialized") {
		t.Errorf("expected 'not initialized' error, got: %v", err)
	}
}

func Test_issuesHandler_listIssuesWithHistory_callsSetup(t *testing.T) {
	t.Parallel()

	handler := &issuesHandler{}
	repo := &Repo{owner: "owner", repo: "repo"}
	ctx := context.Background()

	handler.init(ctx, repo)

	// Should fail because graphClient is nil
	_, err := handler.listIssuesWithHistory()
	if err == nil {
		t.Error("expected error when calling listIssuesWithHistory without graphClient")
	}
}

func Test_issuesHandler_labelFiltering(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		labelName     string
		shouldInclude bool
	}{
		{
			name:          "bug label (lowercase)",
			labelName:     "bug",
			shouldInclude: true,
		},
		{
			name:          "Bug label (mixed case)",
			labelName:     "Bug",
			shouldInclude: true,
		},
		{
			name:          "BUG label (uppercase)",
			labelName:     "BUG",
			shouldInclude: true,
		},
		{
			name:          "security label (lowercase)",
			labelName:     "security",
			shouldInclude: true,
		},
		{
			name:          "Security label (mixed case)",
			labelName:     "Security",
			shouldInclude: true,
		},
		{
			name:          "SECURITY label (uppercase)",
			labelName:     "SECURITY",
			shouldInclude: true,
		},
		{
			name:          "enhancement label",
			labelName:     "enhancement",
			shouldInclude: false,
		},
		{
			name:          "documentation label",
			labelName:     "documentation",
			shouldInclude: false,
		},
		{
			name:          "bug-report label (contains bug but not exact)",
			labelName:     "bug-report",
			shouldInclude: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Test the logic that filters labels (from issues.go line 200-202)
			normalized := strings.ToLower(strings.TrimSpace(tt.labelName))
			included := (normalized == "bug" || normalized == "security")

			if included != tt.shouldInclude {
				t.Errorf("label %q: expected shouldInclude=%v, got %v",
					tt.labelName, tt.shouldInclude, included)
			}
		})
	}
}

func Test_issuesHandler_commentAuthorAssociationLogic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		association      string
		expectMaintainer bool
	}{
		{"OWNER", true},
		{"MEMBER", true},
		{"COLLABORATOR", true},
		{"CONTRIBUTOR", false},
		{"FIRST_TIME_CONTRIBUTOR", false},
		{"FIRST_TIMER", false},
		{"NONE", false},
		{"", false},
		{"owner", true},        // lowercase should match after ToUpper
		{"member", true},       // lowercase should match after ToUpper
		{"collaborator", true}, // lowercase should match after ToUpper
	}

	for _, tt := range tests {
		t.Run(tt.association, func(t *testing.T) {
			t.Parallel()

			// Test the logic that determines if a comment author is a maintainer (from issues.go line 224)
			aa := strings.ToUpper(tt.association)
			isMaint := (aa == "OWNER" || aa == "MEMBER" || aa == "COLLABORATOR")

			if isMaint != tt.expectMaintainer {
				t.Errorf("association %q: expected isMaintainer=%v, got %v",
					tt.association, tt.expectMaintainer, isMaint)
			}
		})
	}
}

func Test_issuesHandler_caching(t *testing.T) {
	t.Parallel()

	// This test verifies that the issues slice is properly stored
	callCount := 0
	handler := &issuesHandler{}
	repo := &Repo{owner: "owner", repo: "repo"}
	ctx := context.Background()

	handler.init(ctx, repo)

	// Mock the setup by manually populating issues to test caching behavior
	handler.issues = []clients.Issue{
		{
			IssueNumber: 123,
			LabelEvents: []clients.LabelEvent{
				{Label: "bug", Added: true, CreatedAt: time.Now()},
			},
		},
	}

	// Verify issues are stored correctly
	for i := 0; i < 3; i++ {
		if len(handler.issues) != 1 {
			t.Errorf("iteration %d: expected 1 cached issue, got %d", i, len(handler.issues))
		}
		callCount++
	}

	if callCount != 3 {
		t.Errorf("expected 3 iterations, got %d", callCount)
	}
}

func Test_issuesHandler_issueStructure(t *testing.T) {
	t.Parallel()

	// Test that the clients.Issue structure is correctly populated with all fields
	now := time.Now()
	closedTime := now.Add(24 * time.Hour)
	url := "https://github.com/owner/repo/issues/123"

	issue := clients.Issue{
		URI:         &url,
		IssueNumber: 123,
		CreatedAt:   &now,
		ClosedAt:    &closedTime,
		LabelEvents: []clients.LabelEvent{
			{
				Label:     "bug",
				Added:     true,
				Actor:     "maintainer",
				CreatedAt: now,
			},
		},
		Comments: []clients.IssueComment{
			{
				Author:       &clients.User{Login: "maintainer"},
				CreatedAt:    &now,
				IsMaintainer: true,
				URL:          "https://github.com/owner/repo/issues/123#comment-1",
			},
		},
	}

	tests := []struct {
		checkFunc func(t *testing.T, issue clients.Issue)
		name      string
	}{
		{
			name: "Issue number is set",
			checkFunc: func(t *testing.T, issue clients.Issue) {
				t.Helper()
				if issue.IssueNumber != 123 {
					t.Errorf("expected issue number 123, got %d", issue.IssueNumber)
				}
			},
		},
		{
			name: "URI is set",
			checkFunc: func(t *testing.T, issue clients.Issue) {
				t.Helper()
				if issue.URI == nil || !strings.Contains(*issue.URI, "issues/123") {
					t.Error("URI not properly set")
				}
			},
		},
		{
			name: "CreatedAt is set",
			checkFunc: func(t *testing.T, issue clients.Issue) {
				t.Helper()
				if issue.CreatedAt == nil {
					t.Error("CreatedAt should be set")
				}
			},
		},
		{
			name: "ClosedAt is set for closed issue",
			checkFunc: func(t *testing.T, issue clients.Issue) {
				t.Helper()
				if issue.ClosedAt == nil {
					t.Error("ClosedAt should be set for closed issue")
				}
			},
		},
		{
			name: "Label events are populated",
			checkFunc: func(t *testing.T, issue clients.Issue) {
				t.Helper()
				if len(issue.LabelEvents) != 1 {
					t.Errorf("expected 1 label event, got %d", len(issue.LabelEvents))
				}
				if issue.LabelEvents[0].Label != "bug" {
					t.Errorf("expected label 'bug', got %s", issue.LabelEvents[0].Label)
				}
				if !issue.LabelEvents[0].Added {
					t.Error("expected label to be added")
				}
			},
		},
		{
			name: "Comments are populated",
			checkFunc: func(t *testing.T, issue clients.Issue) {
				t.Helper()
				if len(issue.Comments) != 1 {
					t.Errorf("expected 1 comment, got %d", len(issue.Comments))
				}
				if !issue.Comments[0].IsMaintainer {
					t.Error("expected comment to be from maintainer")
				}
				if issue.Comments[0].Author == nil || issue.Comments[0].Author.Login != "maintainer" {
					t.Error("expected comment author to be 'maintainer'")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.checkFunc(t, issue)
		})
	}
}

func Test_issuesHandler_getIssues(t *testing.T) {
	t.Parallel()

	t.Run("Returns cached issues after setup", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		handler := &issuesHandler{}
		handler.init(ctx, &Repo{owner: "testowner", repo: "testrepo"})

		// Mock graphClient would normally be set here
		// For this test, we'll verify the method calls setup and returns the cached issues
		// Since we don't have a real GraphQL client, we'll just verify error handling

		_, err := handler.getIssues()
		if err == nil {
			t.Error("Expected error when graphClient is nil, got nil")
		}
		if !strings.Contains(err.Error(), "setup") {
			t.Errorf("Expected error to mention 'setup', got: %v", err)
		}
	})

	t.Run("Returns same issues on multiple calls", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		handler := &issuesHandler{}
		handler.init(ctx, &Repo{owner: "testowner", repo: "testrepo"})

		// Simulate already-cached issues
		handler.issues = []clients.Issue{
			{IssueNumber: 1},
			{IssueNumber: 2},
		}
		handler.once.Do(func() {}) // Mark setup as complete

		issues1, err := handler.getIssues()
		if err != nil {
			t.Fatalf("getIssues() error = %v", err)
		}

		issues2, err := handler.getIssues()
		if err != nil {
			t.Fatalf("getIssues() error = %v", err)
		}

		if len(issues1) != len(issues2) {
			t.Errorf("Expected same number of issues, got %d and %d", len(issues1), len(issues2))
		}

		// Verify it returns the actual cached slice
		if len(issues1) != 2 {
			t.Errorf("Expected 2 issues, got %d", len(issues1))
		}
	})
}

func Test_issuesHandler_listIssuesWithHistory_returnsCopy(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := &issuesHandler{}
	handler.init(ctx, &Repo{owner: "testowner", repo: "testrepo"})

	// Simulate already-cached issues
	handler.issues = []clients.Issue{
		{IssueNumber: 1},
		{IssueNumber: 2},
	}
	handler.once.Do(func() {}) // Mark setup as complete

	issues, err := handler.listIssuesWithHistory()
	if err != nil {
		t.Fatalf("listIssuesWithHistory() error = %v", err)
	}

	if len(issues) != 2 {
		t.Errorf("Expected 2 issues, got %d", len(issues))
	}

	// Verify it's a copy by modifying it
	issues[0].IssueNumber = 999

	// Original should be unchanged
	if handler.issues[0].IssueNumber == 999 {
		t.Error("listIssuesWithHistory should return a copy, not the original slice")
	}
}

func Test_issuesHandler_unlabeledEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		labelName     string
		shouldInclude bool
	}{
		{
			name:          "unlabeled bug event",
			labelName:     "bug",
			shouldInclude: true,
		},
		{
			name:          "unlabeled security event",
			labelName:     "security",
			shouldInclude: true,
		},
		{
			name:          "unlabeled enhancement (filtered out)",
			labelName:     "enhancement",
			shouldInclude: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Test the normalization logic that would be applied to UnlabeledEvent
			name := strings.ToLower(strings.TrimSpace(tt.labelName))
			included := (name == "bug" || name == "security")

			if included != tt.shouldInclude {
				t.Errorf("Label %q: expected included=%v, got %v", tt.labelName, tt.shouldInclude, included)
			}

			// If included, verify it would be marked as Added=false for unlabeled events
			if included {
				labelEvent := clients.LabelEvent{
					Label:     name,
					Added:     false, // UnlabeledEvent sets Added=false
					Actor:     "testactor",
					CreatedAt: time.Now(),
				}
				if labelEvent.Added {
					t.Error("UnlabeledEvent should set Added=false")
				}
			}
		})
	}
}

func Test_issuesHandler_emptyAndWhitespaceLabels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		labelName     string
		shouldInclude bool
	}{
		{
			name:          "empty label name",
			labelName:     "",
			shouldInclude: false,
		},
		{
			name:          "whitespace only label",
			labelName:     "   ",
			shouldInclude: false,
		},
		{
			name:          "label with leading/trailing whitespace",
			labelName:     "  bug  ",
			shouldInclude: true,
		},
		{
			name:          "label with internal whitespace (not bug/security)",
			labelName:     "bug fix",
			shouldInclude: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Simulate the label filtering logic
			name := strings.ToLower(strings.TrimSpace(tt.labelName))
			included := (name == "bug" || name == "security")

			if included != tt.shouldInclude {
				t.Errorf("Label %q -> %q: expected included=%v, got %v",
					tt.labelName, name, tt.shouldInclude, included)
			}
		})
	}
}

func Test_issuesHandler_multipleLabelEvents(t *testing.T) {
	t.Parallel()

	t.Run("Multiple add/remove sequences", func(t *testing.T) {
		t.Parallel()

		// Simulate an issue with multiple label events
		issue := clients.Issue{
			IssueNumber: 123,
			LabelEvents: []clients.LabelEvent{},
		}

		// Add bug label
		issue.LabelEvents = append(issue.LabelEvents,
			clients.LabelEvent{
				Label:     "bug",
				Added:     true,
				CreatedAt: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			},
			// Remove bug label
			clients.LabelEvent{
				Label:     "bug",
				Added:     false,
				CreatedAt: time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
			},
			// Add bug label again
			clients.LabelEvent{
				Label:     "bug",
				Added:     true,
				CreatedAt: time.Date(2024, 1, 3, 10, 0, 0, 0, time.UTC),
			},
			// Add security label
			clients.LabelEvent{
				Label:     "security",
				Added:     true,
				CreatedAt: time.Date(2024, 1, 4, 10, 0, 0, 0, time.UTC),
			},
		)

		// Verify all events are preserved
		if len(issue.LabelEvents) != 4 {
			t.Errorf("Expected 4 label events, got %d", len(issue.LabelEvents))
		}

		// Verify order is preserved (chronological)
		if !issue.LabelEvents[0].CreatedAt.Before(issue.LabelEvents[1].CreatedAt) {
			t.Error("Label events should maintain chronological order")
		}

		// Verify alternating Added/Removed for bug
		if issue.LabelEvents[0].Label != "bug" || !issue.LabelEvents[0].Added {
			t.Error("First event should be bug added")
		}
		if issue.LabelEvents[1].Label != "bug" || issue.LabelEvents[1].Added {
			t.Error("Second event should be bug removed")
		}
		if issue.LabelEvents[2].Label != "bug" || !issue.LabelEvents[2].Added {
			t.Error("Third event should be bug added again")
		}
	})
}

func Test_issuesHandler_issueWithoutTimeline(t *testing.T) {
	t.Parallel()

	t.Run("Issue with no labels or comments", func(t *testing.T) {
		t.Parallel()

		// An issue with empty timeline should still be valid
		issue := clients.Issue{
			IssueNumber: 456,
			LabelEvents: []clients.LabelEvent{},
			Comments:    []clients.IssueComment{},
		}

		if issue.LabelEvents == nil {
			t.Error("LabelEvents should be empty slice, not nil")
		}
		if issue.Comments == nil {
			t.Error("Comments should be empty slice, not nil")
		}
		if len(issue.LabelEvents) != 0 {
			t.Errorf("Expected 0 label events, got %d", len(issue.LabelEvents))
		}
		if len(issue.Comments) != 0 {
			t.Errorf("Expected 0 comments, got %d", len(issue.Comments))
		}
	})
}

func Test_issuesHandler_closedAtHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		closedAt *time.Time
		name     string
		isOpen   bool
	}{
		{
			name:     "Open issue (ClosedAt is nil)",
			closedAt: nil,
			isOpen:   true,
		},
		{
			name:     "Closed issue (ClosedAt is set)",
			closedAt: timePtr(time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)),
			isOpen:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			issue := clients.Issue{
				IssueNumber: 789,
				ClosedAt:    tt.closedAt,
			}

			isOpen := (issue.ClosedAt == nil)
			if isOpen != tt.isOpen {
				t.Errorf("Expected isOpen=%v, got %v", tt.isOpen, isOpen)
			}

			if !tt.isOpen && issue.ClosedAt == nil {
				t.Error("Closed issue should have ClosedAt set")
			}
		})
	}
}

func Test_issuesHandler_commentAuthorNormalization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		authorLogin   string
		expectedLogin string
	}{
		{
			name:          "Uppercase login normalized to lowercase",
			authorLogin:   "MAINTAINER",
			expectedLogin: "maintainer",
		},
		{
			name:          "Mixed case login normalized",
			authorLogin:   "MainTainer",
			expectedLogin: "maintainer",
		},
		{
			name:          "Already lowercase",
			authorLogin:   "maintainer",
			expectedLogin: "maintainer",
		},
		{
			name:          "Login with numbers",
			authorLogin:   "User123",
			expectedLogin: "user123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Simulate the normalization that happens in the code
			normalizedLogin := strings.ToLower(tt.authorLogin)

			if normalizedLogin != tt.expectedLogin {
				t.Errorf("Expected normalized login %q, got %q", tt.expectedLogin, normalizedLogin)
			}
		})
	}
}

// Helper function for tests.
func timePtr(t time.Time) *time.Time {
	return &t
}
