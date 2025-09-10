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
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"

	"github.com/ossf/scorecard/v5/clients"
)

func Test_listIssues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		wantErrStr string
		mockSetup  func(*workItemsHandler)
		want       []clients.Issue
	}{
		{
			name: "happy path",
			mockSetup: func(w *workItemsHandler) {
				workItems := &workitemtracking.WorkItemQueryResult{
					WorkItems: &[]workitemtracking.WorkItemReference{
						{Id: toPtr(1)},
					},
				}
				w.queryWorkItems = func(ctx context.Context, args workitemtracking.QueryByWiqlArgs) (*workitemtracking.WorkItemQueryResult, error) {
					return workItems, nil
				}

				createdDate := "2024-01-01T00:00:00Z"
				workItemDetails := &[]workitemtracking.WorkItem{
					{
						Id:  toPtr(1),
						Url: toPtr("http://example.com"),
						Fields: &map[string]interface{}{
							"System.CreatedDate": createdDate,
							"System.CreatedBy": map[string]interface{}{
								"uniqueName": "test-user",
							},
						},
					},
				}
				w.getWorkItems = func(ctx context.Context, args workitemtracking.GetWorkItemsArgs) (*[]workitemtracking.WorkItem, error) {
					return workItemDetails, nil
				}

				commentTime := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
				comments := &workitemtracking.CommentList{
					Comments: &[]workitemtracking.Comment{
						{CreatedDate: &azuredevops.Time{Time: commentTime}, CreatedBy: &webapi.IdentityRef{UniqueName: toPtr("test-user")}},
					},
				}
				w.getWorkItemComments = func(ctx context.Context, args workitemtracking.GetCommentsArgs) (*workitemtracking.CommentList, error) {
					return comments, nil
				}
			},
			want: []clients.Issue{
				{
					URI:               toPtr("http://example.com"),
					CreatedAt:         toPtr(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
					Author:            &clients.User{Login: "test-user"},
					AuthorAssociation: toPtr(clients.RepoAssociationMember),
					Comments: []clients.IssueComment{
						{
							CreatedAt:         toPtr(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)),
							Author:            &clients.User{Login: "test-user"},
							AuthorAssociation: toPtr(clients.RepoAssociationMember),
						},
					},
				},
			},
		},
		{
			name: "query work items error",
			mockSetup: func(w *workItemsHandler) {
				w.queryWorkItems = func(ctx context.Context, args workitemtracking.QueryByWiqlArgs) (*workitemtracking.WorkItemQueryResult, error) {
					return nil, fmt.Errorf("query error")
				}
			},
			wantErrStr: "error during issuesHandler.setup: error getting work items: query error",
		},
		{
			name: "get work items error",
			mockSetup: func(w *workItemsHandler) {
				workItems := &workitemtracking.WorkItemQueryResult{
					WorkItems: &[]workitemtracking.WorkItemReference{
						{Id: toPtr(1)},
					},
				}
				w.queryWorkItems = func(ctx context.Context, args workitemtracking.QueryByWiqlArgs) (*workitemtracking.WorkItemQueryResult, error) {
					return workItems, nil
				}
				w.getWorkItems = func(ctx context.Context, args workitemtracking.GetWorkItemsArgs) (*[]workitemtracking.WorkItem, error) {
					return nil, fmt.Errorf("get items error")
				}
			},
			wantErrStr: "error during issuesHandler.setup: error getting work item details: get items error",
		},
		{
			name: "get comments error",
			mockSetup: func(w *workItemsHandler) {
				workItems := &workitemtracking.WorkItemQueryResult{
					WorkItems: &[]workitemtracking.WorkItemReference{
						{Id: toPtr(1)},
					},
				}
				w.queryWorkItems = func(ctx context.Context, args workitemtracking.QueryByWiqlArgs) (*workitemtracking.WorkItemQueryResult, error) {
					return workItems, nil
				}
				createdDate := "2024-01-01T00:00:00Z"
				workItemDetails := &[]workitemtracking.WorkItem{
					{
						Id:  toPtr(1),
						Url: toPtr("http://example.com"),
						Fields: &map[string]interface{}{
							"System.CreatedDate": createdDate,
							"System.CreatedBy": map[string]interface{}{
								"uniqueName": "test-user",
							},
						},
					},
				}
				w.getWorkItems = func(ctx context.Context, args workitemtracking.GetWorkItemsArgs) (*[]workitemtracking.WorkItem, error) {
					return workItemDetails, nil
				}
				w.getWorkItemComments = func(ctx context.Context, args workitemtracking.GetCommentsArgs) (*workitemtracking.CommentList, error) {
					return nil, fmt.Errorf("comments error")
				}
			},
			wantErrStr: "error during issuesHandler.setup: error getting comments for work item 1: comments error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := &workItemsHandler{
				ctx:     t.Context(),
				once:    new(sync.Once),
				repourl: &Repo{project: "test-project"},
			}
			tt.mockSetup(w)

			got, err := w.listIssues()
			if tt.wantErrStr != "" {
				if err == nil || err.Error() != tt.wantErrStr {
					t.Errorf("listIssues() error = %v, wantErr %v", err, tt.wantErrStr)
				}
				return
			}
			if err != nil {
				t.Errorf("listIssues() unexpected error: %v", err)
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("listIssues() = %v, want %v", got, tt.want)
			}
		})
	}
}

func toPtr[T any](v T) *T {
	return &v
}
