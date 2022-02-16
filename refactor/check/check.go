// Copyright 2022 Allstar Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package check

import (
	"fmt"
	"log"
	"strings"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks"
	sce "github.com/ossf/scorecard/v4/errors"
	spol "github.com/ossf/scorecard/v4/policy"
)

func GetAll() checker.CheckNameToFnMap {
	// Returns the full list of checks, given any environment variable constraints.
	possibleChecks := checks.AllChecks
	return possibleChecks
}

func GetEnabled(sp *spol.ScorecardPolicy, argsChecks []string,
	requiredRequestTypes []checker.RequestType) (checker.CheckNameToFnMap, error) {
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
		for checkName := range GetAll() {
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

func checksHavePolicies(sp *spol.ScorecardPolicy, enabledChecks checker.CheckNameToFnMap) bool {
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
		checks.AllChecks[checkName].SupportedRequestTypes)
	return len(unsupported) == 0
}

// Enables checks by name.
func enableCheck(checkName string, enabledChecks *checker.CheckNameToFnMap) bool {
	if enabledChecks != nil {
		for key, checkFn := range GetAll() {
			if strings.EqualFold(key, checkName) {
				(*enabledChecks)[key] = checkFn
				return true
			}
		}
	}
	return false
}
