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

package githubrepo

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v53/github"

	"github.com/ossf/scorecard/v5/clients"
)

type secretScanningRoundTripper struct {
	securityAnalysis *github.SecurityAndAnalysis
	workflowFiles    map[string]string
	alertsStatus     int
}

func (s secretScanningRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Handle repository get request
	if strings.Contains(req.URL.Path, "/repos/") &&
		req.Method == http.MethodGet &&
		!strings.Contains(req.URL.Path, "contents") &&
		!strings.Contains(req.URL.Path, "secret-scanning") {
		repo := &github.Repository{
			SecurityAndAnalysis: s.securityAnalysis,
		}
		jsonResp, err := json.Marshal(repo)
		if err != nil {
			return nil, err
		}
		return &http.Response{
			Status:     "200 OK",
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(jsonResp)),
			Header:     make(http.Header),
		}, nil
	}

	// Handle secret-scanning alerts endpoint
	if strings.Contains(req.URL.Path, "secret-scanning/alerts") {
		resp := &http.Response{
			Status:     fmt.Sprintf("%d", s.alertsStatus),
			StatusCode: s.alertsStatus,
			Body:       io.NopCloser(bytes.NewReader([]byte("[]"))),
			Header:     make(http.Header),
		}
		// Return the response - the handler checks the status code to determine
		// if secret scanning is enabled (200 = enabled)
		return resp, nil
	}

	// Handle individual workflow file contents (must come before directory listing)
	if strings.Contains(req.URL.Path, "contents/.github/workflows/") &&
		req.Method == http.MethodGet {
		for name, content := range s.workflowFiles {
			if !strings.HasSuffix(req.URL.Path, name) {
				continue
			}
			fileType := "file"
			path := ".github/workflows/" + name
			// Encode content as base64 as GitHub API does
			encodedContent := base64.StdEncoding.EncodeToString([]byte(content))
			rc := &github.RepositoryContent{
				Type:     &fileType,
				Path:     &path,
				Content:  &encodedContent,
				Encoding: github.String("base64"),
			}
			jsonResp, err := json.Marshal(rc)
			if err != nil {
				return nil, err
			}
			return &http.Response{
				Status:     "200 OK",
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(jsonResp)),
				Header:     make(http.Header),
			}, nil
		}
	}

	// Handle .github/workflows directory listing
	if strings.HasSuffix(req.URL.Path, "contents/.github/workflows") &&
		req.Method == http.MethodGet {
		var entries []*github.RepositoryContent
		for name := range s.workflowFiles {
			fileType := "file"
			path := ".github/workflows/" + name
			entries = append(entries, &github.RepositoryContent{
				Type: &fileType,
				Path: &path,
				Name: &name,
			})
		}
		jsonResp, err := json.Marshal(entries)
		if err != nil {
			return nil, err
		}
		return &http.Response{
			Status:     "200 OK",
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(jsonResp)),
			Header:     make(http.Header),
		}, nil
	}

	return &http.Response{
		Status:     "404 Not Found",
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(bytes.NewReader([]byte("{}"))),
		Header:     make(http.Header),
	}, nil
}

func TestSecretScanningHandler_NativeEnabled(t *testing.T) {
	t.Parallel()
	tests := []struct {
		securityAnalysis *github.SecurityAndAnalysis
		name             string
		alertsStatus     int
		wantNative       clients.TriState
		wantPushProtect  clients.TriState
	}{
		{
			name: "both enabled",
			securityAnalysis: &github.SecurityAndAnalysis{
				SecretScanning: &github.SecretScanning{
					Status: github.String("enabled"),
				},
				SecretScanningPushProtection: &github.SecretScanningPushProtection{
					Status: github.String("enabled"),
				},
			},
			wantNative:      clients.TriTrue,
			wantPushProtect: clients.TriTrue,
		},
		{
			name: "scanning enabled, push protection disabled",
			securityAnalysis: &github.SecurityAndAnalysis{
				SecretScanning: &github.SecretScanning{
					Status: github.String("enabled"),
				},
				SecretScanningPushProtection: &github.SecretScanningPushProtection{
					Status: github.String("disabled"),
				},
			},
			wantNative:      clients.TriTrue,
			wantPushProtect: clients.TriFalse,
		},
		{
			name: "both disabled",
			securityAnalysis: &github.SecurityAndAnalysis{
				SecretScanning: &github.SecretScanning{
					Status: github.String("disabled"),
				},
				SecretScanningPushProtection: &github.SecretScanningPushProtection{
					Status: github.String("disabled"),
				},
			},
			wantNative:      clients.TriFalse,
			wantPushProtect: clients.TriFalse,
		},
		{
			name:             "nil security analysis, alerts 200",
			securityAnalysis: nil,
			alertsStatus:     http.StatusOK,
			wantNative:       clients.TriTrue,
			wantPushProtect:  clients.TriUnknown,
		},
		{
			name:             "nil security analysis, alerts 404",
			securityAnalysis: nil,
			alertsStatus:     http.StatusNotFound,
			wantNative:       clients.TriUnknown,
			wantPushProtect:  clients.TriUnknown,
		},
		{
			name: "empty status strings",
			securityAnalysis: &github.SecurityAndAnalysis{
				SecretScanning:               &github.SecretScanning{},
				SecretScanningPushProtection: &github.SecretScanningPushProtection{},
			},
			alertsStatus:    http.StatusNotFound,
			wantNative:      clients.TriUnknown,
			wantPushProtect: clients.TriUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			httpClient := &http.Client{
				Transport: secretScanningRoundTripper{
					securityAnalysis: tt.securityAnalysis,
					alertsStatus:     tt.alertsStatus,
					workflowFiles:    map[string]string{},
				},
			}
			ghClient := github.NewClient(httpClient)
			ctx := context.Background()

			handler := &secretScanningHandler{
				ctx: ctx,
				gh:  ghClient,
				repourl: &Repo{
					owner: "test-owner",
					repo:  "test-repo",
				},
			}

			result, err := handler.get()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.NativeEnabled != tt.wantNative {
				t.Errorf(
					"NativeEnabled = %v, want %v",
					result.NativeEnabled,
					tt.wantNative,
				)
			}
			if result.PushProtectionEnabled != tt.wantPushProtect {
				t.Errorf(
					"PushProtectionEnabled = %v, want %v",
					result.PushProtectionEnabled,
					tt.wantPushProtect,
				)
			}
		})
	}
}

