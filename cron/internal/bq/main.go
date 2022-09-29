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
	"strings"
	"time"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/ossf/scorecard/v4/cron/data"
	"github.com/ossf/scorecard/v4/cron/internal/config"
)

type shardSummary struct {
	shardMetadata  []byte
	shardsExpected int
	shardsCreated  int
	isTransferred  bool
}

type bucketSummary struct {
	shards map[time.Time]*shardSummary
}

func (summary *bucketSummary) getOrCreate(t time.Time) *shardSummary {
	if summary.shards[t] == nil {
		summary.shards[t] = new(shardSummary)
	}
	return summary.shards[t]
}

// getBucketSummary iterates through all files in a bucket and
// returns `shardSummary` keyed by the shard creation time.
func getBucketSummary(ctx context.Context, bucketURL string) (*bucketSummary, error) {
	keys, err := data.GetBlobKeys(ctx, bucketURL)
	if err != nil {
		return nil, fmt.Errorf("error getting BlobKeys: %w", err)
	}

	summary := bucketSummary{
		shards: make(map[time.Time]*shardSummary),
	}
	for _, key := range keys {
		creationTime, filename, err := data.ParseBlobFilename(key)
		if err != nil {
			return nil, fmt.Errorf("error parsing Blob key: %w", err)
		}
		switch {
		case strings.HasPrefix(filename, "shard-"):
			summary.getOrCreate(creationTime).shardsCreated++
		case filename == config.TransferStatusFilename:
			summary.getOrCreate(creationTime).isTransferred = true
		case filename == config.ShardMetadataFilename:
			keyData, err := data.GetBlobContent(ctx, bucketURL, key)
			if err != nil {
				return nil, fmt.Errorf("error during GetBlobContent: %w", err)
			}
			var metadata data.ShardMetadata
			if err := protojson.Unmarshal(keyData, &metadata); err != nil {
				return nil, fmt.Errorf("error parsing data as ShardMetadata: %w", err)
			}
			summary.getOrCreate(creationTime).shardsExpected = int(metadata.GetNumShard())
			summary.getOrCreate(creationTime).shardMetadata = keyData
		default:
			//nolint: goerr113
			return nil, fmt.Errorf("found unrecognized file: %s", key)
		}
	}
	return &summary, nil
}

// isCompleted checks if the percentage of completed shards is over the desired completion threshold.
// It also returns false to prevent transfers in cases where the expected number of shards is 0,
// as either the .shard_metadata file is missing, or there is nothing to transfer anyway.
func isCompleted(expected, created int, completionThreshold float64) bool {
	completedPercentage := float64(created) / float64(expected)
	return expected > 0 && completedPercentage >= completionThreshold
}

func transferDataToBq(ctx context.Context,
	bucketURL, projectID, datasetName, tableName string, completionThreshold float64, webhookURL string,
	summary *bucketSummary,
) error {
	for creationTime, shards := range summary.shards {
		if shards.isTransferred || !isCompleted(shards.shardsExpected, shards.shardsCreated, completionThreshold) {
			continue
		}

		shardFileURI := data.GetBlobFilename("shard-*", creationTime)
		if err := startDataTransferJob(ctx,
			bucketURL, shardFileURI, projectID, datasetName, tableName,
			creationTime); err != nil {
			return fmt.Errorf("error during StartDataTransferJob: %w", err)
		}

		transferStatusFilename := data.GetTransferStatusFilename(creationTime)
		if err := data.WriteToBlobStore(ctx, bucketURL, transferStatusFilename, nil); err != nil {
			return fmt.Errorf("error during WriteToBlobStore: %w", err)
		}
		if webhookURL == "" {
			continue
		}
		//nolint: noctx, gosec // variable URL is ok here.
		resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(shards.shardMetadata))
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

	summary, err := getBucketSummary(ctx, bucketURL)
	if err != nil {
		panic(err)
	}

	if err := transferDataToBq(ctx,
		bucketURL, projectID, datasetName, tableName, completionThreshold, webhookURL,
		summary); err != nil {
		panic(err)
	}
}
