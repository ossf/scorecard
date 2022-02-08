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

const (
	formatJSON    = "json"
	formatSarif   = "sarif"
	formatDefault = "default"
	formatRaw     = "raw"

	cliEnableSarif = "ENABLE_SARIF"

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

//nolint:gochecknoinits
func init() {
	rootCmd.Flags().StringVar(&flagRepo, "repo", "", "repository to check")
	rootCmd.Flags().StringVar(&flagLocal, "local", "", "local folder to check")
	rootCmd.Flags().StringVar(
		&flagLogLevel,
		"verbosity",
		sclog.DefaultLevel.String(),
		"set the log level",
	)
	rootCmd.Flags().StringVar(
		&flagNPM, "npm", "",
		"npm package to check, given that the npm package has a GitHub repository")
	rootCmd.Flags().StringVar(
		&flagPyPI, "pypi", "",
		"pypi package to check, given that the pypi package has a GitHub repository")
	rootCmd.Flags().StringVar(
		&flagRubyGems, "rubygems", "",
		"rubygems package to check, given that the rubygems package has a GitHub repository")
	rootCmd.Flags().StringSliceVar(
		&flagMetadata, "metadata", []string{}, "metadata for the project. It can be multiple separated by commas")
	rootCmd.Flags().BoolVar(&flagShowDetails, "show-details", false, "show extra details about each check")
	checkNames := []string{}
	for checkName := range getAllChecks() {
		checkNames = append(checkNames, checkName)
	}
	rootCmd.Flags().StringSliceVar(&flagChecksToRun, "checks", []string{},
		fmt.Sprintf("Checks to run. Possible values are: %s", strings.Join(checkNames, ",")))

	if isSarifEnabled() {
		rootCmd.Flags().StringVar(&flagPolicyFile, "policy", "", "policy to enforce")
		rootCmd.Flags().StringVar(&flagFormat, "format", formatDefault,
			"output format allowed values are [default, sarif, json]")
	} else {
		rootCmd.Flags().StringVar(&flagFormat, "format", formatDefault,
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
	validateCmdFlags()

	// Set `repo` from package managers.
	pkgResp, err := fetchGitRepositoryFromPackageManagers(flagNPM, flagPyPI, flagRubyGems)
	if err != nil {
		log.Panic(err)
	}
	if pkgResp.exists {
		if err := cmd.Flags().Set("repo", pkgResp.associatedRepo); err != nil {
			log.Panic(err)
		}
	}

	policy, err := readPolicy()
	if err != nil {
		log.Panicf("readPolicy: %v", err)
	}

	ctx := context.Background()
	logger := sclog.NewLogger(sclog.Level(flagLogLevel))
	repoURI, repoClient, ossFuzzRepoClient, ciiClient, vulnsClient, err := getRepoAccessors(
		ctx, flagRepo, flagLocal, logger)
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
	if flagLocal != "" {
		requiredRequestTypes = append(requiredRequestTypes, checker.FileBased)
	}
	enabledChecks, err := getEnabledChecks(policy, flagChecksToRun, requiredRequestTypes)
	if err != nil {
		log.Panic(err)
	}

	if flagFormat == formatDefault {
		for checkName := range enabledChecks {
			fmt.Fprintf(os.Stderr, "Starting [%s]\n", checkName)
		}
	}

	repoResult, err := pkg.RunScorecards(ctx, repoURI, flagFormat == formatRaw, enabledChecks, repoClient,
		ossFuzzRepoClient, ciiClient, vulnsClient)
	if err != nil {
		log.Panic(err)
	}
	repoResult.Metadata = append(repoResult.Metadata, flagMetadata...)

	// Sort them by name
	sort.Slice(repoResult.Checks, func(i, j int) bool {
		return repoResult.Checks[i].Name < repoResult.Checks[j].Name
	})

	if flagFormat == formatDefault {
		for checkName := range enabledChecks {
			fmt.Fprintf(os.Stderr, "Finished [%s]\n", checkName)
		}
		fmt.Println("\nRESULTS\n-------")
	}

	switch flagFormat {
	case formatDefault:
		err = repoResult.AsString(flagShowDetails, sclog.Level(flagLogLevel), checkDocs, os.Stdout)
	case formatSarif:
		// TODO: support config files and update checker.MaxResultScore.
		err = repoResult.AsSARIF(flagShowDetails, sclog.Level(flagLogLevel), os.Stdout, checkDocs, policy)
	case formatJSON:
		err = repoResult.AsJSON2(flagShowDetails, sclog.Level(flagLogLevel), checkDocs, os.Stdout)
	case formatRaw:
		err = repoResult.AsRawJSON(os.Stdout)
	default:
		err = sce.WithMessage(sce.ErrScorecardInternal,
			fmt.Sprintf("invalid format flag: %v. Expected [default, json]", flagFormat))
	}
	if err != nil {
		log.Panicf("Failed to output results: %v", err)
	}
}

func validateCmdFlags() {
	// Validate exactly one of `--repo`, `--npm`, `--pypi`, `--rubygems`, `--local` is enabled.
	if boolSum(flagRepo != "",
		flagNPM != "",
		flagPyPI != "",
		flagRubyGems != "",
		flagLocal != "") != 1 {
		log.Panic("Exactly one of `--repo`, `--npm`, `--pypi`, `--rubygems` or `--local` must be set")
	}

	// Validate SARIF features are flag-guarded.
	if !isSarifEnabled() {
		if flagFormat == formatSarif {
			log.Panic("sarif format not supported yet")
		}
		if flagPolicyFile != "" {
			log.Panic("policy file not supported yet")
		}
	}

	// Validate V6 features are flag-guarded.
	if !isV6Enabled() {
		if flagFormat == formatRaw {
			log.Panic("raw option not supported yet")
		}
	}

	// Validate format.
	if !validateFormat(flagFormat) {
		log.Panicf("unsupported format '%s'", flagFormat)
	}
}

func boolSum(bools ...bool) int {
	sum := 0
	for _, b := range bools {
		if b {
			sum++
		}
	}
	return sum
}

func isSarifEnabled() bool {
	// UPGRADEv4: remove.
	var sarifEnabled bool
	_, sarifEnabled = os.LookupEnv(cliEnableSarif)
	return sarifEnabled
}

func isV6Enabled() bool {
	var v6 bool
	_, v6 = os.LookupEnv("SCORECARD_V6")
	return v6
}

func validateFormat(format string) bool {
	switch format {
	case formatJSON, formatSarif, formatDefault, formatRaw:
		return true
	default:
		return false
	}
}

func readPolicy() (*spol.ScorecardPolicy, error) {
	if flagPolicyFile != "" {
		data, err := os.ReadFile(flagPolicyFile)
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

func isSupportedCheck(checkName string, requiredRequestTypes []checker.RequestType) bool {
	unsupported := checker.ListUnsupported(
		requiredRequestTypes,
		checks.AllChecks[checkName].SupportedRequestTypes)
	return len(unsupported) == 0
}

func getAllChecks() checker.CheckNameToFnMap {
	// Returns the full list of checks, given any environment variable constraints.
	possibleChecks := checks.AllChecks
	return possibleChecks
}

func getEnabledChecks(sp *spol.ScorecardPolicy, argsChecks []string,
	requiredRequestTypes []checker.RequestType) (checker.CheckNameToFnMap, error) {
	enabledChecks := checker.CheckNameToFnMap{}

	switch {
	case len(argsChecks) != 0:
		// Populate checks to run with the `--repo` CLI argument.
		for _, checkName := range argsChecks {
			if !isSupportedCheck(checkName, requiredRequestTypes) {
				return enabledChecks,
					sce.WithMessage(sce.ErrScorecardInternal,
						fmt.Sprintf("Unsupported RequestType %s by check: %s",
							fmt.Sprint(requiredRequestTypes), checkName))
			}
			if !enableCheck(checkName, &enabledChecks) {
				return enabledChecks,
					sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("invalid check: %s", checkName))
			}
		}
	case sp != nil:
		// Populate checks to run with policy file.
		for checkName := range sp.GetPolicies() {
			if !isSupportedCheck(checkName, requiredRequestTypes) {
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
			if !isSupportedCheck(checkName, requiredRequestTypes) {
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

func getRepoAccessors(ctx context.Context, repoURI, localURI string, logger *sclog.Logger) (
	clients.Repo, // repo
	clients.RepoClient, // repoClient
	clients.RepoClient, // ossFuzzClient
	clients.CIIBestPracticesClient, // ciiClient
	clients.VulnerabilitiesClient, // vulnClient
	error) {
	var githubRepo clients.Repo
	var errGitHub error
	if localURI != "" {
		localRepo, errLocal := localdir.MakeLocalDirRepo(localURI)
		return localRepo, /*repo*/
			localdir.CreateLocalDirClient(ctx, logger), /*repoClient*/
			nil, /*ossFuzzClient*/
			nil, /*ciiClient*/
			nil, /*vulnClient*/
			errLocal
	}

	githubRepo, errGitHub = githubrepo.MakeGithubRepo(repoURI)
	if errGitHub != nil {
		// nolint: wrapcheck
		return githubRepo,
			nil,
			nil,
			nil,
			nil,
			errGitHub
	}

	ossFuzzRepoClient, errOssFuzz := githubrepo.CreateOssFuzzRepoClient(ctx, logger)
	return githubRepo, /*repo*/
		githubrepo.CreateGithubRepoClient(ctx, logger), /*repoClient*/
		ossFuzzRepoClient, /*ossFuzzClient*/
		clients.DefaultCIIBestPracticesClient(), /*ciiClient*/
		clients.DefaultVulnerabilitiesClient(), /*vulnClient*/
		errOssFuzz
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
