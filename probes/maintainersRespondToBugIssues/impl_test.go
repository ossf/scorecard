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

package maintainersRespondToBugIssues

import (
	"testing"
	"time"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
)

func TestRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantOutcomes map[finding.Outcome]int
		name         string
		wantProbe    string
		raw          checker.IssueResponseData
		wantErr      bool
	}{
		{
			name:         "No issues",
			raw:          checker.IssueResponseData{Items: []checker.IssueResponseLag{}},
			wantOutcomes: map[finding.Outcome]int{},
			wantProbe:    Probe,
			wantErr:      false,
		},
		{
			name: "Issue with no bug/security labels",
			raw: checker.IssueResponseData{
				Items: []checker.IssueResponseLag{
					{
						IssueNumber:       1,
						IssueURL:          "https://github.com/owner/repo/issues/1",
						HadLabelIntervals: []checker.LabelInterval{},
					},
				},
			},
			wantOutcomes: map[finding.Outcome]int{
				finding.OutcomeNotApplicable: 1,
			},
			wantProbe: Probe,
			wantErr:   false,
		},
		{
			name: "Issue with bug label, responded quickly",
			raw: checker.IssueResponseData{
				Items: []checker.IssueResponseLag{
					{
						IssueNumber: 1,
						IssueURL:    "https://github.com/owner/repo/issues/1",
						HadLabelIntervals: []checker.LabelInterval{
							{
								Label:               "bug",
								Start:               time.Now().AddDate(0, 0, -10),
								End:                 time.Now().AddDate(0, 0, -5),
								MaintainerResponded: true,
								ResponseAt:          ptrTime(time.Now().AddDate(0, 0, -8)),
								DurationDays:        5,
							},
						},
					},
				},
			},
			wantOutcomes: map[finding.Outcome]int{
				finding.OutcomeTrue: 1,
			},
			wantProbe: Probe,
			wantErr:   false,
		},
		{
			name: "Issue exceeded 180 days without response",
			raw: checker.IssueResponseData{
				Items: []checker.IssueResponseLag{
					{
						IssueNumber: 1,
						IssueURL:    "https://github.com/owner/repo/issues/1",
						HadLabelIntervals: []checker.LabelInterval{
							{
								Label:               "security",
								Start:               time.Now().AddDate(0, 0, -200),
								End:                 time.Now(),
								MaintainerResponded: false,
								ResponseAt:          nil,
								DurationDays:        200,
							},
						},
					},
				},
			},
			wantOutcomes: map[finding.Outcome]int{
				finding.OutcomeFalse: 1,
			},
			wantProbe: Probe,
			wantErr:   false,
		},
		{
			name: "Issue with multiple intervals, one violation",
			raw: checker.IssueResponseData{
				Items: []checker.IssueResponseLag{
					{
						IssueNumber: 1,
						IssueURL:    "https://github.com/owner/repo/issues/1",
						HadLabelIntervals: []checker.LabelInterval{
							{
								Label:               "bug",
								Start:               time.Now().AddDate(0, 0, -10),
								End:                 time.Now().AddDate(0, 0, -5),
								MaintainerResponded: true,
								ResponseAt:          ptrTime(time.Now().AddDate(0, 0, -8)),
								DurationDays:        5,
							},
							{
								Label:               "security",
								Start:               time.Now().AddDate(0, 0, -250),
								End:                 time.Now(),
								MaintainerResponded: false,
								ResponseAt:          nil,
								DurationDays:        250,
							},
						},
					},
				},
			},
			wantOutcomes: map[finding.Outcome]int{
				finding.OutcomeFalse: 1,
			},
			wantProbe: Probe,
			wantErr:   false,
		},
		{
			name: "Multiple issues with mixed outcomes",
			raw: checker.IssueResponseData{
				Items: []checker.IssueResponseLag{
					{
						IssueNumber: 1,
						IssueURL:    "https://github.com/owner/repo/issues/1",
						HadLabelIntervals: []checker.LabelInterval{
							{
								Label:               "bug",
								Start:               time.Now().AddDate(0, 0, -10),
								End:                 time.Now().AddDate(0, 0, -5),
								MaintainerResponded: true,
								ResponseAt:          ptrTime(time.Now().AddDate(0, 0, -8)),
								DurationDays:        5,
							},
						},
					},
					{
						IssueNumber: 2,
						IssueURL:    "https://github.com/owner/repo/issues/2",
						HadLabelIntervals: []checker.LabelInterval{
							{
								Label:               "security",
								Start:               time.Now().AddDate(0, 0, -200),
								End:                 time.Now(),
								MaintainerResponded: false,
								ResponseAt:          nil,
								DurationDays:        200,
							},
						},
					},
					{
						IssueNumber:       3,
						IssueURL:          "https://github.com/owner/repo/issues/3",
						HadLabelIntervals: []checker.LabelInterval{},
					},
				},
			},
			wantOutcomes: map[finding.Outcome]int{
				finding.OutcomeTrue:          1,
				finding.OutcomeFalse:         1,
				finding.OutcomeNotApplicable: 1,
			},
			wantProbe: Probe,
			wantErr:   false,
		},
		{
			name: "Issue with kind/bug label responded quickly",
			raw: checker.IssueResponseData{
				Items: []checker.IssueResponseLag{
					{
						IssueNumber: 1,
						IssueURL:    "https://github.com/owner/repo/issues/1",
						HadLabelIntervals: []checker.LabelInterval{
							{
								Label:               "kind/bug",
								Start:               time.Now().AddDate(0, 0, -10),
								End:                 time.Now().AddDate(0, 0, -5),
								MaintainerResponded: true,
								ResponseAt:          ptrTime(time.Now().AddDate(0, 0, -8)),
								DurationDays:        5,
							},
						},
					},
				},
			},
			wantOutcomes: map[finding.Outcome]int{
				finding.OutcomeTrue: 1,
			},
			wantProbe: Probe,
			wantErr:   false,
		},
		{
			name: "Issue with area/security label exceeded 180 days",
			raw: checker.IssueResponseData{
				Items: []checker.IssueResponseLag{
					{
						IssueNumber: 1,
						IssueURL:    "https://github.com/owner/repo/issues/1",
						HadLabelIntervals: []checker.LabelInterval{
							{
								Label:               "area/security",
								Start:               time.Now().AddDate(0, 0, -190),
								End:                 time.Now(),
								MaintainerResponded: false,
								ResponseAt:          nil,
								DurationDays:        190,
							},
						},
					},
				},
			},
			wantOutcomes: map[finding.Outcome]int{
				finding.OutcomeFalse: 1,
			},
			wantProbe: Probe,
			wantErr:   false,
		},
		{
			name: "Issue with area/product security label exceeded 180 days",
			raw: checker.IssueResponseData{
				Items: []checker.IssueResponseLag{
					{
						IssueNumber: 1,
						IssueURL:    "https://github.com/owner/repo/issues/1",
						HadLabelIntervals: []checker.LabelInterval{
							{
								Label:               "area/product security",
								Start:               time.Now().AddDate(0, 0, -210),
								End:                 time.Now(),
								MaintainerResponded: false,
								ResponseAt:          nil,
								DurationDays:        210,
							},
						},
					},
				},
			},
			wantOutcomes: map[finding.Outcome]int{
				finding.OutcomeFalse: 1,
			},
			wantProbe: Probe,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			raw := &checker.RawResults{
				MaintainedResults: checker.MaintainedData{
					IssueResponseData: tt.raw,
				},
			}
			findings, probeID, err := Run(raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if probeID != tt.wantProbe {
				t.Errorf("Run() probeID = %v, want %v", probeID, tt.wantProbe)
			}

			// Count outcomes
			outcomes := make(map[finding.Outcome]int)
			for _, f := range findings {
				outcomes[f.Outcome]++
			}

			for outcome, count := range tt.wantOutcomes {
				if outcomes[outcome] != count {
					t.Errorf("Run() outcome %v count = %v, want %v", outcome, outcomes[outcome], count)
				}
			}
		})
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
