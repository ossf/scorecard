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
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/ossf/scorecard/checker"
)

const (
	minContributionsPerUser = 5
	minOrganizationCount    = 2
	// CheckContributors is the registered name for Contributors.
	CheckContributors = "Contributors"
)

//nolint:gochecknoinits
func init() {
	registerCheck(CheckContributors, Contributors)
}

func Contributors(c *checker.CheckRequest) checker.CheckResult {
	contribs, _, err := c.Client.Repositories.ListContributors(c.Ctx, c.Owner, c.Repo, &github.ListContributorsOptions{})
	if err != nil {
		return checker.MakeRetryResult(CheckContributors, err)
	}

	companies := map[string]struct{}{}
	for _, contrib := range contribs {
		if contrib.GetContributions() < minContributionsPerUser {
			continue
		}
		u, _, err := c.Client.Users.Get(c.Ctx, contrib.GetLogin())
		if err != nil {
			return checker.MakeRetryResult(CheckContributors, err)
		}
		orgs, _, err := c.Client.Organizations.List(c.Ctx, contrib.GetLogin(), nil)
		if err != nil {
			c.Logf("unable to get org members for %s", contrib.GetLogin())
		} else if len(orgs) > 0 {
			companies[*orgs[0].Login] = struct{}{}
			continue
		}

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
	names := []string{}
	for c := range companies {
		names = append(names, c)
	}
	c.Logf("companies found: %v", strings.Join(names, ","))
	if len(companies) >= minOrganizationCount {
		return checker.CheckResult{
			Name:       CheckContributors,
			Pass:       true,
			Confidence: checker.MaxResultConfidence,
		}
	}
	return checker.CheckResult{
		Name:       CheckContributors,
		Pass:       false,
		Confidence: checker.MaxResultConfidence,
	}
}
