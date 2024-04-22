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
	"io"
	"os"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	scut "github.com/ossf/scorecard/v5/utests"
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
			path: "./raw/testdata/Dockerfile-script-ok",
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

			mockRepo.EXPECT().GetFileReader(gomock.Any()).DoAndReturn(func(fn string) (io.ReadCloser, error) {
				if tt.path == "" {
					return nil, nil
				}
				return os.Open(tt.path)
			}).AnyTimes()

			dl := scut.TestDetailLogger{}
			c := &checker.CheckRequest{
				RepoClient: mockRepo,
				Dlogger:    &dl,
			}

			res := PinningDependencies(c)
			scut.ValidateTestReturn(t, tt.name, &tt.want, &res, &dl)
		})
	}
}
