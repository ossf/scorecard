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

func TestCreateInterval_CloseReopenLogic(t *testing.T) {
	t.Parallel()
	now := time.Now()
	start := now.AddDate(0, 0, -200)
	end := now

	tests := []struct {
		name                    string
		description             string
		stateChanges            []clients.StateChangeEvent
		comments                []clients.IssueComment
		expectedDurationDays    int
		isOngoing               bool
		expectMaintainerRespond bool
	}{
		{
			name: "Issue closed by creator within 180 days, not reopened - should count as response",
			stateChanges: []clients.StateChangeEvent{
				{
					CreatedAt:    start.AddDate(0, 0, 50),
					Actor:        "issue-creator",
					IsMaintainer: false,
					Closed:       true,
				},
			},
			comments:                []clients.IssueComment{},
			isOngoing:               false,
			expectMaintainerRespond: true,
			expectedDurationDays:    50,
			description:             "Closing within 180 days should count as response",
		},
		{
			name: "Issue closed by maintainer within 180 days, not reopened - should count as response",
			stateChanges: []clients.StateChangeEvent{
				{
					CreatedAt:    start.AddDate(0, 0, 100),
					Actor:        "maintainer",
					IsMaintainer: true,
					Closed:       true,
				},
			},
			comments:                []clients.IssueComment{},
			isOngoing:               false,
			expectMaintainerRespond: true,
			expectedDurationDays:    100,
			description:             "Maintainer closing within 180 days should count",
		},
		{
			name: "Issue closed after 180 days without maintainer comment - should be violation",
			stateChanges: []clients.StateChangeEvent{
				{
					CreatedAt:    start.AddDate(0, 0, 190),
					Actor:        "maintainer",
					IsMaintainer: true,
					Closed:       true,
				},
			},
			comments:                []clients.IssueComment{},
			isOngoing:               false,
			expectMaintainerRespond: false,
			expectedDurationDays:    190,
			description:             "Closing after 180 days without comment should not count",
		},
		{
			name: "Issue closed then reopened - total open time exceeds 180 days without maintainer comment",
			stateChanges: []clients.StateChangeEvent{
				{
					CreatedAt:    start.AddDate(0, 0, 50),
					Actor:        "creator",
					IsMaintainer: false,
					Closed:       true,
				},
				{
					CreatedAt:    start.AddDate(0, 0, 100),
					Actor:        "creator",
					IsMaintainer: false,
					Closed:       false, // Reopened
				},
			},
			comments:                []clients.IssueComment{},
			isOngoing:               true,
			expectMaintainerRespond: false,
			expectedDurationDays:    150, // 50 days open + 100 days open after reopen
			description:             "Reopened issue with >180 days open time should be violation",
		},
		{
			name: "Issue closed then reopened with maintainer comment after reopen",
			stateChanges: []clients.StateChangeEvent{
				{
					CreatedAt:    start.AddDate(0, 0, 50),
					Actor:        "creator",
					IsMaintainer: false,
					Closed:       true,
				},
				{
					CreatedAt:    start.AddDate(0, 0, 100),
					Actor:        "creator",
					IsMaintainer: false,
					Closed:       false, // Reopened
				},
			},
			comments: []clients.IssueComment{
				{
					CreatedAt:    ptrTime(start.AddDate(0, 0, 120)),
					IsMaintainer: true,
				},
			},
			isOngoing:               true,
			expectMaintainerRespond: true,
			expectedDurationDays:    150, // Duration still counts, but has response
			description:             "Maintainer comment after reopen should count as response",
		},
		{
			name: "Issue closed, reopened, closed again - multiple cycles",
			stateChanges: []clients.StateChangeEvent{
				{
					CreatedAt:    start.AddDate(0, 0, 30),
					Actor:        "creator",
					IsMaintainer: false,
					Closed:       true,
				},
				{
					CreatedAt:    start.AddDate(0, 0, 50),
					Actor:        "maintainer",
					IsMaintainer: true,
					Closed:       false, // Reopened
				},
				{
					CreatedAt:    start.AddDate(0, 0, 100),
					Actor:        "maintainer",
					IsMaintainer: true,
					Closed:       true, // Closed again
				},
			},
			comments:                []clients.IssueComment{},
			isOngoing:               false,
			expectMaintainerRespond: true,
			expectedDurationDays:    80, // 30 + 50 days of open time
			description:             "Final close within 180 days should count",
		},
		{
			name:         "Issue never closed, no maintainer comment - ongoing violation",
			stateChanges: []clients.StateChangeEvent{},
			comments: []clients.IssueComment{
				{
					CreatedAt:    ptrTime(start.AddDate(0, 0, 10)),
					IsMaintainer: false,
				},
			},
			isOngoing:               true,
			expectMaintainerRespond: false,
			expectedDurationDays:    200,
			description:             "Open issue without maintainer response should be violation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			issue := &clients.Issue{
				Comments:          tt.comments,
				StateChangeEvents: tt.stateChanges,
			}

			interval := createInterval(start, end, issue, "bug", tt.isOngoing)

			if tt.expectMaintainerRespond {
				if !interval.MaintainerResponded {
					t.Errorf("%s: Expected maintainer response to be true, got false", tt.description)
				}
				if interval.ResponseAt == nil {
					t.Errorf("%s: Expected ResponseAt to be set, got nil", tt.description)
				}
			} else if interval.MaintainerResponded {
				t.Errorf("%s: Expected maintainer response to be false, got true", tt.description)
			} // Check duration calculation (allow some tolerance for calculation differences)
			if interval.DurationDays < tt.expectedDurationDays-1 || interval.DurationDays > tt.expectedDurationDays+1 {
				t.Errorf("%s: Expected duration ~%d days, got %d days",
					tt.description, tt.expectedDurationDays, interval.DurationDays)
			}
		})
	}
}

