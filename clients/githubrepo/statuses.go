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

	"github.com/google/go-github/v53/github"

	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

type statusesHandler struct {
	client  *github.Client
	ctx     context.Context
	repourl *repoURL
}

func (handler *statusesHandler) init(ctx context.Context, repourl *repoURL) {
	handler.ctx = ctx
	handler.repourl = repourl
}

func (handler *statusesHandler) listStatuses(ref string) ([]clients.Status, error) {
	statuses, _, err := handler.client.Repositories.ListStatuses(
		handler.ctx, handler.repourl.owner, handler.repourl.repo, ref, &github.ListOptions{})
	if err != nil {
		return nil, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("ListStatuses: %v", err))
	}
	return statusesFrom(statuses), nil
}

func statusesFrom(data []*github.RepoStatus) []clients.Status {
	var statuses []clients.Status
	for _, status := range data {
		statuses = append(statuses, clients.Status{
			State:     status.GetState(),
			Context:   status.GetContext(),
			URL:       status.GetURL(),
			TargetURL: status.GetTargetURL(),
		})
	}
	return statuses
}
