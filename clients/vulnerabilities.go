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

package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/ossf/scorecard/v4/errors"
)

const osvQueryEndpoint = "https://api.osv.dev/v1/query"

type osvQuery struct {
	Commit string `json:"commit"`
}

// VulnerabilitiesClient cheks for vulnerabilities in api.osv.dev.
type VulnerabilitiesClient interface {
	HasUnfixedVulnerabilities(context context.Context, commit string) (VulnerabilitiesResponse, error)
}

// VulnerabilitiesResponse is the response from the OSV API.
type VulnerabilitiesResponse struct {
	Vulns []struct {
		ID string `json:"id"`
	} `json:"vulns"`
}

type vulns struct{}

// DefaultVulnerabilitiesClient is a new Vulnerabilities client.
func DefaultVulnerabilitiesClient() VulnerabilitiesClient {
	return vulns{}
}

// HasUnfixedVulnerabilities runs Vulnerabilities check.
func (v vulns) HasUnfixedVulnerabilities(ctx context.Context, commit string) (VulnerabilitiesResponse, error) {
	query, err := json.Marshal(&osvQuery{
		Commit: commit,
	})
	if err != nil {
		return VulnerabilitiesResponse{}, errors.WithMessage(err, "failed to marshal query")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, osvQueryEndpoint, bytes.NewReader(query))
	if err != nil {
		return VulnerabilitiesResponse{}, errors.WithMessage(err, "failed to create request")
	}

	// Use our own http client as the one from CheckRequest adds GitHub tokens to the headers.
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return VulnerabilitiesResponse{}, errors.WithMessage(err, "failed to send request")
	}
	defer resp.Body.Close()

	var osvResp VulnerabilitiesResponse
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&osvResp); err != nil {
		return VulnerabilitiesResponse{}, errors.WithMessage(err, "failed to decode response")
	}

	return osvResp, nil
}
