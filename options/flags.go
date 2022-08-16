// Copyright OpenSSF Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package options

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ossf/scorecard/v4/checks"
)

const (
	// FlagRepo is the flag name for specifying a repository.
	FlagRepo = "repo"

	// FlagLocal is the flag name for specifying a local run.
	FlagLocal = "local"

	// FlagCommit is the flag name for specifying a commit.
	FlagCommit = "commit"

	// FlagLogLevel is the flag name for specifying the log level.
	FlagLogLevel = "verbosity"

	// FlagNPM is the flag name for specifying a NPM repository.
	FlagNPM = "npm"

	// FlagPyPI is the flag name for specifying a PyPI repository.
	FlagPyPI = "pypi"

	// FlagRubyGems is the flag name for specifying a RubyGems repository.
	FlagRubyGems = "rubygems"

	// FlagMetadata is the flag name for specifying metadata for the project.
	FlagMetadata = "metadata"

	// FlagShowDetails is the flag name for outputting additional check info.
	FlagShowDetails = "show-details"

	// FlagChecks is the flag name for specifying which checks to run.
	FlagChecks = "checks"

	// FlagPolicyFile is the flag name for specifying a policy file.
	FlagPolicyFile = "policy"

	// FlagFormat is the flag name for specifying output format.
	FlagFormat = "format"
)

// Command is an interface for handling options for command-line utilities.
type Command interface {
	// AddFlags adds this options' flags to the cobra command.
	AddFlags(cmd *cobra.Command)
}

// AddFlags adds this options' flags to the cobra command.
func (o *Options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&o.Repo,
		FlagRepo,
		o.Repo,
		"repository to check (valid inputs: \"owner/repo\", \"github.com/owner/repo\", \"https://github.com/repo\")",
	)

	cmd.Flags().StringVar(
		&o.Local,
		FlagLocal,
		o.Local,
		"local folder to check",
	)

	cmd.Flags().StringVar(
		&o.Commit,
		FlagCommit,
		o.Commit,
		"commit to analyze",
	)

	cmd.Flags().StringVar(
		&o.LogLevel,
		FlagLogLevel,
		o.LogLevel,
		"set the log level",
	)

	cmd.Flags().StringVar(
		&o.NPM,
		FlagNPM,
		o.NPM,
		"npm package to check, given that the npm package has a GitHub repository",
	)

	cmd.Flags().StringVar(
		&o.PyPI,
		FlagPyPI,
		o.PyPI,
		"pypi package to check, given that the pypi package has a GitHub repository",
	)

	cmd.Flags().StringVar(
		&o.RubyGems,
		FlagRubyGems,
		o.RubyGems,
		"rubygems package to check, given that the rubygems package has a GitHub repository",
	)

	cmd.Flags().StringSliceVar(
		&o.Metadata,
		FlagMetadata,
		o.Metadata,
		"metadata for the project. It can be multiple separated by commas",
	)

	cmd.Flags().BoolVar(
		&o.ShowDetails,
		FlagShowDetails,
		o.ShowDetails,
		"show extra details about each check",
	)

	checkNames := []string{}
	for checkName := range checks.GetAll() {
		checkNames = append(checkNames, checkName)
	}
	cmd.Flags().StringSliceVar(
		&o.ChecksToRun,
		FlagChecks,
		o.ChecksToRun,
		fmt.Sprintf("Checks to run. Possible values are: %s", strings.Join(checkNames, ",")),
	)

	// TODO(options): Extract logic
	allowedFormats := []string{
		FormatDefault,
		FormatJSON,
	}

	if o.isSarifEnabled() {
		cmd.Flags().StringVar(
			&o.PolicyFile,
			FlagPolicyFile,
			o.PolicyFile,
			"policy to enforce",
		)

		allowedFormats = append(allowedFormats, FormatSarif)
	}

	cmd.Flags().StringVar(
		&o.Format,
		FlagFormat,
		o.Format,
		fmt.Sprintf(
			"output format. Possible values are: %s",
			strings.Join(allowedFormats, ", "),
		),
	)
}
