// Copyright 2026 OpenSSF Scorecard Authors
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

// Package cdn implements clients for CDN operations.
package cdn

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var errPurgeFailed = errors.New("purge failed")

// Purger is the interface for purging URLs from a CDN.
type Purger interface {
	Purge(ctx context.Context, url string) error
}

// FastlyClient implements Purger for Fastly.
type FastlyClient struct {
	token   string
	baseURL string
}

// NewFastlyClient creates a new FastlyClient.
func NewFastlyClient(token, baseURL string) *FastlyClient {
	baseURL = strings.TrimSuffix(baseURL, "/")
	return &FastlyClient{token: token, baseURL: baseURL}
}

// Purge purges the given URL from Fastly.
// It sends a PURGE request to the given URL with the Fastly-Key header.
func (c *FastlyClient) Purge(ctx context.Context, path string) error {
	req, err := http.NewRequestWithContext(ctx, "PURGE", c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("http.NewRequest: %w", err)
	}
	req.Header.Set("Fastly-Key", c.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http.Do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %s", errPurgeFailed, resp.Status)
	}
	return nil
}

// NoOpClient is a no-op implementation of PurgeClient.
type NoOpClient struct{}

// NewNoOpClient creates a new NoOpClient.
func NewNoOpClient() *NoOpClient {
	return &NoOpClient{}
}

// Purge does nothing.
func (c *NoOpClient) Purge(ctx context.Context, url string) error {
	return nil
}
