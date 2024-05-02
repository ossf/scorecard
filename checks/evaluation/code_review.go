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
	"strconv"

	"github.com/ossf/scorecard/v5/checker"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/codeApproved"
)

// TODO(raghavkaul) More partial credit? E.g. approval from non-contributor, discussion liveness,
// number of resolved comments, number of approvers (more eyes on a project).

// CodeReview applies the score policy for the Code-Review check.
func CodeReview(name string, findings []finding.Finding, dl checker.DetailLogger) checker.CheckResult {
	expectedProbes := []string{
		codeApproved.Probe,
	}

	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	for _, f := range findings {
		switch f.Outcome {
		case finding.OutcomeNotApplicable:
			return checker.CreateInconclusiveResult(name, f.Message)
		case finding.OutcomeTrue:
			return checker.CreateMaxScoreResult(name, "all changesets reviewed")
		case finding.OutcomeError:
			return checker.CreateRuntimeErrorResult(name, sce.WithMessage(sce.ErrScorecardInternal, f.Message))
		default:
			approved, err := strconv.Atoi(f.Values[codeApproved.NumApprovedKey])
			if err != nil {
				err = sce.WithMessage(sce.ErrScorecardInternal, "converting approved count: %v")
				return checker.CreateRuntimeErrorResult(name, err)
			}
			total, err := strconv.Atoi(f.Values[codeApproved.NumTotalKey])
			if err != nil {
				err = sce.WithMessage(sce.ErrScorecardInternal, "converting total count: %v")
				return checker.CreateRuntimeErrorResult(name, err)
			}
			return checker.CreateProportionalScoreResult(name, f.Message, approved, total)
		}
	}
	return checker.CreateMaxScoreResult(name, "all changesets reviewed")
}
