// Copyright OpenSSF Authors
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

package options

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v6"

	"github.com/ossf/scorecard-action/github"
	"github.com/ossf/scorecard/v4/options"
)

var (
	// Errors.
	errGithubEventPathEmpty       = errors.New("GitHub event path is empty")
	errResultsPathEmpty           = errors.New("results path is empty")
	errOnlyDefaultBranchSupported = errors.New("only default branch is supported")

	trueStr = "true"
)

// Options are options for running scorecard via GitHub Actions.
type Options struct {
	// Scorecard options.
	ScorecardOpts *options.Options

	// Scorecard command-line options.
	EnabledChecks string `env:"ENABLED_CHECKS"`

	// Scorecard checks.
	EnableLicense           string `env:"ENABLE_LICENSE"`
	EnableDangerousWorkflow string `env:"ENABLE_DANGEROUS_WORKFLOW"`

	// GitHub options.
	// TODO(github): Consider making this a separate options set so we can
	//               encapsulate handling
	GithubEventName  string `env:"GITHUB_EVENT_NAME"`
	GithubEventPath  string `env:"GITHUB_EVENT_PATH"`
	GithubRef        string `env:"GITHUB_REF"`
	GithubRepository string `env:"GITHUB_REPOSITORY"`
	GithubWorkspace  string `env:"GITHUB_WORKSPACE"`

	DefaultBranch string `env:"SCORECARD_DEFAULT_BRANCH"`
	// TODO(options): This may be better as a bool
	IsForkStr string `env:"SCORECARD_IS_FORK"`
	// TODO(options): This may be better as a bool
	PrivateRepoStr string `env:"SCORECARD_PRIVATE_REPOSITORY"`

	// Input parameters
	InputResultsFile    string `env:"INPUT_RESULTS_FILE"`
	InputResultsFormat  string `env:"INPUT_RESULTS_FORMAT"`
	InputPublishResults string `env:"INPUT_PUBLISH_RESULTS"`
}

const (
	defaultScorecardPolicyFile = "/policy.yml"
	formatSarif                = options.FormatSarif
)

// New creates a new options set for running scorecard via GitHub Actions.
func New() (*Options, error) {
	// Enable scorecard command to use SARIF format.
	os.Setenv(options.EnvVarEnableSarif, trueStr)

	opts := &Options{
		ScorecardOpts: options.New(),
	}
	if err := env.Parse(opts); err != nil {
		return opts, fmt.Errorf("parsing entrypoint env vars: %w", err)
	}

	if err := opts.Initialize(); err != nil {
		return opts, fmt.Errorf(
			"initializing scorecard-action options: %w",
			err,
		)
	}

	// TODO(options): Move this set-or-default logic to its own function.
	opts.ScorecardOpts.Format = formatSarif
	opts.ScorecardOpts.EnableSarif = true
	if opts.InputResultsFormat != "" {
		opts.ScorecardOpts.Format = opts.InputResultsFormat
	}

	if opts.ScorecardOpts.Format == formatSarif {
		if opts.ScorecardOpts.PolicyFile == "" {
			// TODO(policy): Should we default or error here?
			opts.ScorecardOpts.PolicyFile = defaultScorecardPolicyFile
		}
	}

	// TODO(scorecard): Reset commit options. Fix this in scorecard.
	opts.ScorecardOpts.Commit = options.DefaultCommit

	if err := opts.ScorecardOpts.Validate(); err != nil {
		return opts, fmt.Errorf("validating scorecard options: %w", err)
	}

	opts.SetPublishResults()

	if opts.ScorecardOpts.ResultsFile == "" {
		opts.ScorecardOpts.ResultsFile = opts.InputResultsFile
	}

	if opts.ScorecardOpts.ResultsFile == "" {
		// TODO(test): Reassess test case for this code path
		return opts, errResultsPathEmpty
	}

	if err := opts.Validate(); err != nil {
		return opts, fmt.Errorf("validating scorecard-action options: %w", err)
	}

	return opts, nil
}

