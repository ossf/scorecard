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

package v20260828

import (
	"strings"
	"testing"

	"github.com/ossf/scorecard/v5/internal/probes"
	// Register every probe, so that the test sees the same registry as the
	// CLI.
	_ "github.com/ossf/scorecard/v5/probes"
)

func TestRequirements(t *testing.T) {
	t.Parallel()
	rel := Release()

	// Counts from the release tag: 65 requirements, 25 of them at
	// maturity-1.
	if got := len(rel.Requirements); got != 65 {
		t.Errorf("release has %d requirements, want 65", got)
	}
	perProfile := map[string]int{}
	seen := map[string]bool{}
	for i := range rel.Requirements {
		req := &rel.Requirements[i]
		if seen[req.ID] {
			t.Errorf("duplicate requirement %s", req.ID)
		}
		seen[req.ID] = true
		if req.Control == "" || !strings.HasPrefix(req.ID, req.Control+".") {
			t.Errorf("%s has control %q", req.ID, req.Control)
		}
		if len(req.Applicability) == 0 {
			t.Errorf("%s has no applicability", req.ID)
		}
		for _, p := range req.Applicability {
			if !rel.HasProfile(p) {
				t.Errorf("%s names undefined profile %q", req.ID, p)
			}
			perProfile[p]++
		}
	}
	want := map[string]int{maturity1: 25, maturity2: 42, maturity3: 63}
	for p, n := range want {
		if perProfile[p] != n {
			t.Errorf("%s has %d requirements, want %d", p, perProfile[p], n)
		}
	}
}

// TestMappingsNameRegisteredProbes fails if a mapping names a probe that
// is not registered, for example after a probe is renamed.
func TestMappingsNameRegisteredProbes(t *testing.T) {
	t.Parallel()
	rel := Release()
	for i := range rel.Mappings {
		m := &rel.Mappings[i]
		if _, ok := rel.Requirement(m.Requirement); !ok {
			t.Errorf("mapping for %s: no such requirement in %s", m.Requirement, rel.Version)
		}
		if len(m.Probes) == 0 || len(m.Steps) == 0 || m.Evaluate == nil {
			t.Errorf("mapping for %s is incomplete", m.Requirement)
		}
		for _, p := range m.Probes {
			if _, err := probes.Get(p); err != nil {
				t.Errorf("mapping for %s names probe %q: %v", m.Requirement, p, err)
			}
		}
	}
}
