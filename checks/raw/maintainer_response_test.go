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

package raw

import (
	"testing"
	"time"

	"github.com/ossf/scorecard/v5/clients"
)

func TestTrackedLabels(t *testing.T) {
	t.Parallel()
	// Test that clients.TrackedIssueLabels contains the expected labels.
	expectedLabels := map[string]bool{
		"bug":                   true,
		"security":              true,
		"kind/bug":              true,
		"area/security":         true,
		"area/product security": true,
	}

	if len(clients.TrackedIssueLabels) != len(expectedLabels) {
		t.Errorf("Expected %d tracked labels, got %d", len(expectedLabels), len(clients.TrackedIssueLabels))
	}

	for _, label := range clients.TrackedIssueLabels {
		if !expectedLabels[label] {
			t.Errorf("Unexpected label in TrackedIssueLabels: %s", label)
		}
	}

	for label := range expectedLabels {
		found := false
		for _, tracked := range clients.TrackedIssueLabels {
			if tracked == label {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected label not found in TrackedIssueLabels: %s", label)
		}
	}
}

func TestProcessIssue_NewLabels(t *testing.T) {
	t.Parallel()
	now := time.Now()

	tests := []struct {
		name          string
		label         string
		expectTracked bool
	}{
		{
			name:          "kind/bug label is tracked",
			label:         "kind/bug",
			expectTracked: true,
		},
		{
			name:          "area/security label is tracked",
			label:         "area/security",
			expectTracked: true,
		},
		{
			name:          "area/product security label is tracked",
			label:         "area/product security",
			expectTracked: true,
		},
		{
			name:          "unrelated label is not tracked",
			label:         "documentation",
			expectTracked: false,
		},
		{
			name:          "enhancement label is not tracked",
			label:         "enhancement",
			expectTracked: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			issue := &clients.Issue{
				IssueNumber: 1,
				CreatedAt:   &now,
				LabelEvents: []clients.LabelEvent{
					{
						Label:     tt.label,
						Added:     true,
						CreatedAt: now.AddDate(0, 0, -10),
					},
				},
			}

			result := processIssue(issue, now)

			// Check if the label should be tracked.
			if !tt.expectTracked {
				if len(result.HadLabelIntervals) > 0 {
					t.Errorf("Expected label %s not to be tracked, but found %d intervals", tt.label, len(result.HadLabelIntervals))
				}
				if result.CurrentlyLabeledBugOrSecurity {
					t.Errorf("Expected CurrentlyLabeledBugOrSecurity to be false for label %s", tt.label)
				}
				return
			}

			// Label should be tracked - verify intervals.
			if len(result.HadLabelIntervals) == 0 {
				t.Errorf("Expected label %s to be tracked, but no intervals found", tt.label)
				return
			}
			if result.HadLabelIntervals[0].Label != tt.label {
				t.Errorf("Expected interval label %s, got %s", tt.label, result.HadLabelIntervals[0].Label)
			}
			if !result.CurrentlyLabeledBugOrSecurity {
				t.Errorf("Expected CurrentlyLabeledBugOrSecurity to be true for label %s", tt.label)
			}
		})
	}
}

func TestProcessIssue_MultipleNewLabels(t *testing.T) {
	t.Parallel()
	now := time.Now()

	issue := &clients.Issue{
		IssueNumber: 1,
		CreatedAt:   &now,
		LabelEvents: []clients.LabelEvent{
			{
				Label:     "kind/bug",
				Added:     true,
				CreatedAt: now.AddDate(0, 0, -200),
			},
			{
				Label:     "area/security",
				Added:     true,
				CreatedAt: now.AddDate(0, 0, -150),
			},
			{
				Label:     "area/product security",
				Added:     true,
				CreatedAt: now.AddDate(0, 0, -100),
			},
		},
	}

	result := processIssue(issue, now)

	// Should have 3 intervals (one for each label).
	if len(result.HadLabelIntervals) != 3 {
		t.Errorf("Expected 3 intervals, got %d", len(result.HadLabelIntervals))
	}

	// Check each label has an interval.
	foundLabels := make(map[string]bool)
	for _, interval := range result.HadLabelIntervals {
		foundLabels[interval.Label] = true
	}

	expectedLabels := []string{"kind/bug", "area/security", "area/product security"}
	for _, label := range expectedLabels {
		if !foundLabels[label] {
			t.Errorf("Expected interval for label %s not found", label)
		}
	}

	// Should be marked as currently labeled.
	if !result.CurrentlyLabeledBugOrSecurity {
		t.Error("Expected CurrentlyLabeledBugOrSecurity to be true")
	}
}

