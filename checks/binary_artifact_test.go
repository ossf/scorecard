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
	"context"
	"io"
	"os"
	"testing"

	"github.com/golang/mock/gomock"

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
		expected    checker.CheckResult
	}{
		{
			name:        "Jar file",
			inputFolder: "testdata/binaryartifacts/jars",
			err:         nil,
			expected: checker.CheckResult{
				Score: 8,
			},
		},
		{
			name:        "non binary file",
			inputFolder: "testdata/licensedir/withlicense",
			err:         nil,
			expected: checker.CheckResult{
				Score: 10,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// logger := log.NewLogger(log.DebugLevel)

			// TODO: Use gMock instead of Localdir here.
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

			ctx := context.Background()

			dl := scut.TestDetailLogger{}

			req := checker.CheckRequest{
				Ctx:        ctx,
				RepoClient: mockRepoClient,
				Dlogger:    &dl,
			}

			result := BinaryArtifacts(&req)
			if result.Score != tt.expected.Score {
				t.Errorf("BinaryArtifacts: %v, expected %v for tests %v", result.Score, tt.expected.Score, tt.name)
			}

			ctrl.Finish()
		})
	}
}
