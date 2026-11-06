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
	"net/http"
	"strings"
	"sync"

	"github.com/google/go-github/v82/github"

	"github.com/ossf/scorecard/v5/clients"
	sce "github.com/ossf/scorecard/v5/errors"
)

const (
	ownerEndpointUser = "users"
	ownerEndpointOrg  = "orgs"

	// attestationLookBack is the number of most recent releases for which
	// attestations are looked up. It mirrors the look back window used by the
	// releasesHaveAttestation probe, so that no API calls are made for releases
	// that are never evaluated.
	attestationLookBack = 5
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
			return
		}
		handler.releases = releasesFrom(releases)
		// Only incur the extra API calls if there is something to look up.
		if handler.hasAttestationCandidates() {
			handler.ownerEndpointPrefix = handler.resolveOwnerEndpointPrefix()
			handler.checkAttestations()
		}
	})
	return handler.errSetup
}

// resolveOwnerEndpointPrefix determines whether the repo owner is a GitHub
// user or organization and returns the corresponding API path prefix.
// It calls the GitHub Users API once so that hasAttestation can use a single
// endpoint per asset instead of trying both.
// If the owner type cannot be determined, an empty prefix is returned and
// hasAttestation will try both endpoints.
func (handler *releasesHandler) resolveOwnerEndpointPrefix() string {
	user, _, err := handler.client.Users.Get(handler.ctx, handler.repourl.owner)
	if err != nil {
		return ""
	}
	if strings.EqualFold(user.GetType(), "Organization") {
		return ownerEndpointOrg
	}
	return ownerEndpointUser
}

// attestationReleases returns the number of releases to look up attestations for.
func (handler *releasesHandler) attestationReleases() int {
	return min(len(handler.releases), attestationLookBack)
}

// hasAttestationCandidates returns whether any release within the look back
// window has an asset with a digest that could have an attestation.
func (handler *releasesHandler) hasAttestationCandidates() bool {
	for i := range handler.attestationReleases() {
		for j := range handler.releases[i].Assets {
			if handler.releases[i].Assets[j].Digest != "" {
				return true
			}
		}
	}
	return false
}

func (handler *releasesHandler) checkAttestations() {
	for i := range handler.attestationReleases() {
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
	prefixes := []string{handler.ownerEndpointPrefix}
	if handler.ownerEndpointPrefix == "" {
		// The owner type is unknown, so try both endpoints.
		prefixes = []string{ownerEndpointUser, ownerEndpointOrg}
	}
	for _, prefix := range prefixes {
		found, notFound := handler.queryAttestation(prefix, digest)
		if !notFound {
			return found
		}
	}
	return false
}

// queryAttestation looks up attestations for the digest using the given owner
// endpoint prefix. notFound is true if the endpoint responded with a 404, which
// indicates the owner may be of the other type.
func (handler *releasesHandler) queryAttestation(prefix, digest string) (found, notFound bool) {
	endpoint := fmt.Sprintf("%s/%s/attestations/%s", prefix, handler.repourl.owner, digest)
	req, err := handler.client.NewRequest("GET", endpoint, nil)
	if err != nil {
		return false, false
	}
	var body attestationResponse
	resp, err := handler.client.Do(handler.ctx, req, &body)
	if err != nil {
		return false, resp != nil && resp.StatusCode == http.StatusNotFound
	}
	return len(body.Attestations) > 0, false
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
			IsImmutable:     r.GetImmutable(),
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
