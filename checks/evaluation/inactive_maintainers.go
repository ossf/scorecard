// Copyright 2026 OpenSSF Scorecard Authors
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

	"github.com/ossf/scorecard/v5/checker"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/hasInactiveMaintainers"
)

// InactiveMaintainers applies the score policy for the
// Inactive-Maintainers check.
func InactiveMaintainers(name string,
	findings []finding.Finding,
	dl checker.DetailLogger,
) checker.CheckResult {
	expectedProbes := []string{
		hasInactiveMaintainers.Probe,
	}

	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	// Handle case where no maintainers found
	if len(findings) == 1 && findings[0].Outcome == finding.OutcomeNotApplicable {
		return checker.CreateInconclusiveResult(name,
			"no maintainers with elevated permissions found in the repository")
	}

	// Count active and inactive maintainers
	totalMaintainers := 0
	inactiveMaintainers := 0
	activeMaintainers := 0

	for i := range findings {
		f := &findings[i]
		if f.Probe == hasInactiveMaintainers.Probe {
			switch f.Outcome {
			case finding.OutcomeTrue: // Inactive maintainer
				inactiveMaintainers++
				totalMaintainers++
				dl.Warn(&checker.LogMessage{
					Text: f.Message,
				})
			case finding.OutcomeFalse: // Active maintainer
				activeMaintainers++
				totalMaintainers++
				dl.Info(&checker.LogMessage{
					Text: f.Message,
				})
			case finding.OutcomeNotApplicable,
				finding.OutcomeNotAvailable,
				finding.OutcomeError,
				finding.OutcomeNotSupported:
				// These outcomes are handled separately or ignored
			}
		}
	}

	// If no inactive maintainers, return max score
	if inactiveMaintainers == 0 && totalMaintainers > 0 {
		reason := fmt.Sprintf(
			"all %d maintainer(s) have been active in the "+
				"last 6 months",
			totalMaintainers,
		)
		return checker.CreateMaxScoreResult(name, reason)
	}

	// If all maintainers are inactive, return min score
	if inactiveMaintainers == totalMaintainers {
		reason := fmt.Sprintf(
			"all %d maintainer(s) have been inactive for the "+
				"last 6 months",
			inactiveMaintainers,
		)
		return checker.CreateMinScoreResult(name, reason)
	}

	// Otherwise, score proportionally based on active maintainers
	// Score = 10 * (active maintainers / total maintainers)
	reason := fmt.Sprintf(
		"%d out of %d maintainer(s) have been inactive for the "+
			"last 6 months",
		inactiveMaintainers,
		totalMaintainers,
	)

	return checker.CreateProportionalScoreResult(name, reason,
		activeMaintainers, totalMaintainers)
}
