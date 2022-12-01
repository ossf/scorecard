// Copyright 2021 OpenSSF Scorecard Authors
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
	sce "github.com/ossf/scorecard/v4/errors"
)

// Webhooks applies the score policy for the Webhooks check.
func Webhooks(name string, dl checker.DetailLogger,
	r *checker.WebhooksData,
) checker.CheckResult {
	if r == nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, "empty raw data")
		return checker.CreateRuntimeErrorResult(name, e)
	}

	if len(r.Webhooks) < 1 {
		return checker.CreateMaxScoreResult(name, "no webhooks defined")
	}

	hasNoSecretCount := 0
	for _, hook := range r.Webhooks {
		if !hook.UsesAuthSecret {
			dl.Warn(&checker.LogMessage{
				Path: hook.Path,
				Type: checker.FileTypeURL,
				Text: "Webhook with no secret configured",
			})
			hasNoSecretCount++
		}
	}

	if hasNoSecretCount == 0 {
		return checker.CreateMaxScoreResult(name, fmt.Sprintf("all %d hook(s) have a secret configured", len(r.Webhooks)))
	}

	if len(r.Webhooks) == hasNoSecretCount {
		return checker.CreateMinScoreResult(name, fmt.Sprintf("%d hook(s) do not have a secret configured", len(r.Webhooks)))
	}

	return checker.CreateProportionalScoreResult(name,
		fmt.Sprintf("%d/%d hook(s) with no secrets configured detected",
			hasNoSecretCount, len(r.Webhooks)), hasNoSecretCount, len(r.Webhooks))
}
