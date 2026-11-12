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

package ciiowner

import "testing"

func TestSameProject(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		candidate string
		badge     string
		want      bool
	}{
		{
			name:      "same org, different repos",
			candidate: "github.com/ossf/scorecard",
			badge:     "github.com/ossf/allstar",
			want:      true,
		},
		{
			name:      "same repo",
			candidate: "github.com/ossf/scorecard",
			badge:     "github.com/ossf/scorecard",
			want:      true,
		},
		{
			name:      "different org",
			candidate: "github.com/evil/scorecard",
			badge:     "github.com/ossf/scorecard",
			want:      false,
		},
		{
			name:      "prefix attack: org vs org-evil",
			candidate: "github.com/ossf-evil/repo",
			badge:     "github.com/ossf/scorecard",
			want:      false,
		},
		{
			name:      "different host",
			candidate: "gitlab.com/ossf/scorecard",
			badge:     "github.com/ossf/scorecard",
			want:      false,
		},
		{
			name:      "case insensitive",
			candidate: "github.com/OSSF/scorecard",
			badge:     "github.com/ossf/allstar",
			want:      true,
		},
		{
			name:      ".git suffix stripped",
			candidate: "github.com/ossf/scorecard.git",
			badge:     "github.com/ossf/allstar",
			want:      true,
		},
		{
			name:      "http vs https normalised",
			candidate: "http://github.com/ossf/scorecard",
			badge:     "https://github.com/ossf/allstar",
			want:      true,
		},
		{
			name:      "port 443 treated as no port",
			candidate: "https://github.com:443/ossf/scorecard",
			badge:     "github.com/ossf/allstar",
			want:      true,
		},
		{
			name:      "non-standard port differs",
			candidate: "https://github.com:8080/ossf/scorecard",
			badge:     "github.com/ossf/allstar",
			want:      false,
		},
		{
			name:      "userinfo trick rejected",
			candidate: "https://github.com@evil.com/ossf/x",
			badge:     "github.com/ossf/scorecard",
			want:      false,
		},
		{
			name:      "empty candidate",
			candidate: "",
			badge:     "github.com/ossf/scorecard",
			want:      false,
		},
		{
			name:      "single-segment path rejected",
			candidate: "github.com/ossf",
			badge:     "github.com/ossf/scorecard",
			want:      false,
		},
		{
			name:      "GitLab same subgroup",
			candidate: "gitlab.com/group/sub1/repo-a",
			badge:     "gitlab.com/group/sub1/repo-b",
			want:      true,
		},
		{
			name:      "GitLab different subgroup",
			candidate: "gitlab.com/group/sub1/repo-a",
			badge:     "gitlab.com/group/sub2/repo-b",
			want:      false,
		},
		{
			name:      "trailing slash stripped",
			candidate: "github.com/ossf/scorecard/",
			badge:     "github.com/ossf/allstar",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := SameProject(tt.candidate, tt.badge)
			if got != tt.want {
				t.Errorf("SameProject(%q, %q) = %v, want %v",
					tt.candidate, tt.badge, got, tt.want)
			}
		})
	}
}
