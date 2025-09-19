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

	"github.com/google/go-github/v53/github"

	"github.com/ossf/scorecard/v5/clients/githubrepo/roundtripper"
	"github.com/ossf/scorecard/v5/log"
)

// ErrNilResponse indicates the GitHub API returned a nil response object.
var ErrNilResponse = errors.New("nil response from GitHub API")

// ListOrgRepos lists all non-archived repositories for a GitHub organization.
func ListOrgRepos(ctx context.Context, orgName string) ([]string, error) {
	// If org is a URL like "github.com/gabrielsoltz", extract just the org name.
	if len(orgName) > 0 {
		if parsed := parseOrgName(orgName); parsed != "" {
			orgName = parsed
		}
	}

	// Use the centralized transport so we respect token rotation, GitHub App
	// auth, rate limiting and instrumentation already implemented in
	// clients/githubrepo/roundtripper.
	logger := log.NewLogger(log.DefaultLevel)
	rt := roundtripper.TransportFactory(ctx, logger)
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

// parseOrgName extracts the organization name from a GitHub URL or returns the input if already an org name.
func parseOrgName(input string) string {
	// Remove "github.com/" prefix if present
	const prefix = "github.com/"
	if len(input) > len(prefix) && input[:len(prefix)] == prefix {
		return input[len(prefix):]
	}
	return input
}
