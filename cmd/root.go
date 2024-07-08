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

// Package cmd implements Scorecard command-line.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/release-utils/version"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/clients/githubrepo"
	"github.com/ossf/scorecard/v5/clients/gitlabrepo"
	"github.com/ossf/scorecard/v5/clients/localdir"
	pmc "github.com/ossf/scorecard/v5/cmd/internal/packagemanager"
	docs "github.com/ossf/scorecard/v5/docs/checks"
	sce "github.com/ossf/scorecard/v5/errors"
	sclog "github.com/ossf/scorecard/v5/log"
	"github.com/ossf/scorecard/v5/options"
	"github.com/ossf/scorecard/v5/pkg/scorecard"
	"github.com/ossf/scorecard/v5/policy"
)

const (
	scorecardLong = "A program that shows the OpenSSF scorecard for an open source software."
	scorecardUse  = `./scorecard (--repo=<repo> | --local=<folder> | --{npm,pypi,rubygems,nuget}=<package_name>)
	 [--checks=check1,...] [--show-details] [--show-annotations]`
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
	var err error
	var repoResult scorecard.ScorecardResult

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

	var repo clients.Repo
	if o.Local != "" {
		repo, err = localdir.MakeLocalDirRepo(o.Local)
		if err != nil {
			return fmt.Errorf("making local dir: %w", err)
		}
	} else {
		repo, err = makeRepo(o.Repo)
		if err != nil {
			return fmt.Errorf("making remote repo: %w", err)
		}
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
	// this call to policy is different from the one in pkg.Run
	// this one is concerned with a policy file, while the pkg.Run call is
	// more concerned with the supported request types
	enabledChecks, err := policy.GetEnabled(pol, o.Checks(), requiredRequestTypes)
	if err != nil {
		return fmt.Errorf("GetEnabled: %w", err)
	}
	checks := make([]string, 0, len(enabledChecks))
	for c := range enabledChecks {
		checks = append(checks, c)
	}

	enabledProbes := o.Probes()
	if o.Format == options.FormatDefault {
		if len(enabledProbes) > 0 {
			printProbeStart(enabledProbes)
		} else {
			printCheckStart(enabledChecks)
		}
	}

	repoResult, err = scorecard.Run(ctx, repo,
		scorecard.WithLogLevel(sclog.ParseLevel(o.LogLevel)),
		scorecard.WithCommitSHA(o.Commit),
		scorecard.WithCommitDepth(o.CommitDepth),
		scorecard.WithProbes(enabledProbes),
		scorecard.WithChecks(checks),
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
		if len(enabledProbes) > 0 {
			printProbeResults(enabledProbes)
		} else {
			printCheckResults(enabledChecks)
		}
	}

	resultsErr := scorecard.FormatResults(
		o,
		&repoResult,
		checkDocs,
		pol,
	)
	if resultsErr != nil {
		return fmt.Errorf("failed to format results: %w", resultsErr)
	}

	// intentionally placed at end to preserve outputting results, even if a check has a runtime error
	for _, result := range repoResult.Checks {
		if result.Error != nil {
			return sce.WithMessage(sce.ErrCheckRuntime, fmt.Sprintf("%s: %v", result.Name, result.Error))
		}
	}
	return nil
}

func printProbeStart(enabledProbes []string) {
	for _, probeName := range enabledProbes {
		fmt.Fprintf(os.Stderr, "Starting probe [%s]\n", probeName)
	}
}

func printCheckStart(enabledChecks checker.CheckNameToFnMap) {
	for checkName := range enabledChecks {
		fmt.Fprintf(os.Stderr, "Starting [%s]\n", checkName)
	}
}

func printProbeResults(enabledProbes []string) {
	for _, probeName := range enabledProbes {
		fmt.Fprintf(os.Stderr, "Finished probe %s\n", probeName)
	}
}

func printCheckResults(enabledChecks checker.CheckNameToFnMap) {
	for checkName := range enabledChecks {
		fmt.Fprintf(os.Stderr, "Finished [%s]\n", checkName)
	}
	fmt.Fprintln(os.Stderr, "\nRESULTS\n-------")
}

// makeRepo helps turn a URI into the appropriate clients.Repo.
// currently this is a decision between GitHub and GitLab,
// but may expand in the future.
func makeRepo(uri string) (clients.Repo, error) {
	var repo clients.Repo
	var errGitHub, errGitLab error
	if repo, errGitHub = githubrepo.MakeGithubRepo(uri); errGitHub != nil {
		repo, errGitLab = gitlabrepo.MakeGitlabRepo(uri)
		if errGitLab != nil {
			return nil, fmt.Errorf("unable to parse as github or gitlab: %w", errors.Join(errGitHub, errGitLab))
		}
	}
	return repo, nil
}
