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

// Package main implements the PubSub controller.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ossf/scorecard/v2/cron/config"
	"github.com/ossf/scorecard/v2/cron/data"
	"github.com/ossf/scorecard/v2/cron/pubsub"
)

const commitSHA = "SCORECARD_COMMIT_SHA"

func publishToRepoRequestTopic(ctx context.Context, iter data.Iterator, datetime time.Time) (int32, error) {
	var shardNum int32
	request := data.ScorecardBatchRequest{
		JobTime:  timestamppb.New(datetime),
		ShardNum: &shardNum,
	}
	topic, err := config.GetRequestTopicURL()
	if err != nil {
		return shardNum, fmt.Errorf("error getting RequestTopicURL: %w", err)
	}
	topicPublisher, err := pubsub.CreatePublisher(ctx, topic)
	if err != nil {
		return shardNum, fmt.Errorf("error running CreatePublisher: %w", err)
	}

	shardSize, err := config.GetShardSize()
	if err != nil {
		return shardNum, fmt.Errorf("error getting ShardSize: %w", err)
	}

	// Create and send batch requests of repoURLs of size `ShardSize`:
	// * Iterate through incoming repoURLs until `request` has len(Repos) of size `ShardSize`.
	// * Publish request to PubSub topic.
	// * Clear request.Repos and increment shardNum.
	for iter.HasNext() {
		repoURL, err := iter.Next()
		if err != nil {
			return shardNum, fmt.Errorf("error reading repoURL: %w", err)
		}
		request.Repos = append(request.GetRepos(), repoURL.URL())
		if len(request.GetRepos()) < shardSize {
			continue
		}
		if err := topicPublisher.Publish(&request); err != nil {
			return shardNum, fmt.Errorf("error running topicPublisher.Publish: %w", err)
		}
		request.Repos = nil
		shardNum++
	}
	// Check if more repoURLs are pending to be sent in `request`.
	if len(request.GetRepos()) > 0 {
		if err := topicPublisher.Publish(&request); err != nil {
			return shardNum, fmt.Errorf("error running topicPublisher.Publish: %w", err)
		}
	}

	if err := topicPublisher.Close(); err != nil {
		return shardNum, fmt.Errorf("error running topicPublisher.Close: %w", err)
	}
	return shardNum, nil
}

func main() {
	ctx := context.Background()
	t := time.Now()

	// nolint: gomnd
	if len(os.Args) != 2 {
		panic("must provide a single argument")
	}
	// nolint: gomnd
	inFile, err := os.OpenFile(os.Args[1], os.O_RDONLY, 0o644)
	if err != nil {
		panic(err)
	}
	reader, err := data.MakeIteratorFrom(inFile)
	if err != nil {
		panic(err)
	}

	bucket, err := config.GetResultDataBucketURL()
	if err != nil {
		panic(err)
	}

	bucket2, err := config.GetResultDataBucketURLV2()
	if err != nil {
		panic(err)
	}

	shardNum, err := publishToRepoRequestTopic(ctx, reader, t)
	if err != nil {
		panic(err)
	}
	// Populate `.shard_metadata` file.
	metadata := data.ShardMetadata{
		NumShard:  new(int32),
		ShardLoc:  new(string),
		CommitSha: new(string),
	}
	*metadata.NumShard = (shardNum + 1)
	*metadata.ShardLoc = bucket + "/" + data.GetBlobFilename("", t)
	*metadata.CommitSha = os.Getenv(commitSHA)
	metadataJSON, err := protojson.Marshal(&metadata)
	if err != nil {
		panic(fmt.Errorf("error during protojson.Marshal: %w", err))
	}
	err = data.WriteToBlobStore(ctx, bucket, data.GetShardMetadataFilename(t), metadataJSON)
	if err != nil {
		panic(fmt.Errorf("error writing to BlobStore: %w", err))
	}

	// UPGRADEv2: to remove.
	*metadata.ShardLoc = bucket2 + "/" + data.GetBlobFilename("", t)
	metadataJSON, err = protojson.Marshal(&metadata)
	if err != nil {
		panic(fmt.Errorf("error during protojson.Marshal2: %w", err))
	}
	err = data.WriteToBlobStore(ctx, bucket2, data.GetShardMetadataFilename(t), metadataJSON)
	if err != nil {
		panic(fmt.Errorf("error writing to BlobStore2: %w", err))
	}
}
