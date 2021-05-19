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
	"net/http"

	cache "github.com/naveensrinivasan/httpcache"
	"github.com/naveensrinivasan/httpcache/diskcache"
	"github.com/peterbourgon/diskv"
)

func MakeBlobCacheTransport(innerTransport http.RoundTripper, blobCache cache.Cache) http.RoundTripper {
	return makeCachedTransport(innerTransport, blobCache)
}

func MakeDiskCacheTransport(innerTransport http.RoundTripper, cachePath string, cacheSize uint64) http.RoundTripper {
	c := diskcache.NewWithDiskv(
		diskv.New(diskv.Options{BasePath: cachePath, CacheSizeMax: cacheSize}))
	return makeCachedTransport(innerTransport, c)
}

func MakeInMemoryCacheTransport(innerTransport http.RoundTripper) http.RoundTripper {
	return makeCachedTransport(innerTransport, cache.NewMemoryCache())
}

func makeCachedTransport(innerTransport http.RoundTripper, backingCache cache.Cache) http.RoundTripper {
	c := cache.NewTransport(backingCache)
	c.Transport = innerTransport
	return c
}
