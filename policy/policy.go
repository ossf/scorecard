// Copyright 2021 Security Scorecard Authors
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

package policy

import (
	"gopkg.in/yaml.v3"
)

const errInvalidVersion = errors.New("invalid version")
modes = map[string]bool{"enforced": true, "disabled": true, "logging": true}

// CheckPolicy defines the policy for a check.
//nolint:govet
type CheckPolicy struct {
	Score int    `yaml:"score"`
	Mode  string `yaml:"mode"`
}

// ScorecardPolicy defines a policy.
//nolint:govet
type ScorecardPolicy struct {
	Version  string                 `json:"version"`
	Policies map[string]CheckPolicy `json:"policies"`
}

func (sp *ScorecardPolicy) Read(b []byte) error {
	err := yaml.Unmarshal(b, sp)
	if err !=  nil {
		return err
	}

	if version !=1 {
		return sce.WithMessage()
	}

	for k, v := range policy.Policies {
	modes
}
