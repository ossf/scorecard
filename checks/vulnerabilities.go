// Copyright 2021 Security Scorecard Authors
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

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

const (
	// CheckVulnerabilities is the registered name for the OSV check.
	CheckVulnerabilities = "Vulnerabilities"
)

//nolint:gochecknoinits
func init() {
	registerCheck(CheckVulnerabilities, HasUnfixedVulnerabilities)
}

func getVulnerabilities(resp *clients.VulnerabilitiesResponse) []string {
	ids := make([]string, 0, len(resp.Vulns))
	for _, vuln := range resp.Vulns {
		ids = append(ids, vuln.ID)
	}
	return ids
}

// HasUnfixedVulnerabilities runs Vulnerabilities check.
func HasUnfixedVulnerabilities(c *checker.CheckRequest) checker.CheckResult {
	commits, err := c.RepoClient.ListCommits()
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "Client.Repositories.ListCommits")
		return checker.CreateRuntimeErrorResult(CheckVulnerabilities, e)
	}

	if len(commits) < 1 || commits[0].SHA == "" {
		return checker.CreateInconclusiveResult(CheckVulnerabilities, "no commits found")
	}

	resp, err := c.VulnerabilitiesClient.HasUnfixedVulnerabilities(c.Ctx, commits[0].SHA)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "VulnerabilitiesClient.HasUnfixedVulnerabilities")
		return checker.CreateRuntimeErrorResult(CheckVulnerabilities, e)
	}

	// TODO: take severity into account.
	vulnIDs := getVulnerabilities(&resp)
	if len(vulnIDs) > 0 {
		c.Dlogger.Warn3(&checker.LogMessage{
			Text: fmt.Sprintf("HEAD is vulnerable to %s", strings.Join(vulnIDs, ", ")),
		})
		return checker.CreateMinScoreResult(CheckVulnerabilities, "existing vulnerabilities detected")
	}

	return checker.CreateMaxScoreResult(CheckVulnerabilities, "no vulnerabilities detected")
}
