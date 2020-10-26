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
	"strings"

	"github.com/dlorenc/scorecard/checker"
	"github.com/google/go-github/v32/github"
)

var releaseLookBack int = 5

func init() {
	registerCheck("Signed-Releases", SignedReleases)
}

func SignedReleases(c checker.Checker) checker.CheckResult {
	releases, _, err := c.Client.Repositories.ListReleases(c.Ctx, c.Owner, c.Repo, &github.ListOptions{})
	if err != nil {
		return checker.RetryResult(err)
	}

	totalReleases := 0
	totalSigned := 0
	for _, r := range releases {
		assets, _, err := c.Client.Repositories.ListReleaseAssets(c.Ctx, c.Owner, c.Repo, r.GetID(), &github.ListOptions{})
		if err != nil {
			return checker.RetryResult(err)
		}
		if len(assets) == 0 {
			continue
		}
		totalReleases++
		signed := false
		for _, asset := range assets {
			for _, suffix := range []string{".asc", ".minisig", ".sig"} {
				if strings.HasSuffix(asset.GetName(), suffix) {
					c.Logf("signed release found: %s, url: %s", asset.GetName(), asset.GetURL())
					signed = true
					break
				}
			}
			if signed {
				totalSigned++
				break
			}
		}
		if totalReleases > releaseLookBack {
			break
		}
	}

	if totalReleases == 0 {
		return checker.InconclusiveResult
	}
	return checker.ProportionalResult(totalSigned, totalReleases, 0.8)
}
