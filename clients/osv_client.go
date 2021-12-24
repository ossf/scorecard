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

	sce "github.com/ossf/scorecard/v3/errors"
)

const osvQueryEndpoint = "https://api.osv.dev/v1/query"

type osvQuery struct {
	Commit string `json:"commit"`
}

// OSVResponse response from OSV.
type OSVResponse struct {
	Vulns []struct {
		ID string `json:"id"`
	} `json:"vulns"`
}

// OSVClient is a client for the OSV vulnerability database.
type OSVClient interface {
	// GetOSV returns the OSV response for the given commit.
	GetOSV(commitID string) (*OSVResponse, error)
}

type osvClient struct {
	ctx context.Context
}

// GetOSV returns the OSV response for the given commit.
func (c *osvClient) GetOSV(commitID string) (*OSVResponse, error) {
	query, err := json.Marshal(&osvQuery{
		Commit: commitID,
	})
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "json.Marshal")
		return nil, e
	}

	req, err := http.NewRequestWithContext(c.ctx, http.MethodPost, osvQueryEndpoint, bytes.NewReader(query))
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "http.NewRequestWithContext")
		return nil, e
	}

	// Use our own http client as the one from CheckRequest adds GitHub tokens to the headers.
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "httpClient.Do")
		return nil, e
	}
	defer resp.Body.Close()

	var osvResp OSVResponse
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&osvResp); err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "decoder.Decode")
		return nil, e
	}

	return &osvResp, nil
}

// NewOSVClient returns a new OSVClient.
func NewOSVClient(ctx context.Context) OSVClient {
	return &osvClient{
		ctx: ctx,
	}
}
