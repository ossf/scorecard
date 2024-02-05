// Copyright 2020 OpenSSF Scorecard Authors
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

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/blocksDeleteOnBranches"
	"github.com/ossf/scorecard/v4/probes/blocksForcePushOnBranches"
	"github.com/ossf/scorecard/v4/probes/branchProtectionAppliesToAdmins"
	"github.com/ossf/scorecard/v4/probes/branchesAreProtected"
	"github.com/ossf/scorecard/v4/probes/dismissesStaleReviews"
	"github.com/ossf/scorecard/v4/probes/requiresApproversForPullRequests"
	"github.com/ossf/scorecard/v4/probes/requiresCodeOwnersReview"
	"github.com/ossf/scorecard/v4/probes/requiresLastPushApproval"
	"github.com/ossf/scorecard/v4/probes/requiresPRsToChangeCode"
	"github.com/ossf/scorecard/v4/probes/requiresUpToDateBranches"
	"github.com/ossf/scorecard/v4/probes/runsStatusChecksBeforeMerging"
)

const (
	minReviews = 2
	// Points incremented at each level.
	basicLevel                  = 3 // Level 1.
	adminNonAdminReviewLevel    = 3 // Level 2.
	nonAdminContextLevel        = 2 // Level 3.
	nonAdminThoroughReviewLevel = 1 // Level 4.
	adminThoroughReviewLevel    = 1 // Level 5.

)

type scoresInfo struct {
	basic               int
	review              int
	adminReview         int
	context             int
	thoroughReview      int
	adminThoroughReview int
	codeownerReview     int
}

// Maximum score depending on whether admin token is used.
type levelScore struct {
	scores scoresInfo // Score result for a branch.
	maxes  scoresInfo // Maximum possible score for a branch.
}

type tier uint8

const (
	Tier1 tier = iota
	Tier2
	Tier3
	Tier4
	Tier5
)

// BranchProtection runs Branch-Protection check.
func BranchProtection(name string,
	findings []finding.Finding, dl checker.DetailLogger,
) checker.CheckResult {
	expectedProbes := []string{
		blocksDeleteOnBranches.Probe,
		blocksForcePushOnBranches.Probe,
		branchesAreProtected.Probe,
		branchProtectionAppliesToAdmins.Probe,
		dismissesStaleReviews.Probe,
		requiresApproversForPullRequests.Probe,
		requiresCodeOwnersReview.Probe,
		requiresLastPushApproval.Probe,
		requiresUpToDateBranches.Probe,
		runsStatusChecksBeforeMerging.Probe,
		requiresPRsToChangeCode.Probe,
	}

	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	// Create a map branches and whether theyare protected
	// Protected field only indates that the branch matches
	// one `Branch protection rules`. All settings may be disabled,
	// so it does not provide any guarantees.
	protectedBranches := make(map[string]bool)
	for i := range findings {
		f := &findings[i]
		if f.Outcome == finding.OutcomeNotApplicable {
			return checker.CreateInconclusiveResult(name,
				"unable to detect any development/release branches")
		}
		branchName, err := getBranchName(f)
		if err != nil {
			return checker.CreateRuntimeErrorResult(name, err)
		}
		// the order of this switch statement matters.
		switch {
		// Sanity check:
		case f.Probe != branchesAreProtected.Probe:
			continue
		// Sanity check:
		case branchName == "":
			e := sce.WithMessage(sce.ErrScorecardInternal, "probe is missing branch name")
			return checker.CreateRuntimeErrorResult(name, e)
		// Now we can check whether the branch is protected:
		case f.Outcome == finding.OutcomeNegative:
			protectedBranches[branchName] = false
			dl.Warn(&checker.LogMessage{
				Text: fmt.Sprintf("branch protection not enabled for branch '%s'", branchName),
			})
		case f.Outcome == finding.OutcomePositive:
			protectedBranches[branchName] = true
		default:
			continue
		}
	}

	branchScores := make(map[string]*levelScore)

	for i := range findings {
		f := &findings[i]
		if f.Outcome == finding.OutcomeNotApplicable {
			return checker.CreateInconclusiveResult(name, "unable to detect any development/release branches")
		}

		branchName, err := getBranchName(f)
		if err != nil {
			return checker.CreateRuntimeErrorResult(name, err)
		}
		if branchName == "" {
			e := sce.WithMessage(sce.ErrScorecardInternal, "probe is missing branch name")
			return checker.CreateRuntimeErrorResult(name, e)
		}

		if _, ok := branchScores[branchName]; !ok {
			branchScores[branchName] = &levelScore{}
		}

		var score, max int

		doLogging := protectedBranches[branchName]
		switch f.Probe {
		case blocksDeleteOnBranches.Probe, blocksForcePushOnBranches.Probe:
			score, max = deleteAndForcePushProtection(f, doLogging, dl)
			branchScores[branchName].scores.basic += score
			branchScores[branchName].maxes.basic += max

		case dismissesStaleReviews.Probe, branchProtectionAppliesToAdmins.Probe:
			score, max = adminThoroughReviewProtection(f, doLogging, dl)
			branchScores[branchName].scores.adminThoroughReview += score
			branchScores[branchName].maxes.adminThoroughReview += max

		case requiresApproversForPullRequests.Probe:
			// Scorecard evaluation scores twice with this probe:
			// Once if the count is above 0
			// Once if the count is above 2
			score, max = nonAdminThoroughReviewProtection(f, doLogging, dl)
			branchScores[branchName].scores.thoroughReview += score
			branchScores[branchName].maxes.thoroughReview += max

			reviewerWeight := 2
			max = reviewerWeight
			if f.Outcome == finding.OutcomePositive && f.Values["numberOfRequiredReviewers"] > 0 {
				branchScores[branchName].scores.review += reviewerWeight
			}
			branchScores[branchName].maxes.review += max

		case requiresCodeOwnersReview.Probe:
			score, max = codeownerBranchProtection(f, doLogging, dl)
			branchScores[branchName].scores.codeownerReview += score
			branchScores[branchName].maxes.codeownerReview += max

		case requiresUpToDateBranches.Probe, requiresLastPushApproval.Probe,
			requiresPRsToChangeCode.Probe:
			score, max = adminReviewProtection(f, doLogging, dl)
			branchScores[branchName].scores.adminReview += score
			branchScores[branchName].maxes.adminReview += max

		case runsStatusChecksBeforeMerging.Probe:
			score, max = nonAdminContextProtection(f, doLogging, dl)
			branchScores[branchName].scores.context += score
			branchScores[branchName].maxes.context += max
		}
	}

	if len(branchScores) == 0 {
		return checker.CreateInconclusiveResult(name, "unable to detect any development/release branches")
	}

	scores := make([]levelScore, 0, len(branchScores))
	for _, v := range branchScores {
		scores = append(scores, *v)
	}

	score, err := computeFinalScore(scores)
	if err != nil {
		return checker.CreateRuntimeErrorResult(name, err)
	}

	switch score {
	case checker.MinResultScore:
		return checker.CreateMinScoreResult(name,
			"branch protection not enabled on development/release branches")
	case checker.MaxResultScore:
		return checker.CreateMaxScoreResult(name,
			"branch protection is fully enabled on development and all release branches")
	default:
		return checker.CreateResultWithScore(name,
			"branch protection is not maximal on development and all release branches", score)
	}
}

