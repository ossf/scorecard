package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"

	goflag "flag"

	"github.com/dlorenc/scorecard/checker"
	"github.com/dlorenc/scorecard/checks"
	"github.com/dlorenc/scorecard/roundtripper"
	"github.com/google/go-github/v32/github"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	repo        string
	checksToRun string
	// This one has to use goflag instead of pflag because it's defined by zap
	logLevel = zap.LevelFlag("verbosity", zap.InfoLevel, "override the default log level")
)

var rootCmd = &cobra.Command{
	Use:   "scorecard",
	Short: "Security scorecards!",
	Long:  `A scorecard program!`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := zap.NewProductionConfig()
		cfg.Level.SetLevel(*logLevel)
		logger, _ := cfg.Build()
		defer logger.Sync() // flushes buffer, if any
		sugar := logger.Sugar()

		split := strings.SplitN(repo, "/", 3)
		if len(split) != 3 {
			log.Fatalf("invalid repo flag: %s, pass the full repository URL", repo)
		}
		host, owner, repo := split[0], split[1], split[2]

		switch host {
		case "github.com":
		default:
			log.Fatalf("unsupported host: %s", host)
		}

		ctx := context.Background()

		// Use our custom roundtripper
		rt := roundtripper.NewTransport(ctx, sugar)

		client := &http.Client{
			Transport: rt,
		}
		ghClient := github.NewClient(client)

		c := checker.Checker{
			Ctx:        ctx,
			Client:     ghClient,
			HttpClient: client,
			Owner:      owner,
			Repo:       repo,
		}

		resultsCh := make(chan result)
		wg := sync.WaitGroup{}
		checksToRunList := []string{}
		if len(checksToRun) > 0 {
			checksToRunList = strings.Split(checksToRun, ",")
		}
		for _, check := range checks.AllChecks {
			check := check
			if !stringInListOrEmpty(check.Name, checksToRunList) {
				continue
			}
			wg.Add(1)
			fmt.Fprintf(os.Stderr, "Starting [%s]\n", check.Name)
			go func() {
				defer wg.Done()
				runner := checker.Runner{Checker: c}
				r := runner.Run(check.Fn)
				resultsCh <- result{
					name: check.Name,
					cr:   r,
				}
			}()
		}
		go func() {
			wg.Wait()
			close(resultsCh)
		}()

		// Collect results
		results := []result{}
		for result := range resultsCh {
			fmt.Fprintf(os.Stderr, "Finished [%s]\n", result.name)
			results = append(results, result)
		}

		// Sort them by name
		sort.Slice(results, func(i, j int) bool {
			return results[i].name < results[j].name
		})

		fmt.Println()
		fmt.Println("RESULTS")
		fmt.Println("-------")
		for _, r := range results {
			fmt.Println(r.name, displayResult(r.cr.Pass), r.cr.Confidence)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type result struct {
	cr   checker.CheckResult
	name string
}

func stringInListOrEmpty(s string, list []string) bool {
	if len(list) == 0 {
		return true
	}
	for _, le := range list {
		if le == s {
			return true
		}
	}
	return false
}

func displayResult(result bool) string {
	if result {
		return "pass"
	} else {
		return "fail"
	}
}

func init() {
	// Add the zap flag manually
	rootCmd.PersistentFlags().AddGoFlagSet(goflag.CommandLine)
	rootCmd.PersistentFlags().StringVar(&repo, "repo", "", "repository to check")
	rootCmd.PersistentFlags().StringVar(&checksToRun, "checks", "", "specific checks to run, instead of all")
}
