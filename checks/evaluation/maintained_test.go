// Copyright 2023 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package evaluation

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	scut "github.com/ossf/scorecard/v4/utests"
)

func Test_hasActivityByCollaboratorOrHigher(t *testing.T) {
	t.Parallel()
	r := clients.RepoAssociationCollaborator
	twentDaysAgo := time.Now().AddDate(0 /*years*/, 0 /*months*/, -20 /*days*/)
	type args struct {
		issue     *clients.Issue
		threshold time.Time
	}
	tests := []struct {
		args args
		name string
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
					CreatedAt:         &twentDaysAgo,
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
							CreatedAt:         &twentDaysAgo,
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

func TestMaintained(t *testing.T) {
	twentyDaysAgo := time.Now().AddDate(0 /*years*/, 0 /*months*/, -20 /*days*/)
	collab := clients.RepoAssociationCollaborator
	t.Parallel()
	type args struct {
		dl   checker.DetailLogger
		r    *checker.MaintainedData
		name string
	}
	tests := []struct {
		name string
		args args
		want checker.CheckResult
	}{
		{
			name: "nil",
			args: args{
				name: "test",
				dl:   nil,
				r:    nil,
			},
			want: checker.CheckResult{
				Name:    "test",
				Version: 2,
				Reason:  "internal error: empty raw data",
				Score:   -1,
			},
		},
		{
			name: "archived",
			args: args{
				name: "test",
				dl:   nil,
				r: &checker.MaintainedData{
					ArchivedStatus: checker.ArchivedStatus{Status: true},
				},
			},
			want: checker.CheckResult{
				Name:    "test",
				Version: 2,
				Reason:  "repo is marked as archived",
				Score:   0,
			},
		},
		{
			name: "no activity",
			args: args{
				name: "test",
				dl:   nil,
				r: &checker.MaintainedData{
					ArchivedStatus: checker.ArchivedStatus{Status: false},
					DefaultBranchCommits: []clients.Commit{
						{
							CommittedDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
						},
					},
				},
			},
			want: checker.CheckResult{
				Name:    "test",
				Version: 2,
				Reason:  "0 commit(s) out of 1 and 0 issue activity out of 0 found in the last 90 days -- score normalized to 0",
				Score:   0,
			},
		},
		{
			name: "commit activity in the last 30 days",
			args: args{
				name: "test",
				dl:   &scut.TestDetailLogger{},
				r: &checker.MaintainedData{
					ArchivedStatus: checker.ArchivedStatus{Status: false},
					DefaultBranchCommits: []clients.Commit{
						{
							CommittedDate: time.Now().AddDate(0 /*years*/, 0 /*months*/, -20 /*days*/),
						},
						{
							CommittedDate: time.Now().AddDate(0 /*years*/, 0 /*months*/, -10 /*days*/),
						},
					},

					Issues: []clients.Issue{
						{
							CreatedAt:         &twentyDaysAgo,
							AuthorAssociation: &collab,
						},
					},
					CreatedAt: time.Now().AddDate(0 /*years*/, 0 /*months*/, -100 /*days*/),
				},
			},
			want: checker.CheckResult{
				Name:    "test",
				Version: 2,
				Reason:  "2 commit(s) out of 2 and 1 issue activity out of 1 found in the last 90 days -- score normalized to 2",
				Score:   2,
			},
		},
		{
			name: "Repo created recently",
			args: args{
				name: "test",
				dl:   &scut.TestDetailLogger{},
				r: &checker.MaintainedData{
					ArchivedStatus: checker.ArchivedStatus{Status: false},
					DefaultBranchCommits: []clients.Commit{
						{
							CommittedDate: time.Now().AddDate(0 /*years*/, 0 /*months*/, -20 /*days*/),
						},
						{
							CommittedDate: time.Now().AddDate(0 /*years*/, 0 /*months*/, -10 /*days*/),
						},
					},

					Issues: []clients.Issue{
						{
							CreatedAt:         &twentyDaysAgo,
							AuthorAssociation: &collab,
						},
					},
					CreatedAt: time.Now().AddDate(0 /*years*/, 0 /*months*/, -10 /*days*/),
				},
			},
			want: checker.CheckResult{
				Name:    "test",
				Version: 2,
				Reason:  "repo was created 10 days ago, not enough maintenance history",
				Score:   0,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := Maintained(tt.args.name, tt.args.dl, tt.args.r); !cmp.Equal(got, tt.want, cmpopts.IgnoreFields(checker.CheckResult{}, "Error")) { //nolint:lll
				t.Errorf("Maintained() = %v, want %v", got, cmp.Diff(got, tt.want, cmpopts.IgnoreFields(checker.CheckResult{}, "Error"))) //nolint:lll
			}
		})
	}
}
