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

package evaluation

import (
	"strings"
	"testing"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/hasGitHubPushProtectionEnabled"
	"github.com/ossf/scorecard/v5/probes/hasGitHubSecretScanningEnabled"
	"github.com/ossf/scorecard/v5/probes/hasGitLabPipelineSecretDetection"
	"github.com/ossf/scorecard/v5/probes/hasGitLabPushRulesPreventSecrets"
	"github.com/ossf/scorecard/v5/probes/hasGitLabSecretPushProtection"
	"github.com/ossf/scorecard/v5/probes/hasThirdPartyDetectSecrets"
	"github.com/ossf/scorecard/v5/probes/hasThirdPartyGGShield"
	"github.com/ossf/scorecard/v5/probes/hasThirdPartyGitSecrets"
	"github.com/ossf/scorecard/v5/probes/hasThirdPartyGitleaks"
	"github.com/ossf/scorecard/v5/probes/hasThirdPartyRepoSupervisor"
	"github.com/ossf/scorecard/v5/probes/hasThirdPartyShhGit"
	"github.com/ossf/scorecard/v5/probes/hasThirdPartyTruffleHog"
	scut "github.com/ossf/scorecard/v5/utests"
)

func TestSecretScanning_GitHub(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		findings []finding.Finding
		raw      *checker.SecretScanningData
		result   scut.TestReturn
	}{
		{
			name: "GitHub native enabled, no third party",
			findings: []finding.Finding{
				{
					Probe:   hasGitHubSecretScanningEnabled.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   hasGitHubPushProtectionEnabled.Probe,
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score: 10,
			},
		},
		{
			name: "GitHub native enabled with push protection",
			findings: []finding.Finding{
				{
					Probe:   hasGitHubSecretScanningEnabled.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   hasGitHubPushProtectionEnabled.Probe,
					Outcome: finding.OutcomeTrue,
				},
			},
			result: scut.TestReturn{
				Score: 10,
			},
		},
		{
			name: "GitHub native enabled with third party tools",
			findings: []finding.Finding{
				{
					Probe:   hasGitHubSecretScanningEnabled.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   hasGitHubPushProtectionEnabled.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   hasThirdPartyGitleaks.Probe,
					Outcome: finding.OutcomeTrue,
					Message: "Gitleaks found at .github/workflows/security.yml",
				},
			},
			result: scut.TestReturn{
				Score: 10,
			},
		},
		{
			name: "GitHub native disabled, no third party",
			findings: []finding.Finding{
				{
					Probe:   hasGitHubSecretScanningEnabled.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasGitHubPushProtectionEnabled.Probe,
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score: 0,
			},
		},
		{
			name: "GitHub native disabled, third party present",
			findings: []finding.Finding{
				{
					Probe:   hasGitHubSecretScanningEnabled.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasGitHubPushProtectionEnabled.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasThirdPartyGitleaks.Probe,
					Outcome: finding.OutcomeTrue,
					Message: "Gitleaks found at .github/workflows/security.yml",
				},
			},
			result: scut.TestReturn{
				Score: 1, // Third-party tool present but no CI run data
			},
		},
		{
			name: "GitHub native disabled, multiple third party tools",
			findings: []finding.Finding{
				{
					Probe:   hasGitHubSecretScanningEnabled.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasGitHubPushProtectionEnabled.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasThirdPartyGitleaks.Probe,
					Outcome: finding.OutcomeTrue,
					Message: "Gitleaks found at .github/workflows/scan1.yml",
				},
				{
					Probe:   hasThirdPartyTruffleHog.Probe,
					Outcome: finding.OutcomeTrue,
					Message: "TruffleHog found at .github/workflows/scan2.yml",
				},
			},
			result: scut.TestReturn{
				Score: 1, // Multiple third-party tools present but no CI run data
			},
		},
		{
			name:     "GitHub permission denied, no third party",
			findings: []finding.Finding{
				// No GitHub native probes because we couldn't check
			},
			raw: &checker.SecretScanningData{
				Platform: "github",
				Evidence: []string{"source:permission_denied"},
			},
			result: scut.TestReturn{
				Score: checker.InconclusiveResultScore,
			},
		},
		{
			name: "GitHub permission denied, with third party",
			findings: []finding.Finding{
				{
					Probe:   hasThirdPartyGitleaks.Probe,
					Outcome: finding.OutcomeTrue,
					Message: "Gitleaks found at .github/workflows/security.yml",
				},
			},
			raw: &checker.SecretScanningData{
				Platform:           "github",
				Evidence:           []string{"source:permission_denied"},
				ThirdPartyGitleaks: true,
			},
			result: scut.TestReturn{
				Score: checker.InconclusiveResultScore,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := &scut.TestDetailLogger{}
			got := SecretScanning("SecretScanning", tt.findings, dl, tt.raw)
			scut.ValidateTestReturn(t, tt.name, &tt.result, &got, dl)
			validatePermissionDenied(t, tt.raw, &got)
		})
	}
}

// validatePermissionDenied checks permission denied cases have correct messaging.
func validatePermissionDenied(t *testing.T, raw *checker.SecretScanningData, got *checker.CheckResult) {
	t.Helper()
	if raw == nil || raw.Platform != "github" {
		return
	}
	for _, evidence := range raw.Evidence {
		if !strings.Contains(evidence, "permission_denied") {
			continue
		}
		// Verify the message is about insufficient permissions, not "disabled"
		if !strings.Contains(got.Reason, "insufficient permissions") {
			t.Errorf(
				"Expected reason to mention 'insufficient permissions', got: %s",
				got.Reason,
			)
		}
		if strings.Contains(got.Reason, "disabled") {
			t.Errorf(
				"Reason should not say 'disabled' for permission errors, got: %s",
				got.Reason,
			)
		}
		if got.Score != checker.InconclusiveResultScore {
			t.Errorf(
				"Expected score %d for permission denied, got: %d",
				checker.InconclusiveResultScore,
				got.Score,
			)
		}
	}
}

func TestSecretScanning_GitLab(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		findings []finding.Finding
		result   scut.TestReturn
	}{
		{
			name: "GitLab all enabled",
			findings: []finding.Finding{
				{
					Probe:   hasGitLabSecretPushProtection.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   hasGitLabPipelineSecretDetection.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   hasGitLabPushRulesPreventSecrets.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   hasThirdPartyGitleaks.Probe,
					Outcome: finding.OutcomeTrue,
					Message: "Gitleaks found at .gitlab-ci.yml",
				},
			},
			result: scut.TestReturn{
				Score: 10, // 4 + 4 + 1 + 1 = 10 (capped)
			},
		},
		{
			name: "GitLab only SPP",
			findings: []finding.Finding{
				{
					Probe:   hasGitLabSecretPushProtection.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   hasGitLabPipelineSecretDetection.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasGitLabPushRulesPreventSecrets.Probe,
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score: 4,
			},
		},
		{
			name: "GitLab only Pipeline",
			findings: []finding.Finding{
				{
					Probe:   hasGitLabSecretPushProtection.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasGitLabPipelineSecretDetection.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   hasGitLabPushRulesPreventSecrets.Probe,
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score: 4,
			},
		},
		{
			name: "GitLab SPP and Pipeline",
			findings: []finding.Finding{
				{
					Probe:   hasGitLabSecretPushProtection.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   hasGitLabPipelineSecretDetection.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   hasGitLabPushRulesPreventSecrets.Probe,
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score: 8,
			},
		},
		{
			name: "GitLab only push rules",
			findings: []finding.Finding{
				{
					Probe:   hasGitLabSecretPushProtection.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasGitLabPipelineSecretDetection.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasGitLabPushRulesPreventSecrets.Probe,
					Outcome: finding.OutcomeTrue,
				},
			},
			result: scut.TestReturn{
				Score: 1,
			},
		},
		{
			name: "GitLab only third party",
			findings: []finding.Finding{
				{
					Probe:   hasGitLabSecretPushProtection.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasGitLabPipelineSecretDetection.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasGitLabPushRulesPreventSecrets.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasThirdPartyGitleaks.Probe,
					Outcome: finding.OutcomeTrue,
				},
			},
			result: scut.TestReturn{
				Score: 1,
			},
		},
		{
			name: "GitLab nothing enabled",
			findings: []finding.Finding{
				{
					Probe:   hasGitLabSecretPushProtection.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasGitLabPipelineSecretDetection.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasGitLabPushRulesPreventSecrets.Probe,
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score: 0,
			},
		},
		{
			name: "GitLab multiple third party tools",
			findings: []finding.Finding{
				{
					Probe:   hasGitLabSecretPushProtection.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasGitLabPipelineSecretDetection.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasGitLabPushRulesPreventSecrets.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasThirdPartyGitleaks.Probe,
					Outcome: finding.OutcomeTrue,
					Message: "Gitleaks found at .gitlab-ci.yml",
				},
				{
					Probe:   hasThirdPartyTruffleHog.Probe,
					Outcome: finding.OutcomeTrue,
					Message: "TruffleHog found at .gitlab-ci.yml",
				},
			},
			result: scut.TestReturn{
				Score: 1, // Still just 1 point for third party
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := &scut.TestDetailLogger{}
			got := SecretScanning("SecretScanning", tt.findings, dl, nil)
			scut.ValidateTestReturn(t, tt.name, &tt.result, &got, dl)
		})
	}
}

func TestSecretScanning_AllThirdPartyTools(t *testing.T) {
	t.Parallel()
	// Test that all seven third-party tools are detected
	findings := []finding.Finding{
		{
			Probe:   hasGitLabSecretPushProtection.Probe,
			Outcome: finding.OutcomeFalse,
		},
		{
			Probe:   hasGitLabPipelineSecretDetection.Probe,
			Outcome: finding.OutcomeFalse,
		},
		{
			Probe:   hasGitLabPushRulesPreventSecrets.Probe,
			Outcome: finding.OutcomeFalse,
		},
		{
			Probe:   hasThirdPartyGitleaks.Probe,
			Outcome: finding.OutcomeTrue,
		},
		{
			Probe:   hasThirdPartyTruffleHog.Probe,
			Outcome: finding.OutcomeTrue,
		},
		{
			Probe:   hasThirdPartyDetectSecrets.Probe,
			Outcome: finding.OutcomeTrue,
		},
		{
			Probe:   hasThirdPartyGitSecrets.Probe,
			Outcome: finding.OutcomeTrue,
		},
		{
			Probe:   hasThirdPartyGGShield.Probe,
			Outcome: finding.OutcomeTrue,
		},
		{
			Probe:   hasThirdPartyShhGit.Probe,
			Outcome: finding.OutcomeTrue,
		},
		{
			Probe:   hasThirdPartyRepoSupervisor.Probe,
			Outcome: finding.OutcomeTrue,
		},
	}

	dl := &scut.TestDetailLogger{}
	got := SecretScanning("SecretScanning", findings, dl, nil)

	// All seven tools present = +1 point (on GitLab path)
	if got.Score != 1 {
		t.Errorf(
			"Expected score 1 with all third-party tools, got %d",
			got.Score,
		)
	}
}

func TestSecretScanning_ScoreCapping(t *testing.T) {
	t.Parallel()
	// Test that GitLab score caps at 10
	findings := []finding.Finding{
		{
			Probe:   hasGitLabSecretPushProtection.Probe,
			Outcome: finding.OutcomeTrue, // 4 points
		},
		{
			Probe:   hasGitLabPipelineSecretDetection.Probe,
			Outcome: finding.OutcomeTrue, // 4 points
		},
		{
			Probe:   hasGitLabPushRulesPreventSecrets.Probe,
			Outcome: finding.OutcomeTrue, // 1 point
		},
		{
			Probe:   hasThirdPartyGitleaks.Probe,
			Outcome: finding.OutcomeTrue, // 1 point = 10 total
		},
	}

	dl := &scut.TestDetailLogger{}
	got := SecretScanning("SecretScanning", findings, dl, nil)

	if got.Score != checker.MaxResultScore {
		t.Errorf(
			"Expected score %d (capped), got %d",
			checker.MaxResultScore,
			got.Score,
		)
	}
}

func TestSecretScanning_ReasonMessages(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name               string
		findings           []finding.Finding
		wantReasonContains []string
	}{
		{
			name: "GitHub native enabled",
			findings: []finding.Finding{
				{
					Probe:   hasGitHubSecretScanningEnabled.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   hasGitHubPushProtectionEnabled.Probe,
					Outcome: finding.OutcomeFalse,
				},
			},
			wantReasonContains: []string{"GitHub native secret scanning is enabled"},
		},
		{
			name: "GitHub with push protection",
			findings: []finding.Finding{
				{
					Probe:   hasGitHubSecretScanningEnabled.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   hasGitHubPushProtectionEnabled.Probe,
					Outcome: finding.OutcomeTrue,
				},
			},
			wantReasonContains: []string{"push protection enabled"},
		},
		{
			name: "GitHub disabled with third party",
			findings: []finding.Finding{
				{
					Probe:   hasGitHubSecretScanningEnabled.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasGitHubPushProtectionEnabled.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasThirdPartyGitleaks.Probe,
					Outcome: finding.OutcomeTrue,
					Message: "Gitleaks found at .github/workflows/security.yml",
				},
			},
			wantReasonContains: []string{"disabled", "Gitleaks found"},
		},
		{
			name: "GitLab all features",
			findings: []finding.Finding{
				{
					Probe:   hasGitLabSecretPushProtection.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   hasGitLabPipelineSecretDetection.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   hasGitLabPushRulesPreventSecrets.Probe,
					Outcome: finding.OutcomeTrue,
				},
			},
			wantReasonContains: []string{
				"Secret Push Protection: on",
				"Pipeline Secret Detection: on",
				"Push rules prevent_secrets: on",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := &scut.TestDetailLogger{}
			got := SecretScanning("SecretScanning", tt.findings, dl, nil)

			for _, want := range tt.wantReasonContains {
				if !stringContains(got.Reason, want) {
					t.Errorf("Expected reason to contain %q, got: %s", want, got.Reason)
				}
			}
		})
	}
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestSecretScanning_ThirdPartyWithCIStats tests scoring with
// CI statistics and execution patterns.
func TestSecretScanning_ThirdPartyWithCIStats(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		findings []finding.Finding
		raw      *checker.SecretScanningData
		result   scut.TestReturn
	}{
		{
			name: "Commit-based tool with 100% coverage",
			findings: []finding.Finding{
				{
					Probe:   hasGitHubSecretScanningEnabled.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasGitHubPushProtectionEnabled.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasThirdPartyGitleaks.Probe,
					Outcome: finding.OutcomeTrue,
				},
			},
			raw: &checker.SecretScanningData{
				Platform:           "github",
				ThirdPartyGitleaks: true,
				ThirdPartyCIInfo: map[string]*checker.ToolCIStats{
					"gitleaks": {
						ToolName:             "gitleaks",
						ExecutionPattern:     "commit-based",
						TotalCommitsAnalyzed: 100,
						CommitsWithToolRun:   100,
						HasRecentRuns:        true,
					},
				},
			},
			result: scut.TestReturn{
				Score: 10, // 100% coverage = 10 points
			},
		},
		{
			name: "Commit-based tool with 70% coverage",
			findings: []finding.Finding{
				{
					Probe:   hasGitHubSecretScanningEnabled.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasGitHubPushProtectionEnabled.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasThirdPartyGitleaks.Probe,
					Outcome: finding.OutcomeTrue,
				},
			},
			raw: &checker.SecretScanningData{
				Platform:           "github",
				ThirdPartyGitleaks: true,
				ThirdPartyCIInfo: map[string]*checker.ToolCIStats{
					"gitleaks": {
						ToolName:             "gitleaks",
						ExecutionPattern:     "commit-based",
						TotalCommitsAnalyzed: 100,
						CommitsWithToolRun:   70,
						HasRecentRuns:        true,
					},
				},
			},
			result: scut.TestReturn{
				Score: 7, // 70-99% coverage = 7 points
			},
		},
		{
			name: "Commit-based tool with 50% coverage",
			findings: []finding.Finding{
				{
					Probe:   hasGitHubSecretScanningEnabled.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasGitHubPushProtectionEnabled.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasThirdPartyTruffleHog.Probe,
					Outcome: finding.OutcomeTrue,
				},
			},
			raw: &checker.SecretScanningData{
				Platform:             "github",
				ThirdPartyTruffleHog: true,
				ThirdPartyCIInfo: map[string]*checker.ToolCIStats{
					"trufflehog": {
						ToolName:             "trufflehog",
						ExecutionPattern:     "commit-based",
						TotalCommitsAnalyzed: 100,
						CommitsWithToolRun:   50,
						HasRecentRuns:        true,
					},
				},
			},
			result: scut.TestReturn{
				Score: 5, // 50-69% coverage = 5 points
			},
		},
		{
			name: "Commit-based tool with 30% coverage",
			findings: []finding.Finding{
				{
					Probe:   hasGitHubSecretScanningEnabled.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasGitHubPushProtectionEnabled.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasThirdPartyDetectSecrets.Probe,
					Outcome: finding.OutcomeTrue,
				},
			},
			raw: &checker.SecretScanningData{
				Platform:                "github",
				ThirdPartyDetectSecrets: true,
				ThirdPartyCIInfo: map[string]*checker.ToolCIStats{
					"detect-secrets": {
						ToolName:             "detect-secrets",
						ExecutionPattern:     "commit-based",
						TotalCommitsAnalyzed: 100,
						CommitsWithToolRun:   30,
						HasRecentRuns:        true,
					},
				},
			},
			result: scut.TestReturn{
				Score: 3, // Less than 50% = 3 points
			},
		},
		{
			name: "Periodic tool with recent runs",
			findings: []finding.Finding{
				{
					Probe:   hasGitHubSecretScanningEnabled.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasGitHubPushProtectionEnabled.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasThirdPartyShhGit.Probe,
					Outcome: finding.OutcomeTrue,
				},
			},
			raw: &checker.SecretScanningData{
				Platform:         "github",
				ThirdPartyShhGit: true,
				ThirdPartyCIInfo: map[string]*checker.ToolCIStats{
					"shhgit": {
						ToolName:             "shhgit",
						ExecutionPattern:     "periodic",
						TotalCommitsAnalyzed: 100,
						HasRecentRuns:        true,
					},
				},
			},
			result: scut.TestReturn{
				Score: 10, // Periodic with recent runs = 10 points
			},
		},
		{
			name: "Periodic tool without recent runs",
			findings: []finding.Finding{
				{
					Probe:   hasGitHubSecretScanningEnabled.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasGitHubPushProtectionEnabled.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasThirdPartyRepoSupervisor.Probe,
					Outcome: finding.OutcomeTrue,
				},
			},
			raw: &checker.SecretScanningData{
				Platform:                 "github",
				ThirdPartyRepoSupervisor: true,
				ThirdPartyCIInfo: map[string]*checker.ToolCIStats{
					"repo-supervisor": {
						ToolName:             "repo-supervisor",
						ExecutionPattern:     "periodic",
						TotalCommitsAnalyzed: 100,
						HasRecentRuns:        false,
					},
				},
			},
			result: scut.TestReturn{
				Score: 1, // Periodic without recent runs = 1 point
			},
		},
		{
			name: "Multiple tools with mixed execution patterns",
			findings: []finding.Finding{
				{
					Probe:   hasGitHubSecretScanningEnabled.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasGitHubPushProtectionEnabled.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasThirdPartyGitleaks.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   hasThirdPartyShhGit.Probe,
					Outcome: finding.OutcomeTrue,
				},
			},
			raw: &checker.SecretScanningData{
				Platform:           "github",
				ThirdPartyGitleaks: true,
				ThirdPartyShhGit:   true,
				ThirdPartyCIInfo: map[string]*checker.ToolCIStats{
					"gitleaks": {
						ToolName:             "gitleaks",
						ExecutionPattern:     "commit-based",
						TotalCommitsAnalyzed: 100,
						CommitsWithToolRun:   80,
						HasRecentRuns:        true,
					},
					"shhgit": {
						ToolName:             "shhgit",
						ExecutionPattern:     "periodic",
						TotalCommitsAnalyzed: 100,
						HasRecentRuns:        true,
					},
				},
			},
			result: scut.TestReturn{
				Score: 10, // Best tool wins: max(7 from gitleaks, 10 from shhgit) = 10
			},
		},
		{
			name: "Tool present but no CI stats",
			findings: []finding.Finding{
				{
					Probe:   hasGitHubSecretScanningEnabled.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasGitHubPushProtectionEnabled.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   hasThirdPartyGitleaks.Probe,
					Outcome: finding.OutcomeTrue,
				},
			},
			raw: &checker.SecretScanningData{
				Platform:           "github",
				ThirdPartyGitleaks: true,
				ThirdPartyCIInfo:   map[string]*checker.ToolCIStats{}, // Empty CI stats
			},
			result: scut.TestReturn{
				Score: 1, // Tool present but no CI data = 1 point
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := &scut.TestDetailLogger{}
			got := SecretScanning("SecretScanning", tt.findings, dl, tt.raw)
			scut.ValidateTestReturn(t, tt.name, &tt.result, &got, dl)
		})
	}
}

// TestScoreForTool tests the scoreForTool function
// with different execution patterns.
func TestScoreForTool(t *testing.T) {
	t.Parallel()
	tests := []struct { //nolint:govet // Test struct alignment not critical
		name          string
		toolName      string
		stats         *checker.ToolCIStats
		expectedScore int
	}{
		{
			name:          "Nil stats returns 1",
			toolName:      "gitleaks",
			stats:         nil,
			expectedScore: 1,
		},
		{
			name:     "Zero commits analyzed returns 1",
			toolName: "gitleaks",
			stats: &checker.ToolCIStats{
				ToolName:             "gitleaks",
				ExecutionPattern:     "commit-based",
				TotalCommitsAnalyzed: 0,
			},
			expectedScore: 1,
		},
		{
			name:     "Periodic with recent runs returns 10",
			toolName: "shhgit",
			stats: &checker.ToolCIStats{
				ToolName:             "shhgit",
				ExecutionPattern:     "periodic",
				TotalCommitsAnalyzed: 100,
				HasRecentRuns:        true,
			},
			expectedScore: 10,
		},
		{
			name:     "Periodic without recent runs returns 1",
			toolName: "repo-supervisor",
			stats: &checker.ToolCIStats{
				ToolName:             "repo-supervisor",
				ExecutionPattern:     "periodic",
				TotalCommitsAnalyzed: 100,
				HasRecentRuns:        false,
			},
			expectedScore: 1,
		},
		{
			name:     "Commit-based with 100% coverage returns 10",
			toolName: "gitleaks",
			stats: &checker.ToolCIStats{
				ToolName:             "gitleaks",
				ExecutionPattern:     "commit-based",
				TotalCommitsAnalyzed: 100,
				CommitsWithToolRun:   100,
			},
			expectedScore: 10,
		},
		{
			name:     "Commit-based with 90% coverage returns 7",
			toolName: "trufflehog",
			stats: &checker.ToolCIStats{
				ToolName:             "trufflehog",
				ExecutionPattern:     "commit-based",
				TotalCommitsAnalyzed: 100,
				CommitsWithToolRun:   90,
			},
			expectedScore: 7,
		},
		{
			name:     "Commit-based with 70% coverage returns 7",
			toolName: "detect-secrets",
			stats: &checker.ToolCIStats{
				ToolName:             "detect-secrets",
				ExecutionPattern:     "commit-based",
				TotalCommitsAnalyzed: 100,
				CommitsWithToolRun:   70,
			},
			expectedScore: 7,
		},
		{
			name:     "Commit-based with 65% coverage returns 5",
			toolName: "git-secrets",
			stats: &checker.ToolCIStats{
				ToolName:             "git-secrets",
				ExecutionPattern:     "commit-based",
				TotalCommitsAnalyzed: 100,
				CommitsWithToolRun:   65,
			},
			expectedScore: 5,
		},
		{
			name:     "Commit-based with 50% coverage returns 5",
			toolName: "ggshield",
			stats: &checker.ToolCIStats{
				ToolName:             "ggshield",
				ExecutionPattern:     "commit-based",
				TotalCommitsAnalyzed: 100,
				CommitsWithToolRun:   50,
			},
			expectedScore: 5,
		},
		{
			name:     "Commit-based with 25% coverage returns 3",
			toolName: "gitleaks",
			stats: &checker.ToolCIStats{
				ToolName:             "gitleaks",
				ExecutionPattern:     "commit-based",
				TotalCommitsAnalyzed: 100,
				CommitsWithToolRun:   25,
			},
			expectedScore: 3,
		},
		{
			name:     "Commit-based with 0% coverage returns 1",
			toolName: "trufflehog",
			stats: &checker.ToolCIStats{
				ToolName:             "trufflehog",
				ExecutionPattern:     "commit-based",
				TotalCommitsAnalyzed: 100,
				CommitsWithToolRun:   0,
			},
			expectedScore: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			score := scoreForTool(tt.toolName, tt.stats)
			if score != tt.expectedScore {
				t.Errorf(
					"Expected score %d, got %d",
					tt.expectedScore,
					score,
				)
			}
		})
	}
}

// TestScoreFromCoverage tests the coverage-to-score mapping.
func TestScoreFromCoverage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		coverage      float64
		expectedScore int
	}{
		{
			name:          "100% coverage",
			coverage:      1.0,
			expectedScore: 10,
		},
		{
			name:          "99% coverage",
			coverage:      0.99,
			expectedScore: 7,
		},
		{
			name:          "70% coverage",
			coverage:      0.70,
			expectedScore: 7,
		},
		{
			name:          "69% coverage",
			coverage:      0.69,
			expectedScore: 5,
		},
		{
			name:          "50% coverage",
			coverage:      0.50,
			expectedScore: 5,
		},
		{
			name:          "49% coverage",
			coverage:      0.49,
			expectedScore: 3,
		},
		{
			name:          "25% coverage",
			coverage:      0.25,
			expectedScore: 3,
		},
		{
			name:          "1% coverage",
			coverage:      0.01,
			expectedScore: 3,
		},
		{
			name:          "0% coverage",
			coverage:      0.0,
			expectedScore: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			score := scoreFromCoverage(tt.coverage)
			if score != tt.expectedScore {
				t.Errorf("For coverage %.2f%%, expected score %d, got %d",
					tt.coverage*100, tt.expectedScore, score)
			}
		})
	}
}

