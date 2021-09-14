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
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/ossf/scorecard/v2/checks"
	sce "github.com/ossf/scorecard/v2/errors"
)

var (
	errInvalidVersion = errors.New("invalid version")
	errInvalidCheck   = errors.New("invalid check name")
	errInvalidScore   = errors.New("invalid score")
	errInvalidMode    = errors.New("invalid mode")
	errRepeatingCheck = errors.New("check has multiple definitions")
)

var modes = map[string]bool{"enforced": true, "disabled": true, "logging": true}

// CheckPolicy defines the policy for a check.
//nolint:govet
type CheckPolicy struct {
	Score int    `yaml:"score"`
	Mode  string `yaml:"mode"`
}

// ScorecardPolicy defines a policy.
//nolint:govet
type ScorecardPolicy struct {
	Version  int                    `json:"version"`
	Policies map[string]CheckPolicy `json:"policies"`
}

func (sp *ScorecardPolicy) Read(b []byte) error {
	err := yaml.Unmarshal(b, sp)
	if err != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}

	if sp.Version != 1 {
		return sce.WithMessage(sce.ErrScorecardInternal, errInvalidVersion.Error())
	}

	checksFound := make(map[string]bool)
	for n, p := range sp.Policies {
		_, exists := checks.AllChecks[n]
		if !exists {
			return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("%v: %v", errInvalidCheck.Error(), n))
		}

		if p.Score < 0 || p.Score > 10 {
			return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("%v: %v", errInvalidScore.Error(), p.Score))
		}

		_, exists = modes[p.Mode]
		if !exists {
			return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("%v: %v", errInvalidMode.Error(), p.Mode))
		}

		_, exists = checksFound[n]
		if exists {
			return sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("%v: %v", errRepeatingCheck.Error(), n))
		}
		checksFound[n] = true
	}

	return nil
}
