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
	"encoding/json"
	"testing"
	"time"

	depsdev "github.com/ossf/scorecard/v5/internal/packageclient"
)

// TestVersionComparison_OfflineValidation tests version comparison logic
// using pre-captured API responses (no live API calls).
// This test validates the newestInfoFor function with real-world test cases,
// covering various scenarios from latest versions to very outdated dependencies.
//
// IMPORTANT: newestInfoFor returns the timestamp of the OLDEST NEWER version,
// not the latest version. This provides a fair "time to update" measurement.
// It also filters out Go pseudo-versions (unreleased commits).
//
//nolint:gocognit // test function with many validation cases
func TestVersionComparison_OfflineValidation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		packageName    string
		currentVersion string
		versionsJSON   string // Raw JSON from deps.dev API
		expectIsLatest bool
		expectDaysOld  int // Expected days between current and OLDEST NEWER version (not latest!)
		daysTolerance  int // Acceptable tolerance due to test time precision
	}{
		// Test cases 1-2: Latest versions
		{
			name:           "go-cmp v0.7.0 is latest",
			packageName:    "github.com/google/go-cmp",
			currentVersion: "0.7.0",
			versionsJSON:   goCmpVersionsJSON,
			expectIsLatest: true,
		},
		{
			name:           "go-git/v5 v5.16.3 is latest",
			packageName:    "github.com/go-git/go-git/v5",
			currentVersion: "5.16.3",
			versionsJSON:   goGitV5VersionsJSON,
			expectIsLatest: true,
		},

		// Test cases 3-10: Recent dependencies (< 200 days old)
		{
			name:           "buildkit v0.25.1 - 1 day behind latest",
			packageName:    "github.com/moby/buildkit",
			currentVersion: "0.25.1",
			versionsJSON:   buildkitVersionsJSON,
			expectIsLatest: false,
			expectDaysOld:  1,
			daysTolerance:  1,
		},
		{
			name:           "cobra v1.10.1 - is latest (pseudo-version v1.10.2-0.xxx filtered)",
			packageName:    "github.com/spf13/cobra",
			currentVersion: "1.10.1",
			versionsJSON:   cobraVersionsJSON,
			expectIsLatest: true, // v1.10.2-0.xxx is a pseudo-version, filtered out
		},
		{
			name:           "go-containerregistry v0.20.6 - latest (next is pseudo-version)",
			packageName:    "github.com/google/go-containerregistry",
			currentVersion: "0.20.6",
			versionsJSON:   goContainerRegistryVersionsJSON,
			expectIsLatest: true,
			expectDaysOld:  0,
			daysTolerance:  1,
		},
		{
			name:           "semver/v3 v3.4.0 - latest (next is pseudo-version)",
			packageName:    "github.com/Masterminds/semver/v3",
			currentVersion: "3.4.0",
			versionsJSON:   semverV3VersionsJSON,
			expectIsLatest: true,
			expectDaysOld:  0,
			daysTolerance:  1,
		},
		{
			name:           "logr v1.4.3 - latest (next is pseudo-version)",
			packageName:    "github.com/go-logr/logr",
			currentVersion: "1.4.3",
			versionsJSON:   logrVersionsJSON,
			expectIsLatest: true,
			expectDaysOld:  0,
			daysTolerance:  2,
		},
		{
			name:           "go-git/v5 v5.16.2 - 102 days behind latest",
			packageName:    "github.com/go-git/go-git/v5",
			currentVersion: "5.16.2",
			versionsJSON:   goGitV5VersionsJSON,
			expectIsLatest: false,
			expectDaysOld:  102,
			daysTolerance:  1,
		},
		{
			name:           "go-git/v5 v5.16.1 - 2 days to v5.16.2 (oldest newer)",
			packageName:    "github.com/go-git/go-git/v5",
			currentVersion: "5.16.1",
			versionsJSON:   goGitV5VersionsJSON,
			expectIsLatest: false,
			expectDaysOld:  2,
			daysTolerance:  2,
		},
		{
			name:           "go-git/v5 v5.16.0 - 11 days to v5.16.1 (oldest newer)",
			packageName:    "github.com/go-git/go-git/v5",
			currentVersion: "5.16.0",
			versionsJSON:   goGitV5VersionsJSON,
			expectIsLatest: false,
			expectDaysOld:  11,
			daysTolerance:  2,
		},

		// Test cases 11-20: Moderately outdated (200-600 days)
		{
			name:           "backoff/v4 v4.3.0 - latest (next is pseudo-version)",
			packageName:    "github.com/cenkalti/backoff/v4",
			currentVersion: "4.3.0",
			versionsJSON:   backoffV4VersionsJSON,
			expectIsLatest: true,
			expectDaysOld:  0,
			daysTolerance:  2,
		},
		{
			name:           "xxhash/v2 v2.3.0 - latest (next is pseudo-version)",
			packageName:    "github.com/cespare/xxhash/v2",
			currentVersion: "2.3.0",
			versionsJSON:   xxhashV2VersionsJSON,
			expectIsLatest: true,
			expectDaysOld:  0,
			daysTolerance:  2,
		},
		{
			name:           "buildkit v0.25.0 - 26 days to v0.25.1 (oldest newer)",
			packageName:    "github.com/moby/buildkit",
			currentVersion: "0.25.0",
			versionsJSON:   buildkitVersionsJSON,
			expectIsLatest: false,
			expectDaysOld:  26,
			daysTolerance:  1,
		},
		{
			name:           "buildkit v0.24.0 - 29 days to v0.25.0 (oldest newer)",
			packageName:    "github.com/moby/buildkit",
			currentVersion: "0.24.0",
			versionsJSON:   buildkitVersionsJSON,
			expectIsLatest: false,
			expectDaysOld:  29,
			daysTolerance:  2,
		},
		{
			name:           "buildkit v0.23.0 - 56 days to v0.24.0 (oldest newer)",
			packageName:    "github.com/moby/buildkit",
			currentVersion: "0.23.0",
			versionsJSON:   buildkitVersionsJSON,
			expectIsLatest: false,
			expectDaysOld:  56,
			daysTolerance:  2,
		},
		{
			name:           "buildkit v0.22.0 - 92 days to v0.23.0 (oldest newer)",
			packageName:    "github.com/moby/buildkit",
			currentVersion: "0.22.0",
			versionsJSON:   buildkitVersionsJSON,
			expectIsLatest: false,
			expectDaysOld:  92,
			daysTolerance:  2,
		},
		{
			name:           "buildkit v0.21.0 - 90 days to v0.22.0 (oldest newer)",
			packageName:    "github.com/moby/buildkit",
			currentVersion: "0.21.0",
			versionsJSON:   buildkitVersionsJSON,
			expectIsLatest: false,
			expectDaysOld:  90,
			daysTolerance:  2,
		},
		{
			name:           "buildkit v0.20.0 - 140 days to v0.21.0 (oldest newer)",
			packageName:    "github.com/moby/buildkit",
			currentVersion: "0.20.0",
			versionsJSON:   buildkitVersionsJSON,
			expectIsLatest: false,
			expectDaysOld:  140,
			daysTolerance:  2,
		},
		{
			name:           "buildkit v0.19.0 - 103 days to v0.20.0 (oldest newer)",
			packageName:    "github.com/moby/buildkit",
			currentVersion: "0.19.0",
			versionsJSON:   buildkitVersionsJSON,
			expectIsLatest: false,
			expectDaysOld:  103,
			daysTolerance:  2,
		},
		{
			name:           "buildkit v0.18.0 - 62 days to v0.19.0 (oldest newer)",
			packageName:    "github.com/moby/buildkit",
			currentVersion: "0.18.0",
			versionsJSON:   buildkitVersionsJSON,
			expectIsLatest: false,
			expectDaysOld:  62,
			daysTolerance:  2,
		},

		// Test cases 21-33: Very outdated (> 600 days)
		{
			name:           "uuid v1.6.0 - latest (next is pseudo-version)",
			packageName:    "github.com/google/uuid",
			currentVersion: "1.6.0",
			versionsJSON:   uuidVersionsJSON,
			expectIsLatest: true,
			expectDaysOld:  0,
			daysTolerance:  2,
		},
		{
			name:           "buildkit v0.17.0 - 69 days to v0.18.0 (oldest newer)",
			packageName:    "github.com/moby/buildkit",
			currentVersion: "0.17.0",
			versionsJSON:   buildkitVersionsJSON,
			expectIsLatest: false,
			expectDaysOld:  69,
			daysTolerance:  2,
		},
		{
			name:           "buildkit v0.16.0 - 63 days to v0.17.0 (oldest newer)",
			packageName:    "github.com/moby/buildkit",
			currentVersion: "0.16.0",
			versionsJSON:   buildkitVersionsJSON,
			expectIsLatest: false,
			expectDaysOld:  63,
			daysTolerance:  2,
		},
		{
			name:           "cobra v1.10.0 - 0 days to v1.10.1 (oldest newer)",
			packageName:    "github.com/spf13/cobra",
			currentVersion: "1.10.0",
			versionsJSON:   cobraVersionsJSON,
			expectIsLatest: false,
			expectDaysOld:  0,
			daysTolerance:  1,
		},
		{
			name:           "cobra v1.9.1 - 194 days to v1.10.0 (oldest newer)",
			packageName:    "github.com/spf13/cobra",
			currentVersion: "1.9.1",
			versionsJSON:   cobraVersionsJSON,
			expectIsLatest: false,
			expectDaysOld:  194,
			daysTolerance:  2,
		},
		{
			name:           "cobra v1.9.0 - 13 days to v1.9.1 (oldest newer)",
			packageName:    "github.com/spf13/cobra",
			currentVersion: "1.9.0",
			versionsJSON:   cobraVersionsJSON,
			expectIsLatest: false,
			expectDaysOld:  13,
			daysTolerance:  2,
		},
		{
			name:           "cobra v1.8.1 - 87 days to v1.9.0 (oldest newer)",
			packageName:    "github.com/spf13/cobra",
			currentVersion: "1.8.1",
			versionsJSON:   cobraVersionsJSON,
			expectIsLatest: false,
			expectDaysOld:  87,
			daysTolerance:  2,
		},
		{
			name:           "cobra v1.8.0 - 284 days to v1.8.1 (oldest newer)",
			packageName:    "github.com/spf13/cobra",
			currentVersion: "1.8.0",
			versionsJSON:   cobraVersionsJSON,
			expectIsLatest: false,
			expectDaysOld:  284,
			daysTolerance:  2,
		},
		{
			name:           "cobra v1.7.0 - 258 days to v1.8.0 (oldest newer)",
			packageName:    "github.com/spf13/cobra",
			currentVersion: "1.7.0",
			versionsJSON:   cobraVersionsJSON,
			expectIsLatest: false,
			expectDaysOld:  258,
			daysTolerance:  2,
		},
		{
			name:           "cobra v1.6.1 - 180 days to v1.7.0 (oldest newer)",
			packageName:    "github.com/spf13/cobra",
			currentVersion: "1.6.1",
			versionsJSON:   cobraVersionsJSON,
			expectIsLatest: false,
			expectDaysOld:  180,
			daysTolerance:  2,
		},
		{
			name:           "cobra v1.6.0 - 15 days to v1.6.1 (oldest newer)",
			packageName:    "github.com/spf13/cobra",
			currentVersion: "1.6.0",
			versionsJSON:   cobraVersionsJSON,
			expectIsLatest: false,
			expectDaysOld:  15,
			daysTolerance:  2,
		},
		{
			name:           "cobra v1.5.0 - 121 days to v1.6.0 (oldest newer)",
			packageName:    "github.com/spf13/cobra",
			currentVersion: "1.5.0",
			versionsJSON:   cobraVersionsJSON,
			expectIsLatest: false,
			expectDaysOld:  121,
			daysTolerance:  2,
		},
		{
			name:           "cobra v1.4.0 - 150 days to v1.5.0 (oldest newer)",
			packageName:    "github.com/spf13/cobra",
			currentVersion: "1.4.0",
			versionsJSON:   cobraVersionsJSON,
			expectIsLatest: false,
			expectDaysOld:  150,
			daysTolerance:  2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Parse the captured API response
			var versions []depsdev.Version
			if err := json.Unmarshal([]byte(tc.versionsJSON), &versions); err != nil {
				t.Fatalf("Failed to parse versions JSON: %v", err)
			}

			t.Logf("Testing %s version %s", tc.packageName, tc.currentVersion)
			t.Logf("Total versions from captured API response: %d", len(versions))

			// Test our newestInfoFor function
			isLatest, oldestNewerPublishedAt := newestInfoFor(tc.currentVersion, versions)

			// Verify isLatest
			if isLatest != tc.expectIsLatest {
				t.Errorf("IsLatest mismatch: got %v, want %v", isLatest, tc.expectIsLatest)
			}

			// Verify days old calculation (to oldest newer version)
			//nolint:nestif // test validation requires nested structure
			if !tc.expectIsLatest {
				if oldestNewerPublishedAt == nil {
					t.Error("Expected oldestNewerPublishedAt to be set for outdated version")
					return
				}

				// Find current version's publish date
				var currentPub time.Time
				for _, v := range versions {
					// Normalize version strings for comparison (handle with/without 'v' prefix)
					versionToMatch := tc.currentVersion
					versionInList := v.VersionKey.Version
					if versionToMatch[0] != 'v' {
						versionToMatch = "v" + versionToMatch
					}
					if versionInList[0] != 'v' && len(versionInList) > 0 {
						versionInList = "v" + versionInList
					}

					if versionInList == versionToMatch {
						var err error
						currentPub, err = time.Parse(time.RFC3339, v.PublishedAt)
						if err != nil {
							t.Errorf("Failed to parse current version publish date: %v", err)
							return
						}
						break
					}
				}

				if currentPub.IsZero() {
					t.Error("Could not find current version in versions list")
					return
				}

				delta := oldestNewerPublishedAt.Sub(currentPub)
				days := int(delta.Hours() / 24)

				t.Logf("Result: %d days between current and oldest newer (expected ~%d)", days, tc.expectDaysOld)
				t.Logf("Current version published: %s", currentPub.Format(time.RFC3339))
				t.Logf("Oldest newer version published: %s", oldestNewerPublishedAt.Format(time.RFC3339))

				diff := days - tc.expectDaysOld
				if diff < 0 {
					diff = -diff
				}

				if diff > tc.daysTolerance {
					t.Errorf("Days old mismatch: got %d days, expected %d days (tolerance: %d)",
						days, tc.expectDaysOld, tc.daysTolerance)
				}
			} else {
				t.Logf("Result: IS LATEST VERSION")
			}
		})
	}
}

