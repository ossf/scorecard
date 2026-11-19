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

package probes_test

import (
	"testing"

	"github.com/ossf/scorecard/v5/checks"
	"github.com/ossf/scorecard/v5/internal/probes"
)

// TestRequiredRawDataIsRegistered requires every check a registered probe
// needs raw data from to be a registered check, so naming any registered
// probe finds a collector. Importing checks registers every probe.
func TestRequiredRawDataIsRegistered(t *testing.T) {
	t.Parallel()
	registered := probes.Registered()
	if _, ok := registered["hasPermissiveLicense"]; !ok {
		t.Fatal("hasPermissiveLicense is not registered; the probe packages did not load")
	}
	all := checks.GetAllWithExperimental()
	for name, p := range registered {
		for _, check := range p.RequiredRawData {
			if _, ok := all[check]; !ok {
				t.Errorf("probe %q needs raw data from %q, which is not a registered check", name, check)
			}
		}
	}
}
