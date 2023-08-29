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
		} else if !*rr.Pinned {
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
	// GitHub actions.
	actionScore, err := createReturnForIsGitHubActionsWorkflowPinned(wp, dl)
	if err != nil {
		return checker.CreateRuntimeErrorResult(name, err)
	}

	// Docker files.
	dockerFromScore, err := createReturnForIsDockerfilePinned(pr, dl)
	if err != nil {
		return checker.CreateRuntimeErrorResult(name, err)
	}

	// Docker downloads.
	dockerDownloadScore, err := createReturnForIsDockerfileFreeOfInsecureDownloads(pr, dl)
	if err != nil {
		return checker.CreateRuntimeErrorResult(name, err)
	}

	// Script downloads.
	scriptScore, err := createReturnForIsShellScriptFreeOfInsecureDownloads(pr, dl)
	if err != nil {
		return checker.CreateRuntimeErrorResult(name, err)
	}

	// Pip installs.
	pipScore, err := createReturnForIsPipInstallPinned(pr, dl)
	if err != nil {
		return checker.CreateRuntimeErrorResult(name, err)
	}

	// Npm installs.
	npmScore, err := createReturnForIsNpmInstallPinned(pr, dl)
	if err != nil {
		return checker.CreateRuntimeErrorResult(name, err)
	}

	// Go installs.
	goScore, err := createReturnForIsGoInstallPinned(pr, dl)
	if err != nil {
		return checker.CreateRuntimeErrorResult(name, err)
	}

	// If no dependencies of an ecossystem are found, it results in an inconclusive score.
	// We filter out inconclusive scores so only applicable ecossystems are considered in
	// the final score.
	scores := []int{actionScore, dockerFromScore, dockerDownloadScore, scriptScore, pipScore, npmScore, goScore}
	conclusiveScores := []int{}
	for i := range scores {
		if scores[i] != checker.InconclusiveResultScore {
			conclusiveScores = append(conclusiveScores, scores[i])
		}
	}

	if len(conclusiveScores) == 0 {
		return checker.CreateInconclusiveResult(name, "no dependencies found")
	}

	score := checker.AggregateScores(conclusiveScores...)

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

// Create the result for scripts.
func createReturnForIsShellScriptFreeOfInsecureDownloads(pr map[checker.DependencyUseType]pinnedResult,
	dl checker.DetailLogger,
) (int, error) {
	return createReturnValues(pr, checker.DependencyUseTypeDownloadThenRun,
		"no insecure (not pinned by hash) dependency downloads found in shell scripts",
		"no dependency downloads found in shell scripts",
		dl)
}

// Create the result for docker containers.
func createReturnForIsDockerfilePinned(pr map[checker.DependencyUseType]pinnedResult,
	dl checker.DetailLogger,
) (int, error) {
	return createReturnValues(pr, checker.DependencyUseTypeDockerfileContainerImage,
		"Dockerfile dependencies are pinned",
		"no Dockerfile dependencies found",
		dl)
}

// Create the result for docker commands.
func createReturnForIsDockerfileFreeOfInsecureDownloads(pr map[checker.DependencyUseType]pinnedResult,
	dl checker.DetailLogger,
) (int, error) {
	return createReturnValues(pr, checker.DependencyUseTypeDownloadThenRun,
		"no insecure (not pinned by hash) dependency downloads found in Dockerfiles",
		"no dependency downloads found in Dockerfiles",
		dl)
}

// Create the result for pip install commands.
func createReturnForIsPipInstallPinned(pr map[checker.DependencyUseType]pinnedResult,
	dl checker.DetailLogger,
) (int, error) {
	return createReturnValues(pr, checker.DependencyUseTypePipCommand,
		"pip installs are pinned",
		"no pip installs found",
		dl)
}

// Create the result for npm install commands.
func createReturnForIsNpmInstallPinned(pr map[checker.DependencyUseType]pinnedResult,
	dl checker.DetailLogger,
) (int, error) {
	return createReturnValues(pr, checker.DependencyUseTypeNpmCommand,
		"npm installs are pinned",
		"no npm installs found",
		dl)
}

// Create the result for go install commands.
func createReturnForIsGoInstallPinned(pr map[checker.DependencyUseType]pinnedResult,
	dl checker.DetailLogger,
) (int, error) {
	return createReturnValues(pr, checker.DependencyUseTypeGoCommand,
		"go installs are pinned",
		"no go installs found",
		dl)
}

func createReturnValues(pr map[checker.DependencyUseType]pinnedResult,
	t checker.DependencyUseType, maxResultMsg string, inconclusiveResultMsg string,
	dl checker.DetailLogger,
) (int, error) {
	// If there are zero dependencies of this type, it will generate an inconclusive score,
	// so we can disconsider it in the aggregated score.
	// If all dependencies of this type are pinned, it will get a maximum score.
	// If 1 or more dependencies of this type are unpinned, it will get a minimum score.
	//nolint
	r := pr[t]

	switch r.total {
	case 0:
		dl.Info(&checker.LogMessage{
			Text: inconclusiveResultMsg,
		})
		return checker.InconclusiveResultScore, nil
	case r.pinned:
		dl.Info(&checker.LogMessage{
			Text: maxResultMsg,
		})
		return checker.MaxResultScore, nil
	default:
		return checker.MinResultScore, nil
	}
}

// Create the result.
func createReturnForIsGitHubActionsWorkflowPinned(wp worklowPinningResult, dl checker.DetailLogger) (int, error) {
	return createReturnValuesForGitHubActionsWorkflowPinned(wp,
		fmt.Sprintf("%ss are pinned", checker.DependencyUseTypeGHAction),
		fmt.Sprintf("%ss found", checker.DependencyUseTypeGHAction),
		dl)
}

func createReturnValuesForGitHubActionsWorkflowPinned(r worklowPinningResult, maxResultMsg string,
	inconclusiveResultMsg string, dl checker.DetailLogger,
) (int, error) {
	if r.gitHubOwned.total == 0 {
		dl.Info(&checker.LogMessage{
			Text: fmt.Sprintf("%s %s", "no GitHub-owned", inconclusiveResultMsg),
		})
	}

	if r.thirdParties.total == 0 {
		dl.Info(&checker.LogMessage{
			Text: fmt.Sprintf("%s %s", "no Third-party", inconclusiveResultMsg),
		})
	}

	if r.gitHubOwned.total == 0 && r.thirdParties.total == 0 {
		return checker.InconclusiveResultScore, nil
	}

	score := checker.MinResultScore

	if r.gitHubOwned.total == r.gitHubOwned.pinned {
		score += 2
		if r.gitHubOwned.total != 0 {
			dl.Info(&checker.LogMessage{
				Type:   finding.FileTypeSource,
				Offset: checker.OffsetDefault,
				Text:   fmt.Sprintf("%s %s", "GitHub-owned", maxResultMsg),
			})
		}
	}

	if r.thirdParties.total == r.thirdParties.pinned {
		score += 8
		if r.thirdParties.total != 0 {
			dl.Info(&checker.LogMessage{
				Type:   finding.FileTypeSource,
				Offset: checker.OffsetDefault,
				Text:   fmt.Sprintf("%s %s", "Third-party", maxResultMsg),
			})
		}
	}

	return score, nil
}
