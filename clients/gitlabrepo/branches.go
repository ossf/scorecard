// Copyright 2022 Security Scorecard Authors
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
	"sync"

	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
)

type branchesHandler struct {
	glClient         *gitlab.Client
	once             *sync.Once
	errSetup         error
	repourl          *repoURL
	defaultBranchRef *clients.BranchRef
}

func (handler *branchesHandler) init(repourl *repoURL) {
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
}

func (handler *branchesHandler) setup() error {
	handler.once.Do(func() {
		if !strings.EqualFold(handler.repourl.commitSHA, clients.HeadSHA) {
			handler.errSetup = fmt.Errorf("%w: branches only supported for HEAD queries", clients.ErrUnsupportedFeature)
			return
		}

		proj, _, err := handler.glClient.Projects.GetProject(handler.repourl.projectID, &gitlab.GetProjectOptions{})
		if err != nil {
			handler.errSetup = fmt.Errorf("requirest for project failed with error %w", err)
			return
		}

		branch, _, err := handler.glClient.Branches.GetBranch(handler.repourl.projectID, proj.DefaultBranch)
		if err != nil {
			handler.errSetup = fmt.Errorf("request for default branch failed with error %w", err)
			return
		}

		if branch.Protected {
			protectedBranch, resp, err := handler.glClient.ProtectedBranches.GetProtectedBranch(
				handler.repourl.projectID, branch.Name)
			if err != nil && resp.StatusCode != 403 {
				handler.errSetup = fmt.Errorf("request for protected branch failed with error %w", err)
				return
			} else if resp.StatusCode == 403 {
				handler.errSetup = fmt.Errorf("incorrect permissions to fully check branch protection %w", err)
				return
			}

			projectStatusChecks, resp, err := handler.glClient.ExternalStatusChecks.ListProjectStatusChecks(
				handler.repourl.projectID, &gitlab.ListOptions{})
			if err != nil && resp.StatusCode != 404 {
				handler.errSetup = fmt.Errorf("request for external status checks failed with error %w", err)
				return
			}

			projectApprovalRule, resp, err := handler.glClient.Projects.GetApprovalConfiguration(handler.repourl.projectID)
			if err != nil && resp.StatusCode != 404 {
				handler.errSetup = fmt.Errorf("request for project approval rule failed with %w", err)
				return
			}

			handler.defaultBranchRef = makeBranchRefFrom(branch, protectedBranch,
				projectStatusChecks, projectApprovalRule)
		} else {
			handler.defaultBranchRef = &clients.BranchRef{
				Name:      &branch.Name,
				Protected: &branch.Protected,
			}
		}
		handler.errSetup = nil
	})
	return handler.errSetup
}

func (handler *branchesHandler) getDefaultBranch() (*clients.BranchRef, error) {
	err := handler.setup()
	if err != nil {
		return nil, fmt.Errorf("error during branchesHandler.setup: %w", err)
	}

	return handler.defaultBranchRef, nil
}

func (handler *branchesHandler) getBranch(branch string) (*clients.BranchRef, error) {
	bran, _, err := handler.glClient.Branches.GetBranch(handler.repourl.projectID, branch)
	if err != nil {
		return nil, fmt.Errorf("error getting branch in branchsHandler.getBranch: %w", err)
	}

	if bran.Protected {
		protectedBranch, _, err := handler.glClient.ProtectedBranches.GetProtectedBranch(handler.repourl.projectID, bran.Name)
		if err != nil {
			return nil, fmt.Errorf("request for protected branch failed with error %w", err)
		}

		projectStatusChecks, resp, err := handler.glClient.ExternalStatusChecks.ListProjectStatusChecks(
			handler.repourl.projectID, &gitlab.ListOptions{})
		if err != nil && resp.StatusCode != 404 {
			return nil, fmt.Errorf("request for external status checks failed with error %w", err)
		}

		projectApprovalRule, resp, err := handler.glClient.Projects.GetApprovalConfiguration(handler.repourl.projectID)
		if err != nil && resp.StatusCode != 404 {
			return nil, fmt.Errorf("request for project approval rule failed with %w", err)
		}

		return makeBranchRefFrom(bran, protectedBranch, projectStatusChecks, projectApprovalRule), nil
	} else {
		ret := &clients.BranchRef{
			Name:      &bran.Name,
			Protected: &bran.Protected,
		}
		return ret, nil
	}
}

func makeContextsFromResp(checks []*gitlab.ProjectStatusCheck) []string {
	ret := make([]string, len(checks))
	for i, statusCheck := range checks {
		ret[i] = statusCheck.Name
	}
	return ret
}

func makeBranchRefFrom(branch *gitlab.Branch, protectedBranch *gitlab.ProtectedBranch,
	projectStatusChecks []*gitlab.ProjectStatusCheck,
	projectApprovalRule *gitlab.ProjectApprovals,
) *clients.BranchRef {
	requiresStatusChecks := newFalse()
	if len(projectStatusChecks) > 0 {
		requiresStatusChecks = newTrue()
	}

	statusChecksRule := clients.StatusChecksRule{
		UpToDateBeforeMerge:  newTrue(),
		RequiresStatusChecks: requiresStatusChecks,
		Contexts:             makeContextsFromResp(projectStatusChecks),
	}

	pullRequestReviewRule := clients.PullRequestReviewRule{
		DismissStaleReviews:     newTrue(),
		RequireCodeOwnerReviews: &protectedBranch.CodeOwnerApprovalRequired,
	}

	if projectApprovalRule != nil {
		requiredApprovalNum := int32(projectApprovalRule.ApprovalsBeforeMerge)
		pullRequestReviewRule.RequiredApprovingReviewCount = &requiredApprovalNum
	}

	ret := &clients.BranchRef{
		Name:      &branch.Name,
		Protected: &branch.Protected,
		BranchProtectionRule: clients.BranchProtectionRule{
			RequiredPullRequestReviews: pullRequestReviewRule,
			AllowDeletions:             newFalse(),
			AllowForcePushes:           &protectedBranch.AllowForcePush,
			EnforceAdmins:              newTrue(),
			CheckRules:                 statusChecksRule,
		},
	}

	return ret
}

func newTrue() *bool {
	b := true
	return &b
}

func newFalse() *bool {
	b := false
	return &b
}
