// Copyright 2024 OpenSSF Scorecard Authors
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

//nolint:stylecheck
package packagedWithNpm

import (
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	"github.com/ossf/scorecard/v5/finding"
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func Test_Run(t *testing.T) {
	t.Parallel()

	t.Run("nil check request", func(t *testing.T) {
		t.Parallel()
		findings, s, err := Run(nil)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if s != "" {
			t.Errorf("expected empty probe name, got %q", s)
		}
		if len(findings) != 0 {
			t.Errorf("expected no findings, got %d", len(findings))
		}
	})

	t.Run("no package.json found", func(t *testing.T) {
		t.Parallel()
		request := setupMockClient(t, []string{"src/main.js", "README.md"}, "", false)
		findings, s, err := Run(request)
		validateResults(t, findings, s, err, []finding.Outcome{finding.OutcomeFalse})
	})

	t.Run("package.json in subdirectory", func(t *testing.T) {
		t.Parallel()
		request := setupMockClient(t, []string{"frontend/package.json", "README.md"}, "", false)
		findings, s, err := Run(request)
		validateResults(t, findings, s, err, []finding.Outcome{finding.OutcomeFalse})
	})

	t.Run("package.json with invalid JSON", func(t *testing.T) {
		t.Parallel()
		request := setupMockClient(t, []string{"package.json"}, `{"name": "test-package",`, true)
		findings, s, err := Run(request)
		validateResults(t, findings, s, err, []finding.Outcome{finding.OutcomeFalse})
	})

	t.Run("package.json without name", func(t *testing.T) {
		t.Parallel()
		request := setupMockClient(t, []string{"package.json"}, `{"version": "1.0.0"}`, true)
		findings, s, err := Run(request)
		validateResults(t, findings, s, err, []finding.Outcome{finding.OutcomeFalse})
	})

	t.Run("package.json with empty name", func(t *testing.T) {
		t.Parallel()
		request := setupMockClient(t, []string{"package.json"}, `{"name": "", "version": "1.0.0"}`, true)
		findings, s, err := Run(request)
		validateResults(t, findings, s, err, []finding.Outcome{finding.OutcomeFalse})
	})
}

func setupMockClient(t *testing.T, files []string, packageContent string, expectFileRead bool) *checker.CheckRequest {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	mockClient := mockrepo.NewMockRepoClient(ctrl)

	// Set up expectations for ListFiles call
	mockClient.EXPECT().ListFiles(gomock.Any()).DoAndReturn(func(predicate func(string) (bool, error)) ([]string, error) {
		var matchedFiles []string
		for _, file := range files {
			match, err := predicate(file)
			if err != nil {
				return nil, err
			}
			if match {
				matchedFiles = append(matchedFiles, file)
			}
		}
		return matchedFiles, nil
	})

	// Set up expectations for file reading if needed
	if expectFileRead {
		reader := nopCloser{strings.NewReader(packageContent)}
		mockClient.EXPECT().GetFileReader("package.json").Return(reader, nil)
	}

	return &checker.CheckRequest{
		RepoClient: mockClient,
	}
}

func validateResults(t *testing.T, findings []finding.Finding, s string, err error, expectedOutcomes []finding.Outcome) {
	t.Helper()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if s != Probe {
		t.Errorf("expected probe name %q, got %q", Probe, s)
	}
	if len(findings) != len(expectedOutcomes) {
		t.Errorf("expected %d findings, got %d", len(expectedOutcomes), len(findings))
		return
	}
	for i, f := range findings {
		if i >= len(expectedOutcomes) {
			break
		}
		if diff := cmp.Diff(expectedOutcomes[i], f.Outcome, cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	}
}
