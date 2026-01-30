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

package gitlabrepo

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func TestSecretScanningHandler_GitLabNative(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                    string
		ciConfig                string
		secretPushProtection    bool
		pushRulesPreventSecrets bool
		wantSPP                 bool
		wantPushRules           bool
		wantPipeline            bool
	}{
		{
			name:                    "all enabled",
			secretPushProtection:    true,
			pushRulesPreventSecrets: true,
			ciConfig: `
include:
  - template: Security/Secret-Detection.gitlab-ci.yml
`,
			wantSPP:       true,
			wantPushRules: true,
			wantPipeline:  true,
		},
		{
			name:                    "only SPP enabled",
			secretPushProtection:    true,
			pushRulesPreventSecrets: false,
			ciConfig:                "",
			wantSPP:                 true,
			wantPushRules:           false,
			wantPipeline:            false,
		},
		{
			name:                    "only push rules enabled",
			secretPushProtection:    false,
			pushRulesPreventSecrets: true,
			ciConfig:                "",
			wantSPP:                 false,
			wantPushRules:           true,
			wantPipeline:            false,
		},
		{
			name: "pipeline template only",
			ciConfig: `
include:
  - template: Security/Secret-Detection.gitlab-ci.yml
`,
			wantSPP:       false,
			wantPushRules: false,
			wantPipeline:  true,
		},
		{
			name: "alternative template name",
			ciConfig: `
include:
  - template: Jobs/Secret-Detection.gitlab-ci.yml
`,
			wantPipeline: true,
		},
		{
			name: "secret_detection job",
			ciConfig: `
secret_detection:
  script:
    - echo "scanning"
`,
			wantPipeline: true,
		},
		{
			name:          "all disabled",
			ciConfig:      "",
			wantSPP:       false,
			wantPushRules: false,
			wantPipeline:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/v4/projects/123/security_settings":
					settings := map[string]interface{}{
						"secret_push_protection_enabled": tt.secretPushProtection,
					}
					_ = json.NewEncoder(w).Encode(settings) //nolint:errcheck // Test mock
				case "/api/v4/projects/123/push_rule":
					if !tt.pushRulesPreventSecrets {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					pushRule := map[string]interface{}{
						"prevent_secrets": tt.pushRulesPreventSecrets,
					}
					_ = json.NewEncoder(w).Encode(pushRule) //nolint:errcheck // Test mock
				case "/api/v4/projects/123/repository/files/.gitlab-ci.yml":
					if tt.ciConfig == "" {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					encoded := base64.StdEncoding.EncodeToString([]byte(tt.ciConfig))
					file := map[string]interface{}{
						"content": encoded,
					}
					_ = json.NewEncoder(w).Encode(file) //nolint:errcheck // Test mock
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			glClient, err := gitlab.NewClient("token", gitlab.WithBaseURL(server.URL+"/api/v4"))
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			handler := &secretScanningHandler{
				glClient: glClient,
				ctx:      context.Background(),
				repourl: &Repo{
					projectID:     "123",
					defaultBranch: "main",
				},
			}

			result, err := handler.get()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.SecretPushProtection != tt.wantSPP {
				t.Errorf("SecretPushProtection = %v, want %v", result.SecretPushProtection, tt.wantSPP)
			}
			if result.PushRulesPreventSecrets != tt.wantPushRules {
				t.Errorf(
					"PushRulesPreventSecrets = %v, want %v",
					result.PushRulesPreventSecrets,
					tt.wantPushRules,
				)
			}
			if result.PipelineSecretDetection != tt.wantPipeline {
				t.Errorf(
					"PipelineSecretDetection = %v, want %v",
					result.PipelineSecretDetection,
					tt.wantPipeline,
				)
			}
		})
	}
}

