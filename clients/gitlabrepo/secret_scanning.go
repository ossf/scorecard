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
	"fmt"
	"net/http"
	"strings"
	"sync"

	gitlab "gitlab.com/gitlab-org/api/client-go"
	"gopkg.in/yaml.v3"

	sce "github.com/ossf/scorecard/v5/errors"
)

// secretScanningHandler aggregates all GitLab-side signals related to secret scanning:
// - Project Security Settings: Secret Push Protection
// - Push rules: prevent_secrets
// - CI config (.gitlab-ci.yml): native pipeline secret detection + third-party scanners
//
// It fetches once (sync.Once), organizes the data into simple booleans and path slices,
// and exposes them via get().
type secretScanningHandler struct {
	errSetup                error
	ctx                     context.Context
	repourl                 *Repo
	glClient                *gitlab.Client
	tpGGShieldPaths         []string
	evidence                []string
	tpRepoSupervisorPaths   []string
	tpShhGitPaths           []string
	tpGitleaksPaths         []string
	tpGitSecretsPaths       []string
	tpDetectSecretsPaths    []string
	tpTruffleHogPaths       []string
	once                    sync.Once
	secretPushProtection    bool
	tpRepoSupervisor        bool
	tpShhGit                bool
	tpGGShield              bool
	tpGitSecrets            bool
	tpDetectSecrets         bool
	tpTruffleHog            bool
	tpGitleaks              bool
	pipelineSecretDetection bool
	pushRulesPreventSecrets bool
}

func (h *secretScanningHandler) init(ctx context.Context, gl *gitlab.Client, repourl *Repo) {
	h.ctx = ctx
	h.glClient = gl
	h.repourl = repourl
}

func (h *secretScanningHandler) setup() error {
	h.once.Do(func() {
		if err := h.fetchProjectSPP(); err != nil {
			h.errSetup = err
			return
		}
		if err := h.fetchPushRules(); err != nil {
			h.errSetup = err
			return
		}
		if err := h.fetchCIConfig(); err != nil {
			h.errSetup = err
			return
		}
	})
	return h.errSetup
}

func (h *secretScanningHandler) get() (struct {
	TpGitleaksPaths         []string
	Evidence                []string
	TpRepoSupervisorPaths   []string
	TpShhGitPaths           []string
	TpGGShieldPaths         []string
	TpGitSecretsPaths       []string
	TpDetectSecretsPaths    []string
	TpTruffleHogPaths       []string
	TpTruffleHog            bool
	TpRepoSupervisor        bool
	TpShhGit                bool
	TpGGShield              bool
	TpGitSecrets            bool
	TpDetectSecrets         bool
	SecretPushProtection    bool
	TpGitleaks              bool
	PipelineSecretDetection bool
	PushRulesPreventSecrets bool
}, error,
) {
	if err := h.setup(); err != nil {
		return struct {
			TpGitleaksPaths         []string
			Evidence                []string
			TpRepoSupervisorPaths   []string
			TpShhGitPaths           []string
			TpGGShieldPaths         []string
			TpGitSecretsPaths       []string
			TpDetectSecretsPaths    []string
			TpTruffleHogPaths       []string
			TpTruffleHog            bool
			TpRepoSupervisor        bool
			TpShhGit                bool
			TpGGShield              bool
			TpGitSecrets            bool
			TpDetectSecrets         bool
			SecretPushProtection    bool
			TpGitleaks              bool
			PipelineSecretDetection bool
			PushRulesPreventSecrets bool
		}{}, err
	}

	return struct {
		TpGitleaksPaths         []string
		Evidence                []string
		TpRepoSupervisorPaths   []string
		TpShhGitPaths           []string
		TpGGShieldPaths         []string
		TpGitSecretsPaths       []string
		TpDetectSecretsPaths    []string
		TpTruffleHogPaths       []string
		TpTruffleHog            bool
		TpRepoSupervisor        bool
		TpShhGit                bool
		TpGGShield              bool
		TpGitSecrets            bool
		TpDetectSecrets         bool
		SecretPushProtection    bool
		TpGitleaks              bool
		PipelineSecretDetection bool
		PushRulesPreventSecrets bool
	}{
		SecretPushProtection:    h.secretPushProtection,
		PushRulesPreventSecrets: h.pushRulesPreventSecrets,
		PipelineSecretDetection: h.pipelineSecretDetection,

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

// fetchProjectSPP queries the Project Security Settings API
// to read Secret Push Protection.
func (h *secretScanningHandler) fetchProjectSPP() error {
	settings, resp, err := h.glClient.ProjectSecuritySettings.ListProjectSecuritySettings(
		h.repourl.projectID,
	)
	if err != nil {
		if resp != nil && resp.Response != nil && resp.StatusCode == http.StatusNotFound {
			// Feature not available / not set; not an error.
			return nil
		}
		return sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("ProjectSecuritySettings.ListProjectSecuritySettings: %v", err))
	}
	if settings != nil && settings.SecretPushProtectionEnabled {
		h.secretPushProtection = true
		h.evidence = append(h.evidence, "project.security_settings.secret_push_protection_enabled=true")
	}
	return nil
}

// fetchPushRules reads project push rules and picks
// the prevent_secrets boolean.
func (h *secretScanningHandler) fetchPushRules() error {
	r, resp, err := h.glClient.Projects.GetProjectPushRules(h.repourl.projectID)
	if err != nil {
		if resp != nil && resp.Response != nil && resp.StatusCode == http.StatusNotFound {
			// No push rules configured.
			return nil
		}
		return sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("Projects.GetProjectPushRules: %v", err))
	}
	if r != nil && r.PreventSecrets {
		h.pushRulesPreventSecrets = true
		h.evidence = append(h.evidence, "push_rule.prevent_secrets=true")
	}
	return nil
}

