// Copyright 2020 Security Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package checks

import (
	"path"

	"github.com/google/go-github/v32/github"
	"github.com/ossf/scorecard/checker"
)

func init() {
	registerCheck("Security-Policy", SecurityPolicy)
}

func SecurityPolicy(c checker.Checker) checker.CheckResult {
	// check repository for repository-specific policy
	const confidence = 10
	for _, securityFile := range []string{"SECURITY.md", "SECURITY.MD", "security.md", "security.MD"} {
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
				Confidence: confidence,
			}
		}

		// alternatively check repository owner for default community health file
		dc, err := c.Client.Repositories.DownloadContents(c.Ctx, c.Owner, ".github", securityFile, &github.RepositoryContentGetOptions{})
		if err == nil {
			dc.Close()
			c.Logf("security policy found (default community health file): %s", securityFile)
			return checker.CheckResult{
				Pass:       true,
				Confidence: confidence,
			}
		}
	}
	// policy wasn't found, return failure case
	return checker.CheckResult{
		Pass:       false,
		Confidence: confidence,
	}
}
