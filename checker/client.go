// Copyright 2022 Security Scorecard Authors
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
	"strings"

	"github.com/ossf/scorecard/v4/clients"
	ghrepo "github.com/ossf/scorecard/v4/clients/githubrepo"
	glrepo "github.com/ossf/scorecard/v4/clients/gitlabrepo"
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
			nil, /*vulnClient*/
			retErr
	}

	if strings.Contains(repoURI, "gitlab.") {
		repo, makeRepoError = glrepo.MakeGitlabRepo(repoURI)
		if makeRepoError != nil {
			return repo,
				nil,
				nil,
				nil,
				nil,
				fmt.Errorf("getting local directory client: %w", makeRepoError)
		}
	} else {
		repo, makeRepoError = ghrepo.MakeGithubRepo(repoURI)
		if makeRepoError != nil {
			return repo,
				nil,
				nil,
				nil,
				nil,
				fmt.Errorf("getting local directory client: %w", makeRepoError)
		}
	}

	ossFuzzRepoClient, errOssFuzz := ghrepo.CreateOssFuzzRepoClient(ctx, logger)
	var retErr error
	if errOssFuzz != nil {
		retErr = fmt.Errorf("getting OSS-Fuzz repo client: %w", errOssFuzz)
	}
	// TODO(repo): Should we be handling the OSS-Fuzz client error like this?
	if strings.Contains(repoURI, "gitlab.") {
		glClient, err := glrepo.CreateGitlabClientWithToken(ctx, os.Getenv("GITLAB_AUTH_TOKEN"), repo)
		if err != nil {
			return repo,
				nil,
				nil,
				nil,
				nil,
				fmt.Errorf("error creating gitlab client: %w", err)
		}
		return repo, /*repo*/
			glClient, /*repoClient*/
			ossFuzzRepoClient, /*ossFuzzClient*/
			clients.DefaultCIIBestPracticesClient(), /*ciiClient*/
			clients.DefaultVulnerabilitiesClient(), /*vulnClient*/
			retErr
	} else {
		return repo, /*repo*/
			ghrepo.CreateGithubRepoClient(ctx, logger), /*repoClient*/
			ossFuzzRepoClient, /*ossFuzzClient*/
			clients.DefaultCIIBestPracticesClient(), /*ciiClient*/
			clients.DefaultVulnerabilitiesClient(), /*vulnClient*/
			retErr
	}
}
