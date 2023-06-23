// Copyright 2022 OpenSSF Scorecard Authors
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

// Package worker implements the generic cron worker logic.
package worker

import (
	"context"
	"flag"
	"fmt"

	"github.com/ossf/scorecard/v4/cron/config"
	"github.com/ossf/scorecard/v4/cron/data"
	"github.com/ossf/scorecard/v4/cron/internal/pubsub"
	"github.com/ossf/scorecard/v4/log"
)

// Worker is the interface used to process batch requests.
//
// Process does the processing for a batch request. Returning an error will cause the request to be nack'd,
// allowing it to be re-processed later. If no error is returned, the request will be ack'd, consuming it.
//
// PostProcess is called only after an error-free call to Process.
type Worker interface {
	Process(ctx context.Context, req *data.ScorecardBatchRequest, bucketURL string) error
	PostProcess()
}

// WorkLoop is the entry point into the common cron worker structure.
type WorkLoop struct {
	worker Worker
}

// NewWorkLoop creates a workloop using a specified worker.
func NewWorkLoop(worker Worker) WorkLoop {
	return WorkLoop{worker: worker}
}

// Run initiates the processing performed by the WorkLoop.
func (wl *WorkLoop) Run() error {
	ctx := context.Background()

	if !flag.Parsed() {
		flag.Parse()
	}

	if err := config.ReadConfig(); err != nil {
		return fmt.Errorf("config.ReadConfig: %w", err)
	}

	subscriptionURL, err := config.GetRequestSubscriptionURL()
	if err != nil {
		return fmt.Errorf("config.GetRequestSubscriptionURL: %w", err)
	}

	subscriber, err := pubsub.CreateSubscriber(ctx, subscriptionURL)
	if err != nil {
		return fmt.Errorf("config.CreateSubscriber: %w", err)
	}

	bucketURL, err := config.GetResultDataBucketURL()
	if err != nil {
		return fmt.Errorf("config.GetResultDataBucketURL: %w", err)
	}

	logger := log.NewCronLogger(log.InfoLevel)

	for {
		req, err := subscriber.SynchronousPull()
		if err != nil {
			return fmt.Errorf("subscriber.SynchronousPull: %w", err)
		}

		logger.Info("Received message from subscription")
		if req == nil {
			// TODO(log): Previously Warn. Consider logging an error here.
			logger.Info("subscription returned nil message during Receive, exiting")
			break
		}

		// don't process requests from jobs without metadata files, as the results will never be transferred.
		// https://github.com/ossf/scorecard/issues/2307
		hasMd, err := hasMetadataFile(ctx, req, bucketURL)
		if err != nil {
			announceError(err, subscriber, logger)
			continue
		}

		if !hasMd {
			// nack the message so it can be tried later, as the metadata file may not have been created yet.
			subscriber.Nack()
			continue
		}

		exists, err := resultExists(ctx, req, bucketURL)
		if err != nil {
			announceError(err, subscriber, logger)
			continue
		}

		// Sanity check - make sure we are not re-processing an already processed request.
		if exists {
			logger.Info(fmt.Sprintf("Skipping already processed request: %s.", req.String()))
		} else {
			if err := wl.worker.Process(ctx, req, bucketURL); err != nil {
				announceError(err, subscriber, logger)
				continue
			}
		}

		wl.worker.PostProcess()
		subscriber.Ack()
	}
	if err := subscriber.Close(); err != nil {
		return fmt.Errorf("subscriber.Close: %w", err)
	}
	return nil
}

func announceError(err error, subscriber pubsub.Subscriber, logger *log.Logger) {
	// TODO(log): Previously Warn. Consider logging an error here.
	logger.Info(fmt.Sprintf("error processing request: %v", err))
	// Nack the message so that another worker can retry.
	subscriber.Nack()
}

func resultExists(ctx context.Context, sbr *data.ScorecardBatchRequest, bucketURL string) (bool, error) {
	exists, err := data.BlobExists(ctx, bucketURL, ResultFilename(sbr))
	if err != nil {
		return false, fmt.Errorf("error during BlobExists: %w", err)
	}
	return exists, nil
}

// ResultFilename returns the filename where the result from processing a batch request should go.
// This naming convention is used to detect duplicate requests, as well as transfer the results to BigQuery.
func ResultFilename(sbr *data.ScorecardBatchRequest) string {
	shardname := fmt.Sprintf("shard-%07d", sbr.GetShardNum())
	return data.GetBlobFilename(shardname, sbr.GetJobTime().AsTime())
}

func hasMetadataFile(ctx context.Context, req *data.ScorecardBatchRequest, bucketURL string) (bool, error) {
	filename := data.GetShardMetadataFilename(req.GetJobTime().AsTime().UTC())
	exists, err := data.BlobExists(ctx, bucketURL, filename)
	if err != nil {
		return false, fmt.Errorf("data.BlobExists: %w", err)
	}
	return exists, nil
}
