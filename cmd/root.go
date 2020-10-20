package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	goflag "flag"

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

		checksToRunList := []string{}
		if len(checksToRun) > 0 {
			checksToRunList = strings.Split(checksToRun, ",")
		}
		ctx := context.Background()
		results := pkg.RunScorecards(ctx, sugar, repo, checksToRunList)

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
