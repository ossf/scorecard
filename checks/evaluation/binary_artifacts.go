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
	"path/filepath"
	"regexp"

	semver "github.com/Masterminds/semver/v3"
	"github.com/rhysd/actionlint"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

var (
	gradleWrapperValidationActionRegex             = regexp.MustCompile(`^gradle\/wrapper-validation-action@v?(.+)$`)
	gradleWrapperValidationActionVersionConstraint = mustParseConstraint(`>= 1.0.0`)
)

var errBadArg = errors.New("bad arg")

// gradleWrapperValidatingWorkflowMetadata contains data about a workflow with
// the gradle/wrapper-validation-action being used in a step.
type gradleWrapperValidatingWorkflowMetadata struct {
	ActionVersion *semver.Version
	Path          string
}

// mustParseConstraint attempts parse of semver constraint, panics if fail.
func mustParseConstraint(c string) *semver.Constraints {
	if c, err := semver.NewConstraint(c); err != nil {
		panic(fmt.Errorf("failed to parse constraint: %w", err))
	} else {
		return c
	}
}

// BinaryArtifacts applies the score policy for the Binary-Artifacts check.
func BinaryArtifacts(name string, dl checker.DetailLogger,
	r *checker.BinaryArtifactData, c clients.RepoClient,
) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	// Apply the policy evaluation.
	if r.Files == nil || len(r.Files) == 0 {
		return checker.CreateMaxScoreResult(name, "no binaries found in the repo")
	}

	// Check if gradle-wrapper.jar present
	hasGradleWrappers := false
	removeGradleWrappers := false
	if len(r.Files) > 0 {
		for _, f := range r.Files {
			if filepath.Base(f.Path) == "gradle-wrapper.jar" {
				hasGradleWrappers = true
				break
			}
		}
	}
	if hasGradleWrappers {
		// Gradle wrapper JARs present, so check that they are validated
		ok, msg, err := gradleWrapperValidated(c)
		if err != nil {
			e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
			return checker.CreateRuntimeErrorResult(name, e)
		}
		if ok {
			// It has been confirmed that latest commit has validated JARs!
			// Remove Gradle wrapper JARs from files.
			removeGradleWrappers = true
			dl.Info(&checker.LogMessage{
				Type: checker.FileTypeNone,
				Text: "Successfully validated use of wrapper-validation-action Action.",
			})
		} else {
			dl.Warn(&checker.LogMessage{
				Type: checker.FileTypeNone,
				Text: "Couldn't locate a valid workflow using the wrapper-validation-action Action. " + msg,
			})
		}
	}

	if removeGradleWrappers {
		filterFiles := []checker.File{}
		for _, f := range r.Files {
			if filepath.Base(f.Path) != "gradle-wrapper.jar" {
				filterFiles = append(filterFiles, f)
			}
		}
		r.Files = filterFiles
	}

	score := checker.MaxResultScore
	for _, f := range r.Files {
		dl.Warn(&checker.LogMessage{
			Path: f.Path, Type: checker.FileTypeBinary,
			Offset: f.Offset,
			Text:   "binary detected",
		})
		// We remove one point for each binary.
		score--
	}

	if score < checker.MinResultScore {
		score = checker.MinResultScore
	}

	return checker.CreateResultWithScore(name, "binaries present in source code", score)
}

// gradleWrapperValidated checks for the gradle-wrapper-verify Action being
// used in a non-failing workflow on the latest commit.
func gradleWrapperValidated(c clients.RepoClient) (bool, string, error) {
	metadata := gradleWrapperValidatingWorkflowMetadata{}
	err := fileparser.OnMatchingFileContentDo(c, fileparser.PathMatcher{
		Pattern:       ".github/workflows/*",
		CaseSensitive: false,
	}, checkWorkflowValidatesGradleWrapper, &metadata)
	if err != nil {
		return false, "", fmt.Errorf("%w", err)
	}
	if metadata.Path == "" {
		return false, "Could not find a workflow using the gradle/wrapper-validation-action Action.", nil
	}
	if !gradleWrapperValidationActionVersionConstraint.Check(metadata.ActionVersion) {
		// Version out of acceptable range.
		fmsg := fmt.Sprintf("The wrapper-validation-action version %s does not match requirement %s.",
			metadata.ActionVersion.String(),
			gradleWrapperValidationActionVersionConstraint.String())
		return false, fmsg, nil
	}
	// If validated, check that latest commit has a relevant successful run
	runs, err := c.ListSuccessfulWorkflowRuns(metadata.Path)
	if err != nil {
		return false, "", fmt.Errorf("failure listing workflow runs: %w", err)
	}
	commits, err := c.ListCommits()
	if err != nil {
		return false, "", fmt.Errorf("failure listing commits: %w", err)
	}
	if len(commits) < 1 || len(runs) < 1 {
		return false, "The repository has no workflow runs.", nil
	}
	for _, r := range runs {
		if *r.HeadSHA == commits[0].SHA {
			// Commit has corresponding successful run!
			return true, "Successfully validated verification.", nil
		}
	}
	return false, "Latest commit is not verified by passing workflow using the Gradle wrapper validation Action.", nil
}

// checkWorkflowValidatesGradleWrapper checks that the current workflow file
// is indeed using the gradle/wrapper-validation-action Action, else continues.
func checkWorkflowValidatesGradleWrapper(path string, content []byte, args ...interface{}) (bool, error) {
	validatingWorkflowMetadata, ok := args[0].(*gradleWrapperValidatingWorkflowMetadata)
	if !ok || validatingWorkflowMetadata == nil {
		return false, errBadArg
	}

	action, errs := actionlint.Parse(content)
	if len(errs) > 0 {
		return true, errs[0]
	}

	for _, j := range action.Jobs {
		for _, s := range j.Steps {
			ea, ok := s.Exec.(*actionlint.ExecAction)
			if !ok {
				continue
			}
			if ea.Uses == nil {
				continue
			}
			sms := gradleWrapperValidationActionRegex.FindStringSubmatch(ea.Uses.Value)
			if len(sms) > 1 {
				v, err := semver.NewVersion(sms[1])
				if err != nil {
					// Couldn't parse version, hopefully another step has
					// a correct one.
					continue
				}
				// OK! This is it.
				validatingWorkflowMetadata.Path = filepath.Base(path)
				validatingWorkflowMetadata.ActionVersion = v
				return true, nil
			}
		}
	}
	return true, nil
}
