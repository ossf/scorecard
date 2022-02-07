// Copyright 2022 Security Scorecard Authors
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
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

var (
	errInputResultFileNotSet     = errors.New("INPUT_RESULTS_FILE is not set")
	errInputResultFileEmpty      = errors.New("INPUT_RESULTS_FILE is empty")
	errInputResultFormatNotSet   = errors.New("INPUT_RESULTS_FORMAT is not set")
	errInputResultFormatEmtpy    = errors.New("INPUT_RESULTS_FORMAT is empty")
	errInputPublishResultsNotSet = errors.New("INPUT_PUBLISH_RESULTS is not set")
	errInputPublishResultsEmpty  = errors.New("INPUT_PUBLISH_RESULTS is empty")
	errRequiredENVNotSet         = errors.New("required environment variables are not set")
	errGitHubEventPath           = errors.New("error getting GITHUB_EVENT_PATH")
	errGitHubEventPathEmpty      = errors.New("GITHUB_EVENT_PATH is empty")
	errGitHubEventPathNotSet     = errors.New("GITHUB_EVENT_PATH is not set")
	errEmptyDefaultBranch        = errors.New("default branch is empty")
)

type repositoryInformation struct {
	DefaultBranch string `json:"default_branch"`
	Private       bool   `json:"private"`
}

// main is the entrypoint for the action.
func main() {
	// TODO - This is a port of the entrypoint.sh script.
	// This is still a work in progress.
	if err := initalizeENVVariables(); err != nil {
		panic(err)
	}
	if err := checkIfRequiredENVSet(); err != nil {
		panic(err)
	}

	repository := os.Getenv("GITHUB_REPOSITORY")
	token := os.Getenv("GITHUB_AUTH_TOKEN")

	repo, err := getRepositoryInformation(repository, token)
	if err != nil {
		panic(err)
	}

	if err := updateRepositoryInformation(repo.Private, repo.DefaultBranch); err != nil {
		panic(err)
	}

	if err := updateEnvVariables(); err != nil {
		panic(err)
	}
}

// initalizeENVVariables is a function to initialize the environment variables required for the action.
//nolint
func initalizeENVVariables() error {
	/*
	 https://docs.github.com/en/actions/learn-github-actions/environment-variables
	   GITHUB_EVENT_PATH contains the json file for the event.
	   GITHUB_SHA contains the commit hash.
	   GITHUB_WORKSPACE contains the repo folder.
	   GITHUB_EVENT_NAME contains the event name.
	   GITHUB_ACTIONS is true in GitHub env.
	*/

	envvars := make(map[string]string)
	envvars["ENABLE_SARIF"] = "1"
	envvars["ENABLE_LICENSE"] = "1"
	envvars["ENABLE_DANGEROUS_WORKFLOW"] = "1"
	envvars["SCORECARD_POLICY_FILE"] = "./policy.yml"
	envvars["SCORECARD_BIN"] = "/scorecard"
	envvars["ENABLED_CHECKS"] = ""

	for key, val := range envvars {
		if err := os.Setenv(key, val); err != nil {
			return fmt.Errorf("error setting %s: %w", key, err)
		}
	}

	if result, exists := os.LookupEnv("INPUT_RESULTS_FILE"); !exists {
		return errInputResultFileNotSet
	} else {
		if result == "" {
			return errInputResultFileEmpty
		}
		if err := os.Setenv("SCORECARD_RESULTS_FILE", result); err != nil {
			return fmt.Errorf("error setting SCORECARD_RESULTS_FILE: %w", err)
		}
	}

	if result, exists := os.LookupEnv("INPUT_RESULTS_FORMAT"); !exists {
		return errInputResultFormatNotSet
	} else {
		if result == "" {
			return errInputResultFormatEmtpy
		}
		if err := os.Setenv("SCORECARD_RESULTS_FORMAT", result); err != nil {
			return fmt.Errorf("error setting SCORECARD_RESULTS_FORMAT: %w", err)
		}
	}

	if result, exists := os.LookupEnv("INPUT_PUBLISH_RESULTS"); !exists {
		return errInputPublishResultsNotSet
	} else {
		if result == "" {
			return errInputPublishResultsEmpty
		}
		if err := os.Setenv("SCORECARD_PUBLISH_RESULTS", result); err != nil {
			return fmt.Errorf("error setting SCORECARD_PUBLISH_RESULTS: %w", err)
		}
	}

	return gitHubEventPath()
}

