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
		contribs, _, err := handler.glClient.Repositories.Contributors(
			handler.repourl.projectID, &gitlab.ListContributorsOptions{})
		if err != nil {
			handler.errSetup = fmt.Errorf("error during ListContributors: %w", err)
			return
		}

		for _, contrib := range contribs {
			if contrib.Name == "" {
				continue
			}

			// In Gitlab users only have one registered organization which is the company they work for, this means that
			// the organizations field will not be filled in and the companies field will be a singular value.
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

			contributor := clients.User{
				Login:            contrib.Email,
				Companies:        []string{users[0].Organization},
				NumContributions: contrib.Commits,
				ID:               int64(users[0].ID),
				IsBot:            users[0].Bot,
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
