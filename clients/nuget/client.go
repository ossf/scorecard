// Copyright 2020 OpenSSF Scorecard Authors
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

// Package nuget implements Nuget API client.
package nuget

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/exp/slices"

	pmc "github.com/ossf/scorecard/v4/clients/packagemanager"
	sce "github.com/ossf/scorecard/v4/errors"
)

type indexResults struct {
	Resources []indexResult `json:"resources"`
}

func (n indexResults) findResourceByType(resultType string) (string, error) {
	resourceIndex := slices.IndexFunc(n.Resources,
		func(n indexResult) bool { return n.Type == resultType })
	if resourceIndex == -1 {
		return "", sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("failed to find %v URI at nuget index json", resultType))
	}

	return n.Resources[resourceIndex].ID, nil
}

type indexResult struct {
	ID   string `json:"@id"`
	Type string `json:"@type"`
}

type packageRegistrationCatalogRoot struct {
	Pages []packageRegistrationCatalogPage `json:"items"`
}

func (n packageRegistrationCatalogRoot) latestVersion(manager pmc.Client) (string, error) {
	for pageIndex := len(n.Pages) - 1; pageIndex >= 0; pageIndex-- {
		page := n.Pages[pageIndex]
		if page.Packages == nil {
			err := decodeResponseFromClient(func() (*http.Response, error) {
				//nolint: wrapcheck
				return manager.GetURI(page.ID)
			},
				func(rc io.ReadCloser) error {
					//nolint: wrapcheck
					return json.NewDecoder(rc).Decode(&page)
				}, "nuget package registration page")
			if err != nil {
				return "", err
			}
		}
		for packageIndex := len(page.Packages) - 1; packageIndex >= 0; packageIndex-- {
			base, preReleaseSuffix := parseNugetSemVer(page.Packages[packageIndex].Entry.Version)
			// skipping non listed and pre-releases
			if page.Packages[packageIndex].Entry.Listed && len(strings.TrimSpace(preReleaseSuffix)) == 0 {
				return base, nil
			}
		}
	}
	return "", sce.WithMessage(sce.ErrScorecardInternal, "failed to get a listed version for package")
}

type packageRegistrationCatalogPage struct {
	ID       string                           `json:"@id"`
	Packages []packageRegistrationCatalogItem `json:"items"`
}

type packageRegistrationCatalogItem struct {
	Entry packageRegistrationCatalogEntry `json:"catalogEntry"`
}

type packageRegistrationCatalogEntry struct {
	Version string `json:"version"`
	Listed  bool   `json:"listed"`
}

func (e *packageRegistrationCatalogEntry) UnmarshalJSON(text []byte) error {
	type Alias packageRegistrationCatalogEntry
	aux := Alias{
		Listed: true, // set the default value before parsing JSON
	}
	if err := json.Unmarshal(text, &aux); err != nil {
		return fmt.Errorf("failed to unmarshal json: %w", err)
	}
	*e = packageRegistrationCatalogEntry(aux)
	return nil
}

type packageNuspec struct {
	XMLName  xml.Name       `xml:"package"`
	Metadata nuspecMetadata `xml:"metadata"`
}

func (p *packageNuspec) projectURL(packageName string) (string, error) {
	for _, projectURL := range []string{p.Metadata.Repository.URL, p.Metadata.ProjectURL} {
		projectURL = strings.TrimSpace(projectURL)
		if projectURL != "" && isSupportedProjectURL(projectURL) {
			projectURL = strings.TrimSuffix(projectURL, "/")
			projectURL = strings.TrimSuffix(projectURL, ".git")
			return projectURL, nil
		}
	}
	return "", sce.WithMessage(sce.ErrScorecardInternal,
		fmt.Sprintf("source repo is not defined for nuget package %v", packageName))
}

type nuspecMetadata struct {
	XMLName    xml.Name         `xml:"metadata"`
	ProjectURL string           `xml:"projectUrl"`
	Repository nuspecRepository `xml:"repository"`
}

type nuspecRepository struct {
	XMLName xml.Name `xml:"repository"`
	URL     string   `xml:"url,attr"`
}

type Client interface {
	GitRepositoryByPackageName(packageName string) (string, error)
}

type NugetClient struct {
	Manager pmc.Client
}

func (c NugetClient) GitRepositoryByPackageName(packageName string) (string, error) {
	packageBaseURL, registrationBaseURL, err := c.baseUrls()
	if err != nil {
		return "", err
	}

	packageSpec, err := c.packageSpec(packageBaseURL, registrationBaseURL, packageName)
	if err != nil {
		return "", err
	}

	packageURL, err := packageSpec.projectURL(packageName)
	if err != nil {
		return "", err
	}
	return packageURL, nil
}