// fetchCIConfig downloads `.gitlab-ci.yml`, detects native secret detection include/job,
// and scans for popular third-party secret scanners. It also records the CI path.
//
//nolint:gocognit // Complex CI config parsing requires multiple checks
func (h *secretScanningHandler) fetchCIConfig() error {
	ref := h.repourl.defaultBranch
	file, resp, err := h.glClient.RepositoryFiles.GetFile(
		h.repourl.projectID,
		".gitlab-ci.yml",
		&gitlab.GetFileOptions{Ref: &ref},
	)
	if err != nil {
		if resp != nil && resp.Response != nil && resp.StatusCode == http.StatusNotFound {
			return nil
		}
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("RepositoryFiles.GetFile: %v", err))
	}

	content, err := base64.StdEncoding.DecodeString(file.Content)
	if err != nil {
		// If base64 decoding fails, treat as if file doesn't exist
		return nil //nolint:nilerr // Intentionally ignoring decode error
	}
	const ciPath = ".gitlab-ci.yml"

	type anyMap = map[string]any
	var root anyMap
	if err := yaml.Unmarshal(content, &root); err != nil {
		// Invalid YAML: ignore for detection purposes.
		return nil //nolint:nilerr // Intentionally ignoring YAML parse error
	}

	// Native template include.
	if inc, ok := root["include"]; ok {
		switch v := inc.(type) {
		case []any:
			for _, it := range v {
				if m, ok := it.(anyMap); ok {
					if tpl, ok := m["template"].(string); ok && isSecretDetectionTemplate(tpl) {
						h.pipelineSecretDetection = true
						h.evidence = append(h.evidence, "include.template="+tpl)
					}
				}
			}
		case anyMap:
			if tpl, ok := v["template"].(string); ok && isSecretDetectionTemplate(tpl) {
				h.pipelineSecretDetection = true
				h.evidence = append(h.evidence, "include.template="+tpl)
			}
		}
	}

	// Walk jobs; detect third-party tools and the native conventional job.
	for k, v := range root {
		// Skip non-job top-level keys.
		if k == "stages" || k == "default" || k == "variables" || k == "include" || k == "workflow" {
			continue
		}
		job, ok := v.(anyMap)
		if !ok {
			continue
		}

		// Native job name found.
		if k == "secret_detection" {
			h.pipelineSecretDetection = true
			h.evidence = append(h.evidence, "job:name=secret_detection")
		}

		// Third-party scanners: script and image
		if scripts, ok := job["script"]; ok {
			setFromScripts(
				ciPath, &h.tpGitleaks, &h.tpGitleaksPaths,
				scripts, "gitleaks", &h.evidence, "script:gitleaks",
			)
			setFromScripts(
				ciPath, &h.tpTruffleHog, &h.tpTruffleHogPaths,
				scripts, "trufflehog", &h.evidence, "script:trufflehog",
			)
			setFromScripts(
				ciPath, &h.tpDetectSecrets, &h.tpDetectSecretsPaths,
				scripts, "detect-secrets", &h.evidence, "script:detect-secrets",
			)
			setFromScripts(
				ciPath, &h.tpGitSecrets, &h.tpGitSecretsPaths,
				scripts, "git-secrets", &h.evidence, "script:git-secrets",
			)
			setFromScripts(
				ciPath, &h.tpGGShield, &h.tpGGShieldPaths,
				scripts, "ggshield", &h.evidence, "script:ggshield",
			)
			setFromScripts(
				ciPath, &h.tpShhGit, &h.tpShhGitPaths,
				scripts, "shhgit", &h.evidence, "script:shhgit",
			)
			setFromScripts(
				ciPath, &h.tpRepoSupervisor, &h.tpRepoSupervisorPaths,
				scripts, "repo-supervisor", &h.evidence, "script:repo-supervisor",
			)
		}
		if img, ok := job["image"].(string); ok {
			l := strings.ToLower(img)
			checkImage(
				ciPath, &h.tpGitleaks, &h.tpGitleaksPaths, l,
				[]string{"zricethezav/gitleaks", "gitleaks/gitleaks"}, &h.evidence,
			)
			checkImage(
				ciPath, &h.tpTruffleHog, &h.tpTruffleHogPaths, l,
				[]string{"trufflesecurity/trufflehog"}, &h.evidence,
			)
			checkImage(
				ciPath, &h.tpGGShield, &h.tpGGShieldPaths, l,
				[]string{"gitguardian/ggshield"}, &h.evidence,
			)
		}
	}

	return nil
}

func isSecretDetectionTemplate(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	return s == "security/secret-detection.gitlab-ci.yml" ||
		s == "jobs/secret-detection.gitlab-ci.yml"
}

func setFromScripts(
	path string, flag *bool, paths *[]string,
	scripts any, needle string, ev *[]string, tag string,
) {
	low := strings.ToLower(needle)
	switch v := scripts.(type) {
	case []any:
		for _, line := range v {
			if str, ok := line.(string); ok && strings.Contains(strings.ToLower(str), low) {
				*flag = true
				*paths = append(*paths, path)
				*ev = append(*ev, tag+"@"+path)
				return
			}
		}
	case string:
		if strings.Contains(strings.ToLower(v), low) {
			*flag = true
			*paths = append(*paths, path)
			*ev = append(*ev, tag+"@"+path)
		}
	}
}

func checkImage(
	path string, flag *bool, paths *[]string,
	img string, needles []string, ev *[]string,
) {
	for _, nd := range needles {
		if strings.Contains(img, nd) {
			*flag = true
			*paths = append(*paths, path)
			*ev = append(*ev, "image:"+nd+"@"+path)
			return
		}
	}
}
