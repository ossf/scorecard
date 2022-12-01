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

// Package checks contains documentation about checks.
package checks

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ossf/scorecard/v4/docs/checks/internal"
	sce "github.com/ossf/scorecard/v4/errors"
)

var errCheckNotExist = errors.New("check does not exist")

const docURL = "https://github.com/ossf/scorecard/blob/%s/docs/checks.md"

// DocImpl implements `Doc` interface and
// contains checks' documentation.
type DocImpl struct {
	internaldoc internal.Doc
}

// Read loads the checks' documentation.
func Read() (Doc, error) {
	m, e := internal.ReadDoc()
	if e != nil {
		d := DocImpl{}
		return &d, fmt.Errorf("internal.ReadDoc: %w", e)
	}

	d := DocImpl{internaldoc: m}
	return &d, nil
}

// GetCheck returns the information for check `name`.
func (d *DocImpl) GetCheck(name string) (CheckDoc, error) {
	ic, exists := d.internaldoc.InternalChecks[name]
	if !exists {
		//nolint: wrapcheck
		return nil, sce.CreateInternal(errCheckNotExist, "")
	}
	// Set the name and URL.
	ic.Name = name
	ic.URL = fmt.Sprintf("%s#%s", docURL, strings.ToLower(name))
	return &CheckDocImpl{internalCheck: ic}, nil
}

// GetChecks returns the information for check `name`.
func (d *DocImpl) GetChecks() []CheckDoc {
	var checks []CheckDoc
	for k := range d.internaldoc.InternalChecks {
		//nolint: errcheck
		check, _ := d.GetCheck(k)
		checks = append(checks, check)
	}

	return checks
}

// CheckExists returns whether the check `name` exists or not.
func (d DocImpl) CheckExists(name string) bool {
	_, exists := d.internaldoc.InternalChecks[name]
	return exists
}

// CheckDocImpl implementts `CheckDoc` interface and
// stores documentation about a check.
type CheckDocImpl struct {
	internalCheck internal.Check
}

// GetName returns the name of the check.
func (c *CheckDocImpl) GetName() string {
	return c.internalCheck.Name
}

// GetRisk returns the risk of the check.
func (c *CheckDocImpl) GetRisk() string {
	return c.internalCheck.Risk
}

// GetShort returns the short description of the check.
func (c *CheckDocImpl) GetShort() string {
	return c.internalCheck.Short
}

// GetDescription returns the full description of the check.
func (c *CheckDocImpl) GetDescription() string {
	return c.internalCheck.Description
}

// GetRemediation returns the remediation of the check.
func (c *CheckDocImpl) GetRemediation() []string {
	return c.internalCheck.Remediation
}

// GetSupportedRepoTypes returns the list of repo
// types the check supports.
func (c *CheckDocImpl) GetSupportedRepoTypes() []string {
	l := strings.Split(c.internalCheck.Repos, ",")
	for i := range l {
		l[i] = strings.TrimSpace(l[i])
	}
	return l
}

// GetTags returns the list of tags or the check.
func (c *CheckDocImpl) GetTags() []string {
	l := strings.Split(c.internalCheck.Tags, ",")
	for i := range l {
		l[i] = strings.TrimSpace(l[i])
	}
	return l
}

// GetDocumentationURL returns the URL for the documentation of check `name`.
func (c *CheckDocImpl) GetDocumentationURL(commitish string) string {
	com := commitish
	if com == "" || com == "unknown" {
		com = "main"
	}
	return fmt.Sprintf(c.internalCheck.URL, com)
}
