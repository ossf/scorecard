// Copyright 2021 OpenSSF Scorecard Authors
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

	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	scut "github.com/ossf/scorecard/v5/utests"
)

func TestBinaryArtifacts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		inputFolder string
		err         error
		expected    scut.TestReturn
	}{
		{
			name:        "Jar file",
			inputFolder: "testdata/binaryartifacts/jars",
			err:         nil,
			expected: scut.TestReturn{
				Score:        8,
				NumberOfInfo: 0,
				NumberOfWarn: 2,
			},
		},
		{
			name:        "non binary file",
			inputFolder: "testdata/licensedir/withlicense",
			err:         nil,
			expected: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 0,
				NumberOfWarn: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)

			mockRepoClient.EXPECT().ListFiles(gomock.Any()).DoAndReturn(func(predicate func(string) (bool, error)) ([]string, error) {
				var files []string
				dirFiles, err := os.ReadDir(tt.inputFolder)
				if err == nil {
					for _, file := range dirFiles {
						files = append(files, file.Name())
					}
					print(files)
				}
				return files, err
			}).AnyTimes()

			mockRepoClient.EXPECT().GetFileReader(gomock.Any()).DoAndReturn(func(file string) (io.ReadCloser, error) {
				return os.Open("./" + tt.inputFolder + "/" + file)
			}).AnyTimes()

			ctx := t.Context()

			dl := scut.TestDetailLogger{}

			req := checker.CheckRequest{
				Ctx:        ctx,
				RepoClient: mockRepoClient,
				Dlogger:    &dl,
			}

			result := BinaryArtifacts(&req)

			scut.ValidateTestReturn(t, tt.name, &tt.expected, &result, &dl)

			ctrl.Finish()
		})
	}
}
