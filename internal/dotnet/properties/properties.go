// Copyright 2024 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package properties

import (
	"encoding/xml"
	"errors"
	"regexp"
)

var errInvalidPropsFile = errors.New("error parsing dotnet props file")

type CPMPropertyGroup struct {
	XMLName                        xml.Name `xml:"PropertyGroup"`
	ManagePackageVersionsCentrally bool     `xml:"ManagePackageVersionsCentrally"`
}

type PackageVersionItemGroup struct {
	XMLName        xml.Name         `xml:"ItemGroup"`
	PackageVersion []packageVersion `xml:"PackageVersion"`
}

type packageVersion struct {
	XMLName xml.Name `xml:"PackageVersion"`
	Version string   `xml:"Version,attr"`
	Include string   `xml:"Include,attr"`
}

type DirectoryPropsProject struct {
	XMLName        xml.Name                  `xml:"Project"`
	PropertyGroups []CPMPropertyGroup        `xml:"PropertyGroup"`
	ItemGroups     []PackageVersionItemGroup `xml:"ItemGroup"`
}

type NugetPackage struct {
	Name    string
	Version string
	IsFixed bool
}

type CentralPackageManagementConfig struct {
	PackageVersions []NugetPackage
	IsCPMEnabled    bool
}

func GetCentralPackageManagementConfig(path string, content []byte) (CentralPackageManagementConfig, error) {
	var project DirectoryPropsProject

	err := xml.Unmarshal(content, &project)
	if err != nil {
		return CentralPackageManagementConfig{}, errInvalidPropsFile
	}

	cpmConfig := CentralPackageManagementConfig{
		IsCPMEnabled: isCentralPackageManagementEnabled(&project),
	}

	if cpmConfig.IsCPMEnabled {
		cpmConfig.PackageVersions = extractNugetPackages(&project)
	}

	return cpmConfig, nil
}

func isCentralPackageManagementEnabled(project *DirectoryPropsProject) bool {
	for _, propertyGroup := range project.PropertyGroups {
		if propertyGroup.ManagePackageVersionsCentrally {
			return true
		}
	}

	return false
}

func extractNugetPackages(project *DirectoryPropsProject) []NugetPackage {
	var nugetPackages []NugetPackage
	for _, itemGroup := range project.ItemGroups {
		for _, packageVersion := range itemGroup.PackageVersion {
			nugetPackages = append(nugetPackages, NugetPackage{
				Name:    packageVersion.Include,
				Version: packageVersion.Version,
				IsFixed: isValidFixedVersion(packageVersion.Version),
			})
		}
	}
	return nugetPackages
}

// isValidFixedVersion checks if the version string is a valid, fixed version.
// more on version numbers here: https://learn.microsoft.com/en-us/nuget/concepts/package-versioning?tabs=semver20sort
// ^: Ensures the match starts at the beginning of the string.
// \[: Matches the opening square bracket [.
// [^\[,]+: Matches one or more characters that are not a comma (,) or a square bracket ([) (to avoid nested brackets).
// \]: Matches the closing square bracket ].
// $: Ensures the match ends at the end of the string.
func isValidFixedVersion(version string) bool {
	pattern := `^\[[^\[,]+\]$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(version)
}
