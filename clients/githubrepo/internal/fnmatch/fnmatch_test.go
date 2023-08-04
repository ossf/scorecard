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
