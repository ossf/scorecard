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
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"

	"github.com/ossf/scorecard/v5/cmd/internal/mcp/provider"
	mcpserver "github.com/ossf/scorecard/v5/cmd/internal/mcp/server"
)

const mcpBaseURLEnv = "SCORECARD_MCP_BASE_URL"

func mcpCmd() *cobra.Command {
	var baseURL string
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Serve Scorecard results to MCP clients over stdio",
		Long: `Start a Model Context Protocol (MCP) server on standard input and output.

The server exposes read-only tools and resources for published Scorecard
results and check documentation. It reads results from the Scorecard REST
API and runs no scans. MCP clients start it as a subprocess.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SilenceUsage = true
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			return runMCP(ctx, baseURL, &mcp.StdioTransport{})
		},
	}
	cmd.Flags().StringVar(&baseURL, "base-url", os.Getenv(mcpBaseURLEnv),
		"base URL of the Scorecard REST API (env "+mcpBaseURLEnv+"; default "+provider.DefaultRESTBaseURL+")")
	return cmd
}

// runMCP serves the MCP server on t until the client disconnects or ctx is
// canceled.
func runMCP(ctx context.Context, baseURL string, t mcp.Transport) error {
	s, err := mcpserver.New(provider.NewCachedREST(baseURL))
	if err != nil {
		return fmt.Errorf("creating MCP server: %w", err)
	}
	if err := s.Run(ctx, t); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("running MCP server: %w", err)
	}
	return nil
}
