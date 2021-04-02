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

package roundtripper

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"gocloud.dev/blob"

	//nolint
	_ "gocloud.dev/blob/gcsblob"
)

// Cache that use the blob storage like GCS or S3.
// TODO - Handle returning errors.
type Cache struct {
	Bucket *blob.Bucket
}

// Get returns the []byte representation of the response and true if present, false if not.The
// error is not returned to maintain compatabiltiy with the httpcache Cache interface.
func (c *Cache) Get(key string) ([]byte, bool) {
	resp, err := c.Bucket.ReadAll(context.Background(), key)
	if err != nil {
		return nil, false
	}
	return resp, true
}

// Set saves response resp to the cache with key. The error is not returned to maintain compatabiltiy
// with the httpcache Cache interface.
func (c *Cache) Set(key string, resp []byte) {
	//nolint:errcheck
	c.Bucket.WriteAll(context.Background(), key, resp, nil)
}

// Delete removes key from the cache.The error is not returned to maintain compatabiltiy
// with the httpcache Cache interface.
func (c *Cache) Delete(key string) {
	//nolint:errcheck
	c.Bucket.Delete(context.Background(), key)
}

// New opens the bucket for caching.
func New(ctx context.Context, bucketKey string) (*Cache, error) {
	b, err := blob.OpenBucket(context.Background(), bucketKey)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("error in opening the bucket %s", bucketKey))
	}
	return &Cache{Bucket: b}, nil
}
