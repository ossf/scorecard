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

package checks

import (
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
	scut "github.com/ossf/scorecard/v4/utests"
)

func Test_isTest(t *testing.T) {
	t.Parallel()
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "appveyor",
			args: args{
				s: "appveyor",
			},
			want: true,
		},
		{
			name: "circleci",
			args: args{
				s: "circleci",
			},
			want: true,
		},
		{
			name: "jenkins",
			args: args{
				s: "jenkins",
			},
			want: true,
		},
		{
			name: "e2e",
			args: args{
				s: "e2e",
			},
			want: true,
		},
		{
			name: "github-actions",
			args: args{
				s: "github-actions",
			},
			want: true,
		},
		{
			name: "mergeable",
			args: args{
				s: "mergeable",
			},
			want: true,
		},
		{
			name: "packit-as-a-service",
			args: args{
				s: "packit-as-a-service",
			},
			want: true,
		},
		{
			name: "semaphoreci",
			args: args{
				s: "semaphoreci",
			},
			want: true,
		},
		{
			name: "test",
			args: args{
				s: "test",
			},
			want: true,
		},
		{
			name: "travis-ci",
			args: args{
				s: "travis-ci",
			},
			want: true,
		},
		{
			name: "non-existing",
			args: args{
				s: "non-existing",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isTest(tt.args.s); got != tt.want {
				t.Errorf("isTest() = %v, want %v for test %v", got, tt.want, tt.name)
			}
		})
	}
}

