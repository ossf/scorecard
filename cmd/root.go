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

package cmd

import (
	"context"
	"encoding/json"
	"errors"
	goflag "flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/ossf/scorecard/checker"
	"github.com/ossf/scorecard/checks"
	"github.com/ossf/scorecard/clients/githubrepo"
	"github.com/ossf/scorecard/pkg"
	"github.com/ossf/scorecard/repos"
	"github.com/ossf/scorecard/roundtripper"
)

var (
	repo        repos.RepoURL
	checksToRun []string
	metaData    []string
	// This one has to use goflag instead of pflag because it's defined by zap.
	logLevel    = zap.LevelFlag("verbosity", zap.InfoLevel, "override the default log level")
	format      string
	npm         string
	pypi        string
	rubygems    string
	showDetails bool
	// ErrorInvalidFormatFlag indicates an invalid option was passed for the 'format' argument.
	ErrorInvalidFormatFlag = errors.New("invalid format flag")
)

const (
	formatCSV     = "csv"
	formatJSON    = "json"
	formatDefault = "default"
)

var rootCmd = &cobra.Command{
	Use: `./scorecard --repo=<repo_url> [--checks=check1,...] [--show-details]
or ./scorecard --{npm,pypi,rubgems}=<package_name> [--checks=check1,...] [--show-details]`,
	Short: "Security Scorecards",
	Long:  "A program that shows security scorecard for an open source software.",
	Run: func(cmd *cobra.Command, args []string) {
		// UPGRADEv2: to remove.
		_, v2 := os.LookupEnv("SCORECARD_V2")
		if v2 {
			fmt.Printf("*** USING v2 MIGRATION CODE ***\n\n")
		}
		cfg := zap.NewProductionConfig()
		cfg.Level.SetLevel(*logLevel)
		logger, err := cfg.Build()
		if err != nil {
			log.Fatalf("unable to construct logger: %v", err)
		}
		// nolint
		defer logger.Sync() // flushes buffer, if any
		sugar := logger.Sugar()

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

		if err := repo.ValidGitHubURL(); err != nil {
			log.Fatal(err)
		}

		enabledChecks := checker.CheckNameToFnMap{}
		if len(checksToRun) != 0 {
			for _, checkToRun := range checksToRun {
				if !enableCheck(checkToRun, &enabledChecks) {
					log.Fatalf("Invalid check: %s", checkToRun)
				}
			}
		} else {
			enabledChecks = checks.AllChecks
		}
		if format == formatDefault {
			for checkName := range enabledChecks {
				fmt.Fprintf(os.Stderr, "Starting [%s]\n", checkName)
			}
		}
		ctx := context.Background()

		rt := roundtripper.NewTransport(ctx, sugar)
		httpClient := &http.Client{
			Transport: rt,
		}
		githubClient := github.NewClient(httpClient)
		graphClient := githubv4.NewClient(httpClient)
		repoClient := githubrepo.CreateGithubRepoClient(ctx, githubClient, graphClient)
		defer repoClient.Close()

		repoResult, err := pkg.RunScorecards(ctx, repo, enabledChecks, repoClient, httpClient, githubClient, graphClient)
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
			// UPGRADEv2: to remove.
			if v2 {
				err = repoResult.AsString2(showDetails, *logLevel, os.Stdout)
			} else {
				err = repoResult.AsString(showDetails, *logLevel, os.Stdout)
			}
		case formatCSV:
			err = repoResult.AsCSV(showDetails, *logLevel, os.Stdout)
		case formatJSON:
			err = repoResult.AsJSON(showDetails, *logLevel, os.Stdout)
		default:
			err = fmt.Errorf("%w %s. allowed values are: [default, csv, json]", ErrorInvalidFormatFlag, format)
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

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Gets the GitHub repository URL for the npm package.
//nolint:noctx,goerr113
func fetchGitRepositoryFromNPM(packageName string) (string, error) {
	npmSearchURL := "https://registry.npmjs.org/-/v1/search?text=%s&size=1"
	const timeout = 10
	client := &http.Client{
		Timeout: timeout * time.Second,
	}
	resp, err := client.Get(fmt.Sprintf(npmSearchURL, packageName))
	if err != nil {
		return "", fmt.Errorf("failed to get npm package json: %w", err)
	}

	defer resp.Body.Close()
	v := &npmSearchResults{}
	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return "", fmt.Errorf("failed to parse npm package json: %w", err)
	}
	if len(v.Objects) == 0 {
		return "", fmt.Errorf("could not find source repo for npm package: %s", packageName)
	}
	return v.Objects[0].Package.Links.Repository, nil
}

// Gets the GitHub repository URL for the pypi package.
//nolint:noctx,goerr113
func fetchGitRepositoryFromPYPI(packageName string) (string, error) {
	pypiSearchURL := "https://pypi.org/pypi/%s/json"
	const timeout = 10
	client := &http.Client{
		Timeout: timeout * time.Second,
	}
	resp, err := client.Get(fmt.Sprintf(pypiSearchURL, packageName))
	if err != nil {
		return "", fmt.Errorf("failed to get pypi package json: %w", err)
	}

	defer resp.Body.Close()
	v := &pypiSearchResults{}
	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return "", fmt.Errorf("failed to parse pypi package json: %w", err)
	}
	if v.Info.ProjectUrls.Source == "" {
		return "", fmt.Errorf("could not find source repo for pypi package: %s", packageName)
	}
	return v.Info.ProjectUrls.Source, nil
}

