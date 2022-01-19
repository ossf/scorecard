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

	"github.com/ossf/scorecard/v4/checks"
	sce "github.com/ossf/scorecard/v4/errors"
)

var (
	errInvalidVersion = errors.New("invalid version")
	errInvalidCheck   = errors.New("invalid check name")
	errInvalidScore   = errors.New("invalid score")
	errInvalidMode    = errors.New("invalid mode")
	errRepeatingCheck = errors.New("check has multiple definitions")
)

var allowedVersions = map[int]bool{1: true}

var modes = map[string]bool{"enforced": true, "disabled": true}

type checkPolicy struct {
	Mode  string `yaml:"mode"`
	Score int    `yaml:"score"`
}

type scorecardPolicy struct {
	Policies map[string]checkPolicy `yaml:"policies"`
	Version  int                    `yaml:"version"`
}

func isAllowedVersion(v int) bool {
	_, exists := allowedVersions[v]
	return exists
}

func modeToProto(m string) CheckPolicy_Mode {
	switch m {
	default:
		panic("will never happen")
	case "enforced":
		return CheckPolicy_ENFORCED
	case "disabled":
		return CheckPolicy_DISABLED
	}
}

// ParseFromYAML parses a policy file and returns
// a scorecardPolicy.
func ParseFromYAML(b []byte) (*ScorecardPolicy, error) {
	// Internal golang for unmarshalling the policy file.
	sp := scorecardPolicy{}
	// Protobuf-defined policy (policy.proto and policy.pb.go).
	retPolicy := ScorecardPolicy{Policies: map[string]*CheckPolicy{}}

	err := yaml.Unmarshal(b, &sp)
	if err != nil {
		return &retPolicy, sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}

	if !isAllowedVersion(sp.Version) {
		return &retPolicy, sce.WithMessage(sce.ErrScorecardInternal, errInvalidVersion.Error())
	}

	// Set version.
	retPolicy.Version = int32(sp.Version)

	checksFound := make(map[string]bool)
	for n, p := range sp.Policies {
		if _, exists := checks.AllChecks[n]; !exists {
			return &retPolicy, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("%v: %v", errInvalidCheck.Error(), n))
		}

		_, exists := modes[p.Mode]
		if !exists {
			return &retPolicy, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("%v: %v", errInvalidMode.Error(), p.Mode))
		}

		if p.Score < 0 || p.Score > 10 {
			return &retPolicy, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("%v: %v", errInvalidScore.Error(), p.Score))
		}

		_, exists = checksFound[n]
		if exists {
			return &retPolicy, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("%v: %v", errRepeatingCheck.Error(), n))
		}
		checksFound[n] = true

		// Add an entry to the policy.
		retPolicy.Policies[n] = &CheckPolicy{
			Score: int32(p.Score),
			Mode:  modeToProto(p.Mode),
		}
	}

	return &retPolicy, nil
}
