// Copyright 2026 OpenSSF Scorecard Authors
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

//go:build gofuzz

package finding

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

// fuzzProbeDefinitionDir is where the OSS-Fuzz harness writes probe definitions.
const fuzzProbeDefinitionDir = "/tmp/probedefinitions"

// readProbeDefinition returns the def.yml written by the OSS-Fuzz harness.
// This file is only compiled into fuzz builds.
func readProbeDefinition(_ embed.FS, probeID string) ([]byte, error) {
	content, err := os.ReadFile(filepath.Join(fuzzProbeDefinitionDir, probeID, "def.yml"))
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return content, nil
}
