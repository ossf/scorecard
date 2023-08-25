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

// Package cmd implements Scorecard commandline.
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	ngt "github.com/ossf/scorecard/v4/cmd/internal/nuget"
	pmc "github.com/ossf/scorecard/v4/cmd/internal/packagemanager"
	sce "github.com/ossf/scorecard/v4/errors"
)

var (
	githubDomainRegexp    = regexp.MustCompile(`^https?://github[.]com/([^/]+)/([^/]+)`)
	githubSubdomainRegexp = regexp.MustCompile(`^https?://([^.]+)[.]github[.]io/([^/]+).*`)
	gitlabDomainRegexp    = regexp.MustCompile(`^https?://gitlab[.]com/([^/]+)/([^/]+)`)
)

func makeGithubRepo(urlAndPathParts []string) string {
	if len(urlAndPathParts) < 3 {
		return ""
	}
	userOrOrg := strings.ToLower(urlAndPathParts[1])
	repoName := strings.TrimSuffix(strings.ToLower(urlAndPathParts[2]), ".git")
	if userOrOrg == "sponsors" {
		return ""
	}
	return fmt.Sprintf("https://github.com/%s/%s", userOrOrg, repoName)
}

// Both GitHub and GitLab are case insensitive (and thus we lowercase those URLS)
// however generic URLs are indeed case sensitive!
var pypiMatchers = []func(string) string{
	func(url string) string {
		return makeGithubRepo(githubDomainRegexp.FindStringSubmatch(url))
	},

	func(url string) string {
		return makeGithubRepo(githubSubdomainRegexp.FindStringSubmatch(url))
	},

	func(url string) string {
		match := gitlabDomainRegexp.FindStringSubmatch(url)
		if len(match) >= 3 {
			return strings.ToLower(fmt.Sprintf("https://gitlab.com/%s/%s", match[1], match[2]))
		}
		return ""
	},
}

type packageMangerResponse struct {
	associatedRepo string
	exists         bool
}

func fetchGitRepositoryFromPackageManagers(npm, pypi, rubygems, nuget string,
	manager pmc.Client,
) (packageMangerResponse, error) {
	if npm != "" {
		gitRepo, err := fetchGitRepositoryFromNPM(npm, manager)
		return packageMangerResponse{
			exists:         true,
			associatedRepo: gitRepo,
		}, err
	}
	if pypi != "" {
		gitRepo, err := fetchGitRepositoryFromPYPI(pypi, manager)
		return packageMangerResponse{
			exists:         true,
			associatedRepo: gitRepo,
		}, err
	}
	if rubygems != "" {
		gitRepo, err := fetchGitRepositoryFromRubyGems(rubygems, manager)
		return packageMangerResponse{
			exists:         true,
			associatedRepo: gitRepo,
		}, err
	}
	if nuget != "" {
		nugetClient := ngt.NugetClient{Manager: manager}
		gitRepo, err := fetchGitRepositoryFromNuget(nuget, nugetClient)
		return packageMangerResponse{
			exists:         true,
			associatedRepo: gitRepo,
		}, err
	}

	return packageMangerResponse{}, nil
}

type npmSearchResults struct {
	Objects []struct {
		Package struct {
			Links struct {
				Repository string `json:"repository"`
			} `json:"links"`
		} `json:"package"`
	} `json:"objects"`
}

type pypiSearchResults struct {
	Info struct {
		ProjectURLs map[string]string `json:"project_urls"`
		ProjectURL  string            `json:"project_url"`
	} `json:"info"`
}

type rubyGemsSearchResults struct {
	SourceCodeURI string `json:"source_code_uri"`
}

// Gets the GitHub repository URL for the npm package.
func fetchGitRepositoryFromNPM(packageName string, packageManager pmc.Client) (string, error) {
	npmSearchURL := "https://registry.npmjs.org/-/v1/search?text=%s&size=1"
	resp, err := packageManager.Get(npmSearchURL, packageName)
	if err != nil {
		return "", sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("failed to get npm package json: %v", err))
	}

	defer resp.Body.Close()
	v := &npmSearchResults{}
	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return "", sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("failed to parse npm package json: %v", err))
	}
	if len(v.Objects) == 0 {
		return "", sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("could not find source repo for npm package: %s", packageName))
	}
	return v.Objects[0].Package.Links.Repository, nil
}

func findGitRepositoryInPYPIResponse(packageName string, response io.Reader) (string, error) {
	v := &pypiSearchResults{}
	err := json.NewDecoder(response).Decode(v)
	if err != nil {
		return "", sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("failed to parse pypi package json: %v", err))
	}

	v.Info.ProjectURLs["key_not_used_and_very_unlikely_to_be_present_already"] = v.Info.ProjectURL
	var validURL string
	for _, url := range v.Info.ProjectURLs {
		for _, matcher := range pypiMatchers {
			repo := matcher(url)
			if repo == "" {
				continue
			}
			if validURL == "" {
				validURL = repo
			} else if validURL != repo {
				return "", sce.WithMessage(sce.ErrScorecardInternal,
					fmt.Sprintf("found too many possible source repos for pypi package: %s", packageName))
			}
		}
	}

	if validURL == "" {
		return "", sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("could not find source repo for pypi package: %s", packageName))
	} else {
		return validURL, nil
	}
}

// Gets the GitHub repository URL for the pypi package.
func fetchGitRepositoryFromPYPI(packageName string, manager pmc.Client) (string, error) {
	pypiSearchURL := "https://pypi.org/pypi/%s/json"
	resp, err := manager.Get(pypiSearchURL, packageName)
	if err != nil {
		return "", sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("failed to get pypi package json: %v", err))
	}

	defer resp.Body.Close()
	return findGitRepositoryInPYPIResponse(packageName, resp.Body)
}

// Gets the GitHub repository URL for the rubygems package.
func fetchGitRepositoryFromRubyGems(packageName string, manager pmc.Client) (string, error) {
	rubyGemsSearchURL := "https://rubygems.org/api/v1/gems/%s.json"
	resp, err := manager.Get(rubyGemsSearchURL, packageName)
	if err != nil {
		return "", sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("failed to get ruby gem json: %v", err))
	}

	defer resp.Body.Close()
	v := &rubyGemsSearchResults{}
	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return "", sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("failed to parse ruby gem json: %v", err))
	}
	if v.SourceCodeURI == "" {
		return "", sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("could not find source repo for ruby gem: %v", err))
	}
	return v.SourceCodeURI, nil
}

// Gets the GitHub repository URL for the nuget package.
func fetchGitRepositoryFromNuget(packageName string, nugetClient ngt.Client) (string, error) {
	repositoryURI, err := nugetClient.GitRepositoryByPackageName(packageName)
	if err != nil {
		return "", sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("could not find source repo for nuget package: %v", err))
	}
	return repositoryURI, nil
}
