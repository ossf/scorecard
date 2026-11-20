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

// Package host contains helpers for probes which depend on the repository host.
package host

import "os"

const defaultGitHubHost = "github.com"

// MetadataKey is the key of the repository host in the raw results metadata.
const MetadataKey = "repository.host"

// IsGitHub returns whether the repository host is GitHub (or the GitHub Enterprise
// host in GH_HOST). An unknown (empty) host is assumed to be GitHub.
func IsGitHub(host string) bool {
	if host == "" || host == defaultGitHubHost {
		return true
	}
	return host == os.Getenv("GH_HOST")
}
