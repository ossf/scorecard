// Copyright 2025 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package packageclient

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func TestGetPackage(t *testing.T) {
	t.Parallel()
	resp := `{
  "packageKey": {
    "system": "NPM",
    "name": "@colors/colors"
  },
  "purl": "pkg:npm/%40colors/colors",
  "versions": [
    {
      "versionKey": {
        "system": "NPM",
        "name": "@colors/colors",
        "version": "1.4.0"
      },
      "purl": "pkg:npm/%40colors/colors@1.4.0",
      "publishedAt": "2022-02-12T06:40:43Z",
      "isDefault": false,
      "isDeprecated": false
    },
    {
      "versionKey": {
        "system": "NPM",
        "name": "@colors/colors",
        "version": "1.5.0"
      },
      "purl": "pkg:npm/%40colors/colors@1.5.0",
      "publishedAt": "2022-02-12T07:39:04Z",
      "isDefault": false,
      "isDeprecated": false
    },
    {
      "versionKey": {
        "system": "NPM",
        "name": "@colors/colors",
        "version": "1.6.0"
      },
      "purl": "pkg:npm/%40colors/colors@1.6.0",
      "publishedAt": "2023-07-10T05:16:15Z",
      "isDefault": true,
      "isDeprecated": false
    }
  ]
}`
	expected := &Package{
		PackageKey: PackageKey{
			PackageSystem: "NPM",
			PackageName:   "@colors/colors",
		},
		Purl: "pkg:npm/%40colors/colors",
		Versions: []Version{
			{
				VersionKey: VersionKey{
					System:  "NPM",
					Name:    "@colors/colors",
					Version: "1.4.0",
				},
				Purl:         "pkg:npm/%40colors/colors@1.4.0",
				PublishedAt:  "2022-02-12T06:40:43Z",
				IsDefault:    false,
				IsDeprecated: false,
			},
			{
				VersionKey: VersionKey{
					System:  "NPM",
					Name:    "@colors/colors",
					Version: "1.5.0",
				},
				Purl:         "pkg:npm/%40colors/colors@1.5.0",
				PublishedAt:  "2022-02-12T07:39:04Z",
				IsDefault:    false,
				IsDeprecated: false,
			},
			{
				VersionKey: VersionKey{
					System:  "NPM",
					Name:    "@colors/colors",
					Version: "1.6.0",
				},
				Purl:         "pkg:npm/%40colors/colors@1.6.0",
				PublishedAt:  "2023-07-10T05:16:15Z",
				IsDefault:    true,
				IsDeprecated: false,
			},
		},
	}

	// Create test client that does not call the deps.dev api
	testClient := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			// Send response to be tested
			Body: io.NopCloser(bytes.NewBufferString(resp)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})
	depsClient := depsDevClient{
		client: testClient,
	}
	depsClient.client = testClient

	r, err := depsClient.GetPackage(context.Background(), "npm", "@colors/colors")
	if err != nil {
		t.Errorf("GetPackage returned an error: %v", err)
	}
	if diff := cmp.Diff(r, expected); diff != "" {
		t.Errorf("setDefaultCommitData() mismatch (-want +got):\n%s", diff)
	}
}
