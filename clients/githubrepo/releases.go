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
	"errors"
	"fmt"
	"net/http"
	"slices"
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
	client   *github.Client
	once     *sync.Once
	ctx      context.Context
	errSetup error
	repourl  *Repo
	releases []clients.Release

	// attestedReleases is a copy of releases with attestation information,
	// populated on demand so that listing releases stays cheap for callers
	// that do not need it.
	attestOnce          *sync.Once
	attestedReleases    []clients.Release
	ownerEndpointPrefix string
}

func (handler *releasesHandler) init(ctx context.Context, repourl *Repo) {
	handler.ctx = ctx
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
	handler.attestOnce = new(sync.Once)
	handler.releases = nil
	handler.attestedReleases = nil
	handler.ownerEndpointPrefix = ""
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
	})
	return handler.errSetup
}

// setupAttestations enriches a copy of the listed releases with attestation information.
// This is separate from setup because it makes additional API calls per release asset,
// which callers that only need to list releases should not have to pay for.
func (handler *releasesHandler) setupAttestations() error {
	if err := handler.setup(); err != nil {
		return err
	}
	handler.attestOnce.Do(func() {
		// Copy so that the releases returned by getReleases are never modified,
		// as they may be in use by other checks running concurrently.
		handler.attestedReleases = copyReleases(handler.releases)
		// Only incur the extra API calls if there is something to look up.
		if hasAttestationCandidates(handler.attestedReleases) {
			handler.ownerEndpointPrefix = handler.resolveOwnerEndpointPrefix()
			handler.checkAttestations(handler.attestedReleases)
		}
	})
	return nil
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

// hasAttestationCandidates returns whether any release within the look back
// window has an asset with a digest that could have an attestation.
func hasAttestationCandidates(releases []clients.Release) bool {
	for i := range min(len(releases), attestationLookBack) {
		for j := range releases[i].Assets {
			if releases[i].Assets[j].Digest != "" {
				return true
			}
		}
	}
	return false
}

// checkAttestations populates the attestation fields of the assets of the
// releases within the look back window.
func (handler *releasesHandler) checkAttestations(releases []clients.Release) {
	for i := range min(len(releases), attestationLookBack) {
		for j := range releases[i].Assets {
			asset := &releases[i].Assets[j]
			if asset.Digest == "" {
				continue
			}
			found, err := handler.hasAttestation(asset.Digest)
			// A failed lookup is not evidence that there is no attestation.
			asset.HasAttestation = found
			asset.AttestationLookupFailed = err != nil
		}
	}
}

func copyReleases(releases []clients.Release) []clients.Release {
	if releases == nil {
		return nil
	}
	cp := make([]clients.Release, len(releases))
	for i := range releases {
		cp[i] = releases[i]
		cp[i].Assets = slices.Clone(releases[i].Assets)
	}
	return cp
}

type attestationResponse struct {
	Attestations []interface{} `json:"attestations"`
}

// errAttestationNotFound is returned when the attestations endpoint responds with a 404.
var errAttestationNotFound = errors.New("attestations not found")

// hasAttestation returns whether the digest has an attestation. A 404 response
// is treated as there being no attestation, but any other failure to query
// the attestations (e.g. rate limiting or server errors) is returned as an error.
func (handler *releasesHandler) hasAttestation(digest string) (bool, error) {
	prefixes := []string{handler.ownerEndpointPrefix}
	if handler.ownerEndpointPrefix == "" {
		// The owner type is unknown, so try both endpoints.
		prefixes = []string{ownerEndpointUser, ownerEndpointOrg}
	}
	for _, prefix := range prefixes {
		found, err := handler.queryAttestation(prefix, digest)
		if errors.Is(err, errAttestationNotFound) {
			// The owner may be of the other type.
			continue
		}
		return found, err
	}
	return false, nil
}

// queryAttestation looks up attestations for the digest using the given owner endpoint prefix.
func (handler *releasesHandler) queryAttestation(prefix, digest string) (bool, error) {
	endpoint := fmt.Sprintf("%s/%s/attestations/%s", prefix, handler.repourl.owner, digest)
	req, err := handler.client.NewRequest("GET", endpoint, nil)
	if err != nil {
		return false, fmt.Errorf("creating attestation request: %w", err)
	}
	var body attestationResponse
	resp, err := handler.client.Do(handler.ctx, req, &body)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return false, errAttestationNotFound
		}
		return false, fmt.Errorf("querying attestations: %w", err)
	}
	return len(body.Attestations) > 0, nil
}

func (handler *releasesHandler) getReleases() ([]clients.Release, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during graphqlHandler.setup: %w", err)
	}
	return handler.releases, nil
}

// getReleasesWithAttestations returns the releases along with attestation information.
func (handler *releasesHandler) getReleasesWithAttestations() ([]clients.Release, error) {
	if err := handler.setupAttestations(); err != nil {
		return nil, fmt.Errorf("error during releasesHandler.setupAttestations: %w", err)
	}
	return handler.attestedReleases, nil
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
