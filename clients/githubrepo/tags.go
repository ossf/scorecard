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
)

type tagsHandler struct {
	ghClient    *github.Client
	graphClient *githubv4.Client
	once        *sync.Once
	ctx         context.Context
	errSetup    error
	repourl     *Repo
	ruleSets    []*repoRuleSet
}

func (handler *tagsHandler) init(ctx context.Context, repourl *Repo) {
	handler.ctx = ctx
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
	handler.ruleSets = nil
}

func (handler *tagsHandler) setup() error {
	handler.once.Do(func() {
		// Tag protection is only supported for HEAD queries
		if handler.repourl.commitSHA != clients.HeadSHA {
			handler.errSetup = fmt.Errorf("%w: tag protection only supported for HEAD queries",
				clients.ErrUnsupportedFeature)
			return
		}

		// Fetch repository rulesets
		vars := map[string]interface{}{
			"owner": githubv4.String(handler.repourl.owner),
			"name":  githubv4.String(handler.repourl.repo),
		}
		rulesData := new(ruleSetData)
		if err := handler.graphClient.Query(
			handler.ctx,
			rulesData,
			vars,
		); err != nil {
			// Check if this is a permissions error
			if !isPermissionsError(err) {
				handler.errSetup = fmt.Errorf("error querying rulesets: %w", err)
				return
			}
			// Permission error - continue with empty rulesets
			// Tags will be reported as unprotected
			handler.ruleSets = nil
			return
		}
		handler.ruleSets = getActiveRuleSetsFrom(rulesData)
	})
	return handler.errSetup
}

func (handler *tagsHandler) getTag(tagName string) (*clients.TagRef, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during tagsHandler.setup: %w", err)
	}

	// Match rules for this tag
	refName := tagName
	rules, err := rulesMatchingRef(
		handler.ruleSets,
		refName,
		false,
		targetTag,
		refPrefixTag,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error matching rules for tag %s: %w",
			tagName,
			err,
		)
	}

	// Create tag ref and apply rules
	tagRef := &clients.TagRef{
		Name:      &tagName,
		Protected: new(bool),
	}
	applyRepoRulesToTag(tagRef, rules)

	// Return the completed tag reference
	return tagRef, nil
}
