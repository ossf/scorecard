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

package clients

import (
	"context"
)

// VulnerabilitiesClient checks for vulnerabilities in vuln DB.
type VulnerabilitiesClient interface {
	ListUnfixedVulnerabilities(
		context context.Context,
		commit string,
		localDir string,
	) (VulnerabilitiesResponse, error)
}

// DefaultVulnerabilitiesClient returns a new OSV Vulnerabilities client.
func DefaultVulnerabilitiesClient() VulnerabilitiesClient {
	return osvClient{local: false}
}

// ExperimentalLocalOSVClient returns an OSV Vulnerabilities client which
// takes advantage of their experimental local database option. As the
// osv-scanner feature is experimental, so is our usage of it. This function
// may be removed without warning.
//
// https://google.github.io/osv-scanner/experimental/offline-mode/#local-database-option
func ExperimentalLocalOSVClient() VulnerabilitiesClient {
	return osvClient{local: true}
}

// VulnerabilitiesResponse is the response from the vuln DB.
type VulnerabilitiesResponse struct {
	Vulnerabilities []Vulnerability
}

// Vulnerability uniquely identifies a reported security vuln.
type Vulnerability struct {
	ID      string
	Aliases []string
}
