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
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"slices"
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

type ownersHandler struct {
	ghClient *github.Client
	once     *sync.Once
	ctx      context.Context
	errSetup error
	repourl  *Repo
	owners   []clients.User
}

func (handler *ownersHandler) init(ctx context.Context, repourl *Repo) {
	handler.ctx = ctx
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
	handler.owners = nil
}

func (handler *ownersHandler) setup(fileReader io.ReadCloser, repoOwner string) error {
	handler.once.Do(func() {
		defer fileReader.Close()

		if !strings.EqualFold(handler.repourl.commitSHA, clients.HeadSHA) {
			handler.errSetup = fmt.Errorf("%w: ListCodeOwners only supported for HEAD queries", clients.ErrUnsupportedFeature)
			return
		}

		ruleset, err := codeowners.ParseFile(fileReader)
		if err != nil {
			handler.errSetup = fmt.Errorf("failed to parse owners file: %w", err)
			return
		}

		verifiedExternalOwners := getVerifiedExternalOwners(fileReader)

		// expanding owners
		owners := make([]*github.User, 0)
		for _, rule := range ruleset {
			for _, owner := range rule.Owners {
				var users []*github.User
				switch owner.Type {
				case codeowners.UsernameOwner:
					users = getUserFromUsername(handler, owner)
				case codeowners.TeamOwner:
					users = getUsersFromTeam(handler, owner)
				case codeowners.EmailOwner:
					users = getUserFromEmail(handler, owner)
				}
				if users != nil {
					owners = append(owners, users...)
				}
			}
		}

		for _, own := range owners {
			owner := clients.User{
				Login: own.GetLogin(),
			}

			// if verified external contributor add repo org to organization list by default
			if ok := slices.Contains(verifiedExternalOwners, owner.Login); ok {
				owner.Organizations = append(owner.Organizations, clients.User{
					Login: repoOwner,
				})
			}

			orgs, _, err := handler.ghClient.Organizations.List(handler.ctx, own.GetLogin(), nil)
			// This call can fail due to token scopes. So ignore error.
			if err == nil {
				for _, org := range orgs {
					owner.Organizations = append(owner.Organizations, clients.User{
						Login: org.GetLogin(),
					})
				}
			}
			owner.Companies = append(owner.Companies, own.GetCompany())

			handler.owners = append(handler.owners, owner)
		}

		handler.errSetup = nil
	})
	return handler.errSetup
}

// get github user from owner username
func getUserFromUsername(handler *ownersHandler, owner codeowners.Owner) []*github.User {
	emptyUser := clients.User{
		Login: owner.Value,
	}
	user, response, err := handler.ghClient.Users.Get(handler.ctx, owner.Value)
	if err == nil && response.StatusCode == http.StatusOK {
		return []*github.User{user}
	}
	// append to owners if cant find the github user
	handler.owners = append(handler.owners, emptyUser)
	return nil
}

// expand team owners to multiple github users
func getUsersFromTeam(handler *ownersHandler, owner codeowners.Owner) []*github.User {
	emptyUser := clients.User{
		Login: owner.Value,
	}
	users, response, err := handler.ghClient.Teams.ListTeamMembersBySlug(
		handler.ctx,
		handler.repourl.owner,
		owner.Value,
		&github.TeamListTeamMembersOptions{},
	)
	if err == nil && response.StatusCode == http.StatusOK {
		return users
	}
	// append to owners if cant find the github team
	handler.owners = append(handler.owners, emptyUser)
	return nil
}

// get github user from email
func getUserFromEmail(handler *ownersHandler, owner codeowners.Owner) []*github.User {
	emptyUser := clients.User{
		Login: owner.Value,
	}
	query := fmt.Sprintf("\"%s\" in:email", owner.String())
	userSearchResults, response, err := handler.ghClient.Search.Users(handler.ctx, query, &github.SearchOptions{})
	if err == nil && response.StatusCode == http.StatusOK && *userSearchResults.Total > 0 {
		return []*github.User{userSearchResults.Users[0]}
	}
	// append to owners if cant find the github user
	handler.owners = append(handler.owners, emptyUser)
	return nil
}

// getting verified external owners by @verified comment
func getVerifiedExternalOwners(fileReader io.ReadCloser) []string {
	verifiedExternalOwners := make([]string, 0)
	r := regexp.MustCompile("^# @verified .*")
	scanner := bufio.NewScanner(fileReader)
	for scanner.Scan() {
		line := scanner.Text()
		match := r.MatchString(line)
		if match {
			usernames := strings.Fields(line)[2:]
			verifiedExternalOwners = append(verifiedExternalOwners, usernames...)
		}
	}
	return verifiedExternalOwners
}

func (handler *ownersHandler) getOwners(fileReader io.ReadCloser, repoOwner string) ([]clients.User, error) {
	if err := handler.setup(fileReader, repoOwner); err != nil {
		return nil, fmt.Errorf("error during ownersHandler.setup: %w", err)
	}
	return handler.owners, nil
}
