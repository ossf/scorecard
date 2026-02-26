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
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/google/go-github/v53/github"

	"github.com/ossf/scorecard/v5/clients"
	sce "github.com/ossf/scorecard/v5/errors"
)

const (
	githubFileType = "file"
)

type secretScanningHandler struct {
	errSetup              error
	ctx                   context.Context
	gh                    *github.Client
	repourl               *Repo
	tpGitleaksPaths       []string
	evidence              []string
	tpRepoSupervisorPaths []string
	tpShhGitPaths         []string
	tpGGShieldPaths       []string
	tpGitSecretsPaths     []string
	tpDetectSecretsPaths  []string
	tpTruffleHogPaths     []string
	nativeEnabled         clients.TriState
	pushProtectionEnabled clients.TriState
	once                  sync.Once
	tpRepoSupervisor      bool
	tpShhGit              bool
	tpGGShield            bool
	tpGitSecrets          bool
	tpDetectSecrets       bool
	tpTruffleHog          bool
	tpGitleaks            bool
}

func (h *secretScanningHandler) init(ctx context.Context, repourl *Repo, gh *github.Client) {
	h.ctx = ctx
	h.repourl = repourl
	h.gh = gh
}

func (h *secretScanningHandler) setup() error {
	h.once.Do(func() {
		if err := h.fetchRepoSecurityAndAnalysis(); err != nil {
			h.errSetup = err
			return
		}
		if err := h.scanActionsWorkflows(); err != nil {
			h.errSetup = err
			return
		}
	})
	return h.errSetup
}

func (h *secretScanningHandler) get() (struct {
	TpTruffleHogPaths     []string
	Evidence              []string
	TpRepoSupervisorPaths []string
	TpShhGitPaths         []string
	TpGGShieldPaths       []string
	TpGitSecretsPaths     []string
	TpDetectSecretsPaths  []string
	TpGitleaksPaths       []string
	PushProtectionEnabled clients.TriState
	NativeEnabled         clients.TriState
	TpDetectSecrets       bool
	TpRepoSupervisor      bool
	TpShhGit              bool
	TpGGShield            bool
	TpGitSecrets          bool
	TpTruffleHog          bool
	TpGitleaks            bool
}, error,
) {
	if err := h.setup(); err != nil {
		return struct {
			TpTruffleHogPaths     []string
			Evidence              []string
			TpRepoSupervisorPaths []string
			TpShhGitPaths         []string
			TpGGShieldPaths       []string
			TpGitSecretsPaths     []string
			TpDetectSecretsPaths  []string
			TpGitleaksPaths       []string
			PushProtectionEnabled clients.TriState
			NativeEnabled         clients.TriState
			TpDetectSecrets       bool
			TpRepoSupervisor      bool
			TpShhGit              bool
			TpGGShield            bool
			TpGitSecrets          bool
			TpTruffleHog          bool
			TpGitleaks            bool
		}{}, err
	}
	return struct {
		TpTruffleHogPaths     []string
		Evidence              []string
		TpRepoSupervisorPaths []string
		TpShhGitPaths         []string
		TpGGShieldPaths       []string
		TpGitSecretsPaths     []string
		TpDetectSecretsPaths  []string
		TpGitleaksPaths       []string
		PushProtectionEnabled clients.TriState
		NativeEnabled         clients.TriState
		TpDetectSecrets       bool
		TpRepoSupervisor      bool
		TpShhGit              bool
		TpGGShield            bool
		TpGitSecrets          bool
		TpTruffleHog          bool
		TpGitleaks            bool
	}{
		NativeEnabled:         h.nativeEnabled,
		PushProtectionEnabled: h.pushProtectionEnabled,

		TpGitleaks:       h.tpGitleaks,
		TpTruffleHog:     h.tpTruffleHog,
		TpDetectSecrets:  h.tpDetectSecrets,
		TpGitSecrets:     h.tpGitSecrets,
		TpGGShield:       h.tpGGShield,
		TpShhGit:         h.tpShhGit,
		TpRepoSupervisor: h.tpRepoSupervisor,

		TpGitleaksPaths:       append([]string{}, h.tpGitleaksPaths...),
		TpTruffleHogPaths:     append([]string{}, h.tpTruffleHogPaths...),
		TpDetectSecretsPaths:  append([]string{}, h.tpDetectSecretsPaths...),
		TpGitSecretsPaths:     append([]string{}, h.tpGitSecretsPaths...),
		TpGGShieldPaths:       append([]string{}, h.tpGGShieldPaths...),
		TpShhGitPaths:         append([]string{}, h.tpShhGitPaths...),
		TpRepoSupervisorPaths: append([]string{}, h.tpRepoSupervisorPaths...),

		Evidence: append([]string{}, h.evidence...),
	}, nil
}

// --- internals ---

