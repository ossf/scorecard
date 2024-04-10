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
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/ossf/scorecard/v4/clients"
)

// RemediationEffort indicates the estimated effort necessary to remediate a finding.
type RemediationEffort int

const (
	// RemediationEffortNone indicates a no remediation effort.
	RemediationEffortNone RemediationEffort = iota
	// RemediationEffortLow indicates a low remediation effort.
	RemediationEffortLow
	// RemediationEffortMedium indicates a medium remediation effort.
	RemediationEffortMedium
	// RemediationEffortHigh indicates a high remediation effort.
	RemediationEffortHigh
)

// Remediation represents the remediation for a finding.
type Remediation struct {
	// Patch for machines.
	Patch *string `json:"patch,omitempty"`
	// Text for humans.
	Text string `json:"text"`
	// Text in markdown format for humans.
	Markdown string `json:"markdown"`
	// Effort to remediate.
	Effort RemediationEffort `json:"effort"`
}

type yamlRemediation struct {
	OnOutcome Outcome           `yaml:"onOutcome"`
	Text      []string          `yaml:"text"`
	Markdown  []string          `yaml:"markdown"`
	Effort    RemediationEffort `yaml:"effort"`
}

type yamlEcosystem struct {
	Languages []string `yaml:"languages"`
	Clients   []string `yaml:"clients"`
}

var supportedClients = map[string]bool{
	"github":   true,
	"gitlab":   true,
	"localdir": true,
}

type yamlProbe struct {
	ID             string          `yaml:"id"`
	Short          string          `yaml:"short"`
	Motivation     string          `yaml:"motivation"`
	Implementation string          `yaml:"implementation"`
	Ecosystem      yamlEcosystem   `yaml:"ecosystem"`
	Remediation    yamlRemediation `yaml:"remediation"`
}

type probe struct {
	ID                 string
	Short              string
	Motivation         string
	Implementation     string
	Remediation        *Remediation
	RemediateOnOutcome Outcome
}

func probeFromBytes(content []byte, probeID string) (*probe, error) {
	r, err := parseFromYAML(content)
	if err != nil {
		return nil, err
	}

	if err := validate(r, probeID); err != nil {
		return nil, err
	}

	return &probe{
		ID:             r.ID,
		Short:          r.Short,
		Motivation:     r.Motivation,
		Implementation: r.Implementation,
		Remediation: &Remediation{
			Text:     strings.Join(r.Remediation.Text, "\n"),
			Markdown: strings.Join(r.Remediation.Markdown, "\n"),
			Effort:   r.Remediation.Effort,
		},
		RemediateOnOutcome: r.Remediation.OnOutcome,
	}, nil
}

// New create a new probe.
func newProbe(loc embed.FS, probeID string) (*probe, error) {
	content, err := loc.ReadFile("def.yml")
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return probeFromBytes(content, probeID)
}

func validate(r *yamlProbe, probeID string) error {
	if err := validateID(r.ID, probeID); err != nil {
		return err
	}
	if err := validateRemediation(r.Remediation); err != nil {
		return err
	}
	if err := validateEcosystem(r.Ecosystem); err != nil {
		return err
	}
	return nil
}

func validateID(actual, expected string) error {
	if actual != expected {
		return fmt.Errorf("%w: ID: read '%v', expected '%v'", errInvalid,
			actual, expected)
	}
	return nil
}

func validateRemediation(r yamlRemediation) error {
	if err := validateRemediationOutcomeTrigger(r.OnOutcome); err != nil {
		return err
	}
	switch r.Effort {
	case RemediationEffortHigh, RemediationEffortMedium, RemediationEffortLow:
		return nil
	default:
		return fmt.Errorf("%w: %v", errInvalid, fmt.Sprintf("remediation '%v'", r))
	}
}

func validateEcosystem(r yamlEcosystem) error {
	if err := validateSupportedLanguages(r); err != nil {
		return err
	}
	if err := validateSupportedClients(r); err != nil {
		return err
	}
	return nil
}

func validateRemediationOutcomeTrigger(o Outcome) error {
	switch o {
	case OutcomeTrue, OutcomeFalse, OutcomeNotApplicable, OutcomeNotAvailable, OutcomeNotSupported, OutcomeError:
		return nil
	default:
		return fmt.Errorf("%w: unknown outcome: %v", errInvalid, o)
	}
}

func validateSupportedLanguages(r yamlEcosystem) error {
	for _, lang := range r.Languages {
		switch clients.LanguageName(lang) {
		case clients.Go, clients.Python, clients.JavaScript,
			clients.Cpp, clients.C, clients.TypeScript,
			clients.Java, clients.CSharp, clients.Ruby,
			clients.PHP, clients.StarLark, clients.Scala,
			clients.Kotlin, clients.Swift, clients.Rust,
			clients.Haskell, clients.All, clients.Dockerfile,
			clients.ObjectiveC:
			continue
		default:
			return fmt.Errorf("%w: %v", errInvalid, fmt.Sprintf("language '%v'", r))
		}
	}
	return nil
}

func validateSupportedClients(r yamlEcosystem) error {
	for _, lang := range r.Clients {
		if _, ok := supportedClients[lang]; !ok {
			return fmt.Errorf("%w: %v", errInvalid, fmt.Sprintf("client '%v'", r))
		}
	}
	return nil
}

func parseFromYAML(content []byte) (*yamlProbe, error) {
	r := yamlProbe{}

	err := yaml.Unmarshal(content, &r)
	if err != nil {
		return nil, fmt.Errorf("unable to parse yaml: %w", err)
	}
	return &r, nil
}

// UnmarshalYAML is a custom unmarshalling function
// to transform the string into an enum.
func (r *RemediationEffort) UnmarshalYAML(n *yaml.Node) error {
	var str string
	if err := n.Decode(&str); err != nil {
		return fmt.Errorf("%w: %w", errInvalid, err)
	}

	switch n.Value {
	case "Low":
		*r = RemediationEffortLow
	case "Medium":
		*r = RemediationEffortMedium
	case "High":
		*r = RemediationEffortHigh
	default:
		return fmt.Errorf("%w: effort:%q", errInvalid, str)
	}
	return nil
}

// String stringifies the enum.
func (r *RemediationEffort) String() string {
	switch *r {
	case RemediationEffortLow:
		return "Low"
	case RemediationEffortMedium:
		return "Medium"
	case RemediationEffortHigh:
		return "High"
	default:
		return ""
	}
}
