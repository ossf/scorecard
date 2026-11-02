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
	"net/url"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ossf/scorecard/v5/cmd/internal/mcp/catalog"
)

const (
	checksResourceURI     = "scorecard://checks"
	checkResourceTemplate = "scorecard://checks/{name}"
	checkResourcePrefix   = "scorecard://checks/"
	mimeJSON              = "application/json"
)

// --- list_checks ---

type listChecksInput struct{}

type listChecksOutput struct {
	ScorecardVersion string              `json:"scorecard_version,omitempty"`
	Checks           []catalog.CheckInfo `json:"checks"`
}

// --- explain_check ---

type explainCheckInput struct {
	Check string `json:"check" jsonschema:"check name, e.g. Branch-Protection (case-insensitive)"`
}

type explainCheckOutput struct {
	ScorecardVersion string              `json:"scorecard_version,omitempty"`
	Check            catalog.CheckDetail `json:"check"`
}

// registerCatalogTools registers the check-catalog capability tools.
func registerCatalogTools(s *mcp.Server, c *catalog.Catalog) {
	mcp.AddTool(s, &mcp.Tool{
		Name: "list_checks",
		Description: "List the available OpenSSF Scorecard checks with, for each, its name, short " +
			"description, risk level, tags, supported platforms, documentation URL, and whether it " +
			"is experimental. Sourced offline from the Scorecard check documentation; does not " +
			"contact the network. Use explain_check for a single check's full methodology and " +
			"remediation.",
		Annotations: readOnlyAnnotations("List Scorecard checks", false),
	}, func(
		_ context.Context,
		_ *mcp.CallToolRequest,
		_ listChecksInput,
	) (*mcp.CallToolResult, listChecksOutput, error) {
		return nil, listChecksOutput{ScorecardVersion: c.ScorecardVersion(), Checks: c.List()}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name: "explain_check",
		Description: "Explain a single OpenSSF Scorecard check: its description, risk, tags, " +
			"remediation guidance, and documentation URL. Use list_checks to discover valid check " +
			"names. Sourced offline from the Scorecard check documentation; does not contact the " +
			"network. Use get_check_result to get a check's score for a specific repository.",
		Annotations: readOnlyAnnotations("Explain a Scorecard check", false),
	}, func(
		_ context.Context,
		_ *mcp.CallToolRequest,
		in explainCheckInput,
	) (*mcp.CallToolResult, explainCheckOutput, error) {
		if in.Check == "" {
			return nil, explainCheckOutput{}, errCheckRequired
		}
		detail, err := c.Explain(in.Check)
		if err != nil {
			return nil, explainCheckOutput{}, fmt.Errorf("explaining check: %w", err)
		}
		return nil, explainCheckOutput{ScorecardVersion: c.ScorecardVersion(), Check: detail}, nil
	})
}

// registerCatalogResources exposes the check documentation as read-only
// resources: an index and a per-check template.
func registerCatalogResources(s *mcp.Server, c *catalog.Catalog) {
	s.AddResource(&mcp.Resource{
		URI:         checksResourceURI,
		Name:        "scorecard-checks",
		Title:       "OpenSSF Scorecard checks",
		Description: "Index of all OpenSSF Scorecard checks with summary documentation.",
		MIMEType:    mimeJSON,
	}, func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		payload := listChecksOutput{ScorecardVersion: c.ScorecardVersion(), Checks: c.List()}
		return jsonResource(req.Params.URI, payload)
	})

	s.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: checkResourceTemplate,
		Name:        "scorecard-check",
		Title:       "OpenSSF Scorecard check",
		Description: "Full documentation for a single OpenSSF Scorecard check, addressed by name.",
		MIMEType:    mimeJSON,
	}, func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		uri := req.Params.URI
		name := strings.TrimPrefix(uri, checkResourcePrefix)
		if name == uri || name == "" {
			return nil, mcp.ResourceNotFoundError(uri)
		}
		if decoded, err := url.PathUnescape(name); err == nil {
			name = decoded
		}
		detail, err := c.Explain(name)
		if err != nil {
			return nil, mcp.ResourceNotFoundError(uri)
		}
		return jsonResource(uri, detail)
	})
}

func jsonResource(uri string, v any) (*mcp.ReadResourceResult, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling resource %s: %w", uri, err)
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      uri,
			MIMEType: mimeJSON,
			Text:     string(b),
		}},
	}, nil
}
