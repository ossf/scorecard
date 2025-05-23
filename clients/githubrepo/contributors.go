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
	"io"
	"maps"
	"net/http"
	"strings"
	"sync"

	"github.com/google/go-github/v53/github"
	"github.com/hmarr/codeowners"

	"github.com/ossf/scorecard/v5/clients"
)

// these are the paths where CODEOWNERS files can be found see
// https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners#codeowners-file-location
//
//nolint:lll
var (
	CodeOwnerPaths []string = []string{"CODEOWNERS", ".github/CODEOWNERS", "docs/CODEOWNERS"}
)

type contributorsHandler struct {
	ghClient     *github.Client
	once         *sync.Once
	ctx          context.Context
	errSetup     error
	repourl      *Repo
	contributors []clients.User
}

func (handler *contributorsHandler) init(ctx context.Context, repourl *Repo) {
	handler.ctx = ctx
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
	handler.contributors = nil
}

func (handler *contributorsHandler) setup(codeOwnerFile io.ReadCloser) error {
	defer codeOwnerFile.Close()
	handler.once.Do(func() {
		if !strings.EqualFold(handler.repourl.commitSHA, clients.HeadSHA) {
			handler.errSetup = fmt.Errorf("%w: ListContributors only supported for HEAD queries", clients.ErrUnsupportedFeature)
			return
		}

		contributors := make(map[string]clients.User)
		mapContributors(handler, contributors)
		if handler.errSetup != nil {
			return
		}
		mapCodeOwners(handler, codeOwnerFile, contributors)
		if handler.errSetup != nil {
			return
		}

		for contributor := range maps.Values(contributors) {
			orgs, _, err := handler.ghClient.Organizations.List(handler.ctx, contributor.Login, nil)
			// This call can fail due to token scopes. So ignore error.
			if err == nil {
				for _, org := range orgs {
					contributor.Organizations = append(contributor.Organizations, clients.User{
						Login: org.GetLogin(),
					})
				}
			}
			user, _, err := handler.ghClient.Users.Get(handler.ctx, contributor.Login)
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

func mapContributors(handler *contributorsHandler, contributors map[string]clients.User) {
	// getting contributors from the github API
	contribs, _, err := handler.ghClient.Repositories.ListContributors(
		handler.ctx, handler.repourl.owner, handler.repourl.repo, &github.ListContributorsOptions{})
	if err != nil {
		handler.errSetup = fmt.Errorf("error during ListContributors: %w", err)
		return
	}

	// adding contributors to contributor map
	for _, contrib := range contribs {
		if contrib.GetLogin() == "" {
			continue
		}
		contributors[contrib.GetLogin()] = clients.User{
			Login: contrib.GetLogin(), NumContributions: contrib.GetContributions(),
			IsCodeOwner: false,
		}
	}
}

func mapCodeOwners(handler *contributorsHandler, codeOwnerFile io.ReadCloser, contributors map[string]clients.User) {
	ruleset, err := codeowners.ParseFile(codeOwnerFile)
	if err != nil {
		handler.errSetup = fmt.Errorf("error during ParseFile: %w", err)
		return
	}

	// expanding owners
	owners := make([]*clients.User, 0)
	for _, rule := range ruleset {
		for _, owner := range rule.Owners {
			switch owner.Type {
			case codeowners.UsernameOwner:
				// if usernameOwner just add to owners list
				owners = append(owners, &clients.User{Login: owner.Value, NumContributions: 0, IsCodeOwner: true})
			case codeowners.TeamOwner:
				// if teamOwner expand and add to owners list (only accessible by org members with read:org token scope)
				splitTeam := strings.Split(owner.Value, "/")
				if len(splitTeam) == 2 {
					users, response, err := handler.ghClient.Teams.ListTeamMembersBySlug(
						handler.ctx,
						splitTeam[0],
						splitTeam[1],
						&github.TeamListTeamMembersOptions{},
					)
					if err == nil && response.StatusCode == http.StatusOK {
						for _, user := range users {
							owners = append(owners, &clients.User{Login: user.GetLogin(), NumContributions: 0, IsCodeOwner: true})
						}
					}
				}
			}
		}
	}

	// adding owners to contributor map and deduping
	for _, owner := range owners {
		if owner.Login == "" {
			continue
		}
		value, ok := contributors[owner.Login]
		if ok {
			// if contributor exists already set IsCodeOwner to true
			value.IsCodeOwner = true
			contributors[owner.Login] = value
		} else {
			// otherwise add new contributor
			contributors[owner.Login] = *owner
		}
	}
}

func (handler *contributorsHandler) getContributors(codeOwnerFile io.ReadCloser) ([]clients.User, error) {
	if err := handler.setup(codeOwnerFile); err != nil {
		return nil, fmt.Errorf("error during contributorsHandler.setup: %w", err)
	}
	return handler.contributors, nil
}
