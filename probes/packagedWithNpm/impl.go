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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/probes"
)

func init() {
	probes.MustRegisterIndependent(Probe, Run)
}

//go:embed *.yml
var fs embed.FS

const Probe = "packagedWithNpm"

type packageJSON struct {
	Name       string            `json:"name"`
	Repository map[string]string `json:"repository"`
}

type npmRegistryResponse struct {
	Name       string            `json:"name"`
	Repository map[string]string `json:"repository"`
}

func Run(c *checker.CheckRequest) ([]finding.Finding, string, error) {
	if c == nil {
		return nil, "", fmt.Errorf("nil check request")
	}

	var findings []finding.Finding

	// Check if package.json exists in the repository root
	matchedFiles, err := c.RepoClient.ListFiles(func(path string) (bool, error) {
		return path == "package.json", nil
	})
	if err != nil {
		return nil, Probe, fmt.Errorf("failed to list files: %w", err)
	}

	if len(matchedFiles) == 0 {
		// No package.json found
		f, err := finding.NewWith(fs, Probe,
			"No package.json file found. Project does not appear to be an npm package.", nil,
			finding.OutcomeFalse)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}

		return []finding.Finding{*f}, Probe, nil
	}

	// Read package.json file
	reader, err := c.RepoClient.GetFileReader(matchedFiles[0])
	if err != nil {
		return nil, Probe, fmt.Errorf("failed to read package.json: %w", err)
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, Probe, fmt.Errorf("failed to read package.json content: %w", err)
	}

	var pkg packageJSON
	if err := json.Unmarshal(content, &pkg); err != nil {
		f, err := finding.NewWith(fs, Probe,
			"Found package.json but failed to parse it. Invalid JSON format.", nil,
			finding.OutcomeFalse)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		loc := &finding.Location{
			Path: matchedFiles[0],
			Type: finding.FileTypeSource,
		}
		f = f.WithLocation(loc)
		return []finding.Finding{*f}, Probe, nil
	}

	if pkg.Name == "" {
		f, err := finding.NewWith(fs, Probe,
			"Found package.json but no package name specified.", nil,
			finding.OutcomeFalse)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		loc := &finding.Location{
			Path: matchedFiles[0],
			Type: finding.FileTypeSource,
		}
		f = f.WithLocation(loc)
		return []finding.Finding{*f}, Probe, nil
	}

	// Check if package exists on npm registry
	exists, repoURL, err := checkNpmPackageExists(pkg.Name)
	if err != nil {
		f, err := finding.NewWith(fs, Probe,
			fmt.Sprintf("Found package.json with name '%s' but failed to check npm registry: %v", pkg.Name, err), nil,
			finding.OutcomeFalse)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		loc := &finding.Location{
			Path: matchedFiles[0],
			Type: finding.FileTypeSource,
		}
		f = f.WithLocation(loc)
		return []finding.Finding{*f}, Probe, nil
	}

	if !exists {
		f, err := finding.NewWith(fs, Probe,
			fmt.Sprintf("Package '%s' not found on npm registry. Project is not published to npm.", pkg.Name), nil,
			finding.OutcomeFalse)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		loc := &finding.Location{
			Path: matchedFiles[0],
			Type: finding.FileTypeSource,
		}
		f = f.WithLocation(loc)
		return []finding.Finding{*f}, Probe, nil
	}

	// Package exists on npm
	message := fmt.Sprintf("Package '%s' is published on npm registry.", pkg.Name)
	if repoURL != "" {
		message += fmt.Sprintf(" Repository URL: %s", repoURL)
	}

	f, err := finding.NewWith(fs, Probe, message, nil, finding.OutcomeTrue)
	if err != nil {
		return nil, Probe, fmt.Errorf("create finding: %w", err)
	}

	loc := &finding.Location{
		Path: matchedFiles[0],
		Type: finding.FileTypeSource,
	}
	f = f.WithLocation(loc)
	findings = append(findings, *f)

	return findings, Probe, nil
}

func checkNpmPackageExists(packageName string) (bool, string, error) {
	url := fmt.Sprintf("https://registry.npmjs.org/%s", packageName)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return false, "", fmt.Errorf("failed to query npm registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, "", nil
	}

	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("npm registry returned status: %d", resp.StatusCode)
	}

	var npmResp npmRegistryResponse
	if err := json.NewDecoder(resp.Body).Decode(&npmResp); err != nil {
		return false, "", fmt.Errorf("failed to parse npm registry response: %w", err)
	}

	repoURL := ""
	if npmResp.Repository != nil {
		if url, ok := npmResp.Repository["url"]; ok {
			// Clean up git+ prefix and .git suffix
			repoURL = strings.TrimPrefix(url, "git+")
			repoURL = strings.TrimSuffix(repoURL, ".git")
		}
	}

	return true, repoURL, nil
}
