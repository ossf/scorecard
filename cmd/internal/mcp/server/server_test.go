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

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ossf/scorecard/v5/cmd/internal/mcp/model"
	"github.com/ossf/scorecard/v5/cmd/internal/mcp/provider"
)

// fakeProvider implements provider.Provider with canned data (no network).
type fakeProvider struct {
	result        *model.Result
	err           error
	omittedChecks []string
}

func (f *fakeProvider) GetResult(_ context.Context, ref model.RepoRef, _ string) (*model.Result, error) {
	if f.err != nil {
		return nil, f.err
	}
	r := *f.result
	r.Repo = ref
	return &r, nil
}

func (f *fakeProvider) Capabilities() provider.Capabilities {
	return provider.Capabilities{Source: model.SourceCachedREST, OmittedChecks: f.omittedChecks}
}

func fakeResult() *model.Result {
	return &model.Result{
		Date:        "2024-01-02T03:04:05Z",
		Commit:      "64febf8c5229f0a5f0a6d2a0f0d4b6a7c8d9e0f1",
		Scorecard:   model.ScorecardInfo{Version: "v5.5.0"},
		Source:      model.SourceCachedREST,
		Score:       8.7,
		Checks:      []model.Check{{Name: "Code-Review", Score: 10, Reason: "ok"}, {Name: "Fuzzing", Score: -1, Reason: "inconclusive"}},
		Caveats:     []string{"cached"},
		Attribution: &model.Attribution{DataLicense: "CDLA-Permissive-2.0", SourceURL: "http://x"},
	}
}

func newTestSession(t *testing.T, p provider.Provider) (*mcp.ClientSession, func()) {
	t.Helper()
	srv, err := New(p)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ct, st := mcp.NewInMemoryTransports()
	ctx := context.Background()
	ss, err := srv.Connect(ctx, st, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	return cs, func() { _ = cs.Close(); _ = ss.Close() }
}

func decodeStructured[T any](t *testing.T, res *mcp.CallToolResult) T {
	t.Helper()
	var out T
	b, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatalf("marshal structured content: %v", err)
	}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal structured content: %v", err)
	}
	return out
}

func TestListToolsAnnotations(t *testing.T) {
	t.Parallel()
	cs, done := newTestSession(t, &fakeProvider{result: fakeResult()})
	defer done()

	res, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	names := map[string]bool{}
	for _, tl := range res.Tools {
		names[tl.Name] = true
		if tl.Annotations == nil || !tl.Annotations.ReadOnlyHint {
			t.Errorf("tool %q missing readOnlyHint", tl.Name)
		}
		if len(tl.Name) > 64 {
			t.Errorf("tool name %q exceeds 64 chars", tl.Name)
		}
	}
	for _, want := range []string{"get_repo_score", "get_check_result", "compare_repos", "list_checks", "explain_check"} {
		if !names[want] {
			t.Errorf("missing tool %q", want)
		}
	}
}

func TestGetRepoScoreTool(t *testing.T) {
	t.Parallel()
	cs, done := newTestSession(t, &fakeProvider{result: fakeResult()})
	defer done()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_repo_score",
		Arguments: map[string]any{"repo": "ossf/scorecard"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected tool error: %v", res.Content)
	}
	out := decodeStructured[repoScoreOutput](t, res)
	if out.Score != 8.7 {
		t.Errorf("score = %v, want 8.7", out.Score)
	}
	if out.Repo.Org != "ossf" || out.Repo.Name != "scorecard" {
		t.Errorf("repo = %+v", out.Repo)
	}
	if len(out.Checks) != 2 {
		t.Errorf("checks = %d, want 2", len(out.Checks))
	}
	if out.Source != model.SourceCachedREST {
		t.Errorf("source = %q", out.Source)
	}
	if len(out.Caveats) == 0 || out.Attribution == nil {
		t.Error("expected caveats and attribution on output")
	}
}

func TestGetRepoScoreInvalidCommit(t *testing.T) {
	t.Parallel()
	cs, done := newTestSession(t, &fakeProvider{result: fakeResult()})
	defer done()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_repo_score",
		Arguments: map[string]any{"repo": "ossf/scorecard", "commit": "not-a-sha"},
	})
	if err != nil {
		t.Fatalf("CallTool transport error: %v", err)
	}
	if !res.IsError {
		t.Error("expected a tool error for an invalid commit")
	}
}

// toolErrorText returns the text of a failed tool call's first content item.
func toolErrorText(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	if !res.IsError {
		t.Fatal("expected a tool error")
	}
	if len(res.Content) == 0 {
		t.Fatal("expected error content")
	}
	tc, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected text content, got %T", res.Content[0])
	}
	return tc.Text
}

