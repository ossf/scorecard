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
	"github.com/ossf/scorecard/v4/rule"
)

type pinnedResult struct {
	pinned int
	total  int
}

// Structure to host information about pinned github
// or third party dependencies.
type worklowPinningResult struct {
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

	// depTypeKey is the Values map key used to fetch the dependency type.
	depTypeKey = "dependencyType"
)

var (
	dependencyTypes = map[checker.DependencyUseType]int{
		checker.DependencyUseTypeGHAction:                 0,
		checker.DependencyUseTypeDockerfileContainerImage: 1,
		checker.DependencyUseTypeDownloadThenRun:          2,
		checker.DependencyUseTypeGoCommand:                3,
		checker.DependencyUseTypeChocoCommand:             4,
		checker.DependencyUseTypeNpmCommand:               5,
		checker.DependencyUseTypePipCommand:               6,
		checker.DependencyUseTypeNugetCommand:             7,
	}
	intToDepType = map[int]checker.DependencyUseType{
		0: checker.DependencyUseTypeGHAction,
		1: checker.DependencyUseTypeDockerfileContainerImage,
		2: checker.DependencyUseTypeDownloadThenRun,
		3: checker.DependencyUseTypeGoCommand,
		4: checker.DependencyUseTypeChocoCommand,
		5: checker.DependencyUseTypeNpmCommand,
		6: checker.DependencyUseTypePipCommand,
		7: checker.DependencyUseTypeNugetCommand,
	}
)

func ruleRemToProbeRem(rem *rule.Remediation) *probe.Remediation {
	return &probe.Remediation{
		Patch:    rem.Patch,
		Text:     rem.Text,
		Markdown: rem.Markdown,
		Effort:   probe.RemediationEffort(rem.Effort),
	}
}

func probeRemToRuleRem(rem *probe.Remediation) *rule.Remediation {
	return &rule.Remediation{
		Patch:    rem.Patch,
		Text:     rem.Text,
		Markdown: rem.Markdown,
		Effort:   rule.RemediationEffort(rem.Effort),
	}
}

func dependenciesToFindings(r *checker.PinningDependenciesData) ([]finding.Finding, error) {
	findings := make([]finding.Finding, 0)

	for i := range r.ProcessingErrors {
		e := r.ProcessingErrors[i]
		f := finding.Finding{
			Message:  generateTextIncompleteResults(e),
			Location: &e.Location,
			Outcome:  finding.OutcomeNotAvailable,
		}
		findings = append(findings, f)
	}

	for i := range r.Dependencies {
		rr := r.Dependencies[i]
		if rr.Location == nil {
			if rr.Msg == nil {
				e := sce.WithMessage(sce.ErrScorecardInternal, "empty File field")
				return findings, e
			}
			f := &finding.Finding{
				Probe:   "",
				Outcome: finding.OutcomeNotApplicable,
				Message: *rr.Msg,
			}
			findings = append(findings, *f)
			continue
		}
		if rr.Msg != nil {
			loc := &finding.Location{
				Type:      rr.Location.Type,
				Path:      rr.Location.Path,
				LineStart: &rr.Location.Offset,
				LineEnd:   &rr.Location.EndOffset,
				Snippet:   &rr.Location.Snippet,
			}
			f := &finding.Finding{
				Probe:    "",
				Outcome:  finding.OutcomeNotApplicable,
				Message:  *rr.Msg,
				Location: loc,
			}
			findings = append(findings, *f)
			continue
		}
		if rr.Pinned == nil {
			loc := &finding.Location{
				Type:      rr.Location.Type,
				Path:      rr.Location.Path,
				LineStart: &rr.Location.Offset,
				LineEnd:   &rr.Location.EndOffset,
				Snippet:   &rr.Location.Snippet,
			}
			f := &finding.Finding{
				Probe:    "",
				Outcome:  finding.OutcomeNotApplicable,
				Message:  fmt.Sprintf("%s has empty Pinned field", rr.Type),
				Location: loc,
			}
			findings = append(findings, *f)
			continue
		}
		if !*rr.Pinned {
			loc := &finding.Location{
				Type:      rr.Location.Type,
				Path:      rr.Location.Path,
				LineStart: &rr.Location.Offset,
				LineEnd:   &rr.Location.EndOffset,
				Snippet:   &rr.Location.Snippet,
			}
			f := &finding.Finding{
				Probe:    "",
				Outcome:  finding.OutcomeNegative,
				Message:  generateTextUnpinned(&rr),
				Location: loc,
			}
			if rr.Remediation != nil {
				f.Remediation = ruleRemToProbeRem(rr.Remediation)
			}
			f = f.WithValues(map[string]int{
				depTypeKey: dependencyTypes[rr.Type],
			})
			findings = append(findings, *f)
		} else {
			loc := &finding.Location{
				Type:      rr.Location.Type,
				Path:      rr.Location.Path,
				LineStart: &rr.Location.Offset,
				LineEnd:   &rr.Location.EndOffset,
				Snippet:   &rr.Location.Snippet,
			}
			f := &finding.Finding{
				Probe:    "",
				Outcome:  finding.OutcomePositive,
				Location: loc,
			}
			f = f.WithValues(map[string]int{
				depTypeKey: dependencyTypes[rr.Type],
			})
			findings = append(findings, *f)
		}
	}
	return findings, nil
}

