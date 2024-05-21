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
	"github.com/ossf/scorecard/v5/checker"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/hasOpenSSFBadge"
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
	if f.Outcome == finding.OutcomeFalse {
		text = "no effort to earn an OpenSSF best practices badge detected"
		return checker.CreateMinScoreResult(name, text)
	}

	level, ok := f.Values[hasOpenSSFBadge.LevelKey]
	if !ok {
		return checker.CreateRuntimeErrorResult(name, sce.WithMessage(sce.ErrScorecardInternal, "no badge level present"))
	}
	switch level {
	case hasOpenSSFBadge.GoldLevel:
		score = checker.MaxResultScore
		text = "badge detected: Gold"
	case hasOpenSSFBadge.SilverLevel:
		score = silverScore
		text = "badge detected: Silver"
	case hasOpenSSFBadge.PassingLevel:
		score = passingScore
		text = "badge detected: Passing"
	case hasOpenSSFBadge.InProgressLevel:
		score = inProgressScore
		text = "badge detected: InProgress"
	default:
		return checker.CreateRuntimeErrorResult(name, sce.WithMessage(sce.ErrScorecardInternal, "unsupported badge detected"))
	}

	return checker.CreateResultWithScore(name, text, score)
}
