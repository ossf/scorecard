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

	"github.com/bradleyfalzon/ghinstallation"
	"go.uber.org/zap"
)

const (
	GithubAuthToken                = "GITHUB_AUTH_TOKEN" // #nosec G101
	GithubAppKeyPath               = "GITHUB_APP_KEY_PATH"
	GithubAppID                    = "GITHUB_APP_ID"
	GithubAppInstallationID        = "GITHUB_APP_INSTALLATION_ID"
	UseDiskCache                   = "USE_DISK_CACHE"
	DiskCachePath                  = "DISK_CACHE_PATH"
	UseBlobCache                   = "USE_BLOB_CACHE"
	BucketURL                      = "BLOB_URL"
	cacheSize               uint64 = 10000 * 1024 * 1024 // 10gb
)

// NewTransport returns a configured http.Transport for use with GitHub.
func NewTransport(ctx context.Context, logger *zap.SugaredLogger) http.RoundTripper {
	transport := http.DefaultTransport

	// Start with oauth
	if token := os.Getenv(GithubAuthToken); token != "" {
		transport = MakeOAuthTransport(ctx, strings.Split(token, ","))
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

	// Wrap that with the rate limiter, HTTP cache and census-enabled transport.
	return MakeCensusTransport(
		cachedTransportFactory(
			MakeRateLimitedTransport(transport, logger)))
}

func cachedTransportFactory(innerTransport http.RoundTripper) http.RoundTripper {
	// uses blob cache like GCS,S3.
	if cachePath, useBlob := shouldUseBlobCache(); useBlob {
		b, e := New(context.Background(), cachePath)
		if e != nil {
			log.Panic(e)
		}
		return MakeBlobCacheTransport(innerTransport, b)
	}

	// uses the disk cache
	if cachePath, useDisk := shouldUseDiskCache(); useDisk {
		return MakeDiskCacheTransport(innerTransport, cachePath, cacheSize)
	}

	// uses memory cache
	return MakeInMemoryCacheTransport(innerTransport)
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
