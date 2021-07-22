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

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	// nolint:gosec
	_ "net/http/pprof"

	"github.com/google/go-github/v32/github"
	"github.com/shurcooL/githubv4"
	"go.opencensus.io/stats/view"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ossf/scorecard/checker"
	"github.com/ossf/scorecard/checks"
	"github.com/ossf/scorecard/clients"
	"github.com/ossf/scorecard/clients/githubrepo"
	"github.com/ossf/scorecard/cron/config"
	"github.com/ossf/scorecard/cron/data"
	"github.com/ossf/scorecard/cron/monitoring"
	"github.com/ossf/scorecard/cron/pubsub"
	"github.com/ossf/scorecard/pkg"
	"github.com/ossf/scorecard/repos"
	"github.com/ossf/scorecard/roundtripper"
	"github.com/ossf/scorecard/stats"
)

var errIgnore *clients.ErrRepoUnavailable

func processRequest(ctx context.Context,
	batchRequest *data.ScorecardBatchRequest, checksToRun checker.CheckNameToFnMap, bucketURL string,
	repoClient clients.RepoClient,
	httpClient *http.Client, githubClient *github.Client, graphClient *githubv4.Client) error {
	filename := data.GetBlobFilename(
		fmt.Sprintf("shard-%05d", batchRequest.GetShardNum()),
		batchRequest.GetJobTime().AsTime())
	// Sanity check - make sure we are not re-processing an already processed request.
	exists, err := data.BlobExists(ctx, bucketURL, filename)
	if err != nil {
		return fmt.Errorf("error during BlobExists: %w", err)
	}
	if exists {
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
		result.Date = batchRequest.GetJobTime().AsTime().Format("2006-01-02")
		if err := result.AsJSON(true /*showDetails*/, zapcore.InfoLevel, &buffer); err != nil {
			return fmt.Errorf("error during result.AsJSON: %w", err)
		}
	}
	if err := data.WriteToBlobStore(ctx, bucketURL, filename, buffer.Bytes()); err != nil {
		return fmt.Errorf("error during WriteToBlobStore: %w", err)
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
	repoClient = githubrepo.CreateGithubRepoClient(ctx, githubClient)
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
		if err := processRequest(ctx, req, checksToRun, bucketURL,
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