// Captured API responses from deps.dev (via depsclient.GetPackage)
// These are real responses captured on 2025-11-10
// Format matches depsdev.Version struct: {VersionKey:{Version:"x.y.z"}, PublishedAt:"RFC3339"}

const goCmpVersionsJSON = `[
  {"versionKey":{"system":"GO","name":"github.com/google/go-cmp","version":"v0.5.8-0.20250114181544-9b12f366a942"},"publishedAt":"2025-01-14T18:15:44Z"},
  {"versionKey":{"system":"GO","name":"github.com/google/go-cmp","version":"v0.7.0"},"publishedAt":"2024-08-23T18:33:36Z"},
  {"versionKey":{"system":"GO","name":"github.com/google/go-cmp","version":"v0.6.0"},"publishedAt":"2023-05-03T19:29:09Z"},
  {"versionKey":{"system":"GO","name":"github.com/google/go-cmp","version":"v0.5.9"},"publishedAt":"2022-11-05T00:04:31Z"},
  {"versionKey":{"system":"GO","name":"github.com/google/go-cmp","version":"v0.5.8"},"publishedAt":"2022-05-12T02:16:49Z"}
]`

const uuidVersionsJSON = `[
  {"versionKey":{"system":"GO","name":"github.com/google/uuid","version":"v1.6.1-0.20241114170450-2d3c2a9cc518"},"publishedAt":"2024-11-14T17:04:50Z"},
  {"versionKey":{"system":"GO","name":"github.com/google/uuid","version":"v0.0.0-20241114170450-2d3c2a9cc518"},"publishedAt":"2024-11-14T17:04:50Z"},
  {"versionKey":{"system":"GO","name":"github.com/google/uuid","version":"v1.6.1-0.20240806143732-cf44fad6f0fe"},"publishedAt":"2024-08-06T14:37:32Z"},
  {"versionKey":{"system":"GO","name":"github.com/google/uuid","version":"v1.6.1-0.20240806143717-0e97ed3b5379"},"publishedAt":"2024-08-06T14:37:17Z"},
  {"versionKey":{"system":"GO","name":"github.com/google/uuid","version":"v1.6.1-0.20240701162350-d55c313874fe"},"publishedAt":"2024-07-01T16:23:50Z"},
  {"versionKey":{"system":"GO","name":"github.com/google/uuid","version":"v0.0.0-20240701162350-d55c313874fe"},"publishedAt":"2024-07-01T16:23:50Z"},
  {"versionKey":{"system":"GO","name":"github.com/google/uuid","version":"v1.6.1-0.20240701161543-e8d82d30a3eb"},"publishedAt":"2024-07-01T16:15:43Z"},
  {"versionKey":{"system":"GO","name":"github.com/google/uuid","version":"v1.6.1-0.20240603173921-53dda83ebe99"},"publishedAt":"2024-06-03T17:39:21Z"},
  {"versionKey":{"system":"GO","name":"github.com/google/uuid","version":"v0.0.0-20240222164149-6e10cd1027e2"},"publishedAt":"2024-02-22T16:41:49Z"},
  {"versionKey":{"system":"GO","name":"github.com/google/uuid","version":"v1.6.1-0.20240222164149-6e10cd1027e2"},"publishedAt":"2024-02-22T16:41:49Z"},
  {"versionKey":{"system":"GO","name":"github.com/google/uuid","version":"v1.6.1-0.20240212160722-b5b9aeb1d146"},"publishedAt":"2024-02-12T16:07:22Z"},
  {"versionKey":{"system":"GO","name":"github.com/google/uuid","version":"v1.6.0"},"publishedAt":"2024-01-16T19:19:00Z"},
  {"versionKey":{"system":"GO","name":"github.com/google/uuid","version":"v1.5.0"},"publishedAt":"2023-12-12T16:22:53Z"}
]`

