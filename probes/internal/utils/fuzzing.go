// Copyright 2023 OpenSSF Scorecard Authors
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

package utils

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
)

func FuzzerRun(raw *checker.RawResults, fs embed.FS, probeID, fuzzerName string) ([]finding.Finding, string, error) {
	var findings []finding.Finding
	fuzzers := raw.FuzzingResults.Fuzzers

	for i := range fuzzers {
		fuzzer := fuzzers[i]
		if fuzzer.Name != fuzzerName {
			continue
		}

		// The current implementation does not provide file location
		// for all fuzzers. Check this first.
		if len(fuzzer.Files) == 0 {
			f, err := finding.NewWith(fs, probeID,
				fmt.Sprintf("%s integration found", fuzzerName), nil,
				finding.OutcomePositive)
			if err != nil {
				return nil, probeID, fmt.Errorf("create finding: %w", err)
			}
			findings = append(findings, *f)
			continue
		}

		// Files are present. Create one results for each file location.
		for j := range fuzzer.Files {
			file := fuzzer.Files[j]
			f, err := finding.NewWith(fs, probeID,
				fmt.Sprintf("%s integration found", fuzzerName), file.Location(),
				finding.OutcomePositive)
			if err != nil {
				return nil, probeID, fmt.Errorf("create finding: %w", err)
			}
			findings = append(findings, *f)
		}

	}

	if len(findings) == 0 {
		f, err := finding.NewNegative(fs, probeID,
			fmt.Sprintf("no %s integration found", fuzzerName), nil)
		if err != nil {
			return nil, probeID, fmt.Errorf("create finding: %w", err)
		}
		findings = append(findings, *f)
	}

	return findings, probeID, nil
}
