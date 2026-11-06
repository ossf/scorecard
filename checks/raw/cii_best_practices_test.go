// Copyright 2026 OpenSSF Scorecard Authors
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

package raw

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
)

// fakeCIIClient implements both CIIBestPracticesClient and CIIProjectClient.
type fakeCIIClient struct {
	directErr   error
	projectErr  error
	projectResp clients.BadgeResponse
	directLevel clients.BadgeLevel
}

func (f *fakeCIIClient) GetBadgeLevel(_ context.Context, _ string) (clients.BadgeLevel, error) {
	return f.directLevel, f.directErr
}

func (f *fakeCIIClient) GetBadgeByProjectID(_ context.Context, _ int) (clients.BadgeResponse, error) {
	return f.projectResp, f.projectErr
}

// blobFakeCIIClient implements only CIIBestPracticesClient (simulates blob/scorecard.dev mode).
type blobFakeCIIClient struct{ level clients.BadgeLevel }

func (b *blobFakeCIIClient) GetBadgeLevel(_ context.Context, _ string) (clients.BadgeLevel, error) {
	return b.level, nil
}

// fakeRepoClient embeds clients.RepoClient so only ListFiles and GetFileReader need to be implemented.
type fakeRepoClient struct {
	clients.RepoClient
	files map[string]string
}

func (f *fakeRepoClient) ListFiles(predicate func(string) (bool, error)) ([]string, error) {
	var result []string
	for name := range f.files {
		ok, err := predicate(name)
		if err != nil {
			return nil, err
		}
		if ok {
			result = append(result, name)
		}
	}
	return result, nil
}

func (f *fakeRepoClient) GetFileReader(filename string) (io.ReadCloser, error) {
	content, ok := f.files[filename]
	if !ok {
		return nil, errors.New("file not found")
	}
	return io.NopCloser(strings.NewReader(content)), nil
}

// fakeRepo embeds clients.Repo so only URI needs to be implemented.
type fakeRepo struct {
	clients.Repo
	uri string
}

func (r fakeRepo) URI() string { return r.uri }

func TestCIIBestPractices(t *testing.T) {
	t.Parallel()
	tests := []struct {
		ciiClient     clients.CIIBestPracticesClient
		readmeContent map[string]string
		name          string
		repoURI       string
		wantInherited string
		wantBadge     clients.BadgeLevel
		wantErr       bool
	}{
		{
			name:      "direct hit returns badge without fallback",
			ciiClient: &fakeCIIClient{directLevel: clients.Gold},
			repoURI:   "github.com/ossf/scorecard",
			wantBadge: clients.Gold,
		},
		{
			name:    "nil client returns error",
			wantErr: true,
		},
		{
			name:      "blob mode: no CIIProjectClient, returns NotFound",
			ciiClient: &blobFakeCIIClient{level: clients.NotFound},
			repoURI:   "github.com/ossf/scorecard",
			wantBadge: clients.NotFound,
		},
		{
			name: "fallback hit: sibling repo in same org has badge",
			ciiClient: &fakeCIIClient{
				directLevel: clients.NotFound,
				projectResp: clients.BadgeResponse{
					BadgeLevel: "gold",
					ID:         1234,
					RepoURL:    "https://github.com/ossf/allstar",
				},
			},
			readmeContent: map[string]string{
				"README.md": "[![badge](https://www.bestpractices.dev/projects/1234/badge)](https://www.bestpractices.dev/projects/1234)",
			},
			repoURI:       "github.com/ossf/scorecard",
			wantBadge:     clients.Gold,
			wantInherited: "https://github.com/ossf/allstar",
		},
		{
			name: "fallback: ownership mismatch is rejected",
			ciiClient: &fakeCIIClient{
				directLevel: clients.NotFound,
				projectResp: clients.BadgeResponse{
					BadgeLevel: "gold",
					ID:         999,
					RepoURL:    "https://github.com/evil-org/repo",
				},
			},
			readmeContent: map[string]string{
				"README.md": "https://bestpractices.dev/projects/999",
			},
			repoURI:   "github.com/ossf/scorecard",
			wantBadge: clients.NotFound,
		},
		{
			name: "fallback: network error does not propagate",
			ciiClient: &fakeCIIClient{
				directLevel: clients.NotFound,
				projectErr:  errors.New("connection refused"),
			},
			readmeContent: map[string]string{
				"README.md": "https://bestpractices.dev/projects/42",
			},
			repoURI:   "github.com/ossf/scorecard",
			wantBadge: clients.NotFound,
		},
		{
			name: "fallback: badge URL inside fenced code block is ignored",
			ciiClient: &fakeCIIClient{
				directLevel: clients.NotFound,
				projectResp: clients.BadgeResponse{BadgeLevel: "gold", RepoURL: "https://github.com/ossf/allstar"},
			},
			readmeContent: map[string]string{
				"README.md": "```\nhttps://bestpractices.dev/projects/1234\n```",
			},
			repoURI:   "github.com/ossf/scorecard",
			wantBadge: clients.NotFound,
		},
		{
			name:          "no readme: fallback not attempted",
			ciiClient:     &fakeCIIClient{directLevel: clients.NotFound},
			readmeContent: map[string]string{},
			repoURI:       "github.com/ossf/scorecard",
			wantBadge:     clients.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := &checker.CheckRequest{
				Ctx:        context.Background(),
				CIIClient:  tt.ciiClient,
				RepoClient: &fakeRepoClient{files: tt.readmeContent},
				Repo:       fakeRepo{uri: tt.repoURI},
			}
			got, err := CIIBestPractices(req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CIIBestPractices() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if got.Badge != tt.wantBadge {
				t.Errorf("Badge = %v, want %v", got.Badge, tt.wantBadge)
			}
			if got.InheritedFrom != tt.wantInherited {
				t.Errorf("InheritedFrom = %q, want %q", got.InheritedFrom, tt.wantInherited)
			}
		})
	}
}

