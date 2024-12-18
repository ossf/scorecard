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

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"

	"github.com/ossf/scorecard/v5/clients"
)

type buildsHandler struct {
	ctx                 context.Context
	repourl             *Repo
	buildClient         build.Client
	getBuildDefinitions fnListBuildDefinitions
	getBuilds           fnGetBuilds
}

type (
	fnListBuildDefinitions func(
		ctx context.Context,
		args build.GetDefinitionsArgs,
	) (*build.GetDefinitionsResponseValue, error)
	fnGetBuilds func(
		ctx context.Context,
		args build.GetBuildsArgs,
	) (*build.GetBuildsResponseValue, error)
)

func (b *buildsHandler) init(ctx context.Context, repourl *Repo) {
	b.ctx = ctx
	b.repourl = repourl
	b.getBuildDefinitions = b.buildClient.GetDefinitions
	b.getBuilds = b.buildClient.GetBuilds
}

func (b *buildsHandler) listSuccessfulBuilds(filename string) ([]clients.WorkflowRun, error) {
	buildDefinitions := make([]build.BuildDefinitionReference, 0)

	includeAllProperties := true
	repositoryType := "TfsGit"
	continuationToken := ""
	for {
		args := build.GetDefinitionsArgs{
			Project:              &b.repourl.project,
			RepositoryId:         &b.repourl.id,
			RepositoryType:       &repositoryType,
			IncludeAllProperties: &includeAllProperties,
			YamlFilename:         &filename,
			ContinuationToken:    &continuationToken,
		}

		response, err := b.getBuildDefinitions(b.ctx, args)
		if err != nil {
			return nil, err
		}

		buildDefinitions = append(buildDefinitions, response.Value...)

		if response.ContinuationToken == "" {
			break
		}
		continuationToken = response.ContinuationToken
	}

	buildIds := make([]int, 0, len(buildDefinitions))
	for i := range buildDefinitions {
		buildIds = append(buildIds, *buildDefinitions[i].Id)
	}

	args := build.GetBuildsArgs{
		Project:      &b.repourl.project,
		Definitions:  &buildIds,
		ResultFilter: &build.BuildResultValues.Succeeded,
	}
	builds, err := b.getBuilds(b.ctx, args)
	if err != nil {
		return nil, err
	}

	workflowRuns := make([]clients.WorkflowRun, 0, len(builds.Value))
	for i := range builds.Value {
		currentBuild := builds.Value[i]
		workflowRuns = append(workflowRuns, clients.WorkflowRun{
			URL:     *currentBuild.Url,
			HeadSHA: currentBuild.SourceVersion,
		})
	}

	return workflowRuns, nil
}