//nolint:gocognit // Comprehensive test with multiple scenarios
func TestSecretScanningHandler_GitLabThirdParty(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name               string
		ciConfig           string
		wantGitleaks       bool
		wantTruffleHog     bool
		wantDetectSecrets  bool
		wantGitSecrets     bool
		wantGGShield       bool
		wantShhGit         bool
		wantRepoSupervisor bool
	}{
		{
			name: "gitleaks in script",
			ciConfig: `
scan:
  script:
    - gitleaks detect
`,
			wantGitleaks: true,
		},
		{
			name: "gitleaks in image",
			ciConfig: `
scan:
  image: zricethezav/gitleaks:latest
  script:
    - echo "scan"
`,
			wantGitleaks: true,
		},
		{
			name: "trufflehog in script",
			ciConfig: `
scan:
  script:
    - trufflehog filesystem .
`,
			wantTruffleHog: true,
		},
		{
			name: "trufflehog in image",
			ciConfig: `
scan:
  image: trufflesecurity/trufflehog:latest
  script:
    - echo "scan"
`,
			wantTruffleHog: true,
		},
		{
			name: "multiple tools in scripts",
			ciConfig: `
scan:
  script:
    - gitleaks detect
    - detect-secrets scan
    - git-secrets --scan
`,
			wantGitleaks:      true,
			wantDetectSecrets: true,
			wantGitSecrets:    true,
		},
		{
			name: "all tools present",
			ciConfig: `
scan1:
  script:
    - gitleaks detect
    - trufflehog filesystem
scan2:
  script:
    - detect-secrets scan
    - git-secrets --scan
    - ggshield scan
scan3:
  script:
    - shhgit
    - repo-supervisor scan
`,
			wantGitleaks:       true,
			wantTruffleHog:     true,
			wantDetectSecrets:  true,
			wantGitSecrets:     true,
			wantGGShield:       true,
			wantShhGit:         true,
			wantRepoSupervisor: true,
		},
		{
			name: "case insensitive",
			ciConfig: `
scan:
  script:
    - GITLEAKS detect
    - TRUFFLEHOG scan
`,
			wantGitleaks:   true,
			wantTruffleHog: true,
		},
		{
			name: "script as string instead of array",
			ciConfig: `
scan:
  script: gitleaks detect
`,
			wantGitleaks: true,
		},
		{
			name: "ggshield image",
			ciConfig: `
scan:
  image: gitguardian/ggshield:latest
  script:
    - echo test
`,
			wantGGShield: true,
		},
		{
			name:         "no tools",
			ciConfig:     `scan:\n  script:\n    - echo "test"`,
			wantGitleaks: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/v4/projects/123/security_settings":
					w.WriteHeader(http.StatusNotFound)
				case "/api/v4/projects/123/push_rule":
					w.WriteHeader(http.StatusNotFound)
				case "/api/v4/projects/123/repository/files/.gitlab-ci.yml":
					encoded := base64.StdEncoding.EncodeToString([]byte(tt.ciConfig))
					file := map[string]interface{}{
						"content": encoded,
					}
					_ = json.NewEncoder(w).Encode(file) //nolint:errcheck // Test mock
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			glClient, err := gitlab.NewClient("token", gitlab.WithBaseURL(server.URL+"/api/v4"))
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			handler := &secretScanningHandler{
				glClient: glClient,
				ctx:      context.Background(),
				repourl: &Repo{
					projectID:     "123",
					defaultBranch: "main",
				},
			}

			result, err := handler.get()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.TpGitleaks != tt.wantGitleaks {
				t.Errorf("TpGitleaks = %v, want %v", result.TpGitleaks, tt.wantGitleaks)
			}
			if result.TpTruffleHog != tt.wantTruffleHog {
				t.Errorf("TpTruffleHog = %v, want %v", result.TpTruffleHog, tt.wantTruffleHog)
			}
			if result.TpDetectSecrets != tt.wantDetectSecrets {
				t.Errorf("TpDetectSecrets = %v, want %v", result.TpDetectSecrets, tt.wantDetectSecrets)
			}
			if result.TpGitSecrets != tt.wantGitSecrets {
				t.Errorf("TpGitSecrets = %v, want %v", result.TpGitSecrets, tt.wantGitSecrets)
			}
			if result.TpGGShield != tt.wantGGShield {
				t.Errorf("TpGGShield = %v, want %v", result.TpGGShield, tt.wantGGShield)
			}
			if result.TpShhGit != tt.wantShhGit {
				t.Errorf("TpShhGit = %v, want %v", result.TpShhGit, tt.wantShhGit)
			}
			if result.TpRepoSupervisor != tt.wantRepoSupervisor {
				t.Errorf("TpRepoSupervisor = %v, want %v", result.TpRepoSupervisor, tt.wantRepoSupervisor)
			}
		})
	}
}

func TestSecretScanningHandler_InvalidYAML(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v4/projects/123/repository/files/.gitlab-ci.yml":
			invalidYAML := "this is: not: valid: yaml: ["
			encoded := base64.StdEncoding.EncodeToString([]byte(invalidYAML))
			file := map[string]interface{}{
				"content": encoded,
			}
			_ = json.NewEncoder(w).Encode(file) //nolint:errcheck // Test mock
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	glClient, err := gitlab.NewClient("token", gitlab.WithBaseURL(server.URL+"/api/v4"))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	handler := &secretScanningHandler{
		glClient: glClient,
		ctx:      context.Background(),
		repourl: &Repo{
			projectID:     "123",
			defaultBranch: "main",
		},
	}

	// Should not error, just skip invalid YAML
	result, err := handler.get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.PipelineSecretDetection {
		t.Error("Should not detect pipeline secret detection in invalid YAML")
	}
}

