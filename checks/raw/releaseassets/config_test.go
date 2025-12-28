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

package releaseassets

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
)

// mockConfig implements the config interface for testing.
type mockConfig struct {
	globalURLs   []string
	releaseRules []mockReleaseRule
}

type mockReleaseRule struct {
	tag  string
	urls []string
}

func (m mockConfig) GetGPGKeys() interface{} {
	return mockGPGKeys(m)
}

type mockGPGKeys struct {
	globalURLs   []string
	releaseRules []mockReleaseRule
}

func (m mockGPGKeys) GetURLs() []string {
	return m.globalURLs
}

func (m mockGPGKeys) GetReleases() interface{} {
	result := make([]interface{}, len(m.releaseRules))
	for i := range m.releaseRules {
		result[i] = m.releaseRules[i]
	}
	return result
}

func (m mockReleaseRule) GetTag() string {
	return m.tag
}

func (m mockReleaseRule) GetURLs() []string {
	return m.urls
}

//nolint:gocognit // Test function requires comprehensive coverage with many subtests
func TestDiscoverGPGKeys_ConfigBased(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		config         interface{}
		releaseTag     string
		wantSource     string
		wantURLs       []string
		wantNumResults int
	}{
		{
			name: "Per-release key matches v1.*",
			config: mockConfig{
				globalURLs: []string{"https://example.com/global.asc"},
				releaseRules: []mockReleaseRule{
					{tag: "v1.*", urls: []string{"https://example.com/v1-key.asc"}},
					{tag: "v2.*", urls: []string{"https://example.com/v2-key.asc"}},
				},
			},
			releaseTag:     "v1.2.3",
			wantSource:     "config-release",
			wantURLs:       []string{"https://example.com/v1-key.asc"},
			wantNumResults: 1,
		},
		{
			name: "Per-release key matches v2.0.*",
			config: mockConfig{
				globalURLs: []string{"https://example.com/global.asc"},
				releaseRules: []mockReleaseRule{
					{tag: "v1.*", urls: []string{"https://example.com/v1-key.asc"}},
					{tag: "v2.0.*", urls: []string{"https://example.com/v2.0-key.asc"}},
				},
			},
			releaseTag:     "v2.0.5",
			wantSource:     "config-release",
			wantURLs:       []string{"https://example.com/v2.0-key.asc"},
			wantNumResults: 1,
		},
		{
			name: "Exact release match",
			config: mockConfig{
				globalURLs: []string{"https://example.com/global.asc"},
				releaseRules: []mockReleaseRule{
					{tag: "v1.0.0", urls: []string{"https://example.com/v1.0.0-key.asc"}},
				},
			},
			releaseTag:     "v1.0.0",
			wantSource:     "config-release",
			wantURLs:       []string{"https://example.com/v1.0.0-key.asc"},
			wantNumResults: 1,
		},
		{
			name: "Beta release wildcard",
			config: mockConfig{
				globalURLs: []string{"https://example.com/global.asc"},
				releaseRules: []mockReleaseRule{
					{tag: "*-beta*", urls: []string{"https://example.com/beta-key.asc"}},
				},
			},
			releaseTag:     "v2.0.0-beta.1",
			wantSource:     "config-release",
			wantURLs:       []string{"https://example.com/beta-key.asc"},
			wantNumResults: 1,
		},
		{
			name: "No pattern match - uses global",
			config: mockConfig{
				globalURLs: []string{"https://example.com/global.asc"},
				releaseRules: []mockReleaseRule{
					{tag: "v1.*", urls: []string{"https://example.com/v1-key.asc"}},
				},
			},
			releaseTag:     "v3.0.0",
			wantSource:     "config-global",
			wantURLs:       []string{"https://example.com/global.asc"},
			wantNumResults: 1,
		},
		{
			name: "Global keys only (no per-release)",
			config: mockConfig{
				globalURLs: []string{
					"https://example.com/key1.asc",
					"https://example.com/key2.asc",
				},
				releaseRules: nil,
			},
			releaseTag:     "v1.0.0",
			wantSource:     "config-global",
			wantURLs:       []string{"https://example.com/key1.asc", "https://example.com/key2.asc"},
			wantNumResults: 1,
		},
		{
			name: "Multiple URLs for matching release",
			config: mockConfig{
				releaseRules: []mockReleaseRule{
					{
						tag: "v1.*",
						urls: []string{
							"https://example.com/key1.asc",
							"https://example.com/key2.asc",
							"https://backup.example.org/key.asc",
						},
					},
				},
			},
			releaseTag: "v1.5.0",
			wantSource: "config-release",
			wantURLs: []string{
				"https://example.com/key1.asc",
				"https://example.com/key2.asc",
				"https://backup.example.org/key.asc",
			},
			wantNumResults: 1,
		},
		{
			name: "First matching pattern wins",
			config: mockConfig{
				releaseRules: []mockReleaseRule{
					{tag: "v1.*", urls: []string{"https://example.com/v1-key.asc"}},
					{tag: "v1.0.*", urls: []string{"https://example.com/v1.0-key.asc"}},
				},
			},
			releaseTag:     "v1.0.5",
			wantSource:     "config-release",
			wantURLs:       []string{"https://example.com/v1-key.asc"},
			wantNumResults: 1,
		},
		{
			name:           "No config provided",
			config:         nil,
			releaseTag:     "v1.0.0",
			wantSource:     "",
			wantURLs:       nil,
			wantNumResults: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockRepo := mockrepo.NewMockRepoClient(ctrl)
			// Return error for GetFileReader to skip KEYS file check
			mockRepo.EXPECT().GetFileReader(gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()

			ctx := context.Background()
			req := &checker.CheckRequest{
				Config:     tt.config,
				RepoClient: mockRepo,
			}
			release := &clients.Release{
				TagName: tt.releaseTag,
			}

			results := DiscoverGPGKeys(ctx, req, release)

			// Filter results to only config-based ones for this test
			configResults := []KeyDiscoveryResult{}
			for _, r := range results {
				if r.Source == "config-release" || r.Source == "config-global" {
					configResults = append(configResults, r)
				}
			}

			if len(configResults) != tt.wantNumResults {
				t.Errorf("DiscoverGPGKeys() returned %d config-based results, want %d",
					len(configResults), tt.wantNumResults)
				return
			}

			if tt.wantNumResults == 0 {
				return
			}

			result := configResults[0]
			if result.Source != tt.wantSource {
				t.Errorf("Source = %q, want %q", result.Source, tt.wantSource)
			}

			if len(result.KeyURLs) != len(tt.wantURLs) {
				t.Errorf("Got %d URLs, want %d", len(result.KeyURLs), len(tt.wantURLs))
			}

			for i, url := range tt.wantURLs {
				if i >= len(result.KeyURLs) {
					t.Errorf("Missing URL at index %d: %s", i, url)
					continue
				}
				if result.KeyURLs[i] != url {
					t.Errorf("URL[%d] = %q, want %q", i, result.KeyURLs[i], url)
				}
			}
		})
	}
}

