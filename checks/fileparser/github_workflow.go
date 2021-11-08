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

package fileparser

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	sce "github.com/ossf/scorecard/v3/errors"
)

// defaultShellNonWindows is the default shell used for GitHub workflow actions for Linux and Mac.
const defaultShellNonWindows = "bash"

// defaultShellWindows is the default shell used for GitHub workflow actions for Windows.
const defaultShellWindows = "pwsh"

// Structure for workflow config.
// We only declare the fields we need.
// Github workflows format: https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions
type GitHubActionWorkflowConfig struct {
	Jobs map[string]GitHubActionWorkflowJob
	Name string `yaml:"name"`
}

// A Github Action Workflow Job.
// We only declare the fields we need.
// Github workflows format: https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions
// nolint: govet
type GitHubActionWorkflowJob struct {
	Name     string                     `yaml:"name"`
	Steps    []GitHubActionWorkflowStep `yaml:"steps"`
	Defaults struct {
		Run struct {
			Shell string `yaml:"shell"`
		} `yaml:"run"`
	} `yaml:"defaults"`
	RunsOn   stringOrSlice `yaml:"runs-on"`
	Strategy struct {
		// In most cases, the 'matrix' field will have a key of 'os' which is an array of strings, but there are
		// some repos that have something like: 'matrix: ${{ fromJson(needs.matrix.outputs.latest) }}'.
		Matrix interface{} `yaml:"matrix"`
	} `yaml:"strategy"`
}

// A Github Action Workflow Step.
// We only declare the fields we need.
// Github workflows format: https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions
type GitHubActionWorkflowStep struct {
	Name  string         `yaml:"name"`
	ID    string         `yaml:"id"`
	Shell string         `yaml:"shell"`
	Run   string         `yaml:"run"`
	If    string         `yaml:"if"`
	Uses  stringWithLine `yaml:"uses"`
}

// stringOrSlice is for fields that can be a single string or a slice of strings. If the field is a single string,
// this value will be a slice with a single string item.
type stringOrSlice []string

func (s *stringOrSlice) UnmarshalYAML(value *yaml.Node) error {
	var stringSlice []string
	err := value.Decode(&stringSlice)
	if err == nil {
		*s = stringSlice
		return nil
	}
	var single string
	err = value.Decode(&single)
	if err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("error decoding stringOrSlice Value: %v", err))
	}
	*s = []string{single}
	return nil
}

// stringWithLine is for when you want to keep track of the line number that the string came from.
type stringWithLine struct {
	Value string
	Line  int
}

func (ws *stringWithLine) UnmarshalYAML(value *yaml.Node) error {
	err := value.Decode(&ws.Value)
	if err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("error decoding stringWithLine Value: %v", err))
	}
	ws.Line = value.Line

	return nil
}

// GetOSesForJob returns the OSes this job runs on.
func GetOSesForJob(job *GitHubActionWorkflowJob) ([]string, error) {
	// The 'runs-on' field either lists the OS'es directly, or it can have an expression '${{ matrix.os }}' which
	// is where the OS'es are actually listed.
	getFromMatrix := len(job.RunsOn) == 1 && strings.Contains(job.RunsOn[0], "matrix.os")
	if !getFromMatrix {
		return job.RunsOn, nil
	}
	jobOSes := make([]string, 0)
	// nolint: nestif
	if m, ok := job.Strategy.Matrix.(map[string]interface{}); ok {
		if osVal, ok := m["os"]; ok {
			if oses, ok := osVal.([]interface{}); ok {
				for _, os := range oses {
					if strVal, ok := os.(string); ok {
						jobOSes = append(jobOSes, strVal)
					}
				}
				return jobOSes, nil
			}
		}
	}
	return jobOSes, sce.WithMessage(sce.ErrScorecardInternal,
		fmt.Sprintf("unable to determine OS for job: %v", job.Name))
}

// JobAlwaysRunsOnWindows returns true if the only OS that this job runs on is Windows.
func JobAlwaysRunsOnWindows(job *GitHubActionWorkflowJob) (bool, error) {
	jobOSes, err := GetOSesForJob(job)
	if err != nil {
		return false, err
	}
	for _, os := range jobOSes {
		if !strings.HasPrefix(strings.ToLower(os), "windows") {
			return false, nil
		}
	}
	return true, nil
}

// GetShellForStep returns the shell that is used to run the given step.
func GetShellForStep(step *GitHubActionWorkflowStep, job *GitHubActionWorkflowJob) (string, error) {
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#using-a-specific-shell.
	if step.Shell != "" {
		return step.Shell, nil
	}
	if job.Defaults.Run.Shell != "" {
		return job.Defaults.Run.Shell, nil
	}

	isStepWindows, err := IsStepWindows(step)
	if err != nil {
		return "", err
	}
	if isStepWindows {
		return defaultShellWindows, nil
	}

	alwaysRunsOnWindows, err := JobAlwaysRunsOnWindows(job)
	if err != nil {
		return "", err
	}
	if alwaysRunsOnWindows {
		return defaultShellWindows, nil
	}

	return defaultShellNonWindows, nil
}

// IsStepWindows returns true if the step will be run on Windows.
func IsStepWindows(step *GitHubActionWorkflowStep) (bool, error) {
	windowsRegexes := []string{
		// Looking for "if: runner.os == 'Windows'" (and variants)
		`(?i)runner\.os\s*==\s*['"]windows['"]`,
		// Looking for "if: ${{ startsWith(runner.os, 'Windows') }}" (and variants)
		`(?i)\$\{\{\s*startsWith\(runner\.os,\s*['"]windows['"]\)`,
		// Looking for "if: matrix.os == 'windows-2019'" (and variants)
		`(?i)matrix\.os\s*==\s*['"]windows-`,
	}

	for _, windowsRegex := range windowsRegexes {
		matches, err := regexp.MatchString(windowsRegex, step.If)
		if err != nil {
			return false, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("error matching Windows regex: %v", err))
		}
		if matches {
			return true, nil
		}
	}

	return false, nil
}

// IsWorkflowFile returns true if this is a GitHub workflow file.
func IsWorkflowFile(pathfn string) bool {
	// From https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions:
	// "Workflow files use YAML syntax, and must have either a .yml or .yaml file extension."
	switch path.Ext(pathfn) {
	case ".yml", ".yaml":
		return true
	default:
		return false
	}
}

// IsGitHubOwnedAction checks if this is a github specific action.
func IsGitHubOwnedAction(actionName string) bool {
	a := strings.HasPrefix(actionName, "actions/")
	c := strings.HasPrefix(actionName, "github/")
	return a || c
}
