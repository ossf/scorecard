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
	"fmt"
	"io/ioutil"
	"os"
)

// main is the entrypoint for the action.
func main() {
	// TODO - This is a port of the entrypoint.sh script.
	// This is still a work in progress.
	if err := initalizeENVVariables(); err != nil {
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
	if err := os.Setenv("ENABLE_SARIF", "1"); err != nil {
		return err
	}

	if err := os.Setenv("ENABLE_LICENSE", "1"); err != nil {
		return err
	}

	if err := os.Setenv("ENABLE_DANGEROUS_WORKFLOW", "1"); err != nil {
		return err
	}

	if err := os.Setenv("SCORECARD_POLICY_FILE", "/policy.yml"); err != nil {
		return err
	}

	if result, exists := os.LookupEnv("INPUT_RESULTS_FILE"); !exists {
		return fmt.Errorf("INPUT_RESULTS_FILE is not set")
	} else {
		if result == "" {
			return fmt.Errorf("INPUT_RESULTS_FILE is empty")
		}
		if err := os.Setenv("SCORECARD_RESULTS_FILE", result); err != nil {
			return err
		}
	}

	if result, exists := os.LookupEnv("INPUT_RESULTS_FORMAT"); !exists {
		return fmt.Errorf("INPUT_RESULTS_FORMAT is not set")
	} else {
		if result == "" {
			return fmt.Errorf("INPUT_RESULTS_FORMAT is empty")
		}
		if err := os.Setenv("SCORECARD_RESULTS_FORMAT", result); err != nil {
			return err
		}
	}

	if result, exists := os.LookupEnv("INPUT_PUBLISH_RESULTS"); !exists {
		return fmt.Errorf("INPUT_PUBLISH_RESULTS is not set")
	} else {
		if result == "" {
			return fmt.Errorf("INPUT_PUBLISH_RESULTS is empty")
		}
		if err := os.Setenv("SCORECARD_PUBLISH_RESULTS", result); err != nil {
			return err
		}
	}

	if err := os.Setenv("SCORECARD_BIN", "/scorecard"); err != nil {
		return err
	}

	if err := os.Setenv("ENABLED_CHECKS", ""); err != nil {
		return err
	}
	return gitHubEventPath()
}

// gitHubEventPath is a function to get the path to the GitHub event
// and sets the SCORECARD_IS_FORK environment variable.
func gitHubEventPath() error {
	if result, exists := os.LookupEnv("GITHUB_EVENT_PATH"); !exists {
		return fmt.Errorf("GITHUB_EVENT_PATH is not set")
	} else {
		if result == "" {
			return fmt.Errorf("GITHUB_EVENT_PATH is empty")
		}
		if err := os.Setenv("GITHUB_EVENT_PATH", result); err != nil {
			return err
		}

		data, err := ioutil.ReadFile(result)
		if err != nil {
			return err
		}

		if isFork, err := scorecardIsFork(string(data)); err != nil {
			return err
		} else {
			if isFork {
				if err := os.Setenv("SCORECARD_IS_FORK", "true"); err != nil {
					return err
				}
			} else {
				if err := os.Setenv("SCORECARD_IS_FORK", "false"); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// scorecardIsFork is a function to check if the current repo is a fork.
func scorecardIsFork(ghEventPath string) (bool, error) {
	if ghEventPath == "" {
		return false, fmt.Errorf("ghEventPath is empty")
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
		return false, err
	}

	return r.Repository.Fork, nil
}
