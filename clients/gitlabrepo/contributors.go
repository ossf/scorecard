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

	"github.com/ossf/scorecard/v5/clients"
)

type contributorsHandler struct {
	fnContributors retrieveContributorFn
	fnUsers        retrieveUserFn
	glClient       *gitlab.Client
	once           *sync.Once
	errSetup       error
	repourl        *Repo
	contributors   []clients.User
}

func (handler *contributorsHandler) init(repourl *Repo) {
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
	handler.fnContributors = handler.retrieveContributors
	handler.fnUsers = handler.retrieveUsers
}

type (
	retrieveContributorFn func(string) ([]*gitlab.Contributor, error)
	retrieveUserFn        func(string) ([]*gitlab.User, error)
)

func (handler *contributorsHandler) retrieveContributors(project string) ([]*gitlab.Contributor, error) {
	var contribs []*gitlab.Contributor
	i := 1

	for {
		c, _, err := handler.glClient.Repositories.Contributors(
			project,
			&gitlab.ListContributorsOptions{
				ListOptions: gitlab.ListOptions{
					Page:    i,
					PerPage: 100,
				},
			},
		)
		if err != nil {
			//nolint:wrapcheck
			return nil, err
		}

		if len(c) == 0 {
			break
		}
		i++
		contribs = append(contribs, c...)
	}
	return contribs, nil
}

func (handler *contributorsHandler) retrieveUsers(queryName string) ([]*gitlab.User, error) {
	users, _, err := handler.glClient.Search.Users(queryName, &gitlab.SearchOptions{})
	if err != nil {
		//nolint:wrapcheck
		return nil, err
	}
	return users, nil
}

// Orginisation is an experimental feature in GitLab. Takes a slice
// pointer to maintain list of domains and email returns an anonymised
// string with the slice index as a unique identifier.
// Pointer is required as we are modifying the slice length.
func getOrginisation(orgDomain *[]string, orgReal string) string {
	// Some repositories date before email was widely used.
	emailSplit := strings.Split(orgReal, "@")
	var orgName string // Orginisation name
	if len(emailSplit) > 1 {
		orgName = emailSplit[1]
	} else {
		orgName = emailSplit[0] // Not an "email"
	}

	*orgDomain = append(*orgDomain, orgName)

	// Search for domain and use index as unique marker
	for i, domain := range *orgDomain {
		if domain == orgName {
			return fmt.Sprint("GitLab-", i)
		}
	}
	return "GitLab"
}

func (handler *contributorsHandler) setup() error {
	var orgDomain []string // Slice of email domains

	handler.once.Do(func() {
		if !strings.EqualFold(handler.repourl.commitSHA, clients.HeadSHA) {
			handler.errSetup = fmt.Errorf("%w: ListContributors only supported for HEAD queries",
				clients.ErrUnsupportedFeature)
			return
		}

		contribs, err := handler.fnContributors(handler.repourl.projectID)
		if err != nil {
			handler.errSetup = fmt.Errorf("error during ListContributors: %w", err)
			return
		}

		for _, contrib := range contribs {
			if contrib.Name == "" {
				continue
			}

			users, err := handler.fnUsers(contrib.Name)
			if err != nil {
				handler.errSetup = fmt.Errorf("error during Users.Get: %w", err)
				return
			} else if len(users) == 0 && contrib.Email != "" {
				// parseEmailToName is declared in commits.go
				users, err = handler.fnUsers(parseEmailToName(contrib.Email))
				if err != nil {
					handler.errSetup = fmt.Errorf("error during Users.Get: %w", err)
					return
				}
			}

			user := &gitlab.User{}

			if len(users) == 0 {
				user.ID = 0
				user.Organization = ""
				user.Bot = false
			} else {
				user = users[0]
			}

			// In case someone is using the experimental feature.
			var orgName string
			if user.Organization == "" {
				orgName = getOrginisation(&orgDomain, contrib.Email)
			} else {
				orgName = user.Organization
			}

			contributor := clients.User{
				Login:            contrib.Email,
				Companies:        []string{orgName},
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