// Gets the GitHub repository URL for the rubygems package.
//nolint:noctx,goerr113
func fetchGitRepositoryFromRubyGems(packageName string) (string, error) {
	rubyGemsSearchURL := "https://rubygems.org/api/v1/gems/%s.json"
	const timeout = 10
	client := &http.Client{
		Timeout: timeout * time.Second,
	}
	resp, err := client.Get(fmt.Sprintf(rubyGemsSearchURL, packageName))
	if err != nil {
		return "", fmt.Errorf("failed to get ruby gem json: %w", err)
	}

	defer resp.Body.Close()
	v := &rubyGemsSearchResults{}
	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return "", fmt.Errorf("failed to parse ruby gem json: %w", err)
	}
	if v.SourceCodeURI == "" {
		return "", fmt.Errorf("could not find source repo for ruby gem: %s", packageName)
	}
	return v.SourceCodeURI, nil
}

// Enables checks by name.
func enableCheck(checkName string, enabledChecks *checker.CheckNameToFnMap) bool {
	if enabledChecks != nil {
		for key, checkFn := range checks.AllChecks {
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
	rootCmd.Flags().Var(&repo, "repo", "repository to check")
	rootCmd.Flags().StringVar(
		&npm, "npm", "",
		"npm package to check, given that the npm package has a GitHub repository")
	rootCmd.Flags().StringVar(
		&pypi, "pypi", "",
		"pypi package to check, given that the pypi package has a GitHub repository")
	rootCmd.Flags().StringVar(
		&rubygems, "rubygems", "",
		"rubygems package to check, given that the rubygems package has a GitHub repository")
	rootCmd.Flags().StringVar(&format, "format", formatDefault, "output format. allowed values are [default, csv, json]")
	rootCmd.Flags().StringSliceVar(
		&metaData, "metadata", []string{}, "metadata for the project.It can be multiple separated by commas")
	rootCmd.Flags().BoolVar(&showDetails, "show-details", false, "show extra details about each check")
	checkNames := []string{}
	for checkName := range checks.AllChecks {
		checkNames = append(checkNames, checkName)
	}
	rootCmd.Flags().StringSliceVar(&checksToRun, "checks", []string{},
		fmt.Sprintf("Checks to run. Possible values are: %s", strings.Join(checkNames, ",")))
}
