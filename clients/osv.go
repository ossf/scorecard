// Copyright 2021 OpenSSF Scorecard Authors
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

package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/osv-scanner/v2/pkg/osvscanner"

	sce "github.com/ossf/scorecard/v5/errors"
)

const (
	ecosystemGolang = "golang"
	ecosystemNPM    = "npm"
	ecosystemPyPI   = "pypi"
	ecosystemMaven  = "maven"
	ecosystemGem    = "gem"
)

var (
	errOSVBatchStatus = errors.New("OSV batch request failed")
	errOSVGetStatus   = errors.New("OSV get vuln request failed")
)

var _ VulnerabilitiesClient = osvClient{}

type osvClient struct {
	local bool
}

// ListUnfixedVulnerabilities implements VulnerabilityClient.ListUnfixedVulnerabilities.
func (v osvClient) ListUnfixedVulnerabilities(
	ctx context.Context,
	commit,
	localPath string,
) (_ VulnerabilitiesResponse, err error) {
	osvscanner.SetLogger(slog.DiscardHandler)
	defer func() {
		if r := recover(); r != nil {
			err = sce.CreateInternal(sce.ErrScorecardInternal, fmt.Sprintf("osv-scanner panic: %v", r))
			fmt.Fprintf(os.Stderr, "osv-scanner panic: %v\n%s\n", r, string(debug.Stack()))
		}
	}()
	directoryPaths := []string{}
	if localPath != "" {
		directoryPaths = append(directoryPaths, localPath)
	}
	gitCommits := []string{}
	if commit != "" {
		gitCommits = append(gitCommits, commit)
	}
	res, err := osvscanner.DoScan(osvscanner.ScannerActions{
		DirectoryPaths:    directoryPaths,
		IncludeGitRoot:    false,
		Recursive:         true,
		GitCommits:        gitCommits,
		CompareOffline:    v.local,
		DownloadDatabases: v.local,
		// swap out the transitive requirements scanning for offline extractor
		ExperimentalScannerActions: osvscanner.ExperimentalScannerActions{
			PluginsEnabled:  []string{"python/requirements"},
			PluginsDisabled: []string{"python/requirementsenhanceable"},
		},
	}) // TODO: Do logging?

	response := VulnerabilitiesResponse{}

	// either no vulns found, or no packages detected by osvscanner, which likely means no vulns
	// while there could still be vulns, not detecting any packages shouldn't be a runtime error.
	if err == nil || errors.Is(err, osvscanner.ErrNoPackagesFound) {
		return response, nil
	}

	// If vulnerabilities are found, err will be set to osvscanner.VulnerabilitiesFoundErr
	if errors.Is(err, osvscanner.ErrVulnerabilitiesFound) {
		vulns := res.Flatten()
		for i := range vulns {
			// ignore Go stdlib vulns. The go directive from the go.mod isn't a perfect metric
			// of which version of Go will be used to build a project.
			if vulns[i].Package.Ecosystem == "Go" && vulns[i].Package.Name == "stdlib" {
				continue
			}
			response.Vulnerabilities = append(response.Vulnerabilities, Vulnerability{
				ID:      vulns[i].Vulnerability.ID,
				Aliases: vulns[i].Vulnerability.Aliases,
			})
			// Remove duplicate vulnerability IDs for now as we don't report information
			// on the source of each vulnerability yet, therefore having multiple identical
			// vuln IDs might be confusing.
			response.Vulnerabilities = removeDuplicate(
				response.Vulnerabilities,
				func(key Vulnerability) string { return key.ID },
			)
		}

		return response, nil
	}

	return VulnerabilitiesResponse{}, fmt.Errorf("osvscanner.DoScan: %w", err)
}

// RemoveDuplicate removes duplicate entries from a slice.
func removeDuplicate[T any, K comparable](sliceList []T, keyExtract func(T) K) []T {
	allKeys := make(map[K]bool)
	list := []T{}
	for _, item := range sliceList {
		key := keyExtract(item)
		if _, value := allKeys[key]; !value {
			allKeys[key] = true
			list = append(list, item)
		}
	}
	return list
}

