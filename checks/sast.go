package checks

import (
	"github.com/dlorenc/scorecard/checker"
	"github.com/google/go-github/v32/github"
)

func init() {
	registerCheck("SAST", checker.MultiCheck(CodeQLActionRuns))
}

func CodeQLActionRuns(c checker.Checker) checker.CheckResult {
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
		if err != nil {
			return checker.RetryResult(err)
		}
		for _, cr := range crs.CheckRuns {
			if cr.GetStatus() != "completed" {
				continue
			}
			if cr.GetConclusion() != "success" {
				continue
			}
			if cr.GetApp().GetSlug() == "github-code-scanning" {
				c.Logf("GitHub code scan found: %s", cr.GetHTMLURL())
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
