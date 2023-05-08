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

		var contribs []*gitlab.Contributor
		i := 1

		for {
			c, _, err := handler.glClient.Repositories.Contributors(
				handler.repourl.project,
				&gitlab.ListContributorsOptions{
					ListOptions: gitlab.ListOptions{
						Page:    i,
						PerPage: 100,
					},
				},
			)
			if err != nil {
				handler.errSetup = fmt.Errorf("error during ListContributors: %w", err)
				return
			}

			if len(c) == 0 {
				break
			}
			i++
			contribs = append(contribs, c...)
		}

		for _, contrib := range contribs {
			if contrib.Name == "" {
				continue
			}

			users, _, err := handler.glClient.Search.Users(contrib.Name, &gitlab.SearchOptions{})
			if err != nil {
				handler.errSetup = fmt.Errorf("error during Users.Get: %w", err)
				return
			} else if len(users) == 0 {
				// parseEmailToName is declared in commits.go
				users, _, err = handler.glClient.Search.Users(parseEmailToName(contrib.Email), &gitlab.SearchOptions{})
				if err != nil {
					handler.errSetup = fmt.Errorf("error during Users.Get: %w", err)
					return
				}
			}

			user := &gitlab.User{}

			if len(users) == 0 {
				user.ID = 0
				user.Organization = ""
				user.Bot = true
			} else {
				user = users[0]
			}

			contributor := clients.User{
				Login:            contrib.Email,
				Companies:        []string{user.Organization},
				NumContributions: contrib.Commits,
				ID:               int64(user.ID),
				IsBot:            user.Bot,
			}
			handler.contributors = append(handler.contributors, contributor)
		}
	})
	return handler.errSetup
}

func (handler *contributorsHandler) getContributors() ([]clients.User, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during contributorsHandler.setup: %w", err)
	}
	return handler.contributors, nil
}