func TestSecretScanningHandler_Paths(t *testing.T) {
	t.Parallel()
	ciConfig := `
scan:
  script:
    - gitleaks detect
    - trufflehog scan
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v4/projects/123/repository/files/.gitlab-ci.yml":
			encoded := base64.StdEncoding.EncodeToString([]byte(ciConfig))
			file := map[string]interface{}{
				"content": encoded,
			}
			_ = json.NewEncoder(w).Encode(file) //nolint:errcheck // Test mock
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	glClient, err := gitlab.NewClient("token", gitlab.WithBaseURL(server.URL+"/api/v4"))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	handler := &secretScanningHandler{
		glClient: glClient,
		ctx:      context.Background(),
		repourl: &Repo{
			projectID:     "123",
			defaultBranch: "main",
		},
	}

	result, err := handler.get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedPath := []string{".gitlab-ci.yml"}
	if !cmp.Equal(result.TpGitleaksPaths, expectedPath) {
		t.Errorf("TpGitleaksPaths = %v, want %v", result.TpGitleaksPaths, expectedPath)
	}
	if !cmp.Equal(result.TpTruffleHogPaths, expectedPath) {
		t.Errorf("TpTruffleHogPaths = %v, want %v", result.TpTruffleHogPaths, expectedPath)
	}
}

func TestSecretScanningHandler_Evidence(t *testing.T) {
	t.Parallel()
	ciConfig := `
include:
  - template: Security/Secret-Detection.gitlab-ci.yml
scan:
  script:
    - gitleaks detect
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v4/projects/123/security_settings":
			settings := map[string]interface{}{
				"secret_push_protection_enabled": true,
			}
			_ = json.NewEncoder(w).Encode(settings) //nolint:errcheck // Test mock
		case "/api/v4/projects/123/push_rule":
			pushRule := map[string]interface{}{
				"prevent_secrets": true,
			}
			_ = json.NewEncoder(w).Encode(pushRule) //nolint:errcheck // Test mock
		case "/api/v4/projects/123/repository/files/.gitlab-ci.yml":
			encoded := base64.StdEncoding.EncodeToString([]byte(ciConfig))
			file := map[string]interface{}{
				"content": encoded,
			}
			_ = json.NewEncoder(w).Encode(file) //nolint:errcheck // Test mock
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	glClient, err := gitlab.NewClient("token", gitlab.WithBaseURL(server.URL+"/api/v4"))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	handler := &secretScanningHandler{
		glClient: glClient,
		ctx:      context.Background(),
		repourl: &Repo{
			projectID:     "123",
			defaultBranch: "main",
		},
	}

	result, err := handler.get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Evidence) == 0 {
		t.Error("Expected evidence to be collected")
	}

	// Should have evidence about various settings
	foundSPP := false
	foundPushRules := false
	foundTemplate := false
	foundGitleaks := false

	for _, e := range result.Evidence {
		if e == "project.security_settings.secret_push_protection_enabled=true" {
			foundSPP = true
		}
		if e == "push_rule.prevent_secrets=true" {
			foundPushRules = true
		}
		// The handler preserves the original casing from the YAML
		if strings.HasPrefix(strings.ToLower(e), "include.template=") &&
			strings.Contains(strings.ToLower(e), "secret-detection.gitlab-ci.yml") {
			foundTemplate = true
		}
		if e == "script:gitleaks@.gitlab-ci.yml" {
			foundGitleaks = true
		}
	}

	if !foundSPP {
		t.Error("Expected SPP evidence")
	}
	if !foundPushRules {
		t.Error("Expected push rules evidence")
	}
	if !foundTemplate {
		t.Error("Expected template evidence")
	}
	if !foundGitleaks {
		t.Error("Expected gitleaks evidence")
	}
}

func TestIsSecretDetectionTemplate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  bool
	}{
		{"Security/Secret-Detection.gitlab-ci.yml", true},
		{"security/secret-detection.gitlab-ci.yml", true},
		{"Jobs/Secret-Detection.gitlab-ci.yml", true},
		{"jobs/secret-detection.gitlab-ci.yml", true},
		{"  Security/Secret-Detection.gitlab-ci.yml  ", true},
		{"Other/Template.gitlab-ci.yml", false},
		{"", false},
		{"security/sast.gitlab-ci.yml", false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("input=%q", tt.input), func(t *testing.T) {
			t.Parallel()
			got := isSecretDetectionTemplate(tt.input)
			if got != tt.want {
				t.Errorf("isSecretDetectionTemplate(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
