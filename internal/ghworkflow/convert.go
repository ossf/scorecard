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

// This file is the only place that imports the backend parser. Everything it
// returns is converted into this package's types.

import (
	actionlint "actionlint.kjanat.dev"
)

func parse(content []byte) (*Workflow, []*Error) {
	w, errs := actionlint.Parse(content)
	return convertWorkflow(w), convertErrors(errs)
}

func convertErrors(errs []*actionlint.Error) []*Error {
	if errs == nil {
		return nil
	}
	out := make([]*Error, 0, len(errs))
	for _, e := range errs {
		if e == nil {
			continue
		}
		out = append(out, &Error{
			Message:  e.Message,
			Filepath: e.Filepath,
			Line:     e.Line,
			Column:   e.Column,
			Kind:     e.Kind,
		})
	}
	return out
}

func convertPos(p *actionlint.Pos) *Pos {
	if p == nil {
		return nil
	}
	return &Pos{Line: p.Line, Col: p.Col}
}

func convertString(s *actionlint.String) *String {
	if s == nil {
		return nil
	}
	return &String{Value: s.Value, Pos: convertPos(s.Pos)}
}

func convertStrings(ss []*actionlint.String) []*String {
	if ss == nil {
		return nil
	}
	out := make([]*String, 0, len(ss))
	for _, s := range ss {
		out = append(out, convertString(s))
	}
	return out
}

func convertWorkflow(w *actionlint.Workflow) *Workflow {
	if w == nil {
		return nil
	}
	out := &Workflow{
		Permissions: convertPermissions(w.Permissions),
		Env:         convertEnv(w.Env),
	}
	if w.On != nil {
		out.On = make([]*Event, 0, len(w.On))
		for _, e := range w.On {
			if e == nil {
				continue
			}
			out.On = append(out.On, &Event{Name: e.EventName()})
		}
	}
	if w.Jobs != nil {
		out.Jobs = make(map[string]*Job, len(w.Jobs))
		for k, j := range w.Jobs {
			out.Jobs[k] = convertJob(j)
		}
	}
	return out
}

func convertPermissions(p *actionlint.Permissions) *Permissions {
	if p == nil {
		return nil
	}
	out := &Permissions{All: convertString(p.All), Pos: convertPos(p.Pos)}
	if p.Scopes != nil {
		out.Scopes = make(map[string]*PermissionScope, len(p.Scopes))
		for k, s := range p.Scopes {
			if s == nil {
				out.Scopes[k] = nil
				continue
			}
			out.Scopes[k] = &PermissionScope{Name: convertString(s.Name), Value: convertString(s.Value)}
		}
	}
	return out
}

func convertEnv(e *actionlint.Env) *Env {
	if e == nil {
		return nil
	}
	out := &Env{}
	if e.Vars != nil {
		out.Vars = make(map[string]*EnvVar, len(e.Vars))
		for k, v := range e.Vars {
			if v == nil {
				out.Vars[k] = nil
				continue
			}
			out.Vars[k] = &EnvVar{Name: convertString(v.Name), Value: convertString(v.Value)}
		}
	}
	return out
}

func convertJob(j *actionlint.Job) *Job {
	if j == nil {
		return nil
	}
	out := &Job{
		ID:          convertString(j.ID),
		Name:        convertString(j.Name),
		Permissions: convertPermissions(j.Permissions),
		Env:         convertEnv(j.Env),
		Pos:         convertPos(j.Pos),
	}
	if j.RunsOn != nil {
		// The backend reports `runs-on: ${{ ... }}` in Expression and keeps
		// LabelsExpr for `runs-on: {labels: ${{ ... }}}`. Both are a single
		// expression in place of the label list.
		labelsExpr := j.RunsOn.Expression
		if labelsExpr == nil {
			labelsExpr = j.RunsOn.LabelsExpr
		}
		out.RunsOn = &Runner{
			Labels:     convertStrings(j.RunsOn.Labels),
			LabelsExpr: convertString(labelsExpr),
		}
	}
	if j.Defaults != nil {
		out.Defaults = &Defaults{}
		if j.Defaults.Run != nil {
			out.Defaults.Run = &DefaultsRun{Shell: convertString(j.Defaults.Run.Shell)}
		}
	}
	if j.Strategy != nil {
		out.Strategy = &Strategy{Matrix: convertMatrix(j.Strategy.Matrix)}
	}
	if j.WorkflowCall != nil {
		out.WorkflowCall = &WorkflowCall{Uses: convertString(j.WorkflowCall.Uses)}
	}
	if j.Steps != nil {
		out.Steps = make([]*Step, 0, len(j.Steps))
		for _, s := range j.Steps {
			out.Steps = append(out.Steps, convertStep(s))
		}
	}
	return out
}

func convertMatrix(m *actionlint.Matrix) *Matrix {
	if m == nil {
		return nil
	}
	out := &Matrix{}
	if m.Rows != nil {
		out.Rows = make(map[string]*MatrixRow, len(m.Rows))
		for k, r := range m.Rows {
			if r == nil {
				out.Rows[k] = nil
				continue
			}
			row := &MatrixRow{Name: convertString(r.Name)}
			if r.Values != nil {
				row.Values = make([]string, 0, len(r.Values))
				for _, v := range r.Values {
					row.Values = append(row.Values, rawValue(v))
				}
			}
			out.Rows[k] = row
		}
	}
	if m.Include != nil {
		out.Include = &MatrixCombinations{}
		if m.Include.Combinations != nil {
			out.Include.Combinations = make([]*MatrixCombination, 0, len(m.Include.Combinations))
			for _, c := range m.Include.Combinations {
				out.Include.Combinations = append(out.Include.Combinations, convertCombination(c))
			}
		}
	}
	return out
}

func convertCombination(c *actionlint.MatrixCombination) *MatrixCombination {
	if c == nil {
		return nil
	}
	out := &MatrixCombination{}
	if c.Assigns != nil {
		out.Assigns = make(map[string]*MatrixAssign, len(c.Assigns))
		for k, a := range c.Assigns {
			if a == nil {
				out.Assigns[k] = nil
				continue
			}
			out.Assigns[k] = &MatrixAssign{Key: convertString(a.Key), Value: rawValue(a.Value)}
		}
	}
	return out
}

// rawValue renders a matrix value. String scalars are returned unquoted;
// other values use the backend's flow-style rendering.
func rawValue(v actionlint.RawYAMLValue) string {
	switch v := v.(type) {
	case nil:
		return ""
	case *actionlint.RawYAMLString:
		return v.Value
	default:
		return v.String()
	}
}

func convertStep(s *actionlint.Step) *Step {
	if s == nil {
		return nil
	}
	out := &Step{
		ID:   convertString(s.ID),
		If:   convertString(s.If),
		Name: convertString(s.Name),
		Pos:  convertPos(s.Pos),
	}
	switch e := s.Exec.(type) {
	case *actionlint.ExecAction:
		if e == nil {
			break
		}
		a := &ExecAction{Uses: convertString(e.Uses)}
		if e.Inputs != nil {
			a.Inputs = make(map[string]*Input, len(e.Inputs))
			for k, in := range e.Inputs {
				if in == nil {
					a.Inputs[k] = nil
					continue
				}
				a.Inputs[k] = &Input{Name: convertString(in.Name), Value: convertString(in.Value)}
			}
		}
		out.Exec = a
	case *actionlint.ExecRun:
		if e == nil {
			break
		}
		out.Exec = &ExecRun{Run: convertString(e.Run), Shell: convertString(e.Shell)}
	}
	return out
}
