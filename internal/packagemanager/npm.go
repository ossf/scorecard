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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ossf/scorecard/v5/checker"
)

var errMissingPackageName = errors.New("package.json does not contain a name field")

// NPMRegistry provides npm registry functionality.
type NPMRegistry struct {
	client *http.Client
}

// NewNPMRegistry creates a new NPM registry checker.
func NewNPMRegistry() *NPMRegistry {
	return &NPMRegistry{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type npmPackageJSON struct {
	Repository map[string]string `json:"repository"`
	Name       string            `json:"name"`
}

type npmRegistryResponse struct {
	Repository map[string]string `json:"repository"`
	Name       string            `json:"name"`
}

// CheckPackageExists checks if a package exists in the npm registry.
func (n *NPMRegistry) CheckPackageExists(ctx context.Context, packageName string) (*checker.PackageRegistry, error) {
	url := fmt.Sprintf("https://registry.npmjs.org/%s", packageName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		errMsg := fmt.Sprintf("failed to create request: %v", err)
		return &checker.PackageRegistry{
			Type:      "npm",
			Published: false,
			Error:     &errMsg,
		}, nil
	}

	resp, err := n.client.Do(req)
	if err != nil {
		errMsg := fmt.Sprintf("failed to query npm registry: %v", err)
		return &checker.PackageRegistry{
			Type:      "npm",
			Published: false,
			Error:     &errMsg,
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &checker.PackageRegistry{
			Type:      "npm",
			Published: false,
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("npm registry returned status: %d", resp.StatusCode)
		return &checker.PackageRegistry{
			Type:      "npm",
			Published: false,
			Error:     &errMsg,
		}, nil
	}

	var npmResp npmRegistryResponse
	if err := json.NewDecoder(resp.Body).Decode(&npmResp); err != nil {
		errMsg := fmt.Sprintf("failed to parse npm registry response: %v", err)
		return &checker.PackageRegistry{
			Type:      "npm",
			Published: false,
			Error:     &errMsg,
		}, nil
	}

	var repoURL *string
	if npmResp.Repository != nil {
		if url, ok := npmResp.Repository["url"]; ok {
			// Clean up git+ prefix and .git suffix
			cleanURL := strings.TrimPrefix(url, "git+")
			cleanURL = strings.TrimSuffix(cleanURL, ".git")
			repoURL = &cleanURL
		}
	}

	return &checker.PackageRegistry{
		Type:          "npm",
		Published:     true,
		RepositoryURL: repoURL,
	}, nil
}

// ParsePackageJSON parses a package.json file and extracts the package name.
func ParsePackageJSON(content []byte) (string, error) {
	var pkg npmPackageJSON
	if err := json.Unmarshal(content, &pkg); err != nil {
		return "", fmt.Errorf("failed to parse package.json: %w", err)
	}

	if pkg.Name == "" {
		return "", errMissingPackageName
	}

	return pkg.Name, nil
}
