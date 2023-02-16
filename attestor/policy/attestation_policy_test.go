// Copyright 2022 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package policy

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
	sclog "github.com/ossf/scorecard/v4/log"
)

func (ap *AttestationPolicy) ToJSON() string {
	jsonbytes, err := json.Marshal(ap)
	if err != nil {
		return ""
	}

	return string(jsonbytes)
}

func TestCheckNoVulnerabilities(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err      error
		raw      *checker.RawResults
		name     string
		expected PolicyResult
	}{
		{
			name:     "test with no vulnerabilities",
			raw:      &checker.RawResults{},
			err:      nil,
			expected: Pass,
		},
		{
			name: "test with vulnerabilities",
			raw: &checker.RawResults{
				VulnerabilitiesResults: checker.VulnerabilitiesData{
					Vulnerabilities: []clients.Vulnerability{
						{ID: "foo"},
					},
				},
			},
			expected: Fail,
		},
	}

	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := sclog.NewLogger(sclog.DefaultLevel)
			actual, err := CheckNoVulnerabilities(tt.raw, logger)

			if !errors.Is(err, tt.err) {
				t.Fatalf("%s: expected %v, got %v", tt.name, tt.err, err)
			}
			if actual != tt.expected {
				t.Fatalf("%s: invalid result", tt.name)
			}
		})
	}
}

func TestCheckPreventBinaryArtifacts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                   string
		raw                    *checker.RawResults
		err                    error
		allowedBinaryArtifacts []string
		expected               PolicyResult
	}{
		{
			name:     "test with no artifacts",
			raw:      &checker.RawResults{},
			expected: Pass,
			err:      nil,
		},
		{
			name: "test with multiple artifacts",
			raw: &checker.RawResults{
				BinaryArtifactResults: checker.BinaryArtifactData{Files: []checker.File{
					{Path: "a"},
					{Path: "b"},
				}},
			},
			expected: Fail,
			err:      nil,
		},
		{
			name:                   "test with multiple ignored artifacts",
			allowedBinaryArtifacts: []string{"a", "b"},
			raw: &checker.RawResults{
				BinaryArtifactResults: checker.BinaryArtifactData{Files: []checker.File{
					{Path: "a"},
					{Path: "b"},
				}},
			},
			expected: Pass,
			err:      nil,
		},
		{
			name:                   "test with some artifacts",
			allowedBinaryArtifacts: []string{"a"},
			raw: &checker.RawResults{
				BinaryArtifactResults: checker.BinaryArtifactData{Files: []checker.File{
					{Path: "a"},
					{Path: "b/a"},
				}},
			},
			expected: Fail,
			err:      nil,
		},
		{
			name:                   "test with glob ignored",
			allowedBinaryArtifacts: []string{"a/*", "b/*"},
			raw: &checker.RawResults{
				BinaryArtifactResults: checker.BinaryArtifactData{Files: []checker.File{
					{Path: "a/c/foo.txt"},
					{Path: "b/c/foo.txt"},
				}},
			},
			expected: Pass,
			err:      nil,
		},
	}

	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := sclog.NewLogger(sclog.DefaultLevel)
			actual, err := CheckPreventBinaryArtifacts(tt.allowedBinaryArtifacts, tt.raw, logger)

			if !errors.Is(err, tt.err) {
				t.Fatalf("%s: expected %v, got %v", tt.name, tt.err, err)
			}
			if err != nil {
				return
			}

			if actual != tt.expected {
				t.Fatalf("%s: invalid result", tt.name)
			}
		})
	}
}

