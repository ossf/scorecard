// Copyright 2025 OpenSSF Scorecard Authors
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

package releasesHaveAttestation

import (
	"embed"
	"fmt"

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
	Probe           = "releasesHaveAttestation"
	ReleaseNameKey  = "releaseName"
	AssetNameKey    = "assetName"
	releaseLookBack = 5
)

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	var findings []finding.Finding

	releases := raw.SignedReleasesResults.Releases

	for i := range releases {
		if i >= releaseLookBack {
			break
		}
		release := releases[i]
		if len(release.Assets) == 0 {
			continue
		}

		// A release is considered attested if all assets have a digest and
		// each digest has a corresponding GitHub artifact attestation.
		allAttested := true
		for j := range release.Assets {
			asset := release.Assets[j]
			if asset.Digest == "" || !asset.HasAttestation {
				allAttested = false
				break
			}
		}

		if allAttested {
			// All assets in this release have attestations — report the first one as representative.
			asset := release.Assets[0]
			loc := &finding.Location{
				Type: finding.FileTypeURL,
				Path: asset.URL,
			}
			f, err := finding.NewWith(fs, Probe,
				fmt.Sprintf("release artifact %s has attestations for all assets", release.TagName),
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
			continue
		}

		// At least one asset is missing a digest or attestation.
		loc := &finding.Location{
			Type: finding.FileTypeURL,
			Path: release.URL,
		}
		f, err := finding.NewWith(fs, Probe,
			fmt.Sprintf("release artifact %s does not have attestations for all assets", release.TagName),
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
			"no GitHub releases found",
			nil,
			finding.OutcomeNotApplicable)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	}
	return findings, Probe, nil
}
