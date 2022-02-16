// Copyright 2022 Allstar Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package repo

import (
	"context"

	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo"
	"github.com/ossf/scorecard/v4/clients/localdir"
	sclog "github.com/ossf/scorecard/v4/log"
)

// TODO(repo): Pass a `http.RoundTripper` here
func GetClients(ctx context.Context, repoURI, localURI string, logger *sclog.Logger) (
	clients.Repo, // repo
	clients.RepoClient, // repoClient
	clients.RepoClient, // ossFuzzClient
	clients.CIIBestPracticesClient, // ciiClient
	clients.VulnerabilitiesClient, // vulnClient
	error) {
	var githubRepo clients.Repo
	var errGitHub error
	if localURI != "" {
		localRepo, errLocal := localdir.MakeLocalDirRepo(localURI)
		return localRepo, /*repo*/
			localdir.CreateLocalDirClient(ctx, logger), /*repoClient*/
			nil, /*ossFuzzClient*/
			nil, /*ciiClient*/
			nil, /*vulnClient*/
			errLocal
	}

	githubRepo, errGitHub = githubrepo.MakeGithubRepo(repoURI)
	if errGitHub != nil {
		// nolint: wrapcheck
		return githubRepo,
			nil,
			nil,
			nil,
			nil,
			errGitHub
	}

	ossFuzzRepoClient, errOssFuzz := githubrepo.CreateOssFuzzRepoClient(ctx, logger)
	return githubRepo, /*repo*/
		githubrepo.CreateGithubRepoClient(ctx, logger), /*repoClient*/
		ossFuzzRepoClient, /*ossFuzzClient*/
		clients.DefaultCIIBestPracticesClient(), /*ciiClient*/
		clients.DefaultVulnerabilitiesClient(), /*vulnClient*/
		errOssFuzz
}
