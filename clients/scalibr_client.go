// Copyright 2025 OpenSSF Scorecard Authors
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
	"context"
	"errors"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	scalibr "github.com/google/osv-scalibr"
	"github.com/google/osv-scalibr/extractor/filesystem/language/golang/gomod"
	"github.com/google/osv-scalibr/extractor/filesystem/language/javascript/packagejson"
	scalibrfs "github.com/google/osv-scalibr/fs"
	"github.com/google/osv-scalibr/plugin"
	pl "github.com/google/osv-scalibr/plugin/list"
	"github.com/package-url/packageurl-go"
)

var (
	errLocalDirRequired = errors.New("localDir is required")
	errScalibrNoResult  = errors.New("scalibr returned no result")
)

// ===== Debug support (opt-in) ================================================

var (
	scalibrDebugFlag atomic.Bool
	scalibrDebugOnce sync.Once
)

func initScalibrDebug() {
	scalibrDebugOnce.Do(func() {
		v := strings.ToLower(strings.TrimSpace(os.Getenv("SCALIBR_DEBUG")))
		scalibrDebugFlag.Store(v == "1" || v == "true" || v == "yes" || v == "on")
	})
}

func SetScalibrDebug(enable bool) {
	initScalibrDebug()
	scalibrDebugFlag.Store(enable)
}

func dbgf(format string, args ...any) {
	initScalibrDebug()
	if scalibrDebugFlag.Load() {
		log.Printf("[scalibr] "+format, args...)
	}
}

// =============================================================================

type DepsClient interface {
	GetDeps(ctx context.Context, localDir string) (DepsResponse, error)
}

type DepsResponse struct {
	Started  time.Time
	Finished time.Time
	Deps     []Dep
}

type Dep struct {
	Ecosystem string // derived from PURL type when available
	Name      string
	Version   string
	PURL      string
	Location  string // left empty; current extractor API doesn’t expose file reliably
}

// NewDirectDepsClient returns a client that scans with whatever plugins are
// available in this build of osv-scalibr, filtered by capabilities to avoid
// network and transitive resolution. Directness is enforced via metadata.
func NewDirectDepsClient() DepsClient {
	return &scalibrDepsClient{
		capab: &plugin.Capabilities{
			DirectFS: true,
			Network:  plugin.NetworkOffline,
		},
	}
}

type scalibrDepsClient struct {
	capab *plugin.Capabilities
}

func (c *scalibrDepsClient) GetDeps(ctx context.Context, localDir string) (DepsResponse, error) {
	if strings.TrimSpace(localDir) == "" {
		return DepsResponse{}, errLocalDirRequired
	}

	// Load all registered plugins, then filter by capabilities.
	allPlugins := pl.All()
	selected := plugin.FilterByCapabilities(allPlugins, c.capab)

	// Configure extractors to prefer direct dependencies only.
	// Different ecosystems have different mechanisms:
	// - Go (gomod): ExcludeIndirect config option
	// - JavaScript: Use package.json (not package-lock.json) with IncludeDependencies
	// - Python/Rust/etc: Manifest files naturally list direct deps
	for i, p := range selected {
		switch p.Name() {
		case "go/gomod":
			// Configure Go extractor to exclude indirect dependencies
			selected[i] = gomod.NewWithConfig(gomod.Config{ExcludeIndirect: true})
			dbgf("GetDeps: configured gomod extractor to exclude indirect dependencies")

		case "javascript/packagejson":
			// Configure to extract dependencies section (which lists direct deps)
			cfg := packagejson.DefaultConfig()
			cfg.IncludeDependencies = true
			selected[i] = packagejson.New(cfg)
			dbgf("GetDeps: configured packagejson extractor to include dependencies")

		case "javascript/packagelockjson":
			// Exclude package-lock.json since it includes transitive dependencies.
			// We rely on package.json instead for direct dependencies only.
			selected[i] = nil
			dbgf("GetDeps: excluding packagelockjson extractor (includes transitive deps)")
		}
	}

	// Remove nil entries (extractors we explicitly excluded)
	filtered := make([]plugin.Plugin, 0, len(selected))
	for _, p := range selected {
		if p != nil {
			filtered = append(filtered, p)
		}
	}
	selected = filtered

	dbgf("GetDeps: scan root=%q, plugins available=%d, selected=%d (DirectFS, offline)",
		localDir, len(allPlugins), len(selected))

	cfg := &scalibr.ScanConfig{
		Plugins:      selected,
		ScanRoots:    scalibrfs.RealFSScanRoots(localDir),
		UseGitignore: true,
	}
	res := scalibr.New().Scan(ctx, cfg)
	if res == nil || res.Inventory.Packages == nil {
		dbgf("GetDeps: scalibr returned no result or empty inventory")
		return DepsResponse{}, errScalibrNoResult
	}

	out := DepsResponse{Started: res.StartTime, Finished: res.EndTime}
	dbgf("GetDeps: inventory packages reported: %d (pre-filter)", len(res.Inventory.Packages))

	skippedIndirect := 0
	for _, p := range res.Inventory.Packages {
		if p == nil {
			continue
		}

		// Additional fallback: filter indirect deps if metadata indicates scope.
		// The gomod extractor is already configured to exclude indirect deps,
		// but other ecosystem extractors may set this metadata field.
		if md, ok := p.Metadata.(map[string]string); ok {
			if strings.EqualFold(md["dependency_scope"], "indirect") {
				skippedIndirect++
				dbgf("GetDeps: skip indirect: name=%q version=%q", p.Name, p.Version)
				continue
			}
		}

		// Derive ecosystem from PURL type when present.
		var purlStr, eco string
		if u := p.PURL(); u != nil {
			purlStr = u.String()
			eco = normalizePURLType(u.Type)
		}
		name := strings.TrimSpace(p.Name)
		version := strings.TrimSpace(p.Version)

		// Synthesize PURL if missing but we have enough info.
		if purlStr == "" && eco != "" && name != "" && version != "" {
			purlStr = packageurl.NewPackageURL(eco, "", name, version, nil, "").ToString()
		}

		out.Deps = append(out.Deps, Dep{
			Ecosystem: eco,
			Name:      name,
			Version:   version,
			PURL:      purlStr,
			Location:  "",
		})
	}

	// Deterministic order.
	sort.Slice(out.Deps, func(i, j int) bool {
		a, b := out.Deps[i], out.Deps[j]
		if a.Ecosystem != b.Ecosystem {
			return a.Ecosystem < b.Ecosystem
		}
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		if a.Version != b.Version {
			return a.Version < b.Version
		}
		return a.PURL < b.PURL
	})

	dbgf("GetDeps: skipped %d indirect deps; returning %d direct deps", skippedIndirect, len(out.Deps))
	for _, d := range out.Deps {
		dbgf("GetDeps: direct dep: eco=%q name=%q ver=%q purl=%q", d.Ecosystem, d.Name, d.Version, d.PURL)
	}

	return out, nil
}

// normalizePURLType maps PURL types to canonical forms we use elsewhere.
func normalizePURLType(t string) string {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "golang", "go", "gomod":
		return "golang"
	case ecosystemNPM, "node", "packagejson":
		return ecosystemNPM
	case ecosystemPyPI, "python", "pyproject", "requirements":
		return ecosystemPyPI
	case "maven", "pom", "pomxml", "gradle":
		return "maven"
	case "cargo", "cargotoml", "rust", "crates.io":
		return "cargo"
	case "nuget", ".net", "nugetproj":
		return "nuget"
	case "gem", "ruby", "rubygems", "gemfile":
		return "gem"
	default:
		return t
	}
}
