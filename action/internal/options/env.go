// Copyright OpenSSF Authors
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

package options

import (
	"errors"
	"fmt"
)

// Environment variables.
// TODO(env): Remove once environment variables are not used for config.
//nolint:revive,nolintlint
const (
	EnvEnableSarif             = "ENABLE_SARIF"
	EnvEnableLicense           = "ENABLE_LICENSE"
	EnvEnableDangerousWorkflow = "ENABLE_DANGEROUS_WORKFLOW"
	EnvGithubEventPath         = "GITHUB_EVENT_PATH"
	EnvGithubEventName         = "GITHUB_EVENT_NAME"
	EnvGithubRepository        = "GITHUB_REPOSITORY"
	EnvGithubRef               = "GITHUB_REF"
	EnvGithubWorkspace         = "GITHUB_WORKSPACE"
	EnvGithubAuthToken         = "GITHUB_AUTH_TOKEN" //nolint:gosec
	EnvScorecardFork           = "SCORECARD_IS_FORK"
	EnvScorecardPrivateRepo    = "SCORECARD_PRIVATE_REPOSITORY"

	// TODO(input): INPUT_ constants should be removed in a future release once
	//              they have replacements in upstream scorecard.
	EnvInputRepoToken      = "INPUT_REPO_TOKEN" //nolint:gosec
	EnvInputResultsFile    = "INPUT_RESULTS_FILE"
	EnvInputResultsFormat  = "INPUT_RESULTS_FORMAT"
	EnvInputPublishResults = "INPUT_PUBLISH_RESULTS"
)

// Errors

var (
	// Errors.
	errEmptyGitHubAuthToken = errEnvVarIsEmptyWithKey(EnvGithubAuthToken)

	errEnvVarIsEmpty = errors.New("env var is empty")
)

func errEnvVarIsEmptyWithKey(envVar string) error {
	return fmt.Errorf("%w: %s", errEnvVarIsEmpty, envVar)
}
