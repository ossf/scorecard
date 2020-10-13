package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/dlorenc/scorecard/checker"
	"github.com/dlorenc/scorecard/checks"
	"github.com/dlorenc/scorecard/roundtripper"
	"github.com/google/go-github/v32/github"
	"github.com/prometheus/common/log"
	"go.uber.org/zap"
)

var repo = flag.String("repo", "", "url to the repo")
var checksToRun = flag.String("checks", "", "specific checks to run, instead of all")

type result struct {
	cr   checker.CheckResult
	name string
}

func main() {
	flag.Parse()
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	split := strings.SplitN(*repo, "/", 3)
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
	for _, check := range checks.AllChecks {
		check := check
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
		fmt.Println(r.name, r.cr.Pass, r.cr.Confidence)
	}
}
