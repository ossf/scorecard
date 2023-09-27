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
	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/hasBadgeNotFound"
	"github.com/ossf/scorecard/v4/probes/hasGoldBadge"
	"github.com/ossf/scorecard/v4/probes/hasInProgressBadge"
	"github.com/ossf/scorecard/v4/probes/hasPassingBadge"
	"github.com/ossf/scorecard/v4/probes/hasSilverBadge"
	"github.com/ossf/scorecard/v4/probes/hasUnknownBadge"
)

const (
	silverScore = 7
	// Note: if this value is changed, please update the action's threshold score
	// https://github.com/ossf/scorecard-action/blob/main/policies/template.yml#L61.
	passingScore    = 5
	inProgressScore = 2
	maxScore        = 10
	minScore        = 0
	errScore        = -1
)

// CIIBestPractices applies the score policy for the CIIBestPractices check.
func CIIBestPractices(name string,
	findings []finding.Finding, dl checker.DetailLogger,
) checker.CheckResult {
	expectedProbes := []string{
		hasBadgeNotFound.Probe,
		hasGoldBadge.Probe,
		hasSilverBadge.Probe,
		hasInProgressBadge.Probe,
		hasPassingBadge.Probe,
		hasUnknownBadge.Probe,
	}

	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	var score, numberOfPositives int
	var text string
	for i := range findings {
		f := &findings[i]
		if f.Outcome == finding.OutcomePositive {
			numberOfPositives++
			switch f.Probe {
			case hasBadgeNotFound.Probe:
				score = minScore
				text = "no effort to earn an OpenSSF best practices badge detected"
			case hasInProgressBadge.Probe:
				score = inProgressScore
				text = "badge detected: InProgress"
			case hasPassingBadge.Probe:
				score = passingScore
				text = "badge detected: Passing"
			case hasSilverBadge.Probe:
				score = silverScore
				text = "badge detected: Silver"
			case hasGoldBadge.Probe:
				score = maxScore
				text = "badge detected: Gold"
			case hasUnknownBadge.Probe:
				score = errScore
				text = "unsupported badge detected"
			}
		}
	}

	if numberOfPositives != 1 {
		errText := "invalid probe results: multiple badges detected"
		e := sce.WithMessage(sce.ErrScorecardInternal, errText)
		return checker.CreateRuntimeErrorResult(name, e)
	}

	if score == -1 {
		e := sce.WithMessage(sce.ErrScorecardInternal, text)
		return checker.CreateRuntimeErrorResult(name, e)
	}

	return checker.CreateResultWithScore(name, text, score)
}
