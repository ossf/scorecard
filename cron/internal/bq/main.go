// Copyright 2021 OpenSSF Scorecard Authors
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

// Package main implements the BQ transfer job.
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/ossf/scorecard/v4/cron/config"
	"github.com/ossf/scorecard/v4/cron/data"
)

func transferDataToBq(ctx context.Context,
	bucketURL, projectID, datasetName, tableName string, completionThreshold float64, webhookURL string,
	summary *data.BucketSummary,
) error {
	for _, shards := range summary.Shards() {
		if shards.IsTransferred() || !shards.IsCompleted(completionThreshold) {
			continue
		}

		shardFileURI := data.GetBlobFilename("shard-*", shards.CreationTime())
		if err := startDataTransferJob(ctx,
			bucketURL, shardFileURI, projectID, datasetName, tableName,
			shards.CreationTime()); err != nil {
			return fmt.Errorf("error during StartDataTransferJob: %w", err)
		}

		if err := shards.MarkTransferred(ctx, bucketURL); err != nil {
			return fmt.Errorf("error during MarkTransferred: %w", err)
		}
		if webhookURL == "" {
			continue
		}
		//nolint:noctx,gosec // variable URL is ok here.
		resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(shards.Metadata()))
		if err != nil {
			return fmt.Errorf("error during http.Post to %s: %w", webhookURL, err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading resp.Body: %w", err)
		}
		log.Printf("Returned status: %s %s", resp.Status, body)
	}
	return nil
}

func getBQConfig() (projectID, datasetName, tableName string, err error) {
	projectID, err = config.GetProjectID()
	if err != nil {
		return projectID, datasetName, tableName, fmt.Errorf("error getting ProjectId: %w", err)
	}
	datasetName, err = config.GetBigQueryDataset()
	if err != nil {
		return projectID, datasetName, tableName, fmt.Errorf("error getting BigQuery dataset: %w", err)
	}
	tableName, err = config.GetBigQueryTable()
	if err != nil {
		return projectID, datasetName, tableName, fmt.Errorf("error getting BigQuery table: %w", err)
	}
	return
}

func main() {
	ctx := context.Background()

	flag.Parse()
	if err := config.ReadConfig(); err != nil {
		panic(err)
	}

	bucketURL, err := config.GetResultDataBucketURL()
	if err != nil {
		panic(err)
	}
	webhookURL, err := config.GetWebhookURL()
	if err != nil {
		panic(err)
	}
	projectID, datasetName, tableName, err := getBQConfig()
	if err != nil {
		panic(err)
	}
	completionThreshold, err := config.GetCompletionThreshold()
	if err != nil {
		panic(err)
	}

	summary, err := data.GetBucketSummary(ctx, bucketURL)
	if err != nil {
		panic(err)
	}

	if err := transferDataToBq(ctx,
		bucketURL, projectID, datasetName, tableName, completionThreshold, webhookURL,
		summary); err != nil {
		panic(err)
	}
}
