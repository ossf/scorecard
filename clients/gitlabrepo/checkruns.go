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

	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
)

type checkrunsHandler struct {
	glClient *gitlab.Client
	repourl  *repoURL
}

func (handler *checkrunsHandler) init(repourl *repoURL) {
	handler.repourl = repourl
}

func (handler *checkrunsHandler) listCheckRunsForRef(ref string) ([]clients.CheckRun, error) {
	pipelines, _, err := handler.glClient.Pipelines.ListProjectPipelines(
		handler.repourl.projectID, &gitlab.ListProjectPipelinesOptions{
			SHA:         &ref,
			ListOptions: gitlab.ListOptions{},
		})
	if err != nil {
		return nil, fmt.Errorf("request for pipelines returned error: %w", err)
	}

	return checkRunsFrom(pipelines), nil
}

// Conclusion does not exist in the pipelines for gitlab so I have a placeholder "" as the value.
func checkRunsFrom(data []*gitlab.PipelineInfo) []clients.CheckRun {
	var checkRuns []clients.CheckRun
	for _, pipelineInfo := range data {
		// TODO: Can get more info from GitLab API here (e.g. pipeline name, URL)
		// https://docs.gitlab.com/ee/api/pipelines.html#get-a-pipelines-test-report
		checkRuns = append(checkRuns, parseGitlabStatus(pipelineInfo))
	}
	return checkRuns
}

// Conclusion does not exist in the pipelines for gitlab,
// so we parse the status to determine the conclusion if it exists.
func parseGitlabStatus(info *gitlab.PipelineInfo) clients.CheckRun {
	checkrun := clients.CheckRun{
		URL: info.WebURL,
	}
	completed := "completed"

	switch info.Status {
	case "created", "waiting_for_resource", "preparing", "pending", "scheduled":
		checkrun.Status = "queued"
	case "running":
		checkrun.Status = "in-progress"
	case "failed":
		checkrun.Status = completed
		checkrun.Conclusion = "failure"
	case "success":
		checkrun.Status = completed
		checkrun.Conclusion = "success"
	case "canceled":
		checkrun.Status = completed
		checkrun.Conclusion = "cancelled"
	case "skipped":
		checkrun.Status = completed
		checkrun.Conclusion = "skipped"
	case "manual":
		checkrun.Status = completed
		checkrun.Conclusion = "action_required"
	default:
		checkrun.Status = info.Status
	}

	return checkrun
}

// Notes:
// ###################################

// # Gitlab Status values
// created: 'Pipeline has been created.',
// waiting_for_resource: 'A resource (for example, a runner) that the pipeline requires to run is unavailable.',
// preparing: 'Pipeline is preparing to run.',
// pending: 'Pipeline has not started running yet.',
// running: 'Pipeline is running.',
// failed: 'At least one stage of the pipeline failed.',
// success: 'Pipeline completed successfully.',
// canceled: 'Pipeline was canceled before completion.',
// skipped: 'Pipeline was skipped.',
// manual: 'Pipeline needs to be manually started.',
// scheduled: 'Pipeline is scheduled to run.'

// # Github Conclusion Values
// "success",
// "failure",
// "neutral",
// "cancelled",
// "skipped",
// "timed_out",
// "action_required",
// null

// # Github Status values
// "queued",
// "in_progress",
// "completed"
