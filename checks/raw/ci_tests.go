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
