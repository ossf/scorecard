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
	"context"
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	"github.com/ossf/scorecard/v4/attestor/command"
	"github.com/ossf/scorecard/v4/attestor/policy"
	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks"
	"github.com/ossf/scorecard/v4/clients"
	sclog "github.com/ossf/scorecard/v4/log"
	"github.com/ossf/scorecard/v4/pkg"
)

var _ = Describe("E2E TEST PAT: scorecard-attestor policy", func() {
	Context("E2E TEST:Validating scorecard attestation policy", func() {
		It("Should attest to known good repos based on policy", func() {
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
			}

			for _, tc := range tt {
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

var _ = Describe("E2E TEST PAT: scorecard-attestor policy", func() {
	Context("E2E TEST:Validating scorecard attestation policy", func() {
		It("Should attest to bad repos based on policy", func() {
			tt := []struct {
				name     string
				repoURL  string
				commit   string
				policy   policy.AttestationPolicy
				expected policy.PolicyResult
			}{
				{
					name: "test bad repo with vulnerabilities prevented but no known vulnerabilities",
					policy: policy.AttestationPolicy{
						PreventKnownVulnerabilities: true,
					},
					expected: policy.Pass,
				},
				{
					name: "test bad repo with ignored binary artifact",
					policy: policy.AttestationPolicy{
						PreventBinaryArtifacts: true,
						AllowedBinaryArtifacts: []string{"test-binary-artifact-*"},
					},
					expected: policy.Pass,
				},
				{
					name: "test bad repo with binary artifact",
					policy: policy.AttestationPolicy{
						PreventBinaryArtifacts: true,
					},
					expected: policy.Fail,
				},
				{
					name: "test bad repo with ignored dep by path",
					policy: policy.AttestationPolicy{
						PreventUnpinnedDependencies: true,
						AllowedUnpinnedDependencies: []policy.Dependency{{Filepath: "Dockerfile"}},
					},
					expected: policy.Pass,
				},
				{
					name: "test bad repo without ignored dep",
					policy: policy.AttestationPolicy{
						PreventUnpinnedDependencies: true,
					},
					expected: policy.Fail,
				},
				{
					name: "test bad repo with ignored dep by name",
					policy: policy.AttestationPolicy{
						PreventUnpinnedDependencies: true,
						AllowedUnpinnedDependencies: []policy.Dependency{{PackageName: "static-debian11"}, {PackageName: "golang"}},
					},
					expected: policy.Pass,
				},
				{
					name: "test bad repo with everything ignored",
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
					name: "test code reviews required but repo doesn't have code reviews",
					policy: policy.AttestationPolicy{
						EnsureCodeReviewed: true,
					},
					expected: policy.Fail,
				},
			}
			results, err := getScorecardResult("https://github.com/ossf-tests/scorecard-binauthz-test-bad")
			Expect(err).Should(BeNil())
			for _, tc := range tt {
				got, err := tc.policy.EvaluateResults(&results.RawResults)
				Expect(err).Should(BeNil())
				Expect(got).Should(BeEquivalentTo(tc.expected))
			}
		})
	})
})

var _ = Describe("E2E TEST PAT: scorecard-attestor policy", func() {
	Context("E2E TEST:Validating scorecard attestation policy", func() {
		It("Should attest to repos based on code review policy", func() {
			tt := []struct {
				name     string
				repoURL  string
				commit   string
				policy   policy.AttestationPolicy
				expected policy.PolicyResult
			}{
				{
					name: "test repo with simple code review requirements",
					policy: policy.AttestationPolicy{
						EnsureCodeReviewed: true,
						CodeReviewRequirements: policy.CodeReviewRequirements{
							MinReviewers: 1,
						},
					},
					expected: policy.Pass,
				},
				{
					name: "test code reviews required with min reviewers",
					policy: policy.AttestationPolicy{
						EnsureCodeReviewed: true,
						CodeReviewRequirements: policy.CodeReviewRequirements{
							MinReviewers: 1,
						},
					},
					expected: policy.Pass,
				},
				{
					name: "test code reviews required with min reviewers and required reviewers",
					policy: policy.AttestationPolicy{
						EnsureCodeReviewed: true,
						CodeReviewRequirements: policy.CodeReviewRequirements{
							MinReviewers:      1,
							RequiredApprovers: []string{"spencerschrock", "laurentsimon", "naveensrinivasan", "azeemshaikh38", "raghavkaul"},
						},
					},
					expected: policy.Pass,
				},
				{
					name: "test code reviews required with too many min reviewers but matching required reviewers",
					policy: policy.AttestationPolicy{
						EnsureCodeReviewed: true,
						CodeReviewRequirements: policy.CodeReviewRequirements{
							MinReviewers:      2,
							RequiredApprovers: []string{"spencerschrock", "laurentsimon", "naveensrinivasan", "azeemshaikh38", "raghavkaul"},
						},
					},
					expected: policy.Fail,
				},
			}
			results, err := getScorecardResult("https://github.com/ossf-tests/scorecard-attestor-code-review-e2e")
			Expect(err).Should(BeNil())
			for _, tc := range tt {
				got, err := tc.policy.EvaluateResults(&results.RawResults)
				Expect(err).Should(BeNil())
				Expect(got).Should(BeEquivalentTo(tc.expected))
			}
		})
	})
})

func getScorecardResult(repoURL string) (pkg.ScorecardResult, error) {
	ctx := context.Background()
	logger := sclog.NewLogger(sclog.DefaultLevel)

	enabledChecks := map[string]checker.Check{
		checks.CheckBinaryArtifacts: {
			Fn: checks.BinaryArtifacts,
		},
		checks.CheckVulnerabilities: {
			Fn: checks.Vulnerabilities,
		},
		checks.CheckCodeReview: {
			Fn: checks.CodeReview,
		},
		checks.CheckPinnedDependencies: {
			Fn: checks.PinningDependencies,
		},
	}
	repo, repoClient, ossFuzzRepoClient, ciiClient, vulnsClient, err := checker.GetClients(
		ctx, repoURL, "", logger)
	if err != nil {
		return pkg.ScorecardResult{}, fmt.Errorf("couldn't set up clients: %w", err)
	}
	//nolint:wrapcheck,lll
	return pkg.RunScorecard(ctx, repo, clients.HeadSHA, 0, enabledChecks, repoClient, ossFuzzRepoClient, ciiClient, vulnsClient)
}
