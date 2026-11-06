// Copyright 2022 OpenSSF Scorecard Authors
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
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ossf/scorecard/v5/checks"
)

const (
	// FlagRepo is the flag name for specifying a repository.
	FlagRepo  = "repo"
	FlagRepos = "repos"
	FlagOrg   = "org"

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

	// FlagNuget is the flag name for specifying a Nuget repository.
	FlagNuget = "nuget"

	// FlagMetadata is the flag name for specifying metadata for the project.
	FlagMetadata = "metadata"

	// FlagShowDetails is the flag name for outputting additional check info.
	FlagShowDetails = "show-details"

	// Flag FlagFileMode is the flag name for specifying how files are fetched for a repository.
	FlagFileMode = "file-mode"

	// FlagShowAnnotations is the flag name for outputting annotations on checks.
	FlagShowAnnotations = "show-annotations"

	// FlagChecks is the flag name for specifying which checks to run.
	FlagChecks = "checks"

	// FlagPolicyFile is the flag name for specifying a policy file.
	FlagPolicyFile = "policy"

	// FlagFormat is the flag name for specifying output format.
	FlagFormat = "format"

	// FlagResultsFile is the flag name for specifying output file.
	FlagResultsFile = "output"

	// ShorthandFlagResultsFile is the shorthand flag name for specifying output file.
	ShorthandFlagResultsFile = "o"

	FlagCommitDepth = "commit-depth"

	FlagProbes = "probes"
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
		"repository to check (valid inputs: \"owner/repo\", \"github.com/owner/repo\", "+
			"\"https://github.com/repo\", \"gitlab.com/owner/repo\")",
	)

	cmd.Flags().StringSliceVar(
		&o.Repos,
		FlagRepos,
		o.Repos,
		"repositories to check (<repo1>,<repo2>, ...). Each repo must be one of:"+
			" \"owner/repo\", \"github.com/owner/repo\", \"https://github.com/repo\","+
			" \"gitlab.com/owner/repo\"",
	)

	cmd.Flags().StringVar(
		&o.Org,
		FlagOrg,
		o.Org,
		"scans all non-archived repositories in an organization. "+
			"Currently only supports GitHub, e.g., 'github.com/ossf' or 'ossf'",
	)

	cmd.Flags().StringVar(
		&o.Local,
		FlagLocal,
		o.Local,
		"local folder to check (runs only file-based checks; repository host"+
			" checks, e.g. Branch-Protection, are skipped)",
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
		"log level: info, debug, or warn",
	)

	cmd.Flags().StringVar(
		&o.NPM,
		FlagNPM,
		o.NPM,
		"npm package to check; the package must map to a GitHub repository",
	)

	cmd.Flags().StringVar(
		&o.PyPI,
		FlagPyPI,
		o.PyPI,
		"pypi package to check; the package must map to a GitHub repository",
	)

	cmd.Flags().StringVar(
		&o.RubyGems,
		FlagRubyGems,
		o.RubyGems,
		"rubygems package to check; the package must map to a GitHub repository",
	)

	cmd.Flags().StringVar(
		&o.Nuget,
		FlagNuget,
		o.Nuget,
		"nuget package to check; the package must map to a GitHub repository",
	)

	cmd.Flags().StringSliceVar(
		&o.Metadata,
		FlagMetadata,
		o.Metadata,
		"comma-separated metadata to attach to the results (included in JSON output)",
	)

	cmd.Flags().BoolVar(
		&o.ShowDetails,
		FlagShowDetails,
		o.ShowDetails,
		"show extra details about each check",
	)

	cmd.Flags().BoolVar(
		&o.ShowAnnotations,
		FlagShowAnnotations,
		o.ShowAnnotations,
		"show maintainers' annotations for checks",
	)

	cmd.Flags().IntVar(
		&o.CommitDepth,
		FlagCommitDepth,
		o.CommitDepth,
		"number of recent commits to analyze, starting from HEAD",
	)

	checkNames := []string{}
	for checkName := range checks.GetAll() {
		checkNames = append(checkNames, checkName)
	}
	sort.Strings(checkNames)
	cmd.Flags().StringSliceVar(
		&o.ChecksToRun,
		FlagChecks,
		o.ChecksToRun,
		fmt.Sprintf("checks to run. Possible values are: %s", strings.Join(checkNames, ", ")),
	)

	cmd.Flags().StringSliceVar(
		&o.ProbesToRun,
		FlagProbes,
		o.ProbesToRun,
		"individual probes to run instead of checks; when set, --checks is ignored"+
			" (probe names appear in --format=probe output)",
	)

	// TODO(options): Extract logic
	allowedFormats := []string{
		FormatDefault,
		FormatJSON,
		FormatProbe,
		FormatInToto,
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

	cmd.Flags().StringVarP(
		&o.ResultsFile,
		FlagResultsFile,
		ShorthandFlagResultsFile,
		o.ResultsFile,
		"write results to this file instead of stdout",
	)

	cmd.Flags().StringVar(
		&o.FileMode,
		FlagFileMode,
		o.FileMode,
		"fetch repository files via \"archive\" (download a tarball) or \"git\" (clone)",
	)
}
