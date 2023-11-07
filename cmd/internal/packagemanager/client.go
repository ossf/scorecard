// Copyright 2021 OpenSSF Scorecard Authors
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

// Package packagemanager implements a packagemanager client
package packagemanager

import (
	"fmt"
	"net/http"
	"time"
)

type Client interface {
	Get(URI string, packagename string) (*http.Response, error)

	GetURI(URI string) (*http.Response, error)
}

type PackageManagerClient struct{}

//nolint:noctx
func (c *PackageManagerClient) Get(url, packageName string) (*http.Response, error) {
	return c.getRemoteURL(fmt.Sprintf(url, packageName))
}

//nolint:noctx
func (c *PackageManagerClient) GetURI(url string) (*http.Response, error) {
	return c.getRemoteURL(url)
}

//nolint:noctx
func (c *PackageManagerClient) getRemoteURL(url string) (*http.Response, error) {
	const timeout = 10
	client := &http.Client{
		Timeout: timeout * time.Second,
	}
	//nolint:wrapcheck
	return client.Get(url)
}
