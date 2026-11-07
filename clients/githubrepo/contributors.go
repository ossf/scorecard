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
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/google/go-github/v82/github"
	"github.com/hmarr/codeowners"

	"github.com/ossf/scorecard/v5/clients"
)

const codeownersLimit = 100

// these are the paths where CODEOWNERS files can be found see
// https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners#codeowners-file-location
//
//nolint:lll
var (
	codeOwnerPaths []string = []string{"CODEOWNERS", ".github/CODEOWNERS", "docs/CODEOWNERS"}
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
	if codeOwnerFile != nil && codeOwnerFile != (*os.File)(nil) {
		defer codeOwnerFile.Close()
	}
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

		for _, contributor := range contributors {
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
	if codeOwnerFile == nil || codeOwnerFile == (*os.File)(nil) {
		return
	}
	ruleset, err := codeowners.ParseFile(codeOwnerFile)
	if err != nil {
		log.Printf("unable to parse CODEOWNERS file: %v. Skipping...", err)
		return
	}

	// adding owners to contributor map and deduping
	for _, login := range handler.codeOwnerLogins(ruleset) {
		if login == "" {
			continue
		}
		value, ok := contributors[login]
		if !ok {
			value = clients.User{Login: login}
		}
		value.IsCodeOwner = true
		contributors[login] = value
	}
}

// codeOwnerLogins lists the owners in ruleset in order, with duplicates,
// expanding each team to its members. It stops at codeownersLimit, which
// also bounds the team lookups.
func (handler *contributorsHandler) codeOwnerLogins(ruleset codeowners.Ruleset) []string {
	var logins []string
	for _, rule := range ruleset {
		for _, owner := range rule.Owners {
			switch owner.Type {
			case codeowners.UsernameOwner:
				logins = append(logins, owner.Value)
			case codeowners.TeamOwner:
				logins = append(logins, handler.teamMembers(owner.Value)...)
			}
			if len(logins) >= codeownersLimit {
				return logins[:codeownersLimit]
			}
		}
	}
	return logins
}

// teamMembers returns the logins of the members of team, written as
// "org/team". Listing them needs the read:org token scope, so a failed
// lookup returns no members.
func (handler *contributorsHandler) teamMembers(team string) []string {
	splitTeam := strings.Split(team, "/")
	if len(splitTeam) != 2 {
		return nil
	}
	users, response, err := handler.ghClient.Teams.ListTeamMembersBySlug(
		handler.ctx,
		splitTeam[0],
		splitTeam[1],
		&github.TeamListTeamMembersOptions{},
	)
	if err != nil || response.StatusCode != http.StatusOK {
		return nil
	}
	logins := make([]string, 0, len(users))
	for _, user := range users {
		logins = append(logins, user.GetLogin())
	}
	return logins
}

func (handler *contributorsHandler) getContributors(codeOwnerFile io.ReadCloser) ([]clients.User, error) {
	if err := handler.setup(codeOwnerFile); err != nil {
		return nil, fmt.Errorf("error during contributorsHandler.setup: %w", err)
	}
	return handler.contributors, nil
}
