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

package ospsbaseline

import (
	"testing"

	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/conformance"
	"github.com/ossf/scorecard/v5/probes/securityPolicyContainsLinks"
	"github.com/ossf/scorecard/v5/probes/securityPolicyPresent"
)

const (
	policy = securityPolicyPresent.Probe
	links  = securityPolicyContainsLinks.Probe
)

func TestVM0201(t *testing.T) {
	t.Parallel()
	runMappingCases(t, VM0201Mapping().Evaluate, []mappingCase{
		{
			name: "policy with a link",
			findings: []finding.Finding{
				fileFinding(policy, finding.OutcomeTrue, "SECURITY.md"),
				fileFinding(links, finding.OutcomeTrue, "SECURITY.md"),
			},
			want:         conformance.NeedsReview,
			wantText:     "may or may not be a contact",
			wantEvidence: 2,
		},
		{
			name: "policy without a link",
			findings: []finding.Finding{
				fileFinding(policy, finding.OutcomeTrue, ".github/SECURITY.md"),
				fileFinding(links, finding.OutcomeFalse, ".github/SECURITY.md"),
			},
			want:         conformance.NeedsReview,
			wantText:     "No URL or email address",
			wantEvidence: 2,
		},
		{
			name: "no policy",
			findings: []finding.Finding{
				repoFinding(policy, finding.OutcomeFalse),
				repoFinding(links, finding.OutcomeFalse),
			},
			want:         conformance.Fail,
			wantText:     "organization's .github repository",
			wantEvidence: 2,
		},
		{
			name:     "no findings",
			want:     conformance.Unknown,
			wantText: policy,
		},
	})
}