func TestCreateInterval_ComplexCloseReopenScenarios(t *testing.T) {
	t.Parallel()
	now := time.Now()
	start := now.AddDate(0, 0, -300)
	end := now

	t.Run("Issue open 100 days, closed 50 days, reopened and open 100 days - total 200 days open", func(t *testing.T) {
		t.Parallel()
		issue := &clients.Issue{
			StateChangeEvents: []clients.StateChangeEvent{
				{
					CreatedAt:    start.AddDate(0, 0, 100),
					Actor:        "creator",
					IsMaintainer: false,
					Closed:       true,
				},
				{
					CreatedAt:    start.AddDate(0, 0, 150),
					Actor:        "creator",
					IsMaintainer: false,
					Closed:       false,
				},
			},
			Comments: []clients.IssueComment{},
		}

		interval := createInterval(start, end, issue, "bug", true)

		// Total open time: 100 (first period) + 150 (second period) = 250 days
		expectedDuration := 250
		if interval.DurationDays < expectedDuration-2 || interval.DurationDays > expectedDuration+2 {
			t.Errorf("Expected duration ~%d days, got %d days", expectedDuration, interval.DurationDays)
		}

		if interval.MaintainerResponded {
			t.Error("Expected no maintainer response for issue without maintainer activity")
		}
	})

	t.Run("Issue with maintainer comment before first close should count", func(t *testing.T) {
		t.Parallel()
		issue := &clients.Issue{
			StateChangeEvents: []clients.StateChangeEvent{
				{
					CreatedAt:    start.AddDate(0, 0, 100),
					Actor:        "creator",
					IsMaintainer: false,
					Closed:       true,
				},
				{
					CreatedAt:    start.AddDate(0, 0, 150),
					Actor:        "creator",
					IsMaintainer: false,
					Closed:       false,
				},
			},
			Comments: []clients.IssueComment{
				{
					CreatedAt:    ptrTime(start.AddDate(0, 0, 50)),
					IsMaintainer: true,
				},
			},
		}

		interval := createInterval(start, end, issue, "bug", true)

		if !interval.MaintainerResponded {
			t.Error("Expected maintainer response from comment before close")
		}
		if interval.ResponseAt == nil {
			t.Error("Expected ResponseAt to be set")
		} else {
			expectedTime := start.AddDate(0, 0, 50)
			if !interval.ResponseAt.Equal(expectedTime) {
				t.Errorf("Expected response at %v, got %v", expectedTime, interval.ResponseAt)
			}
		}
	})
}
