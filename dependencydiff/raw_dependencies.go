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

package dependencydiff

import (
	"fmt"
	"net/http"
	"path"

	"github.com/google/go-github/v38/github"

	"github.com/ossf/scorecard/v4/clients/githubrepo/roundtripper"
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
func fetchRawDependencyDiffData(dCtx *dependencydiffContext) error {
	ghrt := roundtripper.NewTransport(dCtx.ctx, dCtx.logger)
	ghClient := github.NewClient(&http.Client{Transport: ghrt})
	req, err := ghClient.NewRequest(
		"GET",
		path.Join("repos", dCtx.ownerName, dCtx.repoName,
			"dependency-graph", "compare", dCtx.base+"..."+dCtx.head),
		nil,
	)
	if err != nil {
		return fmt.Errorf("request for dependency-diff failed with %w", err)
	}
	_, err = ghClient.Do(dCtx.ctx, req, &dCtx.dependencydiffs)
	if err != nil {
		return fmt.Errorf("error parsing the dependency-diff response: %w", err)
	}
	for _, d := range dCtx.dependencydiffs {
		if !d.ChangeType.IsValid() {
			return fmt.Errorf("%w: change type", errInvalid)
		}
	}
	return nil
}
