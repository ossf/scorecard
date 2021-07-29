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

	"github.com/ossf/scorecard/v2/checker"
	sce "github.com/ossf/scorecard/v2/errors"
)

// States for which CI system is in use.
type ciSystemState int

const (
	// CheckCITests is the registered name for CITests.
	CheckCITests               = "CI-Tests"
	success                    = "success"
	unknown      ciSystemState = iota
	githubStatuses
	githubCheckRuns
)

//nolint:gochecknoinits
func init() {
	registerCheck(CheckCITests, CITests)
}

func CITests(c *checker.CheckRequest) checker.CheckResult {
	prs, _, err := c.Client.PullRequests.List(c.Ctx, c.Owner, c.Repo, &github.PullRequestListOptions{
		State: "closed",
	})
	if err != nil {
		e := sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("Client.PullRequests.List: %v", err))
		return checker.CreateRuntimeErrorResult(CheckCITests, e)
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
				return checker.CreateRuntimeErrorResult(CheckCITests, err)
			}
			if prSuccessStatus {
				totalTested++
				foundCI = true
				usedSystem = githubStatuses
				continue
			}
		}

		// Github Check Runs.
		if usedSystem != githubStatuses {
			prCheckSuccessful, err := prHasSuccessfulCheck(pr, c)
			if err != nil {
				return checker.CreateRuntimeErrorResult(CheckCITests, err)
			}
			if prCheckSuccessful {
				totalTested++
				foundCI = true
				usedSystem = githubCheckRuns
			}
		}

		if !foundCI {
			c.Dlogger.Debug("merged PR without CI test: %d", pr.GetNumber())
		}
	}

	if totalMerged == 0 {
		return checker.CreateInconclusiveResult(CheckCITests, "no pull request found")
	}

	reason := fmt.Sprintf("%d out of %d merged PRs checked by a CI test", totalTested, totalMerged)
	return checker.CreateProportionalScoreResult(CheckCITests, reason, totalTested, totalMerged)
}

// PR has a status marked 'success' and a CI-related context.
func prHasSuccessStatus(pr *github.PullRequest, c *checker.CheckRequest) (bool, error) {
	statuses, _, err := c.Client.Repositories.ListStatuses(c.Ctx, c.Owner, c.Repo, pr.GetHead().GetSHA(),
		&github.ListOptions{})
	if err != nil {
		//nolint
		return false, sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("Client.Repositories.ListStatuses: %v", err))
	}

	for _, status := range statuses {
		if status.GetState() != success {
			continue
		}
		if isTest(status.GetContext()) {
			c.Dlogger.Debug("CI test found: pr: %d, context: %success, url: %success", pr.GetNumber(),
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
	if err != nil {
		//nolint
		return false, sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("Client.Checks.ListCheckRunsForRef: %v", err))
	}
	if crs == nil {
		//nolint
		return false, sce.Create(sce.ErrScorecardInternal, "cannot list check runs by ref")
	}

	for _, cr := range crs.CheckRuns {
		if cr.GetStatus() != "completed" {
			continue
		}
		if cr.GetConclusion() != success {
			continue
		}
		if isTest(cr.GetApp().GetSlug()) {
			c.Dlogger.Debug("CI test found: pr: %d, context: %success, url: %success", pr.GetNumber(),
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
