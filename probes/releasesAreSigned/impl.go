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

package releasesAreSigned

import (
	"embed"
	"fmt"
	"strings"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/checknames"
	"github.com/ossf/scorecard/v5/internal/probes"
	"github.com/ossf/scorecard/v5/probes/internal/utils/uerror"
)

func init() {
	probes.MustRegister(Probe, Run, []checknames.CheckName{checknames.SignedReleases})
}

//go:embed *.yml
var fs embed.FS

const (
	Probe           = "releasesAreSigned"
	ReleaseNameKey  = "releaseName"
	AssetNameKey    = "assetName"
	releaseLookBack = 5
)

var signatureExtensions = []string{".asc", ".minisig", ".sig", ".sign", ".sigstore", ".sigstore.json"}

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	var findings []finding.Finding

	releases := raw.SignedReleasesResults.Releases

	totalReleases := 0
	for releaseIndex, release := range releases {
		if releaseIndex >= releaseLookBack {
			break
		}

		if len(release.Assets) == 0 {
			continue
		}

		totalReleases++
		signed := false
		for j := range release.Assets {
			asset := release.Assets[j]
			for _, suffix := range signatureExtensions {
				if !strings.HasSuffix(asset.Name, suffix) {
					continue
				}
				// Create True Finding
				// with file info
				loc := &finding.Location{
					Type: finding.FileTypeURL,
					Path: asset.URL,
				}
				f, err := finding.NewWith(fs, Probe,
					fmt.Sprintf("signed release artifact: %s", asset.Name),
					loc,
					finding.OutcomeTrue)
				if err != nil {
					return nil, Probe, fmt.Errorf("create finding: %w", err)
				}
				f.Values = map[string]string{
					ReleaseNameKey: release.TagName,
					AssetNameKey:   asset.Name,
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
			finding.OutcomeFalse)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		f = f.WithValue(ReleaseNameKey, release.TagName)
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
