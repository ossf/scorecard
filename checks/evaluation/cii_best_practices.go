// Copyright Security Scorecard Authors
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
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

// Note: exported for unit tests.
const (
	silverScore = 7
	// Note: if this value is changed, please update the action's threshold score
	// https://github.com/ossf/scorecard-action/blob/main/policies/template.yml#L61.
	passingScore    = 5
	inProgressScore = 2
)

// CIIBestPractices applies the score policy for the CIIBestPractices check.
func CIIBestPractices(name string, dl checker.DetailLogger, r *checker.CIIBestPracticesData) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	var results checker.CheckResult
	switch r.Badge {
	case clients.NotFound:
		results = checker.CreateMinScoreResult(name, "no effort to earn an OpenSSF best practices badge detected")
	case clients.InProgress:
		msg := fmt.Sprintf("badge detected: %v", r.Badge)
		results = checker.CreateResultWithScore(name, msg, inProgressScore)
	case clients.Passing:
		msg := fmt.Sprintf("badge detected: %v", r.Badge)
		results = checker.CreateResultWithScore(name, msg, passingScore)
	case clients.Silver:
		msg := fmt.Sprintf("badge detected: %v", r.Badge)
		results = checker.CreateResultWithScore(name, msg, silverScore)
	case clients.Gold:
		msg := fmt.Sprintf("badge detected: %v", r.Badge)
		results = checker.CreateMaxScoreResult(name, msg)
	case clients.Unknown:
		e := sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("unsupported badge: %v", r.Badge))
		results = checker.CreateRuntimeErrorResult(name, e)
	}
	return results
}
