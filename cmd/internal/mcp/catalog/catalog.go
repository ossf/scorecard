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

// Package catalog provides offline access to OpenSSF Scorecard check
// documentation, sourced from the docs embedded in the scorecard module. It
// requires no network access and no scan.
package catalog

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"

	"sigs.k8s.io/release-utils/version"

	"github.com/ossf/scorecard/v5/checks"
	scdocs "github.com/ossf/scorecard/v5/docs/checks"
)

// errUnknownCheck is returned when a check name is not recognized.
var errUnknownCheck = errors.New("unknown check")

// repoTypes maps a repository platform to the repository type the check
// documentation lists in each check's supported repos.
var repoTypes = map[string]string{
	"github.com": "GitHub",
	"gitlab.com": "GitLab",
}

// CheckInfo summarizes a check (used by list_checks and the index resource).
type CheckInfo struct {
	Name               string   `json:"name"`
	Short              string   `json:"short,omitempty"`
	Risk               string   `json:"risk,omitempty" jsonschema:"risk level: Critical, High, Medium, or Low"`
	DocumentationURL   string   `json:"documentation_url,omitempty"`
	Tags               []string `json:"tags,omitempty"`
	SupportedPlatforms []string `json:"supported_platforms,omitempty"`
	Experimental       bool     `json:"experimental" jsonschema:"experimental (off by default in Scorecard)"`
}

// CheckDetail is the full documentation for a check (used by explain_check and
// the per-check resource). It is intentionally flat (no embedding) so its
// inferred JSON schema matches its JSON marshaling.
type CheckDetail struct {
	Name               string   `json:"name"`
	Short              string   `json:"short,omitempty"`
	Risk               string   `json:"risk,omitempty" jsonschema:"risk level: Critical, High, Medium, or Low"`
	DocumentationURL   string   `json:"documentation_url,omitempty"`
	Description        string   `json:"description,omitempty"`
	Tags               []string `json:"tags,omitempty"`
	SupportedPlatforms []string `json:"supported_platforms,omitempty"`
	Remediation        []string `json:"remediation,omitempty"`
	Experimental       bool     `json:"experimental"`
}

// Catalog provides offline access to Scorecard check documentation.
type Catalog struct {
	doc              scdocs.Doc
	commitish        string
	scorecardVersion string
}

// New loads the embedded Scorecard check documentation. Its documentation
// URLs reference the commit the scorecard binary was built from.
func New() (*Catalog, error) {
	info := version.GetVersionInfo()
	return newCatalog(info.GitVersion, info.GitCommit)
}

// newCatalog loads the check documentation with URLs at commit. An empty or
// "unknown" commit falls back to main, as in Scorecard's JSON results.
func newCatalog(scorecardVersion, commit string) (*Catalog, error) {
	doc, err := scdocs.Read()
	if err != nil {
		return nil, fmt.Errorf("reading Scorecard check documentation: %w", err)
	}
	return &Catalog{
		doc:              doc,
		commitish:        commit,
		scorecardVersion: scorecardVersion,
	}, nil
}

// ScorecardVersion returns the version of the scorecard build the catalog is
// part of, as printed by `scorecard version`.
func (c *Catalog) ScorecardVersion() string { return c.scorecardVersion }

// List returns all documented checks, sorted by name.
func (c *Catalog) List() []CheckInfo {
	docs := c.doc.GetChecks()
	out := make([]CheckInfo, 0, len(docs))
	for _, cd := range docs {
		out = append(out, c.toInfo(cd))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Explain returns the full documentation for a single check by name
// (case-insensitive). It returns an error if the check is unknown.
func (c *Catalog) Explain(name string) (CheckDetail, error) {
	canonical, ok := c.resolve(name)
	if !ok {
		return CheckDetail{}, fmt.Errorf("%w %q; use list_checks to see valid check names", errUnknownCheck, name)
	}
	cd, err := c.doc.GetCheck(canonical)
	if err != nil {
		return CheckDetail{}, fmt.Errorf("%w %q; use list_checks to see valid check names", errUnknownCheck, name)
	}
	info := c.toInfo(cd)
	return CheckDetail{
		Name:               info.Name,
		Short:              info.Short,
		Risk:               info.Risk,
		Tags:               info.Tags,
		SupportedPlatforms: info.SupportedPlatforms,
		DocumentationURL:   info.DocumentationURL,
		Experimental:       info.Experimental,
		Description:        cd.GetDescription(),
		Remediation:        cd.GetRemediation(),
	}, nil
}

// ExpectedChecks returns the names of the non-experimental checks that run on
// platform (for example github.com), sorted by name. For a platform the
// documentation does not list, it returns every non-experimental check.
func (c *Catalog) ExpectedChecks(platform string) []string {
	repoType, known := repoTypes[platform]
	var out []string
	for _, info := range c.List() {
		if info.Experimental {
			continue
		}
		if known && !slices.ContainsFunc(info.SupportedPlatforms, func(t string) bool {
			return strings.EqualFold(t, repoType)
		}) {
			continue
		}
		out = append(out, info.Name)
	}
	return out
}

// resolve returns the canonical check name for a case-insensitive input.
func (c *Catalog) resolve(name string) (string, bool) {
	if c.doc.CheckExists(name) {
		return name, true
	}
	for _, cd := range c.doc.GetChecks() {
		if strings.EqualFold(cd.GetName(), name) {
			return cd.GetName(), true
		}
	}
	return "", false
}

func (c *Catalog) toInfo(cd scdocs.CheckDoc) CheckInfo {
	name := cd.GetName()
	return CheckInfo{
		Name:               name,
		Short:              cd.GetShort(),
		Risk:               cd.GetRisk(),
		Tags:               normalizeTags(cd.GetTags()),
		SupportedPlatforms: cd.GetSupportedRepoTypes(),
		DocumentationURL:   cd.GetDocumentationURL(c.commitish),
		Experimental:       checks.IsExperimental(name),
	}
}

// normalizeTags splits any comma-joined tag entries and trims whitespace.
func normalizeTags(tags []string) []string {
	var out []string
	for _, t := range tags {
		for _, p := range strings.Split(t, ",") {
			if p = strings.TrimSpace(p); p != "" {
				out = append(out, p)
			}
		}
	}
	return out
}
