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

package githubrepo

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-github/v82/github"

	"github.com/ossf/scorecard/v5/clients"
)

type releasesRoundTripper struct {
	attestedDigests map[string]bool
	userType        string
}

func (r *releasesRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	path := req.URL.Path

	// Handle GET /users/{owner} — owner type lookup.
	if strings.HasPrefix(path, "/users/") && strings.Count(path, "/") == 2 {
		userType := r.userType
		if userType == "" {
			userType = "User"
		}
		login := strings.TrimPrefix(path, "/users/")
		body := map[string]string{"login": login, "type": userType}
		return jsonRespOK(body)
	}

	// Handle GET /users/{owner}/attestations/{digest} or /orgs/{owner}/attestations/{digest}.
	if strings.Contains(path, "/attestations/") {
		parts := strings.SplitN(path, "/attestations/", 2)
		if len(parts) == 2 {
			digest := parts[1]
			var attestations []interface{}
			if r.attestedDigests[digest] {
				attestations = []interface{}{map[string]string{"id": "1"}}
			}
			return jsonRespOK(map[string]interface{}{"attestations": attestations})
		}
	}

	return &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     http.Header{},
	}, nil
}

// errorRoundTripper always returns a 500 error response.
type errorRoundTripper struct{}

func (e *errorRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     http.Header{},
	}, nil
}

func jsonRespOK(body interface{}) (*http.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       io.NopCloser(bytes.NewReader(data)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}, nil
}

func TestReleasesFrom(t *testing.T) {
	t.Parallel()
	tagName := "v1.0"
	url := "https://api.github.com/repos/owner/repo/releases/1"
	htmlURL := "https://github.com/owner/repo/releases/tag/v1.0"
	commitish := "main"
	assetName := "binary.tar.gz"
	digest := "sha256:abc123"

	releases := releasesFrom([]*github.RepositoryRelease{
		{
			TagName:         &tagName,
			URL:             &url,
			TargetCommitish: &commitish,
			HTMLURL:         &htmlURL,
			Assets: []*github.ReleaseAsset{
				{
					Name:   &assetName,
					Digest: &digest,
				},
			},
		},
	})

	if len(releases) != 1 {
		t.Fatalf("expected 1 release, got %d", len(releases))
	}
	r := releases[0]
	if r.TagName != tagName {
		t.Errorf("expected TagName %q, got %q", tagName, r.TagName)
	}
	if r.URL != url {
		t.Errorf("expected URL %q, got %q", url, r.URL)
	}
	if r.TargetCommitish != commitish {
		t.Errorf("expected TargetCommitish %q, got %q", commitish, r.TargetCommitish)
	}
	if len(r.Assets) != 1 {
		t.Fatalf("expected 1 asset, got %d", len(r.Assets))
	}
	if r.Assets[0].Name != assetName {
		t.Errorf("expected asset Name %q, got %q", assetName, r.Assets[0].Name)
	}
	if r.Assets[0].URL != htmlURL {
		t.Errorf("expected asset URL %q, got %q", htmlURL, r.Assets[0].URL)
	}
	if r.Assets[0].Digest != digest {
		t.Errorf("expected asset Digest %q, got %q", digest, r.Assets[0].Digest)
	}
}

func TestReleasesFrom_NilDigest(t *testing.T) {
	t.Parallel()
	tagName := "v1.0"
	assetName := "binary.tar.gz"
	releases := releasesFrom([]*github.RepositoryRelease{
		{
			TagName: &tagName,
			Assets:  []*github.ReleaseAsset{{Name: &assetName}},
		},
	})
	if len(releases) != 1 {
		t.Fatalf("expected 1 release, got %d", len(releases))
	}
	if releases[0].Assets[0].Digest != "" {
		t.Errorf("expected empty Digest for nil pointer, got %q", releases[0].Assets[0].Digest)
	}
}

func TestReleasesFrom_Empty(t *testing.T) {
	t.Parallel()
	if releases := releasesFrom(nil); releases != nil {
		t.Errorf("expected nil releases for nil input, got %v", releases)
	}
}

