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
	"net/http"
	"strings"
	"sync"

	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
)

type branchesHandler struct {
	glClient                 *gitlab.Client
	once                     *sync.Once
	errSetup                 error
	repourl                  *repoURL
	defaultBranchRef         *clients.BranchRef
	queryProject             fnProject
	queryBranch              fnQueryBranch
	getProtectedBranch       fnProtectedBranch
	getProjectChecks         fnListProjectStatusChecks
	getApprovalConfiguration fnGetApprovalConfiguration
}

func (handler *branchesHandler) init(repourl *repoURL) {
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
	handler.queryProject = handler.glClient.Projects.GetProject
	handler.queryBranch = handler.glClient.Branches.GetBranch
	handler.getProtectedBranch = handler.glClient.ProtectedBranches.GetProtectedBranch
	handler.getProjectChecks = handler.glClient.ExternalStatusChecks.ListProjectStatusChecks
	handler.getApprovalConfiguration = handler.glClient.Projects.GetApprovalConfiguration
}

type (
	fnProject func(pid interface{}, opt *gitlab.GetProjectOptions,
		options ...gitlab.RequestOptionFunc) (*gitlab.Project, *gitlab.Response, error)
	fnQueryBranch func(pid interface{}, branch string,
		options ...gitlab.RequestOptionFunc) (*gitlab.Branch, *gitlab.Response, error)
	fnProtectedBranch func(pid interface{}, branch string,
		options ...gitlab.RequestOptionFunc) (*gitlab.ProtectedBranch, *gitlab.Response, error)
	fnListProjectStatusChecks func(pid interface{}, opt *gitlab.ListOptions,
		options ...gitlab.RequestOptionFunc) ([]*gitlab.ProjectStatusCheck, *gitlab.Response, error)
	fnGetApprovalConfiguration func(pid interface{},
		options ...gitlab.RequestOptionFunc) (*gitlab.ProjectApprovals, *gitlab.Response, error)
)

// nolint: nestif
func (handler *branchesHandler) setup() error {
	handler.once.Do(func() {
		if !strings.EqualFold(handler.repourl.commitSHA, clients.HeadSHA) {
			handler.errSetup = fmt.Errorf("%w: branches only supported for HEAD queries", clients.ErrUnsupportedFeature)
			return
		}

		proj, _, err := handler.queryProject(handler.repourl.projectID, &gitlab.GetProjectOptions{})
		if err != nil {
			handler.errSetup = fmt.Errorf("request for project failed with error %w", err)
			return
		}

		branch, _, err := handler.queryBranch(handler.repourl.projectID, proj.DefaultBranch)
		if err != nil {
			handler.errSetup = fmt.Errorf("request for default branch failed with error %w", err)
			return
		}

		if branch.Protected {
			protectedBranch, resp, err := handler.getProtectedBranch(
				handler.repourl.projectID, branch.Name)
			if err != nil && resp.StatusCode != http.StatusForbidden {
				handler.errSetup = fmt.Errorf("request for protected branch failed with error %w", err)
				return
			} else if resp.StatusCode == http.StatusForbidden {
				handler.errSetup = fmt.Errorf("incorrect permissions to fully check branch protection %w", err)
				return
			}

			projectStatusChecks, resp, err := handler.getProjectChecks(handler.repourl.projectID, &gitlab.ListOptions{})

			if resp.StatusCode != 200 || err != nil {
				handler.errSetup = fmt.Errorf("request for external status checks failed with error %w", err)
			}

			projectApprovalRule, resp, err := handler.getApprovalConfiguration(handler.repourl.projectID)
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
	if strings.Contains(branch, "/-/commit/") {
		// Gitlab's release commitish contains commit and is not easily tied to specific branch
		p, b := true, ""
		ret := &clients.BranchRef{
			Name:      &b,
			Protected: &p,
		}
		return ret, nil
	}

	bran, _, err := handler.queryBranch(handler.repourl.projectID, branch)
	if err != nil {
		return nil, fmt.Errorf("error getting branch in branchesHandler.getBranch: %w", err)
	}

	if bran.Protected {
		protectedBranch, _, err := handler.getProtectedBranch(handler.repourl.projectID, bran.Name)
		if err != nil {
			return nil, fmt.Errorf("request for protected branch failed with error %w", err)
		}

		projectStatusChecks, resp, err := handler.getProjectChecks(
			handler.repourl.projectID, &gitlab.ListOptions{})
		if err != nil && resp.StatusCode != 404 {
			return nil, fmt.Errorf("request for external status checks failed with error %w", err)
		}

		projectApprovalRule, resp, err := handler.getApprovalConfiguration(handler.repourl.projectID)
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
