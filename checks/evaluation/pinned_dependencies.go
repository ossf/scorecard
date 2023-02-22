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
	"regexp"
	"strings"

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

// Structure to host information about downloadThenRun
// for Dockerfiles and shell scripts.
type downloadThenRunPinningResult struct {
	dockerfile  pinnedResult
	shellScript pinnedResult
}

// PinningDependencies applies the score policy for the Pinned-Dependencies check.
func PinningDependencies(name string, c *checker.CheckRequest,
	r *checker.PinningDependenciesData,
) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	var wp = worklowPinningResult{
		thirdParties: pinnedResult{
			pinned: 0,
			total:  0,
		},
		gitHubOwned: pinnedResult{
			pinned: 0,
			total:  0,
		},
	}
	var d = downloadThenRunPinningResult{
		dockerfile: pinnedResult{
			pinned: 0,
			total:  0,
		},
		shellScript: pinnedResult{
			pinned: 0,
			total:  0,
		},
	}
	pr := make(map[checker.DependencyUseType]pinnedResult)
	for _, el := range pr {
		el.pinned = 0
		el.total = 0
	}
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
		err := updatePinningResults(&rr, &wp, &d, pr)
		if err != nil {
			return checker.CreateRuntimeErrorResult(name, err)
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
	dockerDownloadScore, err := createReturnForIsDockerfileFreeOfInsecureDownloads(d, dl)
	if err != nil {
		return checker.CreateRuntimeErrorResult(name, err)
	}

	// Script downloads.
	scriptScore, err := createReturnForIsShellScriptFreeOfInsecureDownloads(d, dl)
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

	// Choco installs.
	chocoScore, err := createReturnForIsChocoInstallPinned(pr, dl)
	if err != nil {
		return checker.CreateRuntimeErrorResult(name, err)
	}

	// Go installs.
	goScore, err := createReturnForIsGoInstallPinned(pr, dl)
	if err != nil {
		return checker.CreateRuntimeErrorResult(name, err)
	}

	score := checker.AggregateScores(actionScore, dockerFromScore,
		dockerDownloadScore, scriptScore, pipScore, npmScore, chocoScore, goScore)

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
	wp *worklowPinningResult, d *downloadThenRunPinningResult, pr map[checker.DependencyUseType]pinnedResult,
) error {
	if rr.Type == checker.DependencyUseTypeGHAction {
		// Note: `Snippet` contains `action/name@xxx`, so we cna use it to infer
		// if it's a GitHub-owned action or not.
		gitHubOwned := fileparser.IsGitHubOwnedAction(rr.Location.Snippet)
		addWorkflowPinnedResult(rr, wp, gitHubOwned)
		return nil
	} else if rr.Type == checker.DependencyUseTypeDownloadThenRun {
		return addDownloadThenRunPinnedResult(rr, d)
	}

	// Update other result types.
	var p = pr[rr.Type]
	addPinnedResult(rr, &p)
	pr[rr.Type] = p
	return nil
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

func addDownloadThenRunPinnedResult(rr *checker.Dependency, d *downloadThenRunPinningResult) error {
	// Identify if is Dockerfile
	if strings.Contains(rr.Location.Path, "Dockerfile") {
		addPinnedResult(rr, &d.dockerfile)
		return nil
	}
	// Identify if is shell script
	match, err := regexp.MatchString("(.+)\\.sh", rr.Location.Path)
	if err != nil {
		return fmt.Errorf("error matching shell script regex: %v", err)
	}
	if match {
		addPinnedResult(rr, &d.shellScript)
		return nil
	}
	return nil
}

// Create the result for scripts.
func createReturnForIsShellScriptFreeOfInsecureDownloads(d downloadThenRunPinningResult,
	dl checker.DetailLogger,
) (int, error) {
	return createReturnValues(d.shellScript,
		"no insecure (not pinned by hash) dependency downloads found in shell scripts",
		dl)
}

// Create the result for docker containers.
func createReturnForIsDockerfilePinned(pr map[checker.DependencyUseType]pinnedResult,
	dl checker.DetailLogger,
) (int, error) {
	return createReturnValues(pr[checker.DependencyUseTypeDockerfileContainerImage],
		"Dockerfile container images are pinned",
		dl)
}

// Create the result for docker commands.
func createReturnForIsDockerfileFreeOfInsecureDownloads(d downloadThenRunPinningResult,
	dl checker.DetailLogger,
) (int, error) {
	return createReturnValues(d.dockerfile,
		"no insecure (not pinned by hash) dependency downloads found in Dockerfiles",
		dl)
}

// Create the result for pip install commands.
func createReturnForIsPipInstallPinned(pr map[checker.DependencyUseType]pinnedResult,
	dl checker.DetailLogger,
) (int, error) {
	return createReturnValues(pr[checker.DependencyUseTypePipCommand],
		"Pip installs are pinned",
		dl)
}

// Create the result for npm install commands.
func createReturnForIsNpmInstallPinned(pr map[checker.DependencyUseType]pinnedResult,
	dl checker.DetailLogger,
) (int, error) {
	return createReturnValues(pr[checker.DependencyUseTypeNpmCommand],
		"Npm installs are pinned",
		dl)
}

// Create the result for choco install commands.
func createReturnForIsChocoInstallPinned(pr map[checker.DependencyUseType]pinnedResult,
	dl checker.DetailLogger,
) (int, error) {
	return createReturnValues(pr[checker.DependencyUseTypeChocoCommand],
		"Choco installs are pinned",
		dl)
}

// Create the result for go install commands.
func createReturnForIsGoInstallPinned(pr map[checker.DependencyUseType]pinnedResult,
	dl checker.DetailLogger,
) (int, error) {
	return createReturnValues(pr[checker.DependencyUseTypeGoCommand],
		"Go installs are pinned",
		dl)
}

func createReturnValues(pr pinnedResult,
	infoMsg string,
	dl checker.DetailLogger,
) (int, error) {
	// Note: we don't check if the entry exists,
	// as it will have the default value which is handled in the switch statement.
	//nolint
	var score int
	if pr.total == 0 {
		score = checker.MaxResultScore
	} else {
		score = checker.CreateProportionalScore(pr.pinned, pr.total)
	}
	if score == checker.MaxResultScore {
		dl.Info(&checker.LogMessage{
			Text: infoMsg,
		})
	}

	return score, nil
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
	var gitHubOwnedScore int
	var thirdPartiesScore int
	if r.gitHubOwned.total == 0 {
		gitHubOwnedScore = checker.MaxResultScore
	} else {
		gitHubOwnedScore = checker.CreateProportionalScore(r.gitHubOwned.pinned, r.gitHubOwned.total)
	}
	if r.thirdParties.total == 0 {
		thirdPartiesScore = checker.MaxResultScore
	} else {
		thirdPartiesScore = checker.CreateProportionalScore(r.thirdParties.pinned, r.thirdParties.total)
	}
	const gitHubOwnedWeight = 2
	const thirdPartiesWeight = 8
	score := checker.AggregateScoresWithWeight(map[int]int{
		gitHubOwnedScore:  gitHubOwnedWeight,
		thirdPartiesScore: thirdPartiesWeight,
	})

	if gitHubOwnedScore == checker.MaxResultScore {
		dl.Info(&checker.LogMessage{
			Type:   finding.FileTypeSource,
			Offset: checker.OffsetDefault,
			Text:   fmt.Sprintf("%s %s", "GitHub-owned", infoMsg),
		})
	}
	if thirdPartiesScore == checker.MaxResultScore {
		dl.Info(&checker.LogMessage{
			Type:   finding.FileTypeSource,
			Offset: checker.OffsetDefault,
			Text:   fmt.Sprintf("%s %s", "Third-party", infoMsg),
		})
	}

	return score, nil
}
