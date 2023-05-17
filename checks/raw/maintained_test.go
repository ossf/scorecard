// Copyright 2023 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package raw

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
)

func TestMaintained(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepoClient := mockrepo.NewMockRepoClient(ctrl)

	ctx := context.Background()
	req := &checker.CheckRequest{
		Ctx:        ctx,
		RepoClient: mockRepoClient,
	}

	t.Run("successfully returns maintained data", func(t *testing.T) {
		createdAt := time.Now().AddDate(-1, 0, 0) // 1 year ago
		archived := false
		commits := []clients.Commit{
			{SHA: "commit1"},
			{SHA: "commit2"},
		}
		issue := "issue1"
		issues := []clients.Issue{
			{URI: &issue},
		}

		mockRepoClient.EXPECT().IsArchived().Return(archived, nil)
		mockRepoClient.EXPECT().ListCommits().Return(commits, nil)
		mockRepoClient.EXPECT().ListIssues().Return(issues, nil)
		mockRepoClient.EXPECT().GetCreatedAt().Return(createdAt, nil)

		data, err := Maintained(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if data.CreatedAt != createdAt {
			t.Errorf("unexpected createdAt: got %v, want %v", data.CreatedAt, createdAt)
		}

		if data.ArchivedStatus.Status != archived {
			t.Errorf("unexpected archived status: got %v, want %v", data.ArchivedStatus.Status, archived)
		}

		if len(data.DefaultBranchCommits) != len(commits) {
			t.Errorf("unexpected number of commits: got %v, want %v", len(data.DefaultBranchCommits), len(commits))
		}

		if len(data.Issues) != len(issues) {
			t.Errorf("unexpected number of issues: got %v, want %v", len(data.Issues), len(issues))
		}
	})

	t.Run("returns error if IsArchived fails", func(t *testing.T) {
		mockRepoClient.EXPECT().IsArchived().Return(false, fmt.Errorf("some error")) // nolint: goerr113

		_, err := Maintained(req)
		if err == nil {
			t.Fatal("expected an error but got none")
		}
	})

	t.Run("returns error if ListCommits fails", func(t *testing.T) {
		mockRepoClient.EXPECT().IsArchived().Return(false, nil)
		mockRepoClient.EXPECT().ListCommits().Return(nil, fmt.Errorf("some error")) // nolint: goerr113

		_, err := Maintained(req)
		if err == nil {
			t.Fatal("expected an error but got none")
		}
	})

	t.Run("returns error if ListIssues fails", func(t *testing.T) {
		mockRepoClient.EXPECT().IsArchived().Return(false, nil)
		mockRepoClient.EXPECT().ListCommits().Return([]clients.Commit{}, nil)
		mockRepoClient.EXPECT().ListIssues().Return(nil, fmt.Errorf("some error")) // nolint: goerr113

		_, err := Maintained(req)
		if err == nil {
			t.Fatal("expected an error but got none")
		}
	})

	t.Run("returns error if GetCreatedAt fails", func(t *testing.T) {
		mockRepoClient.EXPECT().IsArchived().Return(false, nil)
		mockRepoClient.EXPECT().ListCommits().Return([]clients.Commit{}, nil)
		mockRepoClient.EXPECT().ListIssues().Return([]clients.Issue{}, nil)
		mockRepoClient.EXPECT().GetCreatedAt().Return(time.Time{}, fmt.Errorf("some error")) // nolint: goerr113

		_, err := Maintained(req)
		if err == nil {
			t.Fatal("expected an error but got none")
		}
	})
}
