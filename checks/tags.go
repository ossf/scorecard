package checks

import (
	"github.com/dlorenc/scorecard/checker"
	"github.com/google/go-github/v32/github"
)

func init() {
	AllChecks = append(AllChecks, NamedCheck{
		Name: "Signed-Tags",
		Fn:   SignedTags,
	})
}

func SignedTags(c *checker.Checker) CheckResult {
	tags, _, err := c.Client.Repositories.ListTags(c.Ctx, c.Owner, c.Repo, &github.ListOptions{})
	if err != nil {
		return RetryResult(err)
	}

	totalReleases := 0
	totalSigned := 0
	for _, t := range tags {
		totalReleases++
		gt, _, err := c.Client.Git.GetCommit(c.Ctx, c.Owner, c.Repo, t.GetCommit().GetSHA())
		if err != nil {
			return RetryResult(err)
		}
		if gt.GetVerification().GetVerified() {
			totalSigned++
		}
	}

	// Threshold is 3/4 of releases
	actual := float32(totalSigned) / float32(totalReleases)
	if actual >= .75 {
		return CheckResult{
			Pass:       true,
			Confidence: int(actual * 10),
		}
	}
	return CheckResult{
		Pass:       false,
		Confidence: int(10 - int(actual*10)),
	}
}
