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
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ossf/scorecard/cron/config"
	"github.com/ossf/scorecard/cron/data"
)

type shardSummary struct {
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
		case filename == config.ShardNumFilename:
			keyData, err := data.GetBlobContent(ctx, bucketURL, key)
			if err != nil {
				return nil, fmt.Errorf("error during GetBlobContent: %w", err)
			}
			summary.getOrCreate(creationTime).shardsExpected, err = strconv.Atoi(string(keyData))
			if err != nil {
				return nil, fmt.Errorf("error during strconv.Atoi: %w", err)
			}
		case strings.HasPrefix(filename, "shard-"):
			summary.getOrCreate(creationTime).shardsCreated++
		case filename == config.TransferStatusFilename:
			summary.getOrCreate(creationTime).isTransferred = true
		default:
			// nolint: goerr113
			return nil, fmt.Errorf("found unrecognized file: %s", key)
		}
	}
	return &summary, nil
}

func transferDataToBq(ctx context.Context, bucketURL string, summary *bucketSummary) error {
	for creationTime, shards := range summary.shards {
		if shards.isTransferred || shards.shardsExpected != shards.shardsCreated {
			continue
		}

		shardFileURI := data.GetBlobFilename("shard-*", creationTime)
		if err := StartDataTransferJob(ctx, bucketURL, shardFileURI, creationTime); err != nil {
			return fmt.Errorf("error during StartDataTransferJob: %w", err)
		}

		transferStatusFilename := data.GetTransferStatusFilename(creationTime)
		if err := data.WriteToBlobStore(ctx, bucketURL, transferStatusFilename, nil); err != nil {
			return fmt.Errorf("error during WriteToBlobStore: %w", err)
		}
	}
	return nil
}

func main() {
	ctx := context.Background()
	bucketURL, err := config.GetResultDataBucketURL()
	if err != nil {
		panic(err)
	}

	summary, err := getBucketSummary(ctx, bucketURL)
	if err != nil {
		panic(err)
	}

	if err := transferDataToBq(ctx, bucketURL, summary); err != nil {
		panic(err)
	}
}
