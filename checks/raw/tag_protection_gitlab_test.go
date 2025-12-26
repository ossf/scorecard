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

package raw

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
)

func TestTagProtectionGitLabDataCollection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		branches      []string
		tags          []clients.TagRef
		protectedTags []*gitlab.ProtectedTag
		wantBranches  int
		wantTags      int
		wantProtected int
		wantErr       bool
	}{
		{
			name:     "collects branches and tags",
			branches: []string{"main", "develop"},
			tags:     []clients.TagRef{{Name: strPtr("v1.0.0")}},
			protectedTags: []*gitlab.ProtectedTag{
				{
					Name: "v*",
					CreateAccessLevels: []*gitlab.TagAccessDescription{
						{AccessLevel: gitlab.MaintainerPermissions},
					},
				},
			},
			wantBranches:  2,
			wantTags:      1,
			wantProtected: 1,
			wantErr:       false,
		},
		{
			name:          "handles no branches",
			branches:      []string{},
			tags:          []clients.TagRef{{Name: strPtr("v1.0.0")}},
			protectedTags: []*gitlab.ProtectedTag{},
			wantBranches:  0,
			wantTags:      1,
			wantProtected: 0,
			wantErr:       false,
		},
		{
			name:          "handles no tags",
			branches:      []string{"main"},
			tags:          []clients.TagRef{},
			protectedTags: []*gitlab.ProtectedTag{},
			wantBranches:  0, // No branches collected when there are no tags/releases
			wantTags:      0,
			wantProtected: 0,
			wantErr:       false,
		},
		{
			name:     "handles multiple protected patterns",
			branches: []string{"main"},
			tags:     []clients.TagRef{{Name: strPtr("v1.0.0")}},
			protectedTags: []*gitlab.ProtectedTag{
				{
					Name: "v*",
					CreateAccessLevels: []*gitlab.TagAccessDescription{
						{AccessLevel: gitlab.MaintainerPermissions},
					},
				},
				{
					Name: "release-*",
					CreateAccessLevels: []*gitlab.TagAccessDescription{
						{AccessLevel: gitlab.OwnerPermissions},
					},
				},
			},
			wantBranches:  1,
			wantTags:      1,
			wantProtected: 2,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt
			t.Parallel()

			// Create mock clients for GitLab
			mockGitLabClient := &mockGitLabClient{
				branches:      tt.branches,
				tags:          tt.tags,
				protectedTags: tt.protectedTags,
			}

			req := &checker.CheckRequest{
				RepoClient: mockGitLabClient,
			}

			result, err := TagProtection(req)

			if (err != nil) != tt.wantErr {
				t.Errorf("TagProtection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(result.GitLabBranches) != tt.wantBranches {
					t.Errorf("GitLabBranches count = %d, want %d",
						len(result.GitLabBranches), tt.wantBranches)
				}
				if len(result.Tags) != tt.wantTags {
					t.Errorf("Tags count = %d, want %d",
						len(result.Tags), tt.wantTags)
				}
				if len(result.GitLabProtectedTags) != tt.wantProtected {
					t.Errorf("GitLabProtectedTags count = %d, want %d",
						len(result.GitLabProtectedTags), tt.wantProtected)
				}
			}
		})
	}
}

// mockGitLabClient is a minimal mock for testing GitLab-specific functionality.
type mockGitLabClient struct {
	err           error
	branches      []string
	tags          []clients.TagRef
	protectedTags []*gitlab.ProtectedTag
}

func (m *mockGitLabClient) ListBranches() ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.branches, nil
}

func (m *mockGitLabClient) GetProtectedTagPatterns() ([]*gitlab.ProtectedTag, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.protectedTags, nil
}

func (m *mockGitLabClient) GetMinimumAccessLevel([]*gitlab.TagAccessDescription) (gitlab.AccessLevelValue, error) {
	if m.err != nil {
		return 0, m.err
	}
	return gitlab.MaintainerPermissions, nil
}

func (m *mockGitLabClient) ListTags() ([]clients.TagRef, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.tags, nil
}

