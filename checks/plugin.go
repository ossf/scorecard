// Copyright 2024 OpenSSF Scorecard Authors
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

package checks

import (
	"io/ioutil"
	"log"
	"os"
	"plugin"
	"strings"
	"sync"
)

const CheckPluginLocationEnvironmentVariable = "SCORECARD_DYNAMIC_CHECKS"

var pluginMutex sync.Mutex

//nolint:gochecknoinits
func init() {
	pluginsDir, ok := os.LookupEnv(CheckPluginLocationEnvironmentVariable)
	if !ok {
		return
	}

	files, err := ioutil.ReadDir(pluginsDir)
	if err != nil {
		log.Fatal(err)
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
			log.Fatal(err)
		}

		registerCheck, err := p.Lookup("RegisterCheck")
		if err != nil {
			log.Fatal(err)
		}

		err = registerCheck.(func() error)()
		if err != nil {
			log.Fatal(err)
		}
	}
}
