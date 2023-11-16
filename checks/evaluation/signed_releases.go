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
	"fmt"
	"math"

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/releasesAreSigned"
	"github.com/ossf/scorecard/v4/probes/releasesHaveProvenance"
)

// SignedReleases applies the score policy for the Signed-Releases check.
//

func SignedReleases(name string,
	findings []finding.Finding, dl checker.DetailLogger,
) checker.CheckResult {
	expectedProbes := []string{
		releasesAreSigned.Probe,
		releasesHaveProvenance.Probe,
	}

	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	// All probes have OutcomeNotApplicable in case the project has no
	// releases. Therefore, check for any finding with OutcomeNotApplicable.
	for i := range findings {
		f := &findings[i]
		if f.Outcome == finding.OutcomeNotApplicable {
			dl.Warn(&checker.LogMessage{
				Text: "no GitHub releases found",
			})
			// Generic summary.
			return checker.CreateInconclusiveResult(name, "no releases found")
		}
	}

	score := 0
	totalPositive := 0
	totalReleases := 0
	checker.LogFindings(findings, dl)

	for i := range findings {
		f := &findings[i]
		if f.Outcome == finding.OutcomePositive {
			switch f.Probe {
			case releasesAreSigned.Probe:
				totalPositive++
				if !releaseAlsoHasProvenance(f, findings) {
					score += 8
				}
				totalReleases = f.Values["totalReleases"]
			case releasesHaveProvenance.Probe:
				totalPositive++
				score += 10
				totalReleases = f.Values["totalReleases"]
			}
		}
	}

	if totalPositive == 0 {
		return checker.CreateMinScoreResult(name, "Project has not signed or included provenance with any releases.")
	}

	if totalReleases == 0 {
		// This should not happen in production, but it is useful to have
		// for testing.
		return checker.CreateInconclusiveResult(name, "no releases found")
	}
	score = int(math.Floor(float64(score) / float64(totalReleases)))
	reason := fmt.Sprintf("%d out of %d artifacts are signed or have provenance", totalPositive, totalReleases)
	return checker.CreateResultWithScore(name, reason, score)
}

func releaseAlsoHasProvenance(f1 *finding.Finding, findings []finding.Finding) bool {
	for i := range findings {
		f2 := &findings[i]
		if f2.Probe == releasesHaveProvenance.Probe && f2.Outcome == finding.OutcomePositive {
			if f1.Values["releaseIndex"] == f2.Values["releaseIndex"] {
				if f1.Values["assetIndex"] == f2.Values["assetIndex"] {
					return true
				}
			}
		}
	}
	return false
}
