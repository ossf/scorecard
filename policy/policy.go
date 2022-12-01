// Copyright 2021 OpenSSF Scorecard Authors
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
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/ossf/scorecard/v4/checker"
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

// ParseFromFile takes a policy file and returns a `ScorecardPolicy`.
func ParseFromFile(policyFile string) (*ScorecardPolicy, error) {
	if policyFile != "" {
		data, err := os.ReadFile(policyFile)
		if err != nil {
			return nil, sce.WithMessage(sce.ErrScorecardInternal,
				fmt.Sprintf("os.ReadFile: %v", err))
		}

		sp, err := parseFromYAML(data)
		if err != nil {
			return nil,
				sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("spol.ParseFromYAML: %v", err))
		}

		return sp, nil
	}

	return nil, nil
}

// parseFromYAML parses a policy file and returns a `ScorecardPolicy`.
func parseFromYAML(b []byte) (*ScorecardPolicy, error) {
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
	allChecks := checks.GetAllWithExperimental()
	for n, p := range sp.Policies {
		if _, exists := allChecks[n]; !exists {
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

// GetEnabled returns the list of enabled checks.
func GetEnabled(
	sp *ScorecardPolicy,
	argsChecks []string,
	requiredRequestTypes []checker.RequestType,
) (checker.CheckNameToFnMap, error) {
	enabledChecks := checker.CheckNameToFnMap{}

	switch {
	case len(argsChecks) != 0:
		// Populate checks to run with the `--repo` CLI argument.
		for _, checkName := range argsChecks {
			if !isSupportedCheck(checkName, requiredRequestTypes) {
				return enabledChecks,
					sce.WithMessage(sce.ErrScorecardInternal,
						fmt.Sprintf("Unsupported RequestType %s by check: %s",
							fmt.Sprint(requiredRequestTypes), checkName))
			}
			if !enableCheck(checkName, &enabledChecks) {
				return enabledChecks,
					sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("invalid check: %s", checkName))
			}
		}
	case sp != nil:
		// Populate checks to run with policy file.
		for checkName := range sp.GetPolicies() {
			if !isSupportedCheck(checkName, requiredRequestTypes) {
				// We silently ignore the check, like we do
				// for the default case when no argsChecks
				// or policy are present.
				continue
			}

			if !enableCheck(checkName, &enabledChecks) {
				return enabledChecks,
					sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("invalid check: %s", checkName))
			}
		}
	default:
		// Enable all checks that are supported.
		for checkName := range checks.GetAll() {
			if !isSupportedCheck(checkName, requiredRequestTypes) {
				continue
			}
			if !enableCheck(checkName, &enabledChecks) {
				return enabledChecks,
					sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("invalid check: %s", checkName))
			}
		}
	}

	// If a policy was passed as argument, ensure all checks
	// to run have a corresponding policy.
	if sp != nil && !checksHavePolicies(sp, enabledChecks) {
		return enabledChecks, sce.WithMessage(sce.ErrScorecardInternal, "checks don't have policies")
	}

	return enabledChecks, nil
}

func checksHavePolicies(sp *ScorecardPolicy, enabledChecks checker.CheckNameToFnMap) bool {
	for checkName := range enabledChecks {
		_, exists := sp.Policies[checkName]
		if !exists {
			log.Printf("check %s has no policy declared", checkName)
			return false
		}
	}
	return true
}

func isSupportedCheck(checkName string, requiredRequestTypes []checker.RequestType) bool {
	unsupported := checker.ListUnsupported(
		requiredRequestTypes,
		checks.GetAllWithExperimental()[checkName].SupportedRequestTypes)
	return len(unsupported) == 0
}

// Enables checks by name.
func enableCheck(checkName string, enabledChecks *checker.CheckNameToFnMap) bool {
	if enabledChecks != nil {
		for key, checkFn := range checks.GetAllWithExperimental() {
			if strings.EqualFold(key, checkName) {
				(*enabledChecks)[key] = checkFn
				return true
			}
		}
	}
	return false
}
