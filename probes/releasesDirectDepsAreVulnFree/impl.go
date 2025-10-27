// Copyright 2025 OpenSSF Scorecard Authors
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

package releasesDirectDepsAreVulnFree

import (
	"fmt"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/checknames"
	"github.com/ossf/scorecard/v5/internal/probes"
)

func init() {
	probes.MustRegister(Probe, Run, []checknames.CheckName{checknames.ReleasesDirectDepsVulnFree})
}

// Probe is the stable ID used in def.yaml and attached to each Finding.
const Probe = "releasesDirectDepsAreVulnFree"

// Run consumes checker.RawResults populated by the raw collector.
// Returns a slice with one finding per release.
func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	data := raw.ReleaseDirectDepsVulnsResults
	var out []finding.Finding

	for _, r := range data.Releases {
		// Case 1: release is clean (no vulnerable direct deps).
		if len(r.Findings) == 0 {
			out = append(out, finding.Finding{
				Probe:   Probe,
				Outcome: finding.OutcomeTrue,
				Message: fmt.Sprintf("release %s has no known vulnerabilities in direct dependencies", r.Tag),
				Location: &finding.Location{
					// Use a synthetic path under "releases/" so UIs can group these.
					Path: fmt.Sprintf("releases/%s", r.Tag),
				},
			})
			continue
		}

		// Case 2: release has at least one vulnerable direct dep.
		// Summarize the first one in the message; more details are in raw results.
		f0 := r.Findings[0]
		msg := fmt.Sprintf(
			"release %s has vulnerable direct dependency %s@%s (e.g., %v)",
			r.Tag, f0.Name, f0.Version, f0.OSVIDs,
		)
		out = append(out, finding.Finding{
			Probe:   Probe,
			Outcome: finding.OutcomeFalse,
			Message: msg,
			Location: &finding.Location{
				// Point to the manifest path that declared this dependency.
				Path: f0.ManifestPath,
			},
		})
	}

	return out, "checked recent releases for vulnerable direct dependencies", nil
}
