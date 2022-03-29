// Copyright 2022 Security Scorecard Authors
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
	"errors"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
)

var errNoCommitFound = errors.New("no commit found")

// Vulnerabilities retrieves the raw data for the Vulnerabilities check.
func Vulnerabilities(c *checker.CheckRequest) (checker.VulnerabilitiesData, error) {
	commits, err := c.RepoClient.ListCommits()
	if err != nil {
		return checker.VulnerabilitiesData{}, fmt.Errorf("repoClient.ListCommits: %w", err)
	}

	if len(commits) < 1 || commits[0].SHA == "" {
		return checker.VulnerabilitiesData{}, fmt.Errorf("%w", errNoCommitFound)
	}

	resp, err := c.VulnerabilitiesClient.HasUnfixedVulnerabilities(c.Ctx, commits[0].SHA)
	if err != nil {
		return checker.VulnerabilitiesData{}, fmt.Errorf("vulnerabilitiesClient.HasUnfixedVulnerabilities: %w", err)
	}

	vulnIDs := getVulnerabilities(&resp)
	vulns := []checker.Vulnerability{}
	for _, id := range vulnIDs {
		v := checker.Vulnerability{
			ID: id,
			// Note: add fields if needed.
		}
		vulns = append(vulns, v)
	}
	return checker.VulnerabilitiesData{Vulnerabilities: vulns}, nil
}

func getVulnerabilities(resp *clients.VulnerabilitiesResponse) []string {
	ids := make([]string, 0, len(resp.Vulns))
	for _, vuln := range resp.Vulns {
		ids = append(ids, vuln.ID)
	}
	return ids
}
