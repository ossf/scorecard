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

package packagemanager

import (
	"fmt"
	"io"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
)

// RegistryChecker provides platform-independent package registry checking.
type RegistryChecker struct {
	npmRegistry *NPMRegistry
	// Future: pypiRegistry *PyPIRegistry
	// Future: cargoRegistry *CargoRegistry
}

// NewRegistryChecker creates a new registry checker.
func NewRegistryChecker() *RegistryChecker {
	return &RegistryChecker{
		npmRegistry: NewNPMRegistry(),
	}
}

// CheckAllRegistries checks all supported package registries for the given repository.
func (r *RegistryChecker) CheckAllRegistries(c *checker.CheckRequest) ([]checker.Package, error) {
	var packages []checker.Package

	// Check npm registry
	npmPackages, err := r.checkNPMRegistry(c)
	if err != nil {
		return nil, fmt.Errorf("npm registry check failed: %w", err)
	}
	packages = append(packages, npmPackages...)

	// Future: Add other package managers here
	// pypiPackages, err := r.checkPyPIRegistry(c)
	// if err != nil {
	//     return nil, fmt.Errorf("PyPI registry check failed: %w", err)
	// }
	// packages = append(packages, pypiPackages...)

	return packages, nil
}

// checkNPMRegistry checks for package.json and queries npm registry.
func (r *RegistryChecker) checkNPMRegistry(c *checker.CheckRequest) ([]checker.Package, error) {
	var packages []checker.Package

	// Look for package.json files
	matchedFiles, err := c.RepoClient.ListFiles(func(path string) (bool, error) {
		return path == "package.json", nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	if len(matchedFiles) == 0 {
		// No package.json found - add a debug package indicating no npm package
		pkg := checker.Package{
			Msg: stringPointer("No package.json file found"),
			Registry: &checker.PackageRegistry{
				Type:      "npm",
				Published: false,
			},
		}
		packages = append(packages, pkg)
		return packages, nil
	}

	// Read and parse package.json
	reader, err := c.RepoClient.GetFileReader(matchedFiles[0])
	if err != nil {
		return nil, fmt.Errorf("failed to read package.json: %w", err)
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read package.json content: %w", err)
	}

	packageName, err := ParsePackageJSON(content)
	if err != nil {
		// Invalid package.json
		errMsg := err.Error()
		pkg := checker.Package{
			Name: nil,
			File: &checker.File{
				Path:   matchedFiles[0],
				Type:   finding.FileTypeSource,
				Offset: checker.OffsetDefault,
			},
			Msg: stringPointer(fmt.Sprintf("Found package.json but parsing failed: %v", err)),
			Registry: &checker.PackageRegistry{
				Type:      "npm",
				Published: false,
				Error:     &errMsg,
			},
		}
		packages = append(packages, pkg)
		return packages, nil
	}

	// Query npm registry
	registryInfo, err := r.npmRegistry.CheckPackageExists(c.Ctx, packageName)
	if err != nil {
		return nil, fmt.Errorf("failed to check npm registry: %w", err)
	}

	// Create package with registry information
	pkg := checker.Package{
		Name: &packageName,
		File: &checker.File{
			Path:   matchedFiles[0],
			Type:   finding.FileTypeSource,
			Offset: checker.OffsetDefault,
		},
		Registry: registryInfo,
	}

	// Add appropriate message based on registry status
	switch {
	case registryInfo.Published:
		msg := fmt.Sprintf("Package '%s' is published on npm registry", packageName)
		if registryInfo.RepositoryURL != nil {
			msg += fmt.Sprintf(" (repository: %s)", *registryInfo.RepositoryURL)
		}
		pkg.Msg = &msg
	case registryInfo.Error != nil:
		pkg.Msg = stringPointer(fmt.Sprintf("Package '%s' registry check failed: %s", packageName, *registryInfo.Error))
	default:
		pkg.Msg = stringPointer(fmt.Sprintf("Package '%s' not found on npm registry", packageName))
	}

	packages = append(packages, pkg)
	return packages, nil
}

func stringPointer(s string) *string {
	return &s
}
