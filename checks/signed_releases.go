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
	"errors"
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/ossf/scorecard/checker"
)

const (
	signedReleasesStr   = "Signed-Releases"
	releaseLookBackDays = 5
)

var ErrorNoReleases = errors.New("no releases found")

//nolint:gochecknoinits
func init() {
	registerCheck(signedReleasesStr, SignedReleases)
}

func SignedReleases(c *checker.CheckRequest) checker.CheckResult {
	releases, _, err := c.Client.Repositories.ListReleases(c.Ctx, c.Owner, c.Repo, &github.ListOptions{})
	if err != nil {
		return checker.MakeRetryResult(signedReleasesStr, err)
	}

	artifactExtensions := []string{".asc", ".minisig", ".sig"}

	totalReleases := 0
	totalSigned := 0
	for _, r := range releases {
		assets, _, err := c.Client.Repositories.ListReleaseAssets(c.Ctx, c.Owner, c.Repo, r.GetID(), &github.ListOptions{})
		if err != nil {
			return checker.MakeRetryResult(signedReleasesStr, err)
		}
		if len(assets) == 0 {
			continue
		}
		c.Logf("release found: %s", r.GetTagName())
		totalReleases++
		signed := false
		for _, asset := range assets {
			for _, suffix := range artifactExtensions {
				if strings.HasSuffix(asset.GetName(), suffix) {
					c.Logf("signed release artifact found: %s, url: %s", asset.GetName(), asset.GetURL())
					signed = true
					break
				}
			}
			if signed {
				totalSigned++
				break
			}
		}
		if !signed {
			c.Logf("!! release %s has no signed artifacts", r.GetTagName())
		}
		if totalReleases > releaseLookBackDays {
			break
		}
	}

	if totalReleases == 0 {
		c.Logf("no releases found")
		return checker.MakeInconclusiveResult(signedReleasesStr, ErrorNoReleases)
	}

	c.Logf("found signed artifacts for %d out of %d releases", totalSigned, totalReleases)
	return checker.MakeProportionalResult(signedReleasesStr, totalSigned, totalReleases, 0.8)
}
