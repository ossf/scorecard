package scorecardref

import "testing"

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
		{" ossf/scorecard ", "github.com", "ossf", "scorecard", false},
		{"bitbucket.org/foo/bar", "", "", "", true},
		{"justone", "", "", "", true},
		{"a/b/c/d", "", "", "", true},
		{"github.com//scorecard", "", "", "", true},
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
