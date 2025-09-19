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
	"github.com/ossf/scorecard/v5/clients/azuredevopsrepo"
	"github.com/ossf/scorecard/v5/clients/githubrepo"
	"github.com/ossf/scorecard/v5/clients/githubrepo/roundtripper"
	"github.com/ossf/scorecard/v5/clients/gitlabrepo"
	"github.com/ossf/scorecard/v5/clients/localdir"
	orgpkg "github.com/ossf/scorecard/v5/cmd/internal/org"
	pmc "github.com/ossf/scorecard/v5/cmd/internal/packagemanager"
	docs "github.com/ossf/scorecard/v5/docs/checks"
	sclog "github.com/ossf/scorecard/v5/log"
	"github.com/ossf/scorecard/v5/options"
	"github.com/ossf/scorecard/v5/pkg/scorecard"
	"github.com/ossf/scorecard/v5/policy"
)

const (
	scorecardLong = "A program that shows the OpenSSF scorecard for an open source software."
	scorecardUse  = `./scorecard (--repo=<repo> | --local=<folder> | --org=<organization> | ` +
		`--{npm,pypi,rubygems,nuget}=<package_name>) [--checks=check1,...] [--show-details] [--show-annotations]`
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

// Build the list of repositories to scan, honoring --repos > --org > --local > --repo/pkg-managers.
func buildRepoURLs(ctx context.Context, o *options.Options) ([]string, error) {
	// --repos has highest precedence
	if len(o.Repos) > 0 {
		var urls []string
		for _, r := range o.Repos {
			r = strings.TrimSpace(r)
			if r != "" {
				urls = append(urls, r)
			}
		}
		return urls, nil
	}

	// --org: expand to all non-archived repos
	if o.Org != "" {
		// create a transport to respect auth, rate limiting, etc.
		logger := sclog.NewLogger(sclog.DefaultLevel)
		rt := roundtripper.NewTransport(ctx, logger)
		repos, err := orgpkg.ListOrgRepos(ctx, o.Org, rt)
		if err != nil {
			return nil, fmt.Errorf("listing repositories for org %q: %w", o.Org, err)
		}
		return repos, nil
	}

	// --local: single local path
	if o.Local != "" {
		return []string{o.Local}, nil
	}

	// Package managers may override --repo
	p := &pmc.PackageManagerClient{}
	// Set `repo` from package managers.
	pkgResp, err := fetchGitRepositoryFromPackageManagers(o.NPM, o.PyPI, o.RubyGems, o.Nuget, p)
	if err != nil {
		return nil, fmt.Errorf("fetchGitRepositoryFromPackageManagers: %w", err)
	}
	if pkgResp.exists {
		o.Repo = pkgResp.associatedRepo
	}

	return []string{o.Repo}, nil
}

// rootCmd runs scorecard checks given a set of arguments.
func rootCmd(o *options.Options) error {
	ctx := context.Background()

	// Build the list of repos (only split this logic out)
	repoURLs, err := buildRepoURLs(ctx, o)
	if err != nil {
		return err
	}

	// Shared setup
	pol, err := policy.ParseFromFile(o.PolicyFile)
	if err != nil {
		return fmt.Errorf("readPolicy: %w", err)
	}

	// Read docs.
	checkDocs, err := docs.Read()
	if err != nil {
		return fmt.Errorf("cannot read yaml file: %w", err)
	}

	var requiredRequestTypes []checker.RequestType
	// if local option not set add file based
	if o.Local != "" {
		requiredRequestTypes = append(requiredRequestTypes, checker.FileBased)
	}
	// if commit option set to anything other than HEAD add commit based
	if !strings.EqualFold(o.Commit, clients.HeadSHA) {
		requiredRequestTypes = append(requiredRequestTypes, checker.CommitBased)
	}

	// this call to policy is different from the one in scorecard.Run
	// this one is concerned with a policy file, while the scorecard.Run call is
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

	opts := []scorecard.Option{
		scorecard.WithLogLevel(sclog.ParseLevel(o.LogLevel)),
		scorecard.WithCommitSHA(o.Commit),
		scorecard.WithCommitDepth(o.CommitDepth),
		scorecard.WithProbes(enabledProbes),
		scorecard.WithChecks(checks),
	}
	if strings.EqualFold(o.FileMode, options.FileModeGit) {
		opts = append(opts, scorecard.WithFileModeGit())
	}

	var allResults []*scorecard.Result
	// Iterate and scan each repo using a helper to keep rootCmd small.
	for _, uri := range repoURLs {
		res, err := processRepo(ctx, uri, o, enabledProbes, enabledChecks, checks, opts, checkDocs, pol)
		if err != nil {
			// processRepo already logged details; skip this URI.
			fmt.Fprintf(os.Stderr, "Skipping %s: %v\n", uri, err)
			continue
		}
		if o.CombinedOutput && res != nil {
			allResults = append(allResults, res)
		}
	}

	// If combined output requested, render one combined table appended after
	// all per-repo outputs.
	if o.CombinedOutput && len(allResults) > 0 {
		fmt.Fprintln(os.Stdout, "\nCOMBINED RESULTS\n----------------")
		if err := scorecard.FormatCombinedResultsAll(os.Stdout, o, allResults, checkDocs, pol); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to format combined results: %v\n", err)
		}
	}

	return nil
}

