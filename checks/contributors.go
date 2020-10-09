package checks

import (
	"github.com/dlorenc/scorecard/checker"
	"github.com/google/go-github/v32/github"
)

func init() {
	AllChecks = append(AllChecks, NamedCheck{
		Name: "Contributors",
		Fn:   Contributors,
	})
}

func Contributors(c *checker.Checker) CheckResult {
	contribs, _, err := c.Client.Repositories.ListContributors(c.Ctx, c.Owner, c.Repo, &github.ListContributorsOptions{})
	if err != nil {
		return RetryResult(err)
	}

	companies := map[string]struct{}{}
	for _, contrib := range contribs {
		if contrib.GetContributions() >= 5 {
			u, _, err := c.Client.Users.Get(c.Ctx, contrib.GetLogin())
			if err != nil {
				return RetryResult(err)
			}
			if u.GetCompany() != "" {
				companies[u.GetCompany()] = struct{}{}
			}
		}
		if len(companies) > 2 {
			return CheckResult{
				Pass:       true,
				Confidence: 10,
			}
		}
	}
	return CheckResult{
		Pass:       false,
		Confidence: 10,
	}
}
