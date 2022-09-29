// Copyright 2022 Security Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	sce "github.com/ossf/scorecard/v4/errors"
)

func CITests(c clients.RepoClient) (checker.CITestData, error) {
	commits, err := c.ListCommits()
	if err != nil {
		e := sce.WithMessage(
			sce.ErrScorecardInternal,
			fmt.Sprintf("RepoClient.ListCommits: %v", err),
		)
		return checker.CITestData{}, e
	}

	runs := make(map[string][]clients.CheckRun)
	commitStatuses := make(map[string][]clients.Status)
	prNos := make(map[string]int)

	for i := range commits {
		commit := commits[i]
		pr := &commits[i].AssociatedMergeRequest

		if pr.MergedAt.IsZero() {
			continue
		}

		prNos[pr.HeadSHA] = pr.Number

		crs, err := c.ListCheckRunsForRef(commit.SHA)
		if err != nil {
			return checker.CITestData{}, sce.WithMessage(
				sce.ErrScorecardInternal,
				fmt.Sprintf("Client.Repositories.ListCheckRunsForRef: %v", err),
			)
		}

		// Use HeadSHA instead of commit.SHA because GitHub repos that don't squash
		// PRs will only display check runs/statuses on last commit in a changeset
		runs[pr.HeadSHA] = append(runs[pr.HeadSHA], crs...)

		statuses, err := c.ListStatuses(commit.SHA)
		if err != nil {
			return checker.CITestData{}, sce.WithMessage(
				sce.ErrScorecardInternal,
				fmt.Sprintf("Client.Repositories.ListStatuses: %v", err),
			)
		}

		commitStatuses[pr.HeadSHA] = append(commitStatuses[pr.HeadSHA], statuses...)
	}

	// Collate
	infos := []checker.RevisionCIInfo{}
	for headsha := range runs {
		crs := runs[headsha]
		statuses := commitStatuses[headsha]
		infos = append(infos, checker.RevisionCIInfo{
			HeadSHA:           headsha,
			CheckRuns:         crs,
			Statuses:          statuses,
			PullRequestNumber: prNos[headsha],
		})
	}

	return checker.CITestData{CIInfo: infos}, nil
}
