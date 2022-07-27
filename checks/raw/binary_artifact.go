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

package raw

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	semver "github.com/Masterminds/semver/v3"
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
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

// mustParseConstraint attempts parse of semver constraint, panics if fail.
func mustParseConstraint(c string) *semver.Constraints {
	if c, err := semver.NewConstraint(c); err != nil {
		panic(fmt.Errorf("failed to parse constraint: %w", err))
	} else {
		return c
	}
}

// BinaryArtifacts retrieves the raw data for the Binary-Artifacts check.
func BinaryArtifacts(c clients.RepoClient) (checker.BinaryArtifactData, error) {
	files := []checker.File{}
	err := fileparser.OnMatchingFileContentDo(c, fileparser.PathMatcher{
		Pattern:       "*",
		CaseSensitive: false,
	}, checkBinaryFileContent, &files)
	if err != nil {
		return checker.BinaryArtifactData{}, fmt.Errorf("%w", err)
	}
	// Ignore validated gradle-wrapper.jar files if present
	files, err = excludeValidatedGradleWrappers(c, files)
	if err != nil {
		return checker.BinaryArtifactData{}, fmt.Errorf("%w", err)
	}

	// No error, return the files.
	return checker.BinaryArtifactData{Files: files}, nil
}

// excludeValidatedGradleWrappers returns the subset of files not confirmed
// to be Action-validated gradle-wrapper.jar files.
func excludeValidatedGradleWrappers(c clients.RepoClient, files []checker.File) ([]checker.File, error) {
	// Check if gradle-wrapper.jar present
	if !fileExists(files, "gradle-wrapper.jar") {
		return files, nil
	}
	// Gradle wrapper JARs present, so check that they are validated
	ok, err := gradleWrapperValidated(c)
	if err != nil {
		return files, fmt.Errorf(
			"failure checking for Gradle wrapper validating Action: %w", err)
	}
	if !ok {
		// Gradle Wrappers not validated
		return files, nil
	}
	// It has been confirmed that latest commit has validated JARs!
	// Remove Gradle wrapper JARs from files.
	filterFiles := []checker.File{}
	for _, f := range files {
		if filepath.Base(f.Path) != "gradle-wrapper.jar" {
			filterFiles = append(filterFiles, f)
		}
	}
	files = filterFiles
	return files, nil
}

var checkBinaryFileContent fileparser.DoWhileTrueOnFileContent = func(path string, content []byte,
	args ...interface{},
) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf(
			"checkBinaryFileContent requires exactly one argument: %w", errInvalidArgLength)
	}
	pfiles, ok := args[0].(*[]checker.File)
	if !ok {
		return false, fmt.Errorf(
			"checkBinaryFileContent requires argument of type *[]checker.File: %w", errInvalidArgType)
	}

	binaryFileTypes := map[string]bool{
		"crx":    true,
		"deb":    true,
		"dex":    true,
		"dey":    true,
		"elf":    true,
		"o":      true,
		"so":     true,
		"macho":  true,
		"iso":    true,
		"class":  true,
		"jar":    true,
		"bundle": true,
		"dylib":  true,
		"lib":    true,
		"msi":    true,
		"dll":    true,
		"drv":    true,
		"efi":    true,
		"exe":    true,
		"ocx":    true,
		"pyc":    true,
		"pyo":    true,
		"par":    true,
		"rpm":    true,
		"whl":    true,
	}
	var t types.Type
	var err error
	if len(content) == 0 {
		return true, nil
	}
	if t, err = filetype.Get(content); err != nil {
		return false, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("filetype.Get:%v", err))
	}

	exists1 := binaryFileTypes[t.Extension]
	if exists1 {
		*pfiles = append(*pfiles, checker.File{
			Path:   path,
			Type:   checker.FileTypeBinary,
			Offset: checker.OffsetDefault,
		})
		return true, nil
	}

	exists2 := binaryFileTypes[strings.ReplaceAll(filepath.Ext(path), ".", "")]
	if !isText(content) && exists2 {
		*pfiles = append(*pfiles, checker.File{
			Path:   path,
			Type:   checker.FileTypeBinary,
			Offset: checker.OffsetDefault,
		})
	}

	return true, nil
}

// TODO: refine this function.
func isText(content []byte) bool {
	for _, c := range string(content) {
		if c == '\t' || c == '\n' || c == '\r' {
			continue
		}
		if !unicode.IsPrint(c) {
			return false
		}
	}
	return true
}

// gradleWrapperValidated checks for the gradle-wrapper-verify action being
// used in a non-failing workflow on the latest commit.
func gradleWrapperValidated(c clients.RepoClient) (bool, error) {
	gradleWrapperValidatingWorkflowFile := ""
	err := fileparser.OnMatchingFileContentDo(c, fileparser.PathMatcher{
		Pattern:       ".github/workflows/*",
		CaseSensitive: false,
	}, checkWorkflowValidatesGradleWrapper, &gradleWrapperValidatingWorkflowFile)
	if err != nil {
		return false, fmt.Errorf("%w", err)
	}
	if gradleWrapperValidatingWorkflowFile != "" {
		// If validated, check that latest commit has a relevant successful run
		runs, err := c.ListSuccessfulWorkflowRuns(gradleWrapperValidatingWorkflowFile)
		if err != nil {
			return false, fmt.Errorf("failure listing workflow runs: %w", err)
		}
		commits, err := c.ListCommits()
		if err != nil {
			return false, fmt.Errorf("failure listing commits: %w", err)
		}
		if len(commits) < 1 || len(runs) < 1 {
			return false, nil
		}
		for _, r := range runs {
			if *r.HeadSHA == commits[0].SHA {
				// Commit has corresponding successful run!
				return true, nil
			}
		}
	}
	return false, nil
}

// checkWorkflowValidatesGradleWrapper checks that the current workflow file
// is indeed using the gradle/wrapper-validation-action action, else continues.
func checkWorkflowValidatesGradleWrapper(path string, content []byte, args ...interface{}) (bool, error) {
	validatingWorkflowFile, ok := args[0].(*string)
	if !ok {
		return false, fmt.Errorf("checkWorkflowValidatesGradleWrapper expects arg[0] of type *string: %w", errInvalidArgType)
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
				if !gradleWrapperValidationActionVersionConstraint.Check(v) {
					// Version out of acceptable range.
					continue
				}
				// OK! This is it.
				*validatingWorkflowFile = filepath.Base(path)
				return true, nil
			}
		}
	}
	return true, nil
}

// fileExists checks if a file of name name exists, including within
// subdirectories.
func fileExists(files []checker.File, name string) bool {
	for _, f := range files {
		if filepath.Base(f.Path) == name {
			return true
		}
	}
	return false
}
