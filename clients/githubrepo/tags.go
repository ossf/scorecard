// Copyright 2025 OpenSSF Scorecard Authors
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

	"github.com/google/go-github/v53/github"
	"github.com/shurcooL/githubv4"

	"github.com/ossf/scorecard/v5/clients"
	sce "github.com/ossf/scorecard/v5/errors"
)

// GraphQL query to load tag refs.
type tagsQuery struct {
	Repository struct {
		Refs struct {
			Nodes []*ref
		} `graphql:"refs(refPrefix: \"refs/tags/\", first: 100)"`
	} `graphql:"repository(owner: $owner, name: $name)"`
}

type tagsHandler struct {
	ctx         context.Context
	ghClient    *github.Client
	graphClient *githubv4.Client
	repourl     *Repo

	once     *sync.Once
	errSetup error

	ruleSets []*repoRuleSet
	tags     []*ref
}

func (h *tagsHandler) init(ctx context.Context, repourl *Repo) {
	h.ctx = ctx
	h.repourl = repourl
	h.errSetup = nil
	h.once = new(sync.Once)
	h.ruleSets = nil
}

func (h *tagsHandler) setup() error {
	h.once.Do(func() {
		vars := map[string]interface{}{
			"owner": githubv4.String(h.repourl.owner),
			"name":  githubv4.String(h.repourl.repo),
		}
		rulesData := new(ruleSetData)
		if err := h.graphClient.Query(h.ctx, rulesData, vars); err != nil {
			h.errSetup = sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("githubv4.Query: %v", err))
			return
		}
		h.ruleSets = getActiveRuleSetsFrom(rulesData)

		var q tagsQuery
		if err := h.graphClient.Query(h.ctx, &q, vars); err != nil {
			h.errSetup = sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("githubv4.Query tags: %v", err))
			return
		}
		h.tags = q.Repository.Refs.Nodes
	})
	return h.errSetup
}

func (h *tagsHandler) query(tagName string) (*clients.RepoRef, error) {
	if err := h.setup(); err != nil {
		return nil, fmt.Errorf("error during branchesHandler.setup: %w", err)
	}
	rules, err := rulesMatchingBranch(h.ruleSets, tagName, false, "TAG")
	if err != nil {
		return nil, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("rulesMatchingBranch: %v", err))
	}

	for _, r := range h.tags {
		if *r.Name == tagName {
			return getBranchRefFrom(r, rules), nil
		}
	}
	return nil, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("tag not found: %s", tagName))
}

func (h *tagsHandler) getTags() ([]*clients.RepoRef, error) {
	if err := h.setup(); err != nil {
		return nil, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("error during tagsHandler.setup: %v", err))
	}
	var out []*clients.RepoRef
	for _, tag := range h.tags {
		tag, err := h.getTag(*tag.Name)
		if err != nil {
			return nil, fmt.Errorf("getTag: %w", err)
		}
		out = append(out, tag)
	}
	return out, nil
}

func (h *tagsHandler) getTag(tag string) (*clients.RepoRef, error) {
	branchRef, err := h.query(tag)
	if err != nil {
		return nil, fmt.Errorf("error during branchesHandler.query: %w", err)
	}
	return branchRef, nil
}
