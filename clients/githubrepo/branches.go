// Copyright 2021 OpenSSF Scorecard Authors
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

package githubrepo

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/google/go-github/v53/github"
	"github.com/shurcooL/githubv4"

	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

const (
	refPrefix = "refs/heads/"
)

// See https://github.community/t/graphql-api-protected-branch/14380
/* Example of query:
	query {
  repository(owner: "laurentsimon", name: "test3") {
    branchProtectionRules(first: 100) {
		edges{
			node{
				allowsDeletions
				allowsForcePushes
				dismissesStaleReviews
				isAdminEnforced
				...
				pattern
				matchingRefs(first: 100) {
				nodes {
					name
				}
			}
		}
    }
    refs(first: 100, refPrefix: "refs/heads/") {
      nodes {
        name
        refUpdateRule {
          requiredApprovingReviewCount
          allowsForcePushes
		  ...
        }
      }
    }
  }
}
*/

// Used for non-admin settings.
type refUpdateRule struct {
	AllowsDeletions              *bool
	AllowsForcePushes            *bool
	RequiredApprovingReviewCount *int32
	RequiresCodeOwnerReviews     *bool
	RequiresLinearHistory        *bool
	RequiredStatusCheckContexts  []string
}

// Used for all settings, both admin and non-admin ones.
// This only works with an admin token.
type branchProtectionRule struct {
	DismissesStaleReviews        *bool
	IsAdminEnforced              *bool
	RequiresStrictStatusChecks   *bool
	RequiresStatusChecks         *bool
	AllowsDeletions              *bool
	AllowsForcePushes            *bool
	RequiredApprovingReviewCount *int32
	RequiresCodeOwnerReviews     *bool
	RequiresLinearHistory        *bool
	RequireLastPushApproval      *bool
	RequiredStatusCheckContexts  []string
	// TODO: verify there is no conflicts.
	// BranchProtectionRuleConflicts interface{}
}

type branch struct {
	Name                 *string
	RefUpdateRule        *refUpdateRule
	BranchProtectionRule *branchProtectionRule
}
type defaultBranchData struct {
	Repository struct {
		DefaultBranchRef *branch
	} `graphql:"repository(owner: $owner, name: $name)"`
	RateLimit struct {
		Cost *int
	}
}

type branchData struct {
	Repository struct {
		Ref *branch `graphql:"ref(qualifiedName: $branchRefName)"`
	} `graphql:"repository(owner: $owner, name: $name)"`
}

type branchesHandler struct {
	ghClient         *github.Client
	graphClient      *githubv4.Client
	data             *defaultBranchData
	once             *sync.Once
	ctx              context.Context
	errSetup         error
	repourl          *repoURL
	defaultBranchRef *clients.BranchRef
}

func (handler *branchesHandler) init(ctx context.Context, repourl *repoURL) {
	handler.ctx = ctx
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
	handler.defaultBranchRef = nil
	handler.data = nil
}

func (handler *branchesHandler) setup() error {
	handler.once.Do(func() {
		if !strings.EqualFold(handler.repourl.commitSHA, clients.HeadSHA) {
			handler.errSetup = fmt.Errorf("%w: branches only supported for HEAD queries", clients.ErrUnsupportedFeature)
			return
		}
		vars := map[string]interface{}{
			"owner": githubv4.String(handler.repourl.owner),
			"name":  githubv4.String(handler.repourl.repo),
		}
		handler.data = new(defaultBranchData)
		if err := handler.graphClient.Query(handler.ctx, handler.data, vars); err != nil {
			handler.errSetup = sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("githubv4.Query: %v", err))
			return
		}
		handler.defaultBranchRef = getBranchRefFrom(handler.data.Repository.DefaultBranchRef)
	})
	return handler.errSetup
}

func (handler *branchesHandler) query(branchName string) (*clients.BranchRef, error) {
	if !strings.EqualFold(handler.repourl.commitSHA, clients.HeadSHA) {
		return nil, fmt.Errorf("%w: branches only supported for HEAD queries", clients.ErrUnsupportedFeature)
	}
	vars := map[string]interface{}{
		"owner":         githubv4.String(handler.repourl.owner),
		"name":          githubv4.String(handler.repourl.repo),
		"branchRefName": githubv4.String(refPrefix + branchName),
	}
	queryData := new(branchData)
	if err := handler.graphClient.Query(handler.ctx, queryData, vars); err != nil {
		return nil, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("githubv4.Query: %v", err))
	}
	return getBranchRefFrom(queryData.Repository.Ref), nil
}

