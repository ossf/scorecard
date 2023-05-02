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
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
)

func TestGitlabPackagingYamlCheck(t *testing.T) {
	t.Parallel()

	//nolint
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

type MoqRepo struct{}

func (b MoqRepo) URI() string {
	return "myurl.com/owner/project"
}
func (b MoqRepo) Host() string                      { return "" }
func (b MoqRepo) String() string                    { return "" }
func (b MoqRepo) IsValid() error                    { return nil }
func (b MoqRepo) Metadata() []string                { return nil }
func (b MoqRepo) AppendMetadata(metadata ...string) {}

type MoqRepoClient struct{}

var filename string = ""

func (b MoqRepoClient) ListFiles(predicate func(string) (bool, error)) ([]string, error) {
	return []string{filename}, nil
}
func (b MoqRepoClient) GetFileContent(filename string) ([]byte, error) {
	content, err := os.ReadFile(filename)

	if err != nil {
		return nil, errors.New("invalid file")
	}
	return content, nil
}

func (b MoqRepoClient) InitRepo(repo clients.Repo, commitSHA string, commitDepth int) error {
	return nil
}
func (b MoqRepoClient) URI() string                                                  { return "" }
func (b MoqRepoClient) IsArchived() (bool, error)                                    { return false, nil }
func (b MoqRepoClient) LocalPath() (string, error)                                   { return "", nil }
func (b MoqRepoClient) GetBranch(branch string) (*clients.BranchRef, error)          { return nil, nil }
func (b MoqRepoClient) GetCreatedAt() (time.Time, error)                             { return time.Now(), nil }
func (b MoqRepoClient) GetDefaultBranchName() (string, error)                        { return "", nil }
func (b MoqRepoClient) GetDefaultBranch() (*clients.BranchRef, error)                { return nil, nil }
func (b MoqRepoClient) GetOrgRepoClient(context.Context) (clients.RepoClient, error) { return nil, nil }
func (b MoqRepoClient) ListCommits() ([]clients.Commit, error)                       { return nil, nil }
func (b MoqRepoClient) ListIssues() ([]clients.Issue, error)                         { return nil, nil }
func (b MoqRepoClient) ListLicenses() ([]clients.License, error)                     { return nil, nil }
func (b MoqRepoClient) ListReleases() ([]clients.Release, error)                     { return nil, nil }
func (b MoqRepoClient) ListContributors() ([]clients.User, error)                    { return nil, nil }
func (b MoqRepoClient) ListSuccessfulWorkflowRuns(filename string) ([]clients.WorkflowRun, error) {
	return nil, nil
}
func (b MoqRepoClient) ListCheckRunsForRef(ref string) ([]clients.CheckRun, error) { return nil, nil }
func (b MoqRepoClient) ListStatuses(ref string) ([]clients.Status, error)          { return nil, nil }
func (b MoqRepoClient) ListWebhooks() ([]clients.Webhook, error)                   { return nil, nil }
func (b MoqRepoClient) ListProgrammingLanguages() ([]clients.Language, error)      { return nil, nil }
func (b MoqRepoClient) Search(request clients.SearchRequest) (clients.SearchResponse, error) {
	return clients.SearchResponse{}, nil
}
func (b MoqRepoClient) SearchCommits(request clients.SearchCommitsOptions) ([]clients.Commit, error) {
	return nil, nil
}
func (b MoqRepoClient) Close() error { return nil }

func TestGitlabPackagingPackager(t *testing.T) {
	t.Parallel()

	repo := MoqRepo{}
	repoClient := MoqRepoClient{}
	client := checker.CheckRequest{
		Repo:       repo,
		RepoClient: repoClient,
	}

	//nolint
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

			filename = tt.filename

			packagingData, _ := Packaging(&client)

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
