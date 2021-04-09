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
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/ossf/scorecard/checker"
)

const (
	activeStr    = "Active"
	lookbackDays = 90
)

func init() {
	registerCheck(activeStr, IsActive)
}

func IsActive(c checker.CheckRequest) checker.CheckResult {
	commits, _, err := c.Client.Repositories.ListCommits(c.Ctx, c.Owner, c.Repo, &github.CommitsListOptions{})
	if err != nil {
		return checker.MakeRetryResult(activeStr, err)
	}

	tz, _ := time.LoadLocation("UTC")
	threshold := time.Now().In(tz).AddDate(0, 0, -1*lookbackDays)
	totalCommits := 0
	for _, commit := range commits {
		commitFull, _, err := c.Client.Git.GetCommit(c.Ctx, c.Owner, c.Repo, commit.GetSHA())
		if err != nil {
			return checker.MakeRetryResult(activeStr, err)
		}
		if commitFull.GetAuthor().GetDate().After(threshold) {
			totalCommits++
		}
	}
	c.Logf("commits in last %d days: %d", lookbackDays, totalCommits)
	const numCommits = 2
	const confidence = 10
	return checker.CheckResult{
		Name:       activeStr,
		Pass:       totalCommits >= numCommits,
		Confidence: confidence,
	}
}
