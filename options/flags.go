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

// Command is an interface for handling options for command-line utilities.
type Command interface {
	// AddFlags adds this options' flags to the cobra command.
	AddFlags(cmd *cobra.Command)
}

// AddFlags adds this options' flags to the cobra command.
func (o *Options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&o.Repo,
		"repo",
		o.Repo,
		"repository to check",
	)

	cmd.Flags().StringVar(
		&o.Local,
		"local",
		o.Local,
		"local folder to check",
	)

	// TODO(v5): Should this be behind a feature flag?
	cmd.Flags().StringVar(
		&o.Commit,
		"commit",
		o.Commit,
		"commit to analyze",
	)

	cmd.Flags().StringVar(
		&o.LogLevel,
		"verbosity",
		o.LogLevel,
		"set the log level",
	)

	cmd.Flags().StringVar(
		&o.NPM,
		"npm",
		o.NPM,
		"npm package to check, given that the npm package has a GitHub repository",
	)

	cmd.Flags().StringVar(
		&o.PyPI,
		"pypi",
		o.PyPI,
		"pypi package to check, given that the pypi package has a GitHub repository",
	)

	cmd.Flags().StringVar(
		&o.RubyGems,
		"rubygems",
		o.RubyGems,
		"rubygems package to check, given that the rubygems package has a GitHub repository",
	)

	cmd.Flags().StringSliceVar(
		&o.Metadata,
		"metadata",
		o.Metadata,
		"metadata for the project. It can be multiple separated by commas",
	)

	cmd.Flags().BoolVar(
		&o.ShowDetails,
		"show-details",
		o.ShowDetails,
		"show extra details about each check",
	)

	checkNames := []string{}
	for checkName := range checks.GetAll() {
		checkNames = append(checkNames, checkName)
	}
	cmd.Flags().StringSliceVar(
		&o.ChecksToRun,
		"checks",
		o.ChecksToRun,
		fmt.Sprintf("Checks to run. Possible values are: %s", strings.Join(checkNames, ",")),
	)

	// TODO(options): Extract logic
	if o.isSarifEnabled() {
		cmd.Flags().StringVar(
			&o.PolicyFile,
			"policy",
			o.PolicyFile,
			"policy to enforce",
		)

		cmd.Flags().StringVar(
			&o.Format,
			"format",
			o.Format,
			"output format allowed values are [default, sarif, json]",
		)
	} else {
		cmd.Flags().StringVar(
			&o.Format,
			"format",
			o.Format,
			"output format allowed values are [default, json]",
		)
	}
}
