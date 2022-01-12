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

package format

import (
	"strings"

	docs "github.com/ossf/scorecard/v4/docs/checks"
)

type mockCheck struct {
	name, risk, short, description, url string
	tags, remediation, repos            []string
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
	l := make([]string, len(c.tags))
	for i := range c.tags {
		l[i] = strings.TrimSpace(c.tags[i])
	}
	return l
}

func (c *mockCheck) GetSupportedRepoTypes() []string {
	l := make([]string, len(c.repos))
	for i := range c.repos {
		l[i] = strings.TrimSpace(c.repos[i])
	}
	return l
}

func (c *mockCheck) GetDocumentationURL(commitish string) string {
	return c.url
}

type mockDoc struct {
	checks map[string]mockCheck
}

func (d *mockDoc) GetCheck(name string) (docs.CheckDoc, error) {
	// nolint: gosimple
	m, _ := d.checks[name]
	return &m, nil
}

func (d *mockDoc) GetChecks() []docs.CheckDoc {
	return nil
}

func (d *mockDoc) CheckExists(name string) bool {
	_, exists := d.checks[name]
	return exists
}
