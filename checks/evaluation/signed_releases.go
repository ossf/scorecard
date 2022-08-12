// Copyright 2021 Security Scorecard Authors
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

package evaluation

import (
	"fmt"
	"math"
	"strings"

	"github.com/ossf/scorecard/v4/checker"

	sce "github.com/ossf/scorecard/v4/errors"
)

var (
	signatureExtensions  = []string{".asc", ".minisig", ".sig", ".sign"}
	provenanceExtensions = []string{".intoto.jsonl"}
)

const releaseLookBack = 5

// nolint
// SignedReleases applies the score policy for the Signed-Releases check.
func SignedReleases(name string, dl checker.DetailLogger, r *checker.SignedReleasesData) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	totalReleases := 0
	total := 0
	score := 0
	for _, release := range r.Releases {
		if len(release.Assets) == 0 {
			continue
		}

		dl.Debug(&checker.LogMessage{
			Text: fmt.Sprintf("GitHub release found: %s", release.TagName),
		})

		totalReleases++
		signed := false
		hasProvenance := false

		// Check for provenance.
		for _, asset := range release.Assets {
			for _, suffix := range provenanceExtensions {
				if strings.HasSuffix(asset.Name, suffix) {
					dl.Info(&checker.LogMessage{
						Path: asset.URL,
						Type: checker.FileTypeURL,
						Text: fmt.Sprintf("provenance for release artifact: %s", asset.Name),
					})
					hasProvenance = true
					total++
					break
				}
			}
			if hasProvenance {
				// Assign maximum points.
				score += 10
				break
			}
		}

		if hasProvenance {
			continue
		}

		dl.Warn(&checker.LogMessage{
			Path: release.URL,
			Type: checker.FileTypeURL,
			Text: fmt.Sprintf("release artifact %s does not have provenance", release.TagName),
		})

		// No provenance. Try signatures.
		for _, asset := range release.Assets {
			for _, suffix := range signatureExtensions {
				if strings.HasSuffix(asset.Name, suffix) {
					dl.Info(&checker.LogMessage{
						Path: asset.URL,
						Type: checker.FileTypeURL,
						Text: fmt.Sprintf("signed release artifact: %s", asset.Name),
					})
					signed = true
					total++
					break
				}
			}
			if signed {
				// Assign 8 points.
				score += 8
				break
			}
		}

		if !signed {
			dl.Warn(&checker.LogMessage{
				Path: release.URL,
				Type: checker.FileTypeURL,
				Text: fmt.Sprintf("release artifact %s not signed", release.TagName),
			})
		}
		if totalReleases >= releaseLookBack {
			break
		}
	}

	if totalReleases == 0 {
		dl.Warn(&checker.LogMessage{
			Text: "no GitHub releases found",
		})
		// Generic summary.
		return checker.CreateInconclusiveResult(name, "no releases found")
	}

	score = int(math.Floor(float64(score) / float64(totalReleases)))
	reason := fmt.Sprintf("%d out of %d artifacts are signed or have provenance", total, totalReleases)
	return checker.CreateResultWithScore(name, reason, score)
}
