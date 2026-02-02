// Copyright 2022 OpenSSF Scorecard Authors
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

	"github.com/ossf/scorecard/v5/checker"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/hasOSVVulnerabilities"
	"github.com/ossf/scorecard/v5/probes/releasesDirectDepsAreVulnFree"
)

// Vulnerabilities applies the score policy for the Vulnerabilities check.
// It combines two aspects:
// 1. Current state vulnerabilities (60% weight, max -6 points).
// 2. Release vulnerabilities (40% weight, proportional across releases).
func Vulnerabilities(name string,
	findings []finding.Finding,
	dl checker.DetailLogger,
) checker.CheckResult {
	// We expect findings from hasOSVVulnerabilities probe (always runs)
	// and optionally from releasesDirectDepsAreVulnFree (when releases exist).
	// Validate that we have at least the current vulnerabilities probe.
	hasCurrentVulnProbe := false

	for i := range findings {
		if findings[i].Probe == hasOSVVulnerabilities.Probe {
			hasCurrentVulnProbe = true
			break
		}
	}

	if !hasCurrentVulnProbe {
		e := sce.WithMessage(sce.ErrScorecardInternal, "missing hasOSVVulnerabilities probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	// Separate findings by probe type
	var numCurrentVulns int
	var totalReleases, cleanReleases int

	for i := range findings {
		f := &findings[i]
		switch f.Probe {
		case hasOSVVulnerabilities.Probe:
			if f.Outcome == finding.OutcomeTrue {
				numCurrentVulns++
				checker.LogFinding(dl, f, checker.DetailWarn)
			}
		case releasesDirectDepsAreVulnFree.Probe:
			totalReleases++
			if f.Outcome == finding.OutcomeTrue {
				cleanReleases++
			} else if f.Outcome == finding.OutcomeFalse {
				checker.LogFinding(dl, f, checker.DetailWarn)
			}
		}
	}

	// Calculate weighted score using Option A
	// Current vulnerabilities: 60% weight (6 points max penalty)
	currentComponent := 6.0 - math.Min(float64(numCurrentVulns), 6.0)

	// Release vulnerabilities: 40% weight (4 points max)
	releaseComponent := 4.0
	if totalReleases > 0 {
		releaseComponent = 4.0 * (float64(cleanReleases) / float64(totalReleases))
	}

	score := int(math.Round(currentComponent + releaseComponent))
	if score < checker.MinResultScore {
		score = checker.MinResultScore
	}
	if score > checker.MaxResultScore {
		score = checker.MaxResultScore
	}

	// Build informative reason message
	var reason string
	if totalReleases > 0 {
		reason = fmt.Sprintf(
			"%d current vulnerabilities detected, %d/%d recent releases were free of vulnerabilities at time of release",
			numCurrentVulns, cleanReleases, totalReleases)
	} else {
		reason = fmt.Sprintf("%d existing vulnerabilities detected", numCurrentVulns)
	}

	return checker.CreateResultWithScore(name, reason, score)
}
