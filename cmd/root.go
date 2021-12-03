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
	"encoding/json"
	goflag "flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/ossf/scorecard/v3/checker"
	"github.com/ossf/scorecard/v3/checks"
	"github.com/ossf/scorecard/v3/clients"
	"github.com/ossf/scorecard/v3/clients/githubrepo"
	"github.com/ossf/scorecard/v3/clients/localdir"
	docs "github.com/ossf/scorecard/v3/docs/checks"
	sce "github.com/ossf/scorecard/v3/errors"
	"github.com/ossf/scorecard/v3/pkg"
	spol "github.com/ossf/scorecard/v3/policy"
)

var (
	repo        string
	local       string
	checksToRun []string
	metaData    []string
	// This one has to use goflag instead of pflag because it's defined by zap.
	logLevel    = zap.LevelFlag("verbosity", zap.InfoLevel, "override the default log level")
	format      string
	npm         string
	pypi        string
	rubygems    string
	showDetails bool
	policyFile  string
)

const (
	formatJSON    = "json"
	formatSarif   = "sarif"
	formatDefault = "default"
)

// These strings must be the same as the ones used in
// checks.yaml for the "repos" field.
const (
	repoTypeLocal  = "local"
	repoTypeGitHub = "GitHub"
)

const (
	scorecardLong = "A program that shows security scorecard for an open source software."
	scorecardUse  = `./scorecard [--repo=<repo_url>] [--local=folder] [--checks=check1,...]
	 [--show-details] [--policy=file] or ./scorecard --{npm,pypi,rubygems}=<package_name> 
	 [--checks=check1,...] [--show-details] [--policy=file]`
	scorecardShort = "Security Scorecards"
)

func readPolicy() (*spol.ScorecardPolicy, error) {
	if policyFile != "" {
		data, err := os.ReadFile(policyFile)
		if err != nil {
			return nil, sce.WithMessage(sce.ErrScorecardInternal,
				fmt.Sprintf("os.ReadFile: %v", err))
		}
		sp, err := spol.ParseFromYAML(data)
		if err != nil {
			return nil,
				sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("spol.ParseFromYAML: %v", err))
		}
		return sp, nil
	}
	return nil, nil
}

func checksHavePolicies(sp *spol.ScorecardPolicy, enabledChecks checker.CheckNameToFnMap) bool {
	for checkName := range enabledChecks {
		_, exists := sp.Policies[checkName]
		if !exists {
			log.Printf("check %s has no policy declared", checkName)
			return false
		}
	}
	return true
}

func getSupportedChecks(r string, checkDocs docs.Doc) ([]string, error) {
	allChecks := checks.AllChecks
	supportedChecks := []string{}
	for check := range allChecks {
		c, e := checkDocs.GetCheck(check)
		if e != nil {
			return nil, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("checkDocs.GetCheck: %v", e))
		}
		types := c.GetSupportedRepoTypes()
		for _, t := range types {
			if r == t {
				supportedChecks = append(supportedChecks, c.GetName())
			}
		}
	}
	return supportedChecks, nil
}

func isSupportedCheck(names []string, name string) bool {
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}

func getAllChecks() checker.CheckNameToFnMap {
	// Returns the full list of checks, given any environment variable constraints.
	possibleChecks := checks.AllChecks
	// TODO: Remove this to enable the DANGEROUS_WORKFLOW by default in the next release.
	if _, dangerousWorkflowCheck := os.LookupEnv("ENABLE_DANGEROUS_WORKFLOW"); !dangerousWorkflowCheck {
		delete(possibleChecks, checks.CheckDangerousWorkflow)
	}
	// TODO: Remove this to enable the LICENSE_CHECK by default in the next release.
	if _, licenseflowCheck := os.LookupEnv("ENABLE_LICENSE_CHECK"); !licenseflowCheck {
		delete(possibleChecks, checks.LicenseCheckPolicy)
	}
	return possibleChecks
}

