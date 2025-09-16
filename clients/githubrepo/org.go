package githubrepo

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v60/github"

	"github.com/ossf/scorecard/v5/clients/githubrepo/roundtripper"
	"github.com/ossf/scorecard/v5/log"
)

// ListOrgRepos lists all non-archived repositories for a GitHub organization.
func ListOrgRepos(ctx context.Context, org string) ([]string, error) {
	// If org is a URL like "github.com/gabrielsoltz", extract just the org name.
	if len(org) > 0 {
		if parsed := parseOrgName(org); parsed != "" {
			org = parsed
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
		repos, resp, err := client.Repositories.ListByOrg(ctx, org, opt)
		if err != nil {
			return nil, fmt.Errorf("failed to list repos: %w", err)
		}

		for _, r := range repos {
			if r.GetArchived() {
				continue
			}
			urls = append(urls, r.GetHTMLURL())
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
