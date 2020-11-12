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
	registerCheck("CI-Tests", checker.MultiCheck(GithubStatuses, GithubCheckRuns))
}

func GithubStatuses(c checker.Checker) checker.CheckResult {
	prs, _, err := c.Client.PullRequests.List(c.Ctx, c.Owner, c.Repo, &github.PullRequestListOptions{
		State: "closed",
	})
	if err != nil {
		return checker.RetryResult(err)
	}

	totalMerged := 0
	totalTested := 0
	for _, pr := range prs {
		if pr.MergedAt == nil {
			continue
		}
		totalMerged++
		statuses, _, err := c.Client.Repositories.ListStatuses(c.Ctx, c.Owner, c.Repo, pr.GetHead().GetSHA(), &github.ListOptions{})
		if err != nil {
			return checker.RetryResult(err)
		}
		for _, status := range statuses {
			if status.GetState() != "success" {
				continue
			}
			if isTest(status.GetContext()) {
				c.Logf("CI test found: context: %s, url: %s", status.GetContext(), status.GetURL())
				totalTested++
				break
			}
		}
	}
	if totalTested == 0 {
		return checker.InconclusiveResult
	}
	return checker.ProportionalResult(totalTested, totalMerged, .75)
}

func isTest(s string) bool {
	l := strings.ToLower(s)

	// Add more patterns here!
	for _, pattern := range []string{"appveyor", "buildkite", "circleci", "e2e", "github-actions", "mergeable", "test", "travis-ci"} {
		if strings.Contains(l, pattern) {
			return true
		}
	}
	return false
}

func GithubCheckRuns(c checker.Checker) checker.CheckResult {
	prs, _, err := c.Client.PullRequests.List(c.Ctx, c.Owner, c.Repo, &github.PullRequestListOptions{
		State: "closed",
	})
	if err != nil {
		return checker.RetryResult(err)
	}

	totalMerged := 0
	totalTested := 0
	for _, pr := range prs {
		if pr.MergedAt == nil {
			continue
		}
		totalMerged++
		crs, _, err := c.Client.Checks.ListCheckRunsForRef(c.Ctx, c.Owner, c.Repo, pr.GetHead().GetSHA(), &github.ListCheckRunsOptions{})
		if err != nil || crs == nil {
			return checker.RetryResult(err)
		}
		for _, cr := range crs.CheckRuns {
			if cr.GetStatus() != "completed" {
				continue
			}
			if cr.GetConclusion() != "success" {
				continue
			}
			if isTest(cr.GetApp().GetSlug()) {
				c.Logf("CI test found: context: %s, url: %s", cr.GetApp().GetSlug(), cr.GetURL())
				totalTested++
				break
			}
		}
	}
	if totalTested == 0 {
		return checker.InconclusiveResult
	}
	return checker.ProportionalResult(totalTested, totalMerged, .75)
}