func getEnabledChecks(sp *spol.ScorecardPolicy, argsChecks []string,
	supportedChecks []string, repoType string) (checker.CheckNameToFnMap, error) {
	enabledChecks := checker.CheckNameToFnMap{}

	switch {
	case len(argsChecks) != 0:
		// Populate checks to run with the CLI arguments.
		for _, checkName := range argsChecks {
			if !isSupportedCheck(supportedChecks, checkName) {
				return enabledChecks,
					sce.WithMessage(sce.ErrScorecardInternal,
						fmt.Sprintf("repo type %s: unsupported check: %s", repoType, checkName))
			}
			if !enableCheck(checkName, &enabledChecks) {
				return enabledChecks,
					sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("invalid check: %s", checkName))
			}
		}
	case sp != nil:
		// Populate checks to run with policy file.
		for checkName := range sp.GetPolicies() {
			if !isSupportedCheck(supportedChecks, checkName) {
				// We silently ignore the check, like we do
				// for the default case when no argsChecks
				// or policy are present.
				continue
			}

			if !enableCheck(checkName, &enabledChecks) {
				return enabledChecks,
					sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("invalid check: %s", checkName))
			}
		}
	default:
		// Enable all checks that are supported.
		for checkName := range getAllChecks() {
			if !isSupportedCheck(supportedChecks, checkName) {
				continue
			}
			if !enableCheck(checkName, &enabledChecks) {
				return enabledChecks,
					sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("invalid check: %s", checkName))
			}
		}
	}

	// If a policy was passed as argument, ensure all checks
	// to run have a corresponding policy.
	if sp != nil && !checksHavePolicies(sp, enabledChecks) {
		return enabledChecks, sce.WithMessage(sce.ErrScorecardInternal, "checks don't have policies")
	}

	return enabledChecks, nil
}

func validateFormat(format string) bool {
	switch format {
	case "json", "sarif", "default":
		return true
	default:
		return false
	}
}

func getRepoAccessors(ctx context.Context, uri string, logger *zap.Logger) (
	repo clients.Repo,
	repoClient clients.RepoClient,
	ossFuzzRepoClient clients.RepoClient,
	ciiClient clients.CIIBestPracticesClient,
	repoType string,
	err error) {
	var localRepo, githubRepo clients.Repo
	var errLocal, errGitHub error
	if localRepo, errLocal = localdir.MakeLocalDirRepo(uri); errLocal == nil {
		// Local directory.
		repoType = repoTypeLocal
		repo = localRepo
		repoClient = localdir.CreateLocalDirClient(ctx, logger)
		return
	}
	if githubRepo, errGitHub = githubrepo.MakeGithubRepo(uri); errGitHub == nil {
		// GitHub URL.
		repoType = repoTypeGitHub
		repo = githubRepo
		repoClient = githubrepo.CreateGithubRepoClient(ctx, logger)
		ciiClient = clients.DefaultCIIBestPracticesClient()
		ossFuzzRepoClient, err = githubrepo.CreateOssFuzzRepoClient(ctx, logger)
		return
	}
	err = sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("unspported URI: %s: [%v, %v]", uri, errLocal, errGitHub))
	return
}

func getURI(repo, local string) (string, error) {
	if repo != "" && local != "" {
		return "", sce.WithMessage(sce.ErrScorecardInternal,
			"--repo and --local options cannot be used together")
	}
	if local != "" {
		return fmt.Sprintf("file://%s", local), nil
	}
	return repo, nil
}