const backoffV4VersionsJSON = `[
  {"versionKey":{"system":"GO","name":"github.com/cenkalti/backoff/v4","version":"v4.3.1-0.20241216033218-258846da02f4"},"publishedAt":"2024-12-16T03:32:18Z"},
  {"versionKey":{"system":"GO","name":"github.com/cenkalti/backoff/v4","version":"v4.3.0"},"publishedAt":"2024-01-15T22:28:23Z"},
  {"versionKey":{"system":"GO","name":"github.com/cenkalti/backoff/v4","version":"v4.2.1"},"publishedAt":"2023-03-14T09:25:16Z"},
  {"versionKey":{"system":"GO","name":"github.com/cenkalti/backoff/v4","version":"v4.2.0"},"publishedAt":"2022-12-17T13:51:06Z"},
  {"versionKey":{"system":"GO","name":"github.com/cenkalti/backoff/v4","version":"v4.1.3"},"publishedAt":"2022-04-28T07:30:39Z"}
]`

const cobraVersionsJSON = `[
  {"versionKey":{"system":"GO","name":"github.com/spf13/cobra","version":"v1.10.2-0.20251010135735-4c363afb59d5"},"publishedAt":"2025-10-10T13:57:35Z"},
  {"versionKey":{"system":"GO","name":"github.com/spf13/cobra","version":"v1.10.1"},"publishedAt":"2025-09-01T16:19:51Z"},
  {"versionKey":{"system":"GO","name":"github.com/spf13/cobra","version":"v1.10.0"},"publishedAt":"2025-08-31T23:15:08Z"},
  {"versionKey":{"system":"GO","name":"github.com/spf13/cobra","version":"v1.9.1"},"publishedAt":"2025-02-18T19:32:38Z"},
  {"versionKey":{"system":"GO","name":"github.com/spf13/cobra","version":"v1.9.0"},"publishedAt":"2025-02-05T16:51:13Z"},
  {"versionKey":{"system":"GO","name":"github.com/spf13/cobra","version":"v1.8.1"},"publishedAt":"2024-11-10T15:18:04Z"},
  {"versionKey":{"system":"GO","name":"github.com/spf13/cobra","version":"v1.8.0"},"publishedAt":"2024-01-30T18:57:44Z"},
  {"versionKey":{"system":"GO","name":"github.com/spf13/cobra","version":"v1.7.0"},"publishedAt":"2023-05-17T17:16:26Z"},
  {"versionKey":{"system":"GO","name":"github.com/spf13/cobra","version":"v1.6.1"},"publishedAt":"2022-11-17T22:02:53Z"},
  {"versionKey":{"system":"GO","name":"github.com/spf13/cobra","version":"v1.6.0"},"publishedAt":"2022-11-02T17:34:10Z"},
  {"versionKey":{"system":"GO","name":"github.com/spf13/cobra","version":"v1.5.0"},"publishedAt":"2022-07-04T11:01:52Z"},
  {"versionKey":{"system":"GO","name":"github.com/spf13/cobra","version":"v1.4.0"},"publishedAt":"2022-02-03T14:47:36Z"},
  {"versionKey":{"system":"GO","name":"github.com/spf13/cobra","version":"v1.3.0"},"publishedAt":"2021-12-27T13:15:00Z"},
  {"versionKey":{"system":"GO","name":"github.com/spf13/cobra","version":"v1.2.1"},"publishedAt":"2021-08-03T18:01:34Z"},
  {"versionKey":{"system":"GO","name":"github.com/spf13/cobra","version":"v1.2.0"},"publishedAt":"2021-07-28T18:26:26Z"}
]`

