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

type refUpdateRule struct {
	AllowsDeletions              *bool
	AllowsForcePushes            *bool
	RequiredApprovingReviewCount *int32
	RequiresCodeOwnerReviews     *bool
	RequiresLinearHistory        *bool
	RequiredStatusCheckContexts  []string
}

type branchProtectionRule struct {
	DismissesStaleReviews      *bool
	IsAdminEnforced            *bool
	RequiresStrictStatusChecks *bool
}

type branch struct {
	Name                 *string
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

func copyBoolPtr(src *bool, dest **bool) {
	if src != nil {
		*dest = new(bool)
		**dest = *src
	}
}

func copyInt32Ptr(src *int32, dest **int32) {
	if src != nil {
		*dest = new(int32)
		**dest = *src
	}
}

func copyStringSlice(src []string, dest *[]string) {
	*dest = make([]string, len(src))
	copy(*dest, src)
}

func getBranchRefFrom(data branch) *clients.BranchRef {
	branchRef := new(clients.BranchRef)
	if data.Name != nil {
		branchRef.Name = data.Name
	}

	branchRef.Protected = new(bool)
	if data.RefUpdateRule == nil &&
		data.BranchProtectionRule == nil {
		*branchRef.Protected = false
		return branchRef
	}
	*branchRef.Protected = true

	branchRule := &branchRef.BranchProtectionRule
	if data.RefUpdateRule != nil {
		rule := data.RefUpdateRule
		copyBoolPtr(rule.AllowsDeletions, &branchRule.AllowDeletions)
		copyBoolPtr(rule.AllowsForcePushes, &branchRule.AllowForcePushes)
		copyBoolPtr(rule.RequiresLinearHistory, &branchRule.RequireLinearHistory)
		copyInt32Ptr(rule.RequiredApprovingReviewCount, &branchRule.RequiredPullRequestReviews.RequiredApprovingReviewCount)
		copyBoolPtr(rule.RequiresCodeOwnerReviews, &branchRule.RequiredPullRequestReviews.RequireCodeOwnerReviews)
		copyStringSlice(rule.RequiredStatusCheckContexts, &branchRule.RequiredStatusChecks.Contexts)
	}
	if data.BranchProtectionRule != nil {
		rule := data.BranchProtectionRule
		copyBoolPtr(rule.IsAdminEnforced, &branchRule.EnforceAdmins)
		copyBoolPtr(rule.DismissesStaleReviews, &branchRule.RequiredPullRequestReviews.DismissStaleReviews)
		copyBoolPtr(rule.RequiresStrictStatusChecks, &branchRule.RequiredStatusChecks.Strict)
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
