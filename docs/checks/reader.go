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

	"internal/idocs"

	sce "github.com/ossf/scorecard/v2/errors"
)

var errCheckNotExist = errors.New("check does not exist")

const docURL = "https://github.com/ossf/scorecard/blob/main/docs/checks.md"

// DocInterface defines the documentation interface.
type DocInterface interface {
	GetCheck(name string) (CheckDocInterface, error)
	GetChecks() []CheckDocInterface
	CheckExists(name string) bool
}

// Doc contains checks' documentation.
type Doc struct {
	idoc *idocs.InternalDoc
}

// Read loads the checks' documentation.
func Read() (DocInterface, error) {
	m, e := idocs.ReadDoc()
	if e != nil {
		d := Doc{}
		return &d, fmt.Errorf("idocs.ReadDoc: %w", e)
	}

	d := Doc{idoc: &m}
	return &d, nil
}

// GetCheck returns the information for check `name`.
func (d *Doc) GetCheck(name string) (CheckDocInterface, error) {
	ic, exists := d.idoc.InternalChecks[name]
	if !exists {
		//nolint: wrapcheck
		return nil, sce.CreateInternal(errCheckNotExist, "")
	}
	// Set the name and URL.
	ic.Name = name
	ic.URL = fmt.Sprintf("%s#%s", docURL, strings.ToLower(name))
	return &Check{icheck: ic}, nil
}

// GetChecks returns the information for check `name`.
func (d *Doc) GetChecks() []CheckDocInterface {
	var checks []CheckDocInterface
	for k := range d.idoc.InternalChecks {
		//nolint: errcheck
		check, _ := d.GetCheck(k)
		checks = append(checks, check)
	}

	return checks
}

// CheckExists returns whether the check `name` exists or not.
func (d Doc) CheckExists(name string) bool {
	_, exists := d.idoc.InternalChecks[name]
	return exists
}
