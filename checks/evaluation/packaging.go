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
	"github.com/ossf/scorecard/v5/checker"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/packagedWithAutomatedWorkflow"
	"github.com/ossf/scorecard/v5/probes/publishesWithOIDC"
)

// credentialPublishingScore leaves room above a long-lived token so a project
// can still get credit for publishing from CI, and a higher score for trusted publishing.
const credentialPublishingScore = 8

// Packaging applies the score policy for the Packaging check.
func Packaging(name string,
	findings []finding.Finding,
	dl checker.DetailLogger,
) checker.CheckResult {
	expectedProbes := []string{
		packagedWithAutomatedWorkflow.Probe,
		publishesWithOIDC.Probe,
	}

	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	packaged := false
	credential := false
	trusted := false
	for i := range findings {
		f := &findings[i]
		var logLevel checker.DetailType
		switch f.Probe {
		case packagedWithAutomatedWorkflow.Probe:
			switch f.Outcome {
			case finding.OutcomeTrue:
				packaged = true
				logLevel = checker.DetailInfo
			case finding.OutcomeFalse:
				logLevel = checker.DetailWarn
			default:
				logLevel = checker.DetailDebug
			}
		case publishesWithOIDC.Probe:
			switch f.Outcome {
			case finding.OutcomeTrue:
				trusted = true
				logLevel = checker.DetailInfo
			case finding.OutcomeFalse:
				credential = true
				logLevel = checker.DetailWarn
			default:
				logLevel = checker.DetailDebug
			}
		default:
			e := sce.WithMessage(sce.ErrScorecardInternal, "unknown probe results")
			return checker.CreateRuntimeErrorResult(name, e)
		}
		checker.LogFinding(dl, f, logLevel)
	}

	if !packaged {
		return checker.CreateInconclusiveResult(name, "packaging workflow not detected")
	}
	if credential {
		return checker.CreateResultWithScore(name, "publishing workflow uses a long-lived credential", credentialPublishingScore)
	}
	if trusted {
		return checker.CreateMaxScoreResult(name, "trusted publishing workflow detected")
	}
	return checker.CreateMaxScoreResult(name, "packaging workflow detected")
}
