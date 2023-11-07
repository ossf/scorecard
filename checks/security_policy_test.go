// Copyright 2022 OpenSSF Scorecard Authors
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
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v4/checker"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
	scut "github.com/ossf/scorecard/v4/utests"
)

func TestSecurityPolicy(t *testing.T) {
	t.Parallel()
	//nolint:govet
	tests := []struct {
		name    string
		path    string
		files   []string
		wantErr bool
		want    scut.TestReturn
	}{
		{
			name: "security.md",
			path: "./testdata/securitypolicy/10_realworld",
			files: []string{
				"security.md",
			},
			want: scut.TestReturn{
				Score:        10,
				NumberOfInfo: 4,
				NumberOfWarn: 0,
			},
		},
		{
			name: ".github/security.md",
			path: "./testdata/securitypolicy/10_realworldtwo",
			files: []string{
				".github/security.md",
			},
			want: scut.TestReturn{
				Score:        10,
				NumberOfInfo: 4,
				NumberOfWarn: 0,
			},
		},
		{
			name: "docs/security.md",
			path: "./testdata/securitypolicy/04_textAndDisclosureVuls",
			files: []string{
				"docs/security.md",
			},
			want: scut.TestReturn{
				Score:        4,
				NumberOfInfo: 3,
				NumberOfWarn: 1,
			},
		},
		{
			name: "security.rst",
			path: "./testdata/securitypolicy/03_textOnly",
			files: []string{
				"security.rst",
			},
			want: scut.TestReturn{
				Score:        3,
				NumberOfInfo: 2,
				NumberOfWarn: 2,
			},
		},
		{
			name: ".github/security.rst",
			path: "./testdata/securitypolicy/06_urlOnly",
			files: []string{
				".github/security.rst",
			},
			want: scut.TestReturn{
				Score:        6,
				NumberOfInfo: 2,
				NumberOfWarn: 2,
			},
		},
		{
			name: "docs/security.rst",
			path: "./testdata/securitypolicy/06_emailOnly",
			files: []string{
				"docs/security.rst",
			},
			want: scut.TestReturn{
				Score:        6,
				NumberOfInfo: 2,
				NumberOfWarn: 2,
			},
		},
		{
			name: "doc/security.rst",
			path: "./testdata/securitypolicy/06_urlAndEmailOnly",
			files: []string{
				"doc/security.rst",
			},
			want: scut.TestReturn{
				Score:        6,
				NumberOfInfo: 2,
				NumberOfWarn: 2,
			},
		},
		{
			name: "security.adoc",
			path: "./testdata/securitypolicy/09_linkedContentAndText",
			files: []string{
				"security.adoc",
			},
			want: scut.TestReturn{
				Score:        9,
				NumberOfInfo: 3,
				NumberOfWarn: 1,
			},
		},
		{
			name: ".github/security.adoc",
			path: "./testdata/securitypolicy/10_linkedContentAndTextAndDisclosureVuls",
			files: []string{
				".github/security.adoc",
			},
			want: scut.TestReturn{
				Score:        10,
				NumberOfInfo: 4,
				NumberOfWarn: 0,
			},
		},
		{
			name: "docs/security.adoc",
			path: "./testdata/securitypolicy/00_empty",
			files: []string{
				"docs/security.adoc",
			},
			want: scut.TestReturn{
				Score:        0,
				NumberOfInfo: 1,
				NumberOfWarn: 3,
			},
		},
		{
			name: "Pass Case: Case-insensitive testing",
			path: "./testdata/securitypolicy/00_1byte",
			files: []string{
				"dOCs/SeCuRIty.rsT",
			},
			want: scut.TestReturn{
				Score:        0,
				NumberOfInfo: 1,
				NumberOfWarn: 3,
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

			mockRepo.EXPECT().GetFileContent(gomock.Any()).DoAndReturn(func(fn string) ([]byte, error) {
				if tt.path == "" {
					return nil, nil
				}
				content, err := os.ReadFile(tt.path)
				if err != nil {
					return content, fmt.Errorf("%w", err)
				}
				return content, nil
			}).AnyTimes()

			dl := scut.TestDetailLogger{}
			c := &checker.CheckRequest{
				RepoClient: mockRepo,
				Dlogger:    &dl,
			}

			res := SecurityPolicy(c)

			if !scut.ValidateTestReturn(t, tt.name, &tt.want, &res, &dl) {
				t.Errorf("test failed: log message not present: %+v on %+v", tt.want, res)
			}
		})
	}
}
