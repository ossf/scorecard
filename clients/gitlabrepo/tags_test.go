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

package gitlabrepo

import (
	"fmt"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/ossf/scorecard/v5/clients"
)

func verifyTagProtection(
	t *testing.T,
	tagRef *clients.TagRef,
	expectedProtected,
	expectedRestrict bool,
) {
	t.Helper()

	if tagRef == nil {
		t.Fatal("expected non-nil tagRef")
	}

	if tagRef.Protected == nil {
		t.Fatal("expected non-nil Protected field")
	}

	if *tagRef.Protected != expectedProtected {
		t.Errorf(
			"Protected mismatch: got %v, want %v",
			*tagRef.Protected,
			expectedProtected,
		)
	}

	if tagRef.TagProtectionRule.RestrictCreation == nil {
		t.Fatal("expected non-nil RestrictCreation field")
	}

	if *tagRef.TagProtectionRule.RestrictCreation != expectedRestrict {
		t.Errorf(
			"RestrictCreation mismatch: got %v, want %v",
			*tagRef.TagProtectionRule.RestrictCreation,
			expectedRestrict,
		)
	}

	// Verify protected tags have deletion and update blocking
	if expectedProtected {
		if tagRef.TagProtectionRule.AllowDeletions == nil {
			t.Error("expected non-nil AllowDeletions for protected tag")
		} else if *tagRef.TagProtectionRule.AllowDeletions {
			t.Error("expected AllowDeletions=false for protected tag")
		}

		if tagRef.TagProtectionRule.AllowUpdates == nil {
			t.Error("expected non-nil AllowUpdates for protected tag")
		} else if *tagRef.TagProtectionRule.AllowUpdates {
			t.Error("expected AllowUpdates=false for protected tag")
		}
	}
}

