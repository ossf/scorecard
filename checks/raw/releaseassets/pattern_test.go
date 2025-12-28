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

import "testing"

func TestMatchesPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		tag     string
		pattern string
		want    bool
	}{
		{
			name:    "Exact match",
			tag:     "v1.2.3",
			pattern: "v1.2.3",
			want:    true,
		},
		{
			name:    "Wildcard match v1.*",
			tag:     "v1.0.0",
			pattern: "v1.*",
			want:    true,
		},
		{
			name:    "Wildcard match v1.* with subversion",
			tag:     "v1.2.3",
			pattern: "v1.*",
			want:    true,
		},
		{
			name:    "Wildcard no match",
			tag:     "v2.0.0",
			pattern: "v1.*",
			want:    false,
		},
		{
			name:    "Double wildcard",
			tag:     "v2.0.1",
			pattern: "v2.0.*",
			want:    true,
		},
		{
			name:    "Match any with *",
			tag:     "v3.0.0",
			pattern: "*",
			want:    true,
		},
		{
			name:    "Empty pattern matches all",
			tag:     "v1.0.0",
			pattern: "",
			want:    true,
		},
		{
			name:    "Beta tag wildcard",
			tag:     "v1.0.0-beta",
			pattern: "v1.*",
			want:    true,
		},
		{
			name:    "Suffix wildcard",
			tag:     "v1.0.0-security",
			pattern: "*-security",
			want:    true,
		},
		{
			name:    "Middle wildcard",
			tag:     "v1-beta-2",
			pattern: "v1-*-2",
			want:    true,
		},
		{
			name:    "No match different prefix",
			tag:     "release-1.0",
			pattern: "v1.*",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := matchesPattern(tt.tag, tt.pattern)
			if got != tt.want {
				t.Errorf("matchesPattern(%q, %q) = %v, want %v", tt.tag, tt.pattern, got, tt.want)
			}
		})
	}
}
