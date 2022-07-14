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
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
)

// Vulnerabilities retrieves the raw data for the Vulnerabilities check.
func Vulnerabilities(c *checker.CheckRequest) (checker.VulnerabilitiesData, error) {
	commits, err := c.RepoClient.ListCommits()
	if err != nil {
		return checker.VulnerabilitiesData{}, fmt.Errorf("repoClient.ListCommits: %w", err)
	}

	if len(commits) < 1 || allOf(commits, hasEmptySHA) {
		return checker.VulnerabilitiesData{}, nil
	}

	resp, err := c.VulnerabilitiesClient.HasUnfixedVulnerabilities(c.Ctx, commits[0].SHA)
	if err != nil {
		return checker.VulnerabilitiesData{}, fmt.Errorf("vulnerabilitiesClient.HasUnfixedVulnerabilities: %w", err)
	}
	return checker.VulnerabilitiesData{
		Vulnerabilities: resp.Vulnerabilities,
	}, nil
}

type predicateOnCommitFn func(clients.Commit) bool

var hasEmptySHA predicateOnCommitFn = func(c clients.Commit) bool {
	return c.SHA == ""
}

func allOf(commits []clients.Commit, predicate func(clients.Commit) bool) bool {
	for i := range commits {
		if !predicate(commits[i]) {
			return false
		}
	}
	return true
}
