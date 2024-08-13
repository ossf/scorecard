// Copyright 2023 OpenSSF Scorecard Authors
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

package runner

import (
	"context"
	"errors"

	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/clients/githubrepo"
	"github.com/ossf/scorecard/v5/clients/gitlabrepo"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/log"
	"github.com/ossf/scorecard/v5/pkg/scorecard"
)

// Runner holds the clients and configuration needed to run Scorecard on multiple repos.
type Runner struct {
	ctx           context.Context
	githubClient  clients.RepoClient
	gitlabClient  clients.RepoClient
	logger        *log.Logger
	enabledChecks []string
}

// Creates a Runner which will run the listed checks. If no checks are provided, all will run.
func New(enabledChecks []string) Runner {
	ctx := context.Background()
	logger := log.NewLogger(log.DefaultLevel)
	gitlabClient, err := gitlabrepo.CreateGitlabClient(ctx, "gitlab.com")
	if err != nil {
		logger.Error(err, "creating gitlab client")
	}
	return Runner{
		ctx:           ctx,
		logger:        logger,
		githubClient:  githubrepo.CreateGithubRepoClient(ctx, logger),
		gitlabClient:  gitlabClient,
		enabledChecks: enabledChecks,
	}
}

//nolint:wrapcheck
func (r *Runner) Run(repoURI string) (scorecard.Result, error) {
	r.log("processing repo: " + repoURI)
	repoClient := r.githubClient
	repo, err := githubrepo.MakeGithubRepo(repoURI)
	if errors.Is(err, sce.ErrUnsupportedHost) {
		repo, err = gitlabrepo.MakeGitlabRepo(repoURI)
		repoClient = r.gitlabClient
	}
	if err != nil {
		return scorecard.Result{}, err
	}
	return scorecard.Run(r.ctx, repo,
		scorecard.WithRepoClient(repoClient),
		scorecard.WithChecks(r.enabledChecks),
	)
}

// logs only if logger is set.
func (r *Runner) log(msg string) {
	if r.logger != nil {
		r.logger.Info(msg)
	}
}
