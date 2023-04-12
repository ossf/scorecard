// Copyright 2022 OpenSSF Scorecard Authors
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

package gitlabrepo

import (
	"fmt"
	"strings"
	"sync"

	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
)

type contributorsHandler struct {
	glClient     *gitlab.Client
	once         *sync.Once
	errSetup     error
	repourl      *repoURL
	contributors []clients.User
}

func (handler *contributorsHandler) init(repourl *repoURL) {
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
}

func (handler *contributorsHandler) setup() error {
	handler.once.Do(func() {
		if !strings.EqualFold(handler.repourl.commitSHA, clients.HeadSHA) {
			handler.errSetup = fmt.Errorf("%w: ListContributors only supported for HEAD queries",
				clients.ErrUnsupportedFeature)
			return
		}
		order := "commits"
		sort := "desc"
		lco := &gitlab.ListContributorsOptions{
			ListOptions: gitlab.ListOptions{
				// PerPage: 100,
			},
			OrderBy: &order,
			Sort:    &sort,
		}
		contribs, _, err := handler.glClient.Repositories.Contributors(handler.repourl.project, lco)
		if err != nil {
			handler.errSetup = fmt.Errorf("error during ListContributors: %w", err)
			return
		}

		for _, contrib := range contribs {
			if contrib.Name == "" {
				continue
			}

			var user *gitlab.User
			queries := []string{contrib.Email, contrib.Name}
			for _, q := range queries {
				users, _, err := handler.glClient.Search.Users(q, &gitlab.SearchOptions{})
				if err != nil {
					handler.errSetup = fmt.Errorf("error during Search.Users: %w", err)
					return
				}
				if len(users) == 0 {
					continue
				}
				user = exactMatchUser(users, contrib)
				if user != nil {
					break
				}
			}

			var isBot bool
			var id int
			var companies []string
			if user == nil {
				continue
			}

			user, _, err := handler.glClient.Users.GetUser(user.ID, gitlab.GetUsersOptions{})
			if err != nil {
				handler.errSetup = fmt.Errorf("error during Users.Get: %w", err)
				return
			}
			isBot = user.Bot
			id = user.ID
			companies = []string{user.Organization}

			handler.contributors = append(handler.contributors, clients.User{
				Login:            contrib.Email,
				Companies:        companies,
				NumContributions: contrib.Commits,
				ID:               int64(id),
				IsBot:            isBot,
			})
		}
	})
	return handler.errSetup
}

func exactMatchUser(search []*gitlab.User, contrib *gitlab.Contributor) *gitlab.User {
	for i := range search {
		if search[i].Name == contrib.Name || search[i].PublicEmail == contrib.Email ||
			search[i].Email == contrib.Email {
			return search[i]
		}
	}

	return nil
}

func (handler *contributorsHandler) getContributors() ([]clients.User, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during contributorsHandler.setup: %w", err)
	}
	return handler.contributors, nil
}
