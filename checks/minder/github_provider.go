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

package minder

import (
	"context"
	"errors"
	"net/http"

	"github.com/ossf/scorecard/v5/clients"

	git "github.com/go-git/go-git/v5"
	minder "github.com/mindersec/minder/pkg/engine/v1/interfaces"
)

var errUnimplemented = errors.New("unimplemented")

// GitHubProvider implements the GitHub interface
type GitHubProvider struct {
	repo clients.RepoClient
}

var _ minder.Provider = (*GitHubProvider)(nil)
var _ minder.GitProvider = (*GitHubProvider)(nil)

// Clone implements v1.Git.
func (g *GitHubProvider) Clone(ctx context.Context, url string, branch string) (*git.Repository, error) {
	return repositoryFromClient(g.repo)
}

var _ minder.RESTProvider = (*GitHubProvider)(nil)

// Do implements v1.REST.
func (g *GitHubProvider) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	return nil, errUnimplemented
}

// GetBaseURL implements v1.REST.
func (g *GitHubProvider) GetBaseURL() string {
	return ""
}

// NewRequest implements v1.REST.
func (g *GitHubProvider) NewRequest(method string, url string, body any) (*http.Request, error) {
	return nil, errUnimplemented
}
