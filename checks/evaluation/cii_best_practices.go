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
	"github.com/ossf/scorecard/v4/probes/hasOpenSSFBadge"
)

const (
	silverScore = 7
	// Note: if this value is changed, please update the action's threshold score
	// https://github.com/ossf/scorecard-action/blob/main/policies/template.yml#L61.
	passingScore    = 5
	inProgressScore = 2
)

// CIIBestPractices applies the score policy for the CIIBestPractices check.
func CIIBestPractices(name string,
	findings []finding.Finding, dl checker.DetailLogger,
) checker.CheckResult {
	expectedProbes := []string{
		hasOpenSSFBadge.Probe,
	}

	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	var score int
	var text string

	if len(findings) != 1 {
		errText := "invalid probe results: multiple findings detected"
		e := sce.WithMessage(sce.ErrScorecardInternal, errText)
		return checker.CreateRuntimeErrorResult(name, e)
	}

	f := &findings[0]
	if f.Outcome == finding.OutcomeNegative {
		text = "no effort to earn an OpenSSF best practices badge detected"
		return checker.CreateMinScoreResult(name, text)
	}
	//nolint:nestif
	if _, hasKey := f.Values[hasOpenSSFBadge.GoldLevel]; hasKey {
		score = checker.MaxResultScore
		text = "badge detected: Gold"
	} else if _, hasKey := f.Values[hasOpenSSFBadge.SilverLevel]; hasKey {
		score = silverScore
		text = "badge detected: Silver"
	} else if _, hasKey := f.Values[hasOpenSSFBadge.PassingLevel]; hasKey {
		score = passingScore
		text = "badge detected: Passing"
	} else if _, hasKey := f.Values[hasOpenSSFBadge.InProgressLevel]; hasKey {
		score = inProgressScore
		text = "badge detected: InProgress"
	} else if _, hasKey := f.Values[hasOpenSSFBadge.UnknownLevel]; hasKey {
		text = "unknown badge detected"
		e := sce.WithMessage(sce.ErrScorecardInternal, text)
		return checker.CreateRuntimeErrorResult(name, e)
	} else {
		text = "unsupported badge detected"
		e := sce.WithMessage(sce.ErrScorecardInternal, text)
		return checker.CreateRuntimeErrorResult(name, e)
	}

	return checker.CreateResultWithScore(name, text, score)
}
