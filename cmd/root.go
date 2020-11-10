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
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

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
	// This one has to use goflag instead of pflag because it's defined by zap
	logLevel    = zap.LevelFlag("verbosity", zap.InfoLevel, "override the default log level")
	format      string
	showDetails bool
)

const (
	formatCSV     = "csv"
	formatDefault = "default"
)

var rootCmd = &cobra.Command{
	Use:   "./scorecard --repo=<repo_url> [--checks=check1,...] [--show-details]",
	Short: "Security Scorecards",
	Long:  "A program that shows security scorecard for an open source software.",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := zap.NewProductionConfig()
		cfg.Level.SetLevel(*logLevel)
		logger, _ := cfg.Build()
		defer logger.Sync() // flushes buffer, if any
		sugar := logger.Sugar()

		var outputFn func([]pkg.Result)
		switch format {
		case formatCSV:
			outputFn = outputCSV
		case formatDefault:
			outputFn = outputCSV
		default:
			log.Fatalf("invalid format flag %s. allowed values are: [default, csv]", format)
		}

		enabledChecks := []checker.NamedCheck{}
		if len(checksToRun) != 0 {
			checkNames := map[string]struct{}{}
			for _, s := range checksToRun {
				checkNames[s] = struct{}{}
			}
			for _, c := range checks.AllChecks {
				if _, ok := checkNames[c.Name]; ok {
					enabledChecks = append(enabledChecks, c)
				}
			}
		} else {
			enabledChecks = checks.AllChecks
		}
		for _, c := range enabledChecks {
			fmt.Fprintf(os.Stderr, "Starting [%s]\n", c.Name)
		}
		ctx := context.Background()

		resultsCh := pkg.RunScorecards(ctx, sugar, repo, enabledChecks)
		// Collect results
		results := []pkg.Result{}
		for result := range resultsCh {
			fmt.Fprintf(os.Stderr, "Finished [%s]\n", result.Name)
			results = append(results, result)
		}

		// Sort them by name
		sort.Slice(results, func(i, j int) bool {
			return results[i].Name < results[j].Name
		})

		outputFn(results)
	},
}

func outputCSV(results []pkg.Result) {
	w := csv.NewWriter(os.Stdout)
	record := []string{repo.String()}
	columns := []string{"Repository"}
	for _, r := range results {
		columns = append(columns, r.Name+"-Pass", r.Name+"-Confidence")
		record = append(record, strconv.FormatBool(r.Cr.Pass), strconv.Itoa(r.Cr.Confidence))
	}
	fmt.Fprintln(os.Stderr, "CSV COLUMN NAMES")
	fmt.Fprintf(os.Stderr, "%s\n", strings.Join(columns, ","))
	w.Write(record)
	w.Flush()
}

func outputDefault(results []pkg.Result) {
	fmt.Println()
	fmt.Println("RESULTS")
	fmt.Println("-------")
	for _, r := range results {
		fmt.Println(r.Name+":", displayResult(r.Cr.Pass), r.Cr.Confidence)
		if showDetails {
			for _, d := range r.Cr.Details {
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

func init() {
	// Add the zap flag manually
	rootCmd.PersistentFlags().AddGoFlagSet(goflag.CommandLine)
	rootCmd.Flags().Var(&repo, "repo", "repository to check")
	rootCmd.Flags().StringVar(&format, "format", formatDefault, "output format. allowed values are [default, csv]")
	rootCmd.MarkFlagRequired("repo")
	rootCmd.Flags().BoolVar(&showDetails, "show-details", false, "show extra details about each check")
	checkNames := []string{}
	for _, c := range checks.AllChecks {
		checkNames = append(checkNames, c.Name)
	}
	rootCmd.Flags().StringSliceVar(&checksToRun, "checks", []string{},
		fmt.Sprintf("Checks to run. Possible values are: %s", strings.Join(checkNames, ",")))
}
