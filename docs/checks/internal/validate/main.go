// Copyright 2020 Security Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"fmt"
	"strings"

	"github.com/ossf/scorecard/v4/checks"
	docs "github.com/ossf/scorecard/v4/docs/checks"
)

var allowedRisks = map[string]bool{"Critical": true, "High": true, "Medium": true, "Low": true}

func main() {
	m, err := docs.Read()
	if err != nil {
		panic(fmt.Sprintf("docs.Read: %v", err))
	}

	allChecks := checks.GetAll()
	for check := range allChecks {
		c, e := m.GetCheck(check)
		if e != nil {
			panic(fmt.Sprintf("GetCheck: %v: %s", e, check))
		}

		if check != c.GetName() {
			panic(fmt.Sprintf("invalid checkName: %s != %s", check, c.GetName()))
		}
		if c.GetDescription() == "" {
			panic(fmt.Sprintf("description for checkName: %s is empty", check))
		}
		if strings.TrimSpace(strings.Join(c.GetRemediation(), "")) == "" {
			panic(fmt.Sprintf("remediation for checkName: %s is empty", check))
		}
		if c.GetShort() == "" {
			panic(fmt.Sprintf("short for checkName: %s is empty", check))
		}
		if len(c.GetTags()) == 0 {
			panic(fmt.Sprintf("tags for checkName: %s is empty", check))
		}
		r := c.GetRisk()
		if _, exists := allowedRisks[r]; !exists {
			panic(fmt.Sprintf("risk for checkName: %s is invalid: '%s'", check, r))
		}
	}
	for _, check := range m.GetChecks() {
		if _, exists := allChecks[check.GetName()]; !exists {
			panic(fmt.Sprintf("check present in checks.yaml is not part of `checks.GetAll()`: %s", check.GetName()))
		}
	}
}