func TestResolveOwnerEndpointPrefix_User(t *testing.T) {
	t.Parallel()
	rt := &releasesRoundTripper{userType: "User"}
	handler := &releasesHandler{
		client:  github.NewClient(&http.Client{Transport: rt}),
		ctx:     t.Context(),
		repourl: &Repo{owner: "testuser", repo: "repo"},
	}
	if got := handler.resolveOwnerEndpointPrefix(); got != ownerEndpointUser {
		t.Errorf("expected %q, got %q", ownerEndpointUser, got)
	}
}

func TestResolveOwnerEndpointPrefix_Org(t *testing.T) {
	t.Parallel()
	rt := &releasesRoundTripper{userType: "Organization"}
	handler := &releasesHandler{
		client:  github.NewClient(&http.Client{Transport: rt}),
		ctx:     t.Context(),
		repourl: &Repo{owner: "testorg", repo: "repo"},
	}
	if got := handler.resolveOwnerEndpointPrefix(); got != ownerEndpointOrg {
		t.Errorf("expected %q, got %q", ownerEndpointOrg, got)
	}
}

func TestResolveOwnerEndpointPrefix_Error(t *testing.T) {
	t.Parallel()
	handler := &releasesHandler{
		client:  github.NewClient(&http.Client{Transport: &errorRoundTripper{}}),
		ctx:     t.Context(),
		repourl: &Repo{owner: "anyowner", repo: "repo"},
	}
	if got := handler.resolveOwnerEndpointPrefix(); got != ownerEndpointUser {
		t.Errorf("expected fallback to %q, got %q", ownerEndpointUser, got)
	}
}

func TestHasAttestation_Found(t *testing.T) {
	t.Parallel()
	const digest = "sha256:abc123"
	rt := &releasesRoundTripper{attestedDigests: map[string]bool{digest: true}}
	handler := &releasesHandler{
		client:              github.NewClient(&http.Client{Transport: rt}),
		ctx:                 t.Context(),
		repourl:             &Repo{owner: "testuser", repo: "repo"},
		ownerEndpointPrefix: ownerEndpointUser,
	}
	if !handler.hasAttestation(digest) {
		t.Error("expected attestation to be found")
	}
}

func TestHasAttestation_NotFound(t *testing.T) {
	t.Parallel()
	rt := &releasesRoundTripper{attestedDigests: map[string]bool{}}
	handler := &releasesHandler{
		client:              github.NewClient(&http.Client{Transport: rt}),
		ctx:                 t.Context(),
		repourl:             &Repo{owner: "testuser", repo: "repo"},
		ownerEndpointPrefix: ownerEndpointUser,
	}
	if handler.hasAttestation("sha256:notfound") {
		t.Error("expected attestation not to be found")
	}
}

func TestHasAttestation_RequestError(t *testing.T) {
	t.Parallel()
	handler := &releasesHandler{
		client:              github.NewClient(&http.Client{Transport: &errorRoundTripper{}}),
		ctx:                 t.Context(),
		repourl:             &Repo{owner: "testuser", repo: "repo"},
		ownerEndpointPrefix: ownerEndpointUser,
	}
	if handler.hasAttestation("sha256:any") {
		t.Error("expected false on request error")
	}
}

func TestCheckAttestations(t *testing.T) {
	t.Parallel()
	const attestedDigest = "sha256:attested"
	rt := &releasesRoundTripper{attestedDigests: map[string]bool{attestedDigest: true}}
	handler := &releasesHandler{
		client:              github.NewClient(&http.Client{Transport: rt}),
		ctx:                 t.Context(),
		repourl:             &Repo{owner: "testuser", repo: "repo"},
		ownerEndpointPrefix: ownerEndpointUser,
		releases: []clients.Release{
			{
				TagName: "v1.0",
				Assets: []clients.ReleaseAsset{
					{Name: "attested.tar.gz", Digest: attestedDigest},
					{Name: "no-digest.tar.gz", Digest: ""},
					{Name: "not-attested.tar.gz", Digest: "sha256:other"},
				},
			},
		},
	}
	handler.checkAttestations()

	assets := handler.releases[0].Assets
	if !assets[0].HasAttestation {
		t.Errorf("expected asset %q to have attestation", assets[0].Name)
	}
	if assets[1].HasAttestation {
		t.Errorf("expected asset %q (no digest) NOT to have attestation", assets[1].Name)
	}
	if assets[2].HasAttestation {
		t.Errorf("expected asset %q NOT to have attestation", assets[2].Name)
	}
}
