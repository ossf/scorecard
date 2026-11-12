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

	"go.yaml.in/yaml/v3"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks"
	"github.com/ossf/scorecard/v5/clients"
	docs "github.com/ossf/scorecard/v5/docs/checks"
	sce "github.com/ossf/scorecard/v5/errors"
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
	repoType clients.RepoType,
) (checker.CheckNameToFnMap, error) {
	enabledChecks := checker.CheckNameToFnMap{}

	// Build a case-insensitive repo-type lookup map once, only when needed.
	var repoTypeLookup map[string][]string
	if repoType != "" {
		if d, err := docs.Read(); err == nil {
			repoTypeLookup = make(map[string][]string)
			for _, c := range d.GetChecks() {
				repoTypeLookup[strings.ToLower(c.GetName())] = c.GetSupportedRepoTypes()
			}
		}
	}

	switch {
	case len(argsChecks) != 0:
		// Populate checks to run with the `--repo` CLI argument.
		for _, checkName := range argsChecks {
			registered, found := resolveCheckName(checkName)
			if !found {
				return enabledChecks,
					sce.WithMessage(sce.ErrInvalidCheck,
						fmt.Sprintf("%q is not a known check; run 'scorecard --help' to list the available checks",
							checkName))
			}
			// The request-mode lookup below is case-sensitive, so a name that
			// only differs in casing would otherwise be reported as a request
			// type problem rather than a spelling one.
			if len(requiredRequestTypes) != 0 && registered != checkName {
				return enabledChecks,
					sce.WithMessage(sce.ErrInvalidCheck,
						fmt.Sprintf("%q must be written %q when a request type is required",
							checkName, registered))
			}
			check := checks.GetAllWithExperimental()[registered]
			unsupported := checker.ListUnsupported(requiredRequestTypes, check.SupportedRequestTypes)
			if len(unsupported) != 0 {
				return enabledChecks,
					sce.WithMessage(sce.ErrUnsupportedCheck,
						fmt.Sprintf("%q does not support %s data, which this scan requires",
							registered, joinRequestTypes(unsupported)))
			}
			if !isSupportedRepoType(repoTypeLookup, registered, repoType) {
				continue
			}
			if !enableCheck(registered, &enabledChecks) {
				return enabledChecks,
					sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("invalid check: %s", registered))
			}
		}
	case sp != nil:
		// Populate checks to run with policy file.
		for checkName := range sp.GetPolicies() {
			if !isSupportedCheck(checkName, requiredRequestTypes) {
				continue
			}
			if !isSupportedRepoType(repoTypeLookup, checkName, repoType) {
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
			if !isSupportedRepoType(repoTypeLookup, checkName, repoType) {
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
		_, exists := sp.GetPolicies()[checkName]
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

// resolveCheckName returns the registered name of a check, preferring an exact
// match over a case-insensitive one so that the caller can tell a misspelled
// check apart from one written with the wrong casing.
func resolveCheckName(checkName string) (string, bool) {
	allChecks := checks.GetAllWithExperimental()
	if _, exists := allChecks[checkName]; exists {
		return checkName, true
	}
	for name := range allChecks {
		if strings.EqualFold(name, checkName) {
			return name, true
		}
	}
	return "", false
}

// joinRequestTypes renders request types for an error message, e.g.
// "file-based and commit-based".
func joinRequestTypes(types []checker.RequestType) string {
	names := make([]string, 0, len(types))
	for _, t := range types {
		names = append(names, requestTypeName(t))
	}
	return strings.Join(names, " and ")
}

func requestTypeName(t checker.RequestType) string {
	switch t {
	case checker.FileBased:
		return "file-based"
	case checker.CommitBased:
		return "commit-based"
	default:
		return fmt.Sprintf("request type %d", int(t))
	}
}

// isSupportedRepoType checks if a check supports the given repo type
// using a pre-built lookup map. If the map is nil (docs unavailable or
// repoType empty) or the check is not found, it defaults to supported.
func isSupportedRepoType(repoTypeLookup map[string][]string, checkName string, repoType clients.RepoType) bool {
	if repoTypeLookup == nil {
		return true
	}

	supportedTypes, exists := repoTypeLookup[strings.ToLower(checkName)]
	if !exists {
		return true
	}

	for _, t := range supportedTypes {
		if strings.EqualFold(t, string(repoType)) {
			return true
		}
	}
	return false
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
