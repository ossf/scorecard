// Copyright 2023 OpenSSF Scorecard Authors
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

//nolint:stylecheck
package releasesAreSigned

import (
	"embed"
	"fmt"
	"strings"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/internal/utils/uerror"
)

//go:embed *.yml
var fs embed.FS

const (
	Probe           = "releasesAreSigned"
	releaseLookBack = 5
)

var signatureExtensions = []string{".asc", ".minisig", ".sig", ".sign"}

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	var findings []finding.Finding

	releases := raw.SignedReleasesResults.Releases

	totalReleases := 0
	for releaseIndex, release := range releases {
		if len(release.Assets) == 0 {
			continue
		}

		if releaseIndex == releaseLookBack {
			break
		}

		totalReleases++
		signed := false
		for assetIndex, asset := range release.Assets {
			for _, suffix := range signatureExtensions {
				if !strings.HasSuffix(asset.Name, suffix) {
					continue
				}
				// Create Positive Finding
				// with file info
				loc := &finding.Location{
					Type: finding.FileTypeURL,
					Path: asset.URL,
				}
				f, err := finding.NewWith(fs, Probe,
					fmt.Sprintf("signed release artifact: %s", asset.Name),
					loc,
					finding.OutcomePositive)
				if err != nil {
					return nil, Probe, fmt.Errorf("create finding: %w", err)
				}
				f.Values = map[string]int{
					"totalReleases": len(releases),
					"releaseIndex":  releaseIndex,
					"assetIndex":    assetIndex,
				}
				findings = append(findings, *f)
				signed = true
				break
			}
			if signed {
				break
			}
		}
		if signed {
			continue
		}

		// Release is not signed
		loc := &finding.Location{
			Type: finding.FileTypeURL,
			Path: release.URL,
		}
		f, err := finding.NewWith(fs, Probe,
			fmt.Sprintf("release artifact %s not signed", release.TagName),
			loc,
			finding.OutcomeNegative)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		f.Values = map[string]int{
			"totalReleases": len(releases),
		}
		findings = append(findings, *f)
	}

	if len(findings) == 0 {
		f, err := finding.NewWith(fs, Probe,
			"no GitHub/GitLab releases found",
			nil,
			finding.OutcomeNotApplicable)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	}
	return findings, Probe, nil
}
