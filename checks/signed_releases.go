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
	"fmt"
	"strings"

	"github.com/google/go-github/v32/github"

	"github.com/ossf/scorecard/v2/checker"
	sce "github.com/ossf/scorecard/v2/errors"
)

const (
	// CheckSignedReleases is the registered name for SignedReleases.
	CheckSignedReleases = "Signed-Releases"
	releaseLookBack     = 5
)

//nolint:gochecknoinits
func init() {
	registerCheck(CheckSignedReleases, SignedReleases)
}

// SignedReleases runs Signed-Releases check.
func SignedReleases(c *checker.CheckRequest) checker.CheckResult {
	releases, _, err := c.Client.Repositories.ListReleases(c.Ctx, c.Owner, c.Repo, &github.ListOptions{})
	if err != nil {
		e := sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("Client.Repositories.ListReleases: %v", err))
		return checker.CreateRuntimeErrorResult(CheckSignedReleases, e)
	}

	artifactExtensions := []string{".asc", ".minisig", ".sig"}

	totalReleases := 0
	totalSigned := 0
	for _, r := range releases {
		assets, _, err := c.Client.Repositories.ListReleaseAssets(c.Ctx, c.Owner, c.Repo, r.GetID(), &github.ListOptions{})
		if err != nil {
			e := sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("Client.Repositories.ListReleaseAssets: %v", err))
			return checker.CreateRuntimeErrorResult(CheckSignedReleases, e)
		}
		if len(assets) == 0 {
			continue
		}
		c.Dlogger.Debug("GitHub release found: %s", r.GetTagName())
		totalReleases++
		signed := false
		for _, asset := range assets {
			for _, suffix := range artifactExtensions {
				if strings.HasSuffix(asset.GetName(), suffix) {
					c.Dlogger.Info("signed release artifact: %s, url: %s", asset.GetName(), asset.GetURL())
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
			c.Dlogger.Warn("release artifact %s not signed", r.GetTagName())
		}
		if totalReleases >= releaseLookBack {
			break
		}
	}

	if totalReleases == 0 {
		return checker.CreateInconclusiveResult(CheckSignedReleases, "no GitHub releases found")
	}

	reason := fmt.Sprintf("%d out of %d artifacts are signed", totalSigned, totalReleases)
	return checker.CreateProportionalScoreResult(CheckSignedReleases, reason, totalSigned, totalReleases)
}
