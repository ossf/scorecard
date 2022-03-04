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
	"fmt"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"

	"github.com/ossf/scorecard/v4/checker"
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
	rawData, err := raw.WebHook(c)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckWebHooks, e)
	}

	if len(rawData.Webhook) < 1 {
		return checker.CreateMaxScoreResult(CheckWebHooks, "no webhooks defined")
	}

	hasNoSecretCount := 0
	for _, hook := range rawData.Webhook {
		if !hook.UsesAuthSecret {
			c.Dlogger.Warn(&checker.LogMessage{
				Path: fmt.Sprintf("https://%s/settings/hooks/%d", c.RepoClient.URI(), hook.ID),
				Type: checker.FileTypeURL,
				Text: "Webhook with no secret configured",
			})
			hasNoSecretCount++
		}
	}

	generateWebhookTable(c.RepoClient.URI(), rawData.Webhook)

	if hasNoSecretCount == 0 {
		return checker.CreateMaxScoreResult(CheckWebHooks, fmt.Sprintf("all %d hook(s) have a secret configured", len(rawData.Webhook)))
	}

	if len(rawData.Webhook) == hasNoSecretCount {
		return checker.CreateMinScoreResult(CheckWebHooks, fmt.Sprintf("%d hook(s) do not have a secret configured", len(rawData.Webhook)))
	}

	return checker.CreateProportionalScoreResult(CheckWebHooks,
		fmt.Sprintf("%d out of %d hook(s) with no secrets configured detected", hasNoSecretCount, len(rawData.Webhook)), hasNoSecretCount, len(rawData.Webhook))
}

func generateWebhookTable(repo string, data []checker.WebhookData) {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetHeader([]string{"GitHub Webhook", "Uses Auth Secret"})

	for _, hook := range data {
		table.Append([]string{fmt.Sprintf("https://%s/settings/hooks/%d", repo, hook.ID), strconv.FormatBool(hook.UsesAuthSecret)})
		// https: //github.com/cpanato/testing-ci-providers/settings/hooks/289347313
	}

	table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	table.SetCenterSeparator("|")
	table.Render()

	fmt.Println(tableString.String())
}