const logrVersionsJSON = `[
  {"versionKey":{"system":"GO","name":"github.com/go-logr/logr","version":"v1.4.4-0.20251103063727-c29d895a2b3f"},"publishedAt":"2025-11-03T06:37:27Z"},
  {"versionKey":{"system":"GO","name":"github.com/go-logr/logr","version":"v1.4.4-0.20251027064622-e28ca26c967c"},"publishedAt":"2025-10-27T06:46:22Z"},
  {"versionKey":{"system":"GO","name":"github.com/go-logr/logr","version":"v1.4.4-0.20251020155725-604d54d8fcaa"},"publishedAt":"2025-10-20T15:57:25Z"},
  {"versionKey":{"system":"GO","name":"github.com/go-logr/logr","version":"v1.4.4-0.20251015075504-f9181c41b310"},"publishedAt":"2025-10-15T07:55:04Z"},
  {"versionKey":{"system":"GO","name":"github.com/go-logr/logr","version":"v1.4.4-0.20251006060816-034c819a9ec3"},"publishedAt":"2025-10-06T06:08:16Z"},
  {"versionKey":{"system":"GO","name":"github.com/go-logr/logr","version":"v1.4.4-0.20250929160248-622a7cb1577d"},"publishedAt":"2025-09-29T16:02:48Z"},
  {"versionKey":{"system":"GO","name":"github.com/go-logr/logr","version":"v1.4.4-0.20250915042438-41ac1c6cf14c"},"publishedAt":"2025-09-15T04:24:38Z"},
  {"versionKey":{"system":"GO","name":"github.com/go-logr/logr","version":"v1.4.4-0.20250818172122-502aeda2b1ee"},"publishedAt":"2025-08-18T17:21:22Z"},
  {"versionKey":{"system":"GO","name":"github.com/go-logr/logr","version":"v1.4.4-0.20250806110651-2ec7e2a95fcd"},"publishedAt":"2025-08-06T11:06:51Z"},
  {"versionKey":{"system":"GO","name":"github.com/go-logr/logr","version":"v1.4.4-0.20250715095740-f9c496a62c06"},"publishedAt":"2025-07-15T09:57:40Z"},
  {"versionKey":{"system":"GO","name":"github.com/go-logr/logr","version":"v1.4.4-0.20250701115622-4a7b5fce10e0"},"publishedAt":"2025-07-01T11:56:22Z"},
  {"versionKey":{"system":"GO","name":"github.com/go-logr/logr","version":"v1.4.4-0.20250617103518-edf7bdf1fa5a"},"publishedAt":"2025-06-17T10:35:18Z"},
  {"versionKey":{"system":"GO","name":"github.com/go-logr/logr","version":"v1.4.4-0.20250610092451-c067ccaa42e0"},"publishedAt":"2025-06-10T09:24:51Z"},
  {"versionKey":{"system":"GO","name":"github.com/go-logr/logr","version":"v1.4.4-0.20250603034946-e6c09eb8abbc"},"publishedAt":"2025-06-03T03:49:46Z"},
  {"versionKey":{"system":"GO","name":"github.com/go-logr/logr","version":"v1.4.4-0.20250602042353-7e799854f8dd"},"publishedAt":"2025-06-02T04:23:53Z"},
  {"versionKey":{"system":"GO","name":"github.com/go-logr/logr","version":"v1.4.3"},"publishedAt":"2025-05-29T15:04:52Z"},
  {"versionKey":{"system":"GO","name":"github.com/go-logr/logr","version":"v1.4.2"},"publishedAt":"2024-04-29T11:40:52Z"}
]`

