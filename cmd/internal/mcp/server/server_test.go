package server

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ossf/scorecard/v5/cmd/internal/mcp/model"
	"github.com/ossf/scorecard/v5/cmd/internal/mcp/provider"
)

// fakeProvider implements provider.Provider with canned data (no network).
type fakeProvider struct {
	result *model.Result
	err    error
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
	return provider.Capabilities{Source: model.SourceCachedREST}
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
	if !res.IsError {
		t.Error("expected a tool error for an unknown check")
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
