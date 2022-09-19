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
	stdos "os"
	"reflect"
	"strings"
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
			filename:       "../testdata/.github/workflows/github-workflow-shells-all-windows-bash.yaml",
			expectedShells: []string{"bash"},
		},
		{
			name:           "all windows, OSes listed in matrix.os",
			filename:       "../testdata/.github/workflows/github-workflow-shells-all-windows-matrix.yaml",
			expectedShells: []string{"pwsh"},
		},
		{
			name:           "all windows, OSes listed in matrix.include",
			filename:       "../testdata/.github/workflows/github-workflow-shells-all-windows-matrix-include.yaml",
			expectedShells: []string{"pwsh"},
		},
		{
			name:           "all windows, empty matrix.include",
			filename:       "../testdata/.github/workflows/github-workflow-shells-all-windows-matrix-include-empty.yaml",
			expectedShells: []string{"pwsh"},
		},
		{
			name:           "all windows",
			filename:       "../testdata/.github/workflows/github-workflow-shells-all-windows.yaml",
			expectedShells: []string{"pwsh"},
		},
		{
			name:           "macOS defaults to bash",
			filename:       "../testdata/.github/workflows/github-workflow-shells-default-macos.yaml",
			expectedShells: []string{"bash"},
		},
		{
			name:           "ubuntu defaults to bash",
			filename:       "../testdata/.github/workflows/github-workflow-shells-default-ubuntu.yaml",
			expectedShells: []string{"bash"},
		},
		{
			name:           "windows defaults to pwsh",
			filename:       "../testdata/.github/workflows/github-workflow-shells-default-windows.yaml",
			expectedShells: []string{"pwsh"},
		},
		{
			name:           "windows specified in 'if'",
			filename:       "../testdata/.github/workflows/github-workflow-shells-runner-windows-ubuntu.yaml",
			expectedShells: append(repeatItem("pwsh", 7), repeatItem("bash", 4)...),
		},
		{
			name:           "shell specified in job and step",
			filename:       "../testdata/.github/workflows/github-workflow-shells-specified-job-step.yaml",
			expectedShells: []string{"bash"},
		},
		{
			name:           "windows, shell specified in job",
			filename:       "../testdata/.github/workflows/github-workflow-shells-specified-job-windows.yaml",
			expectedShells: []string{"bash"},
		},
		{
			name:           "shell specified in job",
			filename:       "../testdata/.github/workflows/github-workflow-shells-specified-job.yaml",
			expectedShells: []string{"pwsh"},
		},
		{
			name:           "shell specified in step",
			filename:       "../testdata/.github/workflows/github-workflow-shells-speficied-step.yaml",
			expectedShells: []string{"pwsh"},
		},
		{
			name:           "different shells in each step",
			filename:       "../testdata/.github/workflows/github-workflow-shells-two-shells.yaml",
			expectedShells: []string{"bash", "pwsh"},
		},
		{
			name:           "windows step, bash specified",
			filename:       "../testdata/.github/workflows/github-workflow-shells-windows-bash.yaml",
			expectedShells: []string{"bash", "bash"},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			content, err := stdos.ReadFile(tt.filename)
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

func TestIsWorkflowFile(t *testing.T) {
	t.Parallel()
	type args struct {
		pathfn string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "yaml",
			args: args{
				pathfn: "./testdata/.github/workflows/github-workflow-shells-all-windows.yaml",
			},
			want: true,
		},
		{
			name: "yml",
			args: args{
				pathfn: "./testdata/.github/workflows/github-workflow-shells-all-windows.yml",
			},
			want: true,
		},
		{
			name: "json",
			args: args{
				pathfn: "./testdata/.github/workflows/github-workflow-shells-all-windows.json",
			},
			want: false,
		},
		{
			name: "txt",
			args: args{
				pathfn: "./testdata/.github/workflows/github-workflow-shells-all-windows.txt",
			},
			want: false,
		},
		{
			name: "md",
			args: args{
				pathfn: "./testdata/.github/workflows/github-workflow-shells-all-windows.md",
			},
			want: false,
		},
		{
			name: "unknown",
			args: args{
				pathfn: "./testdata/.github/workflows/github-workflow-shells-all-windows.unknown",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := strings.Replace(tt.args.pathfn, "./testdata/", "", 1)
			if got := IsWorkflowFile(p); got != tt.want {
				t.Errorf("IsWorkflowFile() = %v, want %v for tests %v", got, tt.want, tt.name)
			}
		})
	}
}