// OSVAPIClient is a small interface for talking to the public OSV API.
// It is intentionally separate from the existing osv-scanner-based client above.
type OSVAPIClient interface {
	// QueryBatch POSTs /v1/querybatch and returns, per query, the list of vuln IDs.
	QueryBatch(ctx context.Context, queries []OSVQuery) ([][]string, error)
	// GetVuln GETs /v1/vulns/{id} and returns minimal details (timestamps, aliases, etc.).
	GetVuln(ctx context.Context, id string) (*OSVVuln, error)
}

// NewOSVClient returns an OSVAPIClient backed by a default http.Client.
// (Name chosen to be convenient for callers; it does not conflict with the osvClient type.)
func NewOSVClient() OSVAPIClient {
	osvDebugEnabled := false
	envVal := strings.ToLower(strings.TrimSpace(os.Getenv("OSV_DEBUG")))
	if envVal == "1" || envVal == "true" || envVal == "yes" || envVal == "on" {
		osvDebugEnabled = true
	}
	return &osvHTTPClient{
		http:  &http.Client{Timeout: defaultHTTPTimeout},
		base:  osvBaseURL,
		debug: osvDebugEnabled,
	}
}

// If you prefer to inject a shared RoundTripper:
// func NewOSVClientWithHTTPClient(h *http.Client) OSVAPIClient { return &osvHTTPClient{http: h, base: osvBaseURL} }

const (
	osvBaseURL         = "https://api.osv.dev"
	osvQueryBatchPath  = "/v1/querybatch"
	osvGetVulnTemplate = "/v1/vulns/%s"
	defaultHTTPTimeout = 20 * time.Second
)

type osvHTTPClient struct {
	http  *http.Client
	base  string
	debug bool
}

func (c *osvHTTPClient) logf(format string, args ...any) {
	if c.debug {
		fmt.Fprintf(os.Stderr, "[osv] "+format+"\n", args...)
	}
}

// OSVPackage identifies a package either by ecosystem+name or by PURL.
type OSVPackage struct {
	Name      string `json:"name,omitempty"`      // e.g., "jinja2"
	Ecosystem string `json:"ecosystem,omitempty"` // e.g., "PyPI", "npm", "Maven", "Go"
	PURL      string `json:"purl,omitempty"`      // optional, e.g., "pkg:pypi/jinja2@2.4.1"
}

// OSVQuery describes a single package@version lookup used by /v1/querybatch.
type OSVQuery struct {
	Package OSVPackage `json:"package"`
	Version string     `json:"version,omitempty"`
	Commit  string     `json:"commit,omitempty"`
}

// osvBatchItem is the normalized per-query shape OSV expects in /v1/querybatch.
type osvBatchItem struct {
	Package   *OSVPackage `json:"package,omitempty"`
	Version   string      `json:"version,omitempty"`
	Commit    string      `json:"commit,omitempty"`
	PageToken string      `json:"page_token,omitempty"`
}

type osvBatchReq struct {
	Queries []osvBatchItem `json:"queries"`
}

type osvBatchResp struct {
	Results []struct {
		NextPageToken string `json:"next_page_token,omitempty"`
		Vulns         []struct {
			ID string `json:"id"`
		} `json:"vulns"`
	} `json:"results"`
}

// OSVVuln is a reduced view of the OSV vuln record with the fields we need for
// "known-at-release" filtering and basic reporting. Extend as needed.
type OSVVuln struct {
	ID        string     `json:"id"`
	Published *time.Time `json:"published,omitempty"`
	Modified  *time.Time `json:"modified,omitempty"`
	Aliases   []string   `json:"aliases,omitempty"`
	Summary   string     `json:"summary,omitempty"`
	Details   string     `json:"details,omitempty"`
	Severity  []struct {
		Type  string `json:"type,omitempty"`
		Score string `json:"score,omitempty"`
	} `json:"severity,omitempty"`
}

