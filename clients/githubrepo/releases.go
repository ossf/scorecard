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

package githubrepo

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/google/go-github/v82/github"

	"github.com/ossf/scorecard/v5/clients"
	sce "github.com/ossf/scorecard/v5/errors"
)

const (
	ownerEndpointUser = "users"
	ownerEndpointOrg  = "orgs"
)

type releasesHandler struct {
	client              *github.Client
	once                *sync.Once
	ctx                 context.Context
	errSetup            error
	repourl             *Repo
	releases            []clients.Release
	ownerEndpointPrefix string
}

func (handler *releasesHandler) init(ctx context.Context, repourl *Repo) {
	handler.ctx = ctx
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
	handler.releases = nil
}

func (handler *releasesHandler) setup() error {
	handler.once.Do(func() {
		if !strings.EqualFold(handler.repourl.commitSHA, clients.HeadSHA) {
			handler.errSetup = fmt.Errorf("%w: ListReleases only supported for HEAD queries", clients.ErrUnsupportedFeature)
			return
		}
		releases, _, err := handler.client.Repositories.ListReleases(
			handler.ctx, handler.repourl.owner, handler.repourl.repo, &github.ListOptions{})
		if err != nil {
			handler.errSetup = sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("githubv4.Query: %v", err))
		}
		handler.releases = releasesFrom(releases)
		handler.ownerEndpointPrefix = handler.resolveOwnerEndpointPrefix()
		handler.checkAttestations()
	})
	return handler.errSetup
}

// resolveOwnerEndpointPrefix determines whether the repo owner is a GitHub
// user or organization and returns the corresponding API path prefix.
// It calls the GitHub Users API once so that hasAttestation can use a single
// endpoint per asset instead of trying both.
func (handler *releasesHandler) resolveOwnerEndpointPrefix() string {
	user, _, err := handler.client.Users.Get(handler.ctx, handler.repourl.owner)
	if err != nil {
		// Fall back to users; hasAttestation will skip on 404.
		return ownerEndpointUser
	}
	if strings.EqualFold(user.GetType(), "Organization") {
		return ownerEndpointOrg
	}
	return ownerEndpointUser
}

func (handler *releasesHandler) checkAttestations() {
	for i := range handler.releases {
		for j := range handler.releases[i].Assets {
			asset := &handler.releases[i].Assets[j]
			if asset.Digest == "" {
				continue
			}
			asset.HasAttestation = handler.hasAttestation(asset.Digest)
		}
	}
}

type attestationResponse struct {
	Attestations []interface{} `json:"attestations"`
}

func (handler *releasesHandler) hasAttestation(digest string) bool {
	endpoint := fmt.Sprintf("%s/%s/attestations/%s", handler.ownerEndpointPrefix, handler.repourl.owner, digest)
	req, err := handler.client.NewRequest("GET", endpoint, nil)
	if err != nil {
		return false
	}
	var body attestationResponse
	_, err = handler.client.Do(handler.ctx, req, &body)
	if err != nil {
		return false
	}
	return len(body.Attestations) > 0
}

func (handler *releasesHandler) getReleases() ([]clients.Release, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during graphqlHandler.setup: %w", err)
	}
	return handler.releases, nil
}

func releasesFrom(data []*github.RepositoryRelease) []clients.Release {
	var releases []clients.Release
	for _, r := range data {
		release := clients.Release{
			TagName:         r.GetTagName(),
			URL:             r.GetURL(),
			TargetCommitish: r.GetTargetCommitish(),
		}
		for _, a := range r.Assets {
			release.Assets = append(release.Assets, clients.ReleaseAsset{
				Name:   a.GetName(),
				URL:    r.GetHTMLURL(),
				Digest: a.GetDigest(),
			})
		}
		releases = append(releases, release)
	}
	return releases
}
