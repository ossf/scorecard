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

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/finding"
)

var (
	reRootFile = regexp.MustCompile(`^[^.]([^//]*)$`)
	reSBOMFile = regexp.MustCompile(
		`(?i).+\.(cdx.json|cdx.xml|spdx|spdx.json|spdx.xml|spdx.y[a?]ml|spdx.rdf|spdx.rdf.xm)`,
	)
)

// SBOM retrieves the raw data for the SBOM check.
func SBOM(c *checker.CheckRequest) (checker.SBOMData, error) {
	var results checker.SBOMData

	releases, lerr := c.RepoClient.ListReleases()
	if lerr != nil {
		return results, fmt.Errorf("RepoClient.ListReleases: %w", lerr)
	}

	results.SBOMFiles = append(results.SBOMFiles, checkSBOMReleases(releases)...)

	// Look for SBOMs in source
	repoFiles, err := c.RepoClient.ListFiles(func(string) (bool, error) { return true, nil })
	if err != nil {
		return results, fmt.Errorf("error during ListFiles: %w", err)
	}

	results.SBOMFiles = append(results.SBOMFiles, checkSBOMSource(repoFiles)...)

	return results, nil
}

func checkSBOMReleases(releases []clients.Release) []checker.SBOM {
	var foundSBOMs []checker.SBOM

	for i := range releases {
		v := releases[i]

		for _, link := range v.Assets {
			if !reSBOMFile.MatchString(link.Name) {
				continue
			}

			foundSBOMs = append(foundSBOMs,
				checker.SBOM{
					File: checker.File{
						Path: link.URL,
						Type: finding.FileTypeURL,
					},
					Name: link.Name,
				})
		}
	}
	return foundSBOMs
}

func checkSBOMSource(fileList []string) []checker.SBOM {
	var foundSBOMs []checker.SBOM

	for _, file := range fileList {
		if reSBOMFile.MatchString(file) && reRootFile.MatchString(file) {
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
