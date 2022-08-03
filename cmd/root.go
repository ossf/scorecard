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
	"sigs.k8s.io/release-utils/version"

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
	scorecardUse  = `./scorecard (--repo=<repo> | --local=<folder> | --{npm,pypi,rubygems}=<package_name>)
	 [--checks=check1,...] [--show-details]`
	scorecardShort = "Security Scorecards"
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
			return nil
		},
		// TODO(cmd): Consider using RunE here
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd(o)
		},
	}

	o.AddFlags(cmd)

	// Add sub-commands.
	cmd.AddCommand(dependencydiffCmd(o, options.NewDepdiff()))
	cmd.AddCommand(serveCmd(o))
	cmd.AddCommand(version.Version())
	return cmd
}

// rootCmd runs scorecard checks given a set of arguments.
func rootCmd(o *options.Options) {
	p := &packageManager{}
	// Set `repo` from package managers.
	pkgResp, err := fetchGitRepositoryFromPackageManagers(o.NPM, o.PyPI, o.RubyGems, p)
	if err != nil {
		log.Panic(err)
	}
	if pkgResp.exists {
		o.Repo = pkgResp.associatedRepo
	}
	pol, err := policy.ParseFromFile(o.PolicyFile)
	if err != nil {
		log.Panicf("readPolicy: %v", err)
	}
	ctx := context.Background()
	logger := sclog.NewLogger(sclog.ParseLevel(o.LogLevel))
	// Read docs.
	checkDocs, err := docs.Read()
	if err != nil {
		log.Panicf("cannot read yaml file: %v", err)
	}
	// Run the scorecard checks on the repo.
	doScorecardChecks(ctx, o, logger, checkDocs, pol)
}

func doScorecardChecks(ctx context.Context, o *options.Options,
	logger *sclog.Logger, checkDocs docs.Doc, pol *policy.ScorecardPolicy,
) {
	repoURI, repoClient, ossFuzzRepoClient, ciiClient, vulnsClient, err := checker.GetClients(
		ctx, o.Repo, o.Local, logger)
	if err != nil {
		log.Panic(err)
	}
	defer repoClient.Close()
	if ossFuzzRepoClient != nil {
		defer ossFuzzRepoClient.Close()
	}
	var requiredRequestTypes []checker.RequestType
	if o.Local != "" {
		requiredRequestTypes = append(requiredRequestTypes, checker.FileBased)
	}
	if !strings.EqualFold(o.Commit, clients.HeadSHA) {
		requiredRequestTypes = append(requiredRequestTypes, checker.CommitBased)
	}
	enabledChecks, err := policy.GetEnabled(pol, o.ChecksToRun, requiredRequestTypes)
	if err != nil {
		log.Panic(err)
	}
	if o.Format == options.FormatDefault {
		for checkName := range enabledChecks {
			fmt.Fprintf(os.Stderr, "Starting [%s]\n", checkName)
		}
	}
	repoResult, err := pkg.RunScorecards(
		ctx,
		repoURI,
		o.Commit,
		enabledChecks,
		repoClient,
		ossFuzzRepoClient,
		ciiClient,
		vulnsClient,
	)
	if err != nil {
		log.Panic(err)
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
	)
	if resultsErr != nil {
		log.Panicf("Failed to format results: %v", resultsErr)
	}
}
