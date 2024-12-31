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

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/policy"

	"github.com/ossf/scorecard/v5/clients"
)

var errPullRequestNotFound = errors.New("pull request not found")

type policyHandler struct {
	ctx                  context.Context
	repourl              *Repo
	gitClient            git.Client
	policyClient         policy.Client
	getPullRequestQuery  fnGetPullRequestQuery
	getPolicyEvaluations fnGetPolicyEvaluations
}

type fnGetPolicyEvaluations func(
	ctx context.Context,
	args policy.GetPolicyEvaluationsArgs,
) (*[]policy.PolicyEvaluationRecord, error)

func (p *policyHandler) init(ctx context.Context, repourl *Repo) {
	p.ctx = ctx
	p.repourl = repourl
	p.getPullRequestQuery = p.gitClient.GetPullRequestQuery
	p.getPolicyEvaluations = p.policyClient.GetPolicyEvaluations
}

func (p *policyHandler) listCheckRunsForRef(ref string) ([]clients.CheckRun, error) {
	// The equivalent of a check run in Azure DevOps is a policy evaluation.
	// Unfortunately, Azure DevOps does not provide a way to list policy evaluations for a specific ref.
	// Instead, they are associated with a pull request.

	// Get the pull request associated with the ref.
	args := git.GetPullRequestQueryArgs{
		RepositoryId: &p.repourl.id,
		Queries: &git.GitPullRequestQuery{
			Queries: &[]git.GitPullRequestQueryInput{
				{
					Type:  &git.GitPullRequestQueryTypeValues.LastMergeCommit,
					Items: &[]string{ref},
				},
			},
		},
	}
	queryPullRequests, err := p.getPullRequestQuery(p.ctx, args)
	if err != nil {
		return nil, err
	}

	if len(*queryPullRequests.Results) != 1 {
		return nil, errMultiplePullRequests
	}
	result := (*queryPullRequests.Results)[0]
	pullRequests, ok := result[ref]
	if !ok {
		return nil, errPullRequestNotFound
	}

	if len(pullRequests) != 1 {
		return nil, errMultiplePullRequests
	}

	pullRequest := pullRequests[0]

	artifactID := fmt.Sprintf("vstfs:///CodeReview/CodeReviewId/%s/%d", p.repourl.projectID, *pullRequest.PullRequestId)
	argsPolicy := policy.GetPolicyEvaluationsArgs{
		Project:    &p.repourl.project,
		ArtifactId: &artifactID,
	}
	policyEvaluations, err := p.getPolicyEvaluations(p.ctx, argsPolicy)
	if err != nil {
		return nil, err
	}

	const completed = "completed"

	checkRuns := make([]clients.CheckRun, len(*policyEvaluations))
	for i, evaluation := range *policyEvaluations {
		checkrun := clients.CheckRun{}

		switch *evaluation.Status {
		case policy.PolicyEvaluationStatusValues.Queued:
			checkrun.Status = "queued"
		case policy.PolicyEvaluationStatusValues.Running:
			checkrun.Status = "in_progress"
		case policy.PolicyEvaluationStatusValues.Approved:
			checkrun.Status = completed
			checkrun.Conclusion = "success"
		case policy.PolicyEvaluationStatusValues.Rejected, policy.PolicyEvaluationStatusValues.Broken:
			checkrun.Status = completed
			checkrun.Conclusion = "failure"
		case policy.PolicyEvaluationStatusValues.NotApplicable:
			checkrun.Status = completed
			checkrun.Conclusion = "neutral"
		default:
			checkrun.Status = string(*evaluation.Status)
		}

		checkRuns[i] = checkrun
	}

	return checkRuns, nil
}
