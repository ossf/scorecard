package checks

import (
	"strings"

	"github.com/dlorenc/scorecard/checker"
	"github.com/google/go-github/v32/github"
)

func init() {
	registerCheck("Contributors", Contributors)
}

func Contributors(c checker.Checker) checker.CheckResult {
	contribs, _, err := c.Client.Repositories.ListContributors(c.Ctx, c.Owner, c.Repo, &github.ListContributorsOptions{})
	if err != nil {
		return checker.RetryResult(err)
	}

	companies := map[string]struct{}{}
	for _, contrib := range contribs {
		if contrib.GetContributions() >= 5 {
			u, _, err := c.Client.Users.Get(c.Ctx, contrib.GetLogin())
			if err != nil {
				return checker.RetryResult(err)
			}
			if u.GetCompany() != "" {
				company := strings.ToLower(strings.Trim(strings.TrimSpace(u.GetCompany()), "@"))
				companies[company] = struct{}{}
			}
		}
	}
	names := []string{}
	for c := range companies {
		names = append(names, c)
	}
	c.Logf("companies found: %v", strings.Join(names, ","))
	if len(companies) > 2 {
		return checker.CheckResult{
			Pass:       true,
			Confidence: 10,
		}
	}
	return checker.CheckResult{
		Pass:       false,
		Confidence: 10,
	}
}
