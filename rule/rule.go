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

package rule

import (
	"embed"
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// RemediationEffort indicates the estimated effort necessary to remediate a finding.
type RemediationEffort string

const (
	// RemediationEffortLow indicates a low remediation effort.
	RemediationEffortLow RemediationEffort = "Low"
	// RemediationEffortMedium indicates a medium remediation effort.
	RemediationEffortMedium RemediationEffort = "Medium"
	// RemediationEffortHigh indicates a high remediation effort.
	RemediationEffortHigh RemediationEffort = "High"
)

// Remediation represents the remediation for a finding.
type Remediation struct {
	// Text for humans.
	Text string `json:"text"`
	// Text in markdown format for humans.
	Markdown string `json:"markdown"`
	// Patch for machines.
	Patch *string `json:"patch,omitempty"`
	// Effort to remediate.
	Effort RemediationEffort `json:"effort"`
}

// nolint: govet
type jsonRemediation struct {
	Text     []string          `yaml:"text"`
	Markdown []string          `yaml:"markdown"`
	Effort   RemediationEffort `yaml:"effort"`
}

type jsonRule struct {
	Short          string          `yaml:"short"`
	Desc           string          `yaml:"desc"`
	Motivation     string          `yaml:"motivation"`
	Implementation string          `yaml:"implementation"`
	Risk           string          `yaml:"risk"`
	Remediation    jsonRemediation `yaml:"remediation"`
}

// Risk indicates a risk.
type Risk string

const (
	// RiskCritical is a critical risk.
	RiskCritical Risk = "Critical"
	// RiskHigh is a high risk.
	RiskHigh Risk = "High"
	// RiskMedium is a medium risk.
	RiskMedium Risk = "Medium"
	// RiskLow is a low risk.
	RiskLow Risk = "Low"
	// RiskNone is a no risk.
	RiskNone Risk = "None"
)

// nolint: govet
type Rule struct {
	Name        string
	Short       string
	Desc        string
	Motivation  string
	Risk        Risk
	Remediation *Remediation
}

var errInvalid = errors.New("invalid")

func (r *Risk) GreaterThan(rr Risk) bool {
	m := map[Risk]int{
		RiskNone:     0,
		RiskLow:      1,
		RiskMedium:   2,
		RiskHigh:     3,
		RiskCritical: 4,
	}
	v := m[*r]
	vv := m[rr]
	return v > vv
}

func New(loc embed.FS, rule string) (*Rule, error) {
	content, err := loc.ReadFile(fmt.Sprintf("%s.yml", rule))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errInvalid, err)
	}

	r, err := parseFromJSON(content)
	if err != nil {
		return nil, err
	}

	if err := validate(r); err != nil {
		return nil, err
	}

	return &Rule{
		Name:       rule,
		Short:      r.Short,
		Desc:       r.Desc,
		Motivation: r.Motivation,
		Risk:       Risk(r.Risk),
		Remediation: &Remediation{
			Text:     strings.Join(r.Remediation.Text, "\n"),
			Markdown: strings.Join(r.Remediation.Markdown, "\n"),
			Effort:   r.Remediation.Effort,
		},
	}, nil
}

func validate(r *jsonRule) error {
	if err := validateRisk(r.Risk); err != nil {
		return fmt.Errorf("%w: %v", errInvalid, err)
	}
	if err := validateRemediation(r.Remediation); err != nil {
		return fmt.Errorf("%w: %v", errInvalid, err)
	}
	return nil
}

func validateRemediation(r jsonRemediation) error {
	switch r.Effort {
	case RemediationEffortHigh, RemediationEffortMedium, RemediationEffortLow:
		return nil
	default:
		return fmt.Errorf("%w: %v", errInvalid, fmt.Sprintf("remediation '%s'", r))
	}
}

func validateRisk(r string) error {
	switch Risk(r) {
	case RiskNone, RiskLow, RiskHigh, RiskMedium, RiskCritical:
		return nil
	default:
		return fmt.Errorf("%w: %v", errInvalid, fmt.Sprintf("risk '%s'", r))
	}
}

func parseFromJSON(content []byte) (*jsonRule, error) {
	r := jsonRule{}

	err := yaml.Unmarshal(content, &r)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errInvalid, err)
	}
	return &r, nil
}
