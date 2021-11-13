// Copyright 2021 Security Scorecard Authors
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
	"sync"

	"github.com/google/go-github/v38/github"
	"github.com/shurcooL/githubv4"

	"github.com/ossf/scorecard/v3/clients"
	sce "github.com/ossf/scorecard/v3/errors"
)

const (
	refsToAnalyze = 30
	refPrefix     = "refs/heads/"
)

// See https://github.community/t/graphql-api-protected-branch/14380
/* Example of query:
	query {
  repository(owner: "laurentsimon", name: "test3") {
    branchProtectionRules(first: 100) {
      nodes {
        pushAllowances (first:10){
          edges {
            node {
              id
            }
          }
        }
        allowsDeletions
        allowsForcePushes
        dismissesStaleReviews
        isAdminEnforced
        requiresApprovingReviews
        requiredApprovingReviewCount
        requiresStatusChecks
        requiresStrictStatusChecks
        restrictsPushes
        branchProtectionRuleConflicts(first:3) {
          __typename
          nodes {
            __typename
          }
        }
        pattern
        matchingRefs(first: 100) {
          nodes {
            name

          }
        }
      }
    }
	// I don't think `refs` is needed. The query qorks without it.
    refs(first: 100, refPrefix: "refs/heads/") {
      nodes {
        name
        __typename
      }
    }
  }
}
*/
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
	RequiredStatusCheckContexts  []string
	// TODO: verify there is no conflicts.
	// BranchProtectionRuleConflicts interface{}
}

type branch struct {
	Name                 *string
	BranchProtectionRule *branchProtectionRule
}

type branchesData struct {
	Repository struct {
		DefaultBranchRef branch
		Refs             struct {
			Nodes []branch
		} `graphql:"refs(first: $refsToAnalyze, refPrefix: $refPrefix)"`
	} `graphql:"repository(owner: $owner, name: $name)"`
}

type branchesHandler struct {
	ghClient         *github.Client
	graphClient      *githubv4.Client
	data             *branchesData
	once             *sync.Once
	ctx              context.Context
	errSetup         error
	owner            string
	repo             string
	defaultBranchRef *clients.BranchRef
	branches         []*clients.BranchRef
}

func (handler *branchesHandler) init(ctx context.Context, owner, repo string) {
	handler.ctx = ctx
	handler.owner = owner
	handler.repo = repo
	handler.errSetup = nil
	handler.once = new(sync.Once)
}

func (handler *branchesHandler) setup() error {
	handler.once.Do(func() {
		vars := map[string]interface{}{
			"owner":         githubv4.String(handler.owner),
			"name":          githubv4.String(handler.repo),
			"refsToAnalyze": githubv4.Int(refsToAnalyze),
			"refPrefix":     githubv4.String(refPrefix),
		}
		handler.data = new(branchesData)
		if err := handler.graphClient.Query(handler.ctx, handler.data, vars); err != nil {
			handler.errSetup = sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("githubv4.Query: %v", err))
		}
		handler.defaultBranchRef = getBranchRefFrom(handler.data.Repository.DefaultBranchRef)
		handler.branches = getBranchRefsFrom(handler.data.Repository.Refs.Nodes, handler.defaultBranchRef)
	})
	return handler.errSetup
}

func (handler *branchesHandler) getDefaultBranch() (*clients.BranchRef, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during branchesHandler.setup: %w", err)
	}
	return handler.defaultBranchRef, nil
}

func (handler *branchesHandler) listBranches() ([]*clients.BranchRef, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during branchesHandler.setup: %w", err)
	}
	return handler.branches, nil
}

func getBranchRefFrom(data branch) *clients.BranchRef {
	branchRef := new(clients.BranchRef)
	if data.Name != nil {
		branchRef.Name = data.Name
	}

	branchRef.Protected = new(bool)
	*branchRef.Protected = false

	branchRule := &branchRef.BranchProtectionRule
	if data.BranchProtectionRule != nil {
		*branchRef.Protected = true
		rule := data.BranchProtectionRule
		copyBoolPtr(rule.IsAdminEnforced, &branchRule.EnforceAdmins)
		copyBoolPtr(rule.DismissesStaleReviews, &branchRule.RequiredPullRequestReviews.DismissStaleReviews)
		if rule.RequiresStatusChecks != nil && *rule.RequiresStatusChecks {
			// fmt.Printf("%+v %v\n", *rule.RequiresStatusChecks, rule.RequiresStatusChecks)
			branchRule.RequiredStatusChecks = new(clients.StatusChecksRule)
			branchRule.RequiredStatusChecks.UpToDate = *rule.RequiresStrictStatusChecks
			copyStringSlice(rule.RequiredStatusCheckContexts, &branchRule.RequiredStatusChecks.Contexts)
		}

		copyBoolPtr(rule.AllowsDeletions, &branchRule.AllowDeletions)
		copyBoolPtr(rule.AllowsForcePushes, &branchRule.AllowForcePushes)
		copyBoolPtr(rule.RequiresLinearHistory, &branchRule.RequireLinearHistory)
		copyInt32Ptr(rule.RequiredApprovingReviewCount, &branchRule.RequiredPullRequestReviews.RequiredApprovingReviewCount)
		copyBoolPtr(rule.RequiresCodeOwnerReviews, &branchRule.RequiredPullRequestReviews.RequireCodeOwnerReviews)

	}

	return branchRef
}

func getBranchRefsFrom(data []branch, defaultBranch *clients.BranchRef) []*clients.BranchRef {
	var branchRefs []*clients.BranchRef
	var defaultFound bool
	for i, b := range data {
		branchRefs = append(branchRefs, getBranchRefFrom(b))
		if defaultBranch != nil && branchRefs[i].Name == defaultBranch.Name {
			defaultFound = true
		}
	}
	if !defaultFound {
		branchRefs = append(branchRefs, defaultBranch)
	}
	return branchRefs
}
