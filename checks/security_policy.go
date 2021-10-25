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
	"strings"

	"github.com/ossf/scorecard/v3/checker"
)

// CheckSecurityPolicy is the registred name for SecurityPolicy.
const CheckSecurityPolicy = "Security-Policy"

//nolint:gochecknoinits
func init() {
	registerCheck(CheckSecurityPolicy, SecurityPolicy)
}

// SecurityPolicy runs Security-Policy check.
func SecurityPolicy(c *checker.CheckRequest) checker.CheckResult {
	var r bool
	// Check repository for repository-specific policy.
	onFile := func(name string, dl checker.DetailLogger, data FileCbData) (bool, error) {
		pdata := FileGetCbDataAsBoolPointer(data)
		switch {
		case strings.EqualFold(name, "security.md"),
			strings.EqualFold(name, ".github/security.md"),
			strings.EqualFold(name, "docs/security.md"),
			isSecurityRstFound(name):
			c.Dlogger.Info3(&checker.LogMessage{
				Path: name,
				Type: checker.FileTypeSource,
				// Source file must have line number > 0.
				Offset: 1,
				Text:   "security policy detected",
			})
			*pdata = true
			return false, nil
		default:
			return true, nil
		}
	}
	err := CheckIfFileExists(CheckSecurityPolicy, c, onFile, &r)
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckSecurityPolicy, err)
	}
	if r {
		return checker.CreateMaxScoreResult(CheckSecurityPolicy, "security policy file detected")
	}

	return checker.CreateMinScoreResult(CheckSecurityPolicy, "security policy file not detected")
}

func isSecurityRstFound(name string) bool {
	if strings.EqualFold(name, "doc/security.rst") {
		return true
	} else if strings.EqualFold(name, "docs/security.rst") {
		return true
	}
	return false
}
