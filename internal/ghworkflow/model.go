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

// Package ghworkflow is Scorecard's model of a GitHub Actions workflow file.
//
// Checks and probes consume only the types in this package. The YAML parsing
// itself is delegated to an external parser, which is confined to convert.go
// so it can be replaced without touching any consumer.
package ghworkflow

import "fmt"

// Pos is a 1-based position in the workflow file.
type Pos struct {
	Line int
	Col  int
}

// String is a scalar value along with where it appears in the file.
type String struct {
	Pos   *Pos
	Value string
}

// Workflow is a parsed workflow file. Fields are nil when the corresponding
// section is absent.
type Workflow struct {
	Permissions *Permissions
	Env         *Env
	Jobs        map[string]*Job
	On          []*Event
}

// Event is one trigger listed under `on:`.
type Event struct {
	Name string
}

// EventName returns the trigger's name, e.g. "pull_request_target".
func (e *Event) EventName() string {
	return e.Name
}

// Permissions is a `permissions:` section. All is set when the section is a
// single string such as "read-all"; Scopes is set when it is a mapping.
type Permissions struct {
	All    *String
	Scopes map[string]*PermissionScope
	Pos    *Pos
}

// PermissionScope is a single `<scope>: <level>` entry in a permissions mapping.
type PermissionScope struct {
	Name  *String
	Value *String
}

// Env is an `env:` section. Vars is keyed by the variable's name.
type Env struct {
	Vars map[string]*EnvVar
}

// EnvVar is a single environment variable assignment.
type EnvVar struct {
	Name  *String
	Value *String
}

// Defaults is a `defaults:` section.
type Defaults struct {
	Run *DefaultsRun
}

// DefaultsRun is the `defaults.run:` section.
type DefaultsRun struct {
	Shell *String
}

// Job is a single entry under `jobs:`.
type Job struct {
	ID           *String
	Name         *String
	RunsOn       *Runner
	Permissions  *Permissions
	Env          *Env
	Defaults     *Defaults
	Strategy     *Strategy
	WorkflowCall *WorkflowCall
	Pos          *Pos
	Steps        []*Step
}

// Runner is a job's `runs-on:` value. When it is a single `${{ }}`
// expression, LabelsExpr is set instead of Labels.
type Runner struct {
	LabelsExpr *String
	Labels     []*String
}

// WorkflowCall is a job that calls a reusable workflow with `uses:`.
type WorkflowCall struct {
	Uses *String
}

// Strategy is a job's `strategy:` section.
type Strategy struct {
	Matrix *Matrix
}

// Matrix is a job's `strategy.matrix:` section.
type Matrix struct {
	Rows    map[string]*MatrixRow
	Include *MatrixCombinations
}

// MatrixRow is one matrix dimension, such as `os: [ubuntu-latest, windows-latest]`.
// String values are stored unquoted; arrays and objects are stored in a
// flow-style rendering.
type MatrixRow struct {
	Name   *String
	Values []string
}

// MatrixCombinations is a matrix `include:` list.
type MatrixCombinations struct {
	Combinations []*MatrixCombination
}

// MatrixCombination is one entry of a matrix `include:` list. Assigns is keyed
// by the assigned key's name.
type MatrixCombination struct {
	Assigns map[string]*MatrixAssign
}

// MatrixAssign is a single `key: value` pair in a matrix combination. Value
// follows the same rendering rules as MatrixRow.Values.
type MatrixAssign struct {
	Key   *String
	Value string
}

// Step is a single entry under a job's `steps:`.
type Step struct {
	ID   *String
	If   *String
	Name *String
	Exec Exec
	Pos  *Pos
}

// ExecKind says whether a step runs an action or a shell script.
type ExecKind uint8

const (
	// ExecKindAction is a step with `uses:`.
	ExecKindAction ExecKind = iota
	// ExecKindRun is a step with `run:`.
	ExecKindRun
)

// Exec is what a step executes: either an *ExecAction or an *ExecRun.
type Exec interface {
	Kind() ExecKind
}

// ExecAction is a step that runs an action. Inputs holds its `with:` values,
// keyed by input name.
type ExecAction struct {
	Uses   *String
	Inputs map[string]*Input
}

// Kind returns ExecKindAction.
func (e *ExecAction) Kind() ExecKind {
	return ExecKindAction
}

// Input is a single `with:` entry of an action step.
type Input struct {
	Name  *String
	Value *String
}

// ExecRun is a step that runs a shell script.
type ExecRun struct {
	Run   *String
	Shell *String
}

// Kind returns ExecKindRun.
func (e *ExecRun) Kind() ExecKind {
	return ExecKindRun
}

// Error is a problem found while parsing a workflow. Line and Column are
// 1-based.
type Error struct {
	Message  string
	Filepath string
	Kind     string
	Line     int
	Column   int
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s [%s]", e.Filepath, e.Line, e.Column, e.Message, e.Kind)
}
