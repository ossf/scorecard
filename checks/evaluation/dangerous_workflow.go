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
	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/hasDangerousWorkflowScriptInjection"
	"github.com/ossf/scorecard/v4/probes/hasDangerousWorkflowUntrustedCheckout"
)

// DangerousWorkflow applies the score policy for the DangerousWorkflow check.
func DangerousWorkflow(name string,
	findings []finding.Finding, dl checker.DetailLogger,
) checker.CheckResult {
	expectedProbes := []string{
		hasDangerousWorkflowScriptInjection.Probe,
		hasDangerousWorkflowUntrustedCheckout.Probe,
	}

	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	if !hasWorkflows(findings) {
		return checker.CreateInconclusiveResult(name, "no workflows found")
	}

	if hasDWWithUntrustedCheckout(findings) || hasDWWithScriptInjection(findings) {
		return checker.CreateMinScoreResult(name,
			"dangerous workflow patterns detected")
	}

	return checker.CreateMaxScoreResult(name,
		"no dangerous workflow patterns detected")
}

// Both probes return OutcomeNotApplicable, if there project has no workflows.
func hasWorkflows(findings []finding.Finding) bool {
	for i := range findings {
		f := &findings[i]
		if f.Outcome == finding.OutcomeNotApplicable {
			return false
		}
	}
	return true
}

func hasDWWithUntrustedCheckout(findings []finding.Finding) bool {
	for i := range findings {
		f := &findings[i]
		if f.Probe == hasDangerousWorkflowUntrustedCheckout.Probe {
			if f.Outcome == finding.OutcomeNegative {
				return true
			}
		}
	}
	return false
}

func hasDWWithScriptInjection(findings []finding.Finding) bool {
	for i := range findings {
		f := &findings[i]
		if f.Probe == hasDangerousWorkflowScriptInjection.Probe {
			if f.Outcome == finding.OutcomeNegative {
				return true
			}
		}
	}
	return false
}
