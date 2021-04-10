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
	"net/url"
	"strings"
	"sync"

	"github.com/google/go-github/v32/github"
	"github.com/ossf/scorecard/checker"
	"github.com/ossf/scorecard/roundtripper"
	"github.com/shurcooL/githubv4"
	"go.uber.org/zap"
)

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
	// Allow skipping scheme for ease-of-use, default to https.
	if !strings.Contains(s, "://") {
		s = "https://" + s
	}

	u, e := url.Parse(s)
	if e != nil {
		return e
	}

	const splitLen = 2
	split := strings.SplitN(strings.Trim(u.Path, "/"), "/", splitLen)
	if len(split) != splitLen {
		log.Fatalf("invalid repo flag: [%s], pass the full repository URL", s)
	}

	if len(strings.TrimSpace(split[0])) == 0 || len(strings.TrimSpace(split[1])) == 0 {
		log.Fatalf("invalid repo flag: [%s], pass the full repository URL", s)
	}

	r.Host, r.Owner, r.Repo = u.Host, split[0], split[1]

	switch r.Host {
	case "github.com":
		return nil
	default:
		return fmt.Errorf("unsupported host: %s", r.Host)
	}
}

func RunScorecards(ctx context.Context, logger *zap.SugaredLogger,
	repo RepoURL, checksToRun checker.CheckNameToFnMap) <-chan checker.CheckResult {
	resultsCh := make(chan checker.CheckResult)
	wg := sync.WaitGroup{}
	for _, checkFn := range checksToRun {
		checkFn := checkFn
		wg.Add(1)
		go func() {
			// Use our custom roundtripper
			rt := roundtripper.NewTransport(ctx, logger)

			client := &http.Client{
				Transport: rt,
			}
			ghClient := github.NewClient(client)
			graphClient := githubv4.NewClient(client)

			c := checker.CheckRequest{
				Ctx:         ctx,
				Client:      ghClient,
				HttpClient:  client,
				Owner:       repo.Owner,
				Repo:        repo.Repo,
				GraphClient: graphClient,
			}
			defer wg.Done()
			runner := checker.Runner{CheckRequest: c}
			resultsCh <- runner.Run(checkFn)
		}()
	}
	go func() {
		wg.Wait()
		close(resultsCh)
	}()
	return resultsCh
}