// PinningDependencies applies the score policy for the Pinned-Dependencies check.
func PinningDependencies(name string, c *checker.CheckRequest,
	r *checker.PinningDependenciesData,
) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	var wp worklowPinningResult
	pr := make(map[checker.DependencyUseType]pinnedResult)
	dl := c.Dlogger

	findings, err := dependenciesToFindings(r)
	if err != nil {
		return checker.CreateRuntimeErrorResult(name, err)
	}

	for i := range findings {
		f := findings[i]
		switch f.Outcome {
		case finding.OutcomeNotApplicable:
			if f.Location != nil {
				dl.Debug(&checker.LogMessage{
					Path:      f.Location.Path,
					Type:      f.Location.Type,
					Offset:    *f.Location.LineStart,
					EndOffset: *f.Location.LineEnd,
					Text:      f.Message,
					Snippet:   *f.Location.Snippet,
				})
			} else {
				dl.Debug(&checker.LogMessage{
					Text: f.Message,
				})
			}
			continue
		case finding.OutcomeNegative:
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
		case finding.OutcomeNotAvailable:
			dl.Info(&checker.LogMessage{
				Finding: &f,
			})
			continue
		default:
			// ignore
		}
		updatePinningResults(intToDepType[f.Values[depTypeKey]],
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
	wp *worklowPinningResult, pr map[checker.DependencyUseType]pinnedResult,
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

func generateTextUnpinned(rr *checker.Dependency) string {
	if rr.Type == checker.DependencyUseTypeGHAction {
		// Check if we are dealing with a GitHub action or a third-party one.
		gitHubOwned := fileparser.IsGitHubOwnedAction(rr.Location.Snippet)
		owner := generateOwnerToDisplay(gitHubOwned)
		return fmt.Sprintf("%s not pinned by hash", owner)
	}

	return fmt.Sprintf("%s not pinned by hash", rr.Type)
}

func generateTextIncompleteResults(e checker.ElementError) string {
	return fmt.Sprintf("Possibly incomplete results: %s", e.Err)
}

func generateOwnerToDisplay(gitHubOwned bool) string {
	if gitHubOwned {
		return fmt.Sprintf("GitHub-owned %s", checker.DependencyUseTypeGHAction)
	}
	return fmt.Sprintf("third-party %s", checker.DependencyUseTypeGHAction)
}

func addPinnedResult(outcome finding.Outcome, r *pinnedResult) {
	if outcome == finding.OutcomePositive {
		r.pinned += 1
	}
	r.total += 1
}

func addWorkflowPinnedResult(outcome finding.Outcome, w *worklowPinningResult, isGitHub bool) {
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

func createScoreForGitHubActionsWorkflow(wp *worklowPinningResult, dl checker.DetailLogger,
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