// repoLabelFromURI returns "owner/repo" for supported inputs only.
// Supported formats:
//   - owner/repo
//   - github.com/owner/repo
//   - https://github.com/owner/repo   (http also accepted)
//   - gitlab.com/owner/repo
//   - https://gitlab.com/owner/repo   (http also accepted)
func repoLabelFromURI(uri string) string {
	s := strings.TrimSpace(uri)
	if s == "" {
		return uri
	}

	// Strip optional scheme.
	if strings.HasPrefix(s, "https://") {
		s = strings.TrimPrefix(s, "https://")
	} else if strings.HasPrefix(s, "http://") {
		s = strings.TrimPrefix(s, "http://")
	}

	// Strip optional host.
	if strings.HasPrefix(s, "github.com/") {
		s = strings.TrimPrefix(s, "github.com/")
	} else if strings.HasPrefix(s, "gitlab.com/") {
		s = strings.TrimPrefix(s, "gitlab.com/")
	}

	// Expect owner/repo (ignore any extra path segments).
	parts := strings.Split(s, "/")
	if len(parts) >= 2 && parts[0] != "" && parts[1] != "" && !strings.Contains(parts[0], ".") {
		return parts[0] + "/" + parts[1]
	}

	// Not a supported format; return as-is.
	return uri
}

func printProbeStart(repo string, enabledProbes []string) {
	for _, probeName := range enabledProbes {
		fmt.Fprintf(os.Stderr, "Starting (%s) probe [%s]\n", repo, probeName)
	}
}

func printCheckStart(repo string, enabledChecks checker.CheckNameToFnMap) {
	for checkName := range enabledChecks {
		fmt.Fprintf(os.Stderr, "Starting (%s) [%s]\n", repo, checkName)
	}
}

func printProbeResults(repo string, enabledProbes []string) {
	for _, probeName := range enabledProbes {
		fmt.Fprintf(os.Stderr, "Finished (%s) probe %s\n", repo, probeName)
	}
}

func printCheckResults(repo string, enabledChecks checker.CheckNameToFnMap) {
	for checkName := range enabledChecks {
		fmt.Fprintf(os.Stderr, "Finished (%s) [%s]\n", repo, checkName)
	}
}

// makeRepo helps turn a URI into the appropriate clients.Repo.
// currently this is a decision between GitHub, GitLab, and Azure DevOps,
// but may expand in the future.
func makeRepo(uri string) (clients.Repo, error) {
	var repo clients.Repo
	var errGitHub, errGitLab, errAzureDevOps error
	var compositeErr error

	repo, errGitHub = githubrepo.MakeGithubRepo(uri)
	if errGitHub == nil {
		return repo, nil
	}
	compositeErr = errors.Join(compositeErr, errGitHub)

	repo, errGitLab = gitlabrepo.MakeGitlabRepo(uri)
	if errGitLab == nil {
		return repo, nil
	}
	compositeErr = errors.Join(compositeErr, errGitLab)

	_, experimental := os.LookupEnv("SCORECARD_EXPERIMENTAL")
	if experimental {
		repo, errAzureDevOps = azuredevopsrepo.MakeAzureDevOpsRepo(uri)
		if errAzureDevOps == nil {
			return repo, nil
		}
		compositeErr = errors.Join(compositeErr, errAzureDevOps)
	}

	return nil, fmt.Errorf("unable to parse as github, gitlab, or azuredevops: %w", compositeErr)
}

// processRepo performs the scanning and formatting for a single repo URI.
// It returns the Result when successful (or when combined output is requested),
// or an error describing why the URI should be skipped.
func processRepo(
	ctx context.Context,
	uri string,
	o *options.Options,
	enabledProbes []string,
	enabledChecks checker.CheckNameToFnMap,
	checksList []string,
	opts []scorecard.Option,
	checkDocs docs.Doc,
	pol *policy.ScorecardPolicy,
) (*scorecard.Result, error) {
	var repo clients.Repo
	var err error

	if o.Local != "" && uri == o.Local {
		repo, err = localdir.MakeLocalDirRepo(uri)
		if err != nil {
			return nil, fmt.Errorf("localdir: %w", err)
		}
	} else {
		repo, err = makeRepo(uri)
		if err != nil {
			return nil, err
		}
	}

	label := repoLabelFromURI(uri)

	// Start banners with repo label (always show banners even in combined-only mode)
	if o.Format == options.FormatDefault {
		if len(enabledProbes) > 0 {
			printProbeStart(label, enabledProbes)
		} else {
			printCheckStart(label, enabledChecks)
		}
	}

	result, err := scorecard.Run(ctx, repo, opts...)
	if err != nil {
		return nil, fmt.Errorf("run: %w", err)
	}

	result.Metadata = append(result.Metadata, o.Metadata...)

	// Stable order
	sort.Slice(result.Checks, func(i, j int) bool {
		return result.Checks[i].Name < result.Checks[j].Name
	})

	// End banners BEFORE RESULTS (always show banners even in combined-only mode)
	if o.Format == options.FormatDefault {
		if len(enabledProbes) > 0 {
			printProbeResults(label, enabledProbes)
		} else {
			printCheckResults(label, enabledChecks)
			// Only print the RESULTS header when not in combined-only mode.
			if !o.CombinedOutput {
				fmt.Fprintln(os.Stderr, "\nRESULTS\n-------")
			}
		}
	}

	// RESULTS block: render per-repo result only when not in combined-only mode.
	if !o.CombinedOutput {
		if err := scorecard.FormatResults(o, &result, checkDocs, pol); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to format results for %s: %v\n", uri, err)
		}
	}

	// Surface per-check runtime errors (non-fatal)
	for _, r := range result.Checks {
		if r.Error != nil {
			fmt.Fprintf(os.Stderr, "Check %s failed for %s: %v\n", r.Name, uri, r.Error)
		}
	}

	return &result, nil
}
