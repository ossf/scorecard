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

package hasSelfHostedRunners

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	"github.com/ossf/scorecard/v5/finding"
)

func toUint(u uint) *uint { return &u }

var remediation = &finding.Remediation{
	Text: "Confirm the job actually needs a self-hosted runner. GitHub-hosted runners cover most cases and are ephemeral by default.\n" +
		"If a self-hosted runner is unavoidable (for example an architecture GitHub doesn't offer), make it ephemeral and sandboxed, and keep it off workflows exposed to untrusted pull requests.\n" +
		"See the GitHub guidance on [hardening self-hosted runners](https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions#hardening-for-self-hosted-runners).",
	Effort: finding.RemediationEffortLow,
}

func Test_Run(t *testing.T) {
	t.Parallel()
	//nolint:govet
	tests := []struct {
		name      string
		filenames []string
		expected  []finding.Finding
		err       error
	}{
		{
			name:      "no workflows",
			filenames: []string{},
			expected: []finding.Finding{
				{
					Probe:   Probe,
					Message: "no self-hosted runners found in GitHub Actions workflows",
					Outcome: finding.OutcomeFalse,
				},
			},
		},
		{
			name:      "github-hosted runner only",
			filenames: []string{".github/workflows/secure.yml"},
			expected: []finding.Finding{
				{
					Probe:   Probe,
					Message: "no self-hosted runners found in GitHub Actions workflows",
					Outcome: finding.OutcomeFalse,
				},
			},
		},
		{
			name:      "self-hosted label",
			filenames: []string{".github/workflows/self-hosted.yml"},
			expected: []finding.Finding{
				{
					Probe:       Probe,
					Message:     `job "build" runs on a self-hosted runner`,
					Outcome:     finding.OutcomeTrue,
					Location:    &finding.Location{Path: ".github/workflows/self-hosted.yml", LineStart: toUint(5)},
					Remediation: remediation,
				},
			},
		},
		{
			name:      "runner group",
			filenames: []string{".github/workflows/group.yml"},
			expected: []finding.Finding{
				{
					Probe:       Probe,
					Message:     `job "build" runs on a self-hosted runner`,
					Outcome:     finding.OutcomeTrue,
					Location:    &finding.Location{Path: ".github/workflows/group.yml", LineStart: toUint(6)},
					Remediation: remediation,
				},
			},
		},
		{
			name:      "mixed hosted and self-hosted jobs",
			filenames: []string{".github/workflows/mixed.yml"},
			expected: []finding.Finding{
				{
					Probe:       Probe,
					Message:     `job "selfhosted" runs on a self-hosted runner`,
					Outcome:     finding.OutcomeTrue,
					Location:    &finding.Location{Path: ".github/workflows/mixed.yml", LineStart: toUint(9)},
					Remediation: remediation,
				},
			},
		},
		{
			name:      "runs-on expression is not resolved",
			filenames: []string{".github/workflows/expression.yml"},
			expected: []finding.Finding{
				{
					Probe:   Probe,
					Message: "no self-hosted runners found in GitHub Actions workflows",
					Outcome: finding.OutcomeFalse,
				},
			},
		},
		{
			name:      "malformed workflow",
			filenames: []string{".github/workflows/malformed.yml"},
			expected: []finding.Finding{
				{
					Probe:    Probe,
					Message:  "malformed GitHub Actions workflow file",
					Outcome:  finding.OutcomeError,
					Location: &finding.Location{Path: ".github/workflows/malformed.yml"},
				},
			},
		},
		{
			name: "multiple workflows",
			filenames: []string{
				".github/workflows/secure.yml",
				".github/workflows/self-hosted.yml",
				".github/workflows/group.yml",
			},
			expected: []finding.Finding{
				{
					Probe:       Probe,
					Message:     `job "build" runs on a self-hosted runner`,
					Outcome:     finding.OutcomeTrue,
					Location:    &finding.Location{Path: ".github/workflows/self-hosted.yml", LineStart: toUint(5)},
					Remediation: remediation,
				},
				{
					Probe:       Probe,
					Message:     `job "build" runs on a self-hosted runner`,
					Outcome:     finding.OutcomeTrue,
					Location:    &finding.Location{Path: ".github/workflows/group.yml", LineStart: toUint(6)},
					Remediation: remediation,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			raw := &checker.CheckRequest{}
			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			mockRepoClient.EXPECT().ListFiles(gomock.Any()).DoAndReturn(
				func(predicate func(string) (bool, error)) ([]string, error) {
					var matches []string
					for _, f := range tt.filenames {
						match, err := predicate(f)
						if err != nil {
							t.Fatalf("unexpected err: %v", err)
						}
						if match {
							matches = append(matches, f)
						}
					}
					return matches, nil
				}).AnyTimes()
			mockRepoClient.EXPECT().GetFileReader(gomock.Any()).DoAndReturn(
				func(file string) (io.ReadCloser, error) {
					return os.Open(filepath.Join("testdata", filepath.Base(file)))
				}).AnyTimes()
			raw.RepoClient = mockRepoClient

			findings, _, err := Run(raw)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.expected, findings, cmpopts.IgnoreUnexported(finding.Finding{})); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func Test_Run_NilRaw(t *testing.T) {
	t.Parallel()
	if _, _, err := Run(nil); err == nil {
		t.Fatal("expected error for nil raw request")
	}
}

func Test_Run_ListFilesError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
	mockRepoClient.EXPECT().ListFiles(gomock.Any()).Return(nil, fmt.Errorf("listing error")).AnyTimes()
	raw := &checker.CheckRequest{RepoClient: mockRepoClient}
	if _, _, err := Run(raw); err == nil {
		t.Fatal("expected error when ListFiles fails")
	}
}
