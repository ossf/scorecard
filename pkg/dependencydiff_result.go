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
	"io"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/ossf/scorecard/v4/docs/checks"
	"github.com/ossf/scorecard/v4/log"
	"github.com/ossf/scorecard/v4/options"

	sce "github.com/ossf/scorecard/v4/errors"
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
	var err error

	switch opts.Format {
	case options.FormatDefault:
		err = DependencydiffResultsAsString(depdiffResults, opts.ShowDetails, doc, os.Stdout)
	case options.FormatJSON:
		err = DependencydiffResultsAsJSON(depdiffResults, opts.ShowDetails, log.ParseLevel(opts.LogLevel), doc, os.Stdout)
	case options.FormatMarkdown:
		err = DependencydiffResultsAsMarkdown(depdiffResults, opts.Dependencydiff, doc, os.Stdout)
	default:
		err = sce.WithMessage(
			sce.ErrScorecardInternal,
			fmt.Sprintf(
				"invalid format flag: %v. Expected [default, json, markdown]",
				opts.Format,
			),
		)
	}
	if err != nil {
		return fmt.Errorf("failed to output dependencydiff results: %w", err)
	}
	return nil
}

// DependencydiffResultsAsString exports the dependencydiff results as a string.
func DependencydiffResultsAsString(depdiffResults []DependencyCheckResult, showDetails bool,
	doc checks.Doc, writer io.Writer,
) error {
	data := make([][]string, len(depdiffResults))

	for i, dr := range depdiffResults {
		const withdetails = 8
		const withoutdetails = 7
		var x []string

		if showDetails {
			x = make([]string, withdetails)
		} else {
			x = make([]string, withoutdetails)
		}

		if dr.ChangeType.IsValid() {
			x[0] = string(*dr.ChangeType)
		}
		x[1] = dr.Name
		if dr.Ecosystem != nil {
			x[2] = *dr.Ecosystem
		}
		if dr.SourceRepository != nil {
			x[3] = *dr.SourceRepository
		}
		if dr.ManifestPath != nil {
			x[4] = *dr.ManifestPath
		}
		if dr.PackageURL != nil {
			x[5] = *dr.PackageURL
		}
		scResults := dr.ScorecardResultsWithError.ScorecardResults
		if scResults != nil {
			score, err := scResults.GetAggregateScore(doc)
			if err != nil {
				return err
			}
			x[6] = fmt.Sprintf("%.1f", score)
			if showDetails {
				for _, c := range scResults.Checks {
					x[7] += fmt.Sprintf("%s = %d; ", c.Name, c.Score)
				}
			}
		}
		data[i] = x
	}
	table := tablewriter.NewWriter(writer)
	header := []string{
		"Change Type", "Package Name", "Ecosystem", "Source Repository", "Manifest Path",
		"Package URL", "Aggregate Score",
	}
	if showDetails {
		header = append(header, "Scorecard Check Details")
	}
	table.SetHeader(header)
	table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	table.SetRowSeparator("-")
	table.SetRowLine(true)
	table.SetCenterSeparator("|")
	table.AppendBulk(data)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetRowLine(true)
	table.Render()
	return nil
}
