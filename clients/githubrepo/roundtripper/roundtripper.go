// Copyright 2020 OpenSSF Scorecard Authors
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

// Package roundtripper has implementations of http.RoundTripper useful to clients.RepoClient.
package roundtripper

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/bradleyfalzon/ghinstallation/v2"

	"github.com/ossf/scorecard/v4/clients/githubrepo/roundtripper/tokens"
	"github.com/ossf/scorecard/v4/log"
)

const (
	// githubAppKeyPath is the path to file for GitHub App key.
	githubAppKeyPath = "GITHUB_APP_KEY_PATH"
	// githubAppID is the app ID for the GitHub App.
	githubAppID = "GITHUB_APP_ID"
	// githubAppInstallationID is the installation ID for the GitHub App.
	githubAppInstallationID = "GITHUB_APP_INSTALLATION_ID"
)

// NewTransport returns a configured http.Transport for use with GitHub.
func NewTransport(ctx context.Context, logger *log.Logger) http.RoundTripper {
	transport := http.DefaultTransport

	//nolint
	if tokenAccessor := tokens.MakeTokenAccessor(); tokenAccessor != nil {
		// Use GitHub PAT
		transport = makeGitHubTransport(transport, tokenAccessor)
	} else if keyPath := os.Getenv(githubAppKeyPath); keyPath != "" { // Also try a GITHUB_APP
		appID, err := strconv.Atoi(os.Getenv(githubAppID))
		if err != nil {
			logger.Error(err, "getting GitHub application ID from environment")
		}
		installationID, err := strconv.Atoi(os.Getenv(githubAppInstallationID))
		if err != nil {
			logger.Error(err, "getting GitHub application installation ID")
		}
		transport, err = ghinstallation.NewKeyFromFile(transport, int64(appID), int64(installationID), keyPath)
		if err != nil {
			logger.Error(err, "getting a private key from file")
		}
	} else {
		// TODO(log): Improve error message
		logger.Error(fmt.Errorf("an error occurred while getting GitHub credentials"), "GitHub token env var is not set. Please read https://github.com/ossf/scorecard#authentication")
	}

	return MakeCensusTransport(MakeRateLimitedTransport(transport, logger))
}
