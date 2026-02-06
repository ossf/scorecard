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
	"testing"

	"github.com/ossf/scorecard/v5/clients"
)

func TestMatchAssetPairs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		expectedType  string
		assets        []clients.ReleaseAsset
		expectedPairs int
	}{
		{
			name: "GPG signature match",
			assets: []clients.ReleaseAsset{
				{Name: "myapp-1.0.tar.gz", URL: "https://example.com/myapp-1.0.tar.gz"},
				{Name: "myapp-1.0.tar.gz.asc", URL: "https://example.com/myapp-1.0.tar.gz.asc"},
			},
			expectedPairs: 1,
			expectedType:  "gpg",
		},
		{
			name: "Multiple artifacts with signatures",
			assets: []clients.ReleaseAsset{
				{Name: "app.jar", URL: "https://example.com/app.jar"},
				{Name: "app.jar.asc", URL: "https://example.com/app.jar.asc"},
				{Name: "app.tar.gz", URL: "https://example.com/app.tar.gz"},
				{Name: "app.tar.gz.sig", URL: "https://example.com/app.tar.gz.sig"},
			},
			expectedPairs: 2,
			expectedType:  "gpg",
		},
		{
			name: "Sigstore bundle",
			assets: []clients.ReleaseAsset{
				{Name: "release.whl", URL: "https://example.com/release.whl"},
				{Name: "release.whl.sigstore.json", URL: "https://example.com/release.whl.sigstore.json"},
			},
			expectedPairs: 1,
			expectedType:  "sigstore",
		},
		{
			name: "No matching pairs",
			assets: []clients.ReleaseAsset{
				{Name: "app.jar", URL: "https://example.com/app.jar"},
				{Name: "README.md", URL: "https://example.com/README.md"},
			},
			expectedPairs: 0,
		},
		{
			name: "Signature without artifact",
			assets: []clients.ReleaseAsset{
				{Name: "orphan.asc", URL: "https://example.com/orphan.asc"},
			},
			expectedPairs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pairs := MatchAssetPairs(tt.assets)

			if len(pairs) != tt.expectedPairs {
				t.Errorf("Expected %d pairs, got %d", tt.expectedPairs, len(pairs))
			}

			if tt.expectedPairs > 0 && len(pairs) > 0 {
				if pairs[0].SignatureType != tt.expectedType {
					t.Errorf("Expected signature type %s, got %s", tt.expectedType, pairs[0].SignatureType)
				}
			}
		})
	}
}

func TestIsValidFingerprint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fingerprint string
		expected    bool
	}{
		{
			name:        "Valid uppercase fingerprint",
			fingerprint: "ABCD1234ABCD1234ABCD1234ABCD1234ABCD1234",
			expected:    true,
		},
		{
			name:        "Valid lowercase fingerprint",
			fingerprint: "abcd1234abcd1234abcd1234abcd1234abcd1234",
			expected:    true,
		},
		{
			name:        "Valid mixed case fingerprint",
			fingerprint: "AbCd1234aBcD1234AbCd1234aBcD1234AbCd1234",
			expected:    true,
		},
		{
			name:        "Too short",
			fingerprint: "ABCD1234",
			expected:    false,
		},
		{
			name:        "Too long",
			fingerprint: "ABCD1234ABCD1234ABCD1234ABCD1234ABCD1234EXTRA",
			expected:    false,
		},
		{
			name:        "Invalid characters",
			fingerprint: "GHIJ1234ABCD1234ABCD1234ABCD1234ABCD1234",
			expected:    false,
		},
		{
			name:        "Contains spaces",
			fingerprint: "ABCD 1234 ABCD 1234 ABCD 1234 ABCD 1234 ABCD",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := isValidFingerprint(tt.fingerprint)
			if result != tt.expected {
				t.Errorf("Expected %v for %s, got %v", tt.expected, tt.fingerprint, result)
			}
		})
	}
}

func TestExtractKeyFingerprints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		expectedContains string
		text             string
		expectedCount    int
	}{
		{
			name: "Fingerprint in release notes",
			text: `Release v1.0.0
			
Signed with GPG key: ABCD1234ABCD1234ABCD1234ABCD1234ABCD1234

Download and verify.`,
			expectedCount:    1,
			expectedContains: "ABCD1234ABCD1234ABCD1234ABCD1234ABCD1234",
		},
		{
			name: "Multiple fingerprints",
			text: `Keys used:
Key fingerprint = 1111111111111111111111111111111111111111
Fingerprint: 2222222222222222222222222222222222222222`,
			expectedCount:    2,
			expectedContains: "1111111111111111111111111111111111111111",
		},
		{
			name:          "No fingerprints",
			text:          "This is a release with no key information",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fingerprints := extractKeyFingerprints(tt.text)

			if len(fingerprints) != tt.expectedCount {
				t.Errorf("Expected %d fingerprints, got %d", tt.expectedCount, len(fingerprints))
			}

			if tt.expectedCount > 0 && tt.expectedContains != "" {
				found := false
				for _, fp := range fingerprints {
					if fp == tt.expectedContains {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find fingerprint %s", tt.expectedContains)
				}
			}
		})
	}
}

func TestParseGPGKeysFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		content       string
		expectedCount int
	}{
		{
			name: "Apache KEYS file format",
			content: `pub   4096R/12345678 2020-01-01
      Key fingerprint = AAAA1111BBBB2222CCCC3333DDDD4444EEEE5555
uid                  Developer <dev@example.com>

pub   2048R/87654321 2021-01-01
      Key fingerprint = FFFF6666EEEE5555DDDD4444CCCC3333BBBB2222
uid                  Another Dev <other@example.com>`,
			expectedCount: 2,
		},
		{
			name: "Simple fingerprint list",
			content: `1234567890ABCDEF1234567890ABCDEF12345678
ABCDEF1234567890ABCDEF1234567890ABCDEF12`,
			expectedCount: 2,
		},
		{
			name:          "No valid fingerprints",
			content:       "This file contains no valid GPG fingerprints",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fingerprints := parseGPGKeysFile(tt.content)

			if len(fingerprints) != tt.expectedCount {
				t.Errorf("Expected %d fingerprints, got %d", tt.expectedCount, len(fingerprints))
			}
		})
	}
}
