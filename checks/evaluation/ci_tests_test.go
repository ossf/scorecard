// Copyright 2022 OpenSSF Scorecard Authors
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

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
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
			name: "azure-pipelines",
			args: args{
				s: "azure-pipelines",
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
	tests := []struct {
		name    string
		args    checker.RevisionCIInfo
		want    bool
		wantErr bool
	}{
		{
			name: "check run with conclusion success",
			args: checker.RevisionCIInfo{
				PullRequestNumber: 1,
				HeadSHA:           "sha",
				CheckRuns: []clients.CheckRun{
					{
						App:        clients.CheckRunApp{Slug: "test"},
						Conclusion: "success",
						URL:        "url",
						Status:     "completed",
					},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "check run with conclusion not success",
			args: checker.RevisionCIInfo{
				PullRequestNumber: 1,
				HeadSHA:           "sha",
				CheckRuns: []clients.CheckRun{
					{
						App:        clients.CheckRunApp{Slug: "test"},
						Conclusion: "failed",
						URL:        "url",
						Status:     "completed",
					},
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "check run with conclusion not success",
			args: checker.RevisionCIInfo{
				PullRequestNumber: 1,
				HeadSHA:           "sha",
				CheckRuns: []clients.CheckRun{
					{
						App:        clients.CheckRunApp{Slug: "test"},
						Conclusion: "success",
						URL:        "url",
						Status:     "notcompleted",
					},
				},
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		dl := &scut.TestDetailLogger{}

		//nolint:errcheck
		got, _ := prHasSuccessfulCheck(tt.args, dl)
		if got != tt.want {
			t.Errorf("prHasSuccessfulCheck() = %v, want %v", got, tt.want)
		}
	}
}
