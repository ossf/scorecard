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

package fileparser

import (
	"io/ioutil"
	"testing"

	"github.com/rhysd/actionlint"
	"gotest.tools/assert/cmp"
)

func TestGitHubWorkflowShell(t *testing.T) {
	t.Parallel()

	repeatItem := func(item string, count int) []string {
		ret := make([]string, 0, count)
		for i := 0; i < count; i++ {
			ret = append(ret, item)
		}
		return ret
	}

	tests := []struct {
		name     string
		filename string
		// The shells used in each step, listed in order that the steps are listed in the file
		expectedShells []string
	}{
		{
			name:           "all windows, shell specified in step",
			filename:       "../testdata/github-workflow-shells-all-windows-bash.yaml",
			expectedShells: []string{"bash"},
		},
		{
			name:           "all windows, OSes listed in matrix.os",
			filename:       "../testdata/github-workflow-shells-all-windows-matrix.yaml",
			expectedShells: []string{"pwsh"},
		},
		{
			name:           "all windows",
			filename:       "../testdata/github-workflow-shells-all-windows.yaml",
			expectedShells: []string{"pwsh"},
		},
		{
			name:           "macOS defaults to bash",
			filename:       "../testdata/github-workflow-shells-default-macos.yaml",
			expectedShells: []string{"bash"},
		},
		{
			name:           "ubuntu defaults to bash",
			filename:       "../testdata/github-workflow-shells-default-ubuntu.yaml",
			expectedShells: []string{"bash"},
		},
		{
			name:           "windows defaults to pwsh",
			filename:       "../testdata/github-workflow-shells-default-windows.yaml",
			expectedShells: []string{"pwsh"},
		},
		{
			name:           "windows specified in 'if'",
			filename:       "../testdata/github-workflow-shells-runner-windows-ubuntu.yaml",
			expectedShells: append(repeatItem("pwsh", 7), repeatItem("bash", 4)...),
		},
		{
			name:           "shell specified in job and step",
			filename:       "../testdata/github-workflow-shells-specified-job-step.yaml",
			expectedShells: []string{"bash"},
		},
		{
			name:           "windows, shell specified in job",
			filename:       "../testdata/github-workflow-shells-specified-job-windows.yaml",
			expectedShells: []string{"bash"},
		},
		{
			name:           "shell specified in job",
			filename:       "../testdata/github-workflow-shells-specified-job.yaml",
			expectedShells: []string{"pwsh"},
		},
		{
			name:           "shell specified in step",
			filename:       "../testdata/github-workflow-shells-speficied-step.yaml",
			expectedShells: []string{"pwsh"},
		},
		{
			name:           "different shells in each step",
			filename:       "../testdata/github-workflow-shells-two-shells.yaml",
			expectedShells: []string{"bash", "pwsh"},
		},
		{
			name:           "windows step, bash specified",
			filename:       "../testdata/github-workflow-shells-windows-bash.yaml",
			expectedShells: []string{"bash", "bash"},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			content, err := ioutil.ReadFile(tt.filename)
			if err != nil {
				t.Errorf("cannot read file: %v", err)
			}

			workflow, errs := actionlint.Parse(content)
			if len(errs) > 0 && workflow == nil {
				t.Errorf("cannot unmarshal file: %v", errs[0])
			}
			actualShells := make([]string, 0)
			for _, job := range workflow.Jobs {
				job := job
				for _, step := range job.Steps {
					step := step
					shell, err := GetShellForStep(step, job)
					if err != nil {
						t.Errorf("error getting shell: %v", err)
					}
					actualShells = append(actualShells, shell)
				}
			}
			if !cmp.DeepEqual(tt.expectedShells, actualShells)().Success() {
				t.Errorf("%v: Got (%v) expected (%v)", tt.name, actualShells, tt.expectedShells)
			}
		})
	}
}
