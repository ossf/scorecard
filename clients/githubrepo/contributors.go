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

func (handler *contributorsHandler) setup(fileReader io.ReadCloser) error {
	handler.once.Do(func() {
		defer fileReader.Close()
		if !strings.EqualFold(handler.repourl.commitSHA, clients.HeadSHA) {
			handler.errSetup = fmt.Errorf("%w: ListContributors only supported for HEAD queries", clients.ErrUnsupportedFeature)
			return
		}

		handler.contributors = append(handler.contributors, getContributors(handler)...)
		handler.contributors = append(handler.contributors, getCodeOwners(handler, fileReader)...)

		handler.errSetup = nil
	})
	return handler.errSetup
}

func getContributors(handler *contributorsHandler) []clients.User {
	contribs, _, err := handler.ghClient.Repositories.ListContributors(
		handler.ctx, handler.repourl.owner, handler.repourl.repo, &github.ListContributorsOptions{})
	if err != nil {
		handler.errSetup = fmt.Errorf("error during ListContributors: %w", err)
		return nil
	}

	contributorUsers := make([]clients.User, 0)
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
		contributorUsers = append(contributorUsers, contributor)
	}

	return contributorUsers
}

func getCodeOwners(handler *contributorsHandler, fileReader io.ReadCloser) []clients.User {
	ruleset, err := codeowners.ParseFile(fileReader)
	if err != nil {
		return nil
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

	var ownerUsers []clients.User
	for _, own := range owners {
		if own.GetLogin() == "" {
			continue
		}
		owner := clients.User{
			Login:       own.GetLogin(),
			IsCodeOwner: true,
		}

		// if verified external contributor add repo org to organization list by default
		if ok := slices.Contains(verifiedExternalOwners, owner.Login); ok {
			owner.Organizations = append(owner.Organizations, clients.User{
				Login: handler.repourl.owner,
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

		ownerUsers = append(ownerUsers, owner)
	}
	return ownerUsers
}

// expand team owners to multiple github users.
func getUsersFromTeam(handler *contributorsHandler, owner codeowners.Owner) []*github.User {
	users, response, err := handler.ghClient.Teams.ListTeamMembersBySlug(
		handler.ctx,
		handler.repourl.owner,
		owner.Value,
		&github.TeamListTeamMembersOptions{},
	)
	if err == nil && response.StatusCode == http.StatusOK {
		return users
	}
	return nil
}

// get github user from owner username.
func getUserFromUsername(handler *contributorsHandler, owner codeowners.Owner) []*github.User {
	user, response, err := handler.ghClient.Users.Get(handler.ctx, owner.Value)
	if err == nil && response.StatusCode == http.StatusOK {
		return []*github.User{user}
	}
	return nil
}

// get github user from email.
func getUserFromEmail(handler *contributorsHandler, owner codeowners.Owner) []*github.User {
	query := fmt.Sprintf("\"%s\" in:email", owner.String())
	userSearchResults, response, err := handler.ghClient.Search.Users(handler.ctx, query, &github.SearchOptions{})
	if err == nil && response.StatusCode == http.StatusOK && *userSearchResults.Total > 0 {
		return []*github.User{userSearchResults.Users[0]}
	}
	return nil
}

// getting verified external owners by @verified comment.
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

func (handler *contributorsHandler) getContributors(fileReader io.ReadCloser) ([]clients.User, error) {
	if err := handler.setup(fileReader); err != nil {
		return nil, fmt.Errorf("error during contributorsHandler.setup: %w", err)
	}
	return handler.contributors, nil
}
