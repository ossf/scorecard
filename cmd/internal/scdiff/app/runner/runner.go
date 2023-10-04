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
	"strings"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo"
	"github.com/ossf/scorecard/v4/clients/ossfuzz"
	"github.com/ossf/scorecard/v4/log"
	"github.com/ossf/scorecard/v4/pkg"
)

const (
	commit      = clients.HeadSHA
	commitDepth = 0 // default
)

// Runner holds the clients and configuration needed to run Scorecard on multiple repos.
type Runner struct {
	ctx           context.Context
	logger        *log.Logger
	enabledChecks checker.CheckNameToFnMap
	repoClient    clients.RepoClient
	ossFuzz       clients.RepoClient
	cii           clients.CIIBestPracticesClient
	vuln          clients.VulnerabilitiesClient
}

// Creates a Runner which will run the listed checks. If no checks are provided, all will run.
func New(enabledChecks []string) Runner {
	ctx := context.Background()
	logger := log.NewLogger(log.DefaultLevel)
	return Runner{
		ctx:           ctx,
		logger:        logger,
		repoClient:    githubrepo.CreateGithubRepoClient(ctx, logger),
		ossFuzz:       ossfuzz.CreateOSSFuzzClient(ossfuzz.StatusURL),
		cii:           clients.DefaultCIIBestPracticesClient(),
		vuln:          clients.DefaultVulnerabilitiesClient(),
		enabledChecks: parseChecks(enabledChecks),
	}
}

//nolint:wrapcheck
func (r *Runner) Run(repoURI string) (pkg.ScorecardResult, error) {
	r.log("processing repo: " + repoURI)
	// TODO (gitlab?)
	repo, err := githubrepo.MakeGithubRepo(repoURI)
	if err != nil {
		return pkg.ScorecardResult{}, err
	}
	return pkg.RunScorecard(r.ctx, repo, commit, commitDepth, r.enabledChecks, r.repoClient, r.ossFuzz, r.cii, r.vuln)
}

// logs only if logger is set.
func (r *Runner) log(msg string) {
	if r.logger != nil {
		r.logger.Info(msg)
	}
}

func parseChecks(c []string) checker.CheckNameToFnMap {
	all := checks.GetAll()
	if len(c) == 0 {
		return all
	}

	ret := checker.CheckNameToFnMap{}
	for _, requested := range c {
		for key, fn := range all {
			if strings.EqualFold(key, requested) {
				ret[key] = fn
			}
		}
	}
	return ret
}
