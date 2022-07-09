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

// ChangeType is the change type (added, updated, removed) of a dependency.
type ChangeType string

const (
	Added   ChangeType = "added"
	Updated ChangeType = "updated"
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

// Dependency is a dependency.
type Dependency struct {
	// ChangeType indicates whether the dependency is added, updated, or removed.
	ChangeType ChangeType `json:"change_type"`

	// ManifestFileName is the name of the manifest file of the dependency, such as go.mod for Go.
	ManifestFileName string `json:"manifest"`

	// Ecosystem is the name of the package management system, such as NPM, GO, PYPI.
	Ecosystem string `json:"ecosystem"`

	// Name is the name of the dependency.
	Name string `json:"name"`

	// Version is the package version of the dependency.
	Version string `json:"version"`

	// Package URL is a short link for a package.
	PackageURL *string `json:"package_url"`

	// SrcRepoURL is the source repository URL of the dependency.
	SrcRepoURL *string `json:"source_repository_url"`
}
