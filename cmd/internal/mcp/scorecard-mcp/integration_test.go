package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const sampleRESTResult = `{
  "date":"2024-01-02T03:04:05Z",
  "repo":{"name":"github.com/ossf/scorecard","commit":"64febf8c5229f0a5f0a6d2a0f0d4b6a7c8d9e0f1"},
  "scorecard":{"version":"v5.5.0","commit":"abc123"},
  "score":8.7,
  "checks":[{"name":"Code-Review","score":10,"reason":"all good"}]
}`

// TestStdioIntegration builds the real scorecard-mcp binary and drives it
// over a real stdio subprocess (mcp.CommandTransport), the same way a host
// like Claude Desktop or VS Code connects. It also exercises the -base-url
// flag by pointing the binary at a local httptest.Server instead of the
// public API, so the whole run stays hermetic. This guards the actual
// cmd/scorecard-mcp entrypoint wiring, which the in-memory-transport tests
// in internal/server can't reach.
func TestStdioIntegration(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sampleRESTResult))
	}))
	defer ts.Close()

	ctx := context.Background()

	binPath := filepath.Join(t.TempDir(), "scorecard-mcp-test")
	build := exec.CommandContext(ctx, "go", "build", "-o", binPath, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("building scorecard-mcp: %v\n%s", err, out)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	cs, err := client.Connect(ctx, &mcp.CommandTransport{
		Command: exec.CommandContext(ctx, binPath, "-base-url", ts.URL),
	}, nil)
	if err != nil {
		t.Fatalf("connect over stdio: %v", err)
	}
	defer func() { _ = cs.Close() }()

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "get_repo_score",
		Arguments: map[string]any{"repo": "ossf/scorecard"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected tool error: %v", res.Content)
	}

	sc, ok := res.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("structured content = %T, want map[string]any", res.StructuredContent)
	}
	if score, _ := sc["score"].(float64); score != 8.7 {
		t.Errorf("score = %v, want 8.7 (from the -base-url fake server)", sc["score"])
	}
	if src, _ := sc["source"].(string); src != "cached-rest" {
		t.Errorf("source = %q, want cached-rest", src)
	}
}