func getBranchName(f *finding.Finding) (string, error) {
	for k := range f.Values {
		if k == "branchProtected" || k == "numberOfRequiredReviewers" {
			continue
		}
		return k, nil
	}
	return "", sce.WithMessage(sce.ErrScorecardInternal, "no branch name found")
}

func sumUpScoreForTier(t tier, scoresData []levelScore) int {
	sum := 0
	for i := range scoresData {
		score := scoresData[i]
		switch t {
		case Tier1:
			sum += score.scores.basic
		case Tier2:
			sum += score.scores.review + score.scores.adminReview
		case Tier3:
			sum += score.scores.context
		case Tier4:
			sum += score.scores.thoroughReview + score.scores.codeownerReview
		case Tier5:
			sum += score.scores.adminThoroughReview
		}
	}
	return sum
}

func logWithDebug(f *finding.Finding, doLogging bool, dl checker.DetailLogger) {
	switch f.Outcome {
	case finding.OutcomeNotAvailable:
		debug(dl, doLogging, f.Message)
	case finding.OutcomePositive:
		info(dl, doLogging, f.Message)
	case finding.OutcomeNegative:
		warn(dl, doLogging, f.Message)
	default:
		// To satisfy linter
	}
}

func logWithoutDebug(f *finding.Finding, doLogging bool, dl checker.DetailLogger) {
	switch f.Outcome {
	case finding.OutcomePositive:
		info(dl, doLogging, f.Message)
	case finding.OutcomeNegative:
		warn(dl, doLogging, f.Message)
	default:
		// To satisfy linter
	}
}

func logInfoOrWarn(f *finding.Finding, doLogging bool, dl checker.DetailLogger) {
	switch f.Outcome {
	case finding.OutcomePositive:
		info(dl, doLogging, f.Message)
	default:
		warn(dl, doLogging, f.Message)
	}
}

func normalizeScore(score, max, level int) float64 {
	if max == 0 {
		return float64(level)
	}
	return float64(score*level) / float64(max)
}

