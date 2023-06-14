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

	ngt "github.com/ossf/scorecard/v4/clients/nuget"
	pmc "github.com/ossf/scorecard/v4/clients/packagemanager"
	sce "github.com/ossf/scorecard/v4/errors"
)

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
		ProjectUrls struct {
			Source string `json:"Source"`
		} `json:"project_urls"`
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

// Gets the GitHub repository URL for the pypi package.
func fetchGitRepositoryFromPYPI(packageName string, manager pmc.Client) (string, error) {
	pypiSearchURL := "https://pypi.org/pypi/%s/json"
	resp, err := manager.Get(pypiSearchURL, packageName)
	if err != nil {
		return "", sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("failed to get pypi package json: %v", err))
	}

	defer resp.Body.Close()
	v := &pypiSearchResults{}
	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return "", sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("failed to parse pypi package json: %v", err))
	}
	if v.Info.ProjectUrls.Source == "" {
		return "", sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("could not find source repo for pypi package: %s", packageName))
	}
	return v.Info.ProjectUrls.Source, nil
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
