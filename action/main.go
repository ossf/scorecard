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
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	errInputResultFileNotSet      = errors.New("INPUT_RESULTS_FILE is not set")
	errInputResultFileEmpty       = errors.New("INPUT_RESULTS_FILE is empty")
	errInputResultFormatNotSet    = errors.New("INPUT_RESULTS_FORMAT is not set")
	errInputResultFormatEmtpy     = errors.New("INPUT_RESULTS_FORMAT is empty")
	errInputPublishResultsNotSet  = errors.New("INPUT_PUBLISH_RESULTS is not set")
	errInputPublishResultsEmpty   = errors.New("INPUT_PUBLISH_RESULTS is empty")
	errRequiredENVNotSet          = errors.New("required environment variables are not set")
	errGitHubEventPath            = errors.New("error getting GITHUB_EVENT_PATH")
	errGitHubEventPathEmpty       = errors.New("GITHUB_EVENT_PATH is empty")
	errGitHubEventPathNotSet      = errors.New("GITHUB_EVENT_PATH is not set")
	errEmptyDefaultBranch         = errors.New("default branch is empty")
	errEmptyGitHubAuthToken       = errors.New("repo_token variable is empty")
	errOnlyDefaultBranchSupported = errors.New("only default branch is supported")
)

type repositoryInformation struct {
	DefaultBranch string `json:"default_branch"`
	Private       bool   `json:"private"`
}

const (
	enableSarif             = "ENABLE_SARIF"
	enableLicense           = "ENABLE_LICENSE"
	enableDangerousWorkflow = "ENABLE_DANGEROUS_WORKFLOW"
	enabledChecks           = "ENABLED_CHECKS"
	githubEventPath         = "GITHUB_EVENT_PATH"
	githubEventName         = "GITHUB_EVENT_NAME"
	githubRepository        = "GITHUB_REPOSITORY"
	githubRef               = "GITHUB_REF"
	//nolint:gosec
	githubAuthToken            = "GITHUB_AUTH_TOKEN"
	inputresultsfile           = "INPUT_RESULTS_FILE"
	inputresultsformat         = "INPUT_RESULTS_FORMAT"
	inputpublishresults        = "INPUT_PUBLISH_RESULTS"
	scorecardBin               = "SCORECARD_BIN"
	scorecardResultsFormat     = "SCORECARD_RESULTS_FORMAT"
	scorecardPublishResults    = "SCORECARD_PUBLISH_RESULTS"
	scorecardPolicyFile        = "SCORECARD_POLICY_FILE"
	scorecardResultsFile       = "SCORECARD_RESULTS_FILE"
	scorecardFork              = "SCORECARD_IS_FORK"
	scorecardDefaultBranch     = "SCORECARD_DEFAULT_BRANCH"
	scorecardPrivateRepository = "SCORECARD_PRIVATE_REPOSITORY"
	sarif                      = "sarif"
)

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

	repository := os.Getenv(githubRepository)
	token := os.Getenv(githubAuthToken)

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

	printEnvVariables(os.Stdout)

	if err := validate(os.Stderr); err != nil {
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
	envvars[enableSarif] = "1"
	envvars[enableLicense] = "1"
	envvars[enableDangerousWorkflow] = "1"
	envvars[scorecardPolicyFile] = "./policy.yml"
	envvars[scorecardBin] = "/scorecard"
	envvars[enabledChecks] = ""

	for key, val := range envvars {
		if err := os.Setenv(key, val); err != nil {
			return fmt.Errorf("error setting %s: %w", key, err)
		}
	}

	if result, exists := os.LookupEnv(inputresultsfile); !exists {
		return errInputResultFileNotSet
	} else {
		if result == "" {
			return errInputResultFileEmpty
		}
		if err := os.Setenv(scorecardResultsFile, result); err != nil {
			return fmt.Errorf("error setting %s: %w", scorecardResultsFile, err)
		}
	}

	if result, exists := os.LookupEnv(inputresultsformat); !exists {
		return errInputResultFormatNotSet
	} else {
		if result == "" {
			return errInputResultFormatEmtpy
		}
		if err := os.Setenv(scorecardResultsFormat, result); err != nil {
			return fmt.Errorf("error setting %s: %w", scorecardResultsFormat, err)
		}
	}

	if result, exists := os.LookupEnv(inputpublishresults); !exists {
		return errInputPublishResultsNotSet
	} else {
		if result == "" {
			return errInputPublishResultsEmpty
		}
		if err := os.Setenv(scorecardPublishResults, result); err != nil {
			return fmt.Errorf("error setting %s: %w", scorecardPublishResults, err)
		}
	}

	return gitHubEventPath()
}

