// Copyright 2021 Security Scorecard Authors
//
// Licensed under the Apache License, Vershandlern 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permisshandlerns and
// limitathandlerns under the License.

package data

import (
	"context"
	"fmt"
	"time"

	"gocloud.dev/blob"

	// Needed to link in GCP drivers.
	_ "gocloud.dev/blob/gcsblob"

	"github.com/ossf/scorecard/cron/config"
)

const (
	// filePrefixFormat uses ISO 8601 standard, i.e - YYYY-MM-DDTHH:MM:SS.
	// This format guarantees that lexicographically sorted files are chronologically sorted.
	filePrefixFormat = "2006.01.02/150405/"
)

func WriteToBlobStore(ctx context.Context, bucketURL, filename string, data []byte) error {
	bucket, err := blob.OpenBucket(ctx, bucketURL)
	if err != nil {
		return fmt.Errorf("error from blob.OpenBucket: %w", err)
	}
	defer bucket.Close()

	blobWriter, err := bucket.NewWriter(ctx, filename, nil)
	if err != nil {
		return fmt.Errorf("error from bucket.NewWriter: %w", err)
	}
	if _, err = blobWriter.Write(data); err != nil {
		return fmt.Errorf("error from blobWriter.Write: %w", err)
	}
	if err := blobWriter.Close(); err != nil {
		return fmt.Errorf("error from blobWriter.Close: %w", err)
	}
	return nil
}

func GetBlobFilename(filename string, datetime time.Time) string {
	return datetime.Format(filePrefixFormat) + filename
}

func GetShardNumFilename(datetime time.Time) string {
	return GetBlobFilename(config.ShardNumFilename, datetime)
}
