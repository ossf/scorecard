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
	"time"

	"github.com/ossf/scorecard/v3/checker"
	sce "github.com/ossf/scorecard/v3/errors"
)

const (
	// CheckMaintained is the exported check name for Maintained.
	CheckMaintained = "Maintained"
	lookBackDays    = 90
	activityPerWeek = 1
	daysInOneWeek   = 7
)

//nolint:gochecknoinits
func init() {
	registerCheck(CheckMaintained, IsMaintained)
}

// IsMaintained runs Maintained check.
func IsMaintained(c *checker.CheckRequest) checker.CheckResult {
	archived, err := c.RepoClient.IsArchived()
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckMaintained, e)
	}
	if archived {
		return checker.CreateMinScoreResult(CheckMaintained, "repo is marked as archived")
	}

	// If not explicitly marked archived, look for activity in past `lookBackDays`.
	threshold := time.Now().AddDate(0 /*years*/, 0 /*months*/, -1*lookBackDays /*days*/)

	commits, err := c.RepoClient.ListCommits()
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckMaintained, e)
	}
	commitsWithinThreshold := 0
	for _, commit := range commits {
		if commit.CommittedDate.After(threshold) {
			commitsWithinThreshold++
		}
	}

	issues, err := c.RepoClient.ListIssues()
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckMaintained, e)
	}
	issuesUpdatedWithinThreshold := 0
	for _, issue := range issues {
		if issue.UpdatedAt.After(threshold) {
			issuesUpdatedWithinThreshold++
		}
	}

	return checker.CreateProportionalScoreResult(CheckMaintained, fmt.Sprintf(
		"%d commit(s) out of %d and %d issue activity out of %d found in the last %d days",
		commitsWithinThreshold, len(commits), issuesUpdatedWithinThreshold, len(issues), lookBackDays),
		commitsWithinThreshold+issuesUpdatedWithinThreshold, activityPerWeek*lookBackDays/daysInOneWeek)
}
