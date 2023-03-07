// Copyright 2022 OpenSSF Scorecard Authors
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

package gitlabrepo

import (
	"fmt"
	"strings"

	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
)

type workflowsHandler struct {
	glClient *gitlab.Client
	repourl  *repoURL
}

func (handler *workflowsHandler) init(repourl *repoURL) {
	handler.repourl = repourl
}

func (handler *workflowsHandler) listSuccessfulWorkflowRuns(filename string) ([]clients.WorkflowRun, error) {
	var buildStates []gitlab.BuildStateValue
	buildStates = append(buildStates, gitlab.Success)
	jobs, _, err := handler.glClient.Jobs.ListProjectJobs(handler.repourl.project,
		&gitlab.ListJobsOptions{Scope: &buildStates})
	if err != nil {
		return nil, fmt.Errorf("error getting project jobs: %w", err)
	}

	return workflowsRunsFrom(jobs, filename), nil
}

func workflowsRunsFrom(data []*gitlab.Job, filename string) []clients.WorkflowRun {
	var workflowRuns []clients.WorkflowRun
	for _, job := range data {
		// Find a better way to do this.
		for _, artifact := range job.Artifacts {
			if strings.EqualFold(artifact.Filename, filename) {
				workflowRuns = append(workflowRuns, clients.WorkflowRun{
					HeadSHA: &job.Pipeline.Sha,
					URL:     job.WebURL,
				})
				continue
			}
		}
	}
	return workflowRuns
}
