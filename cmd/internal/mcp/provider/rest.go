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

package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"sigs.k8s.io/release-utils/version"

	"github.com/ossf/scorecard/v5/cmd/internal/mcp/model"
)

// DefaultRESTBaseURL is the public cached OpenSSF Scorecard REST API.
const DefaultRESTBaseURL = "https://api.scorecard.dev"

// Sentinel errors returned for non-200 responses.
var (
	errBadRequest       = errors.New("invalid Scorecard REST API request")
	errUnexpectedStatus = errors.New("unexpected Scorecard REST API status")
)

// userAgent identifies requests from this server to the REST API operators.
var userAgent = "scorecard/" + version.GetVersionInfo().GitVersion

var cachedCaveats = []string{
	"The Scorecard REST API serves results that projects publish with publish_results: true " +
		"and results from Scorecard's weekly scan of public repositories.",
	"Scorecard results are heuristic signals, not a guarantee of security; there are false positives and false negatives.",
}

// CachedRESTProvider reads pre-computed results from the public Scorecard REST API.
type CachedRESTProvider struct {
	Client  *http.Client
	BaseURL string
}

// Compile-time check that CachedRESTProvider implements Provider.
var _ Provider = (*CachedRESTProvider)(nil)

// NewCachedREST returns a CachedRESTProvider. If baseURL is empty, the public
// API (DefaultRESTBaseURL) is used.
func NewCachedREST(baseURL string) *CachedRESTProvider {
	if baseURL == "" {
		baseURL = DefaultRESTBaseURL
	}
	return &CachedRESTProvider{
		BaseURL: strings.TrimSuffix(baseURL, "/"),
		Client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// restResult mirrors the api.scorecard.dev getResult response body.
type restResult struct {
	Repo struct {
		Name   string `json:"name"`
		Commit string `json:"commit"`
	} `json:"repo"`
	Scorecard struct {
		Version string `json:"version"`
		Commit  string `json:"commit"`
	} `json:"scorecard"`
	Date   string `json:"date"`
	Checks []struct {
		Documentation *struct {
			Short string `json:"short"`
			URL   string `json:"url"`
		} `json:"documentation"`
		Name    string   `json:"name"`
		Reason  string   `json:"reason"`
		Details []string `json:"details"`
		Score   int      `json:"score"`
	} `json:"checks"`
	Score float64 `json:"score"`
}

// GetResult implements Provider by querying the cached REST API.
func (p *CachedRESTProvider) GetResult(ctx context.Context, ref model.RepoRef, commit string) (*model.Result, error) {
	endpoint := fmt.Sprintf("%s/projects/%s/%s/%s",
		p.BaseURL,
		url.PathEscape(ref.Platform),
		url.PathEscape(ref.Org),
		url.PathEscape(ref.Name),
	)
	if commit != "" {
		endpoint += "?commit=" + url.QueryEscape(commit)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requesting Scorecard results for %s: %w", ref.String(), err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		// proceed
	case http.StatusNotFound:
		return nil, &NotFoundError{Ref: ref}
	case http.StatusBadRequest:
		return nil, fmt.Errorf("%w for %s (HTTP 400); check the repository reference and commit", errBadRequest, ref.String())
	default:
		return nil, fmt.Errorf("%w %d for %s", errUnexpectedStatus, resp.StatusCode, ref.String())
	}

	var rr restResult
	if err := json.NewDecoder(resp.Body).Decode(&rr); err != nil {
		return nil, fmt.Errorf("decoding Scorecard results for %s: %w", ref.String(), err)
	}
	return p.toResult(ref, commit, &rr), nil
}

func (p *CachedRESTProvider) toResult(ref model.RepoRef, commit string, rr *restResult) *model.Result {
	checks := make([]model.Check, 0, len(rr.Checks))
	for _, c := range rr.Checks {
		mc := model.Check{
			Name:    c.Name,
			Score:   c.Score,
			Reason:  c.Reason,
			Details: c.Details,
		}
		if c.Documentation != nil {
			mc.Documentation = &model.Documentation{Short: c.Documentation.Short, URL: c.Documentation.URL}
		}
		checks = append(checks, mc)
	}

	resolvedCommit := rr.Repo.Commit
	if resolvedCommit == "" {
		resolvedCommit = commit
	}

	return &model.Result{
		Repo:      ref,
		Commit:    resolvedCommit,
		Date:      rr.Date,
		Scorecard: model.ScorecardInfo{Version: rr.Scorecard.Version, Commit: rr.Scorecard.Commit},
		Source:    model.SourceCachedREST,
		Score:     rr.Score,
		Checks:    checks,
		Caveats:   cachedCaveats,
		Attribution: &model.Attribution{
			DataLicense: "CDLA-Permissive-2.0",
			SourceURL:   p.BaseURL,
		},
	}
}
