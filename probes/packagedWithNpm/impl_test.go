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
	"testing"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
)

func Test_Run(t *testing.T) {
	t.Parallel()
	t.Run("nil raw results", func(t *testing.T) {
		t.Parallel()
		testNilRawResults(t)
	})
	t.Run("no npm packages in raw data", func(t *testing.T) {
		t.Parallel()
		testNoNpmPackages(t)
	})
	t.Run("npm package published on registry", func(t *testing.T) {
		t.Parallel()
		testNpmPackagePublished(t)
	})
	t.Run("npm package not found on registry", func(t *testing.T) {
		t.Parallel()
		testNpmPackageNotFound(t)
	})
	t.Run("npm registry check failed with error", func(t *testing.T) {
		t.Parallel()
		testNpmRegistryError(t)
	})
	t.Run("mixed registry types - only processes npm", func(t *testing.T) {
		t.Parallel()
		testMixedRegistryTypes(t)
	})
	t.Run("debug messages are skipped", func(t *testing.T) {
		t.Parallel()
		testDebugMessagesSkipped(t)
	})
}

func testNilRawResults(t *testing.T) {
	t.Helper()
	findings, s, err := Run(nil)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if s != "" {
		t.Errorf("expected empty probe name, got %q", s)
	}
	if len(findings) != 0 {
		t.Errorf("expected no findings, got %d", len(findings))
	}
}

func testNoNpmPackages(t *testing.T) {
	t.Helper()
	raw := &checker.RawResults{
		PackagingResults: checker.PackagingData{
			Packages: []checker.Package{},
		},
	}
	findings, s, err := Run(raw)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if s != Probe {
		t.Errorf("expected probe name %q, got %q", Probe, s)
	}
	if len(findings) != 1 {
		t.Errorf("expected 1 finding, got %d", len(findings))
	}
	if len(findings) > 0 && findings[0].Outcome != finding.OutcomeFalse {
		t.Errorf("expected OutcomeFalse, got %v", findings[0].Outcome)
	}
}

func testNpmPackagePublished(t *testing.T) {
	t.Helper()
	packageName := "test-package"
	repositoryURL := "https://github.com/example/test-package"
	raw := &checker.RawResults{
		PackagingResults: checker.PackagingData{
			Packages: []checker.Package{
				{
					Name: &packageName,
					File: &checker.File{
						Path:   "package.json",
						Type:   finding.FileTypeSource,
						Offset: checker.OffsetDefault,
					},
					Registry: &checker.PackageRegistry{
						Type:          "npm",
						Published:     true,
						RepositoryURL: &repositoryURL,
					},
				},
			},
		},
	}
	findings, s, err := Run(raw)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if s != Probe {
		t.Errorf("expected probe name %q, got %q", Probe, s)
	}
	if len(findings) != 1 {
		t.Errorf("expected 1 finding, got %d", len(findings))
	}
	if len(findings) > 0 && findings[0].Outcome != finding.OutcomeTrue {
		t.Errorf("expected OutcomeTrue, got %v", findings[0].Outcome)
	}
}

func testNpmPackageNotFound(t *testing.T) {
	t.Helper()
	packageName := "nonexistent-package"
	raw := &checker.RawResults{
		PackagingResults: checker.PackagingData{
			Packages: []checker.Package{
				{
					Name: &packageName,
					File: &checker.File{
						Path:   "package.json",
						Type:   finding.FileTypeSource,
						Offset: checker.OffsetDefault,
					},
					Registry: &checker.PackageRegistry{
						Type:      "npm",
						Published: false,
					},
				},
			},
		},
	}
	findings, s, err := Run(raw)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if s != Probe {
		t.Errorf("expected probe name %q, got %q", Probe, s)
	}
	if len(findings) != 1 {
		t.Errorf("expected 1 finding, got %d", len(findings))
	}
	if len(findings) > 0 && findings[0].Outcome != finding.OutcomeFalse {
		t.Errorf("expected OutcomeFalse, got %v", findings[0].Outcome)
	}
}

func testNpmRegistryError(t *testing.T) {
	t.Helper()
	packageName := "test-package"
	errorMsg := "network timeout"
	raw := &checker.RawResults{
		PackagingResults: checker.PackagingData{
			Packages: []checker.Package{
				{
					Name: &packageName,
					File: &checker.File{
						Path:   "package.json",
						Type:   finding.FileTypeSource,
						Offset: checker.OffsetDefault,
					},
					Registry: &checker.PackageRegistry{
						Type:      "npm",
						Published: false,
						Error:     &errorMsg,
					},
				},
			},
		},
	}
	findings, s, err := Run(raw)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if s != Probe {
		t.Errorf("expected probe name %q, got %q", Probe, s)
	}
	if len(findings) != 1 {
		t.Errorf("expected 1 finding, got %d", len(findings))
	}
	if len(findings) > 0 && findings[0].Outcome != finding.OutcomeFalse {
		t.Errorf("expected OutcomeFalse, got %v", findings[0].Outcome)
	}
}

func testMixedRegistryTypes(t *testing.T) {
	t.Helper()
	npmPackage := "npm-package"
	raw := &checker.RawResults{
		PackagingResults: checker.PackagingData{
			Packages: []checker.Package{
				{
					Name: &npmPackage,
					File: &checker.File{
						Path:   "package.json",
						Type:   finding.FileTypeSource,
						Offset: checker.OffsetDefault,
					},
					Registry: &checker.PackageRegistry{
						Type:      "npm",
						Published: true,
					},
				},
				{
					// This should be ignored by the npm probe
					Registry: &checker.PackageRegistry{
						Type:      "pypi",
						Published: true,
					},
				},
			},
		},
	}
	findings, s, err := Run(raw)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if s != Probe {
		t.Errorf("expected probe name %q, got %q", Probe, s)
	}
	// Should only find the npm package, not the pypi one
	if len(findings) != 1 {
		t.Errorf("expected 1 finding, got %d", len(findings))
	}
	if len(findings) > 0 && findings[0].Outcome != finding.OutcomeTrue {
		t.Errorf("expected OutcomeTrue, got %v", findings[0].Outcome)
	}
}

func testDebugMessagesSkipped(t *testing.T) {
	t.Helper()
	debugMsg := "No package.json file found"
	raw := &checker.RawResults{
		PackagingResults: checker.PackagingData{
			Packages: []checker.Package{
				{
					Msg: &debugMsg,
					Registry: &checker.PackageRegistry{
						Type:      "npm",
						Published: false,
					},
				},
			},
		},
	}
	findings, s, err := Run(raw)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if s != Probe {
		t.Errorf("expected probe name %q, got %q", Probe, s)
	}
	// Should return the default "no npm packages detected" finding
	if len(findings) != 1 {
		t.Errorf("expected 1 finding, got %d", len(findings))
	}
	if len(findings) > 0 && findings[0].Outcome != finding.OutcomeFalse {
		t.Errorf("expected OutcomeFalse, got %v", findings[0].Outcome)
	}
}
