package cmd

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/exp/slices"

	sce "github.com/ossf/scorecard/v4/errors"
)

type nugetClient struct {
	manager packageManagerClient
}

type nugetIndexResult struct {
	ID   string `json:"@id"`
	Type string `json:"@type"`
}

type nugetIndexResults struct {
	Resources []nugetIndexResult `json:"resources"`
}

func (n nugetIndexResults) getFieldFromIndexResults(resultType string) (string, error) {
	resourceIndex := slices.IndexFunc(n.Resources,
		func(n nugetIndexResult) bool { return n.Type == resultType })
	if resourceIndex == -1 {
		return "", sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("failed to find %v URI at nuget index json", resultType))
	}

	return n.Resources[resourceIndex].ID, nil
}

type nugetPackageRegistrationCatalogRoot struct {
	Pages []nugetPackageRegistrationCatalogPage `json:"items"`
}

func (n nugetPackageRegistrationCatalogRoot) LatestVersion(manager packageManagerClient) (string, error) {
	for pageIndex := len(n.Pages) - 1; pageIndex >= 0; pageIndex-- {
		page := n.Pages[pageIndex]
		if page.Packages == nil {
			respPage, err := manager.GetURI(page.ID)
			if err != nil {
				return "", sce.WithMessage(sce.ErrScorecardInternal,
					fmt.Sprintf("failed to get nuget package registration page: %v", err))
			}
			defer respPage.Body.Close()
			err = json.NewDecoder(respPage.Body).Decode(&page)
			if err != nil {
				return "", sce.WithMessage(sce.ErrScorecardInternal,
					fmt.Sprintf("failed to parse nuget package registration page: %v", err))
			}
		}
		for packageIndex := len(page.Packages) - 1; packageIndex >= 0; packageIndex-- {
			base, preReleaseSuffix := getNugetSemVer(page.Packages[packageIndex].Entry.Version)
			// skipping non listed and pre-releases
			if page.Packages[packageIndex].Entry.Listed && len(strings.TrimSpace(preReleaseSuffix)) == 0 {
				return base, nil
			}
		}
	}
	return "", sce.WithMessage(sce.ErrScorecardInternal, "failed to get a listed version for package")
}

type nugetPackageRegistrationCatalogPage struct {
	ID       string                            `json:"@id"`
	Packages []nugetPackageRegistrationPackage `json:"items"`
}

type nugetPackageRegistrationPackage struct {
	Entry nugetPackageRegistrationCatalogEntry `json:"catalogEntry"`
}

type nugetPackageRegistrationCatalogEntry struct {
	Version string `json:"version"`
	Listed  bool   `json:"listed"`
}

func (e *nugetPackageRegistrationCatalogEntry) UnmarshalJSON(text []byte) error {
	type Alias nugetPackageRegistrationCatalogEntry
	aux := Alias{
		Listed: true, // set the default value before parsing JSON
	}
	if err := json.Unmarshal(text, &aux); err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("failed to unmarshal json: %v", err))
	}
	*e = nugetPackageRegistrationCatalogEntry(aux)
	return nil
}

type nugetNuspec struct {
	XMLName  xml.Name       `xml:"package"`
	Metadata nuspecMetadata `xml:"metadata"`
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

func (c *nugetClient) getPackageSpec(packageBaseURL, registrationBaseURL, packageName string) (nugetNuspec, error) {
	lowerCasePackageName := strings.ToLower(packageName)
	lastPackageVersion, err := getLatestListedVersion(registrationBaseURL,
		lowerCasePackageName, c.manager)
	if err != nil {
		return nugetNuspec{}, err
	}

	respPackageSpec, err := c.manager.Get(
		packageBaseURL+"%[1]v/"+lastPackageVersion+"/%[1]v.nuspec", lowerCasePackageName)
	if err != nil {
		return nugetNuspec{}, sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("failed to get nuget package spec json: %v", err))
	}
	defer respPackageSpec.Body.Close()

	packageSpecResults := &nugetNuspec{}
	err = xml.NewDecoder(respPackageSpec.Body).Decode(packageSpecResults)
	if err != nil {
		return nugetNuspec{}, sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("failed to parse nuget nuspec xml: %v", err))
	}
	if packageSpecResults.Metadata == (nuspecMetadata{}) {
		return nugetNuspec{}, sce.WithMessage(sce.ErrScorecardInternal,
			"Nuget nuspec xml Metadata is empty")
	}

	return *packageSpecResults, nil
}

func getProjectURLFromPackageSpec(p *nugetNuspec, packageName string) (string, error) {
	var projectURL string
	if strings.TrimSpace(p.Metadata.Repository.URL) != "" {
		projectURL = p.Metadata.Repository.URL
	}
	if strings.TrimSpace(p.Metadata.ProjectURL) != "" {
		projectURL = p.Metadata.ProjectURL
	}
	if isSupportedProjectURL(projectURL) {
		return strings.TrimSuffix(projectURL, ".git"), nil
	}
	return "", sce.WithMessage(sce.ErrScorecardInternal,
		fmt.Sprintf("source repo is not defined for nuget package %v", packageName))
}

