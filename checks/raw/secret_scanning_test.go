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

	"github.com/ossf/scorecard/v5/checker"
)

func TestInitializeToolStats(t *testing.T) {
	t.Parallel()
	tests := []struct { //nolint:govet // Test struct alignment not critical
		name     string
		data     *checker.SecretScanningData
		expected map[string]string // map of tool name to expected execution pattern
	}{
		{
			name: "Commit-based tools get commit-based pattern",
			data: &checker.SecretScanningData{
				ThirdPartyGitleaks:      true,
				ThirdPartyTruffleHog:    true,
				ThirdPartyDetectSecrets: true,
				ThirdPartyGitSecrets:    true,
				ThirdPartyGGShield:      true,
			},
			expected: map[string]string{
				"gitleaks":       "commit-based",
				"trufflehog":     "commit-based",
				"detect-secrets": "commit-based",
				"git-secrets":    "commit-based",
				"ggshield":       "commit-based",
			},
		},
		{
			name: "Periodic tools get periodic pattern",
			data: &checker.SecretScanningData{
				ThirdPartyShhGit:         true,
				ThirdPartyRepoSupervisor: true,
			},
			expected: map[string]string{
				"shhgit":          "periodic",
				"repo-supervisor": "periodic",
			},
		},
		{
			name: "Mixed tools get correct patterns",
			data: &checker.SecretScanningData{
				ThirdPartyGitleaks:       true,
				ThirdPartyShhGit:         true,
				ThirdPartyRepoSupervisor: true,
				ThirdPartyTruffleHog:     true,
			},
			expected: map[string]string{
				"gitleaks":        "commit-based",
				"shhgit":          "periodic",
				"repo-supervisor": "periodic",
				"trufflehog":      "commit-based",
			},
		},
		{
			name:     "No tools detected returns empty map",
			data:     &checker.SecretScanningData{},
			expected: map[string]string{},
		},
		{
			name: "Single commit-based tool",
			data: &checker.SecretScanningData{
				ThirdPartyGitleaks: true,
			},
			expected: map[string]string{
				"gitleaks": "commit-based",
			},
		},
		{
			name: "Single periodic tool",
			data: &checker.SecretScanningData{
				ThirdPartyShhGit: true,
			},
			expected: map[string]string{
				"shhgit": "periodic",
			},
		},
		{
			name: "All seven tools",
			data: &checker.SecretScanningData{
				ThirdPartyGitleaks:       true,
				ThirdPartyTruffleHog:     true,
				ThirdPartyDetectSecrets:  true,
				ThirdPartyGitSecrets:     true,
				ThirdPartyGGShield:       true,
				ThirdPartyShhGit:         true,
				ThirdPartyRepoSupervisor: true,
			},
			expected: map[string]string{
				"gitleaks":        "commit-based",
				"trufflehog":      "commit-based",
				"detect-secrets":  "commit-based",
				"git-secrets":     "commit-based",
				"ggshield":        "commit-based",
				"shhgit":          "periodic",
				"repo-supervisor": "periodic",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			stats := initializeToolStats(tt.data)

			// Check that we got the expected number of tools
			if len(stats) != len(tt.expected) {
				t.Errorf("Expected %d tools, got %d", len(tt.expected), len(stats))
			}

			// Check each tool has the correct execution pattern
			for toolName, expectedPattern := range tt.expected {
				toolStats, found := stats[toolName]
				if !found {
					t.Errorf("Tool %q not found in stats", toolName)
					continue
				}

				if toolStats.ExecutionPattern != expectedPattern {
					t.Errorf("Tool %q: expected execution pattern %q, got %q",
						toolName, expectedPattern, toolStats.ExecutionPattern)
				}

				if toolStats.ToolName != toolName {
					t.Errorf("Tool %q: expected ToolName to be %q, got %q",
						toolName, toolName, toolStats.ToolName)
				}
			}
		})
	}
}

