package checks

import (
	"github.com/dlorenc/scorecard/checker"
	"github.com/google/go-github/v32/github"
)

func init() {
	AllChecks = append(AllChecks, checker.NamedCheck{
		Name: "Security-MD",
		Fn:   Securitymd,
	})
}

func Securitymd(c checker.Checker) checker.CheckResult {
	for _, fp := range []string{".github/SECURITY.md", ".github/security.md", "security.md", "SECURITY.md"} {
		dc, err := c.Client.Repositories.DownloadContents(c.Ctx, c.Owner, c.Repo, fp, &github.RepositoryContentGetOptions{})
		if err != nil {
			continue
		}
		dc.Close()
		c.Logf("security policy found: %s", fp)
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
