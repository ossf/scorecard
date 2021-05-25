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
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/go-github/v32/github"
	"github.com/shurcooL/githubv4"
	"go.uber.org/zap"

	"github.com/ossf/scorecard/checks"
	"github.com/ossf/scorecard/cron/config"
	"github.com/ossf/scorecard/cron/data"
	"github.com/ossf/scorecard/cron/pubsub"
	"github.com/ossf/scorecard/pkg"
	"github.com/ossf/scorecard/repos"
	"github.com/ossf/scorecard/roundtripper"
)

func processRequest(ctx context.Context,
	batchRequest *data.ScorecardBatchRequest, bucketURL string,
	httpClient *http.Client, githubClient *github.Client, graphClient *githubv4.Client) error {
	checksToRun := checks.AllChecks
	// nolint
	// FIXME :- deleting branch-protection
	// The branch protection check needs an admin access to the repository.
	// All of the checks from cron would fail and uses another call to the API.
	// This will reduce usage of the API.
	delete(checksToRun, "Branch-Protection")

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
		result := pkg.RunScorecards(ctx, repoURL, checksToRun, httpClient, githubClient, graphClient)
		result.Date = batchRequest.GetJobTime().AsTime().Format("2006-01-02")
		err := result.AsJSON(true /*showDetails*/, &buffer)
		if err != nil {
			return fmt.Errorf("error during result.AsJSON: %w", err)
		}
	}

	filename := data.GetBlobFilename(
		fmt.Sprintf("shard-%05d", batchRequest.GetShardNum()),
		batchRequest.GetJobTime().AsTime())
	if err := data.WriteToBlobStore(ctx, bucketURL, filename, buffer.Bytes()); err != nil {
		return fmt.Errorf("error during WriteToBlobStore: %w", err)
	}
	log.Printf("Write to shard file successful: %s", filename)
	return nil
}

func createNetClients(ctx context.Context) (
	httpClient *http.Client, githubClient *github.Client, graphClient *githubv4.Client, logger *zap.Logger) {
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
	return
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

	usesBlobCache, envIsSet := os.LookupEnv(roundtripper.UseBlobCache)
	if !envIsSet || usesBlobCache == "" {
		// nolint: goerr113
		panic(fmt.Errorf("env_vars %s must be set", roundtripper.UseBlobCache))
	}
	blobCachePath, envIsSet := os.LookupEnv(roundtripper.BucketURL)
	if !envIsSet || blobCachePath == "" {
		// nolint: goerr113
		panic(fmt.Errorf("env_vars %s must be set", roundtripper.BucketURL))
	}

	httpClient, githubClient, graphClient, logger := createNetClients(ctx)

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
		if err := processRequest(ctx, req, bucketURL, httpClient, githubClient, graphClient); err != nil {
			panic(err)
		}
		// nolint: errcheck // flushes buffer
		logger.Sync()
		subscriber.Ack()
	}
	err = subscriber.Close()
	if err != nil {
		panic(err)
	}
}
