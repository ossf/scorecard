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

// Package checks contains util fns for reading input YAML file.
package checks

import (

	// Used to embed `checks.yaml` file.
	_ "embed"
	"fmt"

	"gopkg.in/yaml.v2"
)

//go:embed checks.yaml
var checksYAML []byte

// Check defines expected check definition in checks.yaml.
type Check struct {
	Risk        string   `yaml:"-"`
	Short       string   `yaml:"short"`
	Description string   `yaml:"description"`
	Remediation []string `yaml:"remediation"`
}

// Doc maps to checks.yaml file.
type Doc struct {
	Checks map[string]Check
}

// Read parses `checks.yaml` file and returns a `Doc` struct.
func Read() (Doc, error) {
	var m Doc
	if err := yaml.Unmarshal(checksYAML, &m); err != nil {
		return Doc{}, fmt.Errorf("yaml.Unmarshal: %w", err)
	}
	return m, nil
}