var rootCmd = &cobra.Command{
	Use:   scorecardUse,
	Short: scorecardShort,
	Long:  scorecardLong,
	Run: func(cmd *cobra.Command, args []string) {
		// UPGRADEv4: remove.
		var v4 bool
		_, v4 = os.LookupEnv("SCORECARD_V4")

		if format == formatSarif && !v4 {
			log.Fatal("sarif not supported yet")
		}

		if policyFile != "" && !v4 {
			log.Fatal("policy not supported yet")
		}

		if local != "" && !v4 {
			log.Fatal("--local option not supported yet")
		}

		// Validate format.
		if !validateFormat(format) {
			log.Fatalf("unsupported format '%s'", format)
		}

		policy, err := readPolicy()
		if err != nil {
			log.Fatalf("readPolicy: %v", err)
		}

		// Get the URI.
		uri, err := getURI(repo, local)
		if err != nil {
			log.Fatal(err)
		}
		if npm != "" {
			if git, err := fetchGitRepositoryFromNPM(npm); err != nil {
				log.Fatal(err)
			} else {
				if err := cmd.Flags().Set("repo", git); err != nil {
					log.Fatal(err)
				}
			}
		} else if pypi != "" {
			if git, err := fetchGitRepositoryFromPYPI(pypi); err != nil {
				log.Fatal(err)
			} else {
				if err := cmd.Flags().Set("repo", git); err != nil {
					log.Fatal(err)
				}
			}
		} else if rubygems != "" {
			if git, err := fetchGitRepositoryFromRubyGems(rubygems); err != nil {
				log.Fatal(err)
			} else {
				if err := cmd.Flags().Set("repo", git); err != nil {
					log.Fatal(err)
				}
			}
		} else {
			if err := cmd.MarkFlagRequired("repo"); err != nil {
				log.Fatal(err)
			}
		}

		ctx := context.Background()
		logger, err := githubrepo.NewLogger(*logLevel)
		if err != nil {
			log.Fatal(err)
		}
		// nolint
		defer logger.Sync() // Flushes buffer, if any.

		repoURI, repoClient, ossFuzzRepoClient, ciiClient, repoType, err := getRepoAccessors(ctx, uri, logger)
		if err != nil {
			log.Fatal(err)
		}
		defer repoClient.Close()
		if ossFuzzRepoClient != nil {
			defer ossFuzzRepoClient.Close()
		}

		// Read docs.
		checkDocs, err := docs.Read()
		if err != nil {
			log.Fatalf("cannot read yaml file: %v", err)
		}

		supportedChecks, err := getSupportedChecks(repoType, checkDocs)
		if err != nil {
			log.Fatalf("cannot read supported checks: %v", err)
		}

		enabledChecks, err := getEnabledChecks(policy, checksToRun, supportedChecks, repoType)
		if err != nil {
			panic(err)
		}

		if format == formatDefault {
			for checkName := range enabledChecks {
				fmt.Fprintf(os.Stderr, "Starting [%s]\n", checkName)
			}
		}
		repoResult, err := pkg.RunScorecards(ctx, repoURI, enabledChecks, repoClient, ossFuzzRepoClient, ciiClient)
		if err != nil {
			log.Fatal(err)
		}
		repoResult.Metadata = append(repoResult.Metadata, metaData...)

		// Sort them by name
		sort.Slice(repoResult.Checks, func(i, j int) bool {
			return repoResult.Checks[i].Name < repoResult.Checks[j].Name
		})

		if format == formatDefault {
			for checkName := range enabledChecks {
				fmt.Fprintf(os.Stderr, "Finished [%s]\n", checkName)
			}
			fmt.Println("\nRESULTS\n-------")
		}

		switch format {
		case formatDefault:
			err = repoResult.AsString(showDetails, *logLevel, checkDocs, os.Stdout)
		case formatSarif:
			// TODO: support config files and update checker.MaxResultScore.
			err = repoResult.AsSARIF(showDetails, *logLevel, os.Stdout, checkDocs, policy,
				policyFile)
		case formatJSON:
			// UPGRADEv2: rename.
			err = repoResult.AsJSON2(showDetails, *logLevel, checkDocs, os.Stdout)
		default:
			err = sce.WithMessage(sce.ErrScorecardInternal,
				fmt.Sprintf("invalid format flag: %v. Expected [default, json]", format))
		}
		if err != nil {
			log.Fatalf("Failed to output results: %v", err)
		}
	},
}

type npmSearchResults struct {
	Objects []struct {
		Package struct {
			Links struct {
				Repository string `json:"repository"`
			} `json:"links"`
		} `json:"package"`
	} `json:"objects"`
}

type pypiSearchResults struct {
	Info struct {
		ProjectUrls struct {
			Source string `json:"Source"`
		} `json:"project_urls"`
	} `json:"info"`
}

type rubyGemsSearchResults struct {
	SourceCodeURI string `json:"source_code_uri"`
}

