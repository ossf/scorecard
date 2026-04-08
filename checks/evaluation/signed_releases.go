// Copyright 2021 OpenSSF Scorecard Authors
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
	"errors"
	"fmt"
	"math"

	"github.com/ossf/scorecard/v5/checker"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/releasesAreImmutable"
	"github.com/ossf/scorecard/v5/probes/releasesAreSigned"
	"github.com/ossf/scorecard/v5/probes/releasesHaveProvenance"
	"github.com/ossf/scorecard/v5/probes/releasesHaveVerifiedProvenance"
)

var errNoReleaseFound = errors.New("no release found")

// SignedReleases applies the score policy for the Signed-Releases check.
//
//nolint:gocognit // surpressing for now
func SignedReleases(name string,
	findings []finding.Finding, dl checker.DetailLogger,
) checker.CheckResult {
	expectedProbes := []string{
		releasesAreSigned.Probe,
		releasesAreImmutable.Probe,
		releasesHaveProvenance.Probe,
	}

	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	// keep track of releases which have provenance or signatures so we don't log
	// redundant false findings on our second pass through below
	hasProvenance := make(map[string]bool)
	hasSignature := make(map[string]bool)
	hasImmutable := make(map[string]bool)

	// Debug all releases and check for OutcomeNotApplicable
	// All probes have OutcomeNotApplicable in case the project has no
	// releases. Therefore, check for any finding with OutcomeNotApplicable.
	loggedReleases := make([]string, 0)
	for i := range findings {
		f := &findings[i]

		if f.Probe == releasesHaveVerifiedProvenance.Probe {
			continue
		}

		// Debug release name
		if f.Outcome == finding.OutcomeNotApplicable {
			// Generic summary.
			return checker.CreateInconclusiveResult(name, "no releases found")
		}
		releaseName := getReleaseName(f)
		if releaseName == "" {
			// Generic summary.
			return checker.CreateRuntimeErrorResult(name, errNoReleaseFound)
		}

		if !contains(loggedReleases, releaseName) {
			dl.Debug(&checker.LogMessage{
				Text: fmt.Sprintf("GitHub release found: %s", releaseName),
			})
			loggedReleases = append(loggedReleases, releaseName)
		}

		if f.Outcome == finding.OutcomeTrue {
			switch f.Probe {
			case releasesHaveProvenance.Probe:
				hasProvenance[releaseName] = true
			case releasesAreSigned.Probe:
				hasSignature[releaseName] = true
			case releasesAreImmutable.Probe:
				hasImmutable[releaseName] = true
			}
		}
	}

	totalTrue := 0
	releaseMap := make(map[string]int)
	uniqueReleaseTags := make([]string, 0)

	var logLevel checker.DetailType
	for i := range findings {
		f := &findings[i]

		if f.Probe == releasesHaveVerifiedProvenance.Probe {
			continue
		}

		releaseName := getReleaseName(f)
		if releaseName == "" {
			return checker.CreateRuntimeErrorResult(name, errNoReleaseFound)
		}
		if !contains(uniqueReleaseTags, releaseName) {
			uniqueReleaseTags = append(uniqueReleaseTags, releaseName)
		}

		switch f.Outcome {
		case finding.OutcomeTrue:
			logLevel = checker.DetailInfo
			totalTrue++
			switch f.Probe {
			case releasesAreSigned.Probe, releasesAreImmutable.Probe:
				if _, ok := releaseMap[releaseName]; !ok {
					releaseMap[releaseName] = 8
				}
			case releasesHaveProvenance.Probe:
				releaseMap[releaseName] = 10
			}
		case finding.OutcomeFalse:
			logLevel = checker.DetailWarn
			switch f.Probe {
			case releasesAreSigned.Probe:
				if hasProvenance[releaseName] || hasImmutable[releaseName] {
					continue
				}
			case releasesAreImmutable.Probe:
				if hasProvenance[releaseName] || hasSignature[releaseName] {
					continue
				}
			}
		default:
			logLevel = checker.DetailDebug
		}
		checker.LogFinding(dl, f, logLevel)
	}

	if totalTrue == 0 {
		return checker.CreateMinScoreResult(name, "Project has not signed or included provenance with any releases.")
	}

	totalReleases := len(uniqueReleaseTags)

	// TODO, the evaluation code should be the one limiting to 5, not assuming the probes have done it already
	// however there are some ordering issues to consider, so going with the easy way for now
	if totalReleases > 5 {
		err := sce.CreateInternal(sce.ErrScorecardInternal, "too many releases, please report this")
		return checker.CreateRuntimeErrorResult(name, err)
	}
	if totalReleases == 0 {
		// This should not happen in production, but it is useful to have
		// for testing.
		return checker.CreateInconclusiveResult(name, "no releases found")
	}

	score := 0
	for _, s := range releaseMap {
		score += s
	}

	score = int(math.Floor(float64(score) / float64(totalReleases)))
	reason := fmt.Sprintf("%d out of the last %d releases have a total of %d signed artifacts.",
		len(releaseMap), totalReleases, totalTrue)
	return checker.CreateResultWithScore(name, reason, score)
}

func getReleaseName(f *finding.Finding) string {
	var key string
	// these keys should be the same, but might as handle situations when they're not
	switch f.Probe {
	case releasesAreSigned.Probe:
		key = releasesAreSigned.ReleaseNameKey
	case releasesAreImmutable.Probe:
		key = releasesAreImmutable.ReleaseNameKey
	case releasesHaveProvenance.Probe:
		key = releasesHaveProvenance.ReleaseNameKey
	}
	return f.Values[key]
}

func contains(releases []string, release string) bool {
	for _, r := range releases {
		if r == release {
			return true
		}
	}
	return false
}
