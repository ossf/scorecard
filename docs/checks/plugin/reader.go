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

// Package plugin contains plugin functions for reading input YAML file.
package plugin

import (
	// Used to embed `checks.yaml` file.
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"plugin"
	"strings"
	"sync"

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
	PluginChecks map[string]Check `yaml:"checks"`
}

const CheckPluginLocationEnvironmentVariable = "SCORECARD_DYNAMIC_CHECKS"

var pluginMutex sync.Mutex

// ReadDoc reads documentation from the `checks.yaml` file.
func ReadDoc() (Doc, error) {
	var m Doc
	var c Check

	pluginsDir, ok := os.LookupEnv(CheckPluginLocationEnvironmentVariable)
	if !ok {
		return Doc{}, nil
	}

	files, err := ioutil.ReadDir(pluginsDir)
	if err != nil {
		return Doc{}, fmt.Errorf("docs.checks.plugin.readdir(%s): %w", pluginsDir, err)
	}

	// Avoid `recursive call during initialization - linker skew` via mutex
	pluginMutex.Lock()
	defer pluginMutex.Unlock()

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".so") {
			continue
		}

		p, err := plugin.Open(file.Name())
		if err != nil {
			return Doc{}, fmt.Errorf("docs.checks.plugin.open: %w", err)
		}

		checkYAML, err := p.Lookup("CheckYAML")
		if err != nil {
			return Doc{}, fmt.Errorf("docs.checks.plugin.lookup.CheckYAML: %w", err)
		}

		c = Check{}
		if err := yaml.Unmarshal(checkYAML.([]byte), &c); err != nil {
			return Doc{}, fmt.Errorf("yaml.Unmarshal: %w", err)
		}
	}

	return m, nil
}
