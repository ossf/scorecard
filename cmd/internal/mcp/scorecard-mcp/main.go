// Command scorecard-mcp runs an MCP server that exposes OpenSSF Scorecard
// security-posture data to MCP clients over stdio.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ossf/scorecard/v5/cmd/internal/mcp/provider"
	"github.com/ossf/scorecard/v5/cmd/internal/mcp/server"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("scorecard-mcp: %v", err)
	}
}

func run() error {
	baseURL := flag.String("base-url", os.Getenv("SCORECARD_MCP_BASE_URL"),
		"Base URL of the Scorecard REST API (default: the public api.scorecard.dev)")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	p := provider.NewCachedREST(*baseURL)
	s, err := server.New(p)
	if err != nil {
		return err
	}

	if err := s.Run(ctx, &mcp.StdioTransport{}); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("running server: %w", err)
	}
	return nil
}