func TestCheckCodeReviewed(t *testing.T) {
	t.Parallel()

	// nolint
	tests := []struct {
		err      error
		raw      *checker.RawResults
		reqs     CodeReviewRequirements
		name     string
		expected PolicyResult
	}{
		{
			name: "no review",
			reqs: CodeReviewRequirements{
				MinReviewers: 1,
			},
			raw: &checker.RawResults{
				CodeReviewResults: checker.CodeReviewData{
					DefaultBranchChangesets: []checker.Changeset{
						{
							RevisionID: "1",
							Commits:    []clients.Commit{{SHA: "a"}},
						},
					},
				},
			},
			expected: Fail,
		},
		{
			name: "too few reviews",
			reqs: CodeReviewRequirements{
				MinReviewers: 2,
			},
			raw: &checker.RawResults{
				CodeReviewResults: checker.CodeReviewData{
					DefaultBranchChangesets: []checker.Changeset{
						{
							RevisionID: "1",
							Commits: []clients.Commit{{
								SHA: "a",
								AssociatedMergeRequest: clients.PullRequest{
									Reviews: []clients.Review{
										{Author: &clients.User{Login: "alice"}, State: "APPROVED"},
									},
								},
							}},
						},
					},
				},
			},
			expected: Fail,
		},
		{
			name: "no approvals from the right users",
			reqs: CodeReviewRequirements{
				MinReviewers:      2,
				RequiredApprovers: []string{"bob", "bob-2"},
			},
			raw: &checker.RawResults{
				CodeReviewResults: checker.CodeReviewData{
					DefaultBranchChangesets: []checker.Changeset{
						{
							RevisionID: "1",
							Commits: []clients.Commit{{
								SHA: "a",
								AssociatedMergeRequest: clients.PullRequest{
									Reviews: []clients.Review{
										{Author: &clients.User{Login: "alice"}, State: "APPROVED"},
										{Author: &clients.User{Login: "alice-2"}, State: "APPROVED"},
										{Author: &clients.User{Login: "bob"}, State: "NEEDS_CHANGES"},
									},
								},
							}},
						},
					},
				},
			},
			expected: Fail,
		},
		{
			name: "approvals from one of the required approvers",
			reqs: CodeReviewRequirements{
				MinReviewers:      2,
				RequiredApprovers: []string{"bob", "alice"},
			},
			raw: &checker.RawResults{
				CodeReviewResults: checker.CodeReviewData{
					DefaultBranchChangesets: []checker.Changeset{
						{
							RevisionID: "1",
							Commits: []clients.Commit{{
								SHA: "a",
								AssociatedMergeRequest: clients.PullRequest{
									Reviews: []clients.Review{
										{Author: &clients.User{Login: "alice"}, State: "APPROVED"},
										{Author: &clients.User{Login: "alice-2"}, State: "APPROVED"},
										{Author: &clients.User{Login: "bob"}, State: "NEEDS_CHANGES"},
									},
								},
							}},
						},
						{
							RevisionID: "1",
							Commits: []clients.Commit{{
								SHA: "a",
								AssociatedMergeRequest: clients.PullRequest{
									Reviews: []clients.Review{
										{Author: &clients.User{Login: "alice"}, State: "NEEDS_CHANGES"},
										{Author: &clients.User{Login: "alice-2"}, State: "APPROVED"},
										{Author: &clients.User{Login: "bob"}, State: "APPROVED"},
									},
								},
							}},
						},
					},
				},
			},
			expected: Pass,
		},
		{
			name: "some changesets not reviewed",
			reqs: CodeReviewRequirements{
				MinReviewers:      2,
				RequiredApprovers: []string{"bob"},
			},
			raw: &checker.RawResults{
				CodeReviewResults: checker.CodeReviewData{
					DefaultBranchChangesets: []checker.Changeset{
						{
							RevisionID: "1",
							Commits: []clients.Commit{{
								SHA: "a",
								AssociatedMergeRequest: clients.PullRequest{
									Reviews: []clients.Review{
										{Author: &clients.User{Login: "alice"}, State: "APPROVED"},
										{Author: &clients.User{Login: "bob"}, State: "APPROVED"},
									},
								},
							}},
						},
						{
							RevisionID: "2",
							Commits: []clients.Commit{{
								SHA: "a",
								AssociatedMergeRequest: clients.PullRequest{
									Reviews: []clients.Review{
										{Author: &clients.User{Login: "alice"}, State: "APPROVED"},
										{Author: &clients.User{Login: "bob"}, State: "NEEDS_CHANGES"},
									},
								},
							}},
						},
					},
				},
			},
			expected: Fail,
		},
		{
			name: "code is reviewed sufficiently",
			reqs: CodeReviewRequirements{
				MinReviewers:      2,
				RequiredApprovers: []string{"bob"},
			},
			raw: &checker.RawResults{
				CodeReviewResults: checker.CodeReviewData{
					DefaultBranchChangesets: []checker.Changeset{
						{
							RevisionID: "1",
							Commits: []clients.Commit{
								{
									SHA: "a",
									AssociatedMergeRequest: clients.PullRequest{
										Reviews: []clients.Review{
											{Author: &clients.User{Login: "alice"}, State: "APPROVED"},
										},
									},
								},
								{
									SHA: "b",
									AssociatedMergeRequest: clients.PullRequest{
										Reviews: []clients.Review{
											{Author: &clients.User{Login: "bob"}, State: "APPROVED"},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: Pass,
		},
	}

	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := sclog.NewLogger(sclog.DefaultLevel)
			actual, err := CheckCodeReviewed(tt.reqs, tt.raw, logger)

			if !errors.Is(err, tt.err) {
				t.Fatalf("%s: expected %v, got %v", tt.name, tt.err, err)
			}
			if actual != tt.expected {
				t.Fatalf("%s: invalid result", tt.name)
			}
		})
	}
}

func asStringPointer(s string) *string {
	return &s
}

func TestNoUnpinnedDependencies(t *testing.T) {
	t.Parallel()

	// nolint
	tests := []struct {
		err      error
		raw      *checker.RawResults
		ignores  []Dependency
		name     string
		expected PolicyResult
	}{
		{
			name: "all dependencies pinned",
			raw: &checker.RawResults{
				PinningDependenciesResults: checker.PinningDependenciesData{
					Dependencies: []checker.Dependency{
						{Name: asStringPointer("foo"), PinnedAt: asStringPointer("abcdef")},
					},
				},
			},
			ignores: []Dependency{
				{Filepath: "bar", PackageName: "go-bar", Version: "aaaaaa"},
			},
			expected: Pass,
		},
		{
			name: "some unpinned dependencies",
			raw: &checker.RawResults{
				PinningDependenciesResults: checker.PinningDependenciesData{
					Dependencies: []checker.Dependency{
						{Name: asStringPointer("foo"), PinnedAt: nil},
					},
				},
			},
			ignores: []Dependency{
				{Filepath: "bar", PackageName: "go-bar", Version: "aaaaaa"},
			},
			expected: Fail,
		},
		{
			name: "unpinned dependencies are ignored by name",
			raw: &checker.RawResults{
				PinningDependenciesResults: checker.PinningDependenciesData{
					Dependencies: []checker.Dependency{
						{Name: asStringPointer("foo"), PinnedAt: nil},
					},
				},
			},
			ignores:  []Dependency{{PackageName: "foo"}},
			expected: Pass,
		},
		{
			name: "unpinned dependencies are ignored by path",
			raw: &checker.RawResults{
				PinningDependenciesResults: checker.PinningDependenciesData{
					Dependencies: []checker.Dependency{
						{Name: asStringPointer("second-pkg"), Location: &checker.File{Path: "bar"}, PinnedAt: nil},
					},
				},
			},
			ignores:  []Dependency{{Filepath: "bar"}},
			expected: Pass,
		},
	}

	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := sclog.NewLogger(sclog.DefaultLevel)
			actual, err := CheckNoUnpinnedDependencies(tt.ignores, tt.raw, logger)

			if !errors.Is(err, tt.err) {
				t.Fatalf("%s: expected %v, got %v", tt.name, tt.err, err)
			}
			if actual != tt.expected {
				t.Fatalf("%s: invalid result", tt.name)
			}
		})
	}
}

func TestAttestationPolicyRead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err      error
		name     string
		filename string
		result   AttestationPolicy
	}{
		{
			name:     "default attestation policy with everything on",
			filename: "./testdata/policy-binauthz.yaml",
			err:      nil,
			result: AttestationPolicy{
				PreventBinaryArtifacts:      true,
				AllowedBinaryArtifacts:      []string{},
				PreventKnownVulnerabilities: true,
				PreventUnpinnedDependencies: true,
				AllowedUnpinnedDependencies: []Dependency{},
				EnsureCodeReviewed:          true,
				CodeReviewRequirements:      CodeReviewRequirements{MinReviewers: 1},
			},
		},
		{
			name:     "invalid attestation policy",
			filename: "./testdata/policy-binauthz-invalid.yaml",
			err:      sce.ErrScorecardInternal,
		},
		{
			name:     "policy with allowlists",
			filename: "./testdata/policy-binauthz-allowlist.yaml",
			err:      nil,
			result: AttestationPolicy{
				PreventBinaryArtifacts:      true,
				AllowedBinaryArtifacts:      []string{"/a/b/c", "d"},
				PreventKnownVulnerabilities: false,
				PreventUnpinnedDependencies: false,
				AllowedUnpinnedDependencies: []Dependency{{Filepath: "Dockerfile"}},
				EnsureCodeReviewed:          true,
				CodeReviewRequirements:      CodeReviewRequirements{RequiredApprovers: []string{"alice"}, MinReviewers: 2},
			},
		},
		{
			name:     "policy with a single policy and no policy parameters",
			filename: "./testdata/policy-binauthz-missingparam.yaml",
			err:      nil,
			result: AttestationPolicy{
				PreventBinaryArtifacts: true,
			},
		},
	}

	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p, err := ParseAttestationPolicyFromFile(tt.filename)

			if !errors.Is(err, tt.err) {
				t.Fatalf("%s: expected %v, got %v", tt.name, tt.err, err)
			}
			if err != nil {
				return
			}

			// Compare outputs only if the error is nil.
			// TODO: compare objects.
			if p.ToJSON() != tt.result.ToJSON() {
				fmt.Printf("p.ToJSON(): %v\n", p.ToJSON())
				fmt.Printf("tt.result.ToJSON(): %v\n", tt.result.ToJSON())
				t.Fatalf("%s: invalid result", tt.name)
			}
		})
	}
}
