// Copyright 2022 OpenSSF Authors
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
//
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v42/github"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	kgh "sigs.k8s.io/release-sdk/github"
	"sigs.k8s.io/release-utils/env"
)

// From https://github.com/kubernetes-sigs/release-sdk/blob/e23d2c82bbb41a007cdf019c30930e8fd2649c01/github/github.go

// GitHub is a wrapper around GitHub related functionality.
type GitHub struct {
	client  Client
	options *Options
}

// Client is an interface modeling supported GitHub operations.
type Client interface {
	// TODO(install): Populate interface
	CreateFile(
		context.Context, string, string, string, *github.RepositoryContentFileOptions,
	) (*github.RepositoryContentResponse, *github.Response, error)
	CreateGitRef(
		context.Context, string, string, *github.Reference,
	) (*github.Reference, *github.Response, error)
	CreatePullRequest(
		context.Context, string, string, string, string, string, string,
	) (*github.PullRequest, error)
	GetBranch(
		context.Context, string, string, string, bool,
	) (*github.Branch, *github.Response, error)
	GetContents(
		context.Context, string, string, string, *github.RepositoryContentGetOptions,
	) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error)
	GetRepositoriesByOrg(
		context.Context, string,
	) ([]*github.Repository, *github.Response, error)
	GetRepository(
		context.Context, string, string,
	) (*github.Repository, *github.Response, error)
}

// Options is a set of options to configure the behavior of the GitHub package.
type Options struct {
	// How many items to request in calls to the github API
	// that require pagination.
	ItemsPerPage int
}

// GetItemsPerPage // TODO(github): needs comment.
func (o *Options) GetItemsPerPage() int {
	return o.ItemsPerPage
}

// DefaultOptions return an options struct with commonly used settings.
func DefaultOptions() *Options {
	return &Options{
		ItemsPerPage: 50,
	}
}

// SetClient can be used to manually set the internal GitHub client.
func (g *GitHub) SetClient(client Client) {
	g.client = client
}

// Client can be used to retrieve the Client type.
func (g *GitHub) Client() Client {
	return g.client
}

// SetOptions gets an options set for the GitHub object.
func (g *GitHub) SetOptions(opts *Options) {
	g.options = opts
}

// Options return a pointer to the options struct.
func (g *GitHub) Options() *Options {
	return g.options
}

// TODO: we should clean up the functions listed below and agree on the same
// return type (with or without error):
// - New
// - NewWithToken
// - NewEnterprise
// - NewEnterpriseWithToken

// New creates a new default GitHub client. Tokens set via the $GITHUB_TOKEN
// environment variable will result in an authenticated client.
// If the $GITHUB_TOKEN is not set, then the client will do unauthenticated
// GitHub requests.
func New() *GitHub {
	token := env.Default(kgh.TokenEnvKey, "")
	client, _ := NewWithToken(token) // nolint: errcheck
	return client
}

// NewWithToken can be used to specify a GitHub token through parameters.
// Empty string will result in unauthenticated client, which makes
// unauthenticated requests.
func NewWithToken(token string) (*GitHub, error) {
	ctx := context.Background()
	client := http.DefaultClient
	state := "unauthenticated"
	if token != "" {
		state = strings.TrimPrefix(state, "un")
		client = oauth2.NewClient(ctx, oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		))
	}
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		logrus.Infof("Unable to retrieve user cache dir: %v", err)
		cacheDir = os.TempDir()
	}
	dir := filepath.Join(cacheDir, "kubernetes", "release-sdk", "github")
	logrus.Debugf("Caching GitHub responses in %v", dir)
	t := httpcache.NewTransport(diskcache.New(dir))
	client.Transport = t.Transport

	logrus.Debugf("Using %s GitHub client", state)
	return &GitHub{
		client:  &githubClient{github.NewClient(client)},
		options: DefaultOptions(),
	}, nil
}

// NewEnterprise // TODO(github): needs comment.
func NewEnterprise(baseURL, uploadURL string) (*GitHub, error) {
	token := env.Default(kgh.TokenEnvKey, "")
	return NewEnterpriseWithToken(baseURL, uploadURL, token)
}

