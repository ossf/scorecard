package catalog

import (
	"os"
	"regexp"
	"testing"
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

// TestExperimentalChecksVersionPin fails on a scorecard dependency bump, as a
// reminder to manually re-verify the hand-maintained experimentalChecks map
// (catalog.go) against checks/all_checks.go's getAll(). A test that diffs
// checks.GetAllWithExperimental() against checks.GetAll() directly would
// catch drift automatically, but importing the checks package pulls in
// Scorecard's full client/raw-check dependency graph (git, GitHub/GitLab/
// Azure clients, cloud storage) just for a test — not worth that cost for a
// two-entry, rarely-changing list.
//
// This reads the pin straight from go.mod rather than using
// scorecardModuleVersion() / debug.ReadBuildInfo(): build info dependency
// lists are only populated for real `go build` binaries, not `go test`
// binaries, so that function reliably returns "" under `go test`.
func TestExperimentalChecksVersionPin(t *testing.T) {
	t.Parallel()
	const verifiedVersion = "v5.5.0"
	if v := scorecardGoModVersion(t); v != verifiedVersion {
		t.Fatalf("scorecard dependency is %s, last verified against %s; manually check "+
			"experimentalChecks in catalog.go against checks/all_checks.go's getAll(), "+
			"then update verifiedVersion here", v, verifiedVersion)
	}
}

// scorecardGoModVersion reads the github.com/ossf/scorecard/v5 requirement
// straight out of the repo's go.mod.
func scorecardGoModVersion(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("../../go.mod")
	if err != nil {
		t.Fatalf("reading go.mod: %v", err)
	}
	m := regexp.MustCompile(`(?m)^\tgithub\.com/ossf/scorecard/v5 (\S+)`).FindSubmatch(data)
	if m == nil {
		t.Fatal("github.com/ossf/scorecard/v5 requirement not found in go.mod")
	}
	return string(m[1])
}
