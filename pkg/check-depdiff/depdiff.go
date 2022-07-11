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

package depdiff

import (
	"fmt"
	"net/http"
	"path"
	"time"

	gogh "github.com/google/go-github/v38/github"
)

type DepDiffContext struct {
	OwnerName   string
	RepoName    string
	BaseSHA     string
	HeadSHA     string
	AccessToken string
}

func GetDependencyDiff(ownerName, repoName, baseSHA, headSHA, accessToken string) (string, error) {
	ctx := DepDiffContext{
		OwnerName:   ownerName,
		RepoName:    repoName,
		BaseSHA:     baseSHA,
		HeadSHA:     headSHA,
		AccessToken: accessToken,
	}

	// Fetch dependency diffs using the GitHub Dependency Review API.
	deps, err := FetchDependencyDiffData(ctx)
	if err != nil {
		return "", err
	}
	fmt.Println(deps)

	return "", nil
}

// Get the depednency-diffs between two specified code commits.
func FetchDependencyDiffData(ctx DepDiffContext) ([]Dependency, error) {
	// Currently, the GitHub Dependency Review
	// (https://docs.github.com/en/rest/dependency-graph/dependency-review) API is used.
	// Set a ten-seconds timeout to make sure the client can be created correctly.
	client := gogh.NewClient(&http.Client{Timeout: 10 * time.Second})
	reqURL := path.Join(
		"repos", ctx.OwnerName, ctx.RepoName, "dependency-graph", "compare",
		ctx.BaseSHA+"..."+ctx.HeadSHA,
	)
	req, err := client.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("request for dependency-diff failed with %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	// An access token is required in the request header to be able to use this API.
	req.Header.Set("Authorization", "token "+ctx.AccessToken)

	depDiff := []Dependency{}
	_, err = client.Do(req.Context(), req, &depDiff)
	if err != nil {
		return nil, fmt.Errorf("get response error: %w", err)
	}
	return depDiff, nil
}

func GetAggregateScore(d Dependency) (float32, error) {
	return 0, nil
}
