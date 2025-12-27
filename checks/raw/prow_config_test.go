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
	"testing"

	"gopkg.in/yaml.v3"
)

func TestCommandContainsSASTTool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		command  []string
		expected bool
	}{
		{
			name:     "golangci-lint command",
			command:  []string{"golangci-lint", "run"},
			expected: true,
		},
		{
			name:     "codeql command",
			command:  []string{"codeql", "database", "analyze"},
			expected: true,
		},
		{
			name:     "trivy security scan",
			command:  []string{"trivy", "fs", "--security-checks", "vuln", "."},
			expected: true,
		},
		{
			name:     "regular build command",
			command:  []string{"make", "build"},
			expected: false,
		},
		{
			name:     "test command",
			command:  []string{"go", "test", "./..."},
			expected: false,
		},
		{
			name:     "shellcheck",
			command:  []string{"shellcheck", "scripts/*.sh"},
			expected: true,
		},
		{
			name:     "gosec",
			command:  []string{"gosec", "./..."},
			expected: true,
		},
		{
			name:     "empty command",
			command:  []string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := CommandContainsSASTTool(tt.command)
			if result != tt.expected {
				t.Errorf("CommandContainsSASTTool(%v) = %v, want %v", tt.command, result, tt.expected)
			}
		})
	}
}

func TestJobContainsSAST(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		job      ProwJob
		expected bool
	}{
		{
			name: "job with SAST in command",
			job: ProwJob{
				Name:    "lint",
				Command: []string{"golangci-lint", "run"},
				Args:    []string{},
			},
			expected: true,
		},
		{
			name: "job with SAST in args",
			job: ProwJob{
				Name:    "security-scan",
				Command: []string{"sh", "-c"},
				Args:    []string{"trivy fs ."},
			},
			expected: true,
		},
		{
			name: "job without SAST",
			job: ProwJob{
				Name:    "build",
				Command: []string{"make", "build"},
				Args:    []string{},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := jobContainsSAST(tt.job)
			if result != tt.expected {
				t.Errorf("jobContainsSAST(%v) = %v, want %v", tt.job, result, tt.expected)
			}
		})
	}
}

func TestProwConfigParsing(t *testing.T) {
	t.Parallel()

	yamlContent := `
presubmits:
  org/repo:
    - name: pull-lint
      command:
        - golangci-lint
        - run
    - name: pull-test
      command:
        - make
        - test

postsubmits:
  org/repo:
    - name: post-scan
      command:
        - trivy
        - fs
        - .

periodics:
  - name: nightly-check
    command:
      - codeql
      - analyze
`

	var config ProwConfig
	if err := yaml.Unmarshal([]byte(yamlContent), &config); err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	// Check presubmits
	if len(config.Presubmits["org/repo"]) != 2 {
		t.Errorf("Expected 2 presubmits, got %d", len(config.Presubmits["org/repo"]))
	}

	// Check postsubmits
	if len(config.Postsubmits["org/repo"]) != 1 {
		t.Errorf("Expected 1 postsubmit, got %d", len(config.Postsubmits["org/repo"]))
	}

	// Check periodics
	if len(config.Periodics) != 1 {
		t.Errorf("Expected 1 periodic, got %d", len(config.Periodics))
	}

	// Check if SAST tools are detected
	lintJob := config.Presubmits["org/repo"][0]
	if !jobContainsSAST(lintJob) {
		t.Error("Expected lint job to contain SAST tool")
	}

	testJob := config.Presubmits["org/repo"][1]
	if jobContainsSAST(testJob) {
		t.Error("Expected test job to NOT contain SAST tool")
	}
}
