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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

var errTooManyRequests = errors.New("failed after exponential backoff")

// httpClientCIIBestPractices implements the CIIBestPracticesClient interface.
// A HTTP client with exponential backoff is used to communicate with the CII Best Practices servers.
type httpClientCIIBestPractices struct {
	baseURL string
}

type expBackoffTransport struct {
	numRetries uint8
}

func (transport *expBackoffTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for i := 0; i < int(transport.numRetries); i++ {
		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusTooManyRequests {
			//nolint:wrapcheck
			return resp, err
		}
		time.Sleep(time.Duration(math.Pow(2, float64(i))) * time.Second)
	}
	return nil, errTooManyRequests
}

// GetBadgeLevel implements CIIBestPracticesClient.GetBadgeLevel.
func (client *httpClientCIIBestPractices) GetBadgeLevel(ctx context.Context, uri string) (BadgeLevel, error) {
	repoURI := fmt.Sprintf("https://%s", uri)
	url := fmt.Sprintf("%s/projects.json?url=%s", client.baseURL, repoURI)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Unknown, fmt.Errorf("error during http.NewRequestWithContext: %w", err)
	}

	httpClient := http.Client{
		Transport: &expBackoffTransport{
			numRetries: 3,
		},
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return Unknown, fmt.Errorf("error during http.Do: %w", err)
	}
	defer resp.Body.Close()

	jsonData, err := io.ReadAll(resp.Body)
	if err != nil {
		return Unknown, fmt.Errorf("error during io.ReadAll: %w", err)
	}

	parsedResponse, err := ParseBadgeResponseFromJSON(jsonData)
	if err != nil {
		return Unknown, fmt.Errorf("error during json parsing: %w", err)
	}
	if len(parsedResponse) < 1 {
		return NotFound, nil
	}
	return parsedResponse[0].getBadgeLevel()
}

// GetBadgeByProjectID implements CIIProjectClient.GetBadgeByProjectID.
// It calls /projects/{id}.json and returns the badge data for that project ID.
// A 404 response is treated as "not found" and returns an empty BadgeResponse with no error.
func (client *httpClientCIIBestPractices) GetBadgeByProjectID(ctx context.Context, id int) (BadgeResponse, error) {
	url := fmt.Sprintf("%s/projects/%d.json", client.baseURL, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return BadgeResponse{}, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}
	httpClient := http.Client{
		Transport: &expBackoffTransport{numRetries: 3},
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return BadgeResponse{}, fmt.Errorf("http.Do: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return BadgeResponse{}, nil
	}
	jsonData, err := io.ReadAll(resp.Body)
	if err != nil {
		return BadgeResponse{}, fmt.Errorf("io.ReadAll: %w", err)
	}
	var parsed BadgeResponse
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		return BadgeResponse{}, fmt.Errorf("json.Unmarshal: %w", err)
	}
	return parsed, nil
}
