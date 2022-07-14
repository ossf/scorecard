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

var _ VulnerabilitiesClient = osvClient{}

type osvClient struct{}

const osvQueryEndpoint = "https://api.osv.dev/v1/query"

type osvQuery struct {
	Commit string `json:"commit"`
}

type osvResp struct {
	Vulns []struct {
		ID string `json:"id"`
	} `json:"vulns"`
}

// HasUnfixedVulnerabilities implements VulnerabilityClient.HasUnfixedVulnerabilities.
func (v osvClient) HasUnfixedVulnerabilities(ctx context.Context, commit string) (VulnerabilitiesResponse, error) {
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

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return VulnerabilitiesResponse{}, errors.WithMessage(err, "failed to send request")
	}
	defer resp.Body.Close()

	var osvresp osvResp
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&osvresp); err != nil {
		return VulnerabilitiesResponse{}, errors.WithMessage(err, "failed to decode response")
	}

	var ret VulnerabilitiesResponse
	for _, vuln := range osvresp.Vulns {
		ret.Vulnerabilities = append(ret.Vulnerabilities, Vulnerability{
			ID: vuln.ID,
		})
	}
	return ret, nil
}
