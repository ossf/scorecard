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
	"net/http"

	// nolint:gosec
	_ "net/http/pprof"

	"go.opencensus.io/stats/view"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo"
	githubstats "github.com/ossf/scorecard/v4/clients/githubrepo/stats"
	"github.com/ossf/scorecard/v4/cron/config"
	"github.com/ossf/scorecard/v4/cron/data"
	format "github.com/ossf/scorecard/v4/cron/format"
	"github.com/ossf/scorecard/v4/cron/monitoring"
	"github.com/ossf/scorecard/v4/cron/pubsub"
	docs "github.com/ossf/scorecard/v4/docs/checks"
	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/pkg"
	"github.com/ossf/scorecard/v4/stats"
)

var ignoreRuntimeErrors = flag.Bool("ignoreRuntimeErrors", false, "if set to true any runtime errors will be ignored")

func processRequest(ctx context.Context,
	batchRequest *data.ScorecardBatchRequest, checksToRun checker.CheckNameToFnMap,
	bucketURL, bucketURL2 string, checkDocs docs.Doc,
	repoClient clients.RepoClient, ossFuzzRepoClient clients.RepoClient,
	ciiClient clients.CIIBestPracticesClient,
	vulnsClient clients.VulnerabilitiesClient,
	logger *zap.Logger) error {
	filename := data.GetBlobFilename(
		fmt.Sprintf("shard-%07d", batchRequest.GetShardNum()),
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
		logger.Info(fmt.Sprintf("Already processed shard %s. Nothing to do.", filename))
		// We have already processed this request, nothing to do.
		return nil
	}

	var buffer bytes.Buffer
	var buffer2 bytes.Buffer
	// TODO: run Scorecard for each repo in a separate thread.
	for _, repo := range batchRequest.GetRepos() {
		logger.Info(fmt.Sprintf("Running Scorecard for repo: %s", *repo.Url))
		repo, err := githubrepo.MakeGithubRepo(*repo.Url)
		if err != nil {
			logger.Warn(fmt.Sprintf("invalid GitHub URL: %v", err))
			continue
		}
		repo.AppendMetadata(repo.Metadata()...)
		result, err := pkg.RunScorecards(ctx, repo, false, checksToRun,
			repoClient, ossFuzzRepoClient, ciiClient, vulnsClient)
		if errors.Is(err, sce.ErrRepoUnreachable) {
			// Not accessible repo - continue.
			continue
		}
		if err != nil {
			return fmt.Errorf("error during RunScorecards: %w", err)
		}
		for checkIndex := range result.Checks {
			check := &result.Checks[checkIndex]
			if !errors.Is(check.Error2, sce.ErrScorecardInternal) {
				continue
			}
			errorMsg := fmt.Sprintf("check %s has a runtime error: %v", check.Name, check.Error2)
			if !(*ignoreRuntimeErrors) {
				// nolint: goerr113
				return errors.New(errorMsg)
			}
			logger.Warn(errorMsg)
		}
		result.Date = batchRequest.GetJobTime().AsTime()
		if err := format.AsJSON(&result, true /*showDetails*/, zapcore.InfoLevel, &buffer); err != nil {
			return fmt.Errorf("error during result.AsJSON: %w", err)
		}

		if err := format.AsJSON2(&result, true /*showDetails*/, zapcore.InfoLevel, checkDocs, &buffer2); err != nil {
			return fmt.Errorf("error during result.AsJSON2: %w", err)
		}
	}
	if err := data.WriteToBlobStore(ctx, bucketURL, filename, buffer.Bytes()); err != nil {
		return fmt.Errorf("error during WriteToBlobStore: %w", err)
	}

	if err := data.WriteToBlobStore(ctx, bucketURL2, filename, buffer2.Bytes()); err != nil {
		return fmt.Errorf("error during WriteToBlobStore2: %w", err)
	}

	logger.Info(fmt.Sprintf("Write to shard file successful: %s", filename))

	return nil
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
		&stats.OutgoingHTTPRequests,
		&githubstats.GithubTokens); err != nil {
		return nil, fmt.Errorf("error during view.Register: %w", err)
	}
	return exporter, nil
}

func main() {
	ctx := context.Background()

	flag.Parse()

	checkDocs, err := docs.Read()
	if err != nil {
		panic(err)
	}

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

	blacklistedChecks, err := config.GetBlacklistedChecks()
	if err != nil {
		panic(err)
	}

	ciiDataBucketURL, err := config.GetCIIDataBucketURL()
	if err != nil {
		panic(err)
	}

	logger, err := githubrepo.NewLogger(zap.InfoLevel)
	if err != nil {
		panic(err)
	}
	repoClient := githubrepo.CreateGithubRepoClient(ctx, logger)
	ciiClient := clients.BlobCIIBestPracticesClient(ciiDataBucketURL)
	ossFuzzRepoClient, err := githubrepo.CreateOssFuzzRepoClient(ctx, logger)
	vulnsClient := clients.DefaultVulnerabilitiesClient()
	if err != nil {
		panic(err)
	}
	defer ossFuzzRepoClient.Close()

	exporter, err := startMetricsExporter()
	if err != nil {
		panic(err)
	}
	defer exporter.StopMetricsExporter()

	// Exposed for monitoring runtime profiles
	go func() {
		logger.Fatal(fmt.Sprintf("%v", http.ListenAndServe(":8080", nil)))
	}()

	checksToRun := checks.AllChecks
	for _, check := range blacklistedChecks {
		delete(checksToRun, check)
	}
	for {
		req, err := subscriber.SynchronousPull()
		if err != nil {
			panic(err)
		}
		logger.Info("Received message from subscription")
		if req == nil {
			logger.Warn("subscription returned nil message during Receive, exiting")
			break
		}
		if err := processRequest(ctx, req, checksToRun,
			bucketURL, bucketURL2, checkDocs,
			repoClient, ossFuzzRepoClient, ciiClient, vulnsClient, logger); err != nil {
			logger.Warn(fmt.Sprintf("error processing request: %v", err))
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
