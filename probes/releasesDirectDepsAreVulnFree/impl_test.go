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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
)

//nolint:gocognit // Test function has many test cases with detailed struct initialization
func Test_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		raw      *checker.RawResults
		err      error
		outcomes []finding.Outcome
	}{
		{
			name: "no releases",
			raw: &checker.RawResults{
				ReleaseDirectDepsVulnsResults: checker.ReleaseDirectDepsVulnsData{
					Releases: []checker.ReleaseDepsVulns{},
				},
			},
			outcomes: []finding.Outcome{},
		},
		{
			name: "single clean release",
			raw: &checker.RawResults{
				ReleaseDirectDepsVulnsResults: checker.ReleaseDirectDepsVulnsData{
					Releases: []checker.ReleaseDepsVulns{
						{
							Tag:       "v1.0.0",
							CommitSHA: "abc123",
							DirectDeps: []checker.DirectDep{
								{
									Ecosystem: "Go",
									Name:      "github.com/pkg/errors",
									Version:   "0.9.1",
								},
							},
							Findings: []checker.DepVuln{}, // No vulnerabilities
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
			},
		},
		{
			name: "single release with vulnerabilities",
			raw: &checker.RawResults{
				ReleaseDirectDepsVulnsResults: checker.ReleaseDirectDepsVulnsData{
					Releases: []checker.ReleaseDepsVulns{
						{
							Tag:       "v0.5.0",
							CommitSHA: "def456",
							DirectDeps: []checker.DirectDep{
								{
									Ecosystem: "Go",
									Name:      "golang.org/x/text",
									Version:   "0.3.5",
								},
							},
							Findings: []checker.DepVuln{
								{
									Ecosystem:    "Go",
									Name:         "golang.org/x/text",
									Version:      "0.3.5",
									OSVIDs:       []string{"CVE-2021-38561"},
									ManifestPath: "go.mod",
								},
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeFalse,
			},
		},
		{
			name: "mixed releases - some clean, some vulnerable",
			raw: &checker.RawResults{
				ReleaseDirectDepsVulnsResults: checker.ReleaseDirectDepsVulnsData{
					Releases: []checker.ReleaseDepsVulns{
						{
							Tag:       "v1.0.0",
							CommitSHA: "abc123",
							Findings:  []checker.DepVuln{}, // Clean
						},
						{
							Tag:       "v0.9.0",
							CommitSHA: "def456",
							Findings:  []checker.DepVuln{}, // Clean
						},
						{
							Tag:       "v0.8.0",
							CommitSHA: "ghi789",
							Findings: []checker.DepVuln{
								{
									Ecosystem:    "Go",
									Name:         "github.com/golang-jwt/jwt/v4",
									Version:      "4.5.1",
									OSVIDs:       []string{"CVE-2025-30204"},
									ManifestPath: "go.mod",
								},
							}, // Vulnerable
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeTrue,
				finding.OutcomeTrue,
				finding.OutcomeFalse,
			},
		},
		{
			name: "release with multiple vulnerable dependencies",
			raw: &checker.RawResults{
				ReleaseDirectDepsVulnsResults: checker.ReleaseDirectDepsVulnsData{
					Releases: []checker.ReleaseDepsVulns{
						{
							Tag:       "v0.1.0",
							CommitSHA: "old123",
							Findings: []checker.DepVuln{
								{
									Ecosystem:    "Go",
									Name:         "golang.org/x/text",
									Version:      "0.3.6",
									OSVIDs:       []string{"CVE-2021-38561"},
									ManifestPath: "go.mod",
								},
								{
									Ecosystem:    "Go",
									Name:         "github.com/gorilla/websocket",
									Version:      "1.4.0",
									OSVIDs:       []string{"CVE-2020-27813"},
									ManifestPath: "go.mod",
								},
							},
						},
					},
				},
			},
			outcomes: []finding.Outcome{
				finding.OutcomeFalse, // Only one finding per release
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			findings, _, err := Run(tt.raw)
			if err != nil {
				if tt.err == nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}

			if len(findings) != len(tt.outcomes) {
				t.Errorf("expected %d findings, got %d", len(tt.outcomes), len(findings))
			}

			for i, f := range findings {
				if i >= len(tt.outcomes) {
					break
				}
				if f.Outcome != tt.outcomes[i] {
					t.Errorf("finding %d: expected outcome %v, got %v", i, tt.outcomes[i], f.Outcome)
				}

				// Verify finding structure
				if f.Probe != Probe {
					t.Errorf("finding %d: expected probe %s, got %s", i, Probe, f.Probe)
				}

				if f.Message == "" {
					t.Errorf("finding %d: message should not be empty", i)
				}

				if f.Location == nil {
					t.Errorf("finding %d: location should not be nil", i)
				}
			}

			// Verify specific message content for clean releases
			for i, f := range findings {
				if f.Outcome == finding.OutcomeTrue {
					release := tt.raw.ReleaseDirectDepsVulnsResults.Releases[i]
					expectedMsg := "release " + release.Tag + " has no known vulnerabilities in direct dependencies"
					if f.Message != expectedMsg {
						t.Errorf("finding %d: expected message %q, got %q", i, expectedMsg, f.Message)
					}
				}
			}
		})
	}
}

func TestRun_FindingsOrder(t *testing.T) {
	t.Parallel()

	raw := &checker.RawResults{
		ReleaseDirectDepsVulnsResults: checker.ReleaseDirectDepsVulnsData{
			Releases: []checker.ReleaseDepsVulns{
				{Tag: "v3.0.0", Findings: []checker.DepVuln{}},
				{Tag: "v2.0.0", Findings: []checker.DepVuln{{Name: "vuln-dep", OSVIDs: []string{"CVE-1"}}}},
				{Tag: "v1.0.0", Findings: []checker.DepVuln{}},
			},
		},
	}

	findings, _, err := Run(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(findings) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(findings))
	}

	// Verify order is preserved
	expectedOutcomes := []finding.Outcome{
		finding.OutcomeTrue,  // v3.0.0
		finding.OutcomeFalse, // v2.0.0
		finding.OutcomeTrue,  // v1.0.0
	}

	for i, expected := range expectedOutcomes {
		if findings[i].Outcome != expected {
			t.Errorf("finding %d: expected outcome %v, got %v", i, expected, findings[i].Outcome)
		}
	}
}

func TestRun_CompareWithExpected(t *testing.T) {
	t.Parallel()

	raw := &checker.RawResults{
		ReleaseDirectDepsVulnsResults: checker.ReleaseDirectDepsVulnsData{
			Releases: []checker.ReleaseDepsVulns{
				{
					Tag:       "v1.0.0",
					CommitSHA: "abc123",
					Findings:  []checker.DepVuln{},
				},
			},
		},
	}

	findings, msg, err := Run(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if msg != "checked recent releases for vulnerable direct dependencies" {
		t.Errorf("unexpected summary message: %s", msg)
	}

	expected := []finding.Finding{
		{
			Probe:   Probe,
			Outcome: finding.OutcomeTrue,
			Message: "release v1.0.0 has no known vulnerabilities in direct dependencies",
			Location: &finding.Location{
				Path: "releases/v1.0.0",
			},
		},
	}

	if diff := cmp.Diff(expected, findings,
		cmpopts.IgnoreUnexported(finding.Finding{}),
		cmpopts.IgnoreFields(finding.Finding{}, "Values")); diff != "" {
		t.Errorf("findings mismatch (-want +got):\n%s", diff)
	}
}
