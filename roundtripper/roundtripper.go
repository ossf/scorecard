// Copyright 2020 Security Scorecard Authors
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

package roundtripper

import (
	"context"
	"log"
	"net/http"
	"os"

	"go.uber.org/zap"
)

// nolinter
// GithubAuthToken is for making requests to GiHub's API.
var GithubAuthTokens = []string{"GITHUB_AUTH_TOKEN", "GITHUB_TOKEN", "GH_TOKEN"}

const (

	// GithubAppKeyPath is the path to file for GitHub App key.
	GithubAppKeyPath = "GITHUB_APP_KEY_PATH"
	// GithubAppID is the app ID for the GitHub App.
	GithubAppID = "GITHUB_APP_ID"
	// GithubAppInstallationID is the installation ID for the GitHub App.
	GithubAppInstallationID = "GITHUB_APP_INSTALLATION_ID"
)

func readGitHubToken() (string, bool) {
	for _, name := range GithubAuthTokens {
		if token, exists := os.LookupEnv(name); exists && token != "" {
			return token, exists
		}
	}
	return "", false
}

// NewTransport returns a configured http.Transport for use with GitHub.
func NewTransport(ctx context.Context, logger *zap.SugaredLogger) http.RoundTripper {
	transport := http.DefaultTransport

	token, exists := readGitHubToken()
	if !exists {
		log.Fatalf("No GitHub token env var is not set. " +
			"Please set this to your Github PAT before running " +
			"this command as detail in https://github.com/ossf/scorecard#authentication")
	}

	return MakeCensusTransport(MakeRateLimitedTransport(transport, logger))
}
