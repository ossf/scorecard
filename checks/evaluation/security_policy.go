// Copyright 2021 Security Scorecard Authors
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

package evaluation

import (
	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
)

// SecurityPolicy applies the score policy for the Security-Policy check.
func SecurityPolicy(name string, dl checker.DetailLogger, r *checker.SecurityPolicyData) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	// Apply the policy evaluation.
	if r.Files == nil || len(r.Files) == 0 {
		// If the file is null or has zero lengths, directly return as not detected.
		return checker.CreateMinScoreResult(name, "security policy file not detected")
	}

	for _, f := range r.Files {
		msg := checker.LogMessage{
			Path:   f.Path,
			Type:   f.Type,
			Offset: f.Offset,
		}
		if msg.Type == checker.FileTypeURL {
			msg.Text = "security policy detected in org repo"
		} else {
			msg.Text = "security policy detected in current repo"
		}
		dl.Info(&msg)
	}
	return checker.CreateMaxScoreResult(name, "security policy file detected")
}