//nolint:gocognit // Comprehensive test with multiple scenarios
func TestSecretScanningHandler_ThirdPartyTools(t *testing.T) {
	t.Parallel()
	tests := []struct {
		workflowFiles      map[string]string
		wantPaths          map[string][]string
		name               string
		wantGitleaks       bool
		wantTruffleHog     bool
		wantDetectSecrets  bool
		wantGitSecrets     bool
		wantGGShield       bool
		wantShhGit         bool
		wantRepoSupervisor bool
	}{
		{
			name: "gitleaks in workflow",
			workflowFiles: map[string]string{
				"security.yml": `
name: Security
on: [push]
jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: gitleaks/gitleaks-action@v2
`,
			},
			wantGitleaks: true,
			wantPaths: map[string][]string{
				"gitleaks": {".github/workflows/security.yml"},
			},
		},
		{
			name: "trufflehog in workflow",
			workflowFiles: map[string]string{
				"scan.yml": `
name: Scan
jobs:
  trufflehog:
    runs-on: ubuntu-latest
    steps:
      - uses: trufflesecurity/trufflehog@v3
`,
			},
			wantTruffleHog: true,
			wantPaths: map[string][]string{
				"trufflehog": {".github/workflows/scan.yml"},
			},
		},
		{
			name: "multiple tools",
			workflowFiles: map[string]string{
				"security.yml": `
name: Security
jobs:
  secrets:
    runs-on: ubuntu-latest
    steps:
      - uses: gitleaks/gitleaks-action@v2
      - run: detect-secrets scan
`,
			},
			wantGitleaks:      true,
			wantDetectSecrets: true,
			wantPaths: map[string][]string{
				"gitleaks":       {".github/workflows/security.yml"},
				"detect-secrets": {".github/workflows/security.yml"},
			},
		},
		{
			name: "case insensitive matching",
			workflowFiles: map[string]string{
				"test.yml": `
jobs:
  scan:
    steps:
      - uses: GITLEAKS/gitleaks-action@v2
      - run: TRUFFLEHOG scan
`,
			},
			wantGitleaks:   true,
			wantTruffleHog: true,
		},
		{
			name: "all tools present",
			workflowFiles: map[string]string{
				"scan.yml": `
jobs:
  scan:
    steps:
      - uses: gitleaks/gitleaks-action@v2
      - uses: trufflesecurity/trufflehog@v3
      - run: detect-secrets scan
      - run: git-secrets --scan
      - uses: gitguardian/ggshield@v1
      - run: shhgit
      - run: repo-supervisor scan
`,
			},
			wantGitleaks:       true,
			wantTruffleHog:     true,
			wantDetectSecrets:  true,
			wantGitSecrets:     true,
			wantGGShield:       true,
			wantShhGit:         true,
			wantRepoSupervisor: true,
		},
		{
			name: "tool in comment should still match",
			workflowFiles: map[string]string{
				"test.yml": `
jobs:
  scan:
    steps:
      # Using gitleaks for scanning
      - run: echo "test"
`,
			},
			wantGitleaks: true,
		},
		{
			name:          "no workflows",
			workflowFiles: map[string]string{},
			wantGitleaks:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			httpClient := &http.Client{
				Transport: secretScanningRoundTripper{
					securityAnalysis: &github.SecurityAndAnalysis{
						SecretScanning: &github.SecretScanning{
							Status: github.String("disabled"),
						},
					},
					alertsStatus:  http.StatusNotFound,
					workflowFiles: tt.workflowFiles,
				},
			}
			ghClient := github.NewClient(httpClient)
			ctx := context.Background()

			handler := &secretScanningHandler{
				ctx: ctx,
				gh:  ghClient,
				repourl: &Repo{
					owner: "test-owner",
					repo:  "test-repo",
				},
			}

			result, err := handler.get()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.TpGitleaks != tt.wantGitleaks {
				t.Errorf(
					"TpGitleaks = %v, want %v",
					result.TpGitleaks,
					tt.wantGitleaks,
				)
			}
			if result.TpTruffleHog != tt.wantTruffleHog {
				t.Errorf(
					"TpTruffleHog = %v, want %v",
					result.TpTruffleHog,
					tt.wantTruffleHog,
				)
			}
			if result.TpDetectSecrets != tt.wantDetectSecrets {
				t.Errorf(
					"TpDetectSecrets = %v, want %v",
					result.TpDetectSecrets,
					tt.wantDetectSecrets,
				)
			}
			if result.TpGitSecrets != tt.wantGitSecrets {
				t.Errorf(
					"TpGitSecrets = %v, want %v",
					result.TpGitSecrets,
					tt.wantGitSecrets,
				)
			}
			if result.TpGGShield != tt.wantGGShield {
				t.Errorf(
					"TpGGShield = %v, want %v",
					result.TpGGShield,
					tt.wantGGShield,
				)
			}
			if result.TpShhGit != tt.wantShhGit {
				t.Errorf(
					"TpShhGit = %v, want %v",
					result.TpShhGit,
					tt.wantShhGit,
				)
			}
			if result.TpRepoSupervisor != tt.wantRepoSupervisor {
				t.Errorf(
					"TpRepoSupervisor = %v, want %v",
					result.TpRepoSupervisor,
					tt.wantRepoSupervisor,
				)
			}

			// Verify paths if specified
			for tool, expectedPaths := range tt.wantPaths {
				var actualPaths []string
				switch tool {
				case "gitleaks":
					actualPaths = result.TpGitleaksPaths
				case "trufflehog":
					actualPaths = result.TpTruffleHogPaths
				case "detect-secrets":
					actualPaths = result.TpDetectSecretsPaths
				}
				if !cmp.Equal(actualPaths, expectedPaths) {
					t.Errorf("Paths for %s = %v, want %v", tool, actualPaths, expectedPaths)
				}
			}
		})
	}
}

