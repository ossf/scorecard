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

package raw

import (
	"fmt"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/internal/packagemanager"
)

// PackageRegistries checks for packages in various registries (npm, PyPI, etc.).
// This is platform-independent - it works the same for GitHub, GitLab, or any Git host.
func PackageRegistries(c *checker.CheckRequest) (checker.PackagingData, error) {
	var data checker.PackagingData

	registryChecker := packagemanager.NewRegistryChecker()
	registryPackages, err := registryChecker.CheckAllRegistries(c)
	if err != nil {
		// Don't fail the entire check if registry queries fail
		// Just log the error and continue with empty results
		pkg := checker.Package{
			Msg: stringPointer(fmt.Sprintf("Failed to check package registries: %v", err)),
		}
		data.Packages = append(data.Packages, pkg)
		return data, nil
	}

	data.Packages = registryPackages
	return data, nil
}

func stringPointer(s string) *string {
	return &s
}
