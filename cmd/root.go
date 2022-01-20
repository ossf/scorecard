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
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo"
	"github.com/ossf/scorecard/v4/clients/localdir"
	docs "github.com/ossf/scorecard/v4/docs/checks"
	sce "github.com/ossf/scorecard/v4/errors"
	sclog "github.com/ossf/scorecard/v4/log"
	"github.com/ossf/scorecard/v4/pkg"
	spol "github.com/ossf/scorecard/v4/policy"
)

var (
	repo        string
	raw         bool
	local       string
	checksToRun []string
	metaData    []string
	logLevel    string
	format      string
	npm         string
	pypi        string
	rubygems    string
	showDetails bool
	policyFile  string
	rootCmd     = &cobra.Command{
		Use:   scorecardUse,
		Short: scorecardShort,
		Long:  scorecardLong,
		Run:   scorecardCmd,
	}
)

const (
	formatJSON    = "json"
	formatSarif   = "sarif"
	formatDefault = "default"

	// These strings must be the same as the ones used in
	// checks.yaml for the "repos" field.
	repoTypeLocal  = "local"
	repoTypeGitHub = "GitHub"

	scorecardLong = "A program that shows security scorecard for an open source software."
	scorecardUse  = `./scorecard [--repo=<repo_url>] [--local=folder] [--checks=check1,...]
	 [--show-details] or ./scorecard --{npm,pypi,rubygems}=<package_name> 
	 [--checks=check1,...] [--show-details]`
	scorecardShort = "Security Scorecards"
)

const cliEnableSarif = "ENABLE_SARIF"

//nolint:gochecknoinits
func init() {
	rootCmd.Flags().StringVar(&repo, "repo", "", "repository to check")
	rootCmd.Flags().StringVar(&local, "local", "", "local folder to check")
	rootCmd.Flags().StringVar(
		&logLevel,
		"verbosity",
		sclog.DefaultLevel.String(),
		"set the log level",
	)
	rootCmd.Flags().StringVar(
		&npm, "npm", "",
		"npm package to check, given that the npm package has a GitHub repository")
	rootCmd.Flags().StringVar(
		&pypi, "pypi", "",
		"pypi package to check, given that the pypi package has a GitHub repository")
	rootCmd.Flags().StringVar(
		&rubygems, "rubygems", "",
		"rubygems package to check, given that the rubygems package has a GitHub repository")
	rootCmd.Flags().StringSliceVar(
		&metaData, "metadata", []string{}, "metadata for the project. It can be multiple separated by commas")
	rootCmd.Flags().BoolVar(&showDetails, "show-details", false, "show extra details about each check")
	checkNames := []string{}
	for checkName := range getAllChecks() {
		checkNames = append(checkNames, checkName)
	}
	rootCmd.Flags().StringSliceVar(&checksToRun, "checks", []string{},
		fmt.Sprintf("Checks to run. Possible values are: %s", strings.Join(checkNames, ",")))

	var sarifEnabled bool
	_, sarifEnabled = os.LookupEnv(cliEnableSarif)
	if sarifEnabled {
		rootCmd.Flags().StringVar(&policyFile, "policy", "", "policy to enforce")
		rootCmd.Flags().StringVar(&format, "format", formatDefault,
			"output format allowed values are [default, sarif, json]")
	} else {
		rootCmd.Flags().StringVar(&format, "format", formatDefault,
			"output format allowed values are [default, json]")
	}

	var v6 bool
	_, v6 = os.LookupEnv("SCORECARD_V6")
	if v6 {
		rootCmd.Flags().BoolVar(&raw, "raw", false, "generate raw results")
	}
}

