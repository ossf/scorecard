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
	"github.com/ossf/scorecard/v5/probes/hasFSFOrOSIApprovedLicense"
	"github.com/ossf/scorecard/v5/probes/hasLicenseFile"
)

const (
	approved    = hasFSFOrOSIApprovedLicense.Probe
	licenseFile = hasLicenseFile.Probe
)

func TestLE0201(t *testing.T) {
	t.Parallel()
	runMappingCases(t, LE0201Mapping().Evaluate, []mappingCase{
		{
			name:         "approved license",
			findings:     []finding.Finding{repoFinding(approved, finding.OutcomeTrue)},
			want:         conformance.Pass,
			wantEvidence: 1,
		},
		{
			name:         "license not approved",
			findings:     []finding.Finding{repoFinding(approved, finding.OutcomeFalse)},
			want:         conformance.Fail,
			wantEvidence: 1,
		},
		{
			name:         "no license file",
			findings:     []finding.Finding{repoFinding(approved, finding.OutcomeNotApplicable)},
			want:         conformance.Unknown,
			wantText:     "OSPS-LE-03.01",
			wantEvidence: 1,
		},
		{
			name:     "no findings",
			want:     conformance.Unknown,
			wantText: approved,
		},
	})
}

func TestLE0202(t *testing.T) {
	t.Parallel()
	runMappingCases(t, LE0202Mapping().Evaluate, []mappingCase{
		{
			name:         "approved license on the default branch",
			findings:     []finding.Finding{repoFinding(approved, finding.OutcomeTrue)},
			want:         conformance.NeedsReview,
			wantText:     "not from the released assets",
			wantEvidence: 1,
		},
		{
			name:         "license not approved",
			findings:     []finding.Finding{repoFinding(approved, finding.OutcomeFalse)},
			want:         conformance.Fail,
			wantEvidence: 1,
		},
		{
			name:         "no license file",
			findings:     []finding.Finding{repoFinding(approved, finding.OutcomeNotApplicable)},
			want:         conformance.Unknown,
			wantEvidence: 1,
		},
		{
			name:         "probe error",
			findings:     []finding.Finding{repoFinding(approved, finding.OutcomeError)},
			want:         conformance.Unknown,
			wantText:     "Error",
			wantEvidence: 1,
		},
	})
}

func TestLE0301(t *testing.T) {
	t.Parallel()
	runMappingCases(t, LE0301Mapping().Evaluate, []mappingCase{
		{
			name:         "LICENSE file",
			findings:     []finding.Finding{fileFinding(licenseFile, finding.OutcomeTrue, "LICENSE")},
			want:         conformance.Pass,
			wantText:     "LICENSE",
			wantEvidence: 1,
		},
		{
			name:         "LICENSES directory",
			findings:     []finding.Finding{fileFinding(licenseFile, finding.OutcomeTrue, "LICENSES/MIT.txt")},
			want:         conformance.Pass,
			wantEvidence: 1,
		},
		{
			name: "COPYRIGHT and COPYING",
			findings: []finding.Finding{
				fileFinding(licenseFile, finding.OutcomeTrue, "COPYRIGHT"),
				fileFinding(licenseFile, finding.OutcomeTrue, "COPYING"),
			},
			want:         conformance.Pass,
			wantText:     "COPYING",
			wantEvidence: 2,
		},
		{
			name:         "only a COPYRIGHT file",
			findings:     []finding.Finding{fileFinding(licenseFile, finding.OutcomeTrue, "COPYRIGHT")},
			want:         conformance.NeedsReview,
			wantText:     "COPYRIGHT",
			wantEvidence: 1,
		},
		{
			name:         "no license file",
			findings:     []finding.Finding{repoFinding(licenseFile, finding.OutcomeFalse)},
			want:         conformance.Fail,
			wantEvidence: 1,
		},
		{
			name:     "no findings",
			want:     conformance.Unknown,
			wantText: licenseFile,
		},
	})
}

func TestLE0302(t *testing.T) {
	t.Parallel()
	runMappingCases(t, LE0302Mapping().Evaluate, []mappingCase{
		{
			name:         "license file on the default branch",
			findings:     []finding.Finding{fileFinding(licenseFile, finding.OutcomeTrue, "LICENSE")},
			want:         conformance.NeedsReview,
			wantText:     "release assets were not read",
			wantEvidence: 1,
		},
		{
			name:         "no license file",
			findings:     []finding.Finding{repoFinding(licenseFile, finding.OutcomeFalse)},
			want:         conformance.Fail,
			wantEvidence: 1,
		},
		{
			name:     "no findings",
			want:     conformance.Unknown,
			wantText: licenseFile,
		},
	})
}
