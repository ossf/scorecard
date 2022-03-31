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

package raw

import (
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
)

// WebHook retrieves the raw data for the WebHooks check.
func WebHook(c *checker.CheckRequest) (checker.WebhooksData, error) {
	hooksResp, err := c.RepoClient.ListWebhooks()
	if err != nil {
		return checker.WebhooksData{},
			sce.WithMessage(sce.ErrScorecardInternal, "Client.Repositories.ListWebhooks")
	}

	if len(hooksResp) < 1 {
		return checker.WebhooksData{}, nil
	}

	hooks := []checker.WebhookData{}
	for _, hook := range hooksResp {
		v := checker.WebhookData{
			ID:             hook.ID,
			UsesAuthSecret: hook.UsesAuthSecret,
			Path:           fmt.Sprintf("https://%s/settings/hooks/%d", c.RepoClient.URI(), hook.ID),
			// Note: add fields if needed.
		}
		hooks = append(hooks, v)
	}

	return checker.WebhooksData{Webhook: hooks}, nil
}
