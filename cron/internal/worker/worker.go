// Copyright 2022 Security Scorecard Authors
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

// Package worker implements a generic cron worker job.
package worker

import (
	"context"
	"fmt"

	"github.com/ossf/scorecard/v4/cron/internal/config"
	"github.com/ossf/scorecard/v4/cron/internal/data"
	"github.com/ossf/scorecard/v4/cron/internal/pubsub"
	"github.com/ossf/scorecard/v4/log"
)

type Worker interface {
	Process(ctx context.Context, req *data.ScorecardBatchRequest, bucketURL string, logger *log.Logger) error
	PostProcess()
}

type WorkLoop struct {
	worker Worker
}

func NewWorkLoop(worker Worker) WorkLoop {
	return WorkLoop{worker: worker}
}

func (wl *WorkLoop) Run() error {
	ctx := context.Background()

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

	logger := log.NewLogger(log.InfoLevel)

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

		exists, err := resultExists(ctx, req, bucketURL)
		if err != nil {
			announceError(err, subscriber, logger)
			continue
		}

		// Sanity check - make sure we are not re-processing an already processed request.
		if !exists {
			if err := wl.worker.Process(ctx, req, bucketURL, logger); err != nil {
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

func ResultFilename(sbr *data.ScorecardBatchRequest) string {
	shardname := fmt.Sprintf("shard-%07d", sbr.GetShardNum())
	return data.GetBlobFilename(shardname, sbr.GetJobTime().AsTime())
}
