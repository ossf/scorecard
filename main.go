package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/dlorenc/scorecard/checker"
	"github.com/dlorenc/scorecard/checks"
	"github.com/dlorenc/scorecard/roundtripper"
	"github.com/google/go-github/v32/github"
)

var repo = flag.String("repo", "", "url to the repo")
var checksToRun = flag.String("checks", "", "specific checks to run, instead of all")

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

	wg := sync.WaitGroup{}
	for _, check := range checks.AllChecks {
		check := check
		wg.Add(1)
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
			fmt.Println(check.Name, r.Confidence, r.Pass)
		}()
	}
	wg.Wait()

}
