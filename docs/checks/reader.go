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

// Package checks contains util fns for reading input YAML file.
package checks

import (
	"errors"
	"fmt"
	"strings"

	"internal/internaldocs"

	sce "github.com/ossf/scorecard/v2/errors"
)

var errCheckNotExist = errors.New("check does not exist")

const docURL = "https://github.com/ossf/scorecard/blob/%s/docs/checks.md"

// Doc defines the documentation interface.
type Doc interface {
	GetCheck(name string) (CheckDoc, error)
	GetChecks() []CheckDoc
	CheckExists(name string) bool
}

// DocImpl implements `Doc` interface and
// contains checks' documentation.
type DocImpl struct {
	internaldoc *internaldocs.InternalDoc
}

// Read loads the checks' documentation.
func Read() (Doc, error) {
	m, e := internaldocs.ReadDoc()
	if e != nil {
		d := DocImpl{}
		return &d, fmt.Errorf("internaldocs.ReadDoc: %w", e)
	}

	d := DocImpl{internaldoc: &m}
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
