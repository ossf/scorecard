// Copyright 2021 Security Scorecard Authors
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

package e2e

import (
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	"github.com/ossf/scorecard/v4/attestor/command"
	"github.com/ossf/scorecard/v4/attestor/policy"
)

var _ = Describe("E2E TEST PAT: scorecard-attestor policy", func() {
	Context("E2E TEST:Validating scorecard attestation policy", func() {
		It("Should attest to repos based on policy", func() {
			tt := []struct {
				name     string
				repoURL  string
				commit   string
				policy   policy.AttestationPolicy
				expected policy.PolicyResult
			}{
				{
					name:    "test good repo",
					repoURL: "https://github.com/ossf-tests/scorecard-binauthz-test-good",
					policy: policy.AttestationPolicy{
						PreventBinaryArtifacts:      true,
						PreventKnownVulnerabilities: true,
						PreventUnpinnedDependencies: true,
					},
					expected: policy.Pass,
				},
				{
					name:    "test bad repo with vulnerabilities prevented but no known vulnerabilities",
					repoURL: "https://github.com/ossf-tests/scorecard-binauthz-test-bad",
					policy: policy.AttestationPolicy{
						PreventKnownVulnerabilities: true,
					},
					expected: policy.Pass,
				},
				{
					name:    "test bad repo with ignored binary artifact",
					repoURL: "https://github.com/ossf-tests/scorecard-binauthz-test-bad",
					policy: policy.AttestationPolicy{
						PreventBinaryArtifacts:      true,
						AllowedBinaryArtifacts:      []string{"test-binary-artifact-*"},
						PreventKnownVulnerabilities: true,
					},
					expected: policy.Pass,
				},
				{
					name:    "test bad repo with ignored binary artifact",
					repoURL: "https://github.com/ossf-tests/scorecard-binauthz-test-bad",
					policy: policy.AttestationPolicy{
						PreventBinaryArtifacts:      true,
						PreventKnownVulnerabilities: true,
					},
					expected: policy.Fail,
				},
				{
					name:    "test bad repo with ignored dep by path",
					repoURL: "https://github.com/ossf-tests/scorecard-binauthz-test-bad",
					policy: policy.AttestationPolicy{
						PreventUnpinnedDependencies: true,
						AllowedUnpinnedDependencies: []policy.Dependency{{Filepath: "Dockerfile"}},
					},
					expected: policy.Pass,
				},
				{
					name:    "test bad repo without ignored dep",
					repoURL: "https://github.com/ossf-tests/scorecard-binauthz-test-bad",
					policy: policy.AttestationPolicy{
						PreventUnpinnedDependencies: true,
					},
					expected: policy.Fail,
				},
				{
					name:    "test bad repo with ignored dep by name",
					repoURL: "https://github.com/ossf-tests/scorecard-binauthz-test-bad",
					policy: policy.AttestationPolicy{
						PreventUnpinnedDependencies: true,
						AllowedUnpinnedDependencies: []policy.Dependency{{PackageName: "static-debian11"}, {PackageName: "golang"}},
					},
					expected: policy.Pass,
				},
				{
					name:    "test bad repo with everything ignored",
					repoURL: "https://github.com/ossf-tests/scorecard-binauthz-test-bad",
					policy: policy.AttestationPolicy{
						PreventBinaryArtifacts:      true,
						AllowedBinaryArtifacts:      []string{"test-binary-artifact-*"},
						PreventKnownVulnerabilities: true,
						PreventUnpinnedDependencies: true,
						AllowedUnpinnedDependencies: []policy.Dependency{{Filepath: "Dockerfile"}},
					},
					expected: policy.Pass,
				},
				{
					name:    "test repo with simple code review requirements",
					repoURL: "https://github.com/ossf/scorecard",
					commit:  "fa0592fab28aa92560f04e1ae8649dfff566ae2b",
					policy: policy.AttestationPolicy{
						EnsureCodeReviewed: true,
						CodeReviewRequirements: policy.CodeReviewRequirements{
							MinReviewers: 1,
						},
					},
					expected: policy.Pass,
				},
				{
					name:    "test code reviews required but repo doesn't have code reviews",
					repoURL: "https://github.com/ossf-tests/scorecard-binauthz-test-bad",
					policy: policy.AttestationPolicy{
						PreventBinaryArtifacts:      true,
						PreventKnownVulnerabilities: true,
						PreventUnpinnedDependencies: true,
						EnsureCodeReviewed:          true,
					},
					expected: policy.Fail,
				},
				{
					name:    "test code reviews required with min reviewers",
					repoURL: "https://github.com/ossf/scorecard",
					commit:  "fa0592fab28aa92560f04e1ae8649dfff566ae2b",
					policy: policy.AttestationPolicy{
						PreventBinaryArtifacts:      true,
						PreventKnownVulnerabilities: false,
						PreventUnpinnedDependencies: true,
						EnsureCodeReviewed:          true,
						CodeReviewRequirements: policy.CodeReviewRequirements{
							MinReviewers: 1,
						},
					},
					expected: policy.Pass,
				},
				{
					name:    "test code reviews required with min reviewers and required reviewers",
					repoURL: "https://github.com/ossf/scorecard",
					commit:  "fa0592fab28aa92560f04e1ae8649dfff566ae2b",
					policy: policy.AttestationPolicy{
						PreventBinaryArtifacts:      true,
						PreventKnownVulnerabilities: false,
						PreventUnpinnedDependencies: true,
						EnsureCodeReviewed:          true,
						CodeReviewRequirements: policy.CodeReviewRequirements{
							MinReviewers:      1,
							RequiredApprovers: []string{"spencerschrock", "laurentsimon", "naveensrinivasan", "azeemshaikh38"},
						},
					},
					expected: policy.Pass,
				},
				{
					name:    "test code reviews required with too many min reviewers but matching required reviewers",
					repoURL: "https://github.com/ossf/scorecard",
					commit:  "fa0592fab28aa92560f04e1ae8649dfff566ae2b",
					policy: policy.AttestationPolicy{
						PreventBinaryArtifacts:      true,
						PreventKnownVulnerabilities: false,
						PreventUnpinnedDependencies: true,
						EnsureCodeReviewed:          true,
						CodeReviewRequirements: policy.CodeReviewRequirements{
							MinReviewers:      2,
							RequiredApprovers: []string{"spencerschrock", "laurentsimon", "naveensrinivasan", "azeemshaikh38"},
						},
					},
					expected: policy.Fail,
				},
			}

			for _, tc := range tt {
				fmt.Printf("attestor_policy_test.go: %s\n", tc.name)
				f, err := os.CreateTemp("/tmp", strings.ReplaceAll(tc.name, " ", "-"))
				Expect(err).Should(BeNil())
				defer os.Remove(f.Name())

				buf, err := yaml.Marshal(tc.policy)
				Expect(err).Should(BeNil())

				nbytes, err := f.Write(buf)
				Expect(err).Should(BeNil())
				Expect(nbytes).Should(BeNumerically(">", 0))

				result, err := command.RunCheckWithParams(tc.repoURL, tc.commit, f.Name())
				Expect(err).Should(BeNil())
				Expect(result).Should(BeEquivalentTo(tc.expected))
			}
		})
	})
})
