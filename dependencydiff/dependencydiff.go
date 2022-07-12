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

// GetDependencyDiffResults gets dependency changes between two given code commits BASE and HEAD
// along with the Scorecard check results of the dependencies, and returns a slice of DependencyCheckResult.
func GetDependencyDiffResults(ownerName, repoName, baseSHA, headSHA string) ([]pkg.DependencyCheckResult, error) {
	ctx := context.Background()
	// Fetch dependency diffs using the GitHub Dependency Review API.
	deps, err := FetchDependencyDiffData(ctx, ownerName, repoName, baseSHA, headSHA)
	if err != nil {
		return nil, fmt.Errorf("error fetching dependency changes: %w", err)
	}

	ghRepo, err := githubrepo.MakeGithubRepo(path.Join(ownerName, repoName))
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
		if d.SourceRepository != nil {
			if *d.SourceRepository == "" {
				continue
			}
			// Running scorecard on this dependency and fetch the result.
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
			depCheckResult := pkg.DependencyCheckResult{
				PackageURL:       d.PackageURL,
				SourceRepository: d.SourceRepository,
				ChangeType:       d.ChangeType,
				ManifestPath:     d.ManifestPath,
				Ecosystem:        d.Ecosystem,
				Version:          d.Version,
				ScorecardResults: &scorecardResult,
				Name:             d.Name,
			}
			results = append(results, depCheckResult)
		}
		// Skip those without source repo urls.
		// TODO: use the BigQuery dataset to supplement null source repo URLs
		// so that we can fetch the Scorecard results for them.
	}
	return results, nil
}

// FetchDependencyDiffData fetches the depednency-diffs between the two code commits
// using the GitHub Dependency Review API, and returns a slice of Dependency.
func FetchDependencyDiffData(ctx context.Context, owner, repo, base, head string) ([]Dependency, error) {
	reqURL, err := url.Parse("https://api.github.com")
	if err != nil {
		return nil, fmt.Errorf("error parsing the url: %w", err)
	}
	reqURL.Path = path.Join(
		"repos", owner, repo, "dependency-graph", "compare", base+"..."+head,
	)
	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("request for dependency-diff failed with %w", err)
	}
	ghrt := roundtripper.NewTransport(ctx, nil)
	resp, err := ghrt.RoundTrip(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error receiving the http reponse: %w with resp status code %v", err, resp.StatusCode)
	}
	defer resp.Body.Close()
	depDiff := []Dependency{}
	err = json.NewDecoder(resp.Body).Decode(&depDiff)
	if err != nil {
		return nil, fmt.Errorf("error parsing the http response: %w", err)
	}
	return depDiff, nil
}
