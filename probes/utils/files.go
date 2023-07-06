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

func FilesRun(files []checker.File, metadata map[string]string, fs embed.FS, probeID, filetype string,
	foundOutcome, notFoundOutcome finding.Outcome, match func(file checker.File) bool,
) ([]finding.Finding, string, error) {
	var findings []finding.Finding
	for i := range files {
		file := files[i]
		if !match(file) {
			continue
		}
		f, err := finding.NewWith(fs, probeID, fmt.Sprintf("%s found", filetype),
			file.Location(), foundOutcome)
		if err != nil {
			return nil, probeID, fmt.Errorf("create finding: %w", err)
		}
		f = f.WithRemediationMetadata(metadata)
		findings = append(findings, *f)
	}

	// No file found.
	if len(findings) == 0 {
		f, err := finding.NewWith(fs, probeID, fmt.Sprintf("no %s found", filetype),
			nil, notFoundOutcome)
		if err != nil {
			return nil, probeID, fmt.Errorf("create finding: %w", err)
		}
		f = f.WithRemediationMetadata(metadata)
		findings = append(findings, *f)
	}

	return findings, probeID, nil
}
