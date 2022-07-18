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

package dependencydiff

import (
	"context"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/log"
	"github.com/ossf/scorecard/v4/pkg"
	"github.com/ossf/scorecard/v4/policy"
)

// Depdiff is the exported name for dependency-diff.
const Depdiff = "Dependency-diff"

type dependencydiffContext struct {
	logger                                *log.Logger
	ownerName, repoName, baseSHA, headSHA string
	ctx                                   context.Context
	ghRepoClient                          clients.RepoClient
	ossFuzzClient                         clients.RepoClient
	vulnsClient                           clients.VulnerabilitiesClient
	ciiClient                             clients.CIIBestPracticesClient
	changeTypesToCheck                    map[pkg.ChangeType]bool
	checkNamesToRun                       []string
	dependencydiffs                       []dependency
	results                               []pkg.DependencyCheckResult
}

// GetDependencyDiffResults gets dependency changes between two given code commits BASE and HEAD
// along with the Scorecard check results of the dependencies, and returns a slice of DependencyCheckResult.
// TO use this API, an access token must be set following https://github.com/ossf/scorecard#authentication.
func GetDependencyDiffResults(
	ctx context.Context, ownerName, repoName, baseSHA, headSHA string, scorecardChecksNames []string,
	changeTypesToCheck map[pkg.ChangeType]bool) ([]pkg.DependencyCheckResult, error) {
	// Fetch the raw dependency diffs.
	dCtx := dependencydiffContext{
		logger:             log.NewLogger(log.InfoLevel),
		ownerName:          ownerName,
		repoName:           repoName,
		baseSHA:            baseSHA,
		headSHA:            headSHA,
		ctx:                ctx,
		changeTypesToCheck: changeTypesToCheck,
		checkNamesToRun:    scorecardChecksNames,
	}
	err := fetchRawDependencyDiffData(&dCtx)
	if err != nil {
		return nil, fmt.Errorf("error in fetchRawDependencyDiffData: %w", err)
	}

	if err != nil {
		return nil, fmt.Errorf("error in initClientByChecks: %w", err)
	}
	err = getScorecardCheckResults(&dCtx)
	if err != nil {
		return nil, fmt.Errorf("error getting scorecard check results: %w", err)
	}
	return dCtx.results, nil
}

func initClientByChecks(dCtx *dependencydiffContext, dSrcRepo string) error {
	_, repoClient, ossFuzzClient, ciiClient, vulnsClient, err := checker.GetClients(
		dCtx.ctx, dSrcRepo, "", dCtx.logger,
	)
	if err != nil {
		return fmt.Errorf("error creating the github repo: %w", err)
	}
	// If the caller doesn't specify the checks to run, run all checks and return all the clients.
	if dCtx.checkNamesToRun == nil || len(dCtx.checkNamesToRun) == 0 {
		dCtx.ghRepoClient, dCtx.ossFuzzClient, dCtx.ciiClient, dCtx.vulnsClient =
			repoClient, ossFuzzClient, ciiClient, vulnsClient
		return nil
	}
	dCtx.ghRepoClient = githubrepo.CreateGithubRepoClient(dCtx.ctx, dCtx.logger)
	for _, cn := range dCtx.checkNamesToRun {
		switch cn {
		case checks.CheckFuzzing:
			dCtx.ossFuzzClient = ossFuzzClient
		case checks.CheckCIIBestPractices:
			dCtx.ciiClient = ciiClient
		case checks.CheckVulnerabilities:
			dCtx.vulnsClient = vulnsClient
		}
	}
	return nil
}

func getScorecardCheckResults(dCtx *dependencydiffContext) error {
	// Initialize the checks to run from the caller's input.
	checksToRun, err := policy.GetEnabled(nil, dCtx.checkNamesToRun, nil)
	if err != nil {
		return fmt.Errorf("error init scorecard checks: %w", err)
	}
	for _, d := range dCtx.dependencydiffs {
		depCheckResult := pkg.DependencyCheckResult{
			PackageURL:       d.PackageURL,
			SourceRepository: d.SourceRepository,
			ChangeType:       d.ChangeType,
			ManifestPath:     d.ManifestPath,
			Ecosystem:        d.Ecosystem,
			Version:          d.Version,
			Name:             d.Name,
		}
		// For now we skip those without source repo urls.
		// TODO (#2063): use the BigQuery dataset to supplement null source repo URLs to fetch the Scorecard results for them.
		if d.SourceRepository != nil && *d.SourceRepository != "" {
			if d.ChangeType != nil && (dCtx.changeTypesToCheck[*d.ChangeType] || dCtx.changeTypesToCheck == nil) {
				// Initialize the repo and client(s) corresponding to the checks to run.
				ghRepo, err := githubrepo.MakeGithubRepo(*d.SourceRepository)
				err = initClientByChecks(dCtx, *d.SourceRepository)
				if err != nil {
					return fmt.Errorf("error getting github repo: %w", err)
				}
				err = initClientByChecks(dCtx, *d.SourceRepository)
				if err != nil {
					return fmt.Errorf("error initializing clients: %w", err)
				}
				// Run scorecard on those types of dependencies that the caller would like to check.
				// If the input map changeTypesToCheck is empty, by default, we run checks for all valid types.
				// TODO (#2064): use the Scorecare REST API to retrieve the Scorecard result statelessly.
				scorecardResult, err := pkg.RunScorecards(
					dCtx.ctx,
					ghRepo,
					// TODO (#2065): In future versions, ideally, this should be
					// the commitSHA corresponding to d.Version instead of HEAD.
					clients.HeadSHA,
					checksToRun,
					dCtx.ghRepoClient,
					dCtx.ossFuzzClient,
					dCtx.ciiClient,
					dCtx.vulnsClient,
				)
				// If the run fails, we leave the current dependency scorecard result empty and record the error
				// rather than letting the entire API return nil since we still expect results for other dependencies.
				if err != nil {
					depCheckResult.ScorecardResultsWithError.Error = sce.WithMessage(sce.ErrScorecardInternal,
						fmt.Sprintf("error running the scorecard checks: %v", err))
				} else { // Otherwise, we record the scorecard check results for this dependency.
					depCheckResult.ScorecardResultsWithError.ScorecardResults = &scorecardResult
				}
			}
		}
		dCtx.results = append(dCtx.results, depCheckResult)
	}
	return nil
}
