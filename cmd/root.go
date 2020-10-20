package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	goflag "flag"

	"github.com/dlorenc/scorecard/checker"
	"github.com/dlorenc/scorecard/checks"
	"github.com/dlorenc/scorecard/pkg"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	repo        pkg.RepoURL
	checksToRun string
	// This one has to use goflag instead of pflag because it's defined by zap
	logLevel = zap.LevelFlag("verbosity", zap.InfoLevel, "override the default log level")
)

var rootCmd = &cobra.Command{
	Use:   "./scorecard --repo=<repo_url> [--checks=check1,...]",
	Short: "Open Source Scorecards",
	Long:  "A program that shows scorecard for an open source software.",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := zap.NewProductionConfig()
		cfg.Level.SetLevel(*logLevel)
		logger, _ := cfg.Build()
		defer logger.Sync() // flushes buffer, if any
		sugar := logger.Sugar()

		enabledChecks := []checker.NamedCheck{}

		if len(checksToRun) != 0 {
			split := strings.Split(checksToRun, ",")
			checkNames := map[string]struct{}{}
			for _, s := range split {
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

		fmt.Println()
		fmt.Println("RESULTS")
		fmt.Println("-------")
		for _, r := range results {
			fmt.Println(r.Name+":", displayResult(r.Cr.Pass), r.Cr.Confidence)
		}
	},
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
	rootCmd.PersistentFlags().Var(&repo, "repo", "repository to check")
	rootCmd.MarkPersistentFlagRequired("repo")
	rootCmd.PersistentFlags().StringVar(&checksToRun, "checks", "", "specific checks to run, instead of all")
}
