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

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

const (
	// CheckCITests is the registered name for CITests.
	CheckCITests = "CI-Tests"
	success      = "success"
)

//nolint:gochecknoinits
func init() {
	if err := registerCheck(CheckCITests, CITests); err != nil {
		// this should never happen
		panic(err)
	}
}

// CITests runs CI-Tests check.
func CITests(c *checker.CheckRequest) checker.CheckResult {
	commits, err := c.RepoClient.ListCommits()
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("RepoClient.ListCommits: %v", err))
		return checker.CreateRuntimeErrorResult(CheckCITests, e)
	}

	totalMerged := 0
	totalTested := 0
	for i := range commits {
		pr := &commits[i].AssociatedMergeRequest
		// TODO(#575): We ignore associated PRs if Scorecard is being run on a fork
		// but the PR was created in the original repo.
		if pr.MergedAt.IsZero() || pr.Repository != c.Repo.String() {
			continue
		}
		totalMerged++

		var foundCI bool

		// Github Statuses.
		prSuccessStatus, err := prHasSuccessStatus(pr, c)
		if err != nil {
			return checker.CreateRuntimeErrorResult(CheckCITests, err)
		}
		if prSuccessStatus {
			totalTested++
			foundCI = true
			continue
		}

		// Github Check Runs.
		prCheckSuccessful, err := prHasSuccessfulCheck(pr, c)
		if err != nil {
			return checker.CreateRuntimeErrorResult(CheckCITests, err)
		}
		if prCheckSuccessful {
			totalTested++
			foundCI = true
		}

		if !foundCI {
			c.Dlogger.Debug3(&checker.LogMessage{
				Text: fmt.Sprintf("merged PR without CI test: %d", pr.Number),
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
func prHasSuccessStatus(pr *clients.PullRequest, c *checker.CheckRequest) (bool, error) {
	statuses, err := c.RepoClient.ListStatuses(pr.HeadSHA)
	if err != nil {
		return false, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("Client.Repositories.ListStatuses: %v", err))
	}

	for _, status := range statuses {
		if status.State != success {
			continue
		}
		if isTest(status.Context) || isTest(status.TargetURL) {
			c.Dlogger.Debug3(&checker.LogMessage{
				Path: status.URL,
				Type: checker.FileTypeURL,
				Text: fmt.Sprintf("CI test found: pr: %d, context: %s", pr.Number,
					status.Context),
			})
			return true, nil
		}
	}
	return false, nil
}

// PR has a successful CI-related check.
func prHasSuccessfulCheck(pr *clients.PullRequest, c *checker.CheckRequest) (bool, error) {
	crs, err := c.RepoClient.ListCheckRunsForRef(pr.HeadSHA)
	if err != nil {
		return false, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("Client.Checks.ListCheckRunsForRef: %v", err))
	}

	for _, cr := range crs {
		if cr.Status != "completed" {
			continue
		}
		if cr.Conclusion != success {
			continue
		}
		if isTest(cr.App.Slug) {
			c.Dlogger.Debug3(&checker.LogMessage{
				Path: cr.URL,
				Type: checker.FileTypeURL,
				Text: fmt.Sprintf("CI test found: pr: %d, context: %s", pr.Number,
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
		"flutter-dashboard", "Cirrus CI",
	} {
		if strings.Contains(l, pattern) {
			return true
		}
	}
	return false
}
