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

package checks

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	scut "github.com/ossf/scorecard/v5/utests"
)

func TestSecretScanning_E2E(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		mockData  checker.SecretScanningData
		wantScore int
	}{
		{
			name: "GitHub native enabled",
			mockData: checker.SecretScanningData{
				Platform:        "github",
				GHNativeEnabled: checker.TriTrue,
			},
			wantScore: 10,
		},
		{
			name: "GitHub native disabled, gitleaks present",
			mockData: checker.SecretScanningData{
				Platform:                "github",
				GHNativeEnabled:         checker.TriFalse,
				ThirdPartyGitleaks:      true,
				ThirdPartyGitleaksPaths: []string{".github/workflows/security.yml"},
			},
			wantScore: 1, // Third-party tool present but no CI run data
		},
		{
			name: "GitHub native disabled, no third party",
			mockData: checker.SecretScanningData{
				Platform:        "github",
				GHNativeEnabled: checker.TriFalse,
			},
			wantScore: 0,
		},
		{
			name: "GitLab all features enabled",
			mockData: checker.SecretScanningData{
				Platform:                  "gitlab",
				GLSecretPushProtection:    true,
				GLPipelineSecretDetection: true,
				GLPushRulesPreventSecrets: true,
				ThirdPartyGitleaks:        true,
			},
			wantScore: 10, // 4+4+1+1 capped at 10
		},
		{
			name: "GitLab only SPP",
			mockData: checker.SecretScanningData{
				Platform:               "gitlab",
				GLSecretPushProtection: true,
			},
			wantScore: 4,
		},
		{
			name: "GitLab only pipeline",
			mockData: checker.SecretScanningData{
				Platform:                  "gitlab",
				GLPipelineSecretDetection: true,
			},
			wantScore: 4,
		},
		{
			name: "GitLab only push rules",
			mockData: checker.SecretScanningData{
				Platform:                  "gitlab",
				GLPushRulesPreventSecrets: true,
			},
			wantScore: 1,
		},
		{
			name: "GitLab only third party",
			mockData: checker.SecretScanningData{
				Platform:           "gitlab",
				ThirdPartyGitleaks: true,
			},
			wantScore: 1,
		},
		{
			name: "GitLab nothing enabled",
			mockData: checker.SecretScanningData{
				Platform: "gitlab",
			},
			wantScore: 0,
		},
		{
			name: "GitLab SPP and Pipeline",
			mockData: checker.SecretScanningData{
				Platform:                  "gitlab",
				GLSecretPushProtection:    true,
				GLPipelineSecretDetection: true,
			},
			wantScore: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mockedSecretScanningRepo{
				mockData: tt.mockData,
			}

			req := &checker.CheckRequest{
				Ctx:        context.Background(),
				RepoClient: mockRepo,
				Dlogger:    &scut.TestDetailLogger{},
			}

			result := SecretScanning(req)

			if result.Score != tt.wantScore {
				t.Errorf("Score = %d, want %d. Reason: %s", result.Score, tt.wantScore, result.Reason)
			}
		})
	}
}

// mockedSecretScanningRepo implements the minimal RepoClient interface for testing.
type mockedSecretScanningRepo struct {
	mockData checker.SecretScanningData
}

func (m *mockedSecretScanningRepo) InitRepo(repo clients.Repo, commitSHA string, commitDepth int) error {
	return nil
}

func (m *mockedSecretScanningRepo) URI() string {
	return "https://github.com/test/repo"
}

func (m *mockedSecretScanningRepo) IsArchived() (bool, error) {
	return false, nil
}

func (m *mockedSecretScanningRepo) ListFiles(predicate func(string) (bool, error)) ([]string, error) {
	return nil, nil
}

func (m *mockedSecretScanningRepo) GetFileReader(filename string) (io.ReadCloser, error) {
	return nil, nil
}

func (m *mockedSecretScanningRepo) ListCommits() ([]clients.Commit, error) {
	return nil, nil
}

func (m *mockedSecretScanningRepo) ListIssues() ([]clients.Issue, error) {
	return nil, nil
}

func (m *mockedSecretScanningRepo) ListReleases() ([]clients.Release, error) {
	return nil, nil
}

func (m *mockedSecretScanningRepo) ListContributors() ([]clients.User, error) {
	return nil, nil
}

func (m *mockedSecretScanningRepo) ListSuccessfulWorkflowRuns(filename string) ([]clients.WorkflowRun, error) {
	return nil, nil
}

