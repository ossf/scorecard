// Copyright 2025 OpenSSF Scorecard Authors
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
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks/fileparser"
	"github.com/ossf/scorecard/v5/finding"
)

var errProwInvalidArgs = errors.New("invalid arguments")

// ProwConfig represents a Prow configuration file.
type ProwConfig struct {
	Presubmits  map[string][]ProwJob `yaml:"presubmits"`
	Postsubmits map[string][]ProwJob `yaml:"postsubmits"`
	Periodics   []ProwJob            `yaml:"periodics"`
}

// ProwJob represents a single Prow job definition.
type ProwJob struct {
	Name    string   `yaml:"name"`
	Command []string `yaml:"command"`
	Args    []string `yaml:"args"`
}

// CommandContainsSASTTool checks if a command/args contains SAST tool indicators.
// Uses the same pattern list as checkRun/status detection for consistency.
func CommandContainsSASTTool(command []string) bool {
	commandStr := strings.ToLower(strings.Join(command, " "))
	// Reuse sastToolPatterns from sast.go for consistency
	for _, pattern := range sastToolPatterns {
		if strings.Contains(commandStr, pattern) {
			return true
		}
	}
	// Also check for generic "lint" which is common in commands
	if strings.Contains(commandStr, "lint") {
		return true
	}
	return false
}

// logDebugProwf logs debug messages for Prow detection.
func logDebugProwf(c *checker.CheckRequest, format string, args ...interface{}) {
	if c.Dlogger != nil {
		c.Dlogger.Debug(&checker.LogMessage{
			Text: fmt.Sprintf("[Prow] "+format, args...),
		})
	}
}

// getProwSASTJobs scans local Prow config files for SAST tools.
// This mirrors the GitHub workflow scanning approach.
func getProwSASTJobs(c *checker.CheckRequest) ([]checker.SASTWorkflow, error) {
	var configPaths []string
	var sastWorkflows []checker.SASTWorkflow

	logDebugProwf(c, "Scanning for Prow configuration files...")

	// Scan common Prow config file patterns
	patterns := []string{".prow.yaml", ".prow/*.yaml", "prow/*.yaml"}

	for _, pattern := range patterns {
		logDebugProwf(c, "Scanning pattern: %s", pattern)
		err := fileparser.OnMatchingFileContentDo(c.RepoClient, fileparser.PathMatcher{
			Pattern:       pattern,
			CaseSensitive: false,
		}, searchProwConfigForSAST, &configPaths)
		if err != nil {
			return sastWorkflows, err
		}
	}

	if len(configPaths) > 0 {
		logDebugProwf(c, "Found %d Prow config file(s) with SAST tools:", len(configPaths))
		for _, path := range configPaths {
			logDebugProwf(c, "  ✓ %s", path)
		}
	} else {
		logDebugProwf(c, "No Prow config files with SAST tools found")
	}

	// Convert paths to SASTWorkflow objects
	for _, path := range configPaths {
		sastWorkflow := checker.SASTWorkflow{
			File: checker.File{
				Path:   path,
				Offset: checker.OffsetDefault,
				Type:   finding.FileTypeSource,
			},
			Type: "Prow",
		}
		sastWorkflows = append(sastWorkflows, sastWorkflow)
	}

	return sastWorkflows, nil
}

// searchProwConfigForSAST searches a Prow config file for SAST tools.
var searchProwConfigForSAST fileparser.DoWhileTrueOnFileContent = func(path string,
	content []byte,
	args ...interface{},
) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("searchProwConfigForSAST requires exactly 1 argument: %w", errProwInvalidArgs)
	}

	paths, ok := args[0].(*[]string)
	if !ok {
		return false, fmt.Errorf("searchProwConfigForSAST expects arg[0] of type *[]string: %w", errProwInvalidArgs)
	}

	var config ProwConfig
	if err := yaml.Unmarshal(content, &config); err != nil {
		// Skip files that aren't valid Prow configs
		return true, nil
	}

	// Check all job types for SAST tools
	hasSAST := false

	// Check presubmits
	for _, jobs := range config.Presubmits {
		for _, job := range jobs {
			if jobContainsSAST(job) {
				hasSAST = true
				break
			}
		}
	}

	// Check postsubmits
	if !hasSAST {
		for _, jobs := range config.Postsubmits {
			for _, job := range jobs {
				if jobContainsSAST(job) {
					hasSAST = true
					break
				}
			}
		}
	}

	// Check periodics
	if !hasSAST {
		for _, job := range config.Periodics {
			if jobContainsSAST(job) {
				hasSAST = true
				break
			}
		}
	}

	if hasSAST {
		*paths = append(*paths, path)
	}

	return true, nil
}

// jobContainsSAST checks if a Prow job contains SAST tools.
func jobContainsSAST(job ProwJob) bool {
	return CommandContainsSASTTool(job.Command) || CommandContainsSASTTool(job.Args)
}
