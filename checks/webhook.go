// Copyright 2022 Security Scorecard Authors
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

package checks

import (
	"os"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/evaluation"
	"github.com/ossf/scorecard/v4/checks/raw"
	sce "github.com/ossf/scorecard/v4/errors"
)

const (
	// CheckWebHooks is the registered name for WebHooks.
	CheckWebHooks = "Webhooks"
)

//nolint:gochecknoinits
func init() {
	if err := registerCheck(CheckWebHooks, WebHooks, nil); err != nil {
		// this should never happen
		panic(err)
	}
}

// WebHooks run Webhooks check.
func WebHooks(c *checker.CheckRequest) checker.CheckResult {
	// TODO: remove this check when v6 is released
	_, enabled := os.LookupEnv("SCORECARD_V6")
	if !enabled {
		c.Dlogger.Warn(&checker.LogMessage{
			Text: "SCORECARD_V6 is not set, not running the Webhook check",
		})

		e := sce.WithMessage(sce.ErrorUnsupportedCheck, "SCORECARD_V6 is not set, not running the Webhook check")
		return checker.CreateInconclusiveResult(CheckWebHooks, e.Error())
	}

	rawData, err := raw.WebHook(c)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckWebHooks, e)
	}

	// Set the raw results.
	if c.RawResults != nil {
		c.RawResults.WebhookResults = rawData
	}

	// Return the score evaluation.
	return evaluation.Webhooks(CheckWebHooks, c.Dlogger, &rawData)
}
