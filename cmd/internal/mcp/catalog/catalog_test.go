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

package catalog

import (
	"maps"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/ossf/scorecard/v5/checks"
)

func TestCatalogList(t *testing.T) {
	t.Parallel()
	c, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	list := c.List()
	if len(list) < 10 {
		t.Fatalf("expected many checks, got %d", len(list))
	}
	for i, ci := range list {
		if ci.Name == "" {
			t.Errorf("check at index %d has empty name", i)
		}
		if i > 0 && list[i-1].Name > ci.Name {
			t.Errorf("checks not sorted: %q before %q", list[i-1].Name, ci.Name)
		}
	}
}

func TestExpectedChecks(t *testing.T) {
	t.Parallel()
	c, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	tests := []struct {
		platform string
		include  []string
		exclude  []string
	}{
		{
			platform: "github.com",
			include:  []string{"CI-Tests", "Signed-Releases", "Vulnerabilities"},
			exclude:  []string{"SBOM", "Webhooks"},
		},
		{
			platform: "gitlab.com",
			include:  []string{"CI-Tests", "Vulnerabilities"},
			exclude:  []string{"SBOM", "Signed-Releases", "Webhooks"},
		},
		{
			platform: "example.com",
			include:  []string{"Signed-Releases", "Vulnerabilities"},
			exclude:  []string{"SBOM", "Webhooks"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.platform, func(t *testing.T) {
			t.Parallel()
			got := c.ExpectedChecks(tt.platform)
			if !slices.IsSorted(got) {
				t.Errorf("ExpectedChecks(%q) = %v, want sorted", tt.platform, got)
			}
			for _, name := range tt.include {
				if !slices.Contains(got, name) {
					t.Errorf("ExpectedChecks(%q) = %v, want it to include %q", tt.platform, got, name)
				}
			}
			for _, name := range tt.exclude {
				if slices.Contains(got, name) {
					t.Errorf("ExpectedChecks(%q) = %v, want it to exclude %q", tt.platform, got, name)
				}
			}
		})
	}
}

func TestDocumentationURLCommit(t *testing.T) {
	t.Parallel()
	const commit = "61e84743356275f6ffca23b1b2f2cb945e19eb96"
	tests := []struct {
		commit string
		want   string
	}{
		{commit: commit, want: "/blob/" + commit + "/docs/checks.md"},
		{commit: "unknown", want: "/blob/main/docs/checks.md"},
		{commit: "", want: "/blob/main/docs/checks.md"},
	}
	for _, tt := range tests {
		t.Run(tt.commit, func(t *testing.T) {
			t.Parallel()
			c, err := newCatalog("v0.0.0", tt.commit)
			if err != nil {
				t.Fatalf("newCatalog: %v", err)
			}
			for _, ci := range c.List() {
				if !strings.Contains(ci.DocumentationURL, tt.want) {
					t.Errorf("%s DocumentationURL = %q, want it to contain %q", ci.Name, ci.DocumentationURL, tt.want)
				}
			}
		})
	}
}

func TestCatalogExplain(t *testing.T) {
	t.Parallel()
	c, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Known check, case-insensitive.
	d, err := c.Explain("branch-protection")
	if err != nil {
		t.Fatalf("Explain(branch-protection): %v", err)
	}
	if d.Name == "" || d.Description == "" {
		t.Errorf("expected populated detail, got name=%q descLen=%d", d.Name, len(d.Description))
	}

	// Experimental flag.
	if wh, err := c.Explain("Webhooks"); err == nil && !wh.Experimental {
		t.Errorf("Webhooks should be marked experimental")
	}

	// Unknown check.
	if _, err := c.Explain("Not-A-Real-Check"); err == nil {
		t.Error("expected error for unknown check")
	}
}

// TestExperimentalChecksMatch compares the catalog against the checks package:
// experimental checks are those in
// GetAllWithExperimental() but not in GetAll() when SCORECARD_EXPERIMENTAL is
// unset.
func TestExperimentalChecksMatch(t *testing.T) {
	t.Setenv("SCORECARD_EXPERIMENTAL", "") // restores the original value on cleanup
	if err := os.Unsetenv("SCORECARD_EXPERIMENTAL"); err != nil {
		t.Fatalf("unsetting SCORECARD_EXPERIMENTAL: %v", err)
	}

	want := map[string]bool{}
	defaults := checks.GetAll()
	for name := range checks.GetAllWithExperimental() {
		if _, ok := defaults[name]; !ok {
			want[name] = true
		}
	}
	c, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	got := map[string]bool{}
	for _, check := range c.List() {
		if check.Experimental {
			got[check.Name] = true
		}
	}
	if !maps.Equal(got, want) {
		t.Errorf("catalog experimental checks = %v, want %v (from checks/all_checks.go)", got, want)
	}
}
