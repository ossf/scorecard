// Copyright 2021 Security Scorecard Authors
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
	"errors"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/remediation"
)

var errInvalidValue = errors.New("invalid value")

type pinnedResult int

const (
	pinnedUndefined pinnedResult = iota
	pinned
	notPinned
)

// Structure to host information about pinned github
// or third party dependencies.
type worklowPinningResult struct {
	thirdParties pinnedResult
	gitHubOwned  pinnedResult
}

// PinningDependencies applies the score policy for the Pinned-Dependencies check.
func PinningDependencies(name string, dl checker.DetailLogger,
	r *checker.PinningDependenciesData,
) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	var wp worklowPinningResult
	pr := make(map[checker.DependencyUseType]pinnedResult)

	for i := range r.Dependencies {
		rr := r.Dependencies[i]
		if rr.File == nil {
			e := sce.WithMessage(sce.ErrScorecardInternal, "empty File field")
			return checker.CreateRuntimeErrorResult(name, e)
		}

		if rr.Msg != nil {
			dl.Debug(&checker.LogMessage{
				Path:      rr.File.Path,
				Type:      rr.File.Type,
				Offset:    rr.File.Offset,
				EndOffset: rr.File.EndOffset,
				Text:      *rr.Msg,
				Snippet:   rr.File.Snippet,
			})
		} else {
			// Warn for inpinned dependency.
			text, err := generateText(&rr)
			if err != nil {
				return checker.CreateRuntimeErrorResult(name, err)
			}
			dl.Warn(&checker.LogMessage{
				Path:        rr.File.Path,
				Type:        rr.File.Type,
				Offset:      rr.File.Offset,
				EndOffset:   rr.File.EndOffset,
				Text:        text,
				Snippet:     rr.File.Snippet,
				Remediation: generateRemediation(&rr),
			})

			// Update the pinning status.
			updatePinningResults(&rr, &wp, pr)
		}
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

	// Action script downloads.
	actionScriptScore, err := createReturnForIsGitHubWorkflowScriptFreeOfInsecureDownloads(pr, dl)
	if err != nil {
		return checker.CreateRuntimeErrorResult(name, err)
	}

	// Scores may be inconclusive.
	actionScore = maxScore(0, actionScore)
	dockerFromScore = maxScore(0, dockerFromScore)
	dockerDownloadScore = maxScore(0, dockerDownloadScore)
	scriptScore = maxScore(0, scriptScore)
	actionScriptScore = maxScore(0, actionScriptScore)

	score := checker.AggregateScores(actionScore, dockerFromScore,
		dockerDownloadScore, scriptScore, actionScriptScore)

	if score == checker.MaxResultScore {
		return checker.CreateMaxScoreResult(name, "all dependencies are pinned")
	}

	return checker.CreateProportionalScoreResult(name,
		"dependency not pinned by hash detected", score, checker.MaxResultScore)
}

func generateRemediation(rr *checker.Dependency) *checker.Remediation {
	switch rr.Type {
	case checker.DependencyUseTypeGHAction:
		return remediation.CreateWorkflowPinningRemediation(rr.File.Path)
	}
	return nil
}

func updatePinningResults(rr *checker.Dependency,
	wp *worklowPinningResult, pr map[checker.DependencyUseType]pinnedResult,
) {
	if rr.Type == checker.DependencyUseTypeGHAction {
		gitHubOwned := fileparser.IsGitHubOwnedAction(rr.File.Snippet)
		addWorkflowPinnedResult(wp, false, gitHubOwned)
		return
	}

	// Update other result types.
	var p pinnedResult
	addPinnedResult(&p, false)
	pr[rr.Type] = p
}

func generateText(rr *checker.Dependency) (string, error) {
	if err := validateType(rr.Type); err != nil {
		return "", err
	}
	if rr.Type == checker.DependencyUseTypeGHAction {
		// Check if we are dealing with a GitHub action or a third-party one.
		gitHubOwned := fileparser.IsGitHubOwnedAction(rr.File.Snippet)
		owner := generateOwnerToDisplay(gitHubOwned)
		return fmt.Sprintf("%s %s not pinned by hash", owner, rr.Type), nil
	}

	return fmt.Sprintf("%s not pinned by hash", rr.Type), nil
}

func generateOwnerToDisplay(gitHubOwned bool) string {
	if gitHubOwned {
		return "GitHub-owned"
	}
	return "third-party"
}

