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
	"time"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"

	"github.com/ossf/scorecard/v5/clients"
)

var (
	errSystemCreatedByFieldNotMap        = fmt.Errorf("error: System.CreatedBy field is not a map")
	errSystemCreatedByFieldNotUniqueName = fmt.Errorf("error: System.CreatedBy field does not contain a UniqueName")
	errSystemCreatedDateFieldNotString   = fmt.Errorf("error: System.CreatedDate field is not a string")
)

type (
	fnQueryWorkItems func(
		ctx context.Context,
		args workitemtracking.QueryByWiqlArgs,
	) (*workitemtracking.WorkItemQueryResult, error)

	fnGetWorkItems func(
		ctx context.Context,
		args workitemtracking.GetWorkItemsArgs,
	) (*[]workitemtracking.WorkItem, error)

	fnGetWorkItemComments func(
		ctx context.Context,
		args workitemtracking.GetCommentsArgs,
	) (*workitemtracking.CommentList, error)
)

type workItemsHandler struct {
	ctx                 context.Context
	repourl             *Repo
	once                *sync.Once
	errSetup            error
	workItemsClient     workitemtracking.Client
	queryWorkItems      fnQueryWorkItems
	getWorkItems        fnGetWorkItems
	getWorkItemComments fnGetWorkItemComments
	issues              []clients.Issue
}

func (w *workItemsHandler) init(ctx context.Context, repourl *Repo) {
	w.ctx = ctx
	w.errSetup = nil
	w.once = new(sync.Once)
	w.repourl = repourl
	w.queryWorkItems = w.workItemsClient.QueryByWiql
	w.getWorkItems = w.workItemsClient.GetWorkItems
	w.getWorkItemComments = w.workItemsClient.GetComments
}

func (w *workItemsHandler) setup() error {
	w.once.Do(func() {
		wiql := `
			SELECT [System.Id]
			FROM WorkItems
			WHERE [System.TeamProject] = @project
			ORDER BY [System.Id] DESC
		`
		workItems, err := w.queryWorkItems(w.ctx, workitemtracking.QueryByWiqlArgs{
			Project: &w.repourl.project,
			Wiql: &workitemtracking.Wiql{
				Query: &wiql,
			},
		})
		if err != nil {
			w.errSetup = fmt.Errorf("error getting work items: %w", err)
			return
		}

		ids := make([]int, 0, len(*workItems.WorkItems))
		for _, wi := range *workItems.WorkItems {
			ids = append(ids, *wi.Id)
		}

		// Get details for each work item
		workItemDetails, err := w.getWorkItems(w.ctx, workitemtracking.GetWorkItemsArgs{
			Ids: &ids,
		})
		if err != nil {
			w.errSetup = fmt.Errorf("error getting work item details: %w", err)
			return
		}

		// Get comments for each work item
		for i := range *workItemDetails {
			wi := &(*workItemDetails)[i]

			createdBy, ok := (*wi.Fields)["System.CreatedBy"].(map[string]interface{})
			if !ok {
				w.errSetup = errSystemCreatedByFieldNotMap
				return
			}
			uniqueName, ok := createdBy["uniqueName"].(string)
			if !ok {
				w.errSetup = errSystemCreatedByFieldNotUniqueName
				return
			}
			createdDate, ok := (*wi.Fields)["System.CreatedDate"].(string)
			if !ok {
				w.errSetup = errSystemCreatedDateFieldNotString
				return
			}
			parsedTime, err := time.Parse(time.RFC3339, createdDate)
			if err != nil {
				w.errSetup = fmt.Errorf("error parsing created date: %w", err)
				return
			}
			// There is not currently an official API to get user permissions in Azure DevOps
			// so we will default to RepoAssociationMember for all users.
			repoAssociation := clients.RepoAssociationMember

			issue := clients.Issue{
				URI:               wi.Url,
				CreatedAt:         &parsedTime,
				Author:            &clients.User{Login: uniqueName},
				AuthorAssociation: &repoAssociation,
				Comments:          make([]clients.IssueComment, 0),
			}

			workItemComments, err := w.getWorkItemComments(w.ctx, workitemtracking.GetCommentsArgs{
				Project:    &w.repourl.project,
				WorkItemId: wi.Id,
			})
			if err != nil {
				w.errSetup = fmt.Errorf("error getting comments for work item %d: %w", *wi.Id, err)
				return
			}

			for i := range *workItemComments.Comments {
				workItemComment := &(*workItemComments.Comments)[i]

				// There is not currently an official API to get user permissions in Azure DevOps
				// so we will default to RepoAssociationMember for all users.
				repoAssociation := clients.RepoAssociationMember

				comment := clients.IssueComment{
					CreatedAt:         &workItemComment.CreatedDate.Time,
					Author:            &clients.User{Login: *workItemComment.CreatedBy.UniqueName},
					AuthorAssociation: &repoAssociation,
				}

				issue.Comments = append(issue.Comments, comment)
			}

			w.issues = append(w.issues, issue)
		}
	})

	return w.errSetup
}

func (w *workItemsHandler) listIssues() ([]clients.Issue, error) {
	if err := w.setup(); err != nil {
		return nil, fmt.Errorf("error during issuesHandler.setup: %w", err)
	}

	return w.issues, nil
}
