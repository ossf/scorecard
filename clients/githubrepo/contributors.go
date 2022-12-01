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

	"github.com/google/go-github/v38/github"

	"github.com/ossf/scorecard/v4/clients"
)

type contributorsHandler struct {
	ghClient     *github.Client
	once         *sync.Once
	ctx          context.Context
	errSetup     error
	repourl      *repoURL
	contributors []clients.User
}

func (handler *contributorsHandler) init(ctx context.Context, repourl *repoURL) {
	handler.ctx = ctx
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
}

func (handler *contributorsHandler) setup() error {
	handler.once.Do(func() {
		if !strings.EqualFold(handler.repourl.commitSHA, clients.HeadSHA) {
			handler.errSetup = fmt.Errorf("%w: ListContributors only supported for HEAD queries", clients.ErrUnsupportedFeature)
			return
		}
		contribs, _, err := handler.ghClient.Repositories.ListContributors(
			handler.ctx, handler.repourl.owner, handler.repourl.repo, &github.ListContributorsOptions{})
		if err != nil {
			handler.errSetup = fmt.Errorf("error during ListContributors: %w", err)
			return
		}

		for _, contrib := range contribs {
			if contrib.GetLogin() == "" {
				continue
			}
			contributor := clients.User{
				NumContributions: contrib.GetContributions(),
				Login:            contrib.GetLogin(),
			}
			orgs, _, err := handler.ghClient.Organizations.List(handler.ctx, contrib.GetLogin(), nil)
			// This call can fail due to token scopes. So ignore error.
			if err == nil {
				for _, org := range orgs {
					contributor.Organizations = append(contributor.Organizations, clients.User{
						Login: org.GetLogin(),
					})
				}
			}
			user, _, err := handler.ghClient.Users.Get(handler.ctx, contrib.GetLogin())
			if err != nil {
				handler.errSetup = fmt.Errorf("error during Users.Get: %w", err)
			}
			contributor.Companies = append(contributor.Companies, user.GetCompany())
			handler.contributors = append(handler.contributors, contributor)
		}
		handler.errSetup = nil
	})
	return handler.errSetup
}

func (handler *contributorsHandler) getContributors() ([]clients.User, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during contributorsHandler.setup: %w", err)
	}
	return handler.contributors, nil
}
