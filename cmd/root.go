// Copyright 2020 Security Scorecard Authors
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
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	docs "github.com/ossf/scorecard/v4/docs/checks"
	sclog "github.com/ossf/scorecard/v4/log"
	"github.com/ossf/scorecard/v4/options"
	"github.com/ossf/scorecard/v4/pkg"
	"github.com/ossf/scorecard/v4/policy"
)

const (
	scorecardLong = "A program that shows security scorecard for an open source software."
	scorecardUse  = `./scorecard [--repo=<repo_url>] [--local=folder] [--checks=check1,...]
	 [--show-details] or ./scorecard --{npm,pypi,rubygems}=<package_name> 
	 [--checks=check1,...] [--show-details]`
	scorecardShort = "Security Scorecards"
)

var rootCmd = &cobra.Command{
	Use:   scorecardUse,
	Short: scorecardShort,
	Long:  scorecardLong,
	Run:   scorecardCmd,
}

var opts = options.New()

//nolint:gochecknoinits
func init() {
	rootCmd.Flags().StringVar(&opts.Repo, "repo", "", "repository to check")
	rootCmd.Flags().StringVar(&opts.Local, "local", "", "local folder to check")
	rootCmd.Flags().StringVar(&opts.Commit, "commit", options.DefaultCommit, "commit to analyze")
	rootCmd.Flags().StringVar(
		&opts.LogLevel,
		"verbosity",
		options.DefaultLogLevel,
		"set the log level",
	)
	rootCmd.Flags().StringVar(
		&opts.NPM, "npm", "",
		"npm package to check, given that the npm package has a GitHub repository")
	rootCmd.Flags().StringVar(
		&opts.PyPI, "pypi", "",
		"pypi package to check, given that the pypi package has a GitHub repository")
	rootCmd.Flags().StringVar(
		&opts.RubyGems, "rubygems", "",
		"rubygems package to check, given that the rubygems package has a GitHub repository")
	rootCmd.Flags().StringSliceVar(
		&opts.Metadata, "metadata", []string{}, "metadata for the project. It can be multiple separated by commas")
	rootCmd.Flags().BoolVar(&opts.ShowDetails, "show-details", false, "show extra details about each check")
	checkNames := []string{}
	for checkName := range policy.GetAll() {
		checkNames = append(checkNames, checkName)
	}
	rootCmd.Flags().StringSliceVar(&opts.ChecksToRun, "checks", []string{},
		fmt.Sprintf("Checks to run. Possible values are: %s", strings.Join(checkNames, ",")))

	// TODO(cmd): Extract logic
	if options.IsSarifEnabled() {
		rootCmd.Flags().StringVar(&opts.PolicyFile, "policy", "", "policy to enforce")
		rootCmd.Flags().StringVar(&opts.Format, "format", options.FormatDefault,
			"output format allowed values are [default, sarif, json]")
	} else {
		rootCmd.Flags().StringVar(&opts.Format, "format", options.FormatDefault,
			"output format allowed values are [default, json]")
	}
}

// Execute runs the Scorecard commandline.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func scorecardCmd(cmd *cobra.Command, args []string) {
	RunScorecard(args)
}

// RunScorecard runs scorecard checks given a set of arguments.
// TODO(cmd): Is `args` required?
func RunScorecard(args []string) {
	// TODO(cmd): Catch validation errors
	valErrs := opts.Validate()
	if len(valErrs) != 0 {
		log.Panicf(
			"the following validation errors occurred: %+v",
			valErrs,
		)
	}

	// Set `repo` from package managers.
	pkgResp, err := fetchGitRepositoryFromPackageManagers(opts.NPM, opts.PyPI, opts.RubyGems)
	if err != nil {
		log.Panic(err)
	}
	if pkgResp.exists {
		opts.Repo = pkgResp.associatedRepo
	}

	pol, err := policy.ParseFromFile(opts.PolicyFile)
	if err != nil {
		log.Panicf("readPolicy: %v", err)
	}

	ctx := context.Background()
	logger := sclog.NewLogger(sclog.ParseLevel(opts.LogLevel))
	repoURI, repoClient, ossFuzzRepoClient, ciiClient, vulnsClient, err := checker.GetClients(
		ctx, opts.Repo, opts.Local, logger)
	if err != nil {
		log.Panic(err)
	}
	defer repoClient.Close()
	if ossFuzzRepoClient != nil {
		defer ossFuzzRepoClient.Close()
	}

	// Read docs.
	checkDocs, err := docs.Read()
	if err != nil {
		log.Panicf("cannot read yaml file: %v", err)
	}

	var requiredRequestTypes []checker.RequestType
	if opts.Local != "" {
		requiredRequestTypes = append(requiredRequestTypes, checker.FileBased)
	}
	if !strings.EqualFold(opts.Commit, clients.HeadSHA) {
		requiredRequestTypes = append(requiredRequestTypes, checker.CommitBased)
	}
	enabledChecks, err := policy.GetEnabled(pol, opts.ChecksToRun, requiredRequestTypes)
	if err != nil {
		log.Panic(err)
	}

	if opts.Format == options.FormatDefault {
		for checkName := range enabledChecks {
			fmt.Fprintf(os.Stderr, "Starting [%s]\n", checkName)
		}
	}

	repoResult, err := pkg.RunScorecards(
		ctx,
		repoURI,
		opts.Commit,
		opts.Format == options.FormatRaw,
		enabledChecks,
		repoClient,
		ossFuzzRepoClient,
		ciiClient,
		vulnsClient,
	)
	if err != nil {
		log.Panic(err)
	}
	repoResult.Metadata = append(repoResult.Metadata, opts.Metadata...)

	// Sort them by name
	sort.Slice(repoResult.Checks, func(i, j int) bool {
		return repoResult.Checks[i].Name < repoResult.Checks[j].Name
	})

	if opts.Format == options.FormatDefault {
		for checkName := range enabledChecks {
			fmt.Fprintf(os.Stderr, "Finished [%s]\n", checkName)
		}
		fmt.Println("\nRESULTS\n-------")
	}

	resultsErr := pkg.FormatResults(
		opts,
		&repoResult,
		checkDocs,
		pol,
	)
	if resultsErr != nil {
		log.Panicf("Failed to output results: %v", err)
	}
}
