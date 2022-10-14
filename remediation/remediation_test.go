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

package remediation

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"

	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
)

func TestNewMetadata(t *testing.T) {
	t.Parallel()
	tests := []struct {
		want   Metadata
		name   string
		uri    string
		branch string
	}{
		{
			name:   "Basic",
			uri:    "github.com/owner/repo",
			branch: "main",
			want: Metadata{
				BranchName: "main",
				RepoName:   "owner/repo",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockClient := mockrepo.NewMockRepoClient(ctrl)
			mockClient.EXPECT().URI().Return("github.com/owner/repo").AnyTimes()
			mockClient.EXPECT().GetDefaultBranchName().Return("main", nil).AnyTimes()

			got := NewMetadata(mockClient)
			if !cmp.Equal(got, tt.want) {
				t.Errorf("Got diff: %s", cmp.Diff(tt.want, got))
			}
		})
	}
}

//nolint:lll
func TestCreateWorkflowPinningRemediation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		md   Metadata
		want *Remediation
		name string
		path string
		op   string
	}{
		{
			name: "Basic pin",
			md: Metadata{
				BranchName: "main",
				RepoName:   "owner/repo",
			},
			path: "foo",
			op:   "pin",
			want: &Remediation{
				HelpText:     "update your workflow using https://app.stepsecurity.io/secureworkflow/owner/repo/foo/main?enable=pin",
				HelpMarkdown: "update your workflow using [https://app.stepsecurity.io](https://app.stepsecurity.io/secureworkflow/owner/repo/foo/main?enable=pin)",
			},
		},
		{
			name: "Missing pin information",
			md: Metadata{
				RepoName: "owner/repo",
			},
			path: "foo",
			op:   "pin",
			want: nil,
		},
		{
			name: "Basic permissions",
			md: Metadata{
				BranchName: "main",
				RepoName:   "owner/repo",
			},
			path: "foo",
			op:   "permissions",
			want: &Remediation{
				HelpText:     "update your workflow using https://app.stepsecurity.io/secureworkflow/owner/repo/foo/main?enable=permissions",
				HelpMarkdown: "update your workflow using [https://app.stepsecurity.io](https://app.stepsecurity.io/secureworkflow/owner/repo/foo/main?enable=permissions)",
			},
		},
		{
			name: "Missing permissions information",
			md: Metadata{
				BranchName: "main",
			},
			path: "foo",
			op:   "permissions",
			want: nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got *Remediation
			switch tt.op {
			case "pin":
				got = tt.md.CreateWorkflowPinningRemediation(tt.path)
			case "permissions":
				got = tt.md.CreateWorkflowPermissionRemediation(tt.path)
			default:
				t.Errorf("%s: invalid test op: %s", tt.name, tt.op)
				return
			}

			if !cmp.Equal(got, tt.want) {
				t.Errorf("Got diff: %s", cmp.Diff(tt.want, got))
			}
		})
	}
}
