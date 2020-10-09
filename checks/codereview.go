package checks

import (
	"github.com/dlorenc/scorecard/checker"
	"github.com/google/go-github/v32/github"
)

func init() {
	AllChecks = append(AllChecks, NamedCheck{
		Name: "Code-Review",
		Fn:   DoesCodeReview,
	})
}

// DoesCodeReview attempts to determine whether a project requires review before code gets merged.
// It uses a set of heuristics:
// - Looking at the repo configuration to see if reviews are required
// - Checking if most of the recent merged PRs were "Approved"
// - Looking for other well-known review labels
func DoesCodeReview(c *checker.Checker) CheckResult {
	return MultiCheck(
		IsPrReviewRequired,
		GithubCodeReview,
		ProwCodeReview,
	)(c)
}

func GithubCodeReview(c *checker.Checker) CheckResult {
	// Look at some merged PRs to see if they were reviewed
	prs, _, err := c.Client.PullRequests.List(c.Ctx, c.Owner, c.Repo, &github.PullRequestListOptions{
		State: "closed",
	})
	if err != nil {
		return InconclusiveResult
	}

	totalMerged := 0
	totalReviewed := 0
	for _, pr := range prs {
		if pr.MergedAt == nil {
			continue
		}
		totalMerged++
		// Merged PR!
		reviews, _, err := c.Client.PullRequests.ListReviews(c.Ctx, c.Owner, c.Repo, pr.GetNumber(), &github.ListOptions{})
		if err != nil {
			continue
		}
		for _, r := range reviews {
			if r.GetState() == "APPROVED" {
				totalReviewed++
				break
			}
		}
	}

	// Threshold is 3/4 of merged PRs
	actual := float32(totalReviewed) / float32(totalMerged)
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

func IsPrReviewRequired(c *checker.Checker) CheckResult {
	// Look to see if review is enforced.
	r, _, err := c.Client.Repositories.Get(c.Ctx, c.Owner, c.Repo)
	if err != nil {
		return RetryResult(err)
	}

	// Check the branch protection rules, we may not be able to get these though.
	bp, _, err := c.Client.Repositories.GetBranchProtection(c.Ctx, c.Owner, c.Repo, r.GetDefaultBranch())
	if err != nil {
		return InconclusiveResult
	}
	if bp.GetRequiredPullRequestReviews().RequiredApprovingReviewCount >= 1 {
		return CheckResult{
			Pass:       true,
			Confidence: 10,
		}
	}
	return InconclusiveResult
}

func ProwCodeReview(c *checker.Checker) CheckResult {
	// Look at some merged PRs to see if they were reviewed
	prs, _, err := c.Client.PullRequests.List(c.Ctx, c.Owner, c.Repo, &github.PullRequestListOptions{
		State: "closed",
	})
	if err != nil {
		return InconclusiveResult
	}

	totalMerged := 0
	totalReviewed := 0
	for _, pr := range prs {
		if pr.MergedAt == nil {
			continue
		}
		totalMerged++
		for _, l := range pr.Labels {
			if l.GetName() == "lgtm" || l.GetName() == "approved" {
				totalReviewed++
				break
			}
		}
	}
	// Threshold is 3/4 of merged PRs
	actual := float32(totalReviewed) / float32(totalMerged)
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
