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
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
)

var errNoBranchFound = errors.New("branch given in ref could not be found")

type branchesHandler struct {
	glClient         *gitlab.Client
	once             *sync.Once
	errSetup         error
	repourl          *repoURL
	defaultBranchRef *clients.BranchRef
	branchNames      []string
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
			if err != nil && resp.StatusCode != 404 && resp.StatusCode != 401 {
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

func (handler *branchesHandler) getBranch(commitOrBranch string) (*clients.BranchRef, error) {
	// If the given string is a branch name then the branch can be found by simply getting branch by it's name.
	bran, resp, err := handler.glClient.Branches.GetBranch(handler.repourl.projectID, commitOrBranch)
	if err != nil && resp.StatusCode != 404 {
		return nil, fmt.Errorf("error getting branch in branchsHandler.getBranch: %w", err)
	}

	// Unfortunately it seems that most of the time commitOrBranch will be a release branch commit.
	// In this case, GitLab creates a commit whenever a release is triggered so we can create a time before the commit
	// and after the commit such that the only branch with a commit at that time will be the release branch.
	if bran == nil {
		// Get commit from commitOrBranch string.
		commit, _, err := handler.glClient.Commits.GetCommit(handler.repourl.projectID, commitOrBranch)
		if err != nil {
			return nil, fmt.Errorf("given commit sha did not align with any known commits: %w", err)
		}

		// We need the various branches of the project to query each branch individually for the release branch name.
		if handler.branchNames == nil {
			branches, _, err := handler.glClient.Branches.ListBranches(handler.repourl.projectID, &gitlab.ListBranchesOptions{})
			if err != nil {
				return nil, fmt.Errorf("could not get the branches from the given GitLab project: %w", err)
			}
			for _, branch := range branches {
				handler.branchNames = append(handler.branchNames, branch.Name)
			}
		}

		// Our boundary around the given commit will be 1 minute
		beforeTime := commit.CreatedAt.Add(-1 * time.Minute)
		afterTime := commit.CreatedAt.Add(1 * time.Minute)
		var branchName string
		for i, name := range handler.branchNames {
			// The main branch seems to always be involved in release branches so we will disregard it here.
			if strings.EqualFold(name, "main") {
				continue
			}

			// Above we obtained the commit associated with this release, however the only way to obtain the associated
			// branch of the commit seems to be to query each branch until we find a commit with the same CreatedAt time.
			possibleCommits, resp, err := handler.glClient.Commits.ListCommits(handler.repourl.projectID,
				&gitlab.ListCommitsOptions{
					RefName: &handler.branchNames[i],
					Since:   &beforeTime,
					Until:   &afterTime,
				})
			if err != nil && resp.StatusCode != 404 {
				return nil, fmt.Errorf("error finding possible list of commits to find release branch: %w", err)
			}

			for _, com := range possibleCommits {
				if com.ID == commitOrBranch {
					branchName = name
					break
				}
			}

			if branchName != "" {
				break
			}
		}

		if branchName == "" {
			return nil, errNoBranchFound
		}

		bran, _, err = handler.glClient.Branches.GetBranch(handler.repourl.projectID, branchName)
		if err != nil {
			return nil, fmt.Errorf("could not obtain the branch: %w", err)
		}
	}

	if bran.Protected {
		protectedBranch, _, err := handler.glClient.ProtectedBranches.GetProtectedBranch(handler.repourl.projectID, bran.Name)
		if err != nil {
			return nil, fmt.Errorf("request for protected branch failed with error %w", err)
		}

		projectStatusChecks, resp, err := handler.glClient.ExternalStatusChecks.ListProjectStatusChecks(
			handler.repourl.projectID, &gitlab.ListOptions{})
		// Project Status Checks are only allowed for GitLab ultimate members so we will assume they are
		// null if user does not have permissions.
		if err != nil && resp.StatusCode != 404 && resp.StatusCode != 401 {
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
	if checks == nil {
		return nil
	}
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
	if projectStatusChecks != nil {
		if len(projectStatusChecks) > 0 {
			requiresStatusChecks = newTrue()
		}
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
