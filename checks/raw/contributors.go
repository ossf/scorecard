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

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
)

// Contributors retrieves the raw data for the Contributors check.
func Contributors(c clients.RepoClient) (checker.ContributorsData, error) {
	var users []clients.User

	contribs, err := c.ListContributors()
	if err != nil {
		return checker.ContributorsData{}, fmt.Errorf("Client.Repositories.ListContributors: %w", err)
	}

	for _, contrib := range contribs {
		user := clients.User{
			Login:            contrib.Login,
			NumContributions: contrib.NumContributions,
		}

		for _, org := range contrib.Organizations {
			if org.Login != "" && !orgContains(user.Organizations, org.Login) {
				user.Organizations = append(user.Organizations, org)
			}
		}

		for _, company := range contrib.Companies {
			if company == "" {
				continue
			}
			company = strings.ToLower(company)
			company = strings.ReplaceAll(company, "inc.", "")
			company = strings.ReplaceAll(company, "llc", "")
			company = strings.ReplaceAll(company, ",", "")
			company = strings.TrimLeft(company, "@")
			company = strings.Trim(company, " ")
			if company != "" && !companyContains(user.Companies, company) {
				user.Companies = append(user.Companies, company)
			}
		}

		users = append(users, user)
	}

	return checker.ContributorsData{Users: users}, nil
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