// Execute runs the Scorecard commandline.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// nolint: gocognit, gocyclo
func scorecardCmd(cmd *cobra.Command, args []string) {
	// UPGRADEv4: remove.
	var sarifEnabled bool
	_, sarifEnabled = os.LookupEnv(cliEnableSarif)

	if format == formatSarif && !sarifEnabled {
		log.Panic("sarif not supported yet")
	}

	if policyFile != "" && !sarifEnabled {
		log.Panic("policy not supported yet")
	}

	var v6 bool
	_, v6 = os.LookupEnv("SCORECARD_V6")
	if raw && !v6 {
		log.Panic("--raw option not supported yet")
	}

	// Validate format.
	if !validateFormat(format) {
		log.Panicf("unsupported format '%s'", format)
	}

	policy, err := readPolicy()
	if err != nil {
		log.Panicf("readPolicy: %v", err)
	}

	// Get the URI.
	uri, err := getURI(repo, local)
	if err != nil {
		log.Panic(err)
	}

	// Set `repo` from package managers.
	exists, gitRepo, err := fetchGitRepositoryFromPackageManagers(npm, pypi, rubygems)
	if err != nil {
		log.Panic(err)
	}
	if exists {
		if err := cmd.Flags().Set("repo", gitRepo); err != nil {
			log.Panic(err)
		}
	}

	// Sanity check that `repo` is set.
	if err := cmd.MarkFlagRequired("repo"); err != nil {
		log.Panic(err)
	}

	ctx := context.Background()
	logger, err := githubrepo.NewLogger(sclog.Level(logLevel))
	if err != nil {
		log.Panic(err)
	}
	// nolint: errcheck
	defer logger.Zap.Sync() // Flushes buffer, if any.

	repoURI, repoClient, ossFuzzRepoClient, ciiClient, vulnsClient, repoType, err := getRepoAccessors(ctx, uri, logger)
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

	supportedChecks, err := getSupportedChecks(repoType, checkDocs)
	if err != nil {
		log.Panicf("cannot read supported checks: %v", err)
	}

	enabledChecks, err := getEnabledChecks(policy, checksToRun, supportedChecks, repoType)
	if err != nil {
		log.Panic(err)
	}

	if format == formatDefault {
		for checkName := range enabledChecks {
			fmt.Fprintf(os.Stderr, "Starting [%s]\n", checkName)
		}
	}

	if raw && format != "json" {
		log.Panicf("only json format is supported")
	}

	repoResult, err := pkg.RunScorecards(ctx, repoURI, raw, enabledChecks, repoClient,
		ossFuzzRepoClient, ciiClient, vulnsClient)
	if err != nil {
		log.Panic(err)
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
		err = repoResult.AsString(showDetails, sclog.Level(logLevel), checkDocs, os.Stdout)
	case formatSarif:
		// TODO: support config files and update checker.MaxResultScore.
		err = repoResult.AsSARIF(showDetails, sclog.Level(logLevel), os.Stdout, checkDocs, policy)
	case formatJSON:
		if raw {
			err = repoResult.AsRawJSON(os.Stdout)
		} else {
			err = repoResult.AsJSON2(showDetails, sclog.Level(logLevel), checkDocs, os.Stdout)
		}

	default:
		err = sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("invalid format flag: %v. Expected [default, json]", format))
	}
	if err != nil {
		log.Panicf("Failed to output results: %v", err)
	}
}

func fetchGitRepositoryFromPackageManagers(npm, pypi, rubygems string) (bool, string, error) {
	if npm != "" {
		gitRepo, err := fetchGitRepositoryFromNPM(npm)
		return true, gitRepo, err
	}

	if pypi != "" {
		gitRepo, err := fetchGitRepositoryFromPYPI(pypi)
		return true, gitRepo, err
	}

	if rubygems != "" {
		gitRepo, err := fetchGitRepositoryFromRubyGems(rubygems)
		return false, gitRepo, err
	}

	return false, "", nil
}

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

func getRepoAccessors(ctx context.Context, uri string, logger *sclog.Logger) (
	repo clients.Repo,
	repoClient clients.RepoClient,
	ossFuzzRepoClient clients.RepoClient,
	ciiClient clients.CIIBestPracticesClient,
	vulnerabilityClient clients.VulnerabilitiesClient,
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
		vulnerabilityClient = clients.DefaultVulnerabilitiesClient()
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
