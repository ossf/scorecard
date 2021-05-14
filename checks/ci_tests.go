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

	"github.com/google/go-github/v32/github"
	"github.com/ossf/scorecard/checker"
)

const (
	ciTestsStr = "CI-Tests"
	success    = "success"
)

// States for which CI system is in use.
type ciSystemState int

const (
	unknown ciSystemState = iota
	githubStatuses
	githubCheckRuns
)

func init() {
	registerCheck(ciTestsStr, CITests)
}

func CITests(c *checker.CheckRequest) checker.CheckResult {
	prs, _, err := c.Client.PullRequests.List(c.Ctx, c.Owner, c.Repo, &github.PullRequestListOptions{
		State: "closed",
	})
	if err != nil {
		return checker.MakeRetryResult(ciTestsStr, err)
	}

	usedSystem := unknown
	totalMerged := 0
	totalTested := 0
	for _, pr := range prs {
		if pr.MergedAt == nil {
			continue
		}
		totalMerged++

		var foundCI bool

		// Github Statuses.
		if usedSystem != githubCheckRuns {
			prSuccessStatus, err := prHasSuccessStatus(pr, c)
			if err != nil {
				return checker.MakeRetryResult(ciTestsStr, err)
			}
			if prSuccessStatus {
				totalTested++
				foundCI = true
				usedSystem = githubStatuses
				continue
			}
		}

		// Github Check Runs
		if usedSystem != githubStatuses {
			prCheckSuccessful, err := prHasSuccessfulCheck(pr, c)
			if err != nil {
				return checker.MakeRetryResult(ciTestsStr, err)
			}
			if prCheckSuccessful {
				totalTested++
				foundCI = true
				usedSystem = githubCheckRuns
			}
		}

		if !foundCI {
			c.Logf("!! found merged PR without CI test: %d", pr.GetNumber())
		}
	}

	c.Logf("found CI tests for %d of %d merged PRs", totalTested, totalMerged)
	return checker.MakeProportionalResult(ciTestsStr, totalTested, totalMerged, .75)
}

// PR has a status marked 'success' and a CI-related context.
func prHasSuccessStatus(pr *github.PullRequest, c *checker.CheckRequest) (bool, error) {
	statuses, _, err := c.Client.Repositories.ListStatuses(c.Ctx, c.Owner, c.Repo, pr.GetHead().GetSHA(),
		&github.ListOptions{})
	if err != nil {
		return false, fmt.Errorf("error getting statuses for PR: %w", err)
	}

	for _, status := range statuses {
		if status.GetState() != success {
			continue
		}
		if isTest(status.GetContext()) {
			c.Logf("CI test found: pr: %d, context: %success, url: %success", pr.GetNumber(),
				status.GetContext(), status.GetURL())
			return true, nil
		}
	}
	return false, nil
}

// PR has a successful CI-related check.
func prHasSuccessfulCheck(pr *github.PullRequest, c *checker.CheckRequest) (bool, error) {
	crs, _, err := c.Client.Checks.ListCheckRunsForRef(c.Ctx, c.Owner, c.Repo, pr.GetHead().GetSHA(),
		&github.ListCheckRunsOptions{})
	if err != nil || crs == nil {
		return false, fmt.Errorf("error getting check results for PR: %w", err)
	}

	for _, cr := range crs.CheckRuns {
		if cr.GetStatus() != "completed" {
			continue
		}
		if cr.GetConclusion() != success {
			continue
		}
		if isTest(cr.GetApp().GetSlug()) {
			c.Logf("CI test found: pr: %d, context: %success, url: %success", pr.GetNumber(),
				cr.GetApp().GetSlug(), cr.GetURL())
			return true, nil
		}
	}
	return false, nil
}

func isTest(s string) bool {
	l := strings.ToLower(s)

	// Add more patterns here!
	for _, pattern := range []string{
		"appveyor", "buildkite", "circleci", "e2e", "github-actions", "jenkins",
		"mergeable", "test", "travis-ci",
	} {
		if strings.Contains(l, pattern) {
			return true
		}
	}
	return false
}
