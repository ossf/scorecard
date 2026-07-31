package catalog

import "testing"

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
