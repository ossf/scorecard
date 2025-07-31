// Copyright 2024 OpenSSF Scorecard Authors
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

//nolint:stylecheck
package packagedWithNpm

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/checknames"
	"github.com/ossf/scorecard/v5/internal/probes"
	"github.com/ossf/scorecard/v5/probes/internal/utils/uerror"
)

func init() {
	probes.MustRegister(Probe, Run, []checknames.CheckName{checknames.Packaging})
}

//go:embed *.yml
var fs embed.FS

const Probe = "packagedWithNpm"

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	r := raw.PackagingResults
	var findings []finding.Finding

	// Look for npm registry packages in the raw data
	for _, p := range r.Packages {
		if !isNpmPackage(p) {
			continue
		}

		f, err := createFindingForPackage(p)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}

		findings = append(findings, *f)
	}

	// If no npm packages were found in the raw data, return a false finding
	if len(findings) == 0 {
		f, err := finding.NewWith(fs, Probe,
			"No npm packages detected", nil,
			finding.OutcomeFalse)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	}

	return findings, Probe, nil
}

func isNpmPackage(p checker.Package) bool {
	// Skip packages that don't have registry information or aren't npm packages
	if p.Registry == nil || p.Registry.Type != "npm" {
		return false
	}

	// Skip debug messages (packages with Msg but no registry outcome)
	if p.Msg != nil && !p.Registry.Published && p.Registry.Error == nil {
		return false
	}

	return true
}

func createFindingForPackage(p checker.Package) (*finding.Finding, error) {
	var f *finding.Finding
	var err error

	switch {
	case p.Registry.Published:
		f, err = createPublishedFinding(p)
	case p.Registry.Error != nil:
		f, err = createErrorFinding(p)
	default:
		f, err = createNotFoundFinding(p)
	}

	if err != nil {
		return nil, err
	}

	// Add location information if available
	if p.File != nil {
		loc := &finding.Location{
			Path: p.File.Path,
			Type: p.File.Type,
		}
		if p.File.Offset != checker.OffsetDefault {
			loc.LineStart = &p.File.Offset
		}
		f = f.WithLocation(loc)
	}

	return f, nil
}

func createPublishedFinding(p checker.Package) (*finding.Finding, error) {
	message := "Package is published on npm registry"
	if p.Name != nil {
		message = fmt.Sprintf("Package '%s' is published on npm registry", *p.Name)
	}
	if p.Registry.RepositoryURL != nil {
		message += fmt.Sprintf(" (repository: %s)", *p.Registry.RepositoryURL)
	}

	f, err := finding.NewWith(fs, Probe, message, nil, finding.OutcomeTrue)
	if err != nil {
		return nil, fmt.Errorf("create finding: %w", err)
	}
	return f, nil
}

func createErrorFinding(p checker.Package) (*finding.Finding, error) {
	message := fmt.Sprintf("npm registry check failed: %s", *p.Registry.Error)
	if p.Name != nil {
		message = fmt.Sprintf("Package '%s' registry check failed: %s", *p.Name, *p.Registry.Error)
	}
	f, err := finding.NewWith(fs, Probe, message, nil, finding.OutcomeFalse)
	if err != nil {
		return nil, fmt.Errorf("create finding: %w", err)
	}
	return f, nil
}

func createNotFoundFinding(p checker.Package) (*finding.Finding, error) {
	message := "Package not found on npm registry"
	if p.Name != nil {
		message = fmt.Sprintf("Package '%s' not found on npm registry", *p.Name)
	}
	f, err := finding.NewWith(fs, Probe, message, nil, finding.OutcomeFalse)
	if err != nil {
		return nil, fmt.Errorf("create finding: %w", err)
	}
	return f, nil
}
