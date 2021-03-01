package main

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
func (c *Cache) Set(key string, resp []byte) error {
	return c.Bucket.WriteAll(context.Background(), key, resp, nil)
}

// Delete removes key from the cache.The error is not returned to maintain compatabiltiy
// with the httpcache Cache interface.
func (c *Cache) Delete(key string) error {
	return c.Bucket.Delete(context.Background(), key)
}

// New opens the bucket for caching.
func New(bucketKey string) (*Cache, error) {
	b, err := blob.OpenBucket(context.Background(), bucketKey)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("error in opening the bucket %s", bucketKey))
	}
	return &Cache{Bucket: b}, nil
}
