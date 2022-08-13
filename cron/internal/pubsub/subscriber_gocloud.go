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

package pubsub

import (
	"context"
	"fmt"
	"log"

	"gocloud.dev/pubsub"
	// Needed to link in GCP drivers.
	_ "gocloud.dev/pubsub/gcppubsub"

	"github.com/ossf/scorecard/v4/cron/internal/data"
)

type receiver interface {
	Receive(ctx context.Context) (*pubsub.Message, error)
	Shutdown(ctx context.Context) error
}

type gocloudSubscriber struct {
	ctx          context.Context
	subscription receiver
	msg          *pubsub.Message
}

//nolint:unused,deadcode
func createGocloudSubscriber(ctx context.Context, subscriptionURL string) (*gocloudSubscriber, error) {
	subscription, err := pubsub.OpenSubscription(ctx, subscriptionURL)
	if err != nil {
		return nil, fmt.Errorf("error during pubsub.OpenSubscription: %w", err)
	}
	ret := gocloudSubscriber{
		ctx:          ctx,
		subscription: subscription,
	}
	return &ret, nil
}

func (subscriber *gocloudSubscriber) SynchronousPull() (*data.ScorecardBatchRequest, error) {
	msg, err := subscriber.subscription.Receive(subscriber.ctx)
	if err != nil {
		log.Printf("error during Receive: %v", err)
		return nil, nil
	}
	subscriber.msg = msg
	return parseJSONToRequest(msg.Body)
}

func (subscriber *gocloudSubscriber) Ack() {
	subscriber.msg.Ack()
}

func (subscriber *gocloudSubscriber) Nack() {
	subscriber.msg.Nack()
}

func (subscriber *gocloudSubscriber) Close() error {
	err := subscriber.subscription.Shutdown(subscriber.ctx)
	if err != nil {
		return fmt.Errorf("error during subscription.Shutdown: %w", err)
	}
	return nil
}
