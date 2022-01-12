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

package githubrepo

import (
	"context"
	"fmt"

	"github.com/google/go-github/v38/github"

	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

type workflowsHandler struct {
	client *github.Client
	ctx    context.Context
	owner  string
	repo   string
}

func (handler *workflowsHandler) init(ctx context.Context, owner, repo string) {
	handler.ctx = ctx
	handler.owner = owner
	handler.repo = repo
}

func (handler *workflowsHandler) listSuccessfulWorkflowRuns(filename string) ([]clients.WorkflowRun, error) {
	workflowRuns, _, err := handler.client.Actions.ListWorkflowRunsByFileName(
		handler.ctx, handler.owner, handler.repo, filename, &github.ListWorkflowRunsOptions{
			Status: "success",
		})
	if err != nil {
		return nil, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("ListWorkflowRunsByFileName: %v", err))
	}
	return workflowsRunsFrom(workflowRuns), nil
}

func workflowsRunsFrom(data *github.WorkflowRuns) []clients.WorkflowRun {
	var workflowRuns []clients.WorkflowRun
	for _, workflowRun := range data.WorkflowRuns {
		workflowRuns = append(workflowRuns, clients.WorkflowRun{
			URL: workflowRun.GetURL(),
		})
	}
	return workflowRuns
}
