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
	"slices"

	"gopkg.in/yaml.v2"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/finding"
)

var reRootFile = regexp.MustCompile(`^[^.]([^//]*)$`)

// Sbom retrieves the raw data for the Sbom check.
func Sbom(c *checker.CheckRequest) (checker.SbomData, error) {
	var results checker.SbomData

	sbomsFound, lerr := c.RepoClient.ListSboms()
	if lerr != nil {
		return results, fmt.Errorf("RepoClient.ListSboms: %w", lerr)
	}

	for i := range sbomsFound {
		v := sbomsFound[i]

		results.SbomFiles = append(results.SbomFiles,
			checker.SbomFile{
				File: checker.File{
					Path: v.Path,
					Type: finding.FileTypeURL,
				},
				SbomInformation: checker.Sbom{
					Name:          v.Name,
					Origin:        checker.SbomOriginationType(v.Origin),
					Schema:        v.Schema,
					SchemaVersion: v.SchemaVersion,
				},
			})
	}

	// no sboms found in release artifacts or pipelines, continue looking for files
	repoFiles, err := c.RepoClient.ListFiles(func(string) (bool, error) { return true, nil })
	if err != nil {
		return results, fmt.Errorf("error during ListFiles: %w", err)
	}

	// TODO: Make these two happy path left
	sourceSboms, err := checkSbomSource(repoFiles)
	if err == nil && sourceSboms != nil {
		results.SbomFiles = append(results.SbomFiles, sourceSboms...)
	}

	standardSbom, err := checkSbomStandard(c, repoFiles)
	if err == nil && standardSbom != nil {
		results.SbomFiles = append(results.SbomFiles, *standardSbom)
	}

	return results, nil
}

func checkSbomSource(fileList []string) ([]checker.SbomFile, error) {
	var foundSboms []checker.SbomFile

	for _, file := range fileList {
		if clients.ReSbomFile.MatchString(file) && reRootFile.MatchString(file) {
			// TODO: parse matching file contents to determine schema & version
			foundSboms = append(foundSboms,
				checker.SbomFile{
					File: checker.File{
						Path: file,
						Type: finding.FileTypeSource,
					},
					SbomInformation: checker.Sbom{
						Name:   file,
						Origin: checker.SbomOriginationTypeOther,
					},
				})
		}
	}

	return foundSboms, nil
}

func checkSbomStandard(c *checker.CheckRequest, fileList []string) (*checker.SbomFile, error) {
	foundSbomInfo := checker.SbomFile{}

	idx := slices.IndexFunc(fileList, func(f string) bool { return f == "SECURITY_INSIGHTS.yml" })

	if idx == -1 { // no matches found
		return nil, nil
	}

	standardsFileName := fileList[idx]

	contents, err := c.RepoClient.GetFileContent(standardsFileName)
	if err != nil {
		return nil, fmt.Errorf("error getting fileContent in checkSbomStandard: %w", err)
	}

	securityInsightsFile := clients.SecurityInsightsSchema{}

	err = yaml.Unmarshal(contents, &securityInsightsFile)
	if err != nil {
		return nil, fmt.Errorf("error parsing security insights file: %w", err)
	}

	// TODO: Check for existence of sbom struct, not ID
	if securityInsightsFile.Properties.Dependencies.Properties.Sbom.ID == "" {
		return nil, fmt.Errorf("error parsing security insights file: %w", err)
	}

	sbomInfo := securityInsightsFile.Properties.Dependencies.Properties.Sbom

	foundSbomInfo.SbomInformation.Name = sbomInfo.Items.AnyOf[0].Properties.SbomFile.Description
	foundSbomInfo.SbomInformation.Origin = checker.SbomOriginationTypeStandards
	foundSbomInfo.SbomInformation.Schema = sbomInfo.Items.AnyOf[0].Properties.SbomFormat.Description

	return &foundSbomInfo, nil
}
