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

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/audit"
)

func Test_auditHandler_setup(t *testing.T) {
	t.Parallel()
	tests := []struct {
		queryLog  fnQueryLog
		createdAt time.Time
		name      string
		wantErr   bool
	}{
		{
			name: "successful setup",
			queryLog: func(ctx context.Context, args audit.QueryLogArgs) (*audit.AuditLogQueryResult, error) {
				return &audit.AuditLogQueryResult{
					HasMore:           new(bool),
					ContinuationToken: new(string),
					DecoratedAuditLogEntries: &[]audit.DecoratedAuditLogEntry{
						{
							ActionId:    strptr("Git.CreateRepo"),
							ProjectName: strptr("test-project"),
							Data:        &map[string]interface{}{"RepoName": "test-repo"},
							Timestamp:   &azuredevops.Time{Time: time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)},
						},
					},
				}, nil
			},
			wantErr:   false,
			createdAt: time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "query log error",
			queryLog: func(ctx context.Context, args audit.QueryLogArgs) (*audit.AuditLogQueryResult, error) {
				return nil, errors.New("query log error")
			},
			wantErr:   true,
			createdAt: time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			handler := &auditHandler{
				once:     new(sync.Once),
				queryLog: tt.queryLog,
				repourl: &Repo{
					project: "test-project",
					name:    "test-repo",
				},
			}
			err := handler.setup()
			if (err != nil) != tt.wantErr {
				t.Fatalf("setup() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !handler.createdAt.Equal(tt.createdAt) {
				t.Errorf("setup() createdAt = %v, want %v", handler.createdAt, tt.createdAt)
			}
		})
	}
}

func strptr(s string) *string {
	return &s
}
