// Copyright 2025 OpenSSF Scorecard Authors
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

package org

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v82/github"

	"github.com/ossf/scorecard/v5/clients/githubrepo/roundtripper"
	"github.com/ossf/scorecard/v5/log"
)

// ErrNilResponse indicates the GitHub API returned a nil response object.
var ErrNilResponse = errors.New("nil response from GitHub API")

// ListOrgRepos lists all non-archived repositories for a GitHub organization.
// The caller should provide an http.RoundTripper (rt). If rt is nil, the
// default transport will be created via roundtripper.NewTransport.
func ListOrgRepos(ctx context.Context, orgName string, rt http.RoundTripper) ([]string, error) {
	// Parse org name if needed.
	if len(orgName) > 0 {
		if parsed := parseOrgName(orgName); parsed != "" {
			orgName = parsed
		}
	}

	// Use the centralized transport so we respect token rotation, GitHub App
	// auth, rate limiting and instrumentation already implemented in
	// clients/githubrepo/roundtripper.
	logger := log.NewLogger(log.DefaultLevel)
	if rt == nil {
		rt = roundtripper.NewTransport(ctx, logger)
	}
	httpClient := &http.Client{Transport: rt}
	client := github.NewClient(httpClient)

	opt := &github.RepositoryListByOrgOptions{
		Type: "all",
	}

	var urls []string
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, orgName, opt)
		if err != nil {
			return nil, fmt.Errorf("failed to list repos: %w", err)
		}

		for _, r := range repos {
			if r.GetArchived() {
				continue
			}
			urls = append(urls, r.GetHTMLURL())
		}

		if resp == nil {
			return nil, ErrNilResponse
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return urls, nil
}

// parseOrgName extracts the GitHub organization from a supported input.
// Supported:
//   - owner > owner
//   - github.com/owner > owner
//   - http://github.com/owner > owner
//   - https://github.com/owner > owner
//
// Returns "" if no org can be parsed.
func parseOrgName(input string) string {
	s := strings.TrimSpace(input)
	if s == "" {
		return ""
	}

	// Strip optional scheme.
	switch {
	case strings.HasPrefix(s, "https://"):
		s = strings.TrimPrefix(s, "https://")
	case strings.HasPrefix(s, "http://"):
		s = strings.TrimPrefix(s, "http://")
	}

	// If it's exactly the host, there's no org.
	if s == "github.com" {
		return ""
	}

	// Strip host prefix if present.
	if after, ok := strings.CutPrefix(s, "github.com/"); ok {
		s = after
	}

	// Keep only the first path segment (the org).
	if i := strings.IndexByte(s, '/'); i >= 0 {
		s = s[:i]
	}

	// Basic sanity: org shouldn't contain dots (to avoid host-like values).
	if s == "" || strings.Contains(s, ".") {
		return ""
	}

	return s
}
