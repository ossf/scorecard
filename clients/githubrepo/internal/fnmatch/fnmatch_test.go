// Copyright 2023 OpenSSF Scorecard Authors
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

package fnmatch

import (
	"testing"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		want    bool
	}{
		// Taken from https://ruby-doc.org/core-2.5.1/File.html#method-c-fnmatch, assumes File::FNM_PATHNAME
		{"cat", "cat", true},
		{"cat", "category", false},
		{"c?t", "cat", true},
		{"c??t", "cat", false},
		{"c*", "cats", true},
		{"ca[a-z]", "cat", true},
		{"cat", "CAT", false},
		{"?", "/", false},
		{"*", "/", false},

		{"\\?", "?", true},
		{"\\a", "a", true},

		{"foo.bar", "foo.bar", true},
		{"foo.bar", "foo:bar", false},

		{"**.rb", "main.rb", true},
		{"**.rb", "./main.rb", false},
		{"**.rb", "lib/song.rb", false},

		{"**/foo", "a/b/c/foo", true},
		{"**foo", "a/b/c/foo", false},

		{"main", "main", true},
		{"releases/**/*", "releases/v2", true},
		{"releases/**/**/*", "releases/v2", true},
		{"releases/**/bar/**/qux", "releases/foo/bar/baz/qux", true},
		{"users/**/*", "users/foo/bar", true},
	}
	for _, tt := range tests {
		got, err := Match(tt.pattern, tt.path)
		if err != nil {
			t.Errorf("match: %v", err)
		}
		if got != tt.want {
			t.Errorf("Match(%q, %q) = %t, wanted %t", tt.pattern, tt.path, got, tt.want)
		}
	}
}
