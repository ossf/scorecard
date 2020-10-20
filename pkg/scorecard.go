package pkg

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sort"
	"sync"

	"github.com/dlorenc/scorecard/checker"
	"github.com/dlorenc/scorecard/checks"
	"github.com/dlorenc/scorecard/roundtripper"
	"github.com/google/go-github/v32/github"
	"go.uber.org/zap"
)

type Result struct {
	Cr   checker.CheckResult
	Name string
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

func RunScorecards(ctx context.Context, logger *zap.SugaredLogger, host, owner, repo string, checksToRun []string) []Result {
	// Use our custom roundtripper
	rt := roundtripper.NewTransport(ctx, logger)

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

	resultsCh := make(chan Result)
	wg := sync.WaitGroup{}
	for _, check := range checks.AllChecks {
		check := check
		if !stringInListOrEmpty(check.Name, checksToRun) {
			continue
		}
		wg.Add(1)
		fmt.Fprintf(os.Stderr, "Starting [%s]\n", check.Name)
		go func() {
			defer wg.Done()
			runner := checker.Runner{Checker: c}
			r := runner.Run(check.Fn)
			resultsCh <- Result{
				Name: check.Name,
				Cr:   r,
			}
		}()
	}
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Collect results
	results := []Result{}
	for result := range resultsCh {
		fmt.Fprintf(os.Stderr, "Finished [%s]\n", result.Name)
		results = append(results, result)
	}

	// Sort them by name
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})
	return results
}
