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

	"github.com/ossf/scorecard/v2/checker"
	sce "github.com/ossf/scorecard/v2/errors"
)

const (
	// CheckMaintained is the exported check name for Maintained.
	CheckMaintained = "Maintained"
	lookBackDays    = 90
	commitsPerWeek  = 1
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
		return checker.CreateRuntimeErrorResult(CheckMaintained, err)
	}
	if archived {
		return checker.CreateMinScoreResult(CheckMaintained, "repo is marked as archived")
	}

	commits, err := c.RepoClient.ListCommits()
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckMaintained, err)
	}

	tz, err := time.LoadLocation("UTC")
	if err != nil {
		e := sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("time.LoadLocation: %v", err))
		return checker.CreateRuntimeErrorResult(CheckMaintained, e)
	}
	threshold := time.Now().In(tz).AddDate(0, 0, -1*lookBackDays)
	totalCommits := 0
	for _, commit := range commits {
		if commit.CommittedDate.After(threshold) {
			totalCommits++
		}
	}
	return checker.CreateProportionalScoreResult(CheckMaintained,
		fmt.Sprintf("%d commit(s) found in the last %d days", totalCommits, lookBackDays),
		totalCommits, commitsPerWeek*lookBackDays/daysInOneWeek)
}
