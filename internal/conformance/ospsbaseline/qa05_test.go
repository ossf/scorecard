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
	"github.com/ossf/scorecard/v5/probes/hasBinaryArtifacts"
	"github.com/ossf/scorecard/v5/probes/hasUnverifiedBinaryArtifacts"
)

const (
	binary     = hasBinaryArtifacts.Probe
	unverified = hasUnverifiedBinaryArtifacts.Probe
	wrapperJar = "gradle/wrapper/gradle-wrapper.jar"
)

func TestQA0501(t *testing.T) {
	t.Parallel()
	runMappingCases(t, QA0501Mapping().Evaluate, []mappingCase{
		{
			name: "no binaries",
			findings: []finding.Finding{
				repoFinding(binary, finding.OutcomeFalse),
				repoFinding(unverified, finding.OutcomeFalse),
			},
			want:         conformance.Pass,
			wantEvidence: 2,
		},
		{
			name: "unverified binary",
			findings: []finding.Finding{
				fileFinding(binary, finding.OutcomeTrue, "tool.exe"),
				fileFinding(unverified, finding.OutcomeTrue, "tool.exe"),
			},
			want:         conformance.Fail,
			wantText:     "tool.exe",
			wantEvidence: 2,
		},
		{
			name: "only validated Gradle wrappers",
			findings: []finding.Finding{
				fileFinding(binary, finding.OutcomeTrue, wrapperJar),
				repoFinding(unverified, finding.OutcomeFalse),
			},
			want:         conformance.NeedsReview,
			wantText:     wrapperJar,
			wantEvidence: 2,
		},
		{
			name: "binary probe error",
			findings: []finding.Finding{
				repoFinding(binary, finding.OutcomeError),
				repoFinding(unverified, finding.OutcomeError),
			},
			want:         conformance.Unknown,
			wantEvidence: 2,
		},
		{
			name: "binaries found, unverified finding missing",
			findings: []finding.Finding{
				fileFinding(binary, finding.OutcomeTrue, "tool.exe"),
			},
			want:         conformance.Unknown,
			wantEvidence: 1,
		},
	})
}

func TestQA0502(t *testing.T) {
	t.Parallel()
	runMappingCases(t, QA0502Mapping().Evaluate, []mappingCase{
		{
			name: "no binaries",
			findings: []finding.Finding{
				repoFinding(binary, finding.OutcomeFalse),
				repoFinding(unverified, finding.OutcomeFalse),
			},
			want:         conformance.Pass,
			wantText:     "No binary files",
			wantEvidence: 2,
		},
		{
			name: "only validated Gradle wrappers",
			findings: []finding.Finding{
				fileFinding(binary, finding.OutcomeTrue, wrapperJar),
				repoFinding(unverified, finding.OutcomeFalse),
			},
			want:         conformance.Pass,
			wantText:     wrapperJar,
			wantEvidence: 2,
		},
		{
			name: "unverified binary",
			findings: []finding.Finding{
				fileFinding(binary, finding.OutcomeTrue, "lib/vendor.jar"),
				fileFinding(unverified, finding.OutcomeTrue, "lib/vendor.jar"),
			},
			want:         conformance.Fail,
			wantText:     "lib/vendor.jar",
			wantEvidence: 2,
		},
		{
			name:     "no findings",
			want:     conformance.Unknown,
			wantText: unverified,
		},
	})
}
