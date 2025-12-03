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
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	scalibr "github.com/google/osv-scalibr"
	scalibrfs "github.com/google/osv-scalibr/fs"
	"github.com/google/osv-scalibr/plugin"
	pl "github.com/google/osv-scalibr/plugin/list"
	"github.com/package-url/packageurl-go"
)

// ===== Debug support (opt-in) ================================================

var scalibrDebug = func() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("SCALIBR_DEBUG")))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}()

func SetScalibrDebug(enable bool) { scalibrDebug = enable }

func dbgf(format string, args ...any) {
	if scalibrDebug {
		log.Printf("[scalibr] "+format, args...)
	}
}

// =============================================================================

var (
	ErrLocalDirRequired = errors.New("localDir is required")
	ErrScalibrNoResult  = errors.New("scalibr returned no result")
)

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
	Location  string // left empty; current extractor API doesn't expose file reliably
}

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
		return DepsResponse{}, ErrLocalDirRequired
	}

	// Parse go.mod to identify indirect dependencies
	goIndirectDeps := parseGoModIndirect(localDir)
	log.Printf("[scalibr] GetDeps: parsed %d indirect dependencies from go.mod", len(goIndirectDeps))

	// Load all registered plugins, then filter by capabilities.
	allPlugins := pl.All()
	selected := plugin.FilterByCapabilities(allPlugins, c.capab)
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
		return DepsResponse{}, ErrScalibrNoResult
	}

	out := DepsResponse{Started: res.StartTime, Finished: res.EndTime}
	dbgf("GetDeps: inventory packages reported: %d (pre-filter)", len(res.Inventory.Packages))

	skippedIndirect := 0
	for _, p := range res.Inventory.Packages {
		if p == nil {
			continue
		}

		// Derive ecosystem from PURL type when present.
		var purlStr, eco string
		if u := p.PURL(); u != nil {
			purlStr = u.String()
			eco = normalizePURLType(u.Type)
		}

		// Keep only direct dependencies, when metadata indicates scope.
		if md, ok := p.Metadata.(map[string]string); ok {
			if strings.EqualFold(md["dependency_scope"], "indirect") {
				skippedIndirect++
				dbgf("GetDeps: skip indirect: name=%q version=%q", p.Name, p.Version)
				continue
			}
		}

		// For Go dependencies, also check if marked indirect in go.mod
		if goIndirectDeps[p.Name] {
			skippedIndirect++
			dbgf("GetDeps: skip Go indirect (from go.mod): name=%q version=%q", p.Name, p.Version)
			continue
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
//
//nolint:goconst // String literals are clearer than constants for this mapping function
func normalizePURLType(t string) string {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "golang", "go", "gomod":
		return "golang"
	case "npm", "node", "packagejson":
		return "npm"
	case "pypi", "python", "pyproject", "requirements":
		return "pypi"
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

// parseGoModIndirect recursively parses ALL go.mod files in the given directory
// and returns a set of package names that are marked as indirect dependencies.
// This is necessary because osv-scalibr scans ALL go.mod files in the tree
// (including subdirectories like "tools/"), not just the root one.
func parseGoModIndirect(rootDir string) map[string]bool {
	indirectDeps := make(map[string]bool)

	log.Printf("[scalibr] parseGoModIndirect: searching for go.mod files recursively in: %s", rootDir)

	// First, check if we need to navigate into a tarball-extracted subdirectory
	// (GitHub tarballs extract to "repo-{sha}/" subdirectory)
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		log.Printf("[scalibr] parseGoModIndirect: failed to read directory: %v", err)
		return indirectDeps
	}

	// If there's exactly one subdirectory and no go.mod at root, it's likely a tarball extraction
	_, errRootGoMod := os.Stat(filepath.Join(rootDir, "go.mod"))
	if errRootGoMod != nil && len(entries) == 1 && entries[0].IsDir() {
		// Navigate into the single subdirectory (tarball structure)
		rootDir = filepath.Join(rootDir, entries[0].Name())
		log.Printf("[scalibr] parseGoModIndirect: detected tarball structure, using: %s", rootDir)
	}

	// Now walk the tree and parse ALL go.mod files
	goModCount := 0
	//nolint:nilerr // Intentionally skip unreadable directories and continue walking
	err = filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip directories we can't read
		}

		// Skip hidden directories like .git
		if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}

		// Parse go.mod files
		if !d.IsDir() && d.Name() == "go.mod" {
			goModCount++
			beforeCount := len(indirectDeps)
			if err := parseGoModFileForIndirect(path, indirectDeps); err == nil {
				newIndirect := len(indirectDeps) - beforeCount
				if newIndirect > 0 {
					log.Printf("[scalibr] parseGoModIndirect: found %d indirect deps in %s", newIndirect, path)
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Printf("[scalibr] parseGoModIndirect: error during walk: %v", err)
	}

	log.Printf("[scalibr] parseGoModIndirect: scanned %d go.mod files, found %d total indirect dependencies",
		goModCount, len(indirectDeps))
	return indirectDeps
}

// parseGoModFileForIndirect parses a single go.mod file and populates the indirectDeps map.
func parseGoModFileForIndirect(goModPath string, indirectDeps map[string]bool) error {
	file, err := os.Open(goModPath)
	if err != nil {
		return fmt.Errorf("opening go.mod file %s: %w", goModPath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Look for lines with "// indirect" comment
		if !strings.Contains(line, "// indirect") {
			continue
		}

		// Parse the dependency name from the line
		// Format: "github.com/pkg/name v1.2.3 // indirect"
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			pkgName := parts[0]
			indirectDeps[pkgName] = true
			dbgf("parseGoModIndirect: found indirect dep: %q", pkgName)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanning go.mod file %s: %w", goModPath, err)
	}
	return nil
}
