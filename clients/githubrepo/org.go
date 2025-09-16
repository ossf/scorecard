package githubrepo

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

// ListOrgRepos lists all non-archived repositories for a GitHub organization.
func ListOrgRepos(ctx context.Context, org string) ([]string, error) {
	// If org is a URL like "github.com/gabrielsoltz", extract just the org name.
	if len(org) > 0 {
		if parsed := parseOrgName(org); parsed != "" {
			org = parsed
		}
	}

	token := os.Getenv("GITHUB_AUTH_TOKEN")
	var tc *http.Client
	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc = oauth2.NewClient(ctx, ts)
	}

	client := github.NewClient(tc)

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
