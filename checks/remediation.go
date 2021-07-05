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

package checks

import (

	// Used to embed checks.yaml.
	_ "embed"
	"errors"
	"fmt"

	"gopkg.in/yaml.v2"

	"github.com/ossf/scorecard/checker"
)

var (
	//go:embed checks.yaml
	checksYAMLBytes []byte
	// ErrUnknownCheckName is when the name of a check is not present in checks.yaml.
	ErrUnknownCheckName = errors.New("unknown check name")
)

// CheckInfo is the structure for checks.yaml.
type CheckInfo struct {
	Description string        `yaml:"description"`
	Remediation []Remediation `yaml:"remediation"`
}

type Remediation struct {
	Steps []string `yaml:"steps"`
	Codes []string `yaml:"codes"`
}

// GetRemediationSteps returns a list of steps that can be done to remediate the issues (if there are any).
func GetRemediationSteps(result *checker.CheckResult) ([]string, error) {
	steps := make([]string, 0)
	if !result.DoShowRemediation() {
		return steps, nil
	}
	m := make(map[string]map[string]CheckInfo)
	err := yaml.Unmarshal(checksYAMLBytes, &m)
	if err != nil {
		return steps, fmt.Errorf("error unmarshalling checks.yaml: %w", err)
	}
	checkInfo, found := m["checks"][result.Name]
	if !found {
		return steps, fmt.Errorf("%w: %s", ErrUnknownCheckName, result.Name)
	}

	var remediationNoCodes *Remediation
	for i, remediation := range checkInfo.Remediation {
		if len(remediation.Codes) == 0 {
			remediationNoCodes = &checkInfo.Remediation[i]
		} else if anyItemsMatch(remediation.Codes, result.FailureCodes) {
			steps = append(steps, remediation.Steps...)
		}
	}

	// Whether any codes matched or not, we'll add the "generic" remediation steps.
	if remediationNoCodes != nil {
		steps = append(steps, remediationNoCodes.Steps...)
	}
	return steps, nil
}

func anyItemsMatch(slice1, slice2 []string) bool {
	for _, item1 := range slice1 {
		for _, item2 := range slice2 {
			if item1 == item2 {
				return true
			}
		}
	}
	return false
}
