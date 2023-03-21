// Copyright 2022 OpenSSF Scorecard Authors
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
	"errors"
	"fmt"
	"io"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

const commitsToAnalye = 30

func CITests(c clients.RepoClient) (checker.CITestData, error) {
	commitIter, err := c.ListCommits()
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

	commit, err := commitIter.Next()
	for i := 0; i < commitsToAnalye && err == nil; func() { commit, err = commitIter.Next(); i++ }() {
		pr := &commit.AssociatedMergeRequest

		if pr.MergedAt.IsZero() {
			continue
		}

		prNos[pr.HeadSHA] = pr.Number

		// HeadSHA is the last commit before the merge. if squashing enabled,
		// multiple commit SHAs will map to a single HeadSHA
		if len(runs[pr.HeadSHA]) == 0 {
			crs, err := c.ListCheckRunsForRef(pr.HeadSHA)
			if err != nil {
				return checker.CITestData{}, sce.WithMessage(
					sce.ErrScorecardInternal,
					fmt.Sprintf("Client.Repositories.ListCheckRunsForRef: %v", err),
				)
			}

			runs[pr.HeadSHA] = crs
		}

		statuses, err := c.ListStatuses(pr.HeadSHA)
		if err != nil {
			return checker.CITestData{}, sce.WithMessage(
				sce.ErrScorecardInternal,
				fmt.Sprintf("Client.Repositories.ListStatuses: %v", err),
			)
		}

		commitStatuses[pr.HeadSHA] = append(commitStatuses[pr.HeadSHA], statuses...)
	}
	if err != nil && !errors.Is(err, io.EOF) {
		return checker.CITestData{}, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("commitIter.Next: %v", err))
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