// gitHubEventPath is a function to get the path to the GitHub event
// and sets the SCORECARD_IS_FORK environment variable.
func gitHubEventPath() error {
	var result string
	var exists bool

	if result, exists = os.LookupEnv("GITHUB_EVENT_PATH"); !exists {
		return errGitHubEventPathNotSet
	}

	if result == "" {
		return errGitHubEventPathEmpty
	}

	data, err := ioutil.ReadFile(result)
	if err != nil {
		return fmt.Errorf("error reading GITHUB_EVENT_PATH: %w", err)
	}
	var isFork bool

	if isFork, err = scorecardIsFork(string(data)); err != nil {
		return fmt.Errorf("error checking if scorecard is a fork: %w", err)
	}

	if isFork {
		if err := os.Setenv("SCORECARD_IS_FORK", "true"); err != nil {
			return fmt.Errorf("error setting SCORECARD_IS_FORK: %w", err)
		}
	} else {
		if err := os.Setenv("SCORECARD_IS_FORK", "false"); err != nil {
			return fmt.Errorf("error setting SCORECARD_IS_FORK: %w", err)
		}
	}

	return nil
}

// scorecardIsFork is a function to check if the current repo is a fork.
func scorecardIsFork(ghEventPath string) (bool, error) {
	if ghEventPath == "" {
		return false, errGitHubEventPath
	}
	/*
	 https://docs.github.com/en/actions/reference/workflow-commands-for-github-actions#github_repository_is_fork
	   GITHUB_REPOSITORY_IS_FORK is true if the repository is a fork.
	*/
	type repo struct {
		Repository struct {
			Fork bool `json:"fork"`
		} `json:"repository"`
	}
	var r repo
	if err := json.Unmarshal([]byte(ghEventPath), &r); err != nil {
		return false, fmt.Errorf("error unmarshalling ghEventPath: %w", err)
	}

	return r.Repository.Fork, nil
}

// checkIfRequiredENVSet is a function to check if the required environment variables are set.
func checkIfRequiredENVSet() error {
	envVariables := make(map[string]bool)
	envVariables["GITHUB_REPOSITORY"] = true
	envVariables["GITHUB_AUTH_TOKEN"] = true

	for key := range envVariables {
		if _, exists := os.LookupEnv(key); !exists {
			return errRequiredENVNotSet
		}
	}
	return nil
}

// getRepositoryInformation is a function to get the repository information.
// It is decided to not use the golang GitHub library because of the
// dependency on the github.com/google/go-github/github library
// which will in turn require other dependencies.
func getRepositoryInformation(name, githubauthToken string) (repositoryInformation, error) {
	//nolint
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s", name), nil)
	if err != nil {
		return repositoryInformation{}, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Authorization", githubauthToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return repositoryInformation{}, fmt.Errorf("error creating request: %w", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return repositoryInformation{}, fmt.Errorf("error reading response body: %w", err)
	}
	var r repositoryInformation
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return repositoryInformation{}, fmt.Errorf("error decoding response body: %w", err)
	}
	return r, nil
}

// updateRepositoryInformation is a function to update the repository information into ENV variables.
func updateRepositoryInformation(privateRepo bool, defaultBranch string) error {
	if defaultBranch == "" {
		return errEmptyDefaultBranch
	}

	if err := os.Setenv("SCORECARD_PRIVATE_REPOSITORY", strconv.FormatBool(privateRepo)); err != nil {
		return fmt.Errorf("error setting SCORECARD_PRIVATE_REPOSITORY: %w", err)
	}
	if err := os.Setenv("SCORECARD_DEFAULT_BRANCH", defaultBranch); err != nil {
		return fmt.Errorf("error setting SCORECARD_DEFAULT_BRANCH: %w", err)
	}
	return nil
}

// updateEnvVariables is a function to update the ENV variables based on results format and private repository.
func updateEnvVariables() error {
	resultsFileFormat := os.Getenv("SCORECARD_RESULTS_FORMAT")
	if resultsFileFormat != "sarif" {
		os.Unsetenv("SCORECARD_POLICY_FILE")
	}
	isPrivateRepo := os.Getenv("SCORECARD_PRIVATE_REPOSITORY")
	if isPrivateRepo != "true" {
		if err := os.Setenv("SCORECARD_PUBLISH_RESULTS", "false"); err != nil {
			return fmt.Errorf("error setting SCORECARD_PUBLISH_RESULTS: %w", err)
		}
	}
	return nil
}
