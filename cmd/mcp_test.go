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

package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"sigs.k8s.io/release-utils/version"

	"github.com/ossf/scorecard/v5/options"
)

const sampleRESTResult = `{
  "date":"2024-01-02T03:04:05Z",
  "repo":{"name":"github.com/ossf/scorecard","commit":"64febf8c5229f0a5f0a6d2a0f0d4b6a7c8d9e0f1"},
  "scorecard":{"version":"v5.5.0","commit":"abc123"},
  "score":8.7,
  "checks":[{"name":"Code-Review","score":10,"reason":"all good"}]
}`

// TestMCPStdioIntegration builds the real scorecard binary and drives
// `scorecard mcp`
// over a real stdio subprocess (mcp.CommandTransport), the same way a host
// like Claude Desktop or VS Code connects. It also exercises the --base-url
// flag by pointing the binary at a local httptest.Server instead of the
// public API, so the whole run stays hermetic. This guards the actual
// subcommand wiring, which the in-memory-transport tests in
// cmd/internal/mcp/server can't reach.
func TestMCPStdioIntegration(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(sampleRESTResult)); err != nil {
			t.Errorf("writing response: %v", err)
		}
	}))
	defer ts.Close()

	ctx := context.Background()

	binPath := filepath.Join(t.TempDir(), "scorecard-mcp-test")
	build := exec.CommandContext(ctx, "go", "build", "-o", binPath, "..")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("building scorecard: %v\n%s", err, out)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	cs, err := client.Connect(ctx, &mcp.CommandTransport{
		Command: exec.CommandContext(ctx, binPath, "mcp", "--base-url", ts.URL),
	}, nil)
	if err != nil {
		t.Fatalf("connect over stdio: %v", err)
	}
	defer func() { _ = cs.Close() }()

	out, err := exec.CommandContext(ctx, binPath, "version", "--json").Output()
	if err != nil {
		t.Fatalf("scorecard version: %v", err)
	}
	var info version.Info
	if err := json.Unmarshal(out, &info); err != nil {
		t.Fatalf("parsing scorecard version: %v", err)
	}
	if got := cs.InitializeResult().ServerInfo.Version; got != info.GitVersion {
		t.Errorf("server version = %q, want %q (scorecard version)", got, info.GitVersion)
	}

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
	if score, ok := sc["score"].(float64); !ok || score != 8.7 {
		t.Errorf("score = %v, want 8.7 (from the --base-url fake server)", sc["score"])
	}
	if src, ok := sc["source"].(string); !ok || src != "cached-rest" {
		t.Errorf("source = %v, want cached-rest", sc["source"])
	}
}

func TestMCPRejectsScanFlags(t *testing.T) {
	t.Parallel()
	cmd := New(options.New())
	cmd.SetArgs([]string{"mcp", "--repo=github.com/owner/repo"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err := cmd.ExecuteContext(t.Context())
	if want := "unknown flag: --repo"; err == nil || !strings.Contains(err.Error(), want) {
		t.Errorf("Execute() error = %v, want one containing %q", err, want)
	}
}

func TestRunMCP(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(sampleRESTResult)); err != nil {
			t.Errorf("writing response: %v", err)
		}
	}))
	defer ts.Close()

	ct, st := mcp.NewInMemoryTransports()
	errc := make(chan error, 1)
	go func() { errc <- runMCP(t.Context(), ts.URL, st) }()

	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	cs, err := client.Connect(t.Context(), ct, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	res, err := cs.CallTool(t.Context(), &mcp.CallToolParams{
		Name:      "get_repo_score",
		Arguments: map[string]any{"repo": "ossf/scorecard"},
	})
	if err != nil || res.IsError {
		t.Fatalf("CallTool = %v, %v; want a result", res, err)
	}
	if err := cs.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := <-errc; err != nil {
		t.Errorf("runMCP() error = %v, want nil after the client disconnects", err)
	}
}

var errConnect = errors.New("connect failed")

type failingTransport struct{}

func (failingTransport) Connect(context.Context) (mcp.Connection, error) { return nil, errConnect }

func TestRunMCPTransportError(t *testing.T) {
	t.Parallel()
	err := runMCP(t.Context(), "", failingTransport{})
	if !errors.Is(err, errConnect) || !strings.Contains(err.Error(), "running MCP server") {
		t.Errorf("runMCP() error = %v, want a wrapped %v", err, errConnect)
	}
}
