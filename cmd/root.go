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
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	goflag "flag"

	"github.com/ossf/scorecard/checker"
	"github.com/ossf/scorecard/checks"
	"github.com/ossf/scorecard/pkg"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	repo        pkg.RepoURL
	checksToRun []string
	metaData    []string
	// This one has to use goflag instead of pflag because it's defined by zap.
	logLevel    = zap.LevelFlag("verbosity", zap.InfoLevel, "override the default log level")
	format      string
	npm         string
	pypi        string
	rubygems    string
	showDetails bool
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
		cfg := zap.NewProductionConfig()
		cfg.Level.SetLevel(*logLevel)
		logger, _ := cfg.Build()
		// nolint
		defer logger.Sync() // flushes buffer, if any
		sugar := logger.Sugar()

		var outputFn func([]checker.CheckResult)
		switch format {
		case formatCSV:
			outputFn = outputCSV
		case formatDefault:
			outputFn = outputDefault
		case formatJSON:
			outputFn = outputJSON
		default:
			log.Fatalf("invalid format flag %s. allowed values are: [default, csv, json]", format)
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

		enabledChecks := checker.CheckNameToFnMap{}
		if len(checksToRun) != 0 {
			for _, checkToRun := range checksToRun {
				if checkFn, ok := checks.AllChecks[checkToRun]; ok {
					enabledChecks[checkToRun] = checkFn
				}
			}
		} else {
			enabledChecks = checks.AllChecks
		}
		for checkName := range enabledChecks {
			if format == formatDefault {
				fmt.Fprintf(os.Stderr, "Starting [%s]\n", checkName)
			}
		}
		ctx := context.Background()

		resultsCh := pkg.RunScorecards(ctx, sugar, repo, enabledChecks)
		// Collect results
		results := []checker.CheckResult{}
		for result := range resultsCh {
			if format == formatDefault {
				fmt.Fprintf(os.Stderr, "Finished [%s]\n", result.Name)
			}
			results = append(results, result)
		}

		// Sort them by name
		sort.Slice(results, func(i, j int) bool {
			return results[i].Name < results[j].Name
		})

		outputFn(results)
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

type record struct {
	Repo     string
	Date     string
	Checks   []checker.CheckResult
	MetaData []string
}

func outputJSON(results []checker.CheckResult) {
	d := time.Now()
	or := record{
		Repo:     repo.String(),
		Date:     d.Format("2006-01-02"),
		MetaData: metaData,
	}

	for _, r := range results {
		tmpResult := checker.CheckResult{
			Name:        r.Name,
			Pass:        r.Pass,
			Confidence:  r.Confidence,
			Description: r.Description,
			HelpURL:     r.HelpURL,
		}
		if showDetails {
			tmpResult.Details = r.Details
		}
		or.Checks = append(or.Checks, tmpResult)
	}
	output, err := json.Marshal(or)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println(string(output))
}

func outputCSV(results []checker.CheckResult) {
	w := csv.NewWriter(os.Stdout)
	record := []string{repo.String()}
	columns := []string{"Repository"}
	for _, r := range results {
		columns = append(columns, r.Name+"-Pass", r.Name+"-Confidence")
		record = append(record, strconv.FormatBool(r.Pass), strconv.Itoa(r.Confidence))
	}
	fmt.Fprintln(os.Stderr, "CSV COLUMN NAMES")
	fmt.Fprintf(os.Stderr, "%s\n", strings.Join(columns, ","))
	if err := w.Write(record); err != nil {
		log.Panic(err)
	}
	w.Flush()
}

func outputDefault(results []checker.CheckResult) {
	fmt.Println()
	fmt.Println("RESULTS")
	fmt.Println("-------")
	for _, r := range results {
		fmt.Println(r.Name+":", displayResult(r.Pass), r.Confidence)
		if showDetails {
			for _, d := range r.Details {
				fmt.Println("    " + d)
			}
		}
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func displayResult(result bool) string {
	if result {
		return "Pass"
	} else {
		return "Fail"
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
		return "", fmt.Errorf("failed to get npm package json: %v", err)
	}

	defer resp.Body.Close()
	v := &npmSearchResults{}
	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return "", fmt.Errorf("failed to parse npm package json: %v", err)
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
		return "", fmt.Errorf("failed to get pypi package json: %v", err)
	}

	defer resp.Body.Close()
	v := &pypiSearchResults{}
	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return "", fmt.Errorf("failed to parse pypi package json: %v", err)
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
		return "", fmt.Errorf("failed to get ruby gem json: %v", err)
	}

	defer resp.Body.Close()
	v := &rubyGemsSearchResults{}
	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return "", fmt.Errorf("failed to parse ruby gem json: %v", err)
	}
	if v.SourceCodeURI == "" {
		return "", fmt.Errorf("could not find source repo for ruby gem: %s", packageName)
	}
	return v.SourceCodeURI, nil
}

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
