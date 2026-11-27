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
	"github.com/ossf/scorecard/v5/probes/hasDangerousWorkflowScriptInjection"
	"github.com/ossf/scorecard/v5/probes/hasDangerousWorkflowUntrustedCheckout"
)

const (
	injection = hasDangerousWorkflowScriptInjection.Probe
	checkout  = hasDangerousWorkflowUntrustedCheckout.Probe
	ciYAML    = ".github/workflows/ci.yml"
)

func TestBR0101(t *testing.T) {
	t.Parallel()
	runMappingCases(t, BR0101Mapping().Evaluate, []mappingCase{
		{
			name: "script injection",
			findings: []finding.Finding{
				fileFinding(injection, finding.OutcomeTrue, ciYAML),
				fileFinding(injection, finding.OutcomeTrue, ".github/workflows/release.yml"),
			},
			want:         conformance.Fail,
			wantText:     ciYAML,
			wantEvidence: 2,
		},
		{
			name:         "no known pattern",
			findings:     []finding.Finding{repoFinding(injection, finding.OutcomeFalse)},
			want:         conformance.NeedsReview,
			wantText:     "Only known patterns",
			wantEvidence: 1,
		},
		{
			name:         "no workflows",
			findings:     []finding.Finding{repoFinding(injection, finding.OutcomeNotApplicable)},
			want:         conformance.Unknown,
			wantText:     "no GitHub Actions workflows",
			wantEvidence: 1,
		},
		{
			name:         "probe error",
			findings:     []finding.Finding{repoFinding(injection, finding.OutcomeError)},
			want:         conformance.Unknown,
			wantText:     "Error",
			wantEvidence: 1,
		},
		{
			name:     "no findings",
			want:     conformance.Unknown,
			wantText: "returned no findings",
		},
		{
			name: "other probes ignored",
			findings: []finding.Finding{
				fileFinding(checkout, finding.OutcomeTrue, ciYAML),
				repoFinding(injection, finding.OutcomeFalse),
			},
			want:         conformance.NeedsReview,
			wantEvidence: 1,
		},
	})
}

func TestBR0103(t *testing.T) {
	t.Parallel()
	runMappingCases(t, BR0103Mapping().Evaluate, []mappingCase{
		{
			name:         "untrusted checkout",
			findings:     []finding.Finding{fileFinding(checkout, finding.OutcomeTrue, ciYAML)},
			want:         conformance.Fail,
			wantText:     "Conditions such as a required label, and job permissions, were not read.",
			wantEvidence: 1,
		},
		{
			name:         "no untrusted checkout",
			findings:     []finding.Finding{repoFinding(checkout, finding.OutcomeFalse)},
			want:         conformance.NeedsReview,
			wantText:     "Only those two triggers",
			wantEvidence: 1,
		},
		{
			name:         "no workflows",
			findings:     []finding.Finding{repoFinding(checkout, finding.OutcomeNotApplicable)},
			want:         conformance.Unknown,
			wantText:     "no GitHub Actions workflows",
			wantEvidence: 1,
		},
		{
			name:     "no findings",
			want:     conformance.Unknown,
			wantText: checkout,
		},
	})
}