func TestDiscoverGPGKeys_Integration(t *testing.T) {
	t.Parallel()

	t.Run("Config keys take precedence over other methods", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		mockRepo := mockrepo.NewMockRepoClient(ctrl)
		mockRepo.EXPECT().GetFileReader(gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()

		ctx := context.Background()
		req := &checker.CheckRequest{
			Config: mockConfig{
				globalURLs: []string{"https://example.com/config-key.asc"},
			},
			RepoClient: mockRepo,
		}
		release := &clients.Release{
			TagName: "v1.0.0",
			Body:    "GPG Key Fingerprint: 1234567890ABCDEF1234567890ABCDEF12345678", // Has fingerprint
		}

		results := DiscoverGPGKeys(ctx, req, release)

		// Should have config result first
		if len(results) == 0 {
			t.Fatal("Expected at least one result")
		}

		firstResult := results[0]
		if firstResult.Source != "config-global" {
			t.Errorf("First result source = %q, want config-global", firstResult.Source)
		}

		if len(firstResult.KeyURLs) != 1 || firstResult.KeyURLs[0] != "https://example.com/config-key.asc" {
			t.Errorf("Config key URL not found in first result")
		}
	})

	t.Run("Falls back to release notes when no config match", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		mockRepo := mockrepo.NewMockRepoClient(ctrl)
		mockRepo.EXPECT().GetFileReader(gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()

		ctx := context.Background()
		req := &checker.CheckRequest{
			Config: mockConfig{
				releaseRules: []mockReleaseRule{
					{tag: "v1.*", urls: []string{"https://example.com/v1-key.asc"}},
				},
			},
			RepoClient: mockRepo,
		}
		release := &clients.Release{
			TagName: "v2.0.0", // Doesn't match v1.*
			Body:    "GPG Key Fingerprint: ABCDEF1234567890ABCDEF1234567890ABCDEF12",
		}

		results := DiscoverGPGKeys(ctx, req, release)

		// Should have release_notes result since v2.0.0 doesn't match v1.* and no global keys
		foundReleaseNotes := false
		for _, r := range results {
			if r.Source == "release_notes" {
				foundReleaseNotes = true
				if len(r.KeyFingerprints) == 0 {
					t.Error("Expected fingerprints from release notes")
				}
			}
		}

		if !foundReleaseNotes {
			t.Error("Expected fallback to release_notes discovery")
		}
	})
}

func TestDiscoverGPGKeys_EmptyConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		config interface{}
		name   string
	}{
		{
			name:   "Nil config",
			config: nil,
		},
		{
			name: "Empty global URLs",
			config: mockConfig{
				globalURLs:   []string{},
				releaseRules: nil,
			},
		},
		{
			name: "Empty release rules",
			config: mockConfig{
				globalURLs:   nil,
				releaseRules: []mockReleaseRule{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockRepo := mockrepo.NewMockRepoClient(ctrl)
			mockRepo.EXPECT().GetFileReader(gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()

			ctx := context.Background()
			req := &checker.CheckRequest{
				Config:     tt.config,
				RepoClient: mockRepo,
			}
			release := &clients.Release{
				TagName: "v1.0.0",
			}

			results := DiscoverGPGKeys(ctx, req, release)

			// Should not return any config-based results
			for _, r := range results {
				if r.Source == "config-release" || r.Source == "config-global" {
					t.Errorf("Unexpected config-based result with source %q", r.Source)
				}
			}
		})
	}
}
