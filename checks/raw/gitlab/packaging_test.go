// Copyright 2020 OpenSSF Scorecard Authors
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

package gitlab

import (
	"os"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v4/checker"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
)

func TestGitlabPackagingYamlCheck(t *testing.T) {
	t.Parallel()

	//nolint:govet
	tests := []struct {
		name       string
		lineNumber uint
		filename   string
		exists     bool
	}{
		{
			name:       "No Publishing Detected",
			filename:   "./testdata/no-publishing.yaml",
			lineNumber: 1,
			exists:     false,
		},
		{
			name:       "Docker",
			filename:   "./testdata/docker.yaml",
			lineNumber: 31,
			exists:     true,
		},
		{
			name:       "Nuget",
			filename:   "./testdata/nuget.yaml",
			lineNumber: 21,
			exists:     true,
		},
		{
			name:       "Poetry",
			filename:   "./testdata/poetry.yaml",
			lineNumber: 30,
			exists:     true,
		},
		{
			name:       "Twine",
			filename:   "./testdata/twine.yaml",
			lineNumber: 26,
			exists:     true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var content []byte
			var err error

			content, err = os.ReadFile(tt.filename)
			if err != nil {
				t.Errorf("cannot read file: %v", err)
			}

			file, found := isGitlabPackagingWorkflow(content, tt.filename)

			if tt.exists && !found {
				t.Errorf("Packaging %q should exist", tt.name)
			} else if !tt.exists && found {
				t.Errorf("No packaging information should have been found in %q", tt.name)
			}

			if file.Offset != tt.lineNumber {
				t.Errorf("Expected line number: %d != %d", tt.lineNumber, file.Offset)
			}

			if err != nil {
				return
			}
		})
	}
}

func TestGitlabPackagingPackager(t *testing.T) {
	t.Parallel()

	//nolint:govet
	tests := []struct {
		name       string
		lineNumber uint
		filename   string
		exists     bool
	}{
		{
			name:       "No Publishing Detected",
			filename:   "./testdata/no-publishing.yaml",
			lineNumber: 1,
			exists:     false,
		},
		{
			name:       "Docker",
			filename:   "./testdata/docker.yaml",
			lineNumber: 31,
			exists:     true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			moqRepoClient := mockrepo.NewMockRepoClient(ctrl)
			moqRepo := mockrepo.NewMockRepo(ctrl)

			moqRepoClient.EXPECT().ListFiles(gomock.Any()).
				Return([]string{tt.filename}, nil).AnyTimes()

			moqRepoClient.EXPECT().GetFileContent(tt.filename).
				DoAndReturn(func(b string) ([]byte, error) {
					//nolint: errcheck
					content, _ := os.ReadFile(b)
					return content, nil
				}).AnyTimes()

			if tt.exists {
				moqRepo.EXPECT().URI().Return("myurl.com/owner/project")
			}

			req := checker.CheckRequest{
				RepoClient: moqRepoClient,
				Repo:       moqRepo,
			}

			//nolint: errcheck
			packagingData, _ := Packaging(&req)

			if !tt.exists {
				if len(packagingData.Packages) != 0 {
					t.Errorf("Repo should not contain any packages")
				}
				return
			}

			if len(packagingData.Packages) == 0 {
				t.Fatalf("Repo should contain related packages")
			}

			pkg := packagingData.Packages[0].File

			if pkg.Offset != tt.lineNumber {
				t.Errorf("Expected line number: %d != %d", tt.lineNumber, pkg.Offset)
			}
			if pkg.Path != tt.filename {
				t.Errorf("Expected filename: %v != %v", tt.filename, pkg.Path)
			}

			runs := packagingData.Packages[0].Runs

			if len(runs) != 1 {
				t.Errorf("Expected only a single run count, but received %d", len(runs))
			}

			if runs[0].URL != "myurl.com/owner/project" {
				t.Errorf("URL did not match expected value %q", runs[0].URL)
			}
		})
	}
}
