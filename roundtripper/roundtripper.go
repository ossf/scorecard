// Copyright 2020 Security Scorecard Authors
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
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/bradleyfalzon/ghinstallation"
	cache "github.com/naveensrinivasan/httpcache"
	"github.com/naveensrinivasan/httpcache/diskcache"
	"github.com/peterbourgon/diskv"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const (
	GithubAuthToken         = "GITHUB_AUTH_TOKEN" // #nosec G101
	GithubAppKeyPath        = "GITHUB_APP_KEY_PATH"
	GithubAppID             = "GITHUB_APP_ID"
	GithubAppInstallationID = "GITHUB_APP_INSTALLATION_ID"
	UseDiskCache            = "USE_DISK_CACHE"
	DiskCachePath           = "DISK_CACHE_PATH"
	UseBlobCache            = "USE_BLOB_CACHE"
	BucketURL               = "BLOB_URL"
)

// RateLimitRoundTripper is a rate-limit aware http.Transport for Github.
type RateLimitRoundTripper struct {
	Logger         *zap.SugaredLogger
	InnerTransport http.RoundTripper
}

type RoundRobinTokenSource struct {
	counter      int64
	AccessTokens []string
}

func (r *RoundRobinTokenSource) Token() (*oauth2.Token, error) {
	c := atomic.AddInt64(&r.counter, 1)
	index := c % int64(len(r.AccessTokens))
	return &oauth2.Token{
		AccessToken: r.AccessTokens[index],
	}, nil
}

// NewTransport returns a configured http.Transport for use with GitHub.
func NewTransport(ctx context.Context, logger *zap.SugaredLogger) http.RoundTripper {
	// Start with oauth
	transport := http.DefaultTransport
	if token := os.Getenv(GithubAuthToken); token != "" {
		ts := &RoundRobinTokenSource{
			AccessTokens: strings.Split(token, ","),
		}
		transport = oauth2.NewClient(ctx, ts).Transport
	} else if keyPath := os.Getenv(GithubAppKeyPath); keyPath != "" { // Also try a GITHUB_APP
		appID, err := strconv.Atoi(os.Getenv(GithubAppID))
		if err != nil {
			log.Panic(err)
		}
		installationID, err := strconv.Atoi(os.Getenv(GithubAppInstallationID))
		if err != nil {
			log.Panic(err)
		}
		transport, err = ghinstallation.NewKeyFromFile(transport, int64(appID), int64(installationID), keyPath)
		if err != nil {
			log.Panic(err)
		}
	}

	// Wrap that with the rate limiter
	rateLimit := &RateLimitRoundTripper{
		Logger:         logger,
		InnerTransport: transport,
	}

	// uses blob cache like GCS,S3.
	if cachePath, useBlob := shouldUseBlobCache(); useBlob {
		b, e := New(context.Background(), cachePath)
		if e != nil {
			log.Panic(e)
		}

		c := cache.NewTransport(b)
		c.Transport = rateLimit
		return c
	}

	// uses the disk cache
	if cachePath, useDisk := shouldUseDiskCache(); useDisk {
		const cacheSize uint64 = 10000 * 1024 * 1024 // 10gb
		c := cache.NewTransport(diskcache.NewWithDiskv(
			diskv.New(diskv.Options{BasePath: cachePath, CacheSizeMax: cacheSize})))
		c.Transport = rateLimit
		return c
	}

	// uses memory cache
	c := cache.NewTransport(cache.NewMemoryCache())
	c.Transport = rateLimit
	return c
}

// shouldUseDiskCache checks the env variables USE_DISK_CACHE and DISK_CACHE_PATH to determine if
// disk should be used for caching.
func shouldUseDiskCache() (string, bool) {
	if isDiskCache := os.Getenv(UseDiskCache); isDiskCache != "" {
		if result, err := strconv.ParseBool(isDiskCache); err == nil && result {
			if cachePath := os.Getenv(DiskCachePath); cachePath != "" {
				return cachePath, true
			}
		}
	}
	return "", false
}

// shouldUseBlobCache checks the env variables USE_BLOB_CACHE and BLOB_URL to determine if
// blob should be used for caching.
func shouldUseBlobCache() (string, bool) {
	if result, err := strconv.ParseBool(os.Getenv(UseBlobCache)); err == nil && result {
		if cachePath := os.Getenv(BucketURL); cachePath != "" {
			return cachePath, true
		}
	}
	return "", false
}

// Roundtrip handles caching and ratelimiting of responses from GitHub.
func (gh *RateLimitRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, err := gh.InnerTransport.RoundTrip(r)
	if err != nil {
		return nil, errors.Wrap(err, "error in round trip")
	}

	rateLimit := resp.Header.Get("X-RateLimit-Remaining")
	remaining, err := strconv.Atoi(rateLimit)
	if err != nil {
		return resp, nil
	}

	if remaining <= 0 {
		reset, err := strconv.Atoi(resp.Header.Get("X-RateLimit-Reset"))
		if err != nil {
			return resp, nil
		}

		duration := time.Until(time.Unix(int64(reset), 0))
		gh.Logger.Warnf("Rate limit exceeded. Waiting %s to retry...", duration)

		// Retry
		time.Sleep(duration)
		gh.Logger.Warnf("Rate limit exceeded. Retrying...")
		return gh.RoundTrip(r)
	}
	return resp, nil
}
