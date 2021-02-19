// Copyright 2020 Security Scorecard Authors
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

package checks

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/ossf/scorecard/checker"
)

func init() {
	registerCheck("Contributors", Contributors)
}

func Contributors(c checker.Checker) checker.CheckResult {
	type orgnaizations []struct {
		Login string `json:"login"`
	}
	contribs, _, err := c.Client.Repositories.ListContributors(c.Ctx, c.Owner, c.Repo, &github.ListContributorsOptions{})
	if err != nil {
		return checker.RetryResult(err)
	}

	companies := map[string]struct{}{}
	for _, contrib := range contribs {
		const contributorsCount = 5
		//nolint:nestif
		if contrib.GetContributions() >= contributorsCount {
			u, _, err := c.Client.Users.Get(c.Ctx, contrib.GetLogin())
			if err != nil {
				return checker.RetryResult(err)
			}
			// TODO - Figure out the real API for getting user orgs
			url := fmt.Sprintf("https://api.github.com/users/%s/orgs", contrib.GetLogin())
			resp, e := c.HttpClient.Get(url)

			if e != nil {
				c.Logf("unable to get org members for %s", contrib.GetLogin())
			} else {
				var orgs orgnaizations

				err = json.NewDecoder(resp.Body).Decode(&orgs)

				if err != nil {
					c.Logf("unable to get  decode json for for %s org", contrib.GetLogin())
				} else {
					if len(orgs) > 0 {
						companies[orgs[0].Login] = struct{}{}
						continue
					}
				}
			}

			defer resp.Body.Close()
			company := u.GetCompany()
			if company != "" {
				company = strings.ToLower(company)
				company = strings.ReplaceAll(company, "inc.", "")
				company = strings.ReplaceAll(company, "llc", "")
				company = strings.ReplaceAll(company, ",", "")
				company = strings.TrimLeft(company, "@")
				company = strings.Trim(company, " ")
				companies[company] = struct{}{}
			}
		}
	}
	names := []string{}
	for c := range companies {
		names = append(names, c)
	}
	c.Logf("companies found: %v", strings.Join(names, ","))
	const numContributors = 2
	const confidence = 10
	if len(companies) >= numContributors {
		return checker.CheckResult{
			Pass:       true,
			Confidence: confidence,
		}
	}
	return checker.CheckResult{
		Pass:       false,
		Confidence: confidence,
	}
}
