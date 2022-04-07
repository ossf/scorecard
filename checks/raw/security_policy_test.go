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
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v4/checker"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
	scut "github.com/ossf/scorecard/v4/utests"
)

func Test_isSecurityPolicyFilename(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{
			name:     "test1",
			filename: "test1",
			expected: false,
		},
		{
			name:     "docs/security.rst",
			filename: "docs/security.rst",
			expected: true,
		},
		{
			name:     "doc/security.rst",
			filename: "doc/security.rst",
			expected: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isSecurityPolicyFilename(tt.filename); got != tt.expected {
				t.Errorf("isSecurityPolicyFilename() = %v, want %v for %v", got, tt.expected, tt.name)
			}
		})
	}
}

// TestSecurityPolicy tests the security policy.
func TestSecurityPolicy(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name    string
		files   []string
		result  checker.SecurityPolicyData
		wantErr bool
		want    scut.TestReturn
	}{
		{
			name: "security.md",
			files: []string{
				"security.md",
			},
		},
		{
			name: ".github/security.md",
			files: []string{
				".github/security.md",
			},
		},
		{
			name: "docs/security.md",
			files: []string{
				"docs/security.md",
			},
		},
		{
			name: "docs/security.rst",
			files: []string{
				"docs/security.rst",
			},
		},
		{
			name: "doc/security.rst",
			files: []string{
				"doc/security.rst",
			},
		},
		{
			name:  "early return if org is null",
			files: []string{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			mockRepo := mockrepo.NewMockRepo(ctrl)

			mockRepoClient.EXPECT().ListFiles(gomock.Any()).Return(tt.files, nil).AnyTimes()
			mockRepo.EXPECT().Org().Return(nil).AnyTimes()
			dl := scut.TestDetailLogger{}
			c := checker.CheckRequest{
				RepoClient: mockRepoClient,
				Repo:       mockRepo,
				Dlogger:    &dl,
			}

			res, err := SecurityPolicy(&c)

			if !scut.ValidateTestReturn(t, tt.name, &tt.want, &checker.CheckResult{}, &dl) {
				t.Errorf("test failed: log message not present: %+v , for test %v", tt.want, tt.name)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("SecurityPolicy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(res.Files) != len(tt.files) {
				t.Errorf("test failed: number of files returned is not correct: %+v", res)
			}
		})
	}
}
