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

package scorecardref

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in       string
		platform string
		org      string
		name     string
		wantErr  bool
	}{
		{"ossf/scorecard", "github.com", "ossf", "scorecard", false},
		{"github.com/ossf/scorecard", "github.com", "ossf", "scorecard", false},
		{"gitlab.com/foo/bar", "gitlab.com", "foo", "bar", false},
		{"github.com/ossf/scorecard/", "github.com", "ossf", "scorecard", false},
		{"https://github.com/ossf/scorecard", "github.com", "ossf", "scorecard", false},
		{"https://github.com/ossf/scorecard.git", "github.com", "ossf", "scorecard", false},
		{"HTTPS://github.com/ossf/scorecard", "github.com", "ossf", "scorecard", false},
		{" ossf/scorecard ", "github.com", "ossf", "scorecard", false},
		{"bitbucket.org/foo/bar", "", "", "", true},
		{"justone", "", "", "", true},
		{"a/b/c/d", "", "", "", true},
		{"github.com//scorecard", "", "", "", true},
		{"github.com/../scorecard", "", "", "", true},
		{"github.com/ossf/..", "", "", "", true},
		{"github.com/./scorecard", "", "", "", true},
		{"gitlab.com/group/subgroup/repo", "", "", "", true},
		{"", "", "", "", true},
	}
	for _, tt := range tests {
		got, err := Parse(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Errorf("Parse(%q): expected error, got %+v", tt.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("Parse(%q): unexpected error: %v", tt.in, err)
			continue
		}
		if got.Platform != tt.platform || got.Org != tt.org || got.Name != tt.name {
			t.Errorf("Parse(%q) = %+v, want %s/%s/%s", tt.in, got, tt.platform, tt.org, tt.name)
		}
	}
	_, err := Parse("gitlab.com/group/subgroup/repo")
	if err == nil || !strings.Contains(err.Error(), "GitLab subgroup paths are not supported") {
		t.Errorf("Parse(GitLab subgroup) error = %v, want a specific unsupported-subgroup error", err)
	}
}

func TestValidateCommit(t *testing.T) {
	t.Parallel()
	valid := "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0"
	if err := ValidateCommit(valid); err != nil {
		t.Errorf("ValidateCommit(valid) unexpected error: %v", err)
	}
	if err := ValidateCommit(""); err != nil {
		t.Errorf("ValidateCommit(empty) should be nil, got: %v", err)
	}
	for _, bad := range []string{"abc", "g" + valid[1:], valid + "0", valid[:39]} {
		if err := ValidateCommit(bad); err == nil {
			t.Errorf("ValidateCommit(%q): expected error", bad)
		}
	}
}
