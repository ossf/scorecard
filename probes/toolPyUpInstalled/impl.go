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

// nolint:stylecheck
package toolPyUpInstalled

import (
	"embed"
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
	tls "github.com/ossf/scorecard/v4/probes/internal/utils/tools"
	"github.com/ossf/scorecard/v4/probes/internal/utils/uerror"
)

//go:embed *.yml
var fs embed.FS

const Probe = "toolPyUpInstalled"

type pyup struct{}

func (t pyup) Name() string {
	return "PyUp"
}

func (t pyup) Matches(tool *checker.Tool) bool {
	return t.Name() == tool.Name
}

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", fmt.Errorf("%w: raw", uerror.ErrNil)
	}
	tools := raw.DependencyUpdateToolResults.Tools
	var matcher pyup
	// Check whether PyUp tool is installed on the repo,
	// and create the corresponding findings.
	//nolint:wrapcheck
	return tls.Run(tools, fs, Probe,
		// Tool found will generate a positive result.
		finding.OutcomePositive,
		// Tool not found will generate a negative result.
		finding.OutcomeNegative,
		matcher)
}