func (c *NugetClient) packageSpec(packageBaseURL, registrationBaseURL, packageName string) (packageNuspec, error) {
	lowerCasePackageName := strings.ToLower(packageName)
	lastPackageVersion, err := c.latestListedVersion(registrationBaseURL,
		lowerCasePackageName)
	if err != nil {
		return packageNuspec{}, err
	}
	packageSpecResults := &packageNuspec{}
	err = decodeResponseFromClient(func() (*http.Response, error) {
		//nolint: wrapcheck
		return c.Manager.Get(
			packageBaseURL+"%[1]v/"+lastPackageVersion+"/%[1]v.nuspec", lowerCasePackageName)
	},
		func(rc io.ReadCloser) error {
			//nolint: wrapcheck
			return xml.NewDecoder(rc).Decode(packageSpecResults)
		}, "nuget package spec")

	if err != nil {
		return packageNuspec{}, err
	}
	if packageSpecResults.Metadata == (nuspecMetadata{}) {
		return packageNuspec{}, sce.WithMessage(sce.ErrScorecardInternal,
			"Nuget nuspec xml Metadata is empty")
	}
	return *packageSpecResults, nil
}

func (c *NugetClient) baseUrls() (string, string, error) {
	indexURL := "https://api.nuget.org/v3/index.json"
	indexResults := &indexResults{}
	err := decodeResponseFromClient(func() (*http.Response, error) {
		//nolint: wrapcheck
		return c.Manager.GetURI(indexURL)
	},
		func(rc io.ReadCloser) error {
			//nolint: wrapcheck
			return json.NewDecoder(rc).Decode(indexResults)
		}, "nuget index json")
	if err != nil {
		return "", "", err
	}
	packageBaseURL, err := indexResults.findResourceByType("PackageBaseAddress/3.0.0")
	if err != nil {
		return "", "", err
	}
	registrationBaseURL, err := indexResults.findResourceByType("RegistrationsBaseUrl/3.6.0")
	if err != nil {
		return "", "", err
	}
	return packageBaseURL, registrationBaseURL, nil
}

// Gets the latest listed nuget version of a package, based on the protocol defined at
// https://learn.microsoft.com/en-us/nuget/api/package-base-address-resource#enumerate-package-versions
func (c *NugetClient) latestListedVersion(baseURL, packageName string) (string, error) {
	packageRegistrationCatalogRoot := &packageRegistrationCatalogRoot{}
	err := decodeResponseFromClient(func() (*http.Response, error) {
		//nolint: wrapcheck
		return c.Manager.Get(baseURL+"%s/index.json", packageName)
	},
		func(rc io.ReadCloser) error {
			//nolint: wrapcheck
			return json.NewDecoder(rc).Decode(packageRegistrationCatalogRoot)
		}, "nuget package registration index json")
	if err != nil {
		return "", err
	}
	return packageRegistrationCatalogRoot.latestVersion(c.Manager)
}

func isSupportedProjectURL(projectURL string) bool {
	pattern := `^(?:https?://)?(?:www\.)?(?:github|gitlab)\.com/([A-Za-z0-9_\.-]+)/([A-Za-z0-9_\./-]+)$`
	regex := regexp.MustCompile(pattern)
	return regex.MatchString(projectURL)
}

// Nuget semver diverges from Semantic Versioning.
// This method returns the Nuget represntation of version and pre release strings.
// nolint: lll // long URL
// more info: https://learn.microsoft.com/en-us/nuget/concepts/package-versioning#where-nugetversion-diverges-from-semantic-versioning
func parseNugetSemVer(versionString string) (base, preReleaseSuffix string) {
	metadataAndVersion := strings.Split(versionString, "+")
	prereleaseAndVersions := strings.Split(metadataAndVersion[0], "-")
	if len(prereleaseAndVersions) == 1 {
		return prereleaseAndVersions[0], ""
	}
	return prereleaseAndVersions[0], prereleaseAndVersions[1]
}

func decodeResponseFromClient(getFunc func() (*http.Response, error),
	decodeFunc func(io.ReadCloser) error, name string,
) error {
	response, err := getFunc()
	if err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("failed to get %s: %v", name, err))
	}
	if response.StatusCode != http.StatusOK {
		return sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("failed to get %s with status: %v", name, response.Status))
	}
	defer response.Body.Close()

	err = decodeFunc(response.Body)
	if err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("failed to parse %s: %v", name, err))
	}
	return nil
}