func computeFinalScore(scores []levelScore) (int, error) {
	if len(scores) == 0 {
		return 0, sce.WithMessage(sce.ErrScorecardInternal, "scores are empty")
	}

	score := float64(0)
	maxScore := scores[0].maxes

	// First, check if they all pass the basic (admin and non-admin) checks.
	maxBasicScore := maxScore.basic * len(scores)
	basicScore := sumUpScoreForTier(Tier1, scores)
	score += normalizeScore(basicScore, maxBasicScore, basicLevel)
	if basicScore < maxBasicScore {
		return int(score), nil
	}

	// Second, check the (admin and non-admin) reviews.
	maxReviewScore := maxScore.review * len(scores)
	maxAdminReviewScore := maxScore.adminReview * len(scores)
	adminNonAdminReviewScore := sumUpScoreForTier(Tier2, scores)
	score += normalizeScore(adminNonAdminReviewScore, maxReviewScore+maxAdminReviewScore, adminNonAdminReviewLevel)
	if adminNonAdminReviewScore < maxReviewScore+maxAdminReviewScore {
		return int(score), nil
	}

	// Third, check the use of non-admin context.
	maxContextScore := maxScore.context * len(scores)
	contextScore := sumUpScoreForTier(Tier3, scores)
	score += normalizeScore(contextScore, maxContextScore, nonAdminContextLevel)
	if contextScore < maxContextScore {
		return int(score), nil
	}

	// Fourth, check the thorough non-admin reviews.
	// Also check whether this repo requires codeowner review
	maxThoroughReviewScore := maxScore.thoroughReview * len(scores)
	maxCodeownerReviewScore := maxScore.codeownerReview * len(scores)
	tier4Score := sumUpScoreForTier(Tier4, scores)
	score += normalizeScore(tier4Score, maxThoroughReviewScore+maxCodeownerReviewScore, nonAdminThoroughReviewLevel)
	if tier4Score < maxThoroughReviewScore+maxCodeownerReviewScore {
		return int(score), nil
	}

	// Lastly, check the thorough admin review config.
	// This one is controversial and has usability issues
	// https://github.com/ossf/scorecard/issues/1027, so we may remove it.
	maxAdminThoroughReviewScore := maxScore.adminThoroughReview * len(scores)
	adminThoroughReviewScore := sumUpScoreForTier(Tier5, scores)
	score += normalizeScore(adminThoroughReviewScore, maxAdminThoroughReviewScore, adminThoroughReviewLevel)
	if adminThoroughReviewScore != maxAdminThoroughReviewScore {
		return int(score), nil
	}

	return int(score), nil
}

func info(dl checker.DetailLogger, doLogging bool, desc string, args ...interface{}) {
	if !doLogging {
		return
	}

	dl.Info(&checker.LogMessage{
		Text: fmt.Sprintf(desc, args...),
	})
}

func debug(dl checker.DetailLogger, doLogging bool, desc string, args ...interface{}) {
	if !doLogging {
		return
	}

	dl.Debug(&checker.LogMessage{
		Text: fmt.Sprintf(desc, args...),
	})
}

func warn(dl checker.DetailLogger, doLogging bool, desc string, args ...interface{}) {
	if !doLogging {
		return
	}

	dl.Warn(&checker.LogMessage{
		Text: fmt.Sprintf(desc, args...),
	})
}

func deleteAndForcePushProtection(f *finding.Finding, doLogging bool, dl checker.DetailLogger) (int, int) {
	var score, max int
	logWithoutDebug(f, doLogging, dl)
	if f.Outcome == finding.OutcomePositive {
		score++
	}
	max++

	return score, max
}

func nonAdminContextProtection(f *finding.Finding, doLogging bool, dl checker.DetailLogger) (int, int) {
	var score, max int
	logInfoOrWarn(f, doLogging, dl)
	if f.Outcome == finding.OutcomePositive {
		score++
	}
	max++
	return score, max
}

func adminReviewProtection(f *finding.Finding, doLogging bool, dl checker.DetailLogger) (int, int) {
	var score, max int
	if f.Outcome == finding.OutcomePositive {
		score++
	}
	switch f.Probe {
	case requiresLastPushApproval.Probe,
		requiresUpToDateBranches.Probe:
		logWithDebug(f, doLogging, dl)
		if f.Outcome != finding.OutcomeNotAvailable {
			max++
		}
	default:
		logInfoOrWarn(f, doLogging, dl)
		max++
	}
	return score, max
}

func adminThoroughReviewProtection(f *finding.Finding, doLogging bool, dl checker.DetailLogger) (int, int) {
	var score, max int

	logWithDebug(f, doLogging, dl)
	if f.Outcome == finding.OutcomePositive {
		score++
	}
	if f.Outcome != finding.OutcomeNotAvailable {
		max++
	}
	return score, max
}

func nonAdminThoroughReviewProtection(f *finding.Finding, doLogging bool, dl checker.DetailLogger) (int, int) {
	var score, max int
	if f.Outcome == finding.OutcomePositive {
		if f.Values["numberOfRequiredReviewers"] >= minReviews {
			info(dl, doLogging, f.Message)
			score++
		} else {
			warn(dl, doLogging, f.Message)
		}
	} else if f.Outcome == finding.OutcomeNegative {
		warn(dl, doLogging, f.Message)
	}
	max++
	return score, max
}

func codeownerBranchProtection(f *finding.Finding, doLogging bool, dl checker.DetailLogger) (int, int) {
	var score, max int
	if f.Outcome == finding.OutcomePositive {
		info(dl, doLogging, f.Message)
		if f.Values["CodeownersFiles"] == 0 {
			warn(dl, doLogging, "codeowners branch protection is being ignored - but no codeowners file found in repo")
		} else {
			score++
		}
	} else {
		warn(dl, doLogging, f.Message)
	}
	max++
	return score, max
}
