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
	"net/http"
	"path"
	"time"

	"github.com/google/go-github/v38/github"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo"
	"github.com/ossf/scorecard/v4/clients/githubrepo/roundtripper"
	"github.com/ossf/scorecard/v4/log"
	"github.com/ossf/scorecard/v4/pkg"
)

// dependency is a raw dependency fetched from the GitHub Dependency Review API.
// Fields of a dependnecy correspondings to those of pkg.DependencyCheckResult.
type dependency struct {
	// Package URL is a short link for a package.
	PackageURL *string `json:"package_url"`

	// SourceRepository is the source repository URL of the dependency.
	SourceRepository *string `json:"source_repository_url"`

	// ChangeType indicates whether the dependency is added, updated, or removed.
	ChangeType *pkg.ChangeType `json:"change_type"`

	// ManifestPath is the path of the manifest file of the dependency, such as go.mod for Go.
	ManifestPath *string `json:"manifest"`

	// Ecosystem is the name of the package management system, such as NPM, GO, PYPI.
	Ecosystem *string `json:"ecosystem"`

	// Version is the package version of the dependency.
	Version *string `json:"version"`

	// Name is the name of the dependency.
	Name string `json:"name"`
}

// fetchRawDependencyDiffData fetches the dependency-diffs between the two code commits
// using the GitHub Dependency Review API, and returns a slice of DependencyCheckResult.
func fetchRawDependencyDiffData(
	ctx context.Context, owner, repo, base, head string, checkNamesToRun []string, logger *log.Logger,
) ([]pkg.DependencyCheckResult, error) {
	ghrt := roundtripper.NewTransport(ctx, logger)
	ghClient := github.NewClient(
		&http.Client{
			Transport: ghrt,
			Timeout:   10 * time.Second,
		},
	)
	req, err := ghClient.NewRequest(
		"GET",
		path.Join("repos", owner, repo, "dependency-graph", "compare", base+"..."+head),
		nil,
	)
	if err != nil {
		wrapped := fmt.Errorf("request for dependency-diff failed with %w", err)
		logger.Error(wrapped, "")
		return nil, wrapped
	}
	deps := []dependency{}
	_, err = ghClient.Do(ctx, req, &deps)
	if err != nil {
		wrapped := fmt.Errorf("error receiving the http reponse: %w", err)
		logger.Error(wrapped, "")
		return nil, wrapped
	}

	ghRepo, err := githubrepo.MakeGithubRepo(path.Join(owner, repo))
	if err != nil {
		wrapped := fmt.Errorf("error creating the github repo: %w", err)
		logger.Error(wrapped, "")
		return nil, wrapped
	}

	// Initialize the checks to run from the caller's input.
	checksToRun := checker.CheckNameToFnMap{}
	if checkNamesToRun == nil && len(checkNamesToRun) == 0 {
		// If no check names are provided, we run all the checks for the caller.
		checksToRun = checks.AllChecks
	} else {
		for _, cn := range checkNamesToRun {
			checksToRun[cn] = checks.AllChecks[cn]
		}
	}

	// Initialize the client(s) corresponding to the checks to run.
	ghRepoClient := githubrepo.CreateGithubRepoClient(ctx, logger)
	// Initialize these three clients as nil at first.
	var ossFuzzClient clients.RepoClient
	var vulnsClient clients.VulnerabilitiesClient
	var ciiClient clients.CIIBestPracticesClient
	// Create the corresponding client if the check needs to run.
	for cn := range checksToRun {
		switch cn {
		case checks.CheckFuzzing:
			ossFuzzClient, err = githubrepo.CreateOssFuzzRepoClient(ctx, logger)
			if err != nil {
				wrapped := fmt.Errorf("error initializing the oss fuzz repo client: %v", err)
				logger.Error(wrapped, "")
				return nil, wrapped
			}
		case checks.CheckVulnerabilities:
			vulnsClient = clients.DefaultVulnerabilitiesClient()
		case checks.CheckCIIBestPractices:
			ciiClient = clients.DefaultCIIBestPracticesClient()
		}
	}

	results := []pkg.DependencyCheckResult{}
	for _, d := range deps {
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
		// TODO: use the BigQuery dataset to supplement null source repo URLs
		// so that we can fetch the Scorecard results for them.
		if d.SourceRepository != nil && *d.SourceRepository != "" {
			if d.ChangeType != nil && *d.ChangeType != pkg.Removed {
				// Run scorecard on those added/updated dependencies with valid srcRepo URLs and fetch the result.
				// TODO: use the Scorecare REST API to retrieve the Scorecard result statelessly.
				scorecardResult, err := pkg.RunScorecards(
					ctx,
					ghRepo,
					// TODO: In future versions, ideally, this should be
					// the commitSHA corresponding to d.Version instead of HEAD.
					clients.HeadSHA,
					checksToRun,
					ghRepoClient,
					ossFuzzClient,
					ciiClient,
					vulnsClient,
				)
				// "err==nil" suggests the run succeeds, so that we record the scorecard check results for this dependency.
				// Otherwise, it indicates the run fails and we leave the current dependency scorecard result empty
				// rather than letting the entire API return nil since we still expect results for other dependencies.
				if err != nil {
					logger.Error(
						fmt.Errorf("error running scorecard checks: %v", err),
						fmt.Sprintf("The scorecard checks running for dependency %s failed.", d.Name),
					)
				} else {
					depCheckResult.ScorecardResults = &scorecardResult
				}
			}
		}
		results = append(results, depCheckResult)
	}
	return results, nil
}
