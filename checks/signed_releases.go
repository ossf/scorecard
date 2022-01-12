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

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
)

const (
	// CheckSignedReleases is the registered name for SignedReleases.
	CheckSignedReleases = "Signed-Releases"
	releaseLookBack     = 5
)

var artifactExtensions = []string{".asc", ".minisig", ".sig", ".sign"}

//nolint:gochecknoinits
func init() {
	registerCheck(CheckSignedReleases, SignedReleases)
}

// SignedReleases runs Signed-Releases check.
func SignedReleases(c *checker.CheckRequest) checker.CheckResult {
	releases, err := c.RepoClient.ListReleases()
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("Client.Repositories.ListReleases: %v", err))
		return checker.CreateRuntimeErrorResult(CheckSignedReleases, e)
	}

	totalReleases := 0
	totalSigned := 0
	for _, r := range releases {
		if len(r.Assets) == 0 {
			continue
		}
		c.Dlogger.Debug3(&checker.LogMessage{
			Text: fmt.Sprintf("GitHub release found: %s", r.TagName),
		})
		totalReleases++
		signed := false
		for _, asset := range r.Assets {
			for _, suffix := range artifactExtensions {
				if strings.HasSuffix(asset.Name, suffix) {
					c.Dlogger.Info3(&checker.LogMessage{
						Path: asset.URL,
						Type: checker.FileTypeURL,
						Text: fmt.Sprintf("signed release artifact: %s", asset.Name),
					})
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
			c.Dlogger.Warn3(&checker.LogMessage{
				Path: r.URL,
				Type: checker.FileTypeURL,
				Text: fmt.Sprintf("release artifact %s not signed", r.TagName),
			})
		}
		if totalReleases >= releaseLookBack {
			break
		}
	}

	if totalReleases == 0 {
		c.Dlogger.Warn3(&checker.LogMessage{
			Text: "no GitHub releases found",
		})
		// Generic summary.
		return checker.CreateInconclusiveResult(CheckSignedReleases, "no releases found")
	}

	reason := fmt.Sprintf("%d out of %d artifacts are signed", totalSigned, totalReleases)
	return checker.CreateProportionalScoreResult(CheckSignedReleases, reason, totalSigned, totalReleases)
}