//nolint:nestif // Unavoidable nested logic for checking security settings
func (h *secretScanningHandler) fetchRepoSecurityAndAnalysis() error {
	repo, _, err := h.gh.Repositories.Get(h.ctx, h.repourl.owner, h.repourl.repo)
	if err != nil {
		// Check if this is a permission error
		if strings.Contains(err.Error(), "403") ||
			strings.Contains(err.Error(), "forbidden") {
			// Token doesn't have permissions - leave as unknown but don't fail
			h.evidence = append(h.evidence, "source:permission_denied")
			return nil
		}
		return sce.WithMessage(
			sce.ErrScorecardInternal,
			fmt.Sprintf("Repositories.Get: %v", err),
		)
	}
	if sa := repo.GetSecurityAndAnalysis(); sa != nil {
		if sa.SecretScanning != nil && sa.SecretScanning.Status != nil {
			if sa.SecretScanning.GetStatus() == "enabled" {
				h.nativeEnabled = clients.TriTrue
			} else {
				h.nativeEnabled = clients.TriFalse
			}
		}
		if sa.SecretScanningPushProtection != nil &&
			sa.SecretScanningPushProtection.Status != nil {
			if sa.SecretScanningPushProtection.GetStatus() == "enabled" {
				h.pushProtectionEnabled = clients.TriTrue
			} else {
				h.pushProtectionEnabled = clients.TriFalse
			}
		}
		if h.nativeEnabled != clients.TriUnknown ||
			h.pushProtectionEnabled != clients.TriUnknown {
			h.evidence = append(h.evidence, "source:security_and_analysis")
			return nil
		}
	}

	// Fallback inference: check if secret scanning alerts endpoint is accessible
	// This helps detect if secret scanning is enabled when SecurityAndAnalysis is not available
	_, resp, err := h.gh.SecretScanning.ListAlertsForRepo(
		h.ctx,
		h.repourl.owner,
		h.repourl.repo,
		&github.SecretScanningAlertListOptions{
			ListOptions: github.ListOptions{PerPage: 1},
		},
	)
	if resp != nil && resp.StatusCode == http.StatusOK {
		h.nativeEnabled = clients.TriTrue
		h.evidence = append(h.evidence, "source:alerts-200")
	} else if resp != nil && h.nativeEnabled == clients.TriUnknown {
		h.evidence = append(h.evidence, fmt.Sprintf("source:alerts-%d", resp.StatusCode))
	}
	// Ignore errors from alerts endpoint - treat as unknown
	_ = err
	return nil
}

func (h *secretScanningHandler) scanActionsWorkflows() error {
	// List .github/workflows directory
	_, directoryContent, _, err := h.gh.Repositories.GetContents(
		h.ctx,
		h.repourl.owner,
		h.repourl.repo,
		".github/workflows",
		&github.RepositoryContentGetOptions{},
	)
	if err != nil {
		// 404/no workflows directory is fine
		return nil //nolint:nilerr // Intentionally ignoring error for missing workflows
	}

	for _, e := range directoryContent {
		if e == nil || e.GetType() != githubFileType {
			continue
		}
		rc, _, _, err := h.gh.Repositories.GetContents(
			h.ctx,
			h.repourl.owner,
			h.repourl.repo,
			e.GetPath(),
			&github.RepositoryContentGetOptions{},
		)
		if err != nil || rc == nil {
			continue
		}
		content, err := rc.GetContent()
		if err != nil {
			continue
		}
		low := strings.ToLower(content)
		p := e.GetPath()

		setIfContains(
			&h.tpGitleaks, &h.tpGitleaksPaths, low, p,
			[]string{"gitleaks", "gitleaks/gitleaks", "gitleaks-action"},
			&h.evidence, "workflow:gitleaks",
		)
		setIfContains(
			&h.tpTruffleHog, &h.tpTruffleHogPaths, low, p,
			[]string{"trufflehog", "trufflesecurity/trufflehog"},
			&h.evidence, "workflow:trufflehog",
		)
		setIfContains(
			&h.tpDetectSecrets, &h.tpDetectSecretsPaths, low, p,
			[]string{"detect-secrets"},
			&h.evidence, "workflow:detect-secrets",
		)
		setIfContains(
			&h.tpGitSecrets, &h.tpGitSecretsPaths, low, p,
			[]string{"git-secrets"},
			&h.evidence, "workflow:git-secrets",
		)
		setIfContains(
			&h.tpGGShield, &h.tpGGShieldPaths, low, p,
			[]string{"ggshield", "gitguardian/ggshield"},
			&h.evidence, "workflow:ggshield",
		)
		setIfContains(
			&h.tpShhGit, &h.tpShhGitPaths, low, p,
			[]string{"shhgit"},
			&h.evidence, "workflow:shhgit",
		)
		setIfContains(
			&h.tpRepoSupervisor, &h.tpRepoSupervisorPaths, low, p,
			[]string{"repo-supervisor"},
			&h.evidence, "workflow:repo-supervisor",
		)
	}
	return nil
}

func setIfContains(
	flag *bool, paths *[]string, hay, filePath string,
	needles []string, ev *[]string, tag string,
) {
	for _, n := range needles {
		if strings.Contains(hay, strings.ToLower(n)) {
			*flag = true
			*paths = append(*paths, filePath)
			*ev = append(*ev, tag+"@"+filePath)
			return
		}
	}
}
