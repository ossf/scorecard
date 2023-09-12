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
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/release-utils/version"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	pmc "github.com/ossf/scorecard/v4/cmd/internal/packagemanager"
	docs "github.com/ossf/scorecard/v4/docs/checks"
	sce "github.com/ossf/scorecard/v4/errors"
	sclog "github.com/ossf/scorecard/v4/log"
	"github.com/ossf/scorecard/v4/options"
	"github.com/ossf/scorecard/v4/pkg"
	"github.com/ossf/scorecard/v4/policy"
)

const (
	scorecardLong = "A program that shows the OpenSSF scorecard for an open source software."
	scorecardUse  = `./scorecard (--repo=<repo> | --local=<folder> | --{npm,pypi,rubygems,nuget}=<package_name>)
	 [--checks=check1,...] [--show-details]`
	scorecardShort = "OpenSSF Scorecard"
)

// New creates a new instance of the scorecard command.
func New(o *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   scorecardUse,
		Short: scorecardShort,
		Long:  scorecardLong,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate()
			if err != nil {
				return fmt.Errorf("validating options: %w", err)
			}
			// options are good at this point. silence usage so it doesn't print for runtime errors
			cmd.SilenceUsage = true
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return rootCmd(o)
		},
	}

	o.AddFlags(cmd)

	// Add sub-commands.
	cmd.AddCommand(serveCmd(o))
	cmd.AddCommand(version.Version())
	return cmd
}

// rootCmd runs scorecard checks given a set of arguments.
func rootCmd(o *options.Options) error {
	p := &pmc.PackageManagerClient{}
	// Set `repo` from package managers.
	pkgResp, err := fetchGitRepositoryFromPackageManagers(o.NPM, o.PyPI, o.RubyGems, o.Nuget, p)
	if err != nil {
		return fmt.Errorf("fetchGitRepositoryFromPackageManagers: %w", err)
	}
	if pkgResp.exists {
		o.Repo = pkgResp.associatedRepo
	}

	pol, err := policy.ParseFromFile(o.PolicyFile)
	if err != nil {
		return fmt.Errorf("readPolicy: %w", err)
	}

	ctx := context.Background()
	logger := sclog.NewLogger(sclog.ParseLevel(o.LogLevel))
	repoURI, repoClient, ossFuzzRepoClient, ciiClient, vulnsClient, err := checker.GetClients(
		ctx, o.Repo, o.Local, logger) // MODIFIED
	if err != nil {
		return fmt.Errorf("GetClients: %w", err)
	}

	defer repoClient.Close()
	if ossFuzzRepoClient != nil {
		defer ossFuzzRepoClient.Close()
	}

	// Read docs.
	checkDocs, err := docs.Read()
	if err != nil {
		return fmt.Errorf("cannot read yaml file: %w", err)
	}

	var requiredRequestTypes []checker.RequestType
	if o.Local != "" {
		requiredRequestTypes = append(requiredRequestTypes, checker.FileBased)
	}
	if !strings.EqualFold(o.Commit, clients.HeadSHA) {
		requiredRequestTypes = append(requiredRequestTypes, checker.CommitBased)
	}
	enabledChecks, err := policy.GetEnabled(pol, o.Checks(), requiredRequestTypes)
	if err != nil {
		return fmt.Errorf("GetEnabled: %w", err)
	}

	// Define output to console or file
	var output io.Writer = os.Stdout
	if o.Output != "" {
		outputF, err := os.Create(o.Output)
		if err != nil {
			return fmt.Errorf("unable to create output file: %w", err)
		}
		defer outputF.Close()
		output = outputF
	}

	if o.Format == options.FormatDefault {
		for checkName := range enabledChecks {
			fmt.Fprintf(os.Stderr, "Starting [%s]\n", checkName)
		}
	}

	repoResult, err := pkg.RunScorecard(
		ctx,
		repoURI,
		o.Commit,
		o.CommitDepth,
		enabledChecks,
		repoClient,
		ossFuzzRepoClient,
		ciiClient,
		vulnsClient,
	)
	if err != nil {
		return fmt.Errorf("RunScorecard: %w", err)
	}

	repoResult.Metadata = append(repoResult.Metadata, o.Metadata...)

	// Sort them by name
	sort.Slice(repoResult.Checks, func(i, j int) bool {
		return repoResult.Checks[i].Name < repoResult.Checks[j].Name
	})

	if o.Format == options.FormatDefault {
		for checkName := range enabledChecks {
			fmt.Fprintf(os.Stderr, "Finished [%s]\n", checkName)
		}
		fmt.Println("\nRESULTS\n-------")
	}

	resultsErr := pkg.FormatResults(
		o,
		&repoResult,
		checkDocs,
		pol,
		output,
	)
	if resultsErr != nil {
		return fmt.Errorf("failed to format results: %w", resultsErr)
	}

	// intentionally placed at end to preserve outputting results, even if a check has a runtime error
	for _, result := range repoResult.Checks {
		if result.Error != nil {
			return sce.WithMessage(sce.ErrorCheckRuntime, fmt.Sprintf("%s: %v", result.Name, result.Error))
		}
	}
	return nil
}
