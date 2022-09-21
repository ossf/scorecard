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

// Package internal contains internal functions for reading input YAML file.
package internal

import (
	// Used to embed `checks.yaml` file.
	_ "embed"
	"fmt"

	"gopkg.in/yaml.v2"
)

//go:embed checks.yaml
var checksYAML []byte

// Check stores a check's information.
//
//nolint:govet
type Check struct {
	Risk        string   `yaml:"risk"`
	Short       string   `yaml:"short"`
	Description string   `yaml:"description"`
	Tags        string   `yaml:"tags"`
	Repos       string   `yaml:"repos"`
	Remediation []string `yaml:"remediation"`
	Name        string   `yaml:"-"`
	URL         string   `yaml:"-"`
}

// Doc stores the documentation for all checks.
type Doc struct {
	InternalChecks map[string]Check `yaml:"checks"`
}

// ReadDoc reads documentation from the `checks.yaml` file.
func ReadDoc() (Doc, error) {
	var m Doc
	if err := yaml.Unmarshal(checksYAML, &m); err != nil {
		return Doc{}, fmt.Errorf("yaml.Unmarshal: %w", err)
	}
	return m, nil
}