func TestIsGitHubOwnedAction(t *testing.T) {
	t.Parallel()
	type args struct {
		actionName string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "github/foo/bar",
			args: args{
				actionName: "github/foo/bar",
			},
			want: true,
		},
		{
			name: "actions",
			args: args{
				actionName: "actions/bar/",
			},
			want: true,
		},
		{
			name: "foo/bar/",
			args: args{
				actionName: "foo/bar/",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsGitHubOwnedAction(tt.args.actionName); got != tt.want {
				t.Errorf("IsGitHubOwnedAction() = %v, want %v for test %v", got, tt.want, tt.name)
			}
		})
	}
}

// TestGetJobName tests the GetJobName function.
func TestGetJobName(t *testing.T) {
	t.Parallel()
	type args struct {
		job *actionlint.Job
	}
	var name actionlint.String
	name.Value = "foo"
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "job name",
			args: args{
				job: &actionlint.Job{
					Name: &name,
				},
			},
			want: "foo",
		},
		{
			name: "job name is empty",
			args: args{
				job: &actionlint.Job{},
			},
			want: "",
		},
		{
			name: "job is nil",
			args: args{},
			want: "",
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := GetJobName(tt.args.job); got != tt.want {
				t.Errorf("GetJobName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetStepName(t *testing.T) {
	t.Parallel()
	type args struct {
		step *actionlint.Step
	}
	var name actionlint.String
	name.Value = "foo"
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "step name",
			args: args{
				step: &actionlint.Step{
					Name: &name,
				},
			},
			want: "foo",
		},
		{
			name: "step name is empty",
			args: args{
				step: &actionlint.Step{},
			},
			want: "",
		},
		{
			name: "step is nil",
			args: args{},
			want: "",
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := GetStepName(tt.args.step); got != tt.want {
				t.Errorf("GetStepName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsStepExecKind(t *testing.T) {
	t.Parallel()
	type args struct {
		step *actionlint.Step
		kind actionlint.ExecKind
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "step is nil",
			args: args{},
			want: false,
		},
		{
			name: "step is not nil",
			args: args{
				step: &actionlint.Step{
					Exec: &actionlint.ExecAction{},
				},
			},
			want: true,
		},
		{
			name: "step is not nil, but kind is not equal",
			args: args{
				step: &actionlint.Step{
					Exec: &actionlint.ExecRun{},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsStepExecKind(tt.args.step, tt.args.kind); got != tt.want {
				t.Errorf("IsStepExecKind() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetLineNumber(t *testing.T) {
	t.Parallel()
	type args struct {
		pos *actionlint.Pos
	}
	//nolint
	tests := []struct {
		name string
		args args
		want uint
	}{
		{
			name: "line number",
			args: args{
				pos: &actionlint.Pos{
					Line: 1,
				},
			},
			want: 1,
		},
		{
			name: "line number is empty",
			args: args{
				pos: &actionlint.Pos{
					Line: 1,
				},
			},
			want: 1,
		},
		{
			name: "pos is nil",
			args: args{},
			want: 1,
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := GetLineNumber(tt.args.pos); got != tt.want {
				t.Errorf("GetLineNumber() = %v, want %v for %v", got, tt.want, tt.name)
			}
		})
	}
}

func TestFormatActionlintError(t *testing.T) {
	t.Parallel()
	type args struct {
		errs []*actionlint.Error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "no errors",
			args: args{
				errs: []*actionlint.Error{},
			},
			wantErr: false,
		},
		{
			name: "one error",
			args: args{
				errs: []*actionlint.Error{
					{
						Message: "foo",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := FormatActionlintError(tt.args.errs); (err != nil) != tt.wantErr {
				t.Errorf("FormatActionlintError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetUses(t *testing.T) {
	t.Parallel()
	type args struct {
		step *actionlint.Step
	}
	//nolint
	tests := []struct {
		name string
		args args
		want *actionlint.String
	}{
		{
			name: "uses",
			args: args{
				step: &actionlint.Step{
					Exec: &actionlint.ExecAction{
						Uses: &actionlint.String{
							Value: "foo",
						},
					},
				},
			},
			want: &actionlint.String{
				Value: "foo",
			},
		},
		{
			name: "uses is empty",
			args: args{
				step: &actionlint.Step{
					Exec: &actionlint.ExecAction{
						Uses: &actionlint.String{
							Value: "",
						},
					},
				},
			},
			want: &actionlint.String{
				Value: "",
			},
		},
		{
			name: "step is nil",
			args: args{},
			want: nil,
		},
		{
			name: "step is not nil, but uses is nil",
			args: args{
				step: &actionlint.Step{
					Exec: &actionlint.ExecAction{},
				},
			},
			want: nil,
		},
		{
			name: "step is not nil, but uses is not nil",
			args: args{
				step: &actionlint.Step{},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := GetUses(tt.args.step); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUses() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getWith(t *testing.T) {
	t.Parallel()
	type args struct {
		step *actionlint.Step
	}
	//nolint
	tests := []struct {
		name string
		args args
		want map[string]*actionlint.Input
	}{
		{
			name: "with",
			args: args{
				step: &actionlint.Step{
					Exec: &actionlint.ExecAction{
						Inputs: map[string]*actionlint.Input{
							"foo": {
								Name:  &actionlint.String{Value: "foo"},
								Value: &actionlint.String{Value: "bar"},
							},
						},
					},
				},
			},
			want: map[string]*actionlint.Input{
				"foo": {
					Name:  &actionlint.String{Value: "foo"},
					Value: &actionlint.String{Value: "bar"},
				},
			},
		},
		{
			name: "with is empty",
			args: args{
				step: &actionlint.Step{
					Exec: &actionlint.ExecAction{
						Inputs: map[string]*actionlint.Input{
							"foo": {
								Name:  &actionlint.String{Value: "foo"},
								Value: &actionlint.String{Value: ""},
							},
						},
					},
				},
			},
			want: map[string]*actionlint.Input{
				"foo": {
					Name:  &actionlint.String{Value: "foo"},
					Value: &actionlint.String{Value: ""},
				},
			},
		},
		{
			name: "step is nil",
			args: args{},
			want: nil,
		},
		{
			name: "step is not nil, but with is nil",
			args: args{
				step: &actionlint.Step{
					Exec: &actionlint.ExecAction{},
				},
			},
			want: nil,
		},
		{
			name: "step is not nil, but with is not nil",
			args: args{
				step: &actionlint.Step{},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := getWith(tt.args.step); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getWith() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getRun(t *testing.T) {
	t.Parallel()
	type args struct {
		step *actionlint.Step
	}
	//nolint
	tests := []struct {
		name string
		args args
		want *actionlint.String
	}{
		{
			name: "run",
			args: args{
				step: &actionlint.Step{
					Exec: &actionlint.ExecAction{
						Inputs: map[string]*actionlint.Input{
							"foo": {
								Name:  &actionlint.String{Value: "foo"},
								Value: &actionlint.String{Value: "bar"},
							},
						},
					},
				},
			},
			want: nil,
		},
		{
			name: "run is empty",
			args: args{
				step: &actionlint.Step{
					Exec: &actionlint.ExecAction{
						Inputs: map[string]*actionlint.Input{
							"foo": {
								Name:  &actionlint.String{Value: "foo"},
								Value: &actionlint.String{Value: ""},
							},
						},
					},
				},
			},
			want: nil,
		},
		{
			name: "step is nil",
			args: args{},
			want: nil,
		},
		{
			name: "step is not nil, but run is nil",
			args: args{
				step: &actionlint.Step{
					Exec: &actionlint.ExecAction{},
				},
			},
			want: nil,
		},
		{
			name: "step is not nil, but run is not nil",
			args: args{
				step: &actionlint.Step{},
			},
			want: nil,
		},
		{
			name: "run is not empty",
			args: args{
				step: &actionlint.Step{
					Exec: &actionlint.ExecRun{
						Run: &actionlint.String{Value: "foo"},
					},
				},
			},
			want: &actionlint.String{Value: "foo"},
		},
		{
			name: "run is not empty",
			args: args{
				step: &actionlint.Step{
					Exec: &actionlint.ExecRun{},
				},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := getRun(tt.args.step); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getRun() = %v, want %v for test case %v", got, tt.want, tt.name)
			}
		})
	}
}

func Test_stepsMatch(t *testing.T) {
	t.Parallel()
	type args struct {
		stepToMatch *JobMatcherStep
		step        *actionlint.Step
	}
	//nolint
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "match",
			args: args{
				stepToMatch: &JobMatcherStep{
					Uses: "foo",
					With: map[string]string{
						"foo": "bar",
					},
				},
				step: &actionlint.Step{
					Exec: &actionlint.ExecAction{
						Uses: &actionlint.String{
							Value: "foo@",
						},
						Inputs: map[string]*actionlint.Input{
							"foo": {
								Name:  &actionlint.String{Value: "foo"},
								Value: &actionlint.String{Value: "bar"},
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "match with empty",
			args: args{
				stepToMatch: &JobMatcherStep{
					Uses: "foo",
					With: map[string]string{
						"foo": "bar",
					},
				},
				step: &actionlint.Step{
					Exec: &actionlint.ExecAction{
						Uses: &actionlint.String{
							Value: "foo@",
						},
						Inputs: map[string]*actionlint.Input{
							"foo": {
								Name:  &actionlint.String{Value: "foo"},
								Value: &actionlint.String{Value: ""},
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "match with empty",
			args: args{
				stepToMatch: &JobMatcherStep{
					Uses: "foo",
					With: map[string]string{
						"foo": "bar",
					},
				},
				step: &actionlint.Step{
					Exec: &actionlint.ExecAction{
						Uses: &actionlint.String{
							Value: "foo@",
						},
						Inputs: map[string]*actionlint.Input{
							"foo": {
								Name:  &actionlint.String{Value: "foo"},
								Value: &actionlint.String{Value: ""},
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "match with empty",
			args: args{
				stepToMatch: &JobMatcherStep{
					Uses: "foo",
					With: map[string]string{
						"foo": "bar",
					},
				},
				step: &actionlint.Step{
					Exec: &actionlint.ExecAction{
						Uses: &actionlint.String{
							Value: "foo@",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "match with empty",
			args: args{
				stepToMatch: &JobMatcherStep{
					Uses: "foo",
					With: map[string]string{
						"foo": "bar",
					},
				},
				step: &actionlint.Step{
					Exec: &actionlint.ExecAction{
						Uses: &actionlint.String{
							Value: "foo",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "match with empty",
			args: args{
				stepToMatch: &JobMatcherStep{
					Uses: "foo",
					With: map[string]string{
						"foo": "bar",
					},
				},
			},
			want: false,
		},
		{
			name: "match",
			args: args{
				stepToMatch: &JobMatcherStep{
					Uses: "foo",
					With: map[string]string{
						"foo": "bar",
					},
					Run: "foo",
				},
				step: &actionlint.Step{
					Exec: &actionlint.ExecAction{
						Uses: &actionlint.String{
							Value: "foo@",
						},
						Inputs: map[string]*actionlint.Input{
							"foo": {
								Name:  &actionlint.String{Value: "foo"},
								Value: &actionlint.String{Value: "bar"},
							},
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := stepsMatch(tt.args.stepToMatch, tt.args.step); got != tt.want {
				t.Errorf("stepsMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsPackagingWorkflow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{
			name:     "npmjs.org publish",
			filename: "../testdata/.github/workflows/github-workflow-packaging-npm.yaml",
			expected: true,
		},
		{
			name:     "npm github publish",
			filename: "../testdata/.github/workflows/github-workflow-packaging-npm-github.yaml",
			expected: false, // Should this be false?
		},
		{
			name:     "maven publish",
			filename: "../testdata/.github/workflows/github-workflow-packaging-maven.yaml",
			expected: true,
		},
		{
			name:     "gradle publish",
			filename: "../testdata/.github/workflows/github-workflow-packaging-gradle.yaml",
			expected: true,
		},
		{
			name:     "gem publish",
			filename: "../testdata/.github/workflows/github-workflow-packaging-gem.yaml",
			expected: true,
		},
		{
			name:     "nuget publish",
			filename: "../testdata/.github/workflows/github-workflow-packaging-nuget.yaml",
			expected: true,
		},
		{
			name:     "docker action publish",
			filename: "../testdata/.github/workflows/github-workflow-packaging-docker-action.yaml",
			expected: true,
		},
		{
			name:     "docker push publish",
			filename: "../testdata/.github/workflows/github-workflow-packaging-docker-push.yaml",
			expected: true,
		},
		{
			name:     "pypi publish",
			filename: "../testdata/.github/workflows/github-workflow-packaging-pypi.yaml",
			expected: true,
		},
		{
			name:     "pypi publish minimal",
			filename: "../testdata/.github/workflows/github-workflow-packaging-pypi-minimal.yaml",
			expected: true,
		},
		{
			name:     "pypi publish failing",
			filename: "../testdata/.github/workflows/github-workflow-packaging-pypi-failing.yaml",
			expected: false,
		},
		{
			name:     "python semantic release publish",
			filename: "../testdata/.github/workflows/github-workflow-packaging-python-semantic-release.yaml",
			expected: true,
		},
		{
			name:     "go publish",
			filename: "../testdata/.github/workflows/github-workflow-packaging-go.yaml",
			expected: true,
		},
		{
			name:     "cargo publish",
			filename: "../testdata/.github/workflows/github-workflow-packaging-cargo.yaml",
			expected: true,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			content, err := stdos.ReadFile(tt.filename)
			if err != nil {
				t.Errorf("cannot read file: %v", err)
			}
			workflow, errs := actionlint.Parse(content)
			if len(errs) > 0 && workflow == nil {
				t.Errorf("cannot parse file: %v", err)
			}
			p := strings.Replace(tt.filename, "../testdata/", "", 1)

			_, ok := IsPackagingWorkflow(workflow, p)
			if ok != tt.expected {
				t.Errorf("isPackagingWorkflow() = %v, expected %v", ok, tt.expected)
			}
		})
	}
}
