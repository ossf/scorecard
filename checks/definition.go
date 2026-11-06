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

package checks

import (
	"github.com/ossf/scorecard/v5/checker"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/probes"
	"github.com/ossf/scorecard/v5/probes/zrunner"
)

// evaluateFn turns a check's probe findings into its score.
type evaluateFn func(name string, findings []finding.Finding, dl checker.DetailLogger) checker.CheckResult

// definition wires a check to its raw data collector, the RawResults
// field it fills, its probes, and its evaluation.
type definition struct {
	collectInto func(*checker.CheckRequest, *checker.RawResults) error
	evaluate    evaluateFn
	name        string
	probes      []probes.ProbeImpl
}

// define builds a definition. collect and store are typed together, so a
// collector can only be paired with a RawResults field of its own type.
func define[T any](
	name string,
	collect func(*checker.CheckRequest) (T, error),
	store func(*checker.RawResults, T),
	probeSet []probes.ProbeImpl,
	evaluate evaluateFn,
) definition {
	return definition{
		name:     name,
		probes:   probeSet,
		evaluate: evaluate,
		collectInto: func(c *checker.CheckRequest, r *checker.RawResults) error {
			data, err := collect(c)
			if err != nil {
				return err
			}
			store(r, data)
			return nil
		},
	}
}

// runDefinition runs a check: collect raw data, run the probes on it, and
// evaluate the findings.
func runDefinition(d definition, c *checker.CheckRequest) checker.CheckResult {
	pRawResults := getRawResults(c)
	if err := d.collectInto(c, pRawResults); err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(d.name, e)
	}

	findings, err := zrunner.Run(pRawResults, d.probes)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(d.name, e)
	}

	ret := d.evaluate(d.name, findings, c.Dlogger)
	ret.Findings = findings
	return ret
}