func TestSecretScanningHandler_MultipleWorkflows(t *testing.T) {
	t.Parallel()
	httpClient := &http.Client{
		Transport: secretScanningRoundTripper{
			securityAnalysis: &github.SecurityAndAnalysis{
				SecretScanning: &github.SecretScanning{
					Status: github.String("enabled"),
				},
			},
			alertsStatus: http.StatusOK,
			workflowFiles: map[string]string{
				"scan1.yml": "steps:\n  - uses: gitleaks/gitleaks-action@v2",
				"scan2.yml": "steps:\n  - run: trufflehog scan",
			},
		},
	}
	ghClient := github.NewClient(httpClient)
	ctx := context.Background()

	handler := &secretScanningHandler{
		ctx: ctx,
		gh:  ghClient,
		repourl: &Repo{
			owner: "test-owner",
			repo:  "test-repo",
		},
	}

	result, err := handler.get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.TpGitleaks {
		t.Error("Expected gitleaks to be detected")
	}
	if !result.TpTruffleHog {
		t.Error("Expected trufflehog to be detected")
	}
	if result.NativeEnabled != clients.TriTrue {
		t.Errorf(
			"Expected native scanning enabled, got %v",
			result.NativeEnabled,
		)
	}
}

func TestSecretScanningHandler_Evidence(t *testing.T) {
	t.Parallel()
	httpClient := &http.Client{
		Transport: secretScanningRoundTripper{
			securityAnalysis: &github.SecurityAndAnalysis{
				SecretScanning: &github.SecretScanning{
					Status: github.String("enabled"),
				},
			},
			alertsStatus: http.StatusOK,
			workflowFiles: map[string]string{
				"scan.yml": "steps:\n  - uses: gitleaks/gitleaks-action@v2",
			},
		},
	}
	ghClient := github.NewClient(httpClient)
	ctx := context.Background()

	handler := &secretScanningHandler{
		ctx: ctx,
		gh:  ghClient,
		repourl: &Repo{
			owner: "test-owner",
			repo:  "test-repo",
		},
	}

	result, err := handler.get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Evidence) == 0 {
		t.Error("Expected evidence to be collected")
	}

	// Should have evidence about security analysis
	hasSecurityEvidence := false
	for _, e := range result.Evidence {
		if strings.Contains(e, "security_and_analysis") {
			hasSecurityEvidence = true
			break
		}
	}
	if !hasSecurityEvidence {
		t.Error("Expected security_and_analysis evidence")
	}
}
