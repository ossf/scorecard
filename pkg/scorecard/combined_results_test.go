package scorecard

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ossf/scorecard/v5/checker"
	docChecks "github.com/ossf/scorecard/v5/docs/checks"
	"github.com/ossf/scorecard/v5/options"
)

// fakeDoc implements minimal docChecks.Doc for tests.
type fakeDoc struct{}

func (f *fakeDoc) GetCheck(name string) (docChecks.CheckDoc, error) {
	// Return a dummy check where GetRisk returns "Medium".
	return &fakeCheck{}, nil
}

func (f *fakeDoc) CheckExists(name string) bool {
	return true
}

func (f *fakeDoc) GetChecks() []docChecks.CheckDoc {
	return []docChecks.CheckDoc{&fakeCheck{}}
}

type fakeCheck struct{}

func (f *fakeCheck) GetName() string                             { return "fake" }
func (f *fakeCheck) GetRisk() string                             { return "Medium" }
func (f *fakeCheck) GetShort() string                            { return "" }
func (f *fakeCheck) GetDescription() string                      { return "" }
func (f *fakeCheck) GetRemediation() []string                    { return nil }
func (f *fakeCheck) GetTags() []string                           { return nil }
func (f *fakeCheck) GetSupportedRepoTypes() []string             { return nil }
func (f *fakeCheck) GetDocumentationURL(commitish string) string { return "http://example" }

func TestFormatCombinedResultsAllAggregatedScore(t *testing.T) {
	t.Parallel()
	// Build two results with simple checks.
	r1 := &Result{Repo: RepoInfo{Name: "owner/repo1"}}
	r1.Checks = []checker.CheckResult{{Name: "c1", Score: 10}, {Name: "c2", Score: 5}}

	r2 := &Result{Repo: RepoInfo{Name: "owner/repo2"}}
	r2.Checks = []checker.CheckResult{{Name: "c1", Score: 7}, {Name: "c2", Score: 7}}

	var buf bytes.Buffer
	err := FormatCombinedResultsAll(&buf, &options.Options{ShowDetails: false, ShowAnnotations: false, LogLevel: "info"}, []*Result{r1, r2}, &fakeDoc{}, nil)
	if err != nil {
		t.Fatalf("FormatCombinedResultsAll failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "AGGREGATED SCORE") {
		t.Fatalf("expected AGGREGATED SCORE header, got: %s", out)
	}
	// Expect repo names present.
	if !strings.Contains(out, "owner/repo1") || !strings.Contains(out, "owner/repo2") {
		t.Fatalf("expected repo names in output, got: %s", out)
	}
}
