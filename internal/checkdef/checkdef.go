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

// Package checkdef holds each check's wiring: its raw data collector, the
// RawResults field it fills, its probes, and its evaluation.
package checkdef

import (
	"fmt"
	"os"

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
	// experimentalName, when set, gates the check on
	// SCORECARD_EXPERIMENTAL and names it in the warning and error text.
	experimentalName string
	probes           []probes.ProbeImpl
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

// experimental returns d gated on SCORECARD_EXPERIMENTAL. name is how the
// warning and error text refer to the check.
func (d definition) experimental(name string) definition {
	d.experimentalName = name
	return d
}

// Run runs the check: collect raw data, run the probes on it, and evaluate
// the findings. An experimental check returns a runtime error instead
// unless SCORECARD_EXPERIMENTAL is set.
func (d definition) Run(c *checker.CheckRequest) checker.CheckResult {
	if d.experimentalName != "" {
		if _, enabled := os.LookupEnv("SCORECARD_EXPERIMENTAL"); !enabled {
			msg := fmt.Sprintf("SCORECARD_EXPERIMENTAL is not set, not running the %s check", d.experimentalName)
			c.Dlogger.Warn(&checker.LogMessage{Text: msg})
			e := sce.WithMessage(sce.ErrUnsupportedCheck, msg)
			return checker.CreateRuntimeErrorResult(d.name, e)
		}
	}

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

// getRawResults returns the request's RawResults, or a new one when the
// request has none.
func getRawResults(c *checker.CheckRequest) *checker.RawResults {
	if c.RawResults != nil {
		return c.RawResults
	}
	return &checker.RawResults{}
}
