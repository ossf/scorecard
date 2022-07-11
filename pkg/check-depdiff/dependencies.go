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

import "github.com/ossf/scorecard/v4/pkg"

// ChangeType is the change type (added, updated, removed) of a dependency.
type ChangeType string

const (
	// Added suggests the dependency is a new one.
	Added ChangeType = "added"
	// Updated suggests the dependency is bumped from an old version.
	Updated ChangeType = "updated"
	// Removed suggests the dependency is removed.
	Removed ChangeType = "removed"
)

// IsValid determines if a ChangeType is valid.
func (ct *ChangeType) IsValid() bool {
	switch *ct {
	case Added, Updated, Removed:
		return true
	default:
		return false
	}
}

// rawDependency is the Dependency structure that is used to receive
// the raw results from the GitHub Dependency Review API.
type rawDependency struct {
	// Package URL is a short link for a package.
	PackageURL *string `json:"package_url"`

	// SrcRepoURL is the source repository URL of the dependency.
	SrcRepoURL *string `json:"source_repository_url"`

	// ChangeType indicates whether the dependency is added, updated, or removed.
	ChangeType *ChangeType `json:"change_type"`

	// ManifestFileName is the name of the manifest file of the dependency, such as go.mod for Go.
	ManifestPath *string `json:"manifest"`

	// Ecosystem is the name of the package management system, such as NPM, GO, PYPI.
	Ecosystem *string `json:"ecosystem"`

	// Name is the name of the dependency.
	Name string `json:"name"`

	// Version is the package version of the dependency.
	Version *string `json:"version"`
}

// DependencyCheckResult is the dependency structure used in the returned results.
type DependencyCheckResult struct {
	// Package URL is a short link for a package.
	PackageURL *string `json:"packageUrl"`

	// SrcRepoURL is the source repository URL of the dependency.
	SrcRepoURL *string `json:"srcRepoUrl"`

	// ChangeType indicates whether the dependency is added, updated, or removed.
	ChangeType *ChangeType `json:"changeType"`

	// ManifestFileName is the name of the manifest file of the dependency, such as go.mod for Go.
	ManifestPath *string `json:"manifest"`

	// Ecosystem is the name of the package management system, such as NPM, GO, PYPI.
	Ecosystem *string `json:"ecosystem"`

	// Name is the name of the dependency.
	Name string `json:"name"`

	// Version is the package version of the dependency.
	Version *string `json:"version"`

	// ScReresults is the scorecard result for the dependency repo.
	ScReresults *pkg.ScorecardResult `json:"scorecardResults"`
}
