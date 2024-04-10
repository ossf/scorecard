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
	"strconv"

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/archived"
	"github.com/ossf/scorecard/v4/probes/createdRecently"
	"github.com/ossf/scorecard/v4/probes/hasRecentCommits"
	"github.com/ossf/scorecard/v4/probes/issueActivityByProjectMember"
)

const (
	lookBackDays    = 90
	activityPerWeek = 1
	daysInOneWeek   = 7
)

// Maintained applies the score policy for the Maintained check.
func Maintained(name string,
	findings []finding.Finding, dl checker.DetailLogger,
) checker.CheckResult {
	// We have 4 unique probes, each should have a finding.
	expectedProbes := []string{
		archived.Probe,
		issueActivityByProjectMember.Probe,
		hasRecentCommits.Probe,
		createdRecently.Probe,
	}

	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	var isArchived, recentlyCreated bool

	var commitsWithinThreshold, numberOfIssuesUpdatedWithinThreshold int
	var err error
	for i := range findings {
		f := &findings[i]
		switch f.Outcome {
		case finding.OutcomeTrue:
			switch f.Probe {
			case issueActivityByProjectMember.Probe:
				numberOfIssuesUpdatedWithinThreshold, err = strconv.Atoi(f.Values[issueActivityByProjectMember.NumIssuesKey])
				if err != nil {
					return checker.CreateRuntimeErrorResult(name, sce.WithMessage(sce.ErrScorecardInternal, err.Error()))
				}
			case hasRecentCommits.Probe:
				commitsWithinThreshold, err = strconv.Atoi(f.Values[hasRecentCommits.NumCommitsKey])
				if err != nil {
					return checker.CreateRuntimeErrorResult(name, sce.WithMessage(sce.ErrScorecardInternal, err.Error()))
				}
			case archived.Probe:
				isArchived = true
				checker.LogFinding(dl, f, checker.DetailWarn)
			case createdRecently.Probe:
				recentlyCreated = true
				checker.LogFinding(dl, f, checker.DetailWarn)
			}
		case finding.OutcomeFalse:
			// both archive and created recently are good if false, and the
			// other probes are informational and dont need logged. But we need
			// to specify the case so it doesn't get logged below at the debug level
		default:
			checker.LogFinding(dl, f, checker.DetailDebug)
		}
	}

	if isArchived {
		return checker.CreateMinScoreResult(name, "project is archived")
	}

	if recentlyCreated {
		return checker.CreateMinScoreResult(name, "project was created in last 90 days. please review its contents carefully")
	}

	return checker.CreateProportionalScoreResult(name, fmt.Sprintf(
		"%d commit(s) and %d issue activity found in the last %d days",
		commitsWithinThreshold, numberOfIssuesUpdatedWithinThreshold, lookBackDays),
		commitsWithinThreshold+numberOfIssuesUpdatedWithinThreshold, activityPerWeek*lookBackDays/daysInOneWeek)
}
