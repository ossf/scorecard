// Copyright 2022 OpenSSF Scorecard Authors
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

package githubrepo

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/google/go-github/v38/github"

	"github.com/ossf/scorecard/v4/clients"
)

type webhookHandler struct {
	ghClient *github.Client
	once     *sync.Once
	ctx      context.Context
	errSetup error
	repourl  *repoURL
	webhook  []clients.Webhook
}

func (handler *webhookHandler) init(ctx context.Context, repourl *repoURL) {
	handler.ctx = ctx
	handler.repourl = repourl
	handler.errSetup = nil
	handler.once = new(sync.Once)
}

func (handler *webhookHandler) setup() error {
	handler.once.Do(func() {
		if !strings.EqualFold(handler.repourl.commitSHA, clients.HeadSHA) {
			handler.errSetup = fmt.Errorf("%w: ListWebHooks only supported for HEAD queries", clients.ErrUnsupportedFeature)
			return
		}
		hooks, _, err := handler.ghClient.Repositories.ListHooks(
			handler.ctx, handler.repourl.owner, handler.repourl.repo, &github.ListOptions{})
		if err != nil {
			handler.errSetup = fmt.Errorf("error during ListHooks: %w", err)
			return
		}

		for _, hook := range hooks {
			repoHook := clients.Webhook{
				ID:             hook.GetID(),
				UsesAuthSecret: getAuthSecret(hook.Config),
			}
			handler.webhook = append(handler.webhook, repoHook)
		}
		handler.errSetup = nil
	})
	return handler.errSetup
}

func getAuthSecret(config map[string]interface{}) bool {
	if val, ok := config["secret"]; ok {
		if val != nil {
			return true
		}
	}

	return false
}

func (handler *webhookHandler) listWebhooks() ([]clients.Webhook, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during webhookHandler.setup: %w", err)
	}
	return handler.webhook, nil
}
