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
	"github.com/ossf/scorecard/v4/remediation"
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
)

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
	//nolint:errcheck
	remediationMetadata, _ := remediation.New(c)

	for i := range r.Dependencies {
		rr := r.Dependencies[i]
		if rr.Location == nil {
			if rr.Msg == nil {
				e := sce.WithMessage(sce.ErrScorecardInternal, "empty File field")
				return checker.CreateRuntimeErrorResult(name, e)
			}
			dl.Debug(&checker.LogMessage{
				Text: *rr.Msg,
			})
			continue
		}
		if rr.Msg != nil {
			dl.Debug(&checker.LogMessage{
				Path:      rr.Location.Path,
				Type:      rr.Location.Type,
				Offset:    rr.Location.Offset,
				EndOffset: rr.Location.EndOffset,
				Text:      *rr.Msg,
				Snippet:   rr.Location.Snippet,
			})
			continue
		}
		if rr.Pinned == nil {
			dl.Debug(&checker.LogMessage{
				Path:      rr.Location.Path,
				Type:      rr.Location.Type,
				Offset:    rr.Location.Offset,
				EndOffset: rr.Location.EndOffset,
				Text:      fmt.Sprintf("%s has empty Pinned field", rr.Type),
				Snippet:   rr.Location.Snippet,
			})
			continue
		}
		if !*rr.Pinned {
			dl.Warn(&checker.LogMessage{
				Path:        rr.Location.Path,
				Type:        rr.Location.Type,
				Offset:      rr.Location.Offset,
				EndOffset:   rr.Location.EndOffset,
				Text:        generateText(&rr),
				Snippet:     rr.Location.Snippet,
				Remediation: generateRemediation(remediationMetadata, &rr),
			})
		}
		// Update the pinning status.
		updatePinningResults(&rr, &wp, pr)
	}

	// Generate scores and Info results.
	var scores []checker.ProportionalScoreWeighted
	// Go through all dependency types
	// GitHub Actions need to be handled separately since they are not in pr
	scores = append(scores, createScoreForGitHubActionsWorkflow(&wp)...)
	// Only exisiting dependencies will be found in pr
	// We will only score the ecosystem if there are dependencies
	// This results in only existing ecosystems being included in the final score
	for t := range pr {
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

func generateRemediation(remediationMd *remediation.RemediationMetadata, rr *checker.Dependency) *rule.Remediation {
	switch rr.Type {
	case checker.DependencyUseTypeGHAction:
		return remediationMd.CreateWorkflowPinningRemediation(rr.Location.Path)
	case checker.DependencyUseTypeDockerfileContainerImage:
		return remediation.CreateDockerfilePinningRemediation(rr, remediation.CraneDigester{})
	default:
		return nil
	}
}

func updatePinningResults(rr *checker.Dependency,
	wp *worklowPinningResult, pr map[checker.DependencyUseType]pinnedResult,
) {
	if rr.Type == checker.DependencyUseTypeGHAction {
		// Note: `Snippet` contains `action/name@xxx`, so we cna use it to infer
		// if it's a GitHub-owned action or not.
		gitHubOwned := fileparser.IsGitHubOwnedAction(rr.Location.Snippet)
		addWorkflowPinnedResult(rr, wp, gitHubOwned)
		return
	}

	// Update other result types.
	p := pr[rr.Type]
	addPinnedResult(rr, &p)
	pr[rr.Type] = p
}

func generateText(rr *checker.Dependency) string {
	if rr.Type == checker.DependencyUseTypeGHAction {
		// Check if we are dealing with a GitHub action or a third-party one.
		gitHubOwned := fileparser.IsGitHubOwnedAction(rr.Location.Snippet)
		owner := generateOwnerToDisplay(gitHubOwned)
		return fmt.Sprintf("%s %s not pinned by hash", owner, rr.Type)
	}

	return fmt.Sprintf("%s not pinned by hash", rr.Type)
}

func generateOwnerToDisplay(gitHubOwned bool) string {
	if gitHubOwned {
		return "GitHub-owned"
	}
	return "third-party"
}

func addPinnedResult(rr *checker.Dependency, r *pinnedResult) {
	if *rr.Pinned {
		r.pinned += 1
	}
	r.total += 1
}

func addWorkflowPinnedResult(rr *checker.Dependency, w *worklowPinningResult, isGitHub bool) {
	if isGitHub {
		addPinnedResult(rr, &w.gitHubOwned)
	} else {
		addPinnedResult(rr, &w.thirdParties)
	}
}

func createScoreForGitHubActionsWorkflow(wp *worklowPinningResult) []checker.ProportionalScoreWeighted {
	if wp.gitHubOwned.total == 0 && wp.thirdParties.total == 0 {
		return []checker.ProportionalScoreWeighted{}
	}
	if wp.gitHubOwned.total != 0 && wp.thirdParties.total != 0 {
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
		return []checker.ProportionalScoreWeighted{
			{
				Success: wp.gitHubOwned.pinned,
				Total:   wp.gitHubOwned.total,
				Weight:  normalWeight,
			},
		}
	}
	return []checker.ProportionalScoreWeighted{
		{
			Success: wp.thirdParties.pinned,
			Total:   wp.thirdParties.total,
			Weight:  normalWeight,
		},
	}
}