const goContainerRegistryVersionsJSON = `[
  {"versionKey":{"system":"GO","name":"github.com/google/go-containerregistry","version":"v0.20.7-0.20251028202801-aab7c77e9d78"},"publishedAt":"2025-10-28T20:28:01Z"},
  {"versionKey":{"system":"GO","name":"github.com/google/go-containerregistry","version":"v0.0.0-20251028202801-aab7c77e9d78"},"publishedAt":"2025-10-28T20:28:01Z"},
  {"versionKey":{"system":"GO","name":"github.com/google/go-containerregistry","version":"v0.0.0-20251010214028-cb4a037720c0"},"publishedAt":"2025-10-10T21:40:28Z"},
  {"versionKey":{"system":"GO","name":"github.com/google/go-containerregistry","version":"v0.20.7-0.20251010214028-cb4a037720c0"},"publishedAt":"2025-10-10T21:40:28Z"},
  {"versionKey":{"system":"GO","name":"github.com/google/go-containerregistry","version":"v0.20.7-0.20251003171851-d0099a1a8b77"},"publishedAt":"2025-10-03T17:18:51Z"},
  {"versionKey":{"system":"GO","name":"github.com/google/go-containerregistry","version":"v0.0.0-20251003171851-d0099a1a8b77"},"publishedAt":"2025-10-03T17:18:51Z"},
  {"versionKey":{"system":"GO","name":"github.com/google/go-containerregistry","version":"v0.0.0-20251002232900-5924946662b9"},"publishedAt":"2025-10-02T23:29:00Z"},
  {"versionKey":{"system":"GO","name":"github.com/google/go-containerregistry","version":"v0.20.7-0.20251002232900-5924946662b9"},"publishedAt":"2025-10-02T23:29:00Z"},
  {"versionKey":{"system":"GO","name":"github.com/google/go-containerregistry","version":"v0.20.7-0.20251002230642-8d9c9efa049a"},"publishedAt":"2025-10-02T23:06:42Z"},
  {"versionKey":{"system":"GO","name":"github.com/google/go-containerregistry","version":"v0.20.6"},"publishedAt":"2025-10-02T15:58:44Z"},
  {"versionKey":{"system":"GO","name":"github.com/google/go-containerregistry","version":"v0.20.5"},"publishedAt":"2025-09-24T14:57:13Z"}
]`

