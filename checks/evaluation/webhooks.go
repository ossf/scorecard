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

	"github.com/ossf/scorecard/v5/checker"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/webhooksUseSecrets"
)

// Webhooks applies the score policy for the Webhooks check.
func Webhooks(name string,
	findings []finding.Finding, dl checker.DetailLogger,
) checker.CheckResult {
	expectedProbes := []string{
		webhooksUseSecrets.Probe,
	}

	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	if len(findings) == 1 && findings[0].Outcome == finding.OutcomeNotApplicable {
		return checker.CreateMaxScoreResult(name, "project does not have webhook")
	}

	var webhooksWithNoSecret int

	totalWebhooks := len(findings)

	for i := range findings {
		f := &findings[i]
		if f.Outcome == finding.OutcomeFalse {
			webhooksWithNoSecret++
		}
	}

	if totalWebhooks == webhooksWithNoSecret {
		return checker.CreateMinScoreResult(name, "no hook(s) have a secret configured")
	}

	if webhooksWithNoSecret == 0 {
		msg := fmt.Sprintf("All %d of the projects webhooks are configured with a secret", totalWebhooks)
		return checker.CreateMaxScoreResult(name, msg)
	}

	msg := fmt.Sprintf("%d out of the projects %d webhooks are configured without a secret",
		webhooksWithNoSecret,
		totalWebhooks)

	return checker.CreateProportionalScoreResult(name,
		msg, totalWebhooks-webhooksWithNoSecret, totalWebhooks)
}