// NewEnterpriseWithToken // TODO(github): needs comment.
func NewEnterpriseWithToken(baseURL, uploadURL, token string) (*GitHub, error) {
	ctx := context.Background()
	client := http.DefaultClient
	state := "unauthenticated"
	if token != "" {
		state = strings.TrimPrefix(state, "un")
		client = oauth2.NewClient(ctx, oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		))
	}
	logrus.Debugf("Using %s Enterprise GitHub client", state)
	ghclient, err := github.NewEnterpriseClient(baseURL, uploadURL, client)
	if err != nil {
		return nil, fmt.Errorf("failed to new github client: %w", err)
	}
	return &GitHub{
		client:  &githubClient{ghclient},
		options: DefaultOptions(),
	}, nil
}

type githubClient struct {
	*github.Client
}

func (g *githubClient) GetRepositoriesByOrg(
	ctx context.Context,
	owner string,
) ([]*github.Repository, *github.Response, error) {
	repos, resp, err := g.Repositories.ListByOrg(
		ctx,
		owner,
		// TODO(install): Does this need to parameterized?
		&github.RepositoryListByOrgOptions{
			Type: "all",
		},
	)
	if err != nil {
		return repos, resp, fmt.Errorf("getting repositories: %w", err)
	}

	return repos, resp, nil
}

func (g *githubClient) GetRepository(
	ctx context.Context,
	owner,
	repo string,
) (*github.Repository, *github.Response, error) {
	pr, resp, err := g.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return pr, resp, fmt.Errorf("getting repository: %w", err)
	}

	return pr, resp, nil
}

func (g *githubClient) GetBranch(
	ctx context.Context,
	owner,
	repo,
	branch string,
	followRedirects bool,
) (*github.Branch, *github.Response, error) {
	// TODO: Revisit logic and simplify returns, where possible.
	b, resp, err := g.Repositories.GetBranch(
		ctx,
		owner,
		repo,
		branch,
		followRedirects,
	)
	if err != nil {
		return b, resp, fmt.Errorf("getting branch: %w", err)
	}

	return b, resp, nil
}

func (g *githubClient) GetContents(
	ctx context.Context,
	owner,
	repo,
	path string,
	opts *github.RepositoryContentGetOptions,
) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error) {
	// TODO: Revisit logic and simplify returns, where possible.
	file, dir, resp, err := g.Repositories.GetContents(
		ctx,
		owner,
		repo,
		path,
		opts,
	)
	if err != nil {
		return file, dir, resp, fmt.Errorf("getting repo content: %w", err)
	}

	return file, dir, resp, nil
}

func (g *githubClient) CreateGitRef(
	ctx context.Context,
	owner,
	repo string,
	ref *github.Reference,
) (*github.Reference, *github.Response, error) {
	// TODO: Revisit logic and simplify returns, where possible.
	gRef, resp, err := g.Git.CreateRef(
		ctx,
		owner,
		repo,
		ref,
	)
	if err != nil {
		return gRef, resp, fmt.Errorf("creating git reference: %w", err)
	}

	return gRef, resp, nil
}

func (g *githubClient) CreateFile(
	ctx context.Context,
	owner,
	repo,
	path string,
	opts *github.RepositoryContentFileOptions,
) (*github.RepositoryContentResponse, *github.Response, error) {
	// TODO: Revisit logic and simplify returns, where possible.
	repoContentResp, resp, err := g.Repositories.CreateFile(
		ctx,
		owner,
		repo,
		path,
		opts,
	)
	if err != nil {
		return repoContentResp, resp, fmt.Errorf("creating file: %w", err)
	}

	return repoContentResp, resp, nil
}

func (g *githubClient) CreatePullRequest(
	ctx context.Context,
	owner,
	repo,
	baseBranchName,
	headBranchName,
	title,
	body string,
) (*github.PullRequest, error) {
	newPullRequest := &github.NewPullRequest{
		Title:               &title,
		Head:                &headBranchName,
		Base:                &baseBranchName,
		Body:                &body,
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := g.PullRequests.Create(ctx, owner, repo, newPullRequest)
	if err != nil {
		return pr, fmt.Errorf("creating pull request: %w", err)
	}

	logrus.Infof("Successfully created PR #%d", pr.GetNumber())
	return pr, nil
}
