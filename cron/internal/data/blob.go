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

// Package data contains util fns for reading and writing data handled by the cron job.
package data

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"gocloud.dev/blob"
	// Needed to link in GCP drivers.
	_ "gocloud.dev/blob/gcsblob"

	"github.com/ossf/scorecard/v4/cron/internal/config"
)

const (
	// filePrefixFormat uses ISO 8601 standard, i.e - YYYY-MM-DDTHH:MM:SS.
	// This format guarantees that lexicographically sorted files are chronologically sorted.
	filePrefixFormat = "2006.01.02/150405/"
)

var (
	errShortBlobName = errors.New("input key length is shorter than expected")
	errParseBlobName = errors.New("error parsing input blob name")
)

func blobKeysPrefix(ctx context.Context, bucket *blob.Bucket, filePrefix string) ([]string, error) {
	keys := make([]string, 0)
	iter := bucket.List(&blob.ListOptions{
		Prefix: filePrefix,
	})
	for {
		next, err := iter.Next(ctx)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error during iter.Next: %w", err)
		}
		keys = append(keys, next.Key)
	}
	return keys, nil
}

// GetBlobKeysWithPrefix returns all object keys for a given bucketURL which start with prefix.
// The prefix can be used to specify directories.
func GetBlobKeysWithPrefix(ctx context.Context, bucketURL, filePrefix string) ([]string, error) {
	bucket, err := blob.OpenBucket(ctx, bucketURL)
	if err != nil {
		return nil, fmt.Errorf("error from blob.OpenBucket: %w", err)
	}
	defer bucket.Close()
	return blobKeysPrefix(ctx, bucket, filePrefix)
}

// GetBlobKeys returns all object keys for a given bucketURL.
func GetBlobKeys(ctx context.Context, bucketURL string) ([]string, error) {
	return GetBlobKeysWithPrefix(ctx, bucketURL, "")
}

// GetBlobContent returns the file content given a bucketURL and object key.
func GetBlobContent(ctx context.Context, bucketURL, key string) ([]byte, error) {
	bucket, err := blob.OpenBucket(ctx, bucketURL)
	if err != nil {
		return nil, fmt.Errorf("error from blob.OpenBucket: %w", err)
	}
	defer bucket.Close()

	keyData, err := bucket.ReadAll(ctx, key)
	if err != nil {
		return keyData, fmt.Errorf("error during bucket.ReadAll: %w", err)
	}
	return keyData, nil
}

// BlobExists checks whether a given `bucketURL/key` blob exists.
func BlobExists(ctx context.Context, bucketURL, key string) (bool, error) {
	bucket, err := blob.OpenBucket(ctx, bucketURL)
	if err != nil {
		return false, fmt.Errorf("error from blob.OpenBucket: %w", err)
	}
	defer bucket.Close()

	ret, err := bucket.Exists(ctx, key)
	if err != nil {
		return ret, fmt.Errorf("error during bucket.Exists: %w", err)
	}
	return ret, nil
}

// WriteToBlobStore creates and writes data to filename in bucketURL.
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

// GetBlobFilename returns a blob key for a shard. Takes Time object and filename as input.
func GetBlobFilename(filename string, datetime time.Time) string {
	return datetime.Format(filePrefixFormat) + filename
}

// GetShardNumFilename returns shard_num filename for a shard.
func GetShardNumFilename(datetime time.Time) string {
	return GetBlobFilename(config.ShardNumFilename, datetime)
}

// GetTransferStatusFilename returns transfer_status filename for a shard.
func GetTransferStatusFilename(datetime time.Time) string {
	return GetBlobFilename(config.TransferStatusFilename, datetime)
}

// GetShardMetadataFilename returns shard_metadata filename for a shard.
func GetShardMetadataFilename(datetime time.Time) string {
	return GetBlobFilename(config.ShardMetadataFilename, datetime)
}

// ParseBlobFilename parses a blob key into a Time object.
func ParseBlobFilename(key string) (time.Time, string, error) {
	if len(key) < len(filePrefixFormat) {
		return time.Now(), "", fmt.Errorf("%w: %s", errShortBlobName, key)
	}
	prefix := key[:len(filePrefixFormat)]
	objectName := key[len(filePrefixFormat):]
	t, err := time.Parse(filePrefixFormat, prefix)
	if err != nil {
		return t, "", fmt.Errorf("%w: %v", errParseBlobName, err)
	}
	return t, objectName, nil
}
