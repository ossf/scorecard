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

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/finding/probe"
	"github.com/ossf/scorecard/v4/probes/pinsDependencies"
	"github.com/ossf/scorecard/v4/rule"
)

type pinnedResult struct {
	pinned int
	total  int
}

// Structure to host information about pinned github
// or third party dependencies.
type workflowPinningResult struct {
	thirdParties pinnedResult
	gitHubOwned  pinnedResult
}

// Weights used for proportional score.
// This defines the priority of pinning a dependency over other dependencies.
// The dependencies from all ecosystems are equally prioritized except
// for GitHub Actions. GitHub Actions can be GitHub-owned or from third-party
// development. The GitHub Actions ecosystem has equal priority compared to other
// ecosystems, but, within GitHub Actions, pinning third-party actions has more
// priority than pinning GitHub-owned actions.
// https://github.com/ossf/scorecard/issues/802
const (
	gitHubOwnedActionWeight int = 2
	thirdPartyActionWeight  int = 8
	normalWeight            int = gitHubOwnedActionWeight + thirdPartyActionWeight
)

func probeRemToRuleRem(rem *probe.Remediation) *rule.Remediation {
	return &rule.Remediation{
		Patch:    rem.Patch,
		Text:     rem.Text,
		Markdown: rem.Markdown,
		Effort:   rule.RemediationEffort(rem.Effort),
	}
}

// PinningDependencies applies the score policy for the Pinned-Dependencies check.
func PinningDependencies(name string,
	findings []finding.Finding,
	dl checker.DetailLogger,
) checker.CheckResult {
	expectedProbes := []string{
		pinsDependencies.Probe,
	}

	if !finding.UniqueProbesEqual(findings, expectedProbes) {
		e := sce.WithMessage(sce.ErrScorecardInternal, "invalid probe results")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	var wp workflowPinningResult
	pr := make(map[checker.DependencyUseType]pinnedResult)

	for i := range findings {
		f := findings[i]
		switch f.Outcome {
		case finding.OutcomeNotApplicable:
			return checker.CreateInconclusiveResult(name, "no dependencies found")
		case finding.OutcomeNotSupported:
			dl.Debug(&checker.LogMessage{
				Finding: &f,
			})
			continue
		case finding.OutcomeNegative:
			// we cant use the finding if we want the remediation to show
			// finding.Remediation are currently suppressed (#3349)
			lm := &checker.LogMessage{
				Path:      f.Location.Path,
				Type:      f.Location.Type,
				Offset:    *f.Location.LineStart,
				EndOffset: *f.Location.LineEnd,
				Text:      f.Message,
				Snippet:   *f.Location.Snippet,
			}

			if f.Remediation != nil {
				lm.Remediation = probeRemToRuleRem(f.Remediation)
			}
			dl.Warn(lm)
		case finding.OutcomeError:
			dl.Info(&checker.LogMessage{
				Finding: &f,
			})
			continue
		default:
			// ignore
		}
		updatePinningResults(checker.DependencyUseType(f.Values[pinsDependencies.DepTypeKey]),
			f.Outcome, f.Location.Snippet,
			&wp, pr)
	}

	// Generate scores and Info results.
	var scores []checker.ProportionalScoreWeighted
	// Go through all dependency types
	// GitHub Actions need to be handled separately since they are not in pr
	scores = append(scores, createScoreForGitHubActionsWorkflow(&wp, dl)...)
	// Only existing dependencies will be found in pr
	// We will only score the ecosystem if there are dependencies
	// This results in only existing ecosystems being included in the final score
	for t := range pr {
		logPinnedResult(dl, pr[t], string(t))
		scores = append(scores, checker.ProportionalScoreWeighted{
			Success: pr[t].pinned,
			Total:   pr[t].total,
			Weight:  normalWeight,
		})
	}

	if len(scores) == 0 {
		return checker.CreateInconclusiveResult(name, "no dependencies found")
	}

	score, err := checker.CreateProportionalScoreWeighted(scores...)
	if err != nil {
		return checker.CreateRuntimeErrorResult(name, err)
	}

	if score == checker.MaxResultScore {
		return checker.CreateMaxScoreResult(name, "all dependencies are pinned")
	}

	return checker.CreateProportionalScoreResult(name,
		"dependency not pinned by hash detected", score, checker.MaxResultScore)
}

func updatePinningResults(dependencyType checker.DependencyUseType,
	outcome finding.Outcome, snippet *string,
	wp *workflowPinningResult, pr map[checker.DependencyUseType]pinnedResult,
) {
	if dependencyType == checker.DependencyUseTypeGHAction {
		// Note: `Snippet` contains `action/name@xxx`, so we can use it to infer
		// if it's a GitHub-owned action or not.
		gitHubOwned := fileparser.IsGitHubOwnedAction(*snippet)
		addWorkflowPinnedResult(outcome, wp, gitHubOwned)
		return
	}

	// Update other result types.
	p := pr[dependencyType]
	addPinnedResult(outcome, &p)
	pr[dependencyType] = p
}

func generateOwnerToDisplay(gitHubOwned bool) string {
	if gitHubOwned {
		return fmt.Sprintf("GitHub-owned %s", checker.DependencyUseTypeGHAction)
	}
	return fmt.Sprintf("third-party %s", checker.DependencyUseTypeGHAction)
}

func addPinnedResult(outcome finding.Outcome, r *pinnedResult) {
	if outcome == finding.OutcomeTrue {
		r.pinned += 1
	}
	r.total += 1
}

func addWorkflowPinnedResult(outcome finding.Outcome, w *workflowPinningResult, isGitHub bool) {
	if isGitHub {
		addPinnedResult(outcome, &w.gitHubOwned)
	} else {
		addPinnedResult(outcome, &w.thirdParties)
	}
}

func logPinnedResult(dl checker.DetailLogger, p pinnedResult, name string) {
	dl.Info(&checker.LogMessage{
		Text: fmt.Sprintf("%3d out of %3d %s dependencies pinned", p.pinned, p.total, name),
	})
}

func createScoreForGitHubActionsWorkflow(wp *workflowPinningResult, dl checker.DetailLogger,
) []checker.ProportionalScoreWeighted {
	if wp.gitHubOwned.total == 0 && wp.thirdParties.total == 0 {
		return []checker.ProportionalScoreWeighted{}
	}
	if wp.gitHubOwned.total != 0 && wp.thirdParties.total != 0 {
		logPinnedResult(dl, wp.gitHubOwned, generateOwnerToDisplay(true))
		logPinnedResult(dl, wp.thirdParties, generateOwnerToDisplay(false))
		return []checker.ProportionalScoreWeighted{
			{
				Success: wp.gitHubOwned.pinned,
				Total:   wp.gitHubOwned.total,
				Weight:  gitHubOwnedActionWeight,
			},
			{
				Success: wp.thirdParties.pinned,
				Total:   wp.thirdParties.total,
				Weight:  thirdPartyActionWeight,
			},
		}
	}
	if wp.gitHubOwned.total != 0 {
		logPinnedResult(dl, wp.gitHubOwned, generateOwnerToDisplay(true))
		return []checker.ProportionalScoreWeighted{
			{
				Success: wp.gitHubOwned.pinned,
				Total:   wp.gitHubOwned.total,
				Weight:  normalWeight,
			},
		}
	}
	logPinnedResult(dl, wp.thirdParties, generateOwnerToDisplay(false))
	return []checker.ProportionalScoreWeighted{
		{
			Success: wp.thirdParties.pinned,
			Total:   wp.thirdParties.total,
			Weight:  normalWeight,
		},
	}
}
