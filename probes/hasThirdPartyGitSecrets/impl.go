// Copyright 2025 OpenSSF Scorecard Authors
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

package hasThirdPartyGitSecrets

import (
	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/internal/checknames"
	"github.com/ossf/scorecard/v5/internal/probes"
	"github.com/ossf/scorecard/v5/probes/internal/utils/uerror"
)

const Probe = "hasThirdPartyGitSecrets"

func init() {
	probes.MustRegister(Probe, Run, []checknames.CheckName{checknames.SecretScanning})
}

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	if raw == nil {
		return nil, "", uerror.ErrNil
	}

	f := finding.Finding{
		Probe:   Probe,
		Outcome: finding.OutcomeFalse,
		Message: "git-secrets not detected in CI",
	}
	if raw.SecretScanningResults.ThirdPartyGitSecrets {
		f.Outcome = finding.OutcomeTrue
		if p := first(raw.SecretScanningResults.ThirdPartyGitSecretsPaths); p != "" {
			f.Message = "git-secrets found at " + p
			f.Location = &finding.Location{Type: finding.FileTypeSource, Path: p}
		} else {
			f.Message = "git-secrets detected in CI"
		}
	}
	return []finding.Finding{f}, Probe, nil
}

func first(xs []string) string {
	if len(xs) == 0 {
		return ""
	}
	return xs[0]
}
