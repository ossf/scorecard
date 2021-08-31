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

	"github.com/ossf/scorecard/v2/clients"
	sce "github.com/ossf/scorecard/v2/errors"
)

const (
	refsToAnalyze = 30
	refPrefix     = "refs/heads/"
)

type refUpdateRule struct {
	AllowsDeletions              *githubv4.Boolean
	AllowsForcePushes            *githubv4.Boolean
	RequiredApprovingReviewCount *githubv4.Int
	RequiresCodeOwnerReviews     *githubv4.Boolean
	RequiresLinearHistory        *githubv4.Boolean
	RequiredStatusCheckContexts  []githubv4.String
}

type branchProtectionRule struct {
	DismissesStaleReviews      *githubv4.Boolean
	IsAdminEnforced            *githubv4.Boolean
	RequiresStrictStatusChecks *githubv4.Boolean
}

type branch struct {
	Name                 *githubv4.String
	RefUpdateRule        *refUpdateRule
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
			handler.errSetup = sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("githubv4.Query: %v", err))
		}
		handler.defaultBranchRef = getBranchRefFrom(handler.data.Repository.DefaultBranchRef)
		handler.branches = getBranchRefsFrom(handler.data.Repository.Refs.Nodes)
		// Maybe add defaultBranchRef to branches.
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

func setPullRequestReviewRule(c *clients.BranchProtectionRule) *clients.PullRequestReviewRule {
	if c.GetRequiredPullRequestReviews() == nil {
		c.RequiredPullRequestReviews = new(clients.PullRequestReviewRule)
	}
	return c.GetRequiredPullRequestReviews()
}

func setStatusChecksRule(c *clients.BranchProtectionRule) *clients.StatusChecksRule {
	if c.GetRequiredStatusChecks() == nil {
		c.RequiredStatusChecks = new(clients.StatusChecksRule)
	}
	return c.GetRequiredStatusChecks()
}

func getBranchRefFrom(data branch) *clients.BranchRef {
	branchRef := new(clients.BranchRef)
	if data.Name != nil {
		branchRef.Name = string(*data.Name)
	}

	if data.RefUpdateRule == nil &&
		data.BranchProtectionRule == nil {
		branchRef.Protected = false
		return branchRef
	}
	branchRef.Protected = true

	branchRef.BranchProtectionRule = new(clients.BranchProtectionRule)
	branchRule := branchRef.GetBranchProtectionRule()
	// nolint: nestif
	if data.RefUpdateRule != nil {
		rule := data.RefUpdateRule
		if rule.AllowsDeletions != nil {
			branchRule.AllowDeletions = &clients.AllowDeletions{
				Enabled: bool(*rule.AllowsDeletions),
			}
		}
		if rule.AllowsForcePushes != nil {
			branchRule.AllowForcePushes = &clients.AllowForcePushes{
				Enabled: bool(*rule.AllowsForcePushes),
			}
		}
		if rule.RequiresLinearHistory != nil {
			branchRule.RequireLinearHistory = &clients.RequireLinearHistory{
				Enabled: bool(*rule.RequiresLinearHistory),
			}
		}
		if rule.RequiredApprovingReviewCount != nil {
			setPullRequestReviewRule(branchRule).RequiredApprovingReviewCount = int32(*rule.RequiredApprovingReviewCount)
		}
		if rule.RequiresCodeOwnerReviews != nil {
			setPullRequestReviewRule(branchRule).RequireCodeOwnerReviews = bool(*rule.RequiresCodeOwnerReviews)
		}
		if rule.RequiredStatusCheckContexts != nil {
			var contexts []string
			for _, context := range rule.RequiredStatusCheckContexts {
				contexts = append(contexts, string(context))
			}
			setStatusChecksRule(branchRule).Contexts = contexts
		}
	}

	if data.BranchProtectionRule != nil {
		rule := data.BranchProtectionRule
		if rule.IsAdminEnforced != nil {
			branchRule.EnforceAdmins = &clients.EnforceAdmins{
				Enabled: bool(*rule.IsAdminEnforced),
			}
		}
		if rule.DismissesStaleReviews != nil {
			setPullRequestReviewRule(branchRule).DismissStaleReviews = bool(*rule.DismissesStaleReviews)
		}
		if rule.RequiresStrictStatusChecks != nil {
			setStatusChecksRule(branchRule).Strict = bool(*rule.RequiresStrictStatusChecks)
		}
	}
	return branchRef
}

func getBranchRefsFrom(data []branch) []*clients.BranchRef {
	branchRefs := make([]*clients.BranchRef, len(data))
	for i, b := range data {
		branchRefs[i] = getBranchRefFrom(b)
	}
	return branchRefs
}