func validateType(t checker.DependencyUseType) error {
	switch t {
	case checker.DependencyUseTypeGHAction, checker.DependencyUseTypeDockerfileContainerImage,
		checker.DependencyUseTypeDownloadThenRun, checker.DependencyUseTypeGoCommand,
		checker.DependencyUseTypeChocoCommand, checker.DependencyUseTypeNpmCommand,
		checker.DependencyUseTypePipCommand:
		return nil
	}
	return sce.WithMessage(sce.ErrScorecardInternal,
		fmt.Sprintf("invalid type: '%v'", t))
}

// TODO(laurent): need to support GCB pinning.
//nolint
func maxScore(s1, s2 int) int {
	if s1 > s2 {
		return s1
	}
	return s2
}

// For the 'to' param, true means the file is pinning dependencies (or there are no dependencies),
// false means there are unpinned dependencies.
func addPinnedResult(r *pinnedResult, to bool) {
	// If the result is `notPinned`, we keep it.
	// In other cases, we always update the result.
	if *r == notPinned {
		return
	}

	switch to {
	case true:
		*r = pinned
	case false:
		*r = notPinned
	}
}

func addWorkflowPinnedResult(w *worklowPinningResult, to, isGitHub bool) {
	if isGitHub {
		addPinnedResult(&w.gitHubOwned, to)
	} else {
		addPinnedResult(&w.thirdParties, to)
	}
}

// Create the result for scripts in GH workflows.
func createReturnForIsGitHubWorkflowScriptFreeOfInsecureDownloads(pr map[checker.DependencyUseType]pinnedResult,
	dl checker.DetailLogger,
) (int, error) {
	return createReturnValues(pr, checker.DependencyUseTypeDownloadThenRun,
		"no insecure (not pinned by hash) dependency downloads found in GitHub workflows",
		dl)
}

// Create the result for scripts.
func createReturnForIsShellScriptFreeOfInsecureDownloads(pr map[checker.DependencyUseType]pinnedResult,
	dl checker.DetailLogger,
) (int, error) {
	return createReturnValues(pr, checker.DependencyUseTypeDownloadThenRun,
		"no insecure (not pinned by hash) dependency downloads found in shell scripts",
		dl)
}

// Create the result for docker containers.
func createReturnForIsDockerfilePinned(pr map[checker.DependencyUseType]pinnedResult,
	dl checker.DetailLogger,
) (int, error) {
	return createReturnValues(pr, checker.DependencyUseTypeDockerfileContainerImage,
		"Dockerfile dependencies are pinned",
		dl)
}

// Create the result for docker commands.
func createReturnForIsDockerfileFreeOfInsecureDownloads(pr map[checker.DependencyUseType]pinnedResult,
	dl checker.DetailLogger,
) (int, error) {
	return createReturnValues(pr, checker.DependencyUseTypeDownloadThenRun,
		"no insecure (not pinned by hash) dependency downloads found in Dockerfiles",
		dl)
}

func createReturnValues(pr map[checker.DependencyUseType]pinnedResult,
	t checker.DependencyUseType, infoMsg string,
	dl checker.DetailLogger,
) (int, error) {
	// Note: we don't check if the entry exists,
	// as it will have the default value which is handled in the switch statement.
	//nolint
	r, _ := pr[t]
	switch r {
	default:
		return checker.InconclusiveResultScore, fmt.Errorf("%w: %v", errInvalidValue, r)
	case pinned, pinnedUndefined:
		dl.Info(&checker.LogMessage{
			Text: infoMsg,
		})
		return checker.MaxResultScore, nil
	case notPinned:
		// No logging needed as it's done by the checks.
		return checker.MinResultScore, nil
	}
}

// Create the result.
func createReturnForIsGitHubActionsWorkflowPinned(wp worklowPinningResult, dl checker.DetailLogger) (int, error) {
	return createReturnValuesForGitHubActionsWorkflowPinned(wp,
		fmt.Sprintf("%ss are pinned", checker.DependencyUseTypeGHAction),
		dl)
}

func createReturnValuesForGitHubActionsWorkflowPinned(r worklowPinningResult, infoMsg string,
	dl checker.DetailLogger,
) (int, error) {
	score := checker.MinResultScore

	if r.gitHubOwned != notPinned {
		score += 2
		dl.Info(&checker.LogMessage{
			Type:   checker.FileTypeSource,
			Offset: checker.OffsetDefault,
			Text:   fmt.Sprintf("%s %s", "GitHub-owned", infoMsg),
		})
	}

	if r.thirdParties != notPinned {
		score += 8
		dl.Info(&checker.LogMessage{
			Type:   checker.FileTypeSource,
			Offset: checker.OffsetDefault,
			Text:   fmt.Sprintf("%s %s", "Third-party", infoMsg),
		})
	}

	return score, nil
}