func (m *mockGitLabClient) URI() string {
	return "https://gitlab.com/test/repo"
}

func (m *mockGitLabClient) IsArchived() (bool, error) {
	return false, nil
}

func (m *mockGitLabClient) GetDefaultBranch() (*clients.BranchRef, error) {
	return &clients.BranchRef{Name: strPtr("main")}, nil
}

func (m *mockGitLabClient) GetDefaultBranchName() (string, error) {
	return "main", nil
}

func (m *mockGitLabClient) GetOrgName() (string, error) {
	return "test", nil
}

func (m *mockGitLabClient) GetCreatedAt() (time.Time, error) {
	t, err := time.Parse(time.RFC3339, "2021-01-01T00:00:00Z")
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func (m *mockGitLabClient) ListCommits() ([]clients.Commit, error) {
	return nil, nil
}

func (m *mockGitLabClient) GetBranch(string) (*clients.BranchRef, error) {
	return nil, nil
}

func (m *mockGitLabClient) GetCommitHash(string) (string, error) {
	return "", nil
}

func (m *mockGitLabClient) Search(clients.SearchRequest) (clients.SearchResponse, error) {
	return clients.SearchResponse{}, nil
}

func (m *mockGitLabClient) SearchCommits(clients.SearchCommitsOptions) ([]clients.Commit, error) {
	return nil, nil
}

func (m *mockGitLabClient) Close() error {
	return nil
}

func (m *mockGitLabClient) ListFiles(predicate func(string) (bool, error)) ([]string, error) {
	return nil, nil
}

func (m *mockGitLabClient) GetFileReader(string) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}

func (m *mockGitLabClient) ListMergeRequests() ([]clients.PullRequest, error) {
	return nil, nil
}

func (m *mockGitLabClient) ListReleases() ([]clients.Release, error) {
	if m.err != nil {
		return nil, m.err
	}
	// Return releases based on tags to ensure TagProtection doesn't return early
	releases := make([]clients.Release, 0, len(m.tags))
	for _, tag := range m.tags {
		releases = append(releases, clients.Release{
			TagName: *tag.Name,
		})
	}
	return releases, nil
}

func (m *mockGitLabClient) ListContributors() ([]clients.User, error) {
	return nil, nil
}

func (m *mockGitLabClient) ListSuccessfulWorkflowRuns(string) ([]clients.WorkflowRun, error) {
	return nil, nil
}

func (m *mockGitLabClient) ListCheckRunsForRef(string) ([]clients.CheckRun, error) {
	return nil, nil
}

func (m *mockGitLabClient) ListStatuses(string) ([]clients.Status, error) {
	return nil, nil
}

func (m *mockGitLabClient) ListWebhooks() ([]clients.Webhook, error) {
	return nil, nil
}

func (m *mockGitLabClient) ListProgrammingLanguages() ([]clients.Language, error) {
	return nil, nil
}

func (m *mockGitLabClient) ListLicenses() ([]clients.License, error) {
	return nil, nil
}

func (m *mockGitLabClient) GetCIIBestPracticesBadge() (clients.BadgeLevel, error) {
	return clients.Unknown, nil
}

func (m *mockGitLabClient) GetOrgRepoClient(context.Context) (clients.RepoClient, error) {
	return nil, errors.New("not implemented")
}

func (m *mockGitLabClient) GetTag(tagName string) (*clients.TagRef, error) {
	if m.err != nil {
		return nil, m.err
	}
	// Search for the tag in our list
	for _, tag := range m.tags {
		if tag.Name != nil && *tag.Name == tagName {
			return &tag, nil
		}
	}
	return nil, fmt.Errorf("tag %s not found", tagName)
}

func (m *mockGitLabClient) InitRepo(clients.Repo, string, int) error {
	return nil
}

func (m *mockGitLabClient) ListIssues() ([]clients.Issue, error) {
	return nil, nil
}

func (m *mockGitLabClient) LocalPath() (string, error) {
	return "", errors.New("not implemented")
}

func strPtr(s string) *string {
	return &s
}