func TestGetTag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		tagName           string
		listError         error
		protectedPatterns []*gitlab.ProtectedTag
		expectedProtected bool
		expectedRestrict  bool
	}{
		{
			name:    "Exact match protection",
			tagName: "v1.0.0",
			protectedPatterns: []*gitlab.ProtectedTag{
				{
					Name:               "v1.0.0",
					CreateAccessLevels: []*gitlab.TagAccessDescription{{AccessLevel: 40}},
				},
			},
			expectedProtected: true,
			expectedRestrict:  true,
		},
		{
			name:    "Wildcard pattern match - v*",
			tagName: "v1.2.3",
			protectedPatterns: []*gitlab.ProtectedTag{
				{
					Name:               "v*",
					CreateAccessLevels: []*gitlab.TagAccessDescription{{AccessLevel: 40}},
				},
			},
			expectedProtected: true,
			expectedRestrict:  true,
		},
		{
			name:    "Wildcard pattern match - release-*",
			tagName: "release-2024",
			protectedPatterns: []*gitlab.ProtectedTag{
				{
					Name:               "release-*",
					CreateAccessLevels: []*gitlab.TagAccessDescription{{AccessLevel: 30}},
				},
			},
			expectedProtected: true,
			expectedRestrict:  true,
		},
		{
			name:    "No match - unprotected tag",
			tagName: "test-tag",
			protectedPatterns: []*gitlab.ProtectedTag{
				{
					Name:               "v*",
					CreateAccessLevels: []*gitlab.TagAccessDescription{{AccessLevel: 40}},
				},
			},
			expectedProtected: false,
			expectedRestrict:  false,
		},
		{
			name:              "No protected patterns",
			tagName:           "v1.0.0",
			protectedPatterns: []*gitlab.ProtectedTag{},
			expectedProtected: false,
			expectedRestrict:  false,
		},
		{
			name:    "Multiple patterns - all patterns merged",
			tagName: "v2.0.0",
			protectedPatterns: []*gitlab.ProtectedTag{
				{
					Name:               "v*",
					CreateAccessLevels: []*gitlab.TagAccessDescription{{AccessLevel: 40}},
				},
				{
					Name:               "v2.*",
					CreateAccessLevels: []*gitlab.TagAccessDescription{{AccessLevel: 30}},
				},
			},
			expectedProtected: true,
			expectedRestrict:  true,
		},
		{
			name:    "Multiple patterns - one with restrictions, one without (merged)",
			tagName: "v1.5.0",
			protectedPatterns: []*gitlab.ProtectedTag{
				{
					Name:               "v*",
					CreateAccessLevels: []*gitlab.TagAccessDescription{}, // No restrictions
				},
				{
					Name:               "v1.*",
					CreateAccessLevels: []*gitlab.TagAccessDescription{{AccessLevel: 40}}, // Has restrictions
				},
			},
			expectedProtected: true,
			expectedRestrict:  true, // Should be true because merged patterns have access levels
		},
		{
			name:    "Pattern without create access levels",
			tagName: "v1.0.0",
			protectedPatterns: []*gitlab.ProtectedTag{
				{
					Name:               "v*",
					CreateAccessLevels: []*gitlab.TagAccessDescription{},
				},
			},
			expectedProtected: true,
			expectedRestrict:  false,
		},
		{
			name:              "List error - treats as unprotected",
			tagName:           "v1.0.0",
			protectedPatterns: nil,
			listError:         fmt.Errorf("API error"),
			expectedProtected: false,
			expectedRestrict:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &tagsHandler{
				once: new(sync.Once),
				repourl: &Repo{
					commitSHA: clients.HeadSHA,
					projectID: "test-project",
				},
				listProtectedTags: func(
					pid interface{},
					opt *gitlab.ListProtectedTagsOptions,
					options ...gitlab.RequestOptionFunc,
				) ([]*gitlab.ProtectedTag, *gitlab.Response, error) {
					if tt.listError != nil {
						return nil, nil, tt.listError
					}
					return tt.protectedPatterns, nil, nil
				},
			}

			// Call setup to populate patterns
			if err := handler.setup(); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			// Get the tag
			tagRef, err := handler.getTag(tt.tagName)
			if err != nil {
				t.Fatalf("getTag failed: %v", err)
			}

			verifyTagProtection(t, tagRef, tt.expectedProtected, tt.expectedRestrict)
		})
	}
}

func TestFindMatchingPatterns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		tagName           string
		protectedPatterns []*gitlab.ProtectedTag
		expectedPatterns  []string
	}{
		{
			name:    "Exact match only",
			tagName: "v1.0.0",
			protectedPatterns: []*gitlab.ProtectedTag{
				{Name: "v1.0.0"},
				{Name: "v2.*"},
			},
			expectedPatterns: []string{"v1.0.0"},
		},
		{
			name:    "Wildcard match only",
			tagName: "v1.2.3",
			protectedPatterns: []*gitlab.ProtectedTag{
				{Name: "v*"},
			},
			expectedPatterns: []string{"v*"},
		},
		{
			name:    "Multiple pattern matches",
			tagName: "v1.2.3",
			protectedPatterns: []*gitlab.ProtectedTag{
				{Name: "v*"},
				{Name: "v1.*"},
				{Name: "*.*.*"},
				{Name: "release-*"}, // doesn't match
			},
			expectedPatterns: []string{"v*", "v1.*", "*.*.*"},
		},
		{
			name:    "Question mark wildcard",
			tagName: "v1",
			protectedPatterns: []*gitlab.ProtectedTag{
				{Name: "v?"},
			},
			expectedPatterns: []string{"v?"},
		},
		{
			name:    "Complex pattern",
			tagName: "release-2024-01",
			protectedPatterns: []*gitlab.ProtectedTag{
				{Name: "release-*-*"},
			},
			expectedPatterns: []string{"release-*-*"},
		},
		{
			name:    "No match",
			tagName: "test-tag",
			protectedPatterns: []*gitlab.ProtectedTag{
				{Name: "v*"},
				{Name: "release-*"},
			},
			expectedPatterns: []string{},
		},
		{
			name:              "Empty patterns",
			tagName:           "v1.0.0",
			protectedPatterns: []*gitlab.ProtectedTag{},
			expectedPatterns:  []string{},
		},
		{
			name:    "Pattern with empty name skipped",
			tagName: "v1.0.0",
			protectedPatterns: []*gitlab.ProtectedTag{
				{Name: ""},
				{Name: "v*"},
			},
			expectedPatterns: []string{"v*"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &tagsHandler{
				protectedPatterns: tt.protectedPatterns,
			}

			results := handler.findMatchingPatterns(tt.tagName)

			if len(results) != len(tt.expectedPatterns) {
				t.Errorf(
					"expected %d matches but got %d",
					len(tt.expectedPatterns),
					len(results),
				)
				return
			}

			// Check that all expected patterns are present
			for _, expected := range tt.expectedPatterns {
				found := false
				for _, result := range results {
					if result.Name == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected pattern %q not found in results", expected)
				}
			}
		})
	}
}

