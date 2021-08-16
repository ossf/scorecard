// Copyright 2021 Security Scorecard Authors
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

// Package main implements cron worker job.
package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"

	// nolint:gosec
	_ "net/http/pprof"

	"github.com/google/go-github/v38/github"
	"github.com/shurcooL/githubv4"
	"go.opencensus.io/stats/view"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ossf/scorecard/v2/checker"
	"github.com/ossf/scorecard/v2/checks"
	"github.com/ossf/scorecard/v2/clients"
	"github.com/ossf/scorecard/v2/clients/githubrepo"
	"github.com/ossf/scorecard/v2/cron/config"
	"github.com/ossf/scorecard/v2/cron/data"
	"github.com/ossf/scorecard/v2/cron/monitoring"
	"github.com/ossf/scorecard/v2/cron/pubsub"
	sce "github.com/ossf/scorecard/v2/errors"
	"github.com/ossf/scorecard/v2/pkg"
	"github.com/ossf/scorecard/v2/repos"
	"github.com/ossf/scorecard/v2/roundtripper"
	"github.com/ossf/scorecard/v2/stats"
)

var (
	ignoreRuntimeErrors = flag.Bool("ignoreRuntimeErrors", false, "if set to true any runtime errors will be ignored")
	errIgnore           *clients.ErrRepoUnavailable
)

func processRequest(ctx context.Context,
	batchRequest *data.ScorecardBatchRequest, checksToRun checker.CheckNameToFnMap,
	bucketURL, bucketURL2 string,
	repoClient clients.RepoClient,
	httpClient *http.Client, githubClient *github.Client, graphClient *githubv4.Client) error {
	filename := data.GetBlobFilename(
		fmt.Sprintf("shard-%05d", batchRequest.GetShardNum()),
		batchRequest.GetJobTime().AsTime())
	// Sanity check - make sure we are not re-processing an already processed request.
	exists1, err := data.BlobExists(ctx, bucketURL, filename)
	if err != nil {
		return fmt.Errorf("error during BlobExists: %w", err)
	}

	exists2, err := data.BlobExists(ctx, bucketURL2, filename)
	if err != nil {
		return fmt.Errorf("error during BlobExists: %w", err)
	}
	if exists1 && exists2 {
		log.Printf("Already processed shard %s. Nothing to do.", filename)
		// We have already processed this request, nothing to do.
		return nil
	}

	repoURLs := make([]repos.RepoURL, 0, len(batchRequest.GetRepos()))
	for _, repo := range batchRequest.GetRepos() {
		repoURL := repos.RepoURL{}
		if err := repoURL.Set(repo); err != nil {
			return fmt.Errorf("error setting RepoURL: %w", err)
		}
		if err := repoURL.ValidGitHubURL(); err != nil {
			return fmt.Errorf("url is not a valid GitHub URL: %w", err)
		}
		repoURLs = append(repoURLs, repoURL)
	}

	var buffer bytes.Buffer
	var buffer2 bytes.Buffer
	// TODO: run Scorecard for each repo in a separate thread.
	for _, repoURL := range repoURLs {
		log.Printf("Running Scorecard for repo: %s", repoURL.URL())
		result, err := pkg.RunScorecards(ctx, repoURL, checksToRun, repoClient, httpClient, githubClient, graphClient)
		if errors.As(err, &errIgnore) {
			// Not accessible repo - continue.
			continue
		}
		if err != nil {
			return fmt.Errorf("error during RunScorecards: %w", err)
		}
		for checkIndex := range result.Checks {
			check := &result.Checks[checkIndex]
			if !errors.As(check.Error2, &sce.ErrScorecardInternal) {
				continue
			}
			errorMsg := fmt.Sprintf("check %s has a runtime error: %v", check.Name, check.Error2)
			if !(*ignoreRuntimeErrors) {
				// nolint: goerr113
				return errors.New(errorMsg)
			}
			log.Print(errorMsg)
		}
		result.Date = batchRequest.GetJobTime().AsTime().Format("2006-01-02")
		if err := result.AsJSON(true /*showDetails*/, zapcore.InfoLevel, &buffer); err != nil {
			return fmt.Errorf("error during result.AsJSON: %w", err)
		}

		if err := result.AsJSON2(true /*showDetails*/, zapcore.InfoLevel, &buffer2); err != nil {
			return fmt.Errorf("error during result.AsJSON2: %w", err)
		}
	}
	if err := data.WriteToBlobStore(ctx, bucketURL, filename, buffer.Bytes()); err != nil {
		return fmt.Errorf("error during WriteToBlobStore: %w", err)
	}

	if err := data.WriteToBlobStore(ctx, bucketURL2, filename, buffer2.Bytes()); err != nil {
		return fmt.Errorf("error during WriteToBlobStore2: %w", err)
	}

	log.Printf("Write to shard file successful: %s", filename)

	return nil
}

