package pkg

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
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

type RepoURL struct {
	Host, Owner, Repo string
}

func (r *RepoURL) String() string {
	return fmt.Sprintf("%s/%s/%s", r.Host, r.Owner, r.Repo)
}

func (r *RepoURL) Type() string {
	return "repo"
}

func (r *RepoURL) Set(s string) error {
	rgx, _ := regexp.Compile("^https?://")
	s = rgx.ReplaceAllString(s, "")
	split := strings.SplitN(s, "/", 3)
	if len(split) != 3 {
		log.Fatalf("invalid repo flag: [%s], pass the full repository URL", s)
	}
	r.Host, r.Owner, r.Repo = split[0], split[1], split[2]

	switch r.Host {
	case "github.com":
		return nil
	default:
		return fmt.Errorf("unsupported host: %s", r.Host)
	}
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

func RunScorecards(ctx context.Context, logger *zap.SugaredLogger, repo RepoURL, checksToRun []string) []Result {
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
		Owner:      repo.Owner,
		Repo:       repo.Repo,
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