// Execute runs the Scorecard commandline.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Gets the GitHub repository URL for the npm package.
//nolint:noctx
func fetchGitRepositoryFromNPM(packageName string) (string, error) {
	npmSearchURL := "https://registry.npmjs.org/-/v1/search?text=%s&size=1"
	const timeout = 10
	client := &http.Client{
		Timeout: timeout * time.Second,
	}
	resp, err := client.Get(fmt.Sprintf(npmSearchURL, packageName))
	if err != nil {
		return "", sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("failed to get npm package json: %v", err))
	}

	defer resp.Body.Close()
	v := &npmSearchResults{}
	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return "", sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("failed to parse npm package json: %v", err))
	}
	if len(v.Objects) == 0 {
		return "", sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("could not find source repo for npm package: %s", packageName))
	}
	return v.Objects[0].Package.Links.Repository, nil
}

// Gets the GitHub repository URL for the pypi package.
//nolint:noctx
func fetchGitRepositoryFromPYPI(packageName string) (string, error) {
	pypiSearchURL := "https://pypi.org/pypi/%s/json"
	const timeout = 10
	client := &http.Client{
		Timeout: timeout * time.Second,
	}
	resp, err := client.Get(fmt.Sprintf(pypiSearchURL, packageName))
	if err != nil {
		return "", sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("failed to get pypi package json: %v", err))
	}

	defer resp.Body.Close()
	v := &pypiSearchResults{}
	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return "", sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("failed to parse pypi package json: %v", err))
	}
	if v.Info.ProjectUrls.Source == "" {
		return "", sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("could not find source repo for pypi package: %s", packageName))
	}
	return v.Info.ProjectUrls.Source, nil
}

// Gets the GitHub repository URL for the rubygems package.
//nolint:noctx
func fetchGitRepositoryFromRubyGems(packageName string) (string, error) {
	rubyGemsSearchURL := "https://rubygems.org/api/v1/gems/%s.json"
	const timeout = 10
	client := &http.Client{
		Timeout: timeout * time.Second,
	}
	resp, err := client.Get(fmt.Sprintf(rubyGemsSearchURL, packageName))
	if err != nil {
		return "", sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("failed to get ruby gem json: %v", err))
	}

	defer resp.Body.Close()
	v := &rubyGemsSearchResults{}
	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return "", sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("failed to parse ruby gem json: %v", err))
	}
	if v.SourceCodeURI == "" {
		return "", sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("could not find source repo for ruby gem: %v", err))
	}
	return v.SourceCodeURI, nil
}

// Enables checks by name.
func enableCheck(checkName string, enabledChecks *checker.CheckNameToFnMap) bool {
	if enabledChecks != nil {
		for key, checkFn := range getAllChecks() {
			if strings.EqualFold(key, checkName) {
				(*enabledChecks)[key] = checkFn
				return true
			}
		}
	}
	return false
}

//nolint:gochecknoinits
func init() {
	// Add the zap flag manually
	rootCmd.PersistentFlags().AddGoFlagSet(goflag.CommandLine)
	rootCmd.Flags().StringVar(&repo, "repo", "", "repository to check")
	rootCmd.Flags().StringVar(&local, "local", "", "local folder to check")
	rootCmd.Flags().StringVar(
		&npm, "npm", "",
		"npm package to check, given that the npm package has a GitHub repository")
	rootCmd.Flags().StringVar(
		&pypi, "pypi", "",
		"pypi package to check, given that the pypi package has a GitHub repository")
	rootCmd.Flags().StringVar(
		&rubygems, "rubygems", "",
		"rubygems package to check, given that the rubygems package has a GitHub repository")
	rootCmd.Flags().StringVar(&format, "format", formatDefault,
		"output format. allowed values are [default, sarif, json]")
	rootCmd.Flags().StringSliceVar(
		&metaData, "metadata", []string{}, "metadata for the project. It can be multiple separated by commas")
	rootCmd.Flags().BoolVar(&showDetails, "show-details", false, "show extra details about each check")
	checkNames := []string{}
	for checkName := range getAllChecks() {
		checkNames = append(checkNames, checkName)
	}
	rootCmd.Flags().StringSliceVar(&checksToRun, "checks", []string{},
		fmt.Sprintf("Checks to run. Possible values are: %s", strings.Join(checkNames, ",")))
	rootCmd.Flags().StringVar(&policyFile, "policy", "", "policy to enforce")
}
