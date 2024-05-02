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
	"regexp"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/finding"
)

var reRootFile = regexp.MustCompile(`^[^.]([^//]*)$`)

// SBOM retrieves the raw data for the SBOM check.
func SBOM(c *checker.CheckRequest) (checker.SBOMData, error) {
	var results checker.SBOMData

	SBOMsFound, lerr := c.RepoClient.ListSBOMs()
	if lerr != nil {
		return results, fmt.Errorf("RepoClient.ListSBOMs: %w", lerr)
	}

	for i := range SBOMsFound {
		v := SBOMsFound[i]

		results.SBOMFiles = append(results.SBOMFiles,
			checker.SBOM{
				File: checker.File{
					Path: v.Path,
					Type: finding.FileTypeURL,
				},
				Name:          v.Name,
				Schema:        v.Schema,
				SchemaVersion: v.SchemaVersion,
				URL:           v.URL,
			})
	}

	releases, lerr := c.RepoClient.ListReleases()
	if lerr != nil {
		return results, fmt.Errorf("RepoClient.ListReleases: %w", lerr)
	}

	releaseSBOMs := checkSBOMReleases(releases)
	if releaseSBOMs != nil {
		results.SBOMFiles = append(results.SBOMFiles, releaseSBOMs...)
	}

	// no SBOMs found in release artifacts or pipelines, continue looking for files
	repoFiles, err := c.RepoClient.ListFiles(func(string) (bool, error) { return true, nil })
	if err != nil {
		return results, fmt.Errorf("error during ListFiles: %w", err)
	}

	// TODO: Make these two happy path left
	sourceSBOMs := checkSBOMSource(repoFiles)
	if sourceSBOMs != nil {
		results.SBOMFiles = append(results.SBOMFiles, sourceSBOMs...)
	}

	return results, nil
}

func checkSBOMReleases(releases []clients.Release) []checker.SBOM {
	var foundSBOMs []checker.SBOM

	for i := range releases {
		v := releases[i]

		for _, link := range v.Assets {
			if !clients.ReSBOMFile.Match([]byte(link.Name)) {
				continue
			}

			foundSBOMs = append(foundSBOMs,
				checker.SBOM{
					File: checker.File{
						Path: link.URL,
						Type: finding.FileTypeURL,
					},
					Name: link.Name,
					URL:  link.URL,
				})
		}
	}
	return foundSBOMs
}

func checkSBOMSource(fileList []string) []checker.SBOM {
	var foundSBOMs []checker.SBOM

	for _, file := range fileList {
		if clients.ReSBOMFile.MatchString(file) && reRootFile.MatchString(file) {
			// TODO: parse matching file contents to determine schema & version
			foundSBOMs = append(foundSBOMs,
				checker.SBOM{
					File: checker.File{
						Path: file,
						Type: finding.FileTypeSource,
					},
					Name: file,
				})
		}
	}

	return foundSBOMs
}
