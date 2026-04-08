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

package clients

// PlatformType identifies the hosting platform.
type PlatformType string

const (
	PlatformGitHub PlatformType = "github"
	PlatformGitLab PlatformType = "gitlab"
)

// TriState is a three-valued boolean.
type TriState int

const (
	TriUnknown TriState = iota
	TriFalse
	TriTrue
)

type SecretScanningSignals struct {
	Platform                      PlatformType
	ThirdPartyDetectSecretsPaths  []string
	Evidence                      []string
	ThirdPartyRepoSupervisorPaths []string
	ThirdPartyShhGitPaths         []string
	ThirdPartyGitleaksPaths       []string
	ThirdPartyGGShieldPaths       []string
	ThirdPartyTruffleHogPaths     []string
	ThirdPartyGitSecretsPaths     []string
	GHNativeEnabled               TriState
	GHPushProtectionEnabled       TriState
	GLPushRulesPreventSecrets     bool
	ThirdPartyGitSecrets          bool
	ThirdPartyDetectSecrets       bool
	ThirdPartyGGShield            bool
	ThirdPartyTruffleHog          bool
	ThirdPartyShhGit              bool
	ThirdPartyGitleaks            bool
	ThirdPartyRepoSupervisor      bool
	GLSecretPushProtection        bool
	GLPipelineSecretDetection     bool
}
