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
	"strconv"
	"strings"

	"github.com/bradleyfalzon/ghinstallation"
	"go.uber.org/zap"

	"github.com/ossf/scorecard/clients/githubrepo"
)

// GithubAuthTokens are for making requests to GiHub's API.
var GithubAuthTokens = []string{"GITHUB_AUTH_TOKEN", "GITHUB_TOKEN", "GH_TOKEN", "GH_AUTH_TOKEN"}

const (

	// GithubAppKeyPath is the path to file for GitHub App key.
	GithubAppKeyPath = "GITHUB_APP_KEY_PATH"
	// GithubAppID is the app ID for the GitHub App.
	GithubAppID = "GITHUB_APP_ID"
	// GithubAppInstallationID is the installation ID for the GitHub App.
	GithubAppInstallationID = "GITHUB_APP_INSTALLATION_ID"
)

func readGitHubTokens() (string, bool) {
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

	// nolinter
	if token, exists := readGitHubTokens(); exists {
		// Use GitHub PAT
		transport = githubrepo.MakeGitHubTransport(transport, strings.Split(token, ","))
	} else if keyPath := os.Getenv(GithubAppKeyPath); keyPath != "" { // Also try a GITHUB_APP
		appID, err := strconv.Atoi(os.Getenv(GithubAppID))
		if err != nil {
			log.Panic(err)
		}
		installationID, err := strconv.Atoi(os.Getenv(GithubAppInstallationID))
		if err != nil {
			log.Panic(err)
		}
		transport, err = ghinstallation.NewKeyFromFile(transport, int64(appID), int64(installationID), keyPath)
		if err != nil {
			log.Panic(err)
		}
	} else {
		log.Fatalf("GitHub token env var is not set. " +
			"Please read https://github.com/ossf/scorecard#authentication")
	}

	return MakeCensusTransport(MakeRateLimitedTransport(transport, logger))
}
