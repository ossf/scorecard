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

package raw

import (
	"strings"
	"testing"
)

func TestDotNetCliAddFloatingVersion(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		version  string
		unpinned bool
	}{
		{version: "13.0.1", unpinned: false},
		{version: "[13.0.1]", unpinned: false},
		{version: "13.*", unpinned: true},
		{version: "*", unpinned: true},
		{version: "'[13.0,)'", unpinned: true},
		{version: "(1.0,2.0]", unpinned: true},
	} {
		t.Run(tt.version, func(t *testing.T) {
			t.Parallel()

			cmd := strings.Fields("dotnet add package Newtonsoft.Json -v " + tt.version)
			if got := isUnpinnedDotNetCliAdd(cmd); got != tt.unpinned {
				t.Errorf("isUnpinnedDotNetCliAdd(%q) = %v, want %v", cmd, got, tt.unpinned)
			}
		})
	}
}