func (handler *branchesHandler) getDefaultBranch() (*clients.BranchRef, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during branchesHandler.setup: %w", err)
	}
	return handler.defaultBranchRef, nil
}

func (handler *branchesHandler) getBranch(branch string) (*clients.BranchRef, error) {
	branchRef, err := handler.query(branch)
	if err != nil {
		return nil, fmt.Errorf("error during branchesHandler.query: %w", err)
	}
	return branchRef, nil
}

func copyAdminSettings(src *branchProtectionRule, dst *clients.BranchProtectionRule) {
	copyBoolPtr(src.IsAdminEnforced, &dst.EnforceAdmins)
	copyBoolPtr(src.RequireLastPushApproval, &dst.RequireLastPushApproval)
	copyBoolPtr(src.DismissesStaleReviews, &dst.RequiredPullRequestReviews.DismissStaleReviews)
	if src.RequiresStatusChecks != nil {
		copyBoolPtr(src.RequiresStatusChecks, &dst.CheckRules.RequiresStatusChecks)
		// TODO(#3255): Update when GitHub GraphQL bug is fixed
		// Workaround for GitHub GraphQL bug https://github.com/orgs/community/discussions/59471
		// The setting RequiresStrictStatusChecks should tell if the branch is required
		// to be up to date before merge, but it only returns the correct value if
		// RequiresStatusChecks is true. If RequiresStatusChecks is false, RequiresStrictStatusChecks
		// is wrongly retrieved as true.
		if src.RequiresStrictStatusChecks != nil {
			upToDateBeforeMerge := *src.RequiresStatusChecks && *src.RequiresStrictStatusChecks
			copyBoolPtr(&upToDateBeforeMerge, &dst.CheckRules.UpToDateBeforeMerge)
		}
	}
}

func copyNonAdminSettings(src interface{}, dst *clients.BranchProtectionRule) {
	// TODO: requiresConversationResolution, requiresSignatures, viewerAllowedToDismissReviews, viewerCanPush
	switch v := src.(type) {
	case *branchProtectionRule:
		copyBoolPtr(v.AllowsDeletions, &dst.AllowDeletions)
		copyBoolPtr(v.AllowsForcePushes, &dst.AllowForcePushes)
		copyBoolPtr(v.RequiresLinearHistory, &dst.RequireLinearHistory)
		copyInt32Ptr(v.RequiredApprovingReviewCount, &dst.RequiredPullRequestReviews.RequiredApprovingReviewCount)
		copyBoolPtr(v.RequiresCodeOwnerReviews, &dst.RequiredPullRequestReviews.RequireCodeOwnerReviews)
		copyStringSlice(v.RequiredStatusCheckContexts, &dst.CheckRules.Contexts)

	case *refUpdateRule:
		copyBoolPtr(v.AllowsDeletions, &dst.AllowDeletions)
		copyBoolPtr(v.AllowsForcePushes, &dst.AllowForcePushes)
		copyBoolPtr(v.RequiresLinearHistory, &dst.RequireLinearHistory)
		copyInt32Ptr(v.RequiredApprovingReviewCount, &dst.RequiredPullRequestReviews.RequiredApprovingReviewCount)
		copyBoolPtr(v.RequiresCodeOwnerReviews, &dst.RequiredPullRequestReviews.RequireCodeOwnerReviews)
		copyStringSlice(v.RequiredStatusCheckContexts, &dst.CheckRules.Contexts)
	}
}

func getBranchRefFrom(data *branch) *clients.BranchRef {
	if data == nil {
		return nil
	}
	branchRef := new(clients.BranchRef)
	if data.Name != nil {
		branchRef.Name = data.Name
	}

	// Protected means we found some data,
	// i.e., there's a rule for the branch.
	// It says nothing about what protection is enabled at all.
	branchRef.Protected = new(bool)
	if data.RefUpdateRule == nil &&
		data.BranchProtectionRule == nil {
		*branchRef.Protected = false
		return branchRef
	}

	*branchRef.Protected = true
	branchRule := &branchRef.BranchProtectionRule

	switch {
	// All settings are available. This typically means
	// scorecard is run with a token that has access
	// to admin settings.
	case data.BranchProtectionRule != nil:
		rule := data.BranchProtectionRule

		// Admin settings.
		copyAdminSettings(rule, branchRule)

		// Non-admin settings.
		copyNonAdminSettings(rule, branchRule)

	// Only non-admin settings are available.
	// https://docs.github.com/en/graphql/reference/objects#refupdaterule.
	case data.RefUpdateRule != nil:
		rule := data.RefUpdateRule
		copyNonAdminSettings(rule, branchRule)
	}

	return branchRef
}