func createNetClients(ctx context.Context) (
	repoClient clients.RepoClient,
	httpClient *http.Client,
	githubClient *github.Client, graphClient *githubv4.Client, logger *zap.Logger) {
	cfg := zap.NewProductionConfig()
	cfg.Level.SetLevel(zap.InfoLevel)
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	sugar := logger.Sugar()
	// Use our custom roundtripper
	rt := roundtripper.NewTransport(ctx, sugar)
	httpClient = &http.Client{
		Transport: rt,
	}
	githubClient = github.NewClient(httpClient)
	graphClient = githubv4.NewClient(httpClient)
	repoClient = githubrepo.CreateGithubRepoClient(ctx, githubClient, graphClient)
	return
}

func startMetricsExporter() (monitoring.Exporter, error) {
	exporter, err := monitoring.GetExporter()
	if err != nil {
		return nil, fmt.Errorf("error during monitoring.GetExporter: %w", err)
	}
	if err := exporter.StartMetricsExporter(); err != nil {
		return nil, fmt.Errorf("error in StartMetricsExporter: %w", err)
	}

	if err := view.Register(
		&stats.CheckRuntime,
		&stats.CheckErrorCount,
		&stats.RepoRuntime,
		&stats.OutgoingHTTPRequests,
		&githubrepo.GithubTokens); err != nil {
		return nil, fmt.Errorf("error during view.Register: %w", err)
	}
	return exporter, nil
}

func main() {
	ctx := context.Background()

	flag.Parse()

	subscriptionURL, err := config.GetRequestSubscriptionURL()
	if err != nil {
		panic(err)
	}
	subscriber, err := pubsub.CreateSubscriber(ctx, subscriptionURL)
	if err != nil {
		panic(err)
	}

	bucketURL, err := config.GetResultDataBucketURL()
	if err != nil {
		panic(err)
	}

	bucketURL2, err := config.GetResultDataBucketURLV2()
	if err != nil {
		panic(err)
	}

	repoClient, httpClient, githubClient, graphClient, logger := createNetClients(ctx)
	defer repoClient.Close()

	exporter, err := startMetricsExporter()
	if err != nil {
		panic(err)
	}
	defer exporter.StopMetricsExporter()

	// Exposed for monitoring runtime profiles
	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	checksToRun := checks.AllChecks
	// nolint
	// FIXME :- deleting branch-protection
	// The branch protection check needs an admin access to the repository.
	// All of the checks from cron would fail and uses another call to the API.
	// This will reduce usage of the API.
	delete(checksToRun, checks.CheckBranchProtection)
	for {
		req, err := subscriber.SynchronousPull()
		if err != nil {
			panic(err)
		}
		log.Print("Received message from subscription")
		if req == nil {
			log.Print("subscription returned nil message during Receive, exiting")
			break
		}
		if err := processRequest(ctx, req, checksToRun, bucketURL, bucketURL2,
			repoClient, httpClient, githubClient, graphClient); err != nil {
			log.Printf("error processing request: %v", err)
			// Nack the message so that another worker can retry.
			subscriber.Nack()
			continue
		}
		// nolint: errcheck // flushes buffer
		logger.Sync()
		exporter.Flush()
		subscriber.Ack()
	}
	err = subscriber.Close()
	if err != nil {
		panic(err)
	}
}
