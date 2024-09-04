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
	"os"

	"github.com/ossf/scorecard/v5/clients"
	azdorepo "github.com/ossf/scorecard/v5/clients/azuredevopsrepo"
	ghrepo "github.com/ossf/scorecard/v5/clients/githubrepo"
	glrepo "github.com/ossf/scorecard/v5/clients/gitlabrepo"
	"github.com/ossf/scorecard/v5/clients/localdir"
	"github.com/ossf/scorecard/v5/clients/ossfuzz"
	"github.com/ossf/scorecard/v5/internal/packageclient"
	"github.com/ossf/scorecard/v5/log"
)

// GetClients returns a list of clients for running scorecard checks.
// TODO(repo): Pass a `http.RoundTripper` here.
func GetClients(ctx context.Context, repoURI, localURI string, logger *log.Logger) (
	clients.Repo, // repo
	clients.RepoClient, // repoClient
	clients.RepoClient, // ossFuzzClient
	clients.CIIBestPracticesClient, // ciiClient
	clients.VulnerabilitiesClient, // vulnClient
	packageclient.ProjectPackageClient, // projectClient
	error,
) {
	var repo clients.Repo
	var makeRepoError error

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
			clients.DefaultVulnerabilitiesClient(), /*vulnClient*/
			nil,
			retErr
	}

	_, experimental := os.LookupEnv("SCORECARD_EXPERIMENTAL")
	var repoClient clients.RepoClient

	repo, makeRepoError = glrepo.MakeGitlabRepo(repoURI)
	if repo != nil && makeRepoError == nil {
		repoClient, makeRepoError = glrepo.CreateGitlabClient(ctx, repo.Host())
	}

	if experimental && (makeRepoError != nil || repo == nil) {
		repo, makeRepoError = azdorepo.MakeAzureDevOpsRepo(repoURI)
		if repo != nil && makeRepoError == nil {
			repoClient, makeRepoError = azdorepo.CreateAzureDevOpsClient(ctx, repo)
		}
	}

	if makeRepoError != nil || repo == nil {
		repo, makeRepoError = ghrepo.MakeGithubRepo(repoURI)
		if makeRepoError != nil {
			return repo,
				nil,
				nil,
				nil,
				nil,
				packageclient.CreateDepsDevClient(),
				fmt.Errorf("error making github repo: %w", makeRepoError)
		}
		repoClient = ghrepo.CreateGithubRepoClient(ctx, logger)
	}

	return repo, /*repo*/
		repoClient, /*repoClient*/
		ossfuzz.CreateOSSFuzzClient(ossfuzz.StatusURL), /*ossFuzzClient*/
		clients.DefaultCIIBestPracticesClient(), /*ciiClient*/
		clients.DefaultVulnerabilitiesClient(), /*vulnClient*/
		packageclient.CreateDepsDevClient(),
		nil
}