func (c *nugetClient) fetchGitRepositoryForPackage(packageName string) (string, error) {
	nugetPackageBaseURL, nugetRegistrationBaseURL, err := c.getNugetBaseUrls()
	if err != nil {
		return "", err
	}

	packageSpec, err := c.getPackageSpec(nugetPackageBaseURL, nugetRegistrationBaseURL, packageName)
	if err != nil {
		return "", err
	}
	packageURL, err := getProjectURLFromPackageSpec(&packageSpec, packageName)
	if err != nil {
		return "", err
	}
	return packageURL, nil
}

func (c *nugetClient) getNugetBaseUrls() (string, string, error) {
	nugetIndexURL := "https://api.nuget.org/v3/index.json"
	respIndex, err := c.manager.GetURI(nugetIndexURL)
	if err != nil {
		return "", "", sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("failed to get nuget index json: %v", err))
	}
	defer respIndex.Body.Close()
	nugetIndexResults := &nugetIndexResults{}
	err = json.NewDecoder(respIndex.Body).Decode(nugetIndexResults)
	if err != nil {
		return "", "", sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("failed to parse nuget index json: %v", err))
	}
	nugetPackageBaseURL, err := nugetIndexResults.getFieldFromIndexResults("PackageBaseAddress/3.0.0")
	if err != nil {
		return "", "", err
	}
	nugetRegistrationBaseURL, err := nugetIndexResults.getFieldFromIndexResults("RegistrationsBaseUrl/3.6.0")
	if err != nil {
		return "", "", err
	}
	return nugetPackageBaseURL, nugetRegistrationBaseURL, nil
}

func isSupportedProjectURL(projectURL string) bool {
	pattern := `^(?:https?://)?(?:www\.)?(?:github|gitlab)\.com/([A-Za-z0-9_\.-]+)/([A-Za-z0-9_\./-]+)$`
	regex := regexp.MustCompile(pattern)
	return regex.MatchString(projectURL)
}

// Gets the latest listed nuget version of a package, based on the protocol defined at
// https://learn.microsoft.com/en-us/nuget/api/package-base-address-resource#enumerate-package-versions
func getLatestListedVersion(baseURL, packageName string, manager packageManagerClient) (string, error) {
	resPackageRegistrationIndex, err := manager.Get(baseURL+"%s/index.json", packageName)
	if err != nil {
		return "", sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("failed to get nuget package registration index json: %v", err))
	}
	defer resPackageRegistrationIndex.Body.Close()
	packageRegistrationCatalogRoot := &nugetPackageRegistrationCatalogRoot{}
	err = json.NewDecoder(resPackageRegistrationIndex.Body).Decode(packageRegistrationCatalogRoot)
	if err != nil {
		return "", sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("failed to parse package registration index json: %v", err))
	}
	return getLatestListedVersionFromPackageRegistrationPages(packageRegistrationCatalogRoot.Pages, manager)
}

func getLatestListedVersionFromPackageRegistrationPages(pages []nugetPackageRegistrationCatalogPage,
	manager packageManagerClient,
) (string, error) {
	for pageIndex := len(pages) - 1; pageIndex >= 0; pageIndex-- {
		page := pages[pageIndex]
		if page.Packages == nil {
			respPage, err := manager.GetURI(page.ID)
			if err != nil {
				return "", sce.WithMessage(sce.ErrScorecardInternal,
					fmt.Sprintf("failed to get nuget package registration page: %v", err))
			}
			defer respPage.Body.Close()
			err = json.NewDecoder(respPage.Body).Decode(&page)
			if err != nil {
				return "", sce.WithMessage(sce.ErrScorecardInternal,
					fmt.Sprintf("failed to parse nuget package registration page: %v", err))
			}
		}
		for packageIndex := len(page.Packages) - 1; packageIndex >= 0; packageIndex-- {
			semVerString, preRelease := getNugetSemVer(page.Packages[packageIndex].Entry.Version)
			// skipping non listed and pre-releases
			if page.Packages[packageIndex].Entry.Listed && len(strings.TrimSpace(preRelease)) == 0 {
				return semVerString, nil
			}
		}
	}
	return "", sce.WithMessage(sce.ErrScorecardInternal,
		"failed to get a listed version for package")
}

// Nuget semver diverges from Semantic Versioning.
// This method returns the Nuget represntation of version and pre release strings.
// nolint: lll // long URL
// more info: https://learn.microsoft.com/en-us/nuget/concepts/package-versioning#where-nugetversion-diverges-from-semantic-versioning
func getNugetSemVer(versionString string) (base, preReleaseSuffix string) {
	metadataAndVersion := strings.Split(versionString, "+")
	prereleaseAndVersions := strings.Split(metadataAndVersion[0], "-")
	if len(prereleaseAndVersions) == 1 {
		return prereleaseAndVersions[0], ""
	}
	return prereleaseAndVersions[0], prereleaseAndVersions[1]
}