func TestExtractBadgeIDs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		files   map[string]string
		wantIDs []int
	}{
		{
			name:    "standard bestpractices.dev URL",
			files:   map[string]string{"README.md": "https://www.bestpractices.dev/projects/1234"},
			wantIDs: []int{1234},
		},
		{
			name:    "URL with /en/ segment",
			files:   map[string]string{"README.md": "https://bestpractices.dev/en/projects/5678"},
			wantIDs: []int{5678},
		},
		{
			name:    "legacy coreinfrastructure.org URL",
			files:   map[string]string{"README.md": "https://www.coreinfrastructure.org/projects/999"},
			wantIDs: []int{999},
		},
		{
			name:    "deduplicated IDs",
			files:   map[string]string{"README.md": "https://bestpractices.dev/projects/1\nhttps://bestpractices.dev/projects/1"},
			wantIDs: []int{1},
		},
		{
			name:    "cap at 2 IDs",
			files:   map[string]string{"README.md": "https://bestpractices.dev/projects/1\nhttps://bestpractices.dev/projects/2\nhttps://bestpractices.dev/projects/3"},
			wantIDs: []int{1, 2},
		},
		{
			name:    "look-alike host is rejected",
			files:   map[string]string{"README.md": "https://evil-bestpractices.dev/projects/1\nhttps://bestpractices.dev.attacker.com/projects/2"},
			wantIDs: nil,
		},
		{
			name:    "URL in fenced code block ignored",
			files:   map[string]string{"README.md": "```\nhttps://bestpractices.dev/projects/42\n```"},
			wantIDs: nil,
		},
		{
			name:    "file in subdirectory ignored",
			files:   map[string]string{"docs/README.md": "https://bestpractices.dev/projects/9"},
			wantIDs: nil,
		},
		{
			name:    "no readme",
			files:   map[string]string{},
			wantIDs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := &checker.CheckRequest{
				RepoClient: &fakeRepoClient{files: tt.files},
			}
			got := extractBadgeIDs(req)
			if len(got) != len(tt.wantIDs) {
				t.Fatalf("extractBadgeIDs() = %v, want %v", got, tt.wantIDs)
			}
			for i := range got {
				if got[i] != tt.wantIDs[i] {
					t.Errorf("id[%d] = %d, want %d", i, got[i], tt.wantIDs[i])
				}
			}
		})
	}
}