func Test_prHasSuccessfulCheck(t *testing.T) {
	t.Parallel()
	//enabled nolint because this is a test
	//nolint
	type args struct {
		r   []clients.CheckRun
		pr  clients.PullRequest
		err error
	}
	//enabled nolint because this is a test
	//nolint
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "check run with conclusion success",
			args: args{
				pr: clients.PullRequest{
					HeadSHA: "sha",
					Number:  1,
					Labels: []clients.Label{
						{
							Name: "label",
						},
					},
					Reviews: []clients.Review{
						{
							State: "APPROVED",
						},
					},
					Author: clients.User{
						Login: "author",
					},
				},
				r: []clients.CheckRun{
					{
						App:        clients.CheckRunApp{Slug: "test"},
						Conclusion: "success",
						URL:        "url",
						Status:     "completed",
					},
				},
				err: nil,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "check run with conclusion not success",
			args: args{
				pr: clients.PullRequest{
					HeadSHA: "sha",
					Number:  1,
					Labels: []clients.Label{
						{
							Name: "label",
						},
					},
					Reviews: []clients.Review{
						{
							State: "APPROVED",
						},
					},
					Author: clients.User{
						Login: "author",
					},
				},
				r: []clients.CheckRun{
					{
						App:        clients.CheckRunApp{Slug: "test"},
						Conclusion: "failed",
						URL:        "url",
						Status:     "completed",
					},
				},
				err: nil,
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "check run with conclusion not success",
			args: args{
				pr: clients.PullRequest{
					HeadSHA: "sha",
					Number:  1,
					Labels: []clients.Label{
						{
							Name: "label",
						},
					},
					Reviews: []clients.Review{
						{
							State: "APPROVED",
						},
					},
					Author: clients.User{
						Login: "author",
					},
				},
				r: []clients.CheckRun{
					{
						App:        clients.CheckRunApp{Slug: "test"},
						Conclusion: "success",
						URL:        "url",
						Status:     "notcompleted",
					},
				},
				err: nil,
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "check run with an error",
			args: args{
				err: errors.New("error"),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockRepo := mockrepo.NewMockRepoClient(ctrl)
			mockRepo.EXPECT().ListCheckRunsForRef(gomock.Any()).DoAndReturn(func(sha string) ([]clients.CheckRun, error) {
				if tt.args.err != nil {
					return nil, tt.args.err
				}
				return tt.args.r, tt.args.err
			}).MinTimes(1)

			req := checker.CheckRequest{
				RepoClient: mockRepo,
			}
			req.Dlogger = &scut.TestDetailLogger{}

			//nolint:errcheck
			got, _ := prHasSuccessfulCheck(&tt.args.pr, &req)
			if got != tt.want {
				t.Errorf("prHasSuccessfulCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_prHasSuccessStatus tests that the prHasSuccessStatus function returns the correct value.
func Test_prHasSuccessStatus(t *testing.T) {
	t.Parallel()
	type args struct {
		pr *clients.PullRequest
	}
	//nolint
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
		status  string
	}{
		{
			name: "check run with conclusion success",
			args: args{
				pr: &clients.PullRequest{
					HeadSHA: "sha",
					Number:  1,
					Labels: []clients.Label{
						{
							Name: "label",
						},
					},
					Reviews: []clients.Review{
						{
							State: "success",
						},
					},
				},
			},
			status: "success",
			want:   true,
		},
		{
			name: "check run with conclusion not success",
			args: args{
				pr: &clients.PullRequest{
					HeadSHA: "sha",
					Number:  1,
					Labels: []clients.Label{
						{
							Name: "label",
						},
					},
					Reviews: []clients.Review{
						{
							State: "failure",
						},
					},
				},
			},
			status: "failure",
			want:   false,
		},
		{
			name:    "check run with error",
			wantErr: true,
			args: args{
				pr: &clients.PullRequest{
					HeadSHA: "sha",
					Number:  1,
					Labels: []clients.Label{
						{
							Name: "label",
						},
					},
					Reviews: []clients.Review{
						{
							State: "success",
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockrepoclient := mockrepo.NewMockRepoClient(ctrl)
			mockrepoclient.EXPECT().ListStatuses(gomock.Any()).DoAndReturn(
				func(sha string) ([]clients.Status, error) {
					if tt.wantErr {
						//nolint
						return nil, errors.New("error")
					}
					return []clients.Status{
						{
							State:   tt.status,
							Context: "buildkite",
						},
					}, nil
				}).AnyTimes()
			c := checker.CheckRequest{
				RepoClient: mockrepoclient,
				Dlogger:    &scut.TestDetailLogger{},
			}
			got, err := prHasSuccessStatus(tt.args.pr, &c)
			if (err != nil) != tt.wantErr {
				t.Errorf("prHasSuccessStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("prHasSuccessStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCITests tests that the CITests function returns the correct value.
func TestCITests(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name     string
		want     checker.CheckResult
		status   string
		wantErr  bool
		commit   []clients.Commit
		r        []clients.CheckRun
		expected scut.TestReturn
	}{
		{
			name: "success",
			expected: scut.TestReturn{
				NumberOfDebug: 1,
			},
			commit: []clients.Commit{
				{
					SHA: "sha",
					AssociatedMergeRequest: clients.PullRequest{
						HeadSHA: "sha",
						Number:  1,
						Labels: []clients.Label{
							{
								Name: "label",
							},
						},
						MergedAt: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		},
		{
			name: "commit 0",
			expected: scut.TestReturn{
				Score: -1,
			},
			commit: []clients.Commit{
				{
					SHA: "sha",
					AssociatedMergeRequest: clients.PullRequest{
						HeadSHA: "sha",
						Number:  1,
						Labels: []clients.Label{
							{
								Name: "label",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)

			mockRepoClient.EXPECT().ListCommits().Return(tt.commit, nil)

			mockRepoClient.EXPECT().ListStatuses(gomock.Any()).DoAndReturn(
				func(sha string) ([]clients.Status, error) {
					if tt.wantErr {
						//nolint
						return nil, errors.New("error")
					}
					return []clients.Status{
						{
							State:   tt.status,
							Context: "buildkite",
						},
					}, nil
				}).AnyTimes()

			mockRepoClient.EXPECT().ListCheckRunsForRef(gomock.Any()).DoAndReturn(
				func(sha string) ([]clients.CheckRun, error) {
					return tt.r, nil
				}).AnyTimes()

			dl := scut.TestDetailLogger{}
			c := checker.CheckRequest{
				RepoClient: mockRepoClient,
				Dlogger:    &dl,
			}
			r := CITests(&c)

			if !scut.ValidateTestReturn(t, tt.name, &tt.expected, &r, &dl) {
				t.Fail()
			}
		})
	}
}
