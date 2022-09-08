package policy

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
	scut "github.com/ossf/scorecard/v4/utests"
)

func (a AttestationPolicy) ToJSON() string {
	jsonbytes, err := json.Marshal(a)
	if err != nil {
		return ""
	}

	return string(jsonbytes)
}

func TestCheckPreventBinaryArtifacts(t *testing.T) {
	t.Parallel()

	dl := scut.TestDetailLogger{}

	tests := []struct {
		name                   string
		raw                    *checker.RawResults
		err                    error
		allowedBinaryArtifacts []string
		expected               PolicyResult
	}{
		{
			name: "test with no artifacts",
			raw: &checker.RawResults{
				BinaryArtifactResults: checker.BinaryArtifactData{Files: []checker.File{}},
			},
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
			name:                   "test with entire artifact path ignored",
			allowedBinaryArtifacts: []string{"a/", "b"},
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
			actual, err := CheckPreventBinaryArtifacts(tt.allowedBinaryArtifacts, tt.raw, &dl)

			if !errors.Is(err, tt.err) {
				t.Fatalf("%s: expected %v, got %v", tt.name, tt.err, err)
			}
			if err != nil {
				return
			}

			// Compare outputs only if the error is nil.
			// TODO: compare objects.
			if actual != tt.expected {
				t.Fatalf("%s: invalid result", tt.name)
			}
		})
	}
}

func TestCheckNoVulnerabilities(t *testing.T) {
	t.Parallel()

	dl := scut.TestDetailLogger{}
	tests := []struct {
		err      error
		raw      *checker.RawResults
		name     string
		expected PolicyResult
	}{
		{
			name: "Test with no vulns",
			raw: &checker.RawResults{
				VulnerabilitiesResults: checker.VulnerabilitiesData{Vulnerabilities: []clients.Vulnerability{}},
			},
			expected: Pass,
		},
		{
			name: "Test with vulns",
			raw: &checker.RawResults{
				VulnerabilitiesResults: checker.VulnerabilitiesData{
					Vulnerabilities: []clients.Vulnerability{{ID: "osv-foo"}},
				},
			},
			expected: Fail,
		},
	}

	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			actual, err := CheckNoVulnerabilities(tt.raw, &dl)

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

func TestCheckNoUnpinnedDependencies(t *testing.T) {
	t.Parallel()

	dl := scut.TestDetailLogger{}
	ghaction := "myactionfoo"

	tests := []struct {
		err      error
		raw      *checker.RawResults
		name     string
		expected PolicyResult
	}{
		{
			name: "no unpinned dependencies found",
			raw: &checker.RawResults{PinningDependenciesResults: checker.PinningDependenciesData{
				Dependencies: []checker.Dependency{},
			}},
			expected: Pass,
		},
		{
			name: "unpinned action workflow",
			raw: &checker.RawResults{PinningDependenciesResults: checker.PinningDependenciesData{
				Dependencies: []checker.Dependency{
					{
						Name:     &ghaction,
						Type:     checker.DependencyUseTypeGHAction,
						Location: &checker.File{Path: ".github/workflows/actions.yml"},
					},
				},
			}},
			expected: Fail,
		},
		{
			name: "dependency not pinned by hash",
			raw: &checker.RawResults{PinningDependenciesResults: checker.PinningDependenciesData{
				Dependencies: []checker.Dependency{
					{
						Name:     &ghaction,
						Type:     checker.DependencyUseTypePipCommand,
						Location: &checker.File{Path: "requirements.txt"},
					},
				},
			}},
			expected: Fail,
		},
	}

	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			actual, err := CheckNoUnpinnedDependencies(tt.raw, &dl)

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

func TestCheckCodeReview(t *testing.T) {
	t.Parallel()

	dl := scut.TestDetailLogger{}

	alice := clients.User{Login: "alice"}
	bob := clients.User{Login: "bob"}

	tests := []struct {
		err      error
		raw      *checker.RawResults
		name     string
		expected PolicyResult
	}{
		{
			name: "Commit with no corresponding PR",
			raw: &checker.RawResults{
				CodeReviewResults: checker.CodeReviewData{
					DefaultBranchCommits: []clients.Commit{
						{
							SHA:       "abc123",
							Committer: alice,
						},
					},
				},
			},
			expected: Fail,
		},
		{
			name: "Commit with code reviewed in GitHub PR",
			raw: &checker.RawResults{
				CodeReviewResults: checker.CodeReviewData{
					DefaultBranchCommits: []clients.Commit{
						{
							SHA:       "def123",
							Committer: alice,
							AssociatedMergeRequest: clients.PullRequest{
								Number:   1,
								MergedAt: time.Now(),
								HeadSHA:  "def123",
								Author:   alice,
								Reviews: []clients.Review{
									{
										Author: &bob,
										State:  "APPROVED",
									},
								},
							},
						},
					},
				},
			},
			expected: Pass,
			err:      nil,
		},
	}

	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			actual, err := CheckCodeReviewed(tt.raw, &dl)

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
				PreventBinaryArtifacts:   true,
				AllowedBinaryArtifacts:   []string{},
				EnsureNoVulnerabilities:  true,
				EnsurePinnedDependencies: true,
				EnsureCodeReviewed:       true,
			},
		},
		{
			name:     "invalid attestation policy",
			filename: "./testdata/policy-binauthz-invalid.yaml",
			err:      sce.ErrScorecardInternal,
		},
		{
			name:     "policy with allowlist of binary artifacts",
			filename: "./testdata/policy-binauthz-allowlist.yaml",
			err:      nil,
			result: AttestationPolicy{
				PreventBinaryArtifacts:   true,
				AllowedBinaryArtifacts:   []string{"/a/b/c", "d"},
				EnsureNoVulnerabilities:  true,
				EnsurePinnedDependencies: true,
				EnsureCodeReviewed:       true,
			},
		},
		{
			name:     "policy with allowlist of binary artifacts",
			filename: "./testdata/policy-binauthz-missingparam.yaml",
			err:      nil,
			result: AttestationPolicy{
				PreventBinaryArtifacts:   true,
				AllowedBinaryArtifacts:   nil,
				EnsureNoVulnerabilities:  true,
				EnsurePinnedDependencies: true,
				EnsureCodeReviewed:       false,
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
				t.Fatalf("%s: invalid result", tt.name)
			}
		})
	}
}
