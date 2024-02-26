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

package fuzzed

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	"github.com/ossf/scorecard/v4/probes/internal/utils/uerror"
)

//go:embed *.yml
var fs embed.FS

const (
	Probe   = "fuzzed"
	ToolKey = "tool"
)

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}

	fuzzers := raw.FuzzingResults.Fuzzers

	if len(fuzzers) == 0 {
		f, err := finding.NewNegative(fs, Probe, "no fuzzer integrations found", nil)
		if err != nil {
			return nil, Probe, fmt.Errorf("create finding: %w", err)
		}
		return []finding.Finding{*f}, Probe, nil
	}

	var findings []finding.Finding
	for i := range fuzzers {
		fuzzer := &fuzzers[i]
		// The current implementation does not provide file location
		// for all fuzzers. Check this first.
		if len(fuzzer.Files) == 0 {
			f, err := finding.NewPositive(fs, Probe, fuzzer.Name+" integration found", nil)
			if err != nil {
				return nil, Probe, fmt.Errorf("create finding: %w", err)
			}
			f = f.WithValue(ToolKey, fuzzer.Name)
			findings = append(findings, *f)
		}

		// Files are present. Create one results for each file location.
		for _, file := range fuzzer.Files {
			f, err := finding.NewPositive(fs, Probe, fuzzer.Name+" integration found", file.Location())
			if err != nil {
				return nil, Probe, fmt.Errorf("create finding: %w", err)
			}
			f = f.WithValue(ToolKey, fuzzer.Name)
			findings = append(findings, *f)
		}
	}

	return findings, Probe, nil
}
