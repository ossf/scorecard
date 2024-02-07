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
package sastToolConfigured

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/internal/probes"
	"github.com/ossf/scorecard/v4/probes/internal/utils/uerror"
)

func init() {
	if err := probes.Register(Probe, Run); err != nil {
		panic(err)
	}
}

//go:embed *.yml
var fs embed.FS

const (
	Probe   = "sastToolConfigured"
	ToolKey = "tool"
)

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	r := raw.SASTResults

	if len(r.Workflows) == 0 {
		f, err := finding.NewWith(fs, Probe, "no SAST configuration files detected", nil, finding.OutcomeNegative)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		return []finding.Finding{*f}, Probe, nil
	}

	findings := make([]finding.Finding, len(r.Workflows))
	for i := range r.Workflows {
		tool := string(r.Workflows[i].Type)
		loc := r.Workflows[i].File.Location()
		f, err := finding.NewWith(fs, Probe, "SAST configuration detected: "+tool, loc, finding.OutcomePositive)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		f = f.WithValue(ToolKey, tool)
		findings[i] = *f
	}
	return findings, Probe, nil
}
