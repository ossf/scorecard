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

package fileparser

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	stdOs "os"
	"path/filepath"

	"deps.dev/util/semver"
	"github.com/google/osv-scalibr/extractor/filesystem"
	"github.com/google/osv-scalibr/extractor/filesystem/language/python/requirements"
	"github.com/google/osv-scalibr/extractor/filesystem/list"
	"github.com/google/osv-scalibr/extractor/filesystem/simplefileapi"
	scalibrfs "github.com/google/osv-scalibr/fs"
)

var (
	packageLock = "package-lock.json"
	packageJSON = "package.json"
	pomXML      = "pom.xml"
	goMod       = "go.mod"
	pyReq       = "requirements.txt"
)

type Dependency struct {
	Name       string
	Version    string
	Comparator string
	Ecosystem  semver.System
}

type NPMPackageLock struct {
	Dependencies map[string]NPMDependency `json:"dependencies"`
	Name         string                   `json:"name"`
	Version      string                   `json:"version"`
}

type NPMPackage struct {
	Dependencies map[string]string `json:"dependencies"`
	Name         string            `json:"name"`
	Version      string            `json:"version"`
}

// NPMDependency represents a dependency read from a package-lock.json or
// package.json file.
// Note that this type is recursive. In npm, dependencies may have nested
// dependencies without limit.
type NPMDependency struct {
	Dependencies map[string]NPMDependency `json:"dependencies"`
	Version      string                   `json:"version"`
	Bundled      bool                     `json:"bundled"`
	Dev          bool                     `json:"dev"`
	Optional     bool                     `json:"optional"`
}

type Version struct {
	Name    string
	Version string
}

func IsLockFile(pathfn string) bool {
	switch filepath.Base(pathfn) {
	case packageJSON, packageLock:
		return true
	case pomXML:
		return true
	case goMod:
		return true
	case pyReq:
		return true
	}
	return false
}

// Compare ompares the version of a dependency against
// either the version or version range of another
// dependency. Compare returns -1 if p1 represents an
// earlier version, +1 a later version, and 0 if they
// are equal.
//
// In case p2 is a range (eg if it starts with "^"),
// Compare returns 0 if p1 is within the range and -1 if
// it represents a version earlier than the range.
func (p1 *Dependency) Compare(p2 *Dependency) bool {
	// npm and python support constraints
	if p1.Ecosystem == semver.NPM ||
		p1.Ecosystem == semver.PyPI {
		if c, err := p1.Ecosystem.ParseConstraint(p2.Version); err == nil {
			return c.Match(p1.Version)
		}
	}
	return p1.Ecosystem.Compare(p1.Version, p2.Version) != -1
}

func getNpmDepsFromPackageLockJSON(pathfn string) ([]*Dependency, error) {
	packages := make([]*Dependency, 0)
	packageJSONContents, err := stdOs.ReadFile(pathfn)
	if err != nil {
		return packages, fmt.Errorf("could not read file: %w", err)
	}
	var pl NPMPackageLock
	err = json.Unmarshal(packageJSONContents, &pl)
	if err != nil {
		return nil, fmt.Errorf("parsing error: %w", err)
	}
	for k, v := range pl.Dependencies {
		p := &Dependency{
			Name:    k,
			Version: v.Version,
		}
		packages = append(packages, p)
	}
	return packages, nil
}

func getPackageJSONDeps(pathfn string) ([]*Dependency, error) {
	packages := make([]*Dependency, 0)
	packageJSONContents, err := stdOs.ReadFile(pathfn)
	if err != nil {
		return packages, fmt.Errorf("could not read file: %w", err)
	}
	var pl NPMPackage
	err = json.Unmarshal(packageJSONContents, &pl)
	if err != nil {
		return nil, fmt.Errorf("parsing error: %w", err)
	}
	for k, v := range pl.Dependencies {
		p := &Dependency{
			Name:    k,
			Version: v,
		}
		packages = append(packages, p)
	}
	return packages, nil
}

// ScanLockFile scans a manifest file and returns the dependencies
// declared in it.
func ScanLockFile(pathfn string) ([]*Dependency, error) {
	// We first manually run our own extractors. osv-scalibr handles
	// a lot of the ecosystems we support, but as of currently, it
	// does not support parsing dependencies from all, so we run our
	// own ones in those cases.
	packages := make([]*Dependency, 0)
	switch filepath.Base(pathfn) {
	case packageJSON:
		return getPackageJSONDeps(pathfn)
	case packageLock:
		return getNpmDepsFromPackageLockJSON(pathfn)
	}

	extractors, err := list.ExtractorsFromName("sourcecode")
	if err != nil {
		return nil, fmt.Errorf("could not get external extractors: %w", err)
	}
	fsReader := stdOs.DirFS(".")
	fsStat, err := fs.Stat(fsReader, pathfn)
	if err != nil {
		return nil, fmt.Errorf("could not fs.Stat file: %w", err)
	}
	for _, e := range extractors {
		// These extractors make http calls to get transitive dependencies.
		// Skip these to avoid fetching transitive dependencies.
		if e.Name() == "python/requirementsnet" || e.Name() == "java/pomxmlnet" {
			continue
		}
		f := simplefileapi.New(pathfn, fsStat)
		// Check if the lockfile matches this extractor
		if !e.FileRequired(f) {
			continue
		}
		fp, err := stdOs.Open(pathfn)
		if err != nil {
			return nil, fmt.Errorf("could not open lockfile: %w", err)
		}
		info, err := fp.Stat()
		if err != nil {
			return nil, fmt.Errorf("could not stat file: %w", err)
		}
		inp := &filesystem.ScanInput{
			FS:     scalibrfs.DirFS("."),
			Path:   pathfn,
			Info:   info,
			Reader: fp,
		}
		inv, err := e.Extract(context.Background(), inp)
		if err != nil {
			return nil, fmt.Errorf("could not extract: %w", err)
		}
		for _, p := range inv.Packages {
			pp := &Dependency{
				Name:    p.Name,
				Version: p.Version,
			}
			if md, ok := p.Metadata.(*requirements.Metadata); ok {
				if md.VersionComparator != "" {
					pp.Comparator = md.VersionComparator
				}
			}
			packages = append(packages, pp)
		}
	}
	return packages, nil
}