// Initialize initializes the environment variables required for the action.
func (o *Options) Initialize() error {
	/*
	 https://docs.github.com/en/actions/learn-github-actions/environment-variables
	   GITHUB_EVENT_PATH contains the json file for the event.
	   GITHUB_SHA contains the commit hash.
	   GITHUB_WORKSPACE contains the repo folder.
	   GITHUB_EVENT_NAME contains the event name.
	   GITHUB_ACTIONS is true in GitHub env.
	*/

	// TODO(checks): Do we actually expect to use these?
	// o.EnableLicense = "1"
	// o.EnableDangerousWorkflow = "1"

	_, tokenSet := os.LookupEnv(EnvGithubAuthToken)
	if !tokenSet {
		inputToken := os.Getenv(EnvInputRepoToken)
		os.Setenv(EnvGithubAuthToken, inputToken)
	}

	return o.SetRepoInfo()
}

// Validate validates the scorecard configuration.
func (o *Options) Validate() error {
	if os.Getenv(EnvGithubAuthToken) == "" {
		fmt.Printf("The 'repo_token' variable is empty.\n")
		if o.IsForkStr == trueStr {
			fmt.Printf("We have detected you are running on a fork.\n")
		}

		fmt.Printf(
			"Please follow the instructions at https://github.com/ossf/scorecard-action#authentication to create the read-only PAT token.\n", //nolint:lll
		)

		return errEmptyGitHubAuthToken
	}

	if strings.Contains(o.GithubEventName, "pull_request") &&
		o.GithubRef == o.DefaultBranch {
		fmt.Printf("%s not supported with %s event.\n", o.GithubRef, o.GithubEventName)
		fmt.Printf("Only the default branch %s is supported.\n", o.DefaultBranch)

		return errOnlyDefaultBranchSupported
	}

	return nil
}

// Print is a function to print options.
func (o *Options) Print() {
	fmt.Printf("Event file: %s\n", o.GithubEventPath)
	fmt.Printf("Event name: %s\n", o.GithubEventName)
	fmt.Printf("Ref: %s\n", o.ScorecardOpts.Commit)
	fmt.Printf("Repository: %s\n", o.ScorecardOpts.Repo)
	fmt.Printf("Fork repository: %s\n", o.IsForkStr)
	fmt.Printf("Private repository: %s\n", o.PrivateRepoStr)
	fmt.Printf("Publication enabled: %+v\n", o.ScorecardOpts.PublishResults)
	fmt.Printf("Format: %s\n", o.ScorecardOpts.Format)
	fmt.Printf("Policy file: %s\n", o.ScorecardOpts.PolicyFile)
	fmt.Printf("Default branch: %s\n", o.DefaultBranch)
}

// SetPublishResults sets whether results should be published based on a
// repository's visibility.
func (o *Options) SetPublishResults() {
	privateRepo, err := strconv.ParseBool(o.PrivateRepoStr)
	if err != nil {
		// TODO(options): Consider making this an error.
		fmt.Printf(
			"parsing bool from %s: %+v\n",
			o.PrivateRepoStr,
			err,
		)
	}

	if privateRepo {
		o.ScorecardOpts.PublishResults = false
	}
}

// SetRepoInfo gets the path to the GitHub event and sets the
// SCORECARD_IS_FORK environment variable.
// TODO(options): Check if this actually needs to be exported.
// TODO(options): Choose a more accurate name for what this does.
func (o *Options) SetRepoInfo() error {
	eventPath := o.GithubEventPath
	if eventPath == "" {
		return errGithubEventPathEmpty
	}

	repoInfo, err := ioutil.ReadFile(eventPath)
	if err != nil {
		return fmt.Errorf("reading GitHub event path: %w", err)
	}

	var r github.RepoInfo
	if err := json.Unmarshal(repoInfo, &r); err != nil {
		return fmt.Errorf("unmarshalling repo info: %w", err)
	}

	o.PrivateRepoStr = strconv.FormatBool(r.Repo.Private)
	o.IsForkStr = strconv.FormatBool(r.Repo.Fork)
	o.DefaultBranch = r.Repo.DefaultBranch

	return nil
}