func TestMakeTagRefFrom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		protectedTag   *gitlab.ProtectedTag
		name           string
		tagName        string
		wantProtected  bool
		wantRestricted bool
	}{
		{
			name:    "Tag with create access levels",
			tagName: "v1.0.0",
			protectedTag: &gitlab.ProtectedTag{
				Name:               "v*",
				CreateAccessLevels: []*gitlab.TagAccessDescription{{AccessLevel: 40}},
			},
			wantProtected:  true,
			wantRestricted: true,
		},
		{
			name:    "Tag without create access levels",
			tagName: "v2.0.0",
			protectedTag: &gitlab.ProtectedTag{
				Name:               "v*",
				CreateAccessLevels: []*gitlab.TagAccessDescription{},
			},
			wantProtected:  true,
			wantRestricted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := makeTagRefFrom(tt.tagName, []*gitlab.ProtectedTag{tt.protectedTag})

			if result == nil {
				t.Fatal("expected non-nil result")
			}

			if result.Name == nil || *result.Name != tt.tagName {
				t.Errorf("tag name mismatch: got %v, want %s", result.Name, tt.tagName)
			}

			if result.Protected == nil || *result.Protected != tt.wantProtected {
				t.Errorf(
					"Protected mismatch: got %v, want %v",
					result.Protected,
					tt.wantProtected,
				)
			}

			if result.TagProtectionRule.RestrictCreation == nil ||
				*result.TagProtectionRule.RestrictCreation != tt.wantRestricted {
				t.Errorf(
					"RestrictCreation mismatch: got %v, want %v",
					result.TagProtectionRule.RestrictCreation,
					tt.wantRestricted,
				)
			}

			// Verify GitLab always blocks deletions and updates for protected tags
			wantFalse := false
			if diff := cmp.Diff(
				&wantFalse,
				result.TagProtectionRule.AllowDeletions,
				cmpopts.EquateEmpty(),
			); diff != "" {
				t.Errorf("AllowDeletions mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(
				&wantFalse,
				result.TagProtectionRule.AllowUpdates,
				cmpopts.EquateEmpty(),
			); diff != "" {
				t.Errorf("AllowUpdates mismatch (-want +got):\n%s", diff)
			}

			// Verify GitLab always enforces for admins
			wantTrue := true
			if diff := cmp.Diff(
				&wantTrue,
				result.TagProtectionRule.EnforceAdmins,
				cmpopts.EquateEmpty(),
			); diff != "" {
				t.Errorf("EnforceAdmins mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMergeCreateAccessLevels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		patterns       []*gitlab.ProtectedTag
		expectedLevels []gitlab.AccessLevelValue
	}{
		{
			name:           "Empty patterns",
			patterns:       []*gitlab.ProtectedTag{},
			expectedLevels: []gitlab.AccessLevelValue{},
		},
		{
			name: "Single pattern with one level",
			patterns: []*gitlab.ProtectedTag{
				{
					Name: "v*",
					CreateAccessLevels: []*gitlab.TagAccessDescription{
						{AccessLevel: 40},
					},
				},
			},
			expectedLevels: []gitlab.AccessLevelValue{40},
		},
		{
			name: "Multiple patterns with different levels - merged",
			patterns: []*gitlab.ProtectedTag{
				{
					Name: "v*",
					CreateAccessLevels: []*gitlab.TagAccessDescription{
						{AccessLevel: 40}, // Maintainer
					},
				},
				{
					Name: "v1.*",
					CreateAccessLevels: []*gitlab.TagAccessDescription{
						{AccessLevel: 30}, // Developer
					},
				},
			},
			expectedLevels: []gitlab.AccessLevelValue{30, 40},
		},
		{
			name: "Multiple patterns with duplicate levels - deduplicated",
			patterns: []*gitlab.ProtectedTag{
				{
					Name: "v*",
					CreateAccessLevels: []*gitlab.TagAccessDescription{
						{AccessLevel: 40},
					},
				},
				{
					Name: "v1.*",
					CreateAccessLevels: []*gitlab.TagAccessDescription{
						{AccessLevel: 40}, // Same level
					},
				},
			},
			expectedLevels: []gitlab.AccessLevelValue{40},
		},
		{
			name: "Pattern with empty access levels",
			patterns: []*gitlab.ProtectedTag{
				{
					Name:               "v*",
					CreateAccessLevels: []*gitlab.TagAccessDescription{},
				},
			},
			expectedLevels: []gitlab.AccessLevelValue{},
		},
		{
			name: "Mixed: some patterns with levels, others without",
			patterns: []*gitlab.ProtectedTag{
				{
					Name: "v*",
					CreateAccessLevels: []*gitlab.TagAccessDescription{
						{AccessLevel: 40},
					},
				},
				{
					Name:               "release-*",
					CreateAccessLevels: []*gitlab.TagAccessDescription{},
				},
				{
					Name: "v1.*",
					CreateAccessLevels: []*gitlab.TagAccessDescription{
						{AccessLevel: 30},
					},
				},
			},
			expectedLevels: []gitlab.AccessLevelValue{30, 40},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := mergeCreateAccessLevels(tt.patterns)

			if len(result) != len(tt.expectedLevels) {
				t.Errorf(
					"expected %d access levels but got %d",
					len(tt.expectedLevels),
					len(result),
				)
				return
			}

			// Check that all expected levels are present (order doesn't matter)
			for _, expectedLevel := range tt.expectedLevels {
				found := false
				for _, access := range result {
					if access.AccessLevel == expectedLevel {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected access level %d not found in results", expectedLevel)
				}
			}
		})
	}
}

func TestGetMinimumAccessLevelFromPatterns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		patterns []*gitlab.TagAccessDescription
		want     gitlab.AccessLevelValue
		wantErr  bool
	}{
		{
			name:     "empty patterns returns developer default",
			patterns: []*gitlab.TagAccessDescription{},
			want:     gitlab.DeveloperPermissions,
			wantErr:  false,
		},
		{
			name:     "nil patterns returns developer default",
			patterns: nil,
			want:     gitlab.DeveloperPermissions,
			wantErr:  false,
		},
		{
			name: "single No one access level",
			patterns: []*gitlab.TagAccessDescription{
				{AccessLevel: gitlab.NoPermissions},
			},
			want:    gitlab.NoPermissions,
			wantErr: false,
		},
		{
			name: "single Developer access level",
			patterns: []*gitlab.TagAccessDescription{
				{AccessLevel: gitlab.DeveloperPermissions},
			},
			want:    gitlab.DeveloperPermissions,
			wantErr: false,
		},
		{
			name: "single Maintainer access level",
			patterns: []*gitlab.TagAccessDescription{
				{AccessLevel: gitlab.MaintainerPermissions},
			},
			want:    gitlab.MaintainerPermissions,
			wantErr: false,
		},
		{
			name: "single Owner access level",
			patterns: []*gitlab.TagAccessDescription{
				{AccessLevel: gitlab.OwnerPermissions},
			},
			want:    gitlab.OwnerPermissions,
			wantErr: false,
		},
		{
			name: "multiple levels - returns minimum (most permissive)",
			patterns: []*gitlab.TagAccessDescription{
				{AccessLevel: gitlab.MaintainerPermissions},
				{AccessLevel: gitlab.DeveloperPermissions},
				{AccessLevel: gitlab.OwnerPermissions},
			},
			want:    gitlab.DeveloperPermissions,
			wantErr: false,
		},
		{
			name: "with No one - returns zero",
			patterns: []*gitlab.TagAccessDescription{
				{AccessLevel: gitlab.MaintainerPermissions},
				{AccessLevel: gitlab.NoPermissions},
				{AccessLevel: gitlab.DeveloperPermissions},
			},
			want:    gitlab.NoPermissions,
			wantErr: false,
		},
		{
			name: "nil entries are skipped",
			patterns: []*gitlab.TagAccessDescription{
				nil,
				{AccessLevel: gitlab.MaintainerPermissions},
				nil,
			},
			want:    gitlab.MaintainerPermissions,
			wantErr: false,
		},
		{
			name: "all nil entries return developer default",
			patterns: []*gitlab.TagAccessDescription{
				nil,
				nil,
				nil,
			},
			want:    gitlab.DeveloperPermissions,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &accessLevelHandler{
				repourl: &Repo{projectID: "test-project"},
			}

			got, err := handler.getMinimumAccessLevel(tt.patterns)

			if (err != nil) != tt.wantErr {
				t.Errorf("getMinimumAccessLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getMinimumAccessLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindMatchingPatternsInTags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		tagName  string
		patterns []*gitlab.ProtectedTag
		want     int
	}{
		{
			name:     "no patterns",
			tagName:  "v1.0.0",
			patterns: []*gitlab.ProtectedTag{},
			want:     0,
		},
		{
			name:    "exact match",
			tagName: "v1.0.0",
			patterns: []*gitlab.ProtectedTag{
				{Name: "v1.0.0"},
				{Name: "v2.0.0"},
			},
			want: 1,
		},
		{
			name:    "wildcard prefix",
			tagName: "v1.0.0",
			patterns: []*gitlab.ProtectedTag{
				{Name: "v*"},
			},
			want: 1,
		},
		{
			name:    "wildcard suffix",
			tagName: "release-v1.0.0",
			patterns: []*gitlab.ProtectedTag{
				{Name: "*v1.0.0"},
			},
			want: 1,
		},
		{
			name:    "catch-all wildcard",
			tagName: "anything",
			patterns: []*gitlab.ProtectedTag{
				{Name: "*"},
			},
			want: 1,
		},
		{
			name:    "multiple matches",
			tagName: "v1.0.0",
			patterns: []*gitlab.ProtectedTag{
				{Name: "v*"},
				{Name: "v1.0.0"},
				{Name: "*1.0.0"},
			},
			want: 3,
		},
		{
			name:    "no match",
			tagName: "hotfix-1.0",
			patterns: []*gitlab.ProtectedTag{
				{Name: "v*"},
				{Name: "release-*"},
			},
			want: 0,
		},
		{
			name:    "empty pattern name skipped",
			tagName: "v1.0.0",
			patterns: []*gitlab.ProtectedTag{
				{Name: ""},
				{Name: "v*"},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := &tagsHandler{
				protectedPatterns: tt.patterns,
			}

			got := handler.findMatchingPatterns(tt.tagName)

			if len(got) != tt.want {
				t.Errorf("findMatchingPatterns() returned %d matches, want %d", len(got), tt.want)
			}
		})
	}
}
