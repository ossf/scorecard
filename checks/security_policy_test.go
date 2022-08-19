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
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v4/checker"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
	scut "github.com/ossf/scorecard/v4/utests"
)

func TestSecurityPolicy(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name    string
		files   []string
		wantErr bool
		want    scut.TestReturn
	}{
		{
			name: "security.md",
			files: []string{
				"security.md",
			},
			want: scut.TestReturn{
				Score:        10,
				NumberOfInfo: 1,
			},
		},
		{
			name: ".github/security.md",
			files: []string{
				".github/security.md",
			},
			want: scut.TestReturn{
				Score:        10,
				NumberOfInfo: 1,
			},
		},
		{
			name: "docs/security.md",
			files: []string{
				"docs/security.md",
			},
			want: scut.TestReturn{
				Score:        10,
				NumberOfInfo: 1,
			},
		},
		{
			name: "security.rst",
			files: []string{
				"security.rst",
			},
			want: scut.TestReturn{
				Score:        10,
				NumberOfInfo: 1,
			},
		},
		{
			name: ".github/security.rst",
			files: []string{
				".github/security.rst",
			},
			want: scut.TestReturn{
				Score:        10,
				NumberOfInfo: 1,
			},
		},
		{
			name: "docs/security.rst",
			files: []string{
				"docs/security.rst",
			},
			want: scut.TestReturn{
				Score:        10,
				NumberOfInfo: 1,
			},
		},
		{
			name: "doc/security.rst",
			files: []string{
				"doc/security.rst",
			},
			want: scut.TestReturn{
				Score:        10,
				NumberOfInfo: 1,
			},
		},
		{
			name: "security.adoc",
			files: []string{
				"security.adoc",
			},
			want: scut.TestReturn{
				Score:        10,
				NumberOfInfo: 1,
			},
		},
		{
			name: ".github/security.adoc",
			files: []string{
				".github/security.adoc",
			},
			want: scut.TestReturn{
				Score:        10,
				NumberOfInfo: 1,
			},
		},
		{
			name: "docs/security.adoc",
			files: []string{
				"docs/security.adoc",
			},
			want: scut.TestReturn{
				Score:        10,
				NumberOfInfo: 1,
			},
		},
		{
			name: "Pass Case: Case-insensitive testing",
			files: []string{
				"dOCs/SeCuRIty.rsT",
			},
			want: scut.TestReturn{
				Score:        10,
				NumberOfInfo: 1,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockRepo := mockrepo.NewMockRepoClient(ctrl)

			mockRepo.EXPECT().ListFiles(gomock.Any()).Return(tt.files, nil).AnyTimes()
			dl := scut.TestDetailLogger{}
			c := checker.CheckRequest{
				RepoClient: mockRepo,
				Dlogger:    &dl,
			}

			res := SecurityPolicy(&c)

			if !scut.ValidateTestReturn(t, tt.name, &tt.want, &res, &dl) {
				t.Errorf("test failed: log message not present: %+v", tt.want)
			}
		})
	}
}