// gitHubEventPath is a function to get the path to the GitHub event
// and sets the SCORECARD_IS_FORK environment variable.
func gitHubEventPath() error {
	var result string
	var exists bool

	if result, exists = os.LookupEnv(githubEventPath); !exists {
		return errGitHubEventPathNotSet
	}

	if result == "" {
		return errGitHubEventPathEmpty
	}

	data, err := ioutil.ReadFile(result)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", githubEventPath, err)
	}
	var isFork bool

	if isFork, err = scorecardIsFork(string(data)); err != nil {
		return fmt.Errorf("error checking if scorecard is a fork: %w", err)
	}

	if isFork {
		if err := os.Setenv(scorecardFork, "true"); err != nil {
			return fmt.Errorf("error setting %s: %w", scorecardFork, err)
		}
	} else {
		if err := os.Setenv(scorecardFork, "false"); err != nil {
			return fmt.Errorf("error setting %s: %w", scorecardFork, err)
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
	envVariables[githubRepository] = true
	envVariables[githubAuthToken] = true

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

	if err := os.Setenv(scorecardPrivateRepository, strconv.FormatBool(privateRepo)); err != nil {
		return fmt.Errorf("error setting %s: %w", scorecardPrivateRepository, err)
	}
	if err := os.Setenv(scorecardDefaultBranch, fmt.Sprintf("refs/heads/%s", defaultBranch)); err != nil {
		return fmt.Errorf("error setting %s: %w", scorecardDefaultBranch, err)
	}
	return nil
}

// updateEnvVariables is a function to update the ENV variables based on results format and private repository.
func updateEnvVariables() error {
	resultsFileFormat := os.Getenv(scorecardResultsFormat)
	if resultsFileFormat != sarif {
		if err := os.Unsetenv(scorecardPolicyFile); err != nil {
			return fmt.Errorf("error unsetting %s: %w", scorecardPolicyFile, err)
		}
	}
	isPrivateRepo := os.Getenv(scorecardPrivateRepository)
	if isPrivateRepo != "true" {
		if err := os.Setenv(scorecardPublishResults, "false"); err != nil {
			return fmt.Errorf("error setting %s: %w", scorecardPublishResults, err)
		}
	}
	return nil
}

// printEnvVariables is a function to print the ENV variables.
func printEnvVariables(writer io.Writer) {
	fmt.Fprintf(writer, "GITHUB_EVENT_PATH=%s\n", os.Getenv(githubEventPath))
	fmt.Fprintf(writer, "GITHUB_EVENT_NAME=%s\n", os.Getenv(githubEventName))
	fmt.Fprintf(writer, "GITHUB_REPOSITORY=%s\n", os.Getenv(githubRepository))
	fmt.Fprintf(writer, "SCORECARD_IS_FORK=%s\n", os.Getenv(scorecardFork))
	fmt.Fprintf(writer, "Ref=%s\n", os.Getenv(githubRef))
	fmt.Fprintf(writer, "SCORECARD_PRIVATE_REPOSITORY=%s\n", os.Getenv(scorecardPrivateRepository))
	fmt.Fprintf(writer, "SCORECARD_PUBLISH_RESULTS=%s\n", os.Getenv(scorecardPublishResults))
	fmt.Fprintf(writer, "Format=%s\n", os.Getenv(scorecardResultsFormat))
	fmt.Fprintf(writer, "Policy file=%s\n", os.Getenv(scorecardPolicyFile))
	fmt.Fprintf(writer, "Default branch=%s\n", os.Getenv(scorecardDefaultBranch))
}

// validate is a function to validate the scorecard configuration based on the environment variables.
func validate(writer io.Writer) error {
	if os.Getenv(githubAuthToken) == "" {
		fmt.Fprintf(writer, "The 'repo_token' variable is empty.\n")
		if os.Getenv(scorecardFork) == "true" {
			fmt.Fprintf(writer, "We have detected you are running on a fork.\n")
		}
		//nolint:lll
		fmt.Fprintf(writer,
			"Please follow the instructions at https://github.com/ossf/scorecard-action#authentication to create the read-only PAT token.\n")
		return errEmptyGitHubAuthToken
	}
	if strings.Contains(os.Getenv(githubEventName), "pull_request") &&
		os.Getenv(githubRef) == os.Getenv(scorecardDefaultBranch) {
		fmt.Fprintf(writer, "%s not supported with %s event.\n", os.Getenv(githubRef), os.Getenv(githubEventName))
		fmt.Fprintf(writer, "Only the default branch %s is supported.\n", os.Getenv(scorecardDefaultBranch))
		return errOnlyDefaultBranchSupported
	}
	return nil
}
