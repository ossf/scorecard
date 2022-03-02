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
		"",
		"repository to check",
	)

	cmd.Flags().StringVar(
		&o.Local,
		"local",
		"",
		"local folder to check",
	)

	cmd.Flags().StringVar(
		&o.Commit,
		"commit",
		DefaultCommit,
		"commit to analyze",
	)

	cmd.Flags().StringVar(
		&o.LogLevel,
		"verbosity",
		DefaultLogLevel,
		"set the log level",
	)

	cmd.Flags().StringVar(
		&o.NPM,
		"npm",
		"",
		"npm package to check, given that the npm package has a GitHub repository",
	)

	cmd.Flags().StringVar(
		&o.PyPI,
		"pypi",
		"",
		"pypi package to check, given that the pypi package has a GitHub repository",
	)

	cmd.Flags().StringVar(
		&o.RubyGems,
		"rubygems",
		"",
		"rubygems package to check, given that the rubygems package has a GitHub repository",
	)

	cmd.Flags().StringSliceVar(
		&o.Metadata,
		"metadata",
		[]string{},
		"metadata for the project. It can be multiple separated by commas",
	)

	cmd.Flags().BoolVar(
		&o.ShowDetails,
		"show-details",
		false,
		"show extra details about each check",
	)

	checkNames := []string{}
	for checkName := range checks.GetAll() {
		checkNames = append(checkNames, checkName)
	}
	cmd.Flags().StringSliceVar(
		&o.ChecksToRun,
		"checks",
		[]string{},
		fmt.Sprintf("Checks to run. Possible values are: %s", strings.Join(checkNames, ",")),
	)

	// TODO(options): Extract logic
	if IsSarifEnabled() {
		cmd.Flags().StringVar(
			&o.PolicyFile,
			"policy",
			"",
			"policy to enforce",
		)

		cmd.Flags().StringVar(
			&o.Format,
			"format",
			FormatDefault,
			"output format allowed values are [default, sarif, json]",
		)
	} else {
		cmd.Flags().StringVar(
			&o.Format,
			"format",
			FormatDefault,
			"output format allowed values are [default, json]",
		)
	}
}
