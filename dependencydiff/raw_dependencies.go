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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/ossf/scorecard/v4/checks"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo"
	"github.com/ossf/scorecard/v4/clients/githubrepo/roundtripper"
	"github.com/ossf/scorecard/v4/pkg"
)

// Dependency is a raw dependency fetched from the GitHub Dependency Review API.
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
// using the GitHub Dependency Review API, and returns a slice of Dependency.
func fetchRawDependencyDiffData(ctx context.Context, owner, repo, base, head string) ([]pkg.DependencyCheckResult, error) {
	reqURL, err := url.Parse("https://api.github.com")
	if err != nil {
		return nil, fmt.Errorf("error parsing the url: %w", err)
	}
	reqURL.Path = url.PathEscape(path.Join(
		"repos", owner, repo, "dependency-graph", "compare", base+"..."+head,
	))
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("request for dependency-diff failed with %w", err)
	}
	ghrt := roundtripper.NewTransport(ctx, nil)
	resp, err := ghrt.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("error receiving the http reponse: %w with resp status code %v", err, resp.StatusCode)
	}
	defer resp.Body.Close()
	deps := []dependency{}
	err = json.NewDecoder(resp.Body).Decode(&deps)
	if err != nil {
		return nil, fmt.Errorf("error parsing the http response: %w", err)
	}

	ghRepo, err := githubrepo.MakeGithubRepo(path.Join(owner, repo))
	if err != nil {
		return nil, fmt.Errorf("error creating the github repo: %w", err)
	}
	ghRepoClient := githubrepo.CreateGithubRepoClient(ctx, nil)
	ossFuzzRepoClient, err := githubrepo.CreateOssFuzzRepoClient(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating the oss fuzz repo client: %w", err)
	}
	vulnsClient := clients.DefaultVulnerabilitiesClient()

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
		if d.SourceRepository != nil {
			if *d.SourceRepository != "" {
				// If the srcRepo is valid, run scorecard on this dependency and fetch the result.
				// TODO: use the Scorecare REST API to retrieve the Scorecard result statelessly.
				scorecardResult, err := pkg.RunScorecards(
					ctx,
					ghRepo,
					clients.HeadSHA,
					checks.AllChecks,
					ghRepoClient,
					ossFuzzRepoClient,
					nil,
					vulnsClient,
				)
				if err != nil {
					return nil, fmt.Errorf("error fetching the scorecard result: %w", err)
				}
				depCheckResult.ScorecardResults = &scorecardResult
			} // Append a DependencyCheckResult with an empty ScorecardResult if the srcRepo is an empty string.
		} // Same as before, if the srcRepo field is null, we also return a DependencyCheckResult with an empty ScorecardResult.
		results = append(results, depCheckResult)
	}
	return results, nil
}
