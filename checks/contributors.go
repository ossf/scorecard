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

func init() {
	registerCheck("Contributors", Contributors)
}

func Contributors(c checker.Checker) checker.CheckResult {
	contribs, _, err := c.Client.Repositories.ListContributors(c.Ctx, c.Owner, c.Repo, &github.ListContributorsOptions{})
	if err != nil {
		return checker.RetryResult(err)
	}

	companies := map[string]struct{}{}
	for _, contrib := range contribs {
		if contrib.GetContributions() >= 5 {
			u, _, err := c.Client.Users.Get(c.Ctx, contrib.GetLogin())
			if err != nil {
				return checker.RetryResult(err)
			}
			if u.GetCompany() != "" {
				company := strings.ToLower(strings.Trim(strings.TrimSpace(u.GetCompany()), "@"))
				companies[company] = struct{}{}
			}
		}
	}
	names := []string{}
	for c := range companies {
		names = append(names, c)
	}
	c.Logf("companies found: %v", strings.Join(names, ","))
	if len(companies) >= 2 {
		return checker.CheckResult{
			Pass:       true,
			Confidence: 10,
		}
	}
	return checker.CheckResult{
		Pass:       false,
		Confidence: 10,
	}
}
