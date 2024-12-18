// Copyright 2024 OpenSSF Scorecard Authors
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

package azuredevopsrepo

import (
	"context"
	"sync"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/servicehooks"

	"github.com/ossf/scorecard/v5/clients"
)

var webHooksConsumerID = "webHooks"

type servicehooksHandler struct {
	ctx                context.Context
	once               *sync.Once
	repourl            *Repo
	servicehooksClient servicehooks.Client
	listSubscriptions  fnListSubscriptions
	errSetup           error
	webhooks           []clients.Webhook
}

type fnListSubscriptions func(
	ctx context.Context,
	args servicehooks.ListSubscriptionsArgs,
) (*[]servicehooks.Subscription, error)

func (s *servicehooksHandler) init(ctx context.Context, repourl *Repo) {
	s.ctx = ctx
	s.once = new(sync.Once)
	s.repourl = repourl
	s.errSetup = nil
	s.webhooks = nil
	s.listSubscriptions = s.servicehooksClient.ListSubscriptions
}

func (s *servicehooksHandler) setup() error {
	s.once.Do(func() {
		args := servicehooks.ListSubscriptionsArgs{
			ConsumerId: &webHooksConsumerID,
		}
		subscriptions, err := s.listSubscriptions(s.ctx, args)
		if err != nil {
			s.errSetup = err
			return
		}

		for i := range *subscriptions {
			subscription := (*subscriptions)[i]

			usesAuthSecret := false
			if subscription.ConsumerInputs != nil {
				_, usesAuthSecret = (*subscription.ConsumerInputs)["basicAuthPassword"]
			}

			s.webhooks = append(s.webhooks, clients.Webhook{
				// Azure DevOps uses uuid.UUID for ID, but Scorecard expects int64
				// ID:             *subscription.Id,
				Path:           (*subscription.ConsumerInputs)["url"],
				UsesAuthSecret: usesAuthSecret,
			})
		}
	})
	return s.errSetup
}

func (s *servicehooksHandler) listWebhooks() ([]clients.Webhook, error) {
	if err := s.setup(); err != nil {
		return nil, err
	}
	return s.webhooks, nil
}
