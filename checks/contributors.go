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
	"fmt"
	"strings"

	"github.com/ossf/scorecard/v2/checker"
	sce "github.com/ossf/scorecard/v2/errors"
)

const (
	minContributionsPerUser    = 5
	numberCompaniesForTopScore = 3
	// CheckContributors is the registered name for Contributors.
	CheckContributors = "Contributors"
)

//nolint:gochecknoinits
func init() {
	registerCheck(CheckContributors, Contributors)
}

// Contributors run Contributors check.
func Contributors(c *checker.CheckRequest) checker.CheckResult {
	contribs, err := c.RepoClient.ListContributors()
	if err != nil {
		e := sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("Client.Repositories.ListContributors: %v", err))
		return checker.CreateRuntimeErrorResult(CheckContributors, e)
	}

	companies := map[string]struct{}{}
	for _, contrib := range contribs {
		if contrib.NumContributions < minContributionsPerUser {
			continue
		}

		for _, org := range contrib.Organizations {
			if org.Login != "" {
				companies[org.Login] = struct{}{}
			}
		}

		company := contrib.Company
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

	c.Dlogger.Info("contributors work for: %v", strings.Join(names, ","))

	reason := fmt.Sprintf("%d different companies found", len(companies))
	return checker.CreateProportionalScoreResult(CheckContributors, reason, len(companies), numberCompaniesForTopScore)
}
