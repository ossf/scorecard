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
	"fmt"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"

	"github.com/ossf/scorecard/v5/clients"
)

const azureGitRepositoryType = "TfsGit"

var errBuildData = errors.New("invalid Azure DevOps build data")

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
	repositoryType := azureGitRepositoryType
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

	if len(buildDefinitions) == 0 {
		return []clients.WorkflowRun{}, nil
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

func (b *buildsHandler) listCheckRunsForPullRequest(
	pr *git.GitPullRequest, gitClient git.Client,
) ([]clients.CheckRun, error) {
	if pr.SourceRefName == nil || *pr.SourceRefName == "" {
		return nil, fmt.Errorf("%w: missing source branch for PR %d", errInvalidPullRequest, *pr.PullRequestId)
	}
	runs := []clients.CheckRun{}
	for _, branch := range []string{*pr.SourceRefName, fmt.Sprintf("refs/pull/%d/merge", *pr.PullRequestId)} {
		checks, err := b.listBuildChecks(pr, branch, gitClient)
		if err != nil {
			return nil, err
		}
		runs = append(runs, checks...)
	}
	return runs, nil
}

func (b *buildsHandler) listBuildChecks(
	pr *git.GitPullRequest, branch string, gitClient git.Client,
) ([]clients.CheckRun, error) {
	repositoryType := azureGitRepositoryType
	args := build.GetBuildsArgs{
		Project: &b.repourl.project, RepositoryId: &b.repourl.id, RepositoryType: &repositoryType,
		BranchName: &branch, MaxTime: pr.ClosedDate,
		QueryOrder:   &build.BuildQueryOrderValues.FinishTimeDescending,
		StatusFilter: &build.BuildStatusValues.Completed, ResultFilter: &build.BuildResultValues.Succeeded,
	}
	if branch != *pr.SourceRefName {
		args.ReasonFilter = &build.BuildReasonValues.PullRequest
	}
	runs := []clients.CheckRun{}
	seenTokens := make(map[string]bool)
	for {
		response, err := b.getBuilds(b.ctx, args)
		if err != nil {
			return nil, fmt.Errorf("request for builds for PR %d failed: %w", *pr.PullRequestId, err)
		}
		if response == nil || response.Value == nil {
			return nil, fmt.Errorf("%w: missing builds response", errBuildData)
		}
		for i := range response.Value {
			current := &response.Value[i]
			matches, err := b.isBuildForPullRequest(current, pr, branch, gitClient)
			if err != nil {
				return nil, err
			}
			if !matches {
				continue
			}
			evidenceURL, err := buildEvidenceURL(current)
			if err != nil {
				return nil, err
			}
			runs = append(runs, clients.CheckRun{
				Status: "completed", Conclusion: "success", URL: evidenceURL,
				App: clients.CheckRunApp{Slug: "azure-pipelines"},
			})
		}
		token := response.ContinuationToken
		if token == "" {
			return runs, nil
		}
		if seenTokens[token] {
			return nil, fmt.Errorf("%w: repeated builds continuation token", errBuildData)
		}
		seenTokens[token] = true
		args.ContinuationToken = &token
	}
}

func (b *buildsHandler) isBuildForPullRequest(current *build.Build, pr *git.GitPullRequest,
	branch string, gitClient git.Client,
) (bool, error) {
	if current.Repository == nil || current.Repository.Id == nil || current.Repository.Type == nil ||
		current.SourceBranch == nil || current.SourceVersion == nil || current.Status == nil ||
		*current.Repository.Id == "" || *current.Repository.Type == "" ||
		*current.SourceBranch == "" || *current.SourceVersion == "" || *current.Status == "" {
		return false, fmt.Errorf("%w: missing repository, revision, or status", errBuildData)
	}
	if !strings.EqualFold(*current.Repository.Id, b.repourl.id) || *current.Repository.Type != azureGitRepositoryType ||
		*current.SourceBranch != branch ||
		*current.Status != build.BuildStatusValues.Completed {
		return false, nil
	}
	if current.Result == nil {
		return false, fmt.Errorf("%w: missing completed build result", errBuildData)
	}
	if *current.Result != build.BuildResultValues.Succeeded {
		return false, nil
	}
	if current.FinishTime == nil || current.FinishTime.Time.IsZero() {
		return false, fmt.Errorf("%w: missing successful build finish time", errBuildData)
	}
	if current.FinishTime.Time.After(pr.ClosedDate.Time) {
		return false, nil
	}
	if branch == *pr.SourceRefName {
		return strings.EqualFold(*current.SourceVersion, *pr.LastMergeSourceCommit.CommitId), nil
	}
	return b.isPRMergeBuild(current, pr, gitClient)
}

func (b *buildsHandler) isPRMergeBuild(
	current *build.Build, pr *git.GitPullRequest, gitClient git.Client,
) (bool, error) {
	if current.Reason == nil {
		return false, fmt.Errorf("%w: missing PR build reason", errBuildData)
	}
	if *current.Reason != build.BuildReasonValues.PullRequest {
		return false, nil
	}
	if strings.EqualFold(*current.SourceVersion, *pr.LastMergeCommit.CommitId) {
		return true, nil
	}
	if pr.LastMergeTargetCommit == nil || pr.LastMergeTargetCommit.CommitId == nil ||
		*pr.LastMergeTargetCommit.CommitId == "" {
		return false, fmt.Errorf("%w: missing target commit for PR %d", errInvalidPullRequest, *pr.PullRequestId)
	}
	commit, err := gitClient.GetCommit(b.ctx, git.GetCommitArgs{
		Project: &b.repourl.project, RepositoryId: &b.repourl.id, CommitId: current.SourceVersion,
	})
	if err != nil {
		return false, fmt.Errorf("request for tested merge commit for PR %d failed: %w", *pr.PullRequestId, err)
	}
	if commit == nil || commit.CommitId == nil || !strings.EqualFold(*commit.CommitId, *current.SourceVersion) ||
		commit.Parents == nil {
		return false, fmt.Errorf("%w: incomplete tested merge commit", errBuildData)
	}
	parents := *commit.Parents
	if len(parents) != 2 {
		return false, nil
	}
	source, target := *pr.LastMergeSourceCommit.CommitId, *pr.LastMergeTargetCommit.CommitId
	return (strings.EqualFold(parents[0], source) && strings.EqualFold(parents[1], target)) ||
		(strings.EqualFold(parents[1], source) && strings.EqualFold(parents[0], target)), nil
}

func buildEvidenceURL(current *build.Build) (string, error) {
	if links, ok := current.Links.(map[string]interface{}); ok {
		if web, ok := links["web"].(map[string]interface{}); ok {
			if href, ok := web["href"].(string); ok && href != "" {
				return href, nil
			}
		}
	}
	if current.Url != nil && *current.Url != "" {
		return *current.Url, nil
	}
	return "", fmt.Errorf("%w: missing successful build URL", errBuildData)
}