// TestFormatCICoverageDetails tests the formatting
// of CI coverage details.
func TestFormatCICoverageDetails(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		ciInfo         map[string]*checker.ToolCIStats
		expectedSubstr []string
	}{
		{
			name:           "Empty CI info returns empty string",
			ciInfo:         map[string]*checker.ToolCIStats{},
			expectedSubstr: []string{},
		},
		{
			name: "Periodic tool with recent runs",
			ciInfo: map[string]*checker.ToolCIStats{
				"shhgit": {
					ToolName:             "shhgit",
					ExecutionPattern:     "periodic",
					TotalCommitsAnalyzed: 100,
					HasRecentRuns:        true,
				},
			},
			expectedSubstr: []string{"shhgit: ran recently"},
		},
		{
			name: "Periodic tool without recent runs",
			ciInfo: map[string]*checker.ToolCIStats{
				"repo-supervisor": {
					ToolName:             "repo-supervisor",
					ExecutionPattern:     "periodic",
					TotalCommitsAnalyzed: 100,
					HasRecentRuns:        false,
				},
			},
			expectedSubstr: []string{"repo-supervisor: no recent runs"},
		},
		{
			name: "Commit-based tool with coverage",
			ciInfo: map[string]*checker.ToolCIStats{
				"gitleaks": {
					ToolName:             "gitleaks",
					ExecutionPattern:     "commit-based",
					TotalCommitsAnalyzed: 100,
					CommitsWithToolRun:   75,
				},
			},
			expectedSubstr: []string{"gitleaks: 75% coverage"},
		},
		{
			name: "Multiple tools mixed",
			ciInfo: map[string]*checker.ToolCIStats{
				"gitleaks": {
					ToolName:             "gitleaks",
					ExecutionPattern:     "commit-based",
					TotalCommitsAnalyzed: 100,
					CommitsWithToolRun:   80,
				},
				"shhgit": {
					ToolName:             "shhgit",
					ExecutionPattern:     "periodic",
					TotalCommitsAnalyzed: 100,
					HasRecentRuns:        true,
				},
			},
			expectedSubstr: []string{"coverage", "ran recently"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := formatCICoverageDetails(tt.ciInfo)

			if len(tt.expectedSubstr) == 0 && result != "" {
				t.Errorf("Expected empty string, got %q", result)
				return
			}

			for _, substr := range tt.expectedSubstr {
				if !stringContains(result, substr) {
					t.Errorf("Expected result to contain %q, got: %s", substr, result)
				}
			}
		})
	}
}
