// Copyright 2026 OpenSSF Scorecard Authors
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

package ghworkflow

import (
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// These tests pin the observable behavior Scorecard relies on, independent of
// which backend parser convert.go uses. A backend change that alters any of
// these values changes Scorecard's results.

func readTestdata(t *testing.T, name string) []byte {
	t.Helper()
	content, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("reading %s: %v", name, err)
	}
	return content
}

func str(value string, line, col int) *String {
	return &String{Value: value, Pos: &Pos{Line: line, Col: col}}
}

func TestParse(t *testing.T) {
	t.Parallel()

	want := &Workflow{
		On: []*Event{
			{Name: "push"},
			{Name: "pull_request_target"},
			{Name: "workflow_run"},
		},
		Permissions: &Permissions{
			All: str("read-all", 9, 14),
			Pos: &Pos{Line: 9, Col: 1},
		},
		Env: &Env{Vars: map[string]*EnvVar{
			"title": {Name: str("TITLE", 11, 3), Value: str("${{ github.event.pull_request.title }}", 11, 10)},
		}},
		Jobs: map[string]*Job{
			"matrix": {
				ID:   str("matrix", 13, 3),
				Name: str("Matrix build", 14, 11),
				Pos:  &Pos{Line: 13, Col: 3},
				RunsOn: &Runner{
					LabelsExpr: str("${{ matrix.os }}", 15, 14),
				},
				Permissions: &Permissions{
					Scopes: map[string]*PermissionScope{
						"contents":      {Name: str("contents", 17, 7), Value: str("read", 17, 17)},
						"pull-requests": {Name: str("pull-requests", 18, 7), Value: str("write", 18, 22)},
					},
					Pos: &Pos{Line: 16, Col: 5},
				},
				Env: &Env{Vars: map[string]*EnvVar{
					"job_var": {Name: str("JOB_VAR", 20, 7), Value: str("value", 20, 16)},
				}},
				Defaults: &Defaults{Run: &DefaultsRun{Shell: str("pwsh", 23, 16)}},
				Strategy: &Strategy{Matrix: &Matrix{
					Rows: map[string]*MatrixRow{
						"os": {Name: str("os", 26, 9), Values: []string{"ubuntu-latest", "windows-latest"}},
					},
					Include: &MatrixCombinations{Combinations: []*MatrixCombination{
						{Assigns: map[string]*MatrixAssign{
							"os":           {Key: str("os", 28, 13), Value: "macos-latest"},
							"experimental": {Key: str("experimental", 29, 13), Value: "true"},
						}},
					}},
				}},
				Steps: []*Step{
					{
						ID:   str("checkout", 32, 13),
						Name: str("Checkout", 31, 15),
						Pos:  &Pos{Line: 31, Col: 9},
						Exec: &ExecAction{
							Uses: str("actions/checkout@v4", 33, 15),
							Inputs: map[string]*Input{
								"ref": {
									Name:  str("ref", 35, 11),
									Value: str("${{ github.event.pull_request.head.sha }}", 35, 16),
								},
								"fetch-depth": {Name: str("fetch-depth", 36, 11), Value: str("0", 36, 24)},
							},
						},
					},
					{
						Name: str("Build", 37, 15),
						If:   str("runner.os == 'Windows'", 38, 13),
						Pos:  &Pos{Line: 37, Col: 9},
						Exec: &ExecRun{
							Run:   str(`echo "${{ github.event.issue.title }}"`, 39, 14),
							Shell: str("bash", 40, 16),
						},
					},
				},
			},
			"labels": {
				ID:  str("labels", 41, 3),
				Pos: &Pos{Line: 41, Col: 3},
				RunsOn: &Runner{
					Labels: []*String{str("self-hosted", 42, 15), str("linux", 42, 28)},
				},
				Steps: []*Step{
					{Pos: &Pos{Line: 44, Col: 9}, Exec: &ExecRun{Run: str("make", 44, 14)}},
				},
			},
			"call": {
				ID:           str("call", 45, 3),
				Pos:          &Pos{Line: 45, Col: 3},
				WorkflowCall: &WorkflowCall{Uses: str("octo-org/example-repo/.github/workflows/reusable.yml@v1", 46, 11)},
			},
		},
	}

	got, errs := Parse(readTestdata(t, "workflow.yaml"))
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Parse() mismatch (-want +got):\n%s", diff)
	}
}

func TestParseExecKind(t *testing.T) {
	t.Parallel()

	got, _ := Parse(readTestdata(t, "workflow.yaml"))
	steps := got.Jobs["matrix"].Steps
	if k := steps[0].Exec.Kind(); k != ExecKindAction {
		t.Errorf("uses step: Kind() = %v, want ExecKindAction", k)
	}
	if k := steps[1].Exec.Kind(); k != ExecKindRun {
		t.Errorf("run step: Kind() = %v, want ExecKindRun", k)
	}
}

// Invalid sections must produce errors while still returning the rest of the
// workflow: most callers only give up on a nil workflow.
func TestParsePartial(t *testing.T) {
	t.Parallel()

	got, errs := Parse(readTestdata(t, "invalid-key.yaml"))
	if got == nil {
		t.Fatal("Parse() returned nil workflow for a file with a structural error")
	}
	if len(errs) != 1 {
		t.Fatalf("got %d errors, want 1: %v", len(errs), errs)
	}
	e := errs[0]
	if e.Line != 5 || e.Column != 5 || e.Kind != "syntax-check" {
		t.Errorf("error position/kind = %d:%d [%s], want 5:5 [syntax-check]", e.Line, e.Column, e.Kind)
	}
	if !strings.HasPrefix(e.Message, `unexpected key "not-a-real-key" for "job" section`) {
		t.Errorf("error message = %q", e.Message)
	}

	job := got.Jobs["build"]
	if job == nil || len(job.Steps) != 1 {
		t.Fatalf("job not parsed alongside the error: %+v", job)
	}
	exec, ok := job.Steps[0].Exec.(*ExecAction)
	if !ok || exec.Uses.Value != "actions/checkout@v4" {
		t.Errorf("step not parsed alongside the error: %+v", job.Steps[0].Exec)
	}
}

func TestParseSyntaxError(t *testing.T) {
	t.Parallel()

	got, errs := Parse(readTestdata(t, "syntax-error.yaml"))
	if got != nil {
		t.Errorf("Parse() = %+v, want nil workflow for invalid YAML", got)
	}
	if len(errs) != 1 {
		t.Fatalf("got %d errors, want 1: %v", len(errs), errs)
	}
	// The YAML library reports where it ran out of input: the line after the
	// unterminated flow sequence.
	if errs[0].Kind != "syntax-check" || errs[0].Line != 6 {
		t.Errorf("error = %+v, want syntax-check on line 6", errs[0])
	}
}

func TestErrorString(t *testing.T) {
	t.Parallel()

	e := &Error{Filepath: "ci.yml", Line: 3, Column: 7, Message: "bad", Kind: "syntax-check"}
	if got, want := e.Error(), "ci.yml:3:7: bad [syntax-check]"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}
