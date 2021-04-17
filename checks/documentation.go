// Copyright 2020 Security Scorecard Authors
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
package checks

import (
	//nolint
	_ "embed"

	"gopkg.in/yaml.v2"

	"github.com/pkg/errors"
)

//go:embed checks.yaml
var checks string

type documentation struct {
	Checks struct {
		SecurityPolicy struct {
			Description string `yaml:"description"`
		} `yaml:"Security-Policy"`
		Contributors struct {
			Description string `yaml:"description"`
		} `yaml:"Contributors"`
		FrozenDeps struct {
			Description string `yaml:"description"`
		} `yaml:"Frozen-Deps"`
		SignedTags struct {
			Description string `yaml:"description"`
		} `yaml:"Signed-Tags"`
		SignedReleases struct {
			Description string `yaml:"description"`
		} `yaml:"Signed-Releases"`
		CiTests struct {
			Description string `yaml:"description"`
		} `yaml:"CI-Tests"`
		CodeReview struct {
			Description string `yaml:"description"`
		} `yaml:"Code-Review"`
		CiiBestPractices struct {
			Description string `yaml:"description"`
		} `yaml:"CII-Best-Practices"`
		PullRequests struct {
			Description string `yaml:"description"`
		} `yaml:"Pull-Requests"`
		Fuzzing struct {
			Description string `yaml:"description"`
		} `yaml:"Fuzzing"`
		Sast struct {
			Description string `yaml:"description"`
		} `yaml:"SAST"`
		Active struct {
			Description string `yaml:"description"`
		} `yaml:"Active"`
		BranchProtection struct {
			Description string `yaml:"description"`
		} `yaml:"Branch-Protection"`
	} `yaml:"checks"`
}

// HelpDocumentation provides the documentation for the checks.
type HelpDocumentation struct {
	Description string
	HelpURL     string
}

// ActiveHelpDocumentation provides the help documentation for ActiveHelpDocumentation check.
func ActiveHelpDocumentation() (HelpDocumentation, error) {
	var helpdocs documentation

	if err := yaml.Unmarshal([]byte(checks), &helpdocs); err != nil {
		return HelpDocumentation{}, errors.Wrap(err, "unable to unmarshal the json")
	}

	return HelpDocumentation{
		Description: helpdocs.Checks.Active.Description,
		HelpURL:     "https://github.com/ossf/scorecard/blob/main/checks/checks.md#active",
	}, nil
}
