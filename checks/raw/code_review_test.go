// Copyright 2022 Security Scorecard Authors
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
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
)

// Test_reviews tests the reviews function.
func Test_reviews(t *testing.T) {
	t.Parallel()
	type args struct {
		mr *clients.PullRequest
	}
	tests := []struct {
		name string
		args args
		want []checker.Review
	}{
		{
			name: "Test_reviews",
			args: args{
				mr: &clients.PullRequest{
					Reviews: []clients.Review{},
				},
			},
			want: []checker.Review{},
		},
		{
			name: "Test_reviews",
			args: args{
				mr: &clients.PullRequest{
					Reviews: []clients.Review{
						{
							State: "APPROVED",
							Author: &clients.User{
								Login: "user",
							},
						},
						{
							State: "APPROVED",
							Author: &clients.User{
								Login: "user",
							},
						},
					},
				},
			},
			want: []checker.Review{
				{
					Reviewer: checker.User{
						Login: "user",
					},
					State: "APPROVED",
				},
				{
					Reviewer: checker.User{
						Login: "user",
					},
					State: "APPROVED",
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := reviews(tt.args.mr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("reviews() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_labels tests the labels function.
func Test_labels(t *testing.T) {
	t.Parallel()
	type args struct {
		mr *clients.PullRequest
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Test_labels",
			args: args{
				mr: &clients.PullRequest{
					Labels: []clients.Label{},
				},
			},
			want: []string{},
		},
		{
			name: "Test_labels",
			args: args{
				mr: &clients.PullRequest{
					Labels: []clients.Label{
						{
							Name: "label",
						},
						{
							Name: "label",
						},
					},
				},
			},
			want: []string{"label", "label"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := labels(tt.args.mr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("labels() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_mergeRequest tests the mergeRequest function.
func Test_mergeRequest(t *testing.T) {
	t.Parallel()
	type args struct {
		mr *clients.PullRequest
	}
	//nolint
	tests := []struct {
		name string
		args args
		want *checker.MergeRequest
	}{
		{
			name: "Test_mergeRequest",
			args: args{
				mr: &clients.PullRequest{
					MergedAt: time.Time{},
					HeadSHA:  "sha",
					Labels: []clients.Label{
						{
							Name: "label",
						},
						{
							Name: "label",
						},
					},
				},
			},
			want: &checker.MergeRequest{
				MergedAt: time.Time{},
				Labels:   []string{"label", "label"},
				Author:   checker.User{},
				Reviews:  []checker.Review{},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := mergeRequest(tt.args.mr); !cmp.Equal(got, tt.want) {
				t.Errorf("mergeRequest() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// Test_getRawDataFromCommit tests the getRawDataFromCommit function.
func Test_getRawDataFromCommit(t *testing.T) {
	t.Parallel()
	type args struct {
		c *clients.Commit
	}
	tests := []struct {
		name string
		args args
		want checker.DefaultBranchCommit
	}{
		{
			name: "Test_getRawDataFromCommit",
			args: args{
				c: &clients.Commit{
					CommittedDate: time.Time{},
					Message:       "message",
					SHA:           "sha",
				},
			},
			want: checker.DefaultBranchCommit{
				SHA:           "sha",
				CommitMessage: "message",
				CommitDate:    &time.Time{},
				MergeRequest: &checker.MergeRequest{
					Labels:  []string{},
					Reviews: []checker.Review{},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := getRawDataFromCommit(tt.args.c); !cmp.Equal(got, tt.want) {
				t.Errorf(cmp.Diff(got, tt.want))
			}
		})
	}
}

// TestCodeReviews tests the CodeReviews function.
func TestCodeReview(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		want    checker.CodeReviewData
		wantErr bool
	}{
		{
			name:    "Test_CodeReview",
			wantErr: false,
		},
		{
			name:    "Want error",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mr := mockrepo.NewMockRepoClient(ctrl)
			mr.EXPECT().ListCommits().DoAndReturn(func() ([]*clients.Commit, error) {
				if tt.wantErr {
					//nolint
					return nil, errors.New("error")
				}
				return []*clients.Commit{
					{
						SHA: "sha",
					},
				}, nil
			})
			result, err := CodeReview(mr)
			if (err != nil) != tt.wantErr {
				t.Errorf("CodeReview() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if !tt.wantErr && cmp.Equal(result, tt.want) {
				t.Errorf(cmp.Diff(result, tt.want))
			}
		})
	}
}
