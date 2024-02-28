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
	"regexp"

	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
)

var gitCommitHashRegex = regexp.MustCompile(`^[a-fA-F0-9]{40}$`)

type checkrunsHandler struct {
	glClient *gitlab.Client
	repourl  *repoURL
}

func (handler *checkrunsHandler) init(repourl *repoURL) {
	handler.repourl = repourl
}

func (handler *checkrunsHandler) listCheckRunsForRef(ref string) ([]clients.CheckRun, error) {
	var options gitlab.ListProjectPipelinesOptions

	if gitCommitHashRegex.MatchString(ref) {
		options.SHA = &ref
	} else {
		options.Ref = &ref
	}

	// Notes for GitLab ListProjectPipelines endpoint:
	// Only full SHA works for SHA param, Short SHA does not work
	// Branch names work for Ref Param, tags and SHAs do not work
	// Reference: https://docs.gitlab.com/ee/api/pipelines.html#list-project-pipelines
	pipelines, _, err := handler.glClient.Pipelines.ListProjectPipelines(handler.repourl.projectID, &options)
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
	const completed = "completed"

	switch info.Status {
	case "created", "waiting_for_resource", "preparing", "pending", "scheduled":
		checkrun.Status = "queued"
	case "running":
		checkrun.Status = "in_progress"
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
