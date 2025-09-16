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
	"github.com/ossf/scorecard/v5/clients/gitlabrepo"
	"github.com/ossf/scorecard/v5/clients/localdir"
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
		repos, err := githubrepo.ListOrgRepos(ctx, o.Org)
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

	// Shared setup (unchanged)
	pol, err := policy.ParseFromFile(o.PolicyFile)
	if err != nil {
		return fmt.Errorf("readPolicy: %w", err)
	}

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

	// Iterate and scan each repo (unchanged)
	for _, uri := range repoURLs {
		var repo clients.Repo

		if o.Local != "" && uri == o.Local {
			repo, err = localdir.MakeLocalDirRepo(uri)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Skipping local %s: %v\n", uri, err)
				continue
			}
		} else {
			repo, err = makeRepo(uri)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Skipping %s: %v\n", uri, err)
				continue
			}
		}

		label := repoLabelFromURI(uri)

		// Start banners with repo label
		if o.Format == options.FormatDefault {
			if len(enabledProbes) > 0 {
				printProbeStart(label, enabledProbes)
			} else {
				printCheckStart(label, enabledChecks)
			}
		}

		result, err := scorecard.Run(ctx, repo, opts...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning %s: %v\n", uri, err)
			continue
		}

		result.Metadata = append(result.Metadata, o.Metadata...)

		// Stable order
		sort.Slice(result.Checks, func(i, j int) bool {
			return result.Checks[i].Name < result.Checks[j].Name
		})

		// End banners BEFORE RESULTS
		if o.Format == options.FormatDefault {
			if len(enabledProbes) > 0 {
				printProbeResults(label, enabledProbes)
			} else {
				printCheckResults(label, enabledChecks)
			}
		}

		// RESULTS block
		if err := scorecard.FormatResults(o, &result, checkDocs, pol); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to format results for %s: %v\n", uri, err)
		}

		// Surface per-check runtime errors (non-fatal)
		for _, r := range result.Checks {
			if r.Error != nil {
				fmt.Fprintf(os.Stderr, "Check %s failed for %s: %v\n", r.Name, uri, r.Error)
			}
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
	fmt.Fprintln(os.Stderr, "\nRESULTS\n-------")
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
