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

	"github.com/google/go-github/v38/github"

	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

type checkrunsHandler struct {
	client  *github.Client
	ctx     context.Context
	repourl *repoURL
}

func (handler *checkrunsHandler) init(ctx context.Context, repourl *repoURL) {
	handler.ctx = ctx
	handler.repourl = repourl
}

func (handler *checkrunsHandler) listCheckRunsForRef(ref string) ([]clients.CheckRun, error) {
	checkRuns, _, err := handler.client.Checks.ListCheckRunsForRef(
		handler.ctx, handler.repourl.owner, handler.repourl.repo, ref, &github.ListCheckRunsOptions{})
	if err != nil {
		return nil, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("ListCheckRunsForRef: %v", err))
	}
	return checkRunsFrom(checkRuns), nil
}

func checkRunsFrom(data *github.ListCheckRunsResults) []clients.CheckRun {
	var checkRuns []clients.CheckRun
	for _, checkRun := range data.CheckRuns {
		checkRuns = append(checkRuns, clients.CheckRun{
			Status:     checkRun.GetStatus(),
			Conclusion: checkRun.GetConclusion(),
			URL:        checkRun.GetURL(),
			App: clients.CheckRunApp{
				Slug: checkRun.GetApp().GetSlug(),
			},
		})
	}
	return checkRuns
}
