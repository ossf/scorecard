// Copyright 2023 OpenSSF Scorecard Authors
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

//nolint:stylecheck
package issueActivityByProjectMember

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/finding"
)

var (
	collab        = clients.RepoAssociationCollaborator
	firstTimeUser = clients.RepoAssociationFirstTimeContributor
)

func fiveIssuesInThreshold() []clients.Issue {
	fiveIssuesInThreshold := make([]clients.Issue, 0)
	for i := 0; i < 5; i++ {
		createdAt := time.Now().AddDate(0 /*years*/, 0 /*months*/, -1*i /*days*/)
		commit := clients.Issue{
			CreatedAt:         &createdAt,
			AuthorAssociation: &collab,
		}
		fiveIssuesInThreshold = append(fiveIssuesInThreshold, commit)
	}
	return fiveIssuesInThreshold
}

func fiveInThresholdByCollabAndFiveByFirstTimeUser() []clients.Issue {
	fiveInThresholdByCollabAndFiveByFirstTimeUser := make([]clients.Issue, 0)
	for i := 0; i < 10; i++ {
		createdAt := time.Now().AddDate(0 /*years*/, 0 /*months*/, -1*i /*days*/)
		commit := clients.Issue{
			CreatedAt: &createdAt,
		}
		if i > 4 {
			commit.AuthorAssociation = &collab
		} else {
			commit.AuthorAssociation = &firstTimeUser
		}
		fiveInThresholdByCollabAndFiveByFirstTimeUser = append(fiveInThresholdByCollabAndFiveByFirstTimeUser, commit)
	}
	return fiveInThresholdByCollabAndFiveByFirstTimeUser
}

func twentyIssuesInThresholdAndTwentyNot() []clients.Issue {
	twentyIssuesInThresholdAndTwentyNot := make([]clients.Issue, 0)
	for i := 70; i < 111; i++ {
		createdAt := time.Now().AddDate(0 /*years*/, 0 /*months*/, -1*i /*days*/)
		commit := clients.Issue{
			CreatedAt:         &createdAt,
			AuthorAssociation: &collab,
		}
		twentyIssuesInThresholdAndTwentyNot = append(twentyIssuesInThresholdAndTwentyNot, commit)
	}
	return twentyIssuesInThresholdAndTwentyNot
}

func Test_Run(t *testing.T) {
	t.Parallel()
	//nolint:govet
	tests := []struct {
		name     string
		values   map[string]int
		raw      *checker.RawResults
		outcomes []finding.Outcome
		err      error
	}{
		{
			name: "Has no issues in threshold",
			raw: &checker.RawResults{
				MaintainedResults: checker.MaintainedData{
					Issues: []clients.Issue{},
				},
			},
			outcomes: []finding.Outcome{finding.OutcomeNegative},
		},
		{
			name: "Has 5 issues in threshold",
			raw: &checker.RawResults{
				MaintainedResults: checker.MaintainedData{
					Issues: fiveIssuesInThreshold(),
				},
			},
			values: map[string]int{
				"numberOfIssuesUpdatedWithinThreshold": 5,
			},
			outcomes: []finding.Outcome{finding.OutcomePositive},
		},
		{
			name: "Has 20 issues in threshold",
			raw: &checker.RawResults{
				MaintainedResults: checker.MaintainedData{
					Issues: twentyIssuesInThresholdAndTwentyNot(),
				},
			},
			values: map[string]int{
				"numberOfIssuesUpdatedWithinThreshold": 20,
			},
			outcomes: []finding.Outcome{finding.OutcomePositive},
		},
		{
			name: "Has 5 issues by collaborator and 5 by first time user",
			raw: &checker.RawResults{
				MaintainedResults: checker.MaintainedData{
					Issues: fiveInThresholdByCollabAndFiveByFirstTimeUser(),
				},
			},
			values: map[string]int{
				"numberOfIssuesUpdatedWithinThreshold": 5,
			},
			outcomes: []finding.Outcome{finding.OutcomePositive},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			findings, s, err := Run(tt.raw)
			if !cmp.Equal(tt.err, err, cmpopts.EquateErrors()) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(tt.err, err, cmpopts.EquateErrors()))
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(Probe, s); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(len(tt.outcomes), len(findings)); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
			for i := range findings {
				outcome := &tt.outcomes[i]
				f := &findings[i]
				if tt.values != nil {
					if diff := cmp.Diff(tt.values, f.Values); diff != "" {
						t.Errorf("mismatch (-want +got):\n%s", diff)
					}
				}
				if diff := cmp.Diff(*outcome, f.Outcome); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func Test_hasActivityByCollaboratorOrHigher(t *testing.T) {
	t.Parallel()
	r := clients.RepoAssociationCollaborator
	twentyDaysAgo := time.Now().AddDate(0 /*years*/, 0 /*months*/, -20 /*days*/)
	type args struct {
		issue     *clients.Issue
		threshold time.Time
	}
	tests := []struct { //nolint:govet
		name string
		args args
		want bool
	}{
		{
			name: "nil issue",
			args: args{
				issue:     nil,
				threshold: time.Now(),
			},
			want: false,
		},
		{
			name: "repo-association collaborator",
			args: args{
				issue: &clients.Issue{
					CreatedAt:         nil,
					AuthorAssociation: &r,
				},
			},
			want: false,
		},
		{
			name: "twentyDaysAgo",
			args: args{
				issue: &clients.Issue{
					CreatedAt:         &twentyDaysAgo,
					AuthorAssociation: &r,
				},
			},
			want: true,
		},
		{
			name: "repo-association collaborator with comment",
			args: args{
				issue: &clients.Issue{
					CreatedAt:         nil,
					AuthorAssociation: &r,
					Comments: []clients.IssueComment{
						{
							CreatedAt:         &twentyDaysAgo,
							AuthorAssociation: &r,
						},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := hasActivityByCollaboratorOrHigher(tt.args.issue, tt.args.threshold); got != tt.want {
				t.Errorf("hasActivityByCollaboratorOrHigher() = %v, want %v", got, tt.want)
			}
		})
	}
}
