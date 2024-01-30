// Copyright 2024 OpenSSF Scorecard Authors
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

func TestPinningDependencies(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		path    string
		files   []string
		want    scut.TestReturn
		wantErr bool
	}{
		{
			name: "Dockerfile",
			path: "./testdata/pinneddependencies/Dockerfile-script-ok",
			files: []string{
				"Dockerfile-script-ok",
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
			mockRepo.EXPECT().GetDefaultBranchName().Return("main", nil).AnyTimes()
            mockRepo.EXPECT().URI().Return("github.com/ossf/scorecard").AnyTimes()
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

			res := PinningDependencies(c)

			if !scut.ValidateTestReturn(t, tt.name, &tt.want, &res, &dl) {
				t.Errorf("test failed: log message not present: %+v on %+v", tt.want, res)
			}
		})
	}
}
