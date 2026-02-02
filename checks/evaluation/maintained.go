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

	"github.com/ossf/scorecard/v5/checker"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/archived"
	"github.com/ossf/scorecard/v5/probes/createdRecently"
	"github.com/ossf/scorecard/v5/probes/hasRecentCommits"
	"github.com/ossf/scorecard/v5/probes/issueActivityByProjectMember"
	"github.com/ossf/scorecard/v5/probes/maintainersRespondToBugIssues"
)

const (
	lookBackDays    = 90
	activityPerWeek = 1
	daysInOneWeek   = 7
)

// penaltyMultiplierInfo calculates the penalty multiplier and message for bug/security response violations.
func penaltyMultiplierInfo(bugResponseViolations, bugResponseEvaluated int) (float64, string) {
	violationPercent := float64(bugResponseViolations) / float64(bugResponseEvaluated) * 100.0
	var multiplier float64
	var penaltyMsg string

	switch {
	case violationPercent > 40.0:
		// Severe: >40% violations, multiply by 0.5
		multiplier = 0.5
		penaltyMsg = fmt.Sprintf(
			". Score reduced by 50%% due to %.1f%% of bug/security issues (%d/%d) "+
				"exceeding 180 days without maintainer response",
			violationPercent, bugResponseViolations, bugResponseEvaluated)
	case violationPercent > 20.0:
		// Moderate: 20-40% violations, multiply by 0.75
		multiplier = 0.75
		penaltyMsg = fmt.Sprintf(
			". Score reduced by 25%% due to %.1f%% of bug/security issues (%d/%d) "+
				"exceeding 180 days without maintainer response",
			violationPercent, bugResponseViolations, bugResponseEvaluated)
	default:
		// Good: <20% violations, no penalty
		multiplier = 1.0
		if bugResponseViolations > 0 {
			penaltyMsg = fmt.Sprintf(". %d/%d bug/security issues had timely maintainer response (%.1f%% compliant)",
				bugResponseEvaluated-bugResponseViolations, bugResponseEvaluated, 100.0-violationPercent)
		} else {
			penaltyMsg = fmt.Sprintf(". All %d bug/security issues had timely maintainer response", bugResponseEvaluated)
		}
	}

	return multiplier, penaltyMsg
}

// Maintained applies the score policy for the Maintained check.
func Maintained(name string,
	findings []finding.Finding, dl checker.DetailLogger,
) checker.CheckResult {
	// We have 5 unique probes, each should have a finding.
	expectedProbes := []string{
		archived.Probe,
		issueActivityByProjectMember.Probe,
		hasRecentCommits.Probe,
		createdRecently.Probe,
		maintainersRespondToBugIssues.Probe,
	}

	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	var isArchived, recentlyCreated bool

	var commitsWithinThreshold, numberOfIssuesUpdatedWithinThreshold int
	var bugResponseEvaluated, bugResponseViolations int
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
			case maintainersRespondToBugIssues.Probe:
				// Issue had maintainer response within 180 days
				bugResponseEvaluated++
			}
		case finding.OutcomeFalse:
			if f.Probe == maintainersRespondToBugIssues.Probe {
				// Issue exceeded 180 days without response (violation)
				bugResponseEvaluated++
				bugResponseViolations++
				checker.LogFinding(dl, f, checker.DetailWarn)
			}
			// both archive and created recently are good if false, and the
			// other probes are informational and dont need logged. But we need
			// to specify the case so it doesn't get logged below at the debug level
		default:
			// OutcomeNotApplicable for maintainersRespondToBugIssues means issue had no tracked labels
			// These are excluded from evaluation
			checker.LogFinding(dl, f, checker.DetailDebug)
		}
	}

	if isArchived {
		return checker.CreateMinScoreResult(name, "project is archived")
	}

	if recentlyCreated {
		return checker.CreateMinScoreResult(name,
			"project was created within the last 90 days. Please review its contents carefully")
	}

	// Calculate base score from activity (commits + issues)
	baseScore := checker.CreateProportionalScoreResult(name, fmt.Sprintf(
		"%d commit(s) and %d issue activity found in the last %d days",
		commitsWithinThreshold, numberOfIssuesUpdatedWithinThreshold, lookBackDays),
		commitsWithinThreshold+numberOfIssuesUpdatedWithinThreshold, activityPerWeek*lookBackDays/daysInOneWeek)

	// Apply penalty multiplier based on bug/security response violations
	if bugResponseEvaluated > 0 {
		multiplier, penaltyMsg := penaltyMultiplierInfo(bugResponseViolations, bugResponseEvaluated)

		// Apply multiplier to score
		finalScore := int(float64(baseScore.Score) * multiplier)
		if finalScore < 0 {
			finalScore = 0
		}
		if finalScore > checker.MaxResultScore {
			finalScore = checker.MaxResultScore
		}

		baseScore.Score = finalScore
		baseScore.Reason += penaltyMsg
	}

	return baseScore
}
