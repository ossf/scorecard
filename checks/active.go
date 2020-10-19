package checks

import (
	"time"

	"github.com/dlorenc/scorecard/checker"
	"github.com/google/go-github/v32/github"
)

var lookbackDays int = 90

func init() {
	registerCheck("Active", IsActive)
}

func IsActive(c checker.Checker) checker.CheckResult {
	return checker.MultiCheck(
		PeriodicCommits,
		PeriodicReleases,
	)(c)
}

func PeriodicCommits(c checker.Checker) checker.CheckResult {
	commits, _, err := c.Client.Repositories.ListCommits(c.Ctx, c.Owner, c.Repo, &github.CommitsListOptions{})
	if err != nil {
		return checker.RetryResult(err)
	}

	tz, _ := time.LoadLocation("UTC")
	threshold := time.Now().In(tz).AddDate(0, 0, -1*lookbackDays)
	totalCommits := 0
	for _, commit := range commits {
		commitFull, _, err := c.Client.Git.GetCommit(c.Ctx, c.Owner, c.Repo, commit.GetSHA())
		if err != nil {
			return checker.RetryResult(err)
		}
		if commitFull.GetAuthor().GetDate().After(threshold) {
			totalCommits++
		}
	}

	return checker.CheckResult{
		Pass:       totalCommits >= 2,
		Confidence: 7,
	}
}

func PeriodicReleases(c checker.Checker) checker.CheckResult {
	releases, _, err := c.Client.Repositories.ListReleases(c.Ctx, c.Owner, c.Repo, &github.ListOptions{})
	if err != nil {
		return checker.RetryResult(err)
	}

	tz, _ := time.LoadLocation("UTC")
	threshold := time.Now().In(tz).AddDate(0, 0, -1*lookbackDays)
	totalReleases := 0
	for _, r := range releases {
		if r.GetCreatedAt().After(threshold) {
			totalReleases++
		}
	}

	return checker.CheckResult{
		Pass:       totalReleases > 0,
		Confidence: 10,
	}
}
