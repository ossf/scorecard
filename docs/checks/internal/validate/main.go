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
package main

import (
	"fmt"
	"strings"

	"github.com/ossf/scorecard/v2/checks"
	docs "github.com/ossf/scorecard/v2/docs/checks"
)

func main() {
	m, err := docs.Read()
	if err != nil {
		panic(fmt.Errorf("docs.Read: %w", err))
	}

	allChecks := checks.AllChecks
	for check := range allChecks {
		c, e := m.GetCheck(check)
		if e != nil {
			panic(fmt.Errorf("GetCheck: %w: %s", e, check))
		}

		if c.GetDescription() == "" {
			// nolint: goerr113
			panic(fmt.Errorf("description for checkName: %s is empty", check))
		}
		if strings.TrimSpace(strings.Join(c.GetRemediation(), "")) == "" {
			// nolint: goerr113
			panic(fmt.Errorf("remediation for checkName: %s is empty", check))
		}
		if c.GetShort() == "" {
			// nolint: goerr113
			panic(fmt.Errorf("short for checkName: %s is empty", check))
		}
		if len(c.GetTags()) == 0 {
			// nolint: goerr113
			panic(fmt.Errorf("tags for checkName: %s is empty", check))
		}
	}
	for _, check := range m.GetChecks() {
		if _, exists := allChecks[check.GetName()]; !exists {
			// nolint: goerr113
			panic(fmt.Errorf("check present in checks.yaml is not part of `checks.AllChecks`: %s", check.GetName()))
		}
	}
}