func TestGetCheckResultUnknown(t *testing.T) {
	t.Parallel()
	cs, done := newTestSession(t, &fakeProvider{result: fakeResult()})
	defer done()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_check_result",
		Arguments: map[string]any{"repo": "ossf/scorecard", "check": "No-Such-Check"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	text := toolErrorText(t, res)
	if !strings.Contains(text, "not a Scorecard check") {
		t.Errorf("error = %q, want mention of an unknown check", text)
	}
}

func TestGetCheckResultExperimental(t *testing.T) {
	t.Parallel()
	cs, done := newTestSession(t, &fakeProvider{result: fakeResult()})
	defer done()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_check_result",
		Arguments: map[string]any{"repo": "ossf/scorecard", "check": "SBOM"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	text := toolErrorText(t, res)
	if !strings.Contains(text, "experimental") {
		t.Errorf("error = %q, want mention that SBOM is experimental", text)
	}
}

func TestGetCheckResultOmitted(t *testing.T) {
	t.Parallel()
	cs, done := newTestSession(t, &fakeProvider{
		result:        fakeResult(),
		omittedChecks: []string{"CI-Tests", "Contributors", "Dependency-Update-Tool"},
	})
	defer done()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_check_result",
		Arguments: map[string]any{"repo": "ossf/scorecard", "check": "CI-Tests"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	text := toolErrorText(t, res)
	if !strings.Contains(text, "omits") {
		t.Errorf("error = %q, want mention that this provider omits the check", text)
	}
}

func TestGetCheckResultAbsentGeneric(t *testing.T) {
	t.Parallel()
	cs, done := newTestSession(t, &fakeProvider{result: fakeResult()})
	defer done()

	// License is a real, non-experimental check, not in fakeResult()'s
	// checks and not in this fake provider's (empty) OmittedChecks.
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_check_result",
		Arguments: map[string]any{"repo": "ossf/scorecard", "check": "License"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	text := toolErrorText(t, res)
	if !strings.Contains(text, "no result for") {
		t.Errorf("error = %q, want a generic no-result message", text)
	}
}

func TestCompareReposInlineError(t *testing.T) {
	t.Parallel()
	cs, done := newTestSession(t, &fakeProvider{result: fakeResult()})
	defer done()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "compare_repos",
		Arguments: map[string]any{"repos": []string{"ossf/scorecard", "bad-ref"}},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if res.IsError {
		t.Fatalf("compare_repos should not fail wholesale: %v", res.Content)
	}
	out := decodeStructured[compareReposOutput](t, res)
	if len(out.Results) != 2 {
		t.Fatalf("results = %d, want 2", len(out.Results))
	}
	if out.Results[0].Error != "" || out.Results[0].Score == nil {
		t.Errorf("first entry should succeed: %+v", out.Results[0])
	}
	if out.Results[1].Error == "" {
		t.Error("second entry should carry an inline parse error")
	}
}

// resultWithDetails returns a result whose single check carries n detail lines.
func resultWithDetails(n int) *model.Result {
	details := make([]string, n)
	for i := range details {
		details[i] = fmt.Sprintf("detail line %d", i)
	}
	return &model.Result{
		Date:      "2024-01-02T03:04:05Z",
		Scorecard: model.ScorecardInfo{Version: "v5.5.0"},
		Source:    model.SourceCachedREST,
		Score:     7.0,
		Checks:    []model.Check{{Name: "Pinned-Dependencies", Score: 5, Reason: "some pinned", Details: details}},
	}
}

func getCheckDetail(t *testing.T, p provider.Provider, args map[string]any) checkDetail {
	t.Helper()
	cs, done := newTestSession(t, p)
	defer done()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "get_check_result", Arguments: args})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected tool error: %v", res.Content)
	}
	return decodeStructured[getCheckResultOutput](t, res).Check
}

func TestGetCheckResultDetailsTruncatedByDefault(t *testing.T) {
	t.Parallel()
	c := getCheckDetail(t, &fakeProvider{result: resultWithDetails(60)},
		map[string]any{"repo": "ossf/scorecard", "check": "Pinned-Dependencies"})
	if c.DetailsTotal != 60 {
		t.Errorf("details_total = %d, want 60", c.DetailsTotal)
	}
	if !c.DetailsTruncated {
		t.Error("details_truncated = false, want true")
	}
	if len(c.Details) != defaultMaxDetails {
		t.Errorf("details len = %d, want %d (default)", len(c.Details), defaultMaxDetails)
	}
}

func TestGetCheckResultMaxDetailsRespected(t *testing.T) {
	t.Parallel()
	c := getCheckDetail(t, &fakeProvider{result: resultWithDetails(60)},
		map[string]any{"repo": "ossf/scorecard", "check": "Pinned-Dependencies", "max_details": 5})
	if len(c.Details) != 5 || !c.DetailsTruncated || c.DetailsTotal != 60 {
		t.Errorf("details=%d truncated=%v total=%d, want 5/true/60", len(c.Details), c.DetailsTruncated, c.DetailsTotal)
	}
}

func TestGetCheckResultMaxDetailsClampedHigh(t *testing.T) {
	t.Parallel()
	// A value above the cap is clamped to maxMaxDetails; with fewer details than
	// the cap, nothing is truncated.
	c := getCheckDetail(t, &fakeProvider{result: resultWithDetails(60)},
		map[string]any{"repo": "ossf/scorecard", "check": "Pinned-Dependencies", "max_details": 100000})
	if len(c.Details) != 60 || c.DetailsTruncated || c.DetailsTotal != 60 {
		t.Errorf("details=%d truncated=%v total=%d, want 60/false/60", len(c.Details), c.DetailsTruncated, c.DetailsTotal)
	}
}

func TestGetCheckResultNoDetailsNotTruncated(t *testing.T) {
	t.Parallel()
	c := getCheckDetail(t, &fakeProvider{result: resultWithDetails(3)},
		map[string]any{"repo": "ossf/scorecard", "check": "Pinned-Dependencies"})
	if len(c.Details) != 3 || c.DetailsTruncated || c.DetailsTotal != 3 {
		t.Errorf("details=%d truncated=%v total=%d, want 3/false/3", len(c.Details), c.DetailsTruncated, c.DetailsTotal)
	}
}

func TestGetRepoScoreValidCommitAccepted(t *testing.T) {
	t.Parallel()
	cs, done := newTestSession(t, &fakeProvider{result: fakeResult()})
	defer done()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_repo_score",
		Arguments: map[string]any{"repo": "ossf/scorecard", "commit": "64febf8c5229f0a5f0a6d2a0f0d4b6a7c8d9e0f1"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if res.IsError {
		t.Fatalf("valid 40-hex commit should be accepted: %v", res.Content)
	}
}

func TestCompareReposTruncatesNotRejects(t *testing.T) {
	t.Parallel()
	cs, done := newTestSession(t, &fakeProvider{result: fakeResult()})
	defer done()

	repos := make([]string, maxCompareRepos+2)
	for i := range repos {
		repos[i] = fmt.Sprintf("ossf/repo-%d", i)
	}
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "compare_repos",
		Arguments: map[string]any{"repos": repos},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if res.IsError {
		t.Fatalf("more than the max should truncate, not reject: %v", res.Content)
	}
	out := decodeStructured[compareReposOutput](t, res)
	if !out.Truncated {
		t.Error("truncated = false, want true")
	}
	if len(out.Results) != maxCompareRepos {
		t.Errorf("results = %d, want %d", len(out.Results), maxCompareRepos)
	}
	// Order is preserved despite concurrent fetch.
	for i, r := range out.Results {
		if want := fmt.Sprintf("ossf/repo-%d", i); r.Input != want {
			t.Errorf("results[%d].Input = %q, want %q", i, r.Input, want)
		}
	}
}

func TestListChecksTool(t *testing.T) {
	t.Parallel()
	cs, done := newTestSession(t, &fakeProvider{result: fakeResult()})
	defer done()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "list_checks", Arguments: map[string]any{}})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if res.IsError {
		t.Fatalf("list_checks error: %v", res.Content)
	}
	out := decodeStructured[listChecksOutput](t, res)
	if len(out.Checks) < 10 {
		t.Errorf("checks = %d, want many", len(out.Checks))
	}
}

func TestReadCheckResource(t *testing.T) {
	t.Parallel()
	cs, done := newTestSession(t, &fakeProvider{result: fakeResult()})
	defer done()

	rr, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: "scorecard://checks/Fuzzing"})
	if err != nil {
		t.Fatalf("ReadResource: %v", err)
	}
	if len(rr.Contents) == 0 || rr.Contents[0].Text == "" {
		t.Fatal("expected resource content")
	}
	var d map[string]any
	if err := json.Unmarshal([]byte(rr.Contents[0].Text), &d); err != nil {
		t.Fatalf("unmarshal resource: %v", err)
	}
	if d["name"] != "Fuzzing" {
		t.Errorf("resource name = %v, want Fuzzing", d["name"])
	}
}
