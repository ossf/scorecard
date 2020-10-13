package checks

import (
	"github.com/dlorenc/scorecard/checker"
	"github.com/google/go-github/v32/github"
)

func init() {
	registerCheck("Signed-Tags", SignedTags)
}

func SignedTags(c checker.Checker) checker.CheckResult {
	tags, _, err := c.Client.Repositories.ListTags(c.Ctx, c.Owner, c.Repo, &github.ListOptions{})
	if err != nil {
		return checker.RetryResult(err)
	}

	totalReleases := 0
	totalSigned := 0
	for _, t := range tags {
		totalReleases++
		gt, _, err := c.Client.Git.GetCommit(c.Ctx, c.Owner, c.Repo, t.GetCommit().GetSHA())
		if err != nil {
			return checker.RetryResult(err)
		}
		if gt.GetVerification().GetVerified() {
			c.Logf("signed tag found: %s", t.GetCommit().GetSHA())
			totalSigned++
		}
	}

	return checker.ProportionalResult(totalSigned, totalReleases, .75)
}
