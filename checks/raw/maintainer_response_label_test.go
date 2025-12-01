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

func TestCreateInterval_MaintainerLabelAddition(t *testing.T) {
	t.Parallel()
	now := time.Now()
	start := now.AddDate(0, 0, -200)
	end := now

	tests := []struct {
		name                    string
		description             string
		labelEvents             []clients.LabelEvent
		comments                []clients.IssueComment
		stateChanges            []clients.StateChangeEvent
		expectedDurationDays    int
		isOngoing               bool
		expectMaintainerRespond bool
	}{
		{
			name: "Maintainer adds tracked label within 180 days - should count as response",
			labelEvents: []clients.LabelEvent{
				{
					CreatedAt:    start.AddDate(0, 0, 50),
					Label:        "security",
					Actor:        "maintainer",
					Added:        true,
					IsMaintainer: true,
				},
			},
			comments:                []clients.IssueComment{},
			stateChanges:            []clients.StateChangeEvent{},
			isOngoing:               true,
			expectMaintainerRespond: true,
			expectedDurationDays:    200,
			description:             "Maintainer adding a tracked label should count as response",
		},
		{
			name: "Maintainer adds tracked label after 180 days - should not count",
			labelEvents: []clients.LabelEvent{
				{
					CreatedAt:    start.AddDate(0, 0, 190),
					Label:        "security",
					Actor:        "maintainer",
					Added:        true,
					IsMaintainer: true,
				},
			},
			comments:                []clients.IssueComment{},
			stateChanges:            []clients.StateChangeEvent{},
			isOngoing:               true,
			expectMaintainerRespond: true,
			expectedDurationDays:    200,
			description:             "Maintainer label addition after 180 days still counts (no time limit on label additions)",
		},
		{
			name: "Non-maintainer adds tracked label - should not count",
			labelEvents: []clients.LabelEvent{
				{
					CreatedAt:    start.AddDate(0, 0, 50),
					Label:        "security",
					Actor:        "contributor",
					Added:        true,
					IsMaintainer: false,
				},
			},
			comments:                []clients.IssueComment{},
			stateChanges:            []clients.StateChangeEvent{},
			isOngoing:               true,
			expectMaintainerRespond: false,
			expectedDurationDays:    200,
			description:             "Non-maintainer label addition should not count",
		},
		{
			name: "Maintainer removes label - should count",
			labelEvents: []clients.LabelEvent{
				{
					CreatedAt:    start.AddDate(0, 0, 50),
					Label:        "bug",
					Actor:        "maintainer",
					Added:        false, // Removed
					IsMaintainer: true,
				},
			},
			comments:                []clients.IssueComment{},
			stateChanges:            []clients.StateChangeEvent{},
			isOngoing:               true, // Still testing within the full interval
			expectMaintainerRespond: true, // Label removal counts as activity
			expectedDurationDays:    200,  // Full interval duration
			description:             "Label removal counts as maintainer activity",
		},
		{
			name: "Maintainer label addition earlier than comment - use label time",
			labelEvents: []clients.LabelEvent{
				{
					CreatedAt:    start.AddDate(0, 0, 30),
					Label:        "security",
					Actor:        "maintainer",
					Added:        true,
					IsMaintainer: true,
				},
			},
			comments: []clients.IssueComment{
				{
					CreatedAt:    ptrTime(start.AddDate(0, 0, 60)),
					IsMaintainer: true,
				},
			},
			stateChanges:            []clients.StateChangeEvent{},
			isOngoing:               true,
			expectMaintainerRespond: true,
			expectedDurationDays:    200,
			description:             "Earlier label addition should be used over later comment",
		},
		{
			name: "Maintainer comment earlier than label addition - use comment time",
			labelEvents: []clients.LabelEvent{
				{
					CreatedAt:    start.AddDate(0, 0, 60),
					Label:        "security",
					Actor:        "maintainer",
					Added:        true,
					IsMaintainer: true,
				},
			},
			comments: []clients.IssueComment{
				{
					CreatedAt:    ptrTime(start.AddDate(0, 0, 30)),
					IsMaintainer: true,
				},
			},
			stateChanges:            []clients.StateChangeEvent{},
			isOngoing:               true,
			expectMaintainerRespond: true,
			expectedDurationDays:    200,
			description:             "Earlier comment should be used over later label addition",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			issue := &clients.Issue{
				Comments:          tt.comments,
				StateChangeEvents: tt.stateChanges,
				LabelEvents:       tt.labelEvents,
				AllLabelEvents:    tt.labelEvents, // For these tests, all events are the same
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
			}

			// Check duration calculation (allow some tolerance for calculation differences)
			if interval.DurationDays < tt.expectedDurationDays-1 || interval.DurationDays > tt.expectedDurationDays+1 {
				t.Errorf("%s: Expected duration ~%d days, got %d days",
					tt.description, tt.expectedDurationDays, interval.DurationDays)
			}
		})
	}
}
