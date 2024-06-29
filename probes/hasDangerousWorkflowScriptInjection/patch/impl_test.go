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

package patch

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"slices"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/rhysd/actionlint"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks/raw"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
)

const (
	testDir = "./testdata"
)

func Test_patchWorkflow(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		filePath   string
		duplicates map[int]int // mark findings as duplicates of others, use same fix
	}{
		{
			// Extracted from real Angular fix: https://github.com/angular/angular/pull/51026/files
			name:     "Real Example 1",
			filePath: "realExample1.yaml",
		},
		{
			// Inspired on a real fix: https://github.com/googleapis/google-cloud-go/pull/9011/files
			name:     "Real Example 2",
			filePath: "realExample2.yaml",
		},
		{
			// Inspired from a real lit/lit fix: https://github.com/lit/lit/pull/3669/files
			name:     "Real Example 3",
			filePath: "realExample3.yaml",
		},
		{
			name:     "User's input is assigned to a variable before used",
			filePath: "userInputAssignedToVariable.yaml",
		},
		{
			name:     "Two incidences in different jobs",
			filePath: "twoInjectionsDifferentJobs.yaml",
		},
		{
			name:     "Two incidences in same job",
			filePath: "twoInjectionsSameJob.yaml",
		},
		{
			name:     "Two incidences in same step",
			filePath: "twoInjectionsSameStep.yaml",
			duplicates: map[int]int{
				2: 1, // finding #2 is a duplicate of #1
			},
		},
		{
			name:     "4-spaces indentation is kept the same",
			filePath: "fourSpacesIndentationExistentEnvVar.yaml",
		},
		{
			name:     "Crazy but valid indentation is kept the same",
			filePath: "crazyButValidIndentation.yaml",
		},
		{
			name:     "Newline on EOF is kept",
			filePath: "newlineOnEOF.yaml",
		},
		{
			name:     "Ignore if user input regex is just part of a comment",
			filePath: "ignorePatternInsideComments.yaml",
		},
		{
			name:     "Reuse existent workflow level env var, if has the same name we'd give",
			filePath: "reuseWorkflowLevelEnvVars.yaml",
		},
		{
			name:     "Reuse existent workflow level env var, if it DOES NOT have the same name we'd give",
			filePath: "reuseEnvVarWithDiffName.yaml",
		},
		// Test currently failing because we don't look for existent env vars on smaller scopes -- job-level or step-level.
		// In this case, we're always creating a new workflow-level env var. Note that this could lead to creation of env vars shadowed
		// by the ones in smaller scope.
		// Once proper behavior is implemented, enable this test
		// {
		// 	name:             "Reuse env var already existent on smaller scope, it converts case of same or different names",
		// 	inputFilepath:    "reuseEnvVarSmallerScope.yaml",
		// },
		// Test currently failing due to lack of style awareness. Currently we always add a blank line after
		// the env block.
		// Once proper behavior is implemented, enable this test.
		// {
		// 	name:             "Keep style if file doesn't use blank lines between blocks",
		// 	inputFilepath:    "noLineBreaksBetweenBlocks.yaml",
		// 	expectedFilepath: "noLineBreaksBetweenBlocks_fixed.yaml",
		// },
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dws := detectDangerousWorkflows(tt.filePath, t)

			inputContent, err := os.ReadFile(path.Join(testDir, tt.filePath))
			if err != nil {
				t.Errorf("Couldn't read input test file. Error:\n%s", err)
			}

			workflow, inputErrs := actionlint.Parse(inputContent)
			if len(inputErrs) > 0 && workflow == nil {
				t.Errorf("Couldn't parse file as workflow. Error:\n%s", inputErrs[0])
			}

			numFindings := len(dws)
			for i, dw := range dws {
				i = i + 1 // Only used for error messages, increment for legibility

				if dw.Type == checker.DangerousWorkflowUntrustedCheckout {
					t.Errorf("Patching of untrusted checkout (finding #%d) is not implemented.", i)
				}

				output, err := patchWorkflow(dw.File, string(inputContent), workflow)
				if err != nil {
					t.Errorf("Couldn't patch workflow for finding #%d.", i)
				}

				patchedErrs := validatePatchedWorkflow(output, inputErrs)
				if len(patchedErrs) > 0 {
					t.Errorf("Patched workflow for finding #%d is invalid. Error:\n%s", i, patchedErrs[0])
				}

				if dup, ok := tt.duplicates[i]; ok {
					i = dup
				}

				expected, err := getExpected(tt.filePath, numFindings, i)
				if err != nil {
					t.Errorf("Couldn't read expected output file for finding #%d. Error:\n%s", i, err)
				}

				if diff := cmp.Diff(expected, output); diff != "" {
					t.Errorf("mismatch for finding #%d. (-want +got):\n%s", i, diff)
				}
			}

		})
	}
}

func getExpected(filePath string, numFindings, findingIndex int) (string, error) {
	// build path to fixed version
	dot := strings.LastIndex(filePath, ".")
	fixedPath := filePath[:dot] + "_fixed"
	if numFindings > 1 {
		fixedPath = fmt.Sprintf("%s_%d", fixedPath, findingIndex)
	}
	fixedPath = fixedPath + filePath[dot:]

	content, err := os.ReadFile(path.Join(testDir, fixedPath))
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func detectDangerousWorkflows(filePath string, t *testing.T) []checker.DangerousWorkflow {
	ctrl := gomock.NewController(t)
	mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
	mockRepoClient.EXPECT().ListFiles(gomock.Any()).Return(
		// Pretend the file is in the workflow directory to pass a check deep in
		// raw.DangerousWorkflow
		[]string{path.Join(".github/workflows/", filePath)}, nil,
	)
	mockRepoClient.EXPECT().GetFileReader(gomock.Any()).DoAndReturn(func(file string) (io.ReadCloser, error) {
		return os.Open("./testdata/" + filePath)
	}).AnyTimes()

	req := &checker.CheckRequest{
		Ctx:        context.Background(),
		RepoClient: mockRepoClient,
	}

	dw, err := raw.DangerousWorkflow(req)

	if err != nil {
		t.Errorf("Error running raw.DangerousWorkflow. Error:\n%s", err)
	}

	// Sort findings by position. This ensures each finding is compared to its
	// respective "fixed" workflow.
	slices.SortFunc(dw.Workflows, func(a, b checker.DangerousWorkflow) int {
		aPos := a.File.Offset
		bPos := b.File.Offset
		if aPos < bPos {
			return -1
		}
		if aPos > bPos {
			return +1
		}
		return 0
	})

	return dw.Workflows
}
