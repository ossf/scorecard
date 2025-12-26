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

package raw

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
)

func ptrString(s string) *string {
	return &s
}

var errTagTest = errors.New("test error")

func setupMockRepo(ctrl *gomock.Controller, releases []clients.Release, releasesErr error,
	tags map[string]*clients.TagRef, getTagErr error,
) *mockrepo.MockRepoClient {
	mockRepo := mockrepo.NewMockRepoClient(ctrl)

	mockRepo.EXPECT().ListReleases().Return(releases, releasesErr).AnyTimes()

	for tagName, tagRef := range tags {
		mockRepo.EXPECT().GetTag(tagName).Return(tagRef, getTagErr).AnyTimes()
	}

	// For tags not in the map, return nil
	mockRepo.EXPECT().GetTag(gomock.Any()).Return(nil, getTagErr).AnyTimes()

	return mockRepo
}

func verifyTagResults(
	t *testing.T,
	result checker.TagProtectionsData,
	expectedLen int,
	expectedTags map[string]*clients.TagRef,
) {
	t.Helper()

	if len(result.Tags) != expectedLen {
		t.Errorf("Expected %d tags, got %d", expectedLen, len(result.Tags))
	}

	// Verify tag data matches expectations
	for _, tag := range result.Tags {
		if tag.Name == nil {
			t.Errorf("Tag has nil name")
			continue
		}
		expectedTag, ok := expectedTags[*tag.Name]
		if !ok {
			t.Errorf("Unexpected tag %s in results", *tag.Name)
			continue
		}
		if diff := cmp.Diff(expectedTag, &tag); diff != "" {
			t.Errorf("Tag mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestTagProtection(t *testing.T) {
	t.Parallel()
	trueVal := true
	falseVal := false
	tagName1 := "v1.0.0"
	tagName2 := "v2.0.0"
	tagName3 := "v3.0.0"

	tests := []struct {
		releasesErr     error
		getTagErr       error
		tags            map[string]*clients.TagRef
		name            string
		releases        []clients.Release
		expectedTagsLen int
		wantErr         bool
	}{
		{
			name:            "No releases",
			releases:        []clients.Release{},
			tags:            map[string]*clients.TagRef{},
			expectedTagsLen: 0,
			wantErr:         false,
		},
		{
			name: "Single release with protected tag",
			releases: []clients.Release{
				{TagName: tagName1},
			},
			tags: map[string]*clients.TagRef{
				tagName1: {
					Name:      &tagName1,
					Protected: &trueVal,
					TagProtectionRule: clients.TagProtectionRule{
						RefProtectionRule: clients.RefProtectionRule{
							AllowDeletions:   &falseVal,
							AllowForcePushes: &falseVal,
							EnforceAdmins:    &trueVal,
						},
						AllowUpdates:      &falseVal,
						RestrictCreation:  &trueVal,
						RequireSignatures: &trueVal,
					},
				},
			},
			expectedTagsLen: 1,
			wantErr:         false,
		},
		{
			name: "Multiple releases with same tag (deduplication)",
			releases: []clients.Release{
				{TagName: tagName1},
				{TagName: tagName1},
				{TagName: tagName1},
			},
			tags: map[string]*clients.TagRef{
				tagName1: {
					Name:      &tagName1,
					Protected: &trueVal,
				},
			},
			expectedTagsLen: 1,
			wantErr:         false,
		},
		{
			name: "Multiple different tags",
			releases: []clients.Release{
				{TagName: tagName1},
				{TagName: tagName2},
				{TagName: tagName3},
			},
			tags: map[string]*clients.TagRef{
				tagName1: {
					Name:      &tagName1,
					Protected: &trueVal,
				},
				tagName2: {
					Name:      &tagName2,
					Protected: &falseVal,
				},
				tagName3: {
					Name:      &tagName3,
					Protected: &trueVal,
				},
			},
			expectedTagsLen: 3,
			wantErr:         false,
		},
		{
			name: "Release with empty tag name (skipped)",
			releases: []clients.Release{
				{TagName: ""},
				{TagName: tagName1},
			},
			tags: map[string]*clients.TagRef{
				tagName1: {
					Name:      &tagName1,
					Protected: &trueVal,
				},
			},
			expectedTagsLen: 1,
			wantErr:         false,
		},
		{
			name: "GetTag returns error (skips tag)",
			releases: []clients.Release{
				{TagName: tagName1},
			},
			tags:            map[string]*clients.TagRef{},
			getTagErr:       errTagTest,
			expectedTagsLen: 0,
			wantErr:         false, // Changed: now we skip tags with errors instead of failing
		},
		{
			name: "GetTag returns ErrUnsupportedFeature",
			releases: []clients.Release{
				{TagName: tagName1},
			},
			tags:            map[string]*clients.TagRef{},
			getTagErr:       clients.ErrUnsupportedFeature,
			expectedTagsLen: 0,
			wantErr:         true, // Still returns error for unsupported features
		},
		{
			name:            "ListReleases returns error",
			releases:        nil,
			releasesErr:     errTagTest,
			tags:            map[string]*clients.TagRef{},
			expectedTagsLen: 0,
			wantErr:         true,
		},
		{
			name: "Mixed protected and unprotected tags",
			releases: []clients.Release{
				{TagName: tagName1},
				{TagName: tagName2},
			},
			tags: map[string]*clients.TagRef{
				tagName1: {
					Name:      &tagName1,
					Protected: &trueVal,
					TagProtectionRule: clients.TagProtectionRule{
						RefProtectionRule: clients.RefProtectionRule{
							AllowDeletions:   &falseVal,
							AllowForcePushes: &falseVal,
						},
					},
				},
				tagName2: {
					Name:      &tagName2,
					Protected: &falseVal,
				},
			},
			expectedTagsLen: 2,
			wantErr:         false,
		},
		{
			name: "Tag with nil Protected field",
			releases: []clients.Release{
				{TagName: tagName1},
			},
			tags: map[string]*clients.TagRef{
				tagName1: {
					Name:      &tagName1,
					Protected: nil, // nil protected field
				},
			},
			expectedTagsLen: 1,
			wantErr:         false,
		},
		{
			name: "Multiple releases with duplicates and unique tags",
			releases: []clients.Release{
				{TagName: tagName1},
				{TagName: tagName1},
				{TagName: tagName2},
				{TagName: tagName3},
				{TagName: tagName2},
			},
			tags: map[string]*clients.TagRef{
				tagName1: {
					Name:      &tagName1,
					Protected: &trueVal,
				},
				tagName2: {
					Name:      &tagName2,
					Protected: &trueVal,
				},
				tagName3: {
					Name:      &tagName3,
					Protected: &falseVal,
				},
			},
			expectedTagsLen: 3,
			wantErr:         false,
		},
		{
			name: "Tag with partial protection data",
			releases: []clients.Release{
				{TagName: tagName1},
			},
			tags: map[string]*clients.TagRef{
				tagName1: {
					Name:      &tagName1,
					Protected: &trueVal,
					TagProtectionRule: clients.TagProtectionRule{
						RefProtectionRule: clients.RefProtectionRule{
							AllowDeletions: &falseVal,
							// Other fields nil
						},
						AllowUpdates: &trueVal,
						// Other tag-specific fields nil
					},
				},
			},
			expectedTagsLen: 1,
			wantErr:         false,
		},
		{
			name: "Multiple empty tag names interspersed",
			releases: []clients.Release{
				{TagName: tagName1},
				{TagName: ""},
				{TagName: tagName2},
				{TagName: ""},
				{TagName: tagName3},
			},
			tags: map[string]*clients.TagRef{
				tagName1: {
					Name:      &tagName1,
					Protected: &trueVal,
				},
				tagName2: {
					Name:      &tagName2,
					Protected: &trueVal,
				},
				tagName3: {
					Name:      &tagName3,
					Protected: &falseVal,
				},
			},
			expectedTagsLen: 3,
			wantErr:         false,
		},
		{
			// This test simulates the behavior when a client (GitHub/GitLab) returns
			// pattern-matched protection results. For example, a GitLab pattern "v*"
			// or GitHub ruleset "refs/tags/v*" would protect all tags starting with "v".
			// The check should correctly handle tags that are protected via pattern
			// matching vs exact rules.
			name: "Pattern-matched tags (simulating v* pattern protection)",
			releases: []clients.Release{
				{TagName: "v1.0.0"},
				{TagName: "v2.0.0"},
				{TagName: "v3.1.5"},
				{TagName: "release-2024"}, // Different pattern, not protected
			},
			tags: map[string]*clients.TagRef{
				"v1.0.0": {
					Name:      ptrString("v1.0.0"),
					Protected: &trueVal,
					TagProtectionRule: clients.TagProtectionRule{
						RefProtectionRule: clients.RefProtectionRule{
							AllowDeletions:   &falseVal,
							AllowForcePushes: &falseVal,
							EnforceAdmins:    &trueVal,
						},
						AllowUpdates:     &falseVal,
						RestrictCreation: &trueVal,
					},
				},
				"v2.0.0": {
					Name:      ptrString("v2.0.0"),
					Protected: &trueVal,
					TagProtectionRule: clients.TagProtectionRule{
						RefProtectionRule: clients.RefProtectionRule{
							AllowDeletions:   &falseVal,
							AllowForcePushes: &falseVal,
							EnforceAdmins:    &trueVal,
						},
						AllowUpdates:     &falseVal,
						RestrictCreation: &trueVal,
					},
				},
				"v3.1.5": {
					Name:      ptrString("v3.1.5"),
					Protected: &trueVal,
					TagProtectionRule: clients.TagProtectionRule{
						RefProtectionRule: clients.RefProtectionRule{
							AllowDeletions:   &falseVal,
							AllowForcePushes: &falseVal,
							EnforceAdmins:    &trueVal,
						},
						AllowUpdates:     &falseVal,
						RestrictCreation: &trueVal,
					},
				},
				"release-2024": {
					Name:      ptrString("release-2024"),
					Protected: &falseVal,
					TagProtectionRule: clients.TagProtectionRule{
						RefProtectionRule: clients.RefProtectionRule{
							AllowDeletions:   &trueVal,
							AllowForcePushes: &trueVal,
							EnforceAdmins:    &falseVal,
						},
					},
				},
			},
			expectedTagsLen: 4,
			wantErr:         false,
		},
		{
			// This test verifies the check correctly processes a mix of pattern-matched
			// (protected) and non-matching (unprotected) tags. It simulates a scenario
			// where some tags match a protection pattern (e.g., "v*") while others don't.
			name: "Mixed patterns - some tags match pattern, others don't",
			releases: []clients.Release{
				{TagName: "v1.0.0"},   // Matches v* pattern
				{TagName: "test-tag"}, // Doesn't match any pattern
				{TagName: "v2.5.0"},   // Matches v* pattern
				{TagName: "random-1"}, // Doesn't match any pattern
			},
			tags: map[string]*clients.TagRef{
				"v1.0.0": {
					Name:      ptrString("v1.0.0"),
					Protected: &trueVal,
					TagProtectionRule: clients.TagProtectionRule{
						RefProtectionRule: clients.RefProtectionRule{
							AllowDeletions: &falseVal,
							EnforceAdmins:  &trueVal,
						},
						RestrictCreation: &trueVal,
					},
				},
				"test-tag": {
					Name:      ptrString("test-tag"),
					Protected: &falseVal,
				},
				"v2.5.0": {
					Name:      ptrString("v2.5.0"),
					Protected: &trueVal,
					TagProtectionRule: clients.TagProtectionRule{
						RefProtectionRule: clients.RefProtectionRule{
							AllowDeletions: &falseVal,
							EnforceAdmins:  &trueVal,
						},
						RestrictCreation: &trueVal,
					},
				},
				"random-1": {
					Name:      ptrString("random-1"),
					Protected: &falseVal,
				},
			},
			expectedTagsLen: 4,
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockRepo := setupMockRepo(
				ctrl,
				tt.releases,
				tt.releasesErr,
				tt.tags,
				tt.getTagErr,
			)

			req := &checker.CheckRequest{
				RepoClient: mockRepo,
			}

			result, err := TagProtection(req)

			if (err != nil) != tt.wantErr {
				t.Errorf("TagProtection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				verifyTagResults(t, result, tt.expectedTagsLen, tt.tags)
			}

			ctrl.Finish()
		})
	}
}
