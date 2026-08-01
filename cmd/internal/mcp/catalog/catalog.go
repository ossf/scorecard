// Package catalog provides offline access to OpenSSF Scorecard check
// documentation, sourced from the docs embedded in the scorecard module. It
// requires no network access and no scan.
package catalog

import (
	"errors"
	"fmt"
	"runtime/debug"
	"sort"
	"strings"

	scdocs "github.com/ossf/scorecard/v5/docs/checks"
)

// errUnknownCheck is returned when a check name is not recognized.
var errUnknownCheck = errors.New("unknown check")

// experimentalChecks lists checks excluded from Scorecard's default set unless
// SCORECARD_EXPERIMENTAL is set. Source of truth: ossf/scorecard
// checks/all_checks.go (getAll deletes these). This is version-sensitive;
// catalog_test.go's TestExperimentalChecksVersionPin fails on a scorecard
// version bump as a reminder to re-verify this list by hand. A test that
// diffs checks.GetAllWithExperimental() against checks.GetAll() directly
// would catch drift automatically, but importing the checks package pulls in
// Scorecard's full client/raw-check dependency graph just for a test — not
// worth that cost for a two-entry, rarely-changing list.
var experimentalChecks = map[string]bool{
	"Webhooks": true,
	"SBOM":     true,
}

// CheckInfo summarizes a check (used by list_checks and the index resource).
type CheckInfo struct {
	Name               string   `json:"name"`
	Short              string   `json:"short,omitempty"`
	Risk               string   `json:"risk,omitempty" jsonschema:"risk level: Critical, High, Medium, or Low"`
	Tags               []string `json:"tags,omitempty"`
	SupportedPlatforms []string `json:"supported_platforms,omitempty"`
	DocumentationURL   string   `json:"documentation_url,omitempty"`
	Experimental       bool     `json:"experimental" jsonschema:"experimental (off by default in Scorecard)"`
}

// CheckDetail is the full documentation for a check (used by explain_check and
// the per-check resource). It is intentionally flat (no embedding) so its
// inferred JSON schema matches its JSON marshaling.
type CheckDetail struct {
	Name               string   `json:"name"`
	Short              string   `json:"short,omitempty"`
	Risk               string   `json:"risk,omitempty" jsonschema:"risk level: Critical, High, Medium, or Low"`
	Tags               []string `json:"tags,omitempty"`
	SupportedPlatforms []string `json:"supported_platforms,omitempty"`
	DocumentationURL   string   `json:"documentation_url,omitempty"`
	Experimental       bool     `json:"experimental"`
	Description        string   `json:"description,omitempty"`
	Remediation        []string `json:"remediation,omitempty"`
}

// Catalog provides offline access to Scorecard check documentation.
type Catalog struct {
	doc              scdocs.Doc
	commitish        string
	scorecardVersion string
}

// New loads the embedded Scorecard check documentation.
func New() (*Catalog, error) {
	doc, err := scdocs.Read()
	if err != nil {
		return nil, fmt.Errorf("reading Scorecard check documentation: %w", err)
	}
	return &Catalog{
		doc:              doc,
		commitish:        "main",
		scorecardVersion: scorecardModuleVersion(),
	}, nil
}

// ScorecardVersion returns the scorecard module version the catalog was built
// against, or "" if unavailable.
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
		Experimental:       experimentalChecks[name],
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

func scorecardModuleVersion() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	for _, d := range bi.Deps {
		if d.Path == "github.com/ossf/scorecard/v5" {
			return d.Version
		}
	}
	return ""
}
