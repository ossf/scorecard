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

package azuredevopsrepo

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/audit"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
)

func TestIsArchived(t *testing.T) {
	t.Parallel()
	tests := []struct {
		err        error
		name       string
		isDisabled bool
		want       bool
		wantErr    bool
	}{
		{
			name:       "repository is archived",
			isDisabled: true,
			want:       true,
			wantErr:    false,
		},
		{
			name:       "repository is not archived",
			isDisabled: false,
			want:       false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := &Client{
				repo: &git.GitRepository{
					IsDisabled: &tt.isDisabled,
				},
				repourl: &Repo{id: "test-repo-id"},
			}

			got, err := client.IsArchived()
			if (err != nil) != tt.wantErr {
				t.Errorf("IsArchived() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsArchived() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCreatedAt_AuditError_FallsBackToFirstCommit(t *testing.T) {
	t.Parallel()

	expectedTime := time.Date(2025, time.February, 28, 3, 0, 0, 0, time.UTC)

	// Create an audit handler that always errors (simulates permission denied)
	auditOnce := new(sync.Once)
	auditOnce.Do(func() {}) // mark as already executed
	ah := &auditHandler{
		once:     auditOnce,
		errSetup: errors.New("Access Denied: needs View audit log permission"),
		repourl:  &Repo{project: "test", name: "test"},
	}

	// Create a commits handler with a pre-set firstCommitCreatedAt
	commitsOnce := new(sync.Once)
	commitsOnce.Do(func() {}) // mark as already executed
	ch := &commitsHandler{
		once:                 commitsOnce,
		firstCommitCreatedAt: expectedTime,
		repourl:              &Repo{id: "test-id"},
	}

	client := &Client{
		audit:   ah,
		commits: ch,
	}

	got, err := client.GetCreatedAt()
	if err != nil {
		t.Fatalf("GetCreatedAt() unexpected error: %v", err)
	}
	if !got.Equal(expectedTime) {
		t.Errorf("GetCreatedAt() = %v, want %v", got, expectedTime)
	}
}

func TestGetCreatedAt_AuditSuccess(t *testing.T) {
	t.Parallel()

	expectedTime := time.Date(2025, time.February, 28, 3, 0, 0, 0, time.UTC)

	ah := &auditHandler{
		once: new(sync.Once),
		queryLog: func(ctx context.Context, args audit.QueryLogArgs) (*audit.AuditLogQueryResult, error) {
			hasMore := false
			return &audit.AuditLogQueryResult{
				HasMore:                  &hasMore,
				DecoratedAuditLogEntries: &[]audit.DecoratedAuditLogEntry{},
			}, nil
		},
		createdAt: expectedTime,
		repourl:   &Repo{project: "test", name: "test"},
	}

	client := &Client{
		audit: ah,
	}

	got, err := client.GetCreatedAt()
	if err != nil {
		t.Fatalf("GetCreatedAt() unexpected error: %v", err)
	}
	if !got.Equal(expectedTime) {
		t.Errorf("GetCreatedAt() = %v, want %v", got, expectedTime)
	}
}
