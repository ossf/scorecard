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

package evaluation

import (
	"testing"

	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes/blocksDeleteOnTags"
	"github.com/ossf/scorecard/v5/probes/blocksForcePushOnTags"
	"github.com/ossf/scorecard/v5/probes/blocksUpdateOnTags"
	"github.com/ossf/scorecard/v5/probes/requiresSignedTags"
	"github.com/ossf/scorecard/v5/probes/restrictsTagCreation"
	"github.com/ossf/scorecard/v5/probes/tagProtectionAppliesToAdmins"
	"github.com/ossf/scorecard/v5/probes/tagsAreProtected"
	scut "github.com/ossf/scorecard/v5/utests"
)

func TestTagProtection(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		findings []finding.Finding
		result   scut.TestReturn
	}{
		{
			name: "No tags - inconclusive",
			findings: []finding.Finding{
				{
					Probe:   tagsAreProtected.Probe,
					Outcome: finding.OutcomeNotApplicable,
				},
				{
					Probe:   blocksDeleteOnTags.Probe,
					Outcome: finding.OutcomeNotApplicable,
				},
				{
					Probe:   blocksForcePushOnTags.Probe,
					Outcome: finding.OutcomeNotApplicable,
				},
				{
					Probe:   blocksUpdateOnTags.Probe,
					Outcome: finding.OutcomeNotApplicable,
				},
				{
					Probe:   tagProtectionAppliesToAdmins.Probe,
					Outcome: finding.OutcomeNotApplicable,
				},
				{
					Probe:   restrictsTagCreation.Probe,
					Outcome: finding.OutcomeNotApplicable,
				},
				{
					Probe:   requiresSignedTags.Probe,
					Outcome: finding.OutcomeNotApplicable,
				},
			},
			result: scut.TestReturn{
				Score:         -1,
				NumberOfInfo:  0,
				NumberOfWarn:  0,
				NumberOfDebug: 0,
			},
		},
		{
			name: "All protections enabled - max score",
			findings: []finding.Finding{
				{
					Probe:   tagsAreProtected.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksDeleteOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksForcePushOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksUpdateOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   tagProtectionAppliesToAdmins.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   restrictsTagCreation.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   requiresSignedTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
			},
			result: scut.TestReturn{
				Score:        10,
				NumberOfInfo: 7,
				NumberOfWarn: 0,
			},
		},
		{
			name: "No protections - zero score",
			findings: []finding.Finding{
				{
					Probe:   tagsAreProtected.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   blocksDeleteOnTags.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   blocksForcePushOnTags.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   blocksUpdateOnTags.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   tagProtectionAppliesToAdmins.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   restrictsTagCreation.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   requiresSignedTags.Probe,
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score:         0,
				NumberOfInfo:  0,
				NumberOfWarn:  1, // Only one warning: not all tags protected
				NumberOfDebug: 1, // 1 debug: tag lacks protection
			},
		},
		{
			name: "Only tags protected - 3 points",
			findings: []finding.Finding{
				{
					Probe:   tagsAreProtected.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksDeleteOnTags.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   blocksForcePushOnTags.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   blocksUpdateOnTags.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   tagProtectionAppliesToAdmins.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   restrictsTagCreation.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   requiresSignedTags.Probe,
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score:         3,
				NumberOfInfo:  1, // protected only
				NumberOfWarn:  2, // delete and force push warnings
				NumberOfDebug: 2, // 2 debug: tags lack protections
			},
		},
		{
			name: "Basic protections - 7 points",
			findings: []finding.Finding{
				{
					Probe:   tagsAreProtected.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksDeleteOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksForcePushOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksUpdateOnTags.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   tagProtectionAppliesToAdmins.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   restrictsTagCreation.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   requiresSignedTags.Probe,
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score:         6, // Tier 2 complete
				NumberOfInfo:  3, // protected, delete, force push
				NumberOfWarn:  1, // update warning
				NumberOfDebug: 1, // 1 debug: tag lacks update protection
			},
		},
		{
			name: "Multiple tags with mixed results",
			findings: []finding.Finding{
				{
					Probe:   tagsAreProtected.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   tagsAreProtected.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   blocksDeleteOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksDeleteOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksForcePushOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksForcePushOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksUpdateOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksUpdateOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   tagProtectionAppliesToAdmins.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   tagProtectionAppliesToAdmins.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   restrictsTagCreation.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   restrictsTagCreation.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   requiresSignedTags.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   requiresSignedTags.Probe,
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score:         0, // Not all tags protected = 0
				NumberOfInfo:  0, // Early return, no info messages
				NumberOfWarn:  1, // Only one warning: not all tags protected
				NumberOfDebug: 1, // 1 debug: tag lacks protection
			},
		},
		{
			name: "All tags fully protected",
			findings: []finding.Finding{
				{
					Probe:   tagsAreProtected.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   tagsAreProtected.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksDeleteOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksDeleteOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksForcePushOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksForcePushOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksUpdateOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksUpdateOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   tagProtectionAppliesToAdmins.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   tagProtectionAppliesToAdmins.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   restrictsTagCreation.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   restrictsTagCreation.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   requiresSignedTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   requiresSignedTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
			},
			result: scut.TestReturn{
				Score:        10,
				NumberOfInfo: 7,
				NumberOfWarn: 0,
			},
		},
		{
			name: "Score 5 - protected + delete blocked only",
			findings: []finding.Finding{
				{
					Probe:   tagsAreProtected.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksDeleteOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksForcePushOnTags.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   blocksUpdateOnTags.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   tagProtectionAppliesToAdmins.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   restrictsTagCreation.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   requiresSignedTags.Probe,
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score:         3, // Tier 2 incomplete - missing force push
				NumberOfInfo:  2, // protected, delete
				NumberOfWarn:  1, // force push warning
				NumberOfDebug: 1, // 1 debug: tag lacks force-push protection
			},
		},
		{
			name: "Score 6 - protected + delete and force push blocked (Tier 2)",
			findings: []finding.Finding{
				{
					Probe:   tagsAreProtected.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksDeleteOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksForcePushOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksUpdateOnTags.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   tagProtectionAppliesToAdmins.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   restrictsTagCreation.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   requiresSignedTags.Probe,
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score:         6, // Tier 2 complete
				NumberOfInfo:  3, // protected, delete, force push
				NumberOfWarn:  1, // update warning
				NumberOfDebug: 1, // 1 debug: tag lacks update protection
			},
		},
		{
			name: "Score 3 - protected + restrict creation only (Tier 1)",
			findings: []finding.Finding{
				{
					Probe:   tagsAreProtected.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksDeleteOnTags.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   blocksForcePushOnTags.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   blocksUpdateOnTags.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   tagProtectionAppliesToAdmins.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   restrictsTagCreation.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   requiresSignedTags.Probe,
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score:         3, // Tier 2 incomplete - missing delete and force push
				NumberOfInfo:  1, // protected only
				NumberOfWarn:  2, // delete and force push warnings
				NumberOfDebug: 2, // 2 debug: tags lack protections
			},
		},
		{
			name: "Score 10 - all protections (Tier 4 complete)",
			findings: []finding.Finding{
				{
					Probe:   tagsAreProtected.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksDeleteOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksForcePushOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksUpdateOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   tagProtectionAppliesToAdmins.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   restrictsTagCreation.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   requiresSignedTags.Probe,
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score:         10, // Full score (signed tags don't affect score)
				NumberOfInfo:  6,
				NumberOfWarn:  0,
				NumberOfDebug: 1,
			},
		},
		{
			name: "Partial protections across multiple tags",
			findings: []finding.Finding{
				{
					Probe:   tagsAreProtected.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   tagsAreProtected.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   tagsAreProtected.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksDeleteOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksDeleteOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksDeleteOnTags.Probe,
					Outcome: finding.OutcomeFalse, // One tag doesn't block delete
				},
				{
					Probe:   blocksForcePushOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksForcePushOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksForcePushOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksUpdateOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksUpdateOnTags.Probe,
					Outcome: finding.OutcomeFalse, // One tag allows updates
				},
				{
					Probe:   blocksUpdateOnTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   tagProtectionAppliesToAdmins.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   tagProtectionAppliesToAdmins.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   tagProtectionAppliesToAdmins.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   restrictsTagCreation.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   restrictsTagCreation.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   restrictsTagCreation.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   requiresSignedTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   requiresSignedTags.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   requiresSignedTags.Probe,
					Outcome: finding.OutcomeTrue,
				},
			},
			result: scut.TestReturn{
				Score:         3, // Tier 2 incomplete - missing delete on all tags
				NumberOfInfo:  2, // protected, force push
				NumberOfWarn:  1, // delete warning
				NumberOfDebug: 1, // 1 debug: tag lacks delete protection
			},
		},
		{
			name: "Single tag with minimal protection",
			findings: []finding.Finding{
				{
					Probe:   tagsAreProtected.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   blocksDeleteOnTags.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   blocksForcePushOnTags.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   blocksUpdateOnTags.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   tagProtectionAppliesToAdmins.Probe,
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   restrictsTagCreation.Probe,
					Outcome: finding.OutcomeTrue,
				},
				{
					Probe:   requiresSignedTags.Probe,
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score:         3, // Tier 2 incomplete - missing delete and force push
				NumberOfInfo:  1, // protected only
				NumberOfWarn:  2, // delete and force push warnings
				NumberOfDebug: 2, // 2 debug: tags lack protections
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			got := TagProtection("Tag-Protection", tt.findings, &dl)
			scut.ValidateTestReturn(t, tt.name, &tt.result, &got, &dl)
		})
	}
}

func TestGitLabTagProtectionScoring(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		findings []finding.Finding
		result   scut.TestReturn
	}{
		{
			name: "GitLab maximum score - both components perfect (10 points)",
			findings: []finding.Finding{
				{
					Probe:   "tagsCannotDuplicateBranchNames",
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						"branchName":      "main",
						"protectionLevel": "strongest",
					},
				},
				{
					Probe:   "gitlabReleaseTagsAreProtected",
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						"tagName":         "v1.0.0",
						"protectionLevel": "strongest",
					},
				},
			},
			result: scut.TestReturn{
				Score:        10,
				NumberOfInfo: 2,
			},
		},
		{
			name: "GitLab Component 1 only - branches protected strongest (2 points)",
			findings: []finding.Finding{
				{
					Probe:   "tagsCannotDuplicateBranchNames",
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						"branchName":      "main",
						"protectionLevel": "strongest",
					},
				},
				{
					Probe:   "gitlabReleaseTagsAreProtected",
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score:        2,
				NumberOfInfo: 1,
				NumberOfWarn: 1,
			},
		},
		{
			name: "GitLab Component 1 strong - 1 point",
			findings: []finding.Finding{
				{
					Probe:   "tagsCannotDuplicateBranchNames",
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						"branchName":      "main",
						"protectionLevel": "strong",
					},
				},
				{
					Probe:   "gitlabReleaseTagsAreProtected",
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score:        1,
				NumberOfInfo: 1,
				NumberOfWarn: 1,
			},
		},
		{
			name: "GitLab Component 2 only - release tags strongest (8 points)",
			findings: []finding.Finding{
				{
					Probe:   "tagsCannotDuplicateBranchNames",
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   "gitlabReleaseTagsAreProtected",
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						"tagName":         "v1.0.0",
						"protectionLevel": "strongest",
					},
				},
			},
			result: scut.TestReturn{
				Score:        8,
				NumberOfInfo: 1,
				NumberOfWarn: 1,
			},
		},
		{
			name: "GitLab Component 2 strong - 4 points",
			findings: []finding.Finding{
				{
					Probe:   "tagsCannotDuplicateBranchNames",
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   "gitlabReleaseTagsAreProtected",
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						"tagName":         "v1.0.0",
						"protectionLevel": "strong",
					},
				},
			},
			result: scut.TestReturn{
				Score:        4,
				NumberOfInfo: 1,
				NumberOfWarn: 1,
			},
		},
		{
			name: "GitLab both strong - 5 points (1 + 4)",
			findings: []finding.Finding{
				{
					Probe:   "tagsCannotDuplicateBranchNames",
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						"branchName":      "main",
						"protectionLevel": "strong",
					},
				},
				{
					Probe:   "gitlabReleaseTagsAreProtected",
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						"tagName":         "v1.0.0",
						"protectionLevel": "strong",
					},
				},
			},
			result: scut.TestReturn{
				Score:        5,
				NumberOfInfo: 2,
			},
		},
		{
			name: "GitLab branch strongest + release strong - 6 points (2 + 4)",
			findings: []finding.Finding{
				{
					Probe:   "tagsCannotDuplicateBranchNames",
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						"branchName":      "main",
						"protectionLevel": "strongest",
					},
				},
				{
					Probe:   "gitlabReleaseTagsAreProtected",
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						"tagName":         "v1.0.0",
						"protectionLevel": "strong",
					},
				},
			},
			result: scut.TestReturn{
				Score:        6,
				NumberOfInfo: 2,
			},
		},
		{
			name: "GitLab branch strong + release strongest - 9 points (1 + 8)",
			findings: []finding.Finding{
				{
					Probe:   "tagsCannotDuplicateBranchNames",
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						"branchName":      "main",
						"protectionLevel": "strong",
					},
				},
				{
					Probe:   "gitlabReleaseTagsAreProtected",
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						"tagName":         "v1.0.0",
						"protectionLevel": "strongest",
					},
				},
			},
			result: scut.TestReturn{
				Score:        9,
				NumberOfInfo: 2,
			},
		},
		{
			name: "GitLab both components fail - 0 points",
			findings: []finding.Finding{
				{
					Probe:   "tagsCannotDuplicateBranchNames",
					Outcome: finding.OutcomeFalse,
				},
				{
					Probe:   "gitlabReleaseTagsAreProtected",
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score:        0,
				NumberOfWarn: 2,
			},
		},
		{
			name: "GitLab mixed branch protection - weakest wins",
			findings: []finding.Finding{
				{
					Probe:   "tagsCannotDuplicateBranchNames",
					Outcome: finding.OutcomeTrue,
					Values: map[string]string{
						"branchName":      "main",
						"protectionLevel": "strongest",
					},
				},
				{
					Probe:   "tagsCannotDuplicateBranchNames",
					Outcome: finding.OutcomeFalse,
					Values: map[string]string{
						"branchName":      "develop",
						"protectionLevel": "none",
					},
				},
				{
					Probe:   "gitlabReleaseTagsAreProtected",
					Outcome: finding.OutcomeFalse,
				},
			},
			result: scut.TestReturn{
				Score:        0,
				NumberOfInfo: 1,
				NumberOfWarn: 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dl := scut.TestDetailLogger{}
			got := TagProtection("Tag-Protection", tt.findings, &dl)
			scut.ValidateTestReturn(t, tt.name, &tt.result, &got, &dl)
		})
	}
}