// QueryBatch implements POST /v1/querybatch.
// It normalizes queries to OSV’s strict rules:
//   - If Package.PURL is present -> USE ONLY PURL (omit top-level "version").
//   - Else -> USE (ecosystem, name, version). For Go, ensure 'v' prefix on version.
//   - Normalize ecosystem names (e.g., "golang" -> "Go") when not using PURL.
//
// It preserves input order: the N-th returned slice corresponds to the N-th query.
func (c *osvHTTPClient) QueryBatch(ctx context.Context, queries []OSVQuery) ([][]string, error) {
	items := make([]osvBatchItem, 0, len(queries))

	for _, q := range queries {
		var it osvBatchItem

		// Commit query (rare)
		if strings.TrimSpace(q.Commit) != "" {
			it.Commit = strings.TrimSpace(q.Commit)
			items = append(items, it)
			continue
		}

		// Prefer PURL-only when present (omit top-level version).
		if purl := strings.TrimSpace(q.Package.PURL); purl != "" {
			it.Package = &OSVPackage{PURL: purl}
			items = append(items, it)
			continue
		}

		// Fallback: ecosystem+name+version
		pkg := OSVPackage{
			Name:      strings.TrimSpace(q.Package.Name),
			Ecosystem: strings.TrimSpace(q.Package.Ecosystem),
		}
		ver := strings.TrimSpace(q.Version)

		// Normalize ecosystem names if the caller used alternate forms.
		// For OSV, Go must be "Go".
		switch strings.ToLower(pkg.Ecosystem) {
		case ecosystemGolang, "go", "gomod":
			pkg.Ecosystem = "Go"
		case ecosystemPyPI, "python":
			pkg.Ecosystem = "PyPI"
		case ecosystemMaven, "pom", "gradle":
			pkg.Ecosystem = "Maven"
		case "rubygems", ecosystemGem:
			pkg.Ecosystem = "RubyGems"
			// npm, crates.io, nuget already match OSV values typically.
		}

		// Go: ensure v-prefix when not using PURL
		if pkg.Ecosystem == "Go" {
			if ver != "" && !strings.HasPrefix(ver, "v") && !strings.HasPrefix(ver, "V") {
				ver = "v" + ver
			}
		}

		it.Package = &pkg
		it.Version = ver
		items = append(items, it)
	}

	reqBody, err := json.Marshal(osvBatchReq{Queries: items})
	if err != nil {
		return nil, fmt.Errorf("marshal osv batch: %w", err)
	}
	c.logf("POST %s%s body=%s", c.base, osvQueryBatchPath, string(reqBody))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+osvQueryBatchPath, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("new osv batch request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("osv batch do: %w", err)
	}
	defer resp.Body.Close()

	raw, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		c.logf("RESP %d (failed to read body: %w)", resp.StatusCode, readErr)
	} else {
		c.logf("RESP %d %s", resp.StatusCode, string(raw))
	}

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("%w: %s", errOSVBatchStatus, resp.Status)
	}

	var out osvBatchResp
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode osv batch: %w", err)
	}

	ids := make([][]string, 0, len(out.Results))
	for _, r := range out.Results {
		row := make([]string, 0, len(r.Vulns))
		for _, v := range r.Vulns {
			row = append(row, v.ID)
		}
		ids = append(ids, row)
	}
	return ids, nil
}

// GetVuln implements GET /v1/vulns/{id} for timestamp/details lookup.
func (c *osvHTTPClient) GetVuln(ctx context.Context, id string) (*OSVVuln, error) {
	url := fmt.Sprintf(c.base+osvGetVulnTemplate, strings.TrimSpace(id))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("new osv vuln request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("osv get vuln do: %w", err)
	}
	defer resp.Body.Close()

	raw, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		c.logf("GET %s -> %d (failed to read body: %w)", url, resp.StatusCode, readErr)
	} else {
		c.logf("GET %s -> %d %s", url, resp.StatusCode, string(raw))
	}

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("%w: %s", errOSVGetStatus, resp.Status)
	}

	var v OSVVuln
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, fmt.Errorf("decode osv vuln: %w", err)
	}
	return &v, nil
}
