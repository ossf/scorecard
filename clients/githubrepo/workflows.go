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
	"errors"
	"fmt"
<<<<<<< HEAD
	"strings"
=======
	"net/url"
	"reflect"
>>>>>>> adbef21 (use check suite ID)

	"github.com/google/go-github/v38/github"
	"github.com/google/go-querystring/query"

	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

type workflowsHandler struct {
	client  *github.Client
	ctx     context.Context
	repourl *repoURL
}

var (
	errWorkflowEmptyID   = errors.New("workflow ID is empty")
	errWorkflowEmptyName = errors.New("workflow Name is empty")
	errWorkflowEmptyPath = errors.New("workflow Path is empty")
)

func (handler *workflowsHandler) init(ctx context.Context, repourl *repoURL) {
	handler.ctx = ctx
	handler.repourl = repourl
}

func (handler *workflowsHandler) GetWorkflowByFileName(filename string) (clients.Workflow, error) {
	workflow, _, err := handler.client.Actions.GetWorkflowByFileName(
		handler.ctx, handler.owner, handler.repo, filename)
	if err != nil {
		return clients.Workflow{}, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("GetWorkflowByFileName: %v", err))
	}
	if workflow.ID == nil {
		return clients.Workflow{}, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("GetWorkflowByFileName: %v", errWorkflowEmptyID))
	}
	if workflow.Name == nil {
		return clients.Workflow{}, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("GetWorkflowByFileName: %v", errWorkflowEmptyName))
	}
	if workflow.Path == nil {
		return clients.Workflow{}, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("GetWorkflowByFileName: %v", errWorkflowEmptyPath))
	}

	return clients.Workflow{
		ID:   *workflow.ID,
		Path: *workflow.Path,
		Name: *workflow.Name,
	}, nil
}

func addOptions(s string, opts interface{}) (string, error) {
	v := reflect.ValueOf(opts)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	qs, err := query.Values(opts)
	if err != nil {
		return s, err
	}

	u.RawQuery = qs.Encode()
	return u.String(), nil
}

func (handler *workflowsHandler) listWorkflowRuns(opts *clients.ListWorkflowRunOptions) ([]clients.WorkflowRun, error) {
	endpoint := fmt.Sprintf("repos/%s/%s/actions/runs", handler.owner, handler.repo)
	u, err := addOptions(endpoint, opts)
	if err != nil {
		return nil, err
	}

	req, err := handler.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("ListWorkflowRuns: %v", err))
	}

	workflowRuns := new(github.WorkflowRuns)
	_, err = handler.client.Do(handler.ctx, req, &workflowRuns)
	if err != nil {
		return nil, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("ListWorkflowRuns: %v", err))
	}

	return workflowsRunsFrom(workflowRuns), nil
}

func (handler *workflowsHandler) listSuccessfulWorkflowRuns(filename string) ([]clients.WorkflowRun, error) {
	if !strings.EqualFold(handler.repourl.commitSHA, clients.HeadSHA) {
		return nil, fmt.Errorf(
			"%w: ListWorkflowRunsByFileName only supported for HEAD queries", clients.ErrUnsupportedFeature)
	}
	workflowRuns, _, err := handler.client.Actions.ListWorkflowRunsByFileName(
		handler.ctx, handler.repourl.owner, handler.repourl.repo, filename, &github.ListWorkflowRunsOptions{
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
		r := clients.WorkflowRun{
			URL:        workflowRun.GetURL(),
			WorkflowID: workflowRun.WorkflowID,
		}

		/*prs := workflowRun.PullRequests
		fmt.Println(len(prs))
		for _, pr := range prs {
			cp := clients.PullRequest{
				// TODO: fill up the rest of the structure.
				Number: pr.GetNumber(),
			}
			r.PullRequests = append(r.PullRequests, cp)
		}
		*/
		workflowRuns = append(workflowRuns, r)
	}
	return workflowRuns
}