const buildkitVersionsJSON = `[
  {"versionKey":{"system":"GO","name":"github.com/moby/buildkit","version":"v0.26.0-rc1.0.20251107065059-474cae70a594"},"publishedAt":"2025-11-07T06:50:59Z"},
  {"versionKey":{"system":"GO","name":"github.com/moby/buildkit","version":"v0.26.0-rc1"},"publishedAt":"2025-11-05T23:49:43Z"},
  {"versionKey":{"system":"GO","name":"github.com/moby/buildkit","version":"v0.25.2"},"publishedAt":"2025-11-05T10:29:03Z"},
  {"versionKey":{"system":"GO","name":"github.com/moby/buildkit","version":"v0.25.1"},"publishedAt":"2025-11-05T09:48:00Z"},
  {"versionKey":{"system":"GO","name":"github.com/moby/buildkit","version":"v0.25.0"},"publishedAt":"2025-10-10T02:22:03Z"},
  {"versionKey":{"system":"GO","name":"github.com/moby/buildkit","version":"v0.24.0"},"publishedAt":"2025-09-10T19:55:56Z"},
  {"versionKey":{"system":"GO","name":"github.com/moby/buildkit","version":"v0.23.0"},"publishedAt":"2025-07-16T09:51:57Z"},
  {"versionKey":{"system":"GO","name":"github.com/moby/buildkit","version":"v0.22.0"},"publishedAt":"2025-04-15T07:26:29Z"},
  {"versionKey":{"system":"GO","name":"github.com/moby/buildkit","version":"v0.21.0"},"publishedAt":"2025-01-15T02:45:20Z"},
  {"versionKey":{"system":"GO","name":"github.com/moby/buildkit","version":"v0.20.0"},"publishedAt":"2024-08-27T19:08:23Z"},
  {"versionKey":{"system":"GO","name":"github.com/moby/buildkit","version":"v0.19.0"},"publishedAt":"2024-05-16T02:56:39Z"},
  {"versionKey":{"system":"GO","name":"github.com/moby/buildkit","version":"v0.18.0"},"publishedAt":"2024-03-14T03:02:18Z"},
  {"versionKey":{"system":"GO","name":"github.com/moby/buildkit","version":"v0.17.0"},"publishedAt":"2024-01-04T22:25:05Z"},
  {"versionKey":{"system":"GO","name":"github.com/moby/buildkit","version":"v0.16.0"},"publishedAt":"2023-11-02T03:03:05Z"},
  {"versionKey":{"system":"GO","name":"github.com/moby/buildkit","version":"v0.15.0"},"publishedAt":"2023-10-06T02:17:12Z"}
]`

