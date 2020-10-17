package checks

import (
	"path"

	"github.com/dlorenc/scorecard/checker"
	"github.com/google/go-github/v32/github"
)

func init() {
	registerCheck("Security-Policy", SecurityPolicy)
}

func SecurityPolicy(c checker.Checker) checker.CheckResult {
	for _, securityFile := range []string{"SECURITY.md", "security.md"} {
		for _, dirs := range []string{"", ".github", "docs"} {
			fp := path.Join(dirs, securityFile)
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
	}
	return checker.CheckResult{
		Pass:       false,
		Confidence: 10,
	}
}
