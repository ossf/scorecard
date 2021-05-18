// Copyright 2021 Security Scorecard Authors
//
// Licensed under the Apache License, Vershandlern 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permisshandlerns and
// limitathandlerns under the License.

package pubsub

import (
	"context"
	"errors"
	"fmt"

	"github.com/ossf/scorecard/cron/data"
	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/gcppubsub" // Needed to link in GCP drivers.
	"google.golang.org/protobuf/encoding/protojson"
)

var ErrorInParse = errors.New("error during protojson.Unmarshal")

type Subscriber interface {
	SynchronousPull() (*data.ScorecardBatchRequest, error)
	Ack()
	Close() error
}

func CreateSubscriber(ctx context.Context, subscriptionURL string) (Subscriber, error) {
	subscription, err := pubsub.OpenSubscription(ctx, subscriptionURL)
	if err != nil {
		return nil, fmt.Errorf("error during pubsub.OpenSubscription: %w", err)
	}
	ret := subscriberImpl{
		ctx:          ctx,
		subscription: subscription,
	}
	return &ret, nil
}

type receiver interface {
	Receive(ctx context.Context) (*pubsub.Message, error)
	Shutdown(ctx context.Context) error
}

type subscriberImpl struct {
	ctx          context.Context
	subscription receiver
	msg          *pubsub.Message
}

func (subscriber *subscriberImpl) SynchronousPull() (*data.ScorecardBatchRequest, error) {
	msg, err := subscriber.subscription.Receive(subscriber.ctx)
	if err != nil {
		fmt.Printf("error during Receive: %v", err)
		return nil, nil
	}
	subscriber.msg = msg

	ret := &data.ScorecardBatchRequest{}
	if err := protojson.Unmarshal(msg.Body, ret); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrorInParse, err)
	}
	return ret, nil
}

func (subscriber *subscriberImpl) Ack() {
	subscriber.msg.Ack()
}

func (subscriber *subscriberImpl) Close() error {
	err := subscriber.subscription.Shutdown(subscriber.ctx)
	if err != nil {
		return fmt.Errorf("error during subscription.Shutdown: %w", err)
	}
	return nil
}