const xxhashV2VersionsJSON = `[
  {"versionKey":{"system":"GO","name":"github.com/cespare/xxhash/v2","version":"v2.3.1-0.20240703180136-ab37246c889f"},"publishedAt":"2024-07-03T18:01:36Z"},
  {"versionKey":{"system":"GO","name":"github.com/cespare/xxhash/v2","version":"v2.3.0"},"publishedAt":"2024-04-04T20:00:10Z"},
  {"versionKey":{"system":"GO","name":"github.com/cespare/xxhash/v2","version":"v2.2.1-0.20230726053707-21fc82b0b9b5"},"publishedAt":"2023-07-26T05:37:07Z"},
  {"versionKey":{"system":"GO","name":"github.com/cespare/xxhash/v2","version":"v2.2.1-0.20230403140943-66b140914239"},"publishedAt":"2023-04-03T14:09:43Z"},
  {"versionKey":{"system":"GO","name":"github.com/cespare/xxhash/v2","version":"v2.2.0"},"publishedAt":"2022-12-04T02:06:23Z"}
]`

const semverV3VersionsJSON = `[
  {"versionKey":{"system":"GO","name":"github.com/Masterminds/semver/v3","version":"v3.4.1-0.20250707152403-bf01c6184597"},"publishedAt":"2025-07-07T15:24:03Z"},
  {"versionKey":{"system":"GO","name":"github.com/Masterminds/semver/v3","version":"v3.4.0"},"publishedAt":"2025-06-27T14:48:33Z"},
  {"versionKey":{"system":"GO","name":"github.com/Masterminds/semver/v3","version":"v3.3.1"},"publishedAt":"2024-11-19T20:00:22Z"},
  {"versionKey":{"system":"GO","name":"github.com/Masterminds/semver/v3","version":"v3.3.0"},"publishedAt":"2024-08-02T19:12:08Z"},
  {"versionKey":{"system":"GO","name":"github.com/Masterminds/semver/v3","version":"v3.2.1"},"publishedAt":"2023-05-19T15:30:05Z"}
]`

const goGitV5VersionsJSON = `[
  {"versionKey":{"system":"GO","name":"github.com/go-git/go-git/v5","version":"v5.16.3"},"publishedAt":"2025-09-17T10:22:49Z"},
  {"versionKey":{"system":"GO","name":"github.com/go-git/go-git/v5","version":"v5.16.2"},"publishedAt":"2025-06-06T16:15:07Z"},
  {"versionKey":{"system":"GO","name":"github.com/go-git/go-git/v5","version":"v5.16.1"},"publishedAt":"2025-06-04T09:28:06Z"},
  {"versionKey":{"system":"GO","name":"github.com/go-git/go-git/v5","version":"v5.16.0"},"publishedAt":"2025-05-23T10:57:28Z"},
  {"versionKey":{"system":"GO","name":"github.com/go-git/go-git/v5","version":"v5.15.0"},"publishedAt":"2025-03-04T19:19:04Z"},
  {"versionKey":{"system":"GO","name":"github.com/go-git/go-git/v5","version":"v5.14.0"},"publishedAt":"2024-11-08T12:11:36Z"}
]`
