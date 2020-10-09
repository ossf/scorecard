package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/dlorenc/scorecard/checker"
	"github.com/dlorenc/scorecard/checks"
	"github.com/dlorenc/scorecard/roundtripper"
	"github.com/google/go-github/v32/github"
)

var repo = flag.String("repo", "", "url to the repo")
var checksToRun = flag.String("checks", "", "specific checks to run, instead of all")

type result struct {
	cr   checks.CheckResult
	name string
}

func main() {
	flag.Parse()

	split := strings.SplitN(*repo, "/", 3)
	host, owner, repo := split[0], split[1], split[2]

	switch host {
	case "github.com":
	default:
		log.Fatalf("unsupported host: %s", host)
	}

	ctx := context.Background()

	// Use our custom roundtripper
	rt := roundtripper.NewTransport(ctx)

	client := &http.Client{
		Transport: rt,
	}
	ghClient := github.NewClient(client)

	c := &checker.Checker{
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
		log.Printf("Starting [%s]\n", check.Name)
		go func() {
			defer wg.Done()
			var r checks.CheckResult
			for retriesRemaining := 3; retriesRemaining > 0; retriesRemaining-- {
				r = check.Fn(c)
				if r.ShouldRetry {
					log.Println(r.Error)
					continue
				}
				break
			}
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
		log.Printf("Finished [%s]\n", result.name)
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
