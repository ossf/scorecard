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

package mttu

import (
	"testing"
	"time"

	"github.com/ossf/scorecard/v5/checker"
)

// Convenience: encode a duration using epoch+duration.
func enc(d time.Duration) time.Time { return epoch.Add(d) }

//nolint:paralleltest // utility test is independent
func Test_decodeDuration(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		if got := decodeDuration(time.Time{}); got != 0 {
			t.Fatalf("want 0, got %v", got)
		}
	})
	t.Run("negative-clamped", func(t *testing.T) {
		neg := epoch.Add(-time.Hour)
		if got := decodeDuration(neg); got != 0 {
			t.Fatalf("want 0, got %v", got)
		}
	})
	t.Run("positive", func(t *testing.T) {
		one := enc(time.Hour)
		if got := decodeDuration(one); got != time.Hour {
			t.Fatalf("want 1h, got %v", got)
		}
	})
}

//nolint:paralleltest // utility test is independent
func TestMeanTimeSinceFirstNewer(t *testing.T) {
	boolp := func(b bool) *bool { return &b }

	tests := []struct {
		name string
		deps []checker.LockDependency
		want time.Duration
		err  bool
	}{
		{
			name: "all-latest-contribute-zero",
			deps: []checker.LockDependency{
				{Name: "a", Version: "1.0.0", IsLatest: boolp(true)},
				{Name: "b", Version: "2.0.0", IsLatest: boolp(true)},
			},
			want: 0,
		},
		{
			name: "mix-latest-and-stale",
			deps: []checker.LockDependency{
				{Name: "a", Version: "1.0.0", IsLatest: boolp(true)}, // 0
				{Name: "b", Version: "1.0.0", IsLatest: boolp(false), TimeSinceOldestReleast: enc(7 * 24 * time.Hour)},
			},
			want: (0 + 7*24*time.Hour) / 2,
		},
		{
			name: "missing-duration-for-stale-dep-is-ignored",
			deps: []checker.LockDependency{
				{Name: "a", Version: "1.0.0", IsLatest: boolp(false)}, // no encoded time → skipped
				{Name: "b", Version: "1.0.0", IsLatest: boolp(false), TimeSinceOldestReleast: enc(48 * time.Hour)},
			},
			want: 48 * time.Hour,
		},
		{
			name: "no-usable-data",
			deps: []checker.LockDependency{
				{Name: "a", Version: "1.0.0", IsLatest: nil, TimeSinceOldestReleast: time.Time{}},
			},
			err: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			raw := &checker.RawResults{
				MTTUDependenciesResults: checker.MTTUDependenciesData{
					Dependencies: tc.deps,
				},
			}
			got, _, _, err := MeanTimeSinceFirstNewer(raw)
			if tc.err {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("mean mismatch: want %v, got %v", tc.want, got)
			}
		})
	}
}
