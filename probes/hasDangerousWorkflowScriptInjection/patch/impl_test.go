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
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"

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
		name     string
		filePath string
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
		// Test currently failing because we don't look for existent env vars pointing to the same content.
		// Once proper behavior is implemented, enable this test
		// {
		// 	name:             "Reuse existent workflow level env var, if it DOES NOT have the same name we'd give",
		// 	inputFilepath:    "reuseEnvVarWithDiffName.yaml",
		// },
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

			numFindings := len(dws)

			for i, dw := range dws {
				if dw.Type == checker.DangerousWorkflowUntrustedCheckout {
					// Patching not yet implemented
					continue
				}

				// Only used for error messages, increment by 1 for human legibility of
				// errors
				i = i + 1

				output, err := patchWorkflow(dw.File, string(inputContent))
				if err != nil {
					t.Errorf("Couldn't patch workflow for finding #%d.", i)
				}

				// build path to fixed version
				dot := strings.LastIndex(tt.filePath, ".")
				fixedPath := tt.filePath[:dot] + "_fixed"
				if numFindings > 1 {
					fixedPath = fmt.Sprintf("%s_%d", fixedPath, i)
				}
				fixedPath = fixedPath + tt.filePath[dot:]

				expectedContent, err := os.ReadFile(path.Join(testDir, fixedPath))
				if err != nil {
					t.Errorf("Couldn't read expected output file for finding #%d. Error:\n%s", i, err)
				}
				expected := string(expectedContent)

				if diff := cmp.Diff(expected, output); diff != "" {
					t.Errorf("mismatch for finding #%d. (-want +got):\n%s", i, diff)
				}
			}

		})
	}
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

// This function parses the diff file and makes a few changes necessary to make a
// valid comparison with the output of GeneratePatch.
//
// For example, the following diff file created with `git diff`:
//
//	diff --git a/testdata/foo.yaml b/testdata/foo_fixed.yaml
//	index 843d0c71..cced3454 100644
//	--- a/testdata/foo.yaml
//	+++ b/testdata/foo_fixed.yaml
//	@@ -6,6 +6,9 @@ jobs:
//	< ... the diff ... >
//
// becomes:
//
//	--- a/testdata/foo.yaml
//	+++ b/testdata/foo_fixed.yaml
//	@@ -6,6 +6,9 @@
//	< ... the diff ... >
//
// Note that, despite the differences between our output and the official
// `git diff`, our output is still valid and can be passed to
// `patch -p1 < path/to/file.diff` to apply the fix to the workflow.
func parseDiffFile(filepath string) (string, error) {
	c, err := os.ReadFile(path.Join("./testdata", filepath))
	if err != nil {
		return "", err
	}

	// The real `git diff` includes multiple "headers" (`diff --git ...`, `index ...`)
	// Our diff does not include these headers; it starts with the "in/out" headers of
	// --- a/path/to/file
	// +++ b/path/to/file
	// We must therefore remove any previous headers from the `git diff`.
	lines := strings.Split(string(c), "\n")
	i := 0
	var line string
	for i, line = range lines {
		if strings.HasPrefix(line, "--- ") {
			break
		}
	}
	content := strings.Join(lines[i:], "\n")

	// The real `git diff` adds contents after the `@@` anchors (the text of the line on
	// which the anchor is placed):
	// 		i.e. `@@ 1,2 3,4 @@ jobs:`
	// while ours does not
	//		i.e. `@@ 1,2 3,4 @@`
	// We must therefore remove that extra content to compare with our diff.
	r := regexp.MustCompile(`(@@[ \d,+-]+@@).*`)
	return r.ReplaceAllString(string(content), "$1"), nil
}
