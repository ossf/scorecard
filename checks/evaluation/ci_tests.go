// Copyright 2022 OpenSSF Scorecard Authors
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

package evaluation

import (
	"fmt"
	"strings"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
)

const (
	// CheckCITests is the registered name for CITests.
	CheckCITests = "CI-Tests"
	success      = "success"
)

func CITests(name string, c *checker.CITestData, dl checker.DetailLogger) checker.CheckResult {
	totalMerged := 0
	totalTested := 0
	for i := range c.CIInfo {
		r := c.CIInfo[i]
		totalMerged++

		var foundCI bool

		// Github Statuses.
		prSuccessStatus, err := prHasSuccessStatus(r, dl)
		if err != nil {
			return checker.CreateRuntimeErrorResult(CheckCITests, err)
		}
		if prSuccessStatus {
			totalTested++
			foundCI = true
			continue
		}

		// Github Check Runs.
		prCheckSuccessful, err := prHasSuccessfulCheck(r, dl)
		if err != nil {
			return checker.CreateRuntimeErrorResult(CheckCITests, err)
		}
		if prCheckSuccessful {
			totalTested++
			foundCI = true
		}

		if !foundCI {
			// Log message says commit, but really we only care about PRs, and
			// use only one commit (branch HEAD) to refer to all commits in a PR
			dl.Debug(&checker.LogMessage{
				Text: fmt.Sprintf("merged PR without CI test at HEAD: %s", r.HeadSHA),
			})
		}
	}

	if totalMerged == 0 {
		return checker.CreateInconclusiveResult(CheckCITests, "no pull request found")
	}

	reason := fmt.Sprintf("%d out of %d merged PRs checked by a CI test", totalTested, totalMerged)
	return checker.CreateProportionalScoreResult(CheckCITests, reason, totalTested, totalMerged)
}

// PR has a status marked 'success' and a CI-related context.
//
//nolint:unparam
func prHasSuccessStatus(r checker.RevisionCIInfo, dl checker.DetailLogger) (bool, error) {
	for _, status := range r.Statuses {
		if status.State != success {
			continue
		}
		if isTest(status.Context) || isTest(status.TargetURL) {
			dl.Debug(&checker.LogMessage{
				Path: status.URL,
				Type: finding.FileTypeURL,
				Text: fmt.Sprintf("CI test found: pr: %s, context: %s", r.HeadSHA,
					status.Context),
			})
			return true, nil
		}
	}
	return false, nil
}

// PR has a successful CI-related check.
//
//nolint:unparam
func prHasSuccessfulCheck(r checker.RevisionCIInfo, dl checker.DetailLogger) (bool, error) {
	for _, cr := range r.CheckRuns {
		if cr.Status != "completed" {
			continue
		}
		if cr.Conclusion != success {
			continue
		}
		if isTest(cr.App.Slug) {
			dl.Debug(&checker.LogMessage{
				Path: cr.URL,
				Type: finding.FileTypeURL,
				Text: fmt.Sprintf("CI test found: pr: %d, context: %s", r.PullRequestNumber,
					cr.App.Slug),
			})
			return true, nil
		}
	}
	return false, nil
}

// isTest returns true if the given string is a CI test.
func isTest(s string) bool {
	l := strings.ToLower(s)

	// Add more patterns here!
	for _, pattern := range []string{
		"appveyor", "buildkite", "circleci", "e2e", "github-actions", "jenkins",
		"mergeable", "packit-as-a-service", "semaphoreci", "test", "travis-ci",
		"flutter-dashboard", "Cirrus CI", "azure-pipelines",
	} {
		if strings.Contains(l, pattern) {
			return true
		}
	}
	return false
}