func TestGetDetectedTools(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		data          *checker.SecretScanningData
		expectedTools []string
	}{
		{
			name:          "No tools detected",
			data:          &checker.SecretScanningData{},
			expectedTools: []string{},
		},
		{
			name: "Only gitleaks",
			data: &checker.SecretScanningData{
				ThirdPartyGitleaks: true,
			},
			expectedTools: []string{"gitleaks"},
		},
		{
			name: "All commit-based tools",
			data: &checker.SecretScanningData{
				ThirdPartyGitleaks:      true,
				ThirdPartyTruffleHog:    true,
				ThirdPartyDetectSecrets: true,
				ThirdPartyGitSecrets:    true,
				ThirdPartyGGShield:      true,
			},
			expectedTools: []string{
				"gitleaks", "trufflehog", "detect-secrets",
				"git-secrets", "ggshield",
			},
		},
		{
			name: "All periodic tools",
			data: &checker.SecretScanningData{
				ThirdPartyShhGit:         true,
				ThirdPartyRepoSupervisor: true,
			},
			expectedTools: []string{"shhgit", "repo-supervisor"},
		},
		{
			name: "All seven tools",
			data: &checker.SecretScanningData{
				ThirdPartyGitleaks:       true,
				ThirdPartyTruffleHog:     true,
				ThirdPartyDetectSecrets:  true,
				ThirdPartyGitSecrets:     true,
				ThirdPartyGGShield:       true,
				ThirdPartyShhGit:         true,
				ThirdPartyRepoSupervisor: true,
			},
			expectedTools: []string{
				"gitleaks", "trufflehog", "detect-secrets",
				"git-secrets", "ggshield", "shhgit", "repo-supervisor",
			},
		},
		{
			name: "Mixed tools",
			data: &checker.SecretScanningData{
				ThirdPartyGitleaks:   true,
				ThirdPartyTruffleHog: true,
				ThirdPartyShhGit:     true,
			},
			expectedTools: []string{"gitleaks", "trufflehog", "shhgit"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			detectedTools := getDetectedTools(tt.data)

			if len(detectedTools) != len(tt.expectedTools) {
				t.Errorf(
					"Expected %d tools, got %d",
					len(tt.expectedTools),
					len(detectedTools),
				)
			}

			for _, expectedTool := range tt.expectedTools {
				if !detectedTools[expectedTool] {
					t.Errorf("Expected tool %q to be detected", expectedTool)
				}
			}
		})
	}
}

func TestPeriodicToolsMap(t *testing.T) {
	t.Parallel()
	// This test verifies that the periodicTools map contains exactly the expected tools
	// It's a documentation test to ensure the map doesn't drift from expectations

	expectedPeriodic := map[string]bool{
		"shhgit":          true,
		"repo-supervisor": true,
	}

	expectedCommitBased := []string{
		"gitleaks",
		"trufflehog",
		"detect-secrets",
		"git-secrets",
		"ggshield",
	}

	// Test that periodic tools are correctly identified
	data := &checker.SecretScanningData{
		ThirdPartyShhGit:         true,
		ThirdPartyRepoSupervisor: true,
	}
	stats := initializeToolStats(data)

	for toolName := range expectedPeriodic {
		if stats[toolName].ExecutionPattern != "periodic" {
			t.Errorf(
				"Tool %q should be periodic, got %q",
				toolName,
				stats[toolName].ExecutionPattern,
			)
		}
	}

	// Test that commit-based tools are correctly identified
	data2 := &checker.SecretScanningData{
		ThirdPartyGitleaks:      true,
		ThirdPartyTruffleHog:    true,
		ThirdPartyDetectSecrets: true,
		ThirdPartyGitSecrets:    true,
		ThirdPartyGGShield:      true,
	}
	stats2 := initializeToolStats(data2)

	for _, toolName := range expectedCommitBased {
		if stats2[toolName].ExecutionPattern != "commit-based" {
			t.Errorf(
				"Tool %q should be commit-based, got %q",
				toolName,
				stats2[toolName].ExecutionPattern,
			)
		}
	}
}
