// Copyright 2022 OpenSSF Scorecard Authors
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

package checker

import (
	"context"
	"fmt"

	"github.com/ossf/scorecard/v4/clients"
	ghrepo "github.com/ossf/scorecard/v4/clients/githubrepo"
	"github.com/ossf/scorecard/v4/clients/localdir"
	"github.com/ossf/scorecard/v4/log"
)

// GetClients returns a list of clients for running scorecard checks.
// TODO(repo): Pass a `http.RoundTripper` here.
func GetClients(ctx context.Context, repoURI, localURI string, logger *log.Logger) (
	clients.Repo, // repo
	clients.RepoClient, // repoClient
	clients.RepoClient, // ossFuzzClient
	clients.CIIBestPracticesClient, // ciiClient
	clients.VulnerabilitiesClient, // vulnClient
	error,
) {
	var githubRepo clients.Repo
	if localURI != "" {
		localRepo, errLocal := localdir.MakeLocalDirRepo(localURI)
		var retErr error
		if errLocal != nil {
			retErr = fmt.Errorf("getting local directory client: %w", errLocal)
		}
		return localRepo, /*repo*/
			localdir.CreateLocalDirClient(ctx, logger), /*repoClient*/
			nil, /*ossFuzzClient*/
			nil, /*ciiClient*/
			nil, /*vulnClient*/
			retErr
	}

	githubRepo, errGitHub := ghrepo.MakeGithubRepo(repoURI)
	if errGitHub != nil {
		return githubRepo,
			nil,
			nil,
			nil,
			nil,
			fmt.Errorf("getting local directory client: %w", errGitHub)
	}

	ossFuzzRepoClient, errOssFuzz := ghrepo.CreateOssFuzzRepoClient(ctx, logger)
	var retErr error
	if errOssFuzz != nil {
		retErr = fmt.Errorf("getting OSS-Fuzz repo client: %w", errOssFuzz)
	}
	// TODO(repo): Should we be handling the OSS-Fuzz client error like this?
	return githubRepo, /*repo*/
		ghrepo.CreateGithubRepoClient(ctx, logger), /*repoClient*/
		ossFuzzRepoClient, /*ossFuzzClient*/
		clients.DefaultCIIBestPracticesClient(), /*ciiClient*/
		clients.DefaultVulnerabilitiesClient(), /*vulnClient*/
		retErr
}
