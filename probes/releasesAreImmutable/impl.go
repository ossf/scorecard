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

package releasesAreImmutable

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
	Probe           = "releasesAreImmutable"
	ReleaseNameKey  = "releaseName"
	releaseLookBack = 5
)

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

		loc := &finding.Location{
			Type: finding.FileTypeURL,
			Path: release.URL,
		}

		if release.IsImmutable {
			f, err := finding.NewWith(fs, Probe,
				fmt.Sprintf("immutable release: %s", release.TagName),
				loc,
				finding.OutcomeTrue)
			if err != nil {
				return nil, Probe, fmt.Errorf("create finding: %w", err)
			}
			f = f.WithValue(ReleaseNameKey, release.TagName)
			findings = append(findings, *f)
			continue
		}

		// Release is not immutable
		f, err := finding.NewWith(fs, Probe,
			fmt.Sprintf("release artifact %s is not immutable", release.TagName),
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
