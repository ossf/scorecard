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
	"strings"

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
)

var artifactExtensions = []string{".asc", ".minisig", ".sig", ".sign"}

const releaseLookBack = 5

// SignedReleases applies the score policy for the Signed-Releases check.
func SignedReleases(name string, dl checker.DetailLogger, r *checker.SignedReleasesData) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	totalReleases := 0
	totalSigned := 0
	totalSigstoreSigned := 0

	for _, release := range r.Releases {
		if len(release.Assets) == 0 {
			continue
		}

		dl.Debug(&checker.LogMessage{
			Text: fmt.Sprintf("GitHub release found: %s", release.Tag),
		})

		totalReleases++
		signed := false
		sigstoreSigned := false

		for _, asset := range release.Assets {
			// Check if signed with extension
			if !signed {
				for _, suffix := range artifactExtensions {
					if strings.HasSuffix(asset.Name, suffix) {
						dl.Info(&checker.LogMessage{
							Path: asset.URL,
							Type: checker.FileTypeURL,
							Text: fmt.Sprintf("signed release artifact: %s", asset.Name),
						})
						signed = true
						totalSigned++
						break
					}
				}
			}
			// Check if signed with rekor
			if !sigstoreSigned {
				if len(asset.RekorEntries) != 0 {
					// TODO: Do we want to check for a keyless signature? Check if the entry
					// contains a Fulcio-issued code-signing certificate.
					sigstoreSigned = true
					totalSigstoreSigned++
				}
			}
			if sigstoreSigned && signed {
				break
			}
		}

		if !signed {
			dl.Warn(&checker.LogMessage{
				Path: release.URL,
				Type: checker.FileTypeURL,
				Text: fmt.Sprintf("release artifact %s not signed", release.Tag),
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

	reason := fmt.Sprintf("%d out of %d artifacts are signed, %d artifacts are sigstore signed", totalSigned, totalReleases, totalSigstoreSigned)
	return checker.CreateProportionalScoreResult(name, reason, totalSigned, totalReleases)
}
