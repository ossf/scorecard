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

package pkg

import (
	"fmt"
	"os"

	"github.com/ossf/scorecard/v4/docs/checks"
	"github.com/ossf/scorecard/v4/log"
	"github.com/ossf/scorecard/v4/options"
)

// ChangeType is the change type (added, updated, removed) of a dependency.
type ChangeType string

const (
	// Added suggests the dependency is a newly added one.
	Added ChangeType = "added"
	// Updated suggests the dependency is updated from an old version.
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

// ScorecardResultsWithError is used for the dependency-diff module to record scorecard results and their errors.
type ScorecardResultsWithError struct {
	// ScorecardResults is the scorecard result for the dependency repo.
	ScorecardResults *ScorecardResult

	// Error is an error returned when running the scorecard checks. A nil Error indicates the run succeeded.
	Error error
}

// DependencyCheckResult is the dependency structure used in the returned results.
type DependencyCheckResult struct {
	// ChangeType indicates whether the dependency is added, updated, or removed.
	ChangeType *ChangeType `json:"changeType"`

	// Package URL is a short link for a package.
	PackageURL *string `json:"packageUrl"`

	// SourceRepository is the source repository URL of the dependency.
	SourceRepository *string `json:"sourceRepository"`

	// ManifestPath is the path of the manifest file of the dependency, such as go.mod for Go.
	ManifestPath *string `json:"manifestPath"`

	// Ecosystem is the name of the package management system, such as NPM, GO, PYPI.
	Ecosystem *string `json:"ecosystem"`

	// Version is the package version of the dependency.
	Version *string `json:"version"`

	// ScorecardResultsWithError is the scorecard checking results of the dependency.
	ScorecardResultsWithError ScorecardResultsWithError `json:"scorecardResultsWithError"`

	// Name is the name of the dependency.
	Name string `json:"name"`
}

// FormatDependencydiffResults formats dependencydiff results.
func FormatDependencydiffResults(
	opts *options.Options,
	depdiffResults []DependencyCheckResult,
	doc checks.Doc,
) error {
	err := DependencydiffResultsAsJSON(depdiffResults, log.ParseLevel(opts.LogLevel), doc, os.Stdout)
	if err != nil {
		return fmt.Errorf("failed to output dependencydiff results: %w", err)
	}
	return nil
}
