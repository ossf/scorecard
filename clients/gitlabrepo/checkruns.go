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

type checkrunsHandler struct {
	glClient *gitlab.Client
	repourl  *repoURL
}

func (handler *checkrunsHandler) init(repourl *repoURL) {
	handler.repourl = repourl
}

func (handler *checkrunsHandler) listCheckRunsForRef(ref string) ([]clients.CheckRun, error) {
	pipelines, _, err := handler.glClient.Pipelines.ListProjectPipelines(
		handler.repourl.project, &gitlab.ListProjectPipelinesOptions{})
	if err != nil {
		return nil, fmt.Errorf("request for pipelines returned error: %w", err)
	}

	return checkRunsFrom(pipelines, ref), nil
}

// Conclusion does not exist in the pipelines for gitlab so I have a placeholder "" as the value.
func checkRunsFrom(data []*gitlab.PipelineInfo, ref string) []clients.CheckRun {
	var checkRuns []clients.CheckRun
	for _, pipelineInfo := range data {
		if strings.EqualFold(pipelineInfo.Ref, ref) {
			checkRuns = append(checkRuns, clients.CheckRun{
				Status:     pipelineInfo.Status,
				Conclusion: "",
				URL:        pipelineInfo.WebURL,
				App:        clients.CheckRunApp{Slug: pipelineInfo.Source},
			})
		}
	}
	return checkRuns
}
