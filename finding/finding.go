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
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/ossf/scorecard/v4/finding/probe"
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
)

// Location represents the location of a finding.
// nolint: govet
type Location struct {
	Type      FileType `json:"type"`
	Path      string   `json:"path"`
	LineStart *uint    `json:"lineStart,omitempty"`
	LineEnd   *uint    `json:"lineEnd,omitempty"`
	Snippet   *string  `json:"snippet,omitempty"`
}

// Outcome is the result of a finding.
type Outcome int

// TODO(#2928): re-visit the finding definitions.
const (
	// NOTE: The additional '_' are intended for future use.
	// This allows adding outcomes without breaking the values
	// of existing outcomes.
	// OutcomeNegative indicates a negative outcome.
	OutcomeNegative Outcome = iota
	_
	_
	_
	// OutcomeNotAvailable indicates an unavailable outcome,
	// typically because an API call did not return an answer.
	OutcomeNotAvailable
	_
	_
	_
	// OutcomeError indicates an errors while running.
	// The results could not be determined.
	OutcomeError
	_
	_
	_
	// OutcomePositive indicates a positive outcome.
	OutcomePositive
	_
	_
	_
	// OutcomeNotSupported indicates a non-supported outcome.
	OutcomeNotSupported
)

// Finding represents a finding.
// nolint: govet
type Finding struct {
	Probe       string             `json:"probe"`
	Outcome     Outcome            `json:"outcome"`
	Message     string             `json:"message"`
	Location    *Location          `json:"location,omitempty"`
	Remediation *probe.Remediation `json:"remediation,omitempty"`
}

// AnonymousFinding is a finding without a corerpsonding probe ID.
type AnonymousFinding struct {
	Finding
	// Remove the probe ID from
	// the structure until the probes are GA.
	Probe string `json:"probe,omitempty"`
}

var errInvalid = errors.New("invalid")

// FromBytes creates a finding for a probe given its config file's content.
func FromBytes(content []byte, probeID string) (*Finding, error) {
	p, err := probe.FromBytes(content, probeID)
	if err != nil {
		// nolint
		return nil, err
	}
	f := &Finding{
		Probe:       p.ID,
		Outcome:     OutcomeNegative,
		Remediation: p.Remediation,
	}
	return f, nil
}

// New creates a new finding.
func New(loc embed.FS, probeID string) (*Finding, error) {
	p, err := probe.New(loc, probeID)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	f := &Finding{
		Probe:       p.ID,
		Outcome:     OutcomeNegative,
		Remediation: p.Remediation,
	}
	return f, nil
}

// NewWith create a finding with the desried location and outcome.
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

// NewWith create a negative finding with the desried location.
func NewNegative(efs embed.FS, probeID, text string, loc *Location,
) (*Finding, error) {
	f, err := NewWith(efs, probeID, text, loc, OutcomeNegative)
	if err != nil {
		return nil, fmt.Errorf("finding.NewWith: %w", err)
	}

	f = f.WithMessage(text)
	return f, nil
}

// NewNotAvailable create a finding with a NotAvailable outcome and the desried location.
func NewNotAvailable(efs embed.FS, probeID, text string, loc *Location,
) (*Finding, error) {
	f, err := NewWith(efs, probeID, text, loc, OutcomeNotAvailable)
	if err != nil {
		return nil, fmt.Errorf("finding.NewWith: %w", err)
	}

	f = f.WithMessage(text)
	return f, nil
}

// NewPositive create a positive finding with the desried location.
func NewPositive(efs embed.FS, probeID, text string, loc *Location,
) (*Finding, error) {
	f, err := NewWith(efs, probeID, text, loc, OutcomePositive)
	if err != nil {
		return nil, fmt.Errorf("finding.NewWith: %w", err)
	}

	f = f.WithMessage(text)
	return f, nil
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

// WithLocation adds a location to an existing finding.
// No copy is made.
func (f *Finding) WithLocation(loc *Location) *Finding {
	f.Location = loc
	if f.Remediation != nil && loc != nil {
		// Replace location data.
		f.Remediation.Text = strings.Replace(f.Remediation.Text,
			"${{ finding.location.path }}", loc.Path, -1)
		f.Remediation.Markdown = strings.Replace(f.Remediation.Markdown,
			"${{ finding.location.path }}", loc.Path, -1)
	}
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
	// Positive is not negative, remove the remediation.
	if o != OutcomeNegative {
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
			f.Remediation.Text = strings.Replace(f.Remediation.Text,
				fmt.Sprintf("${{ metadata.%s }}", k), v, -1)
			f.Remediation.Markdown = strings.Replace(f.Remediation.Markdown,
				fmt.Sprintf("${{ metadata.%s }}", k), v, -1)
		}
	}
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
	case "Negative":
		*o = OutcomeNegative
	case "Positive":
		*o = OutcomePositive
	case "NotAvailable":
		*o = OutcomeNotAvailable
	case "NotSupported":
		*o = OutcomeNotSupported
	case "Error":
		*o = OutcomeError
	default:
		return fmt.Errorf("%w: %q", errInvalid, str)
	}
	return nil
}
