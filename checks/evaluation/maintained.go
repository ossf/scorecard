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
	"github.com/ossf/scorecard/v5/probes/hasInactiveMaintainers"
	"github.com/ossf/scorecard/v5/probes/hasRecentCommits"
	"github.com/ossf/scorecard/v5/probes/issueActivityByProjectMember"
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
	// We have 5 unique probes, each should have a finding.
	expectedProbes := []string{
		archived.Probe,
		issueActivityByProjectMember.Probe,
		hasRecentCommits.Probe,
		createdRecently.Probe,
		hasInactiveMaintainers.Probe,
	}

	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	var isArchived, recentlyCreated bool
	var commitsWithinThreshold, numberOfIssuesUpdatedWithinThreshold int
	var activeMaintainers, totalMaintainers int

	for i := range findings {
		f := &findings[i]
		if err := processFinding(f, dl, &isArchived, &recentlyCreated, &commitsWithinThreshold,
			&numberOfIssuesUpdatedWithinThreshold, &activeMaintainers, &totalMaintainers); err != nil {
			return checker.CreateRuntimeErrorResult(name, err)
		}
	}

	if isArchived {
		return checker.CreateMinScoreResult(name, "project is archived")
	}

	if recentlyCreated {
		return checker.CreateMinScoreResult(name,
			"project was created within the last 90 days. Please review its contents carefully")
	}

	baseActivityScore := commitsWithinThreshold + numberOfIssuesUpdatedWithinThreshold
	expectedActivity := activityPerWeek * lookBackDays / daysInOneWeek
	maintainerFactor := calculateMaintainerFactor(activeMaintainers, totalMaintainers)

	baseScore := float64(baseActivityScore) / float64(expectedActivity)
	if baseScore > 1.0 {
		baseScore = 1.0
	}

	finalScore := int(baseScore * maintainerFactor * checker.MaxResultScore)
	if finalScore > checker.MaxResultScore {
		finalScore = checker.MaxResultScore
	}

	reason := buildReasonMessage(commitsWithinThreshold, numberOfIssuesUpdatedWithinThreshold,
		activeMaintainers, totalMaintainers, lookBackDays)

	return checker.CreateResultWithScore(name, reason, finalScore)
}

func processFinding(f *finding.Finding, dl checker.DetailLogger,
	isArchived, recentlyCreated *bool,
	commitsWithinThreshold, numberOfIssuesUpdatedWithinThreshold *int,
	activeMaintainers, totalMaintainers *int,
) error {
	switch f.Outcome {
	case finding.OutcomeTrue:
		return processTrueFinding(f, dl, isArchived, recentlyCreated, commitsWithinThreshold,
			numberOfIssuesUpdatedWithinThreshold, totalMaintainers)
	case finding.OutcomeFalse:
		return processFalseFinding(f, dl, activeMaintainers, totalMaintainers)
	case finding.OutcomeNotApplicable:
		return nil
	default:
		checker.LogFinding(dl, f, checker.DetailDebug)
		return nil
	}
}

func processTrueFinding(f *finding.Finding, dl checker.DetailLogger,
	isArchived, recentlyCreated *bool,
	commitsWithinThreshold, numberOfIssuesUpdatedWithinThreshold, totalMaintainers *int,
) error {
	switch f.Probe {
	case issueActivityByProjectMember.Probe:
		val, err := strconv.Atoi(f.Values[issueActivityByProjectMember.NumIssuesKey])
		if err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
		*numberOfIssuesUpdatedWithinThreshold = val
	case hasRecentCommits.Probe:
		val, err := strconv.Atoi(f.Values[hasRecentCommits.NumCommitsKey])
		if err != nil {
			return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
		*commitsWithinThreshold = val
	case archived.Probe:
		*isArchived = true
		checker.LogFinding(dl, f, checker.DetailWarn)
	case createdRecently.Probe:
		*recentlyCreated = true
		checker.LogFinding(dl, f, checker.DetailWarn)
	case hasInactiveMaintainers.Probe:
		*totalMaintainers++
		checker.LogFinding(dl, f, checker.DetailWarn)
	}
	return nil
}

func processFalseFinding(f *finding.Finding, dl checker.DetailLogger,
	activeMaintainers, totalMaintainers *int,
) error {
	if f.Probe == hasInactiveMaintainers.Probe {
		*activeMaintainers++
		*totalMaintainers++
		checker.LogFinding(dl, f, checker.DetailInfo)
	}
	return nil
}

func calculateMaintainerFactor(activeMaintainers, totalMaintainers int) float64 {
	if totalMaintainers == 0 {
		return 1.0
	}

	activityRatio := float64(activeMaintainers) / float64(totalMaintainers)

	switch {
	case activityRatio == 1.0:
		return 1.0
	case activityRatio >= 0.5:
		return 0.7 + (0.3 * activityRatio)
	default:
		return 0.5 + (0.2 * activityRatio)
	}
}

func buildReasonMessage(commitsWithinThreshold, numberOfIssuesUpdatedWithinThreshold,
	activeMaintainers, totalMaintainers, lookBackDays int,
) string {
	reason := fmt.Sprintf(
		"%d commit(s) and %d issue activity found in the last %d days",
		commitsWithinThreshold, numberOfIssuesUpdatedWithinThreshold, lookBackDays)

	if totalMaintainers > 0 {
		inactiveMaintainers := totalMaintainers - activeMaintainers
		if inactiveMaintainers > 0 {
			reason += fmt.Sprintf(" -- %d of %d maintainer(s) inactive in last 6 months",
				inactiveMaintainers, totalMaintainers)
		} else {
			reason += fmt.Sprintf(" -- all %d maintainer(s) active in last 6 months",
				totalMaintainers)
		}
	}
	return reason
}
