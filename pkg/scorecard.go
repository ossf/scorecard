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

package pkg

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/dlorenc/scorecard/checker"
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

func RunScorecards(ctx context.Context, logger *zap.SugaredLogger, repo RepoURL, checksToRun []checker.NamedCheck) <-chan (Result) {
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
	for _, check := range checksToRun {
		check := check
		wg.Add(1)
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
	return resultsCh
}