func TestCurrentlyHasLabel_NewLabels(t *testing.T) {
	t.Parallel()
	now := time.Now()

	tests := []struct {
		name     string
		label    string
		events   []clients.LabelEvent
		expected bool
	}{
		{
			name:  "kind/bug added",
			label: "kind/bug",
			events: []clients.LabelEvent{
				{Label: "kind/bug", Added: true, CreatedAt: now},
			},
			expected: true,
		},
		{
			name:  "area/security added then removed",
			label: "area/security",
			events: []clients.LabelEvent{
				{Label: "area/security", Added: true, CreatedAt: now.AddDate(0, 0, -10)},
				{Label: "area/security", Added: false, CreatedAt: now},
			},
			expected: false,
		},
		{
			name:  "area/product security added",
			label: "area/product security",
			events: []clients.LabelEvent{
				{Label: "area/product security", Added: true, CreatedAt: now},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := currentlyHasLabel(tt.events, tt.label)
			if result != tt.expected {
				t.Errorf("Expected currentlyHasLabel to be %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMaintainerCommentFiltering(t *testing.T) {
	t.Parallel()
	now := time.Now()
	start := now.AddDate(0, 0, -10)

	tests := []struct {
		expectedResponseAt *time.Time
		name               string
		comments           []clients.IssueComment
	}{
		{
			name: "Only maintainer comment should count",
			comments: []clients.IssueComment{
				{
					CreatedAt:    ptrTime(now.AddDate(0, 0, -5)),
					IsMaintainer: true,
				},
			},
			expectedResponseAt: ptrTime(now.AddDate(0, 0, -5)),
		},
		{
			name: "Non-maintainer comment should not count",
			comments: []clients.IssueComment{
				{
					CreatedAt:    ptrTime(now.AddDate(0, 0, -5)),
					IsMaintainer: false,
				},
			},
			expectedResponseAt: nil,
		},
		{
			name: "Maintainer comment after non-maintainer should count",
			comments: []clients.IssueComment{
				{
					CreatedAt:    ptrTime(now.AddDate(0, 0, -8)),
					IsMaintainer: false,
				},
				{
					CreatedAt:    ptrTime(now.AddDate(0, 0, -5)),
					IsMaintainer: true,
				},
			},
			expectedResponseAt: ptrTime(now.AddDate(0, 0, -5)),
		},
		{
			name: "Earliest maintainer comment should be returned",
			comments: []clients.IssueComment{
				{
					CreatedAt:    ptrTime(now.AddDate(0, 0, -3)),
					IsMaintainer: true,
				},
				{
					CreatedAt:    ptrTime(now.AddDate(0, 0, -7)),
					IsMaintainer: true,
				},
			},
			expectedResponseAt: ptrTime(now.AddDate(0, 0, -7)),
		},
		{
			name: "Comment before start should not count",
			comments: []clients.IssueComment{
				{
					CreatedAt:    ptrTime(now.AddDate(0, 0, -15)), // Before start
					IsMaintainer: true,
				},
			},
			expectedResponseAt: nil,
		},
		{
			name: "Mixed maintainer and non-maintainer, only maintainer counts",
			comments: []clients.IssueComment{
				{
					CreatedAt:    ptrTime(now.AddDate(0, 0, -9)),
					IsMaintainer: false,
				},
				{
					CreatedAt:    ptrTime(now.AddDate(0, 0, -8)),
					IsMaintainer: false,
				},
				{
					CreatedAt:    ptrTime(now.AddDate(0, 0, -6)),
					IsMaintainer: true,
				},
				{
					CreatedAt:    ptrTime(now.AddDate(0, 0, -4)),
					IsMaintainer: false,
				},
			},
			expectedResponseAt: ptrTime(now.AddDate(0, 0, -6)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := firstMaintainerCommentOnOrAfter(tt.comments, start)

			if tt.expectedResponseAt == nil {
				if result != nil {
					t.Errorf("Expected nil response, got %v", result)
				}
			} else {
				if result == nil {
					t.Errorf("Expected response at %v, got nil", tt.expectedResponseAt)
				} else if !result.Equal(*tt.expectedResponseAt) {
					t.Errorf("Expected response at %v, got %v", tt.expectedResponseAt, result)
				}
			}
		})
	}
}

func TestCreateInterval_MaintainerOnly(t *testing.T) {
	t.Parallel()
	now := time.Now()
	start := now.AddDate(0, 0, -200)
	end := now

	tests := []struct {
		closedAt                *time.Time
		name                    string
		comments                []clients.IssueComment
		isOngoing               bool
		expectMaintainerRespond bool
	}{
		{
			name: "Maintainer comment within interval should count",
			comments: []clients.IssueComment{
				{
					CreatedAt:    ptrTime(now.AddDate(0, 0, -100)),
					IsMaintainer: true,
				},
			},
			isOngoing:               false,
			expectMaintainerRespond: true,
		},
		{
			name: "Non-maintainer comment should NOT count as response",
			comments: []clients.IssueComment{
				{
					CreatedAt:    ptrTime(now.AddDate(0, 0, -100)),
					IsMaintainer: false,
				},
			},
			isOngoing:               true,
			expectMaintainerRespond: false,
		},
		{
			name: "Unlabeling counts as reaction even without maintainer comment",
			comments: []clients.IssueComment{
				{
					CreatedAt:    ptrTime(now.AddDate(0, 0, -100)),
					IsMaintainer: false,
				},
			},
			isOngoing:               false,
			expectMaintainerRespond: true, // End time counts
		},
		{
			name: "Issue closing counts as reaction",
			comments: []clients.IssueComment{
				{
					CreatedAt:    ptrTime(now.AddDate(0, 0, -250)), // Before start
					IsMaintainer: false,
				},
			},
			closedAt:                ptrTime(now.AddDate(0, 0, -50)),
			isOngoing:               false,
			expectMaintainerRespond: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			issue := &clients.Issue{
				Comments: tt.comments,
				ClosedAt: tt.closedAt,
			}

			interval := createInterval(start, end, issue, "bug", tt.isOngoing)

			//nolint:nestif // test validation logic is clear despite nesting
			if tt.expectMaintainerRespond {
				if !interval.MaintainerResponded {
					t.Errorf("Expected maintainer response to be true, got false")
				}
				if interval.ResponseAt == nil {
					t.Error("Expected ResponseAt to be set, got nil")
				}
			} else {
				if interval.MaintainerResponded {
					t.Errorf("Expected maintainer response to be false, got true")
				}
				if interval.ResponseAt != nil {
					t.Errorf("Expected ResponseAt to be nil, got %v", interval.ResponseAt)
				}
			}
		})
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
