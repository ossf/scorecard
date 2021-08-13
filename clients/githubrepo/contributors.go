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

	"github.com/google/go-github/v38/github"

	"github.com/ossf/scorecard/v2/clients"
)

type contributorsHandler struct {
	ghClient     *github.Client
	contributors []clients.Contributor
}

func (handler *contributorsHandler) init(ctx context.Context, owner, repo string) error {
	contribs, _, err := handler.ghClient.Repositories.ListContributors(ctx, owner, repo, &github.ListContributorsOptions{})
	if err != nil {
		return fmt.Errorf("error during ListContributors: %w", err)
	}

	for _, contrib := range contribs {
		if contrib.GetLogin() == "" {
			continue
		}
		contributor := clients.Contributor{
			NumContributions: contrib.GetContributions(),
			User: clients.User{
				Login: contrib.GetLogin(),
			},
		}
		orgs, _, err := handler.ghClient.Organizations.List(ctx, contrib.GetLogin(), nil)
		// This call can fail due to token scopes. So ignore error.
		if err == nil {
			for _, org := range orgs {
				contributor.Organizations = append(contributor.Organizations, clients.User{
					Login: org.GetLogin(),
				})
			}
		}
		user, _, err := handler.ghClient.Users.Get(ctx, contrib.GetLogin())
		if err != nil {
			return fmt.Errorf("error during Users.Get: %w", err)
		}
		contributor.Company = user.GetCompany()
		handler.contributors = append(handler.contributors, contributor)
	}
	return nil
}

func (handler *contributorsHandler) getContributors() ([]clients.Contributor, error) {
	return handler.contributors, nil
}
