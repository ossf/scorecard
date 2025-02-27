// Copyright 2020 OpenSSF Scorecard Authors
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

package raw

import (
	"fmt"
	"strings"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
)

// Contributors retrieves the raw data for the Contributors check.
func Contributors(cr *checker.CheckRequest) (checker.ContributorsData, error) {
	c := cr.RepoClient
	var contributors []clients.User
	var owners []clients.User

	contribs, err := c.ListContributors()
	if err != nil {
		return checker.ContributorsData{}, fmt.Errorf("Client.Repositories.ListContributors: %w", err)
	}

	for _, contrib := range contribs {
		user := cleanCompaniesOrgs(&contrib)
		contributors = append(contributors, user)
	}

	// ignore error so we dont break other probes
	owns, _ := c.ListCodeOwners()
	for _, own := range owns {
		user := cleanCompaniesOrgs(&own)
		owners = append(owners, user)
	}

	// ignore error so we dont break other probes
	repoOwner, _ := c.RepoOwner()

	return checker.ContributorsData{Contributors: contributors, CodeOwners: owners, RepoOwner: repoOwner}, nil
}

func cleanCompaniesOrgs(user *clients.User) clients.User {
	cleanUser := clients.User{
		Login:            user.Login,
		NumContributions: user.NumContributions,
	}

	// removes duplicate orgs
	for _, org := range user.Organizations {
		if org.Login != "" && !orgContains(cleanUser.Organizations, org.Login) {
			cleanUser.Organizations = append(cleanUser.Organizations, org)
		}
	}

	// cleans up company names and removes duplicates
	for _, company := range user.Companies {
		if company == "" {
			continue
		}
		company = strings.ToLower(company)
		company = strings.ReplaceAll(company, "inc.", "")
		company = strings.ReplaceAll(company, "llc", "")
		company = strings.ReplaceAll(company, ",", "")
		company = strings.TrimLeft(company, "@")
		company = strings.Trim(company, " ")
		if company != "" && !companyContains(cleanUser.Companies, company) {
			cleanUser.Companies = append(cleanUser.Companies, company)
		}
	}

	return cleanUser
}

func companyContains(cs []string, name string) bool {
	for _, a := range cs {
		if a == name {
			return true
		}
	}
	return false
}

func orgContains(os []clients.User, login string) bool {
	for _, a := range os {
		if a.Login == login {
			return true
		}
	}
	return false
}
