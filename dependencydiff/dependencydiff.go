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

func GetDependencyDiff(owner, repo, base, head string) ([]pkg.DependencyCheckResult, error) {
	// Fetch dependency diffs using the GitHub Dependency Review API.
	deps, err := FetchDependencyDiffData(owner, repo, base, head)
	if err != nil {
		return nil, err
	}
	// PrintDependencies(deps)

	ctx := context.Background()
	ghRepo, err := githubrepo.MakeGithubRepo(path.Join(owner, repo))
	if err != nil {
		return nil, err
	}
	ghRepoClient := githubrepo.CreateGithubRepoClient(ctx, nil)
	ossFuzzRepoClient, err := githubrepo.CreateOssFuzzRepoClient(ctx, nil)
	vulnsClient := clients.DefaultVulnerabilitiesClient()
	if err != nil {
		panic(err)
	}
	results := []pkg.DependencyCheckResult{}
	count := 0
	for _, d := range deps {
		if d.SourceRepository != nil {
			// Running scorecard on this dependency and fetch the result
			fmt.Printf("Running Scorecard checks for %s\n", d.Name)
			scResult, err := pkg.RunScorecards(
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
				return nil, err
			}
			dd := pkg.DependencyCheckResult{
				PackageURL:       d.PackageURL,
				SourceRepository: d.SourceRepository,
				ChangeType:       d.ChangeType,
				ManifestPath:     d.ManifestPath,
				Ecosystem:        d.Ecosystem,
				Version:          d.Version,
				ScorecardResults: &scResult,
				Name:             d.Name,
			}
			results = append(results, dd)
			count += 1
			if count == 3 {
				break
			}
		} else {
			fmt.Printf("Skipping %s, no source repo found\n", d.Name)
		}
		// Skip those without source repo urls.
		// TODO: use BigQuery to supplement null source repo URLs and fetch the Scorecard results for them.
	}
	return results, nil
}

// Get the depednency-diffs between two specified code commits.
func FetchDependencyDiffData(owner, repo, base, head string) ([]Dependency, error) {
	reqURL, err := url.Parse(
		path.Join(
			"api.github.com",
			"repos", owner, repo, "dependency-graph", "compare", base+"..."+head,
		),
	)
	if err != nil {
		return nil, err
	}
	reqURL.Scheme = "https"
	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("request for dependency-diff failed with %w", err)
	}
	ctx := context.Background()
	ghrt := roundtripper.NewTransport(ctx, nil)
	resp, err := ghrt.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	depDiff := []Dependency{}
	err = json.NewDecoder(resp.Body).Decode(&depDiff)
	if err != nil {
		return nil, fmt.Errorf("parse response error: %w", err)
	}
	return depDiff, nil
}
