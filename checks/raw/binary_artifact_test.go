// Copyright 2021 Security Scorecard Authors
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
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"

	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
)

// TestBinaryArtifact tests the BinaryArtifact checker.
func TestBinaryArtifacts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		err    error
		files  []string
		expect int
	}{
		{
			name: "Jar file",
			err:  nil,
			files: []string{
				"../testdata/binaryartifacts/jars/aws-java-sdk-core-1.11.571.jar",
			},
			expect: 1,
		},
		{
			name: "Mach-O ARM64 executable",
			err:  nil,
			files: []string{
				"../testdata/binaryartifacts/executables/darwin-arm64-bt",
			},
			expect: 1,
		},
		{
			name: "non binary file",
			err:  nil,
			files: []string{
				"../testdata/licensedir/withlicense/LICENSE",
			},
		},
		{
			name: "non binary file",
			err:  nil,
			files: []string{
				"../doesnotexist",
			},
		},
		{
			name: "printable character .lib",
			err:  nil,
			files: []string{
				"../testdata/binaryartifacts/printable.lib",
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			mockRepoClient.EXPECT().ListFiles(gomock.Any()).Return(tt.files, nil)
			mockRepoClient.EXPECT().GetFileContent(gomock.Any()).DoAndReturn(func(file string) ([]byte, error) {
				// This will read the file and return the content
				content, err := os.ReadFile(file)
				if err != nil {
					return content, fmt.Errorf("%w", err)
				}
				return content, nil
			})

			f, err := BinaryArtifacts(mockRepoClient)

			if tt.err != nil {
				// If we expect an error, make sure it is the same
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				// Check that the expected number of files are returned
				if len(f.Files) != tt.expect {
					t.Errorf("expected %d files, got %d test %v", tt.expect, len(f.Files), tt.name)
				}
			}
		})
	}
}
