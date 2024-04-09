// Copyright 2023 OpenSSF Scorecard Authors
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

package finding

import (
	"embed"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// FileType is the type of a file.
type FileType int

const (
	// FileTypeNone must be `0`.
	FileTypeNone FileType = iota
	// FileTypeSource is for source code files.
	FileTypeSource
	// FileTypeBinary is for binary files.
	FileTypeBinary
	// FileTypeText is for text files.
	FileTypeText
	// FileTypeURL for URLs.
	FileTypeURL
	// FileTypeBinaryVerified for verified binary files.
	FileTypeBinaryVerified
)

// Location represents the location of a finding.
type Location struct {
	LineStart *uint    `json:"lineStart,omitempty"`
	LineEnd   *uint    `json:"lineEnd,omitempty"`
	Snippet   *string  `json:"snippet,omitempty"`
	Path      string   `json:"path"`
	Type      FileType `json:"type"`
}

// Outcome is the result of a finding.
type Outcome string

// TODO(#2928): re-visit the finding definitions.
const (
	// OutcomeFalse indicates the answer to the probe's question is "false" or "no".
	OutcomeFalse Outcome = "False"
	// OutcomeNotAvailable indicates an unavailable outcome,
	// typically because an API call did not return an answer.
	OutcomeNotAvailable Outcome = "NotAvailable"
	// OutcomeError indicates an errors while running.
	// The results could not be determined.
	OutcomeError Outcome = "Error"
	// OutcomeTrue indicates the answer to the probe's question is "true" or "yes".
	OutcomeTrue Outcome = "True"
	// OutcomeNotSupported indicates a non-supported outcome.
	OutcomeNotSupported Outcome = "NotSupported"
	// OutcomeNotApplicable indicates if a finding should not
	// be considered in evaluation.
	OutcomeNotApplicable Outcome = "NotApplicable"
)

// Finding represents a finding.
type Finding struct {
	Location    *Location         `json:"location,omitempty"`
	Remediation *Remediation      `json:"remediation,omitempty"`
	Values      map[string]string `json:"values,omitempty"`
	Probe       string            `json:"probe"`
	Message     string            `json:"message"`
	Outcome     Outcome           `json:"outcome"`
}

// AnonymousFinding is a finding without a corresponding probe ID.
type AnonymousFinding struct {
	Probe string `json:"probe,omitempty"`
	Finding
}

var errInvalid = errors.New("invalid")

// FromBytes creates a finding for a probe given its config file's content.
func FromBytes(content []byte, probeID string) (*Finding, error) {
	p, err := probeFromBytes(content, probeID)
	if err != nil {
		return nil, err
	}
	f := &Finding{
		Probe:       p.ID,
		Outcome:     OutcomeFalse,
		Remediation: p.Remediation,
	}
	return f, nil
}

// New creates a new finding.
func New(loc embed.FS, probeID string) (*Finding, error) {
	p, err := NewProbe(loc, probeID)
	if err != nil {
		return nil, err
	}

	f := &Finding{
		Probe:       p.ID,
		Outcome:     OutcomeFalse,
		Remediation: p.Remediation,
	}
	return f, nil
}

// NewWith create a finding with the desired location and outcome.
func NewWith(efs embed.FS, probeID, text string, loc *Location,
	o Outcome,
) (*Finding, error) {
	f, err := New(efs, probeID)
	if err != nil {
		return nil, fmt.Errorf("finding.New: %w", err)
	}

	f = f.WithMessage(text).WithOutcome(o).WithLocation(loc)
	return f, nil
}

// NewFalse create a false finding with the desired location.
func NewFalse(efs embed.FS, probeID, text string, loc *Location,
) (*Finding, error) {
	return NewWith(efs, probeID, text, loc, OutcomeFalse)
}

// NewNotAvailable create a finding with a NotAvailable outcome and the desired location.
func NewNotAvailable(efs embed.FS, probeID, text string, loc *Location,
) (*Finding, error) {
	return NewWith(efs, probeID, text, loc, OutcomeNotAvailable)
}

// NewTrue create a true finding with the desired location.
func NewTrue(efs embed.FS, probeID, text string, loc *Location,
) (*Finding, error) {
	return NewWith(efs, probeID, text, loc, OutcomeTrue)
}

// Anonymize removes the probe ID and outcome
// from the finding. It is a temporary solution
// to integrate the code in the details without exposing
// too much information.
func (f *Finding) Anonymize() *AnonymousFinding {
	return &AnonymousFinding{Finding: *f}
}

// WithMessage adds a message to an existing finding.
// No copy is made.
func (f *Finding) WithMessage(text string) *Finding {
	f.Message = text
	return f
}

// UniqueProbesEqual checks the probe names present in a list of findings
// and compare them against an expected list.
func UniqueProbesEqual(findings []Finding, probes []string) bool {
	// Collect unique probes from findings.
	fm := make(map[string]bool)
	for i := range findings {
		f := &findings[i]
		fm[f.Probe] = true
	}
	// Collect probes from list.
	pm := make(map[string]bool)
	for i := range probes {
		p := &probes[i]
		pm[*p] = true
	}
	return reflect.DeepEqual(pm, fm)
}

// WithLocation adds a location to an existing finding.
// No copy is made.
func (f *Finding) WithLocation(loc *Location) *Finding {
	f.Location = loc
	if f.Remediation != nil && f.Location != nil {
		// Replace location data.
		f.Remediation.Text = strings.ReplaceAll(f.Remediation.Text,
			"${{ finding.location.path }}", f.Location.Path)
		f.Remediation.Markdown = strings.ReplaceAll(f.Remediation.Markdown,
			"${{ finding.location.path }}", f.Location.Path)
	}
	return f
}

// WithValues sets the values to an existing finding.
// No copy is made.
func (f *Finding) WithValues(values map[string]string) *Finding {
	f.Values = values
	return f
}

// WithPatch adds a patch to an existing finding.
// No copy is made.
func (f *Finding) WithPatch(patch *string) *Finding {
	f.Remediation.Patch = patch
	// NOTE: we will update the remediation section
	// using patch information, e.g. ${{ patch.content }}.
	return f
}

// WithOutcome adds an outcome to an existing finding.
// No copy is made.
// WARNING: this function should be called at most once for a finding.
func (f *Finding) WithOutcome(o Outcome) *Finding {
	f.Outcome = o
	// Currently only false probes have remediations.
	// TODO(#3654) this is a temporary mechanical conversion.
	if o != OutcomeFalse {
		f.Remediation = nil
	}

	return f
}

// WithRemediationMetadata adds remediation metadata to an existing finding.
// No copy is made.
func (f *Finding) WithRemediationMetadata(values map[string]string) *Finding {
	if f.Remediation != nil {
		// Replace all dynamic values.
		for k, v := range values {
			// Replace metadata.
			f.Remediation.Text = strings.ReplaceAll(f.Remediation.Text,
				fmt.Sprintf("${{ metadata.%s }}", k), v)
			f.Remediation.Markdown = strings.ReplaceAll(f.Remediation.Markdown,
				fmt.Sprintf("${{ metadata.%s }}", k), v)
		}
	}
	return f
}

// WithValue adds a value to f.Values.
// No copy is made.
func (f *Finding) WithValue(k, v string) *Finding {
	if f.Values == nil {
		f.Values = make(map[string]string)
	}
	f.Values[k] = v
	return f
}

// UnmarshalYAML is a custom unmarshalling function
// to transform the string into an enum.
func (o *Outcome) UnmarshalYAML(n *yaml.Node) error {
	var str string
	if err := n.Decode(&str); err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	switch n.Value {
	case "False":
		*o = OutcomeFalse
	case "True":
		*o = OutcomeTrue
	case "NotAvailable":
		*o = OutcomeNotAvailable
	case "NotSupported":
		*o = OutcomeNotSupported
	case "NotApplicable":
		*o = OutcomeNotApplicable
	case "Error":
		*o = OutcomeError
	default:
		return fmt.Errorf("%w: %q", errInvalid, str)
	}
	return nil
}
