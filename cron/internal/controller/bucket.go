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

// Package main implements the PubSub controller.
package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/ossf/scorecard/v4/cron/config"
	"github.com/ossf/scorecard/v4/cron/data"
)

// getPrefix returns the prefix used when reading input files from a bucket.
// If "prefix" is set, the value is used irrespective of the value of "prefix-file".
// Otherwise, the contents of "prefix-file" (if set) are used.
func getPrefix(ctx context.Context, bucket string) (string, error) {
	prefix, err := config.GetInputBucketPrefix()
	if err != nil {
		return "", fmt.Errorf("config.GetInputBucketPrefix: %w", err)
	}
	if prefix != "" {
		// prioritize prefix if set
		return prefix, nil
	}

	prefixFile, err := config.GetInputBucketPrefixFile()
	if err != nil {
		return "", fmt.Errorf("config.GetInputBucketPrefixFile: %w", err)
	}
	if prefixFile == "" {
		// cant read a file which doesnt exist, but the value is optional so no error
		return "", nil
	}

	b, err := data.GetBlobContent(ctx, bucket, prefixFile)
	if err != nil {
		return "", fmt.Errorf("fetching contents of prefix-file: %w", err)
	}
	s := string(b)
	return strings.TrimSpace(s), nil
}

func bucketFiles(ctx context.Context) data.Iterator {
	var iters []data.Iterator

	bucket, err := config.GetInputBucketURL()
	if err != nil {
		panic(err)
	}
	prefix, err := getPrefix(ctx, bucket)
	if err != nil {
		panic(err)
	}

	files, err := data.GetBlobKeysWithPrefix(ctx, bucket, prefix)
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		b, err := data.GetBlobContent(ctx, bucket, f)
		if err != nil {
			panic(err)
		}
		r := bytes.NewReader(b)
		i, err := data.MakeIteratorFrom(r)
		if err != nil {
			panic(err)
		}
		iters = append(iters, i)
	}
	iter, err := data.MakeNestedIterator(iters)
	if err != nil {
		panic(err)
	}
	return iter
}
