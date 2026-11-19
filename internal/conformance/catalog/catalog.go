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

// Package catalog lists the catalog releases built into Scorecard and
// resolves the --catalog option to one of them.
package catalog

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/ossf/scorecard/v5/internal/conformance"
	ospsv20260828 "github.com/ossf/scorecard/v5/internal/conformance/ospsbaseline/v20260828"
)

var (
	// ErrUnsupportedCatalog means the catalog ID is not built in.
	ErrUnsupportedCatalog = errors.New("unsupported catalog")
	// ErrUnsupportedVersion means the catalog is built in, but not at the
	// requested release version.
	ErrUnsupportedVersion = errors.New("unsupported catalog version")
)

// releases returns every built-in release, oldest first within each
// catalog.
func releases() []*conformance.Release {
	return []*conformance.Release{
		ospsv20260828.Release(),
	}
}

// IDs returns the sorted IDs of the built-in catalogs.
func IDs() []string {
	var ids []string
	for _, r := range releases() {
		ids = append(ids, r.Catalog)
	}
	slices.Sort(ids)
	return slices.Compact(ids)
}

// Lookup resolves a --catalog value, "<id>" or "<id>@<version>", to a
// built-in release. Without a version it returns the newest built-in
// release of the catalog.
func Lookup(value string) (*conformance.Release, error) {
	id, version, hasVersion := strings.Cut(value, "@")
	var matches []*conformance.Release
	for _, r := range releases() {
		if r.Catalog == id {
			matches = append(matches, r)
		}
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("%w %q; valid catalogs: %s", ErrUnsupportedCatalog, id, strings.Join(IDs(), ", "))
	}
	if !hasVersion {
		return matches[len(matches)-1], nil
	}
	versions := make([]string, 0, len(matches))
	for _, r := range matches {
		if r.Version == version {
			return r, nil
		}
		versions = append(versions, r.Version)
	}
	return nil, fmt.Errorf("%w %q for %s; valid versions: %s",
		ErrUnsupportedVersion, version, id, strings.Join(versions, ", "))
}
