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

package pkg

import docs "github.com/ossf/scorecard/v2/docs/checks"

type mockCheck struct {
	name, risk, short, description, url string
	tags, remediation                   []string
}

func (c *mockCheck) GetName() string {
	return c.name
}

func (c *mockCheck) GetRisk() string {
	return c.risk
}

func (c *mockCheck) GetShort() string {
	return c.short
}

func (c *mockCheck) GetDescription() string {
	return c.description
}

func (c *mockCheck) GetRemediation() []string {
	return c.remediation
}

func (c *mockCheck) GetTags() []string {
	return c.tags
}

func (c *mockCheck) GetDocumentationURL() string {
	return c.url
}

type mockDoc struct {
	checks map[string]mockCheck
}

func (d *mockDoc) GetCheck(name string) (docs.CheckDocInterface, error) {
	m, _ := d.checks[name]
	return &m, nil
}

func (d *mockDoc) GetChecks() []docs.CheckDocInterface {
	return nil
}

func (d *mockDoc) CheckExists(name string) bool {
	_, exists := d.checks[name]
	return exists
}