func (m *mockedSecretScanningRepo) ListCheckRunsForRef(ref string) ([]clients.CheckRun, error) {
	return nil, nil
}

func (m *mockedSecretScanningRepo) ListStatuses(ref string) ([]clients.Status, error) {
	return nil, nil
}

func (m *mockedSecretScanningRepo) ListWebhooks() ([]clients.Webhook, error) {
	return nil, nil
}

func (m *mockedSecretScanningRepo) ListProgrammingLanguages() ([]clients.Language, error) {
	return nil, nil
}

func (m *mockedSecretScanningRepo) ListLicenses() ([]clients.License, error) {
	return nil, nil
}

func (m *mockedSecretScanningRepo) GetDefaultBranch() (*clients.BranchRef, error) {
	return nil, nil
}

func (m *mockedSecretScanningRepo) GetDefaultBranchName() (string, error) {
	return "main", nil
}

func (m *mockedSecretScanningRepo) GetBranch(branch string) (*clients.BranchRef, error) {
	return nil, nil
}

func (m *mockedSecretScanningRepo) GetCreatedAt() (time.Time, error) {
	return time.Now(), nil
}

func (m *mockedSecretScanningRepo) GetOrgRepoClient(ctx context.Context) (clients.RepoClient, error) {
	return nil, nil
}

func (m *mockedSecretScanningRepo) Search(request clients.SearchRequest) (clients.SearchResponse, error) {
	return clients.SearchResponse{}, nil
}

func (m *mockedSecretScanningRepo) SearchCommits(request clients.SearchCommitsOptions) ([]clients.Commit, error) {
	return nil, nil
}

func (m *mockedSecretScanningRepo) Close() error {
	return nil
}

func (m *mockedSecretScanningRepo) LocalPath() (string, error) {
	return "", nil
}

func (m *mockedSecretScanningRepo) ListMergeRequests() ([]clients.PullRequest, error) {
	return nil, nil
}

func (m *mockedSecretScanningRepo) GetFileContent(filename string) ([]byte, error) {
	return nil, nil
}

// This is the key method we're mocking.
func (m *mockedSecretScanningRepo) GetSecretScanningSignals() (
	clients.SecretScanningSignals,
	error,
) {
	return clients.SecretScanningSignals{
		Platform:                      clients.PlatformType(m.mockData.Platform),
		GHNativeEnabled:               toClientsTri(m.mockData.GHNativeEnabled),
		GHPushProtectionEnabled:       toClientsTri(m.mockData.GHPushProtectionEnabled),
		GLPipelineSecretDetection:     m.mockData.GLPipelineSecretDetection,
		GLSecretPushProtection:        m.mockData.GLSecretPushProtection,
		GLPushRulesPreventSecrets:     m.mockData.GLPushRulesPreventSecrets,
		ThirdPartyGitleaks:            m.mockData.ThirdPartyGitleaks,
		ThirdPartyGitleaksPaths:       m.mockData.ThirdPartyGitleaksPaths,
		ThirdPartyTruffleHog:          m.mockData.ThirdPartyTruffleHog,
		ThirdPartyTruffleHogPaths:     m.mockData.ThirdPartyTruffleHogPaths,
		ThirdPartyDetectSecrets:       m.mockData.ThirdPartyDetectSecrets,
		ThirdPartyDetectSecretsPaths:  m.mockData.ThirdPartyDetectSecretsPaths,
		ThirdPartyGitSecrets:          m.mockData.ThirdPartyGitSecrets,
		ThirdPartyGitSecretsPaths:     m.mockData.ThirdPartyGitSecretsPaths,
		ThirdPartyGGShield:            m.mockData.ThirdPartyGGShield,
		ThirdPartyGGShieldPaths:       m.mockData.ThirdPartyGGShieldPaths,
		ThirdPartyShhGit:              m.mockData.ThirdPartyShhGit,
		ThirdPartyShhGitPaths:         m.mockData.ThirdPartyShhGitPaths,
		ThirdPartyRepoSupervisor:      m.mockData.ThirdPartyRepoSupervisor,
		ThirdPartyRepoSupervisorPaths: m.mockData.ThirdPartyRepoSupervisorPaths,
		Evidence:                      m.mockData.Evidence,
	}, nil
}

func toClientsTri(t checker.TriState) clients.TriState {
	switch t {
	case checker.TriTrue:
		return clients.TriTrue
	case checker.TriFalse:
		return clients.TriFalse
	default:
		return clients.TriUnknown
	}
}
