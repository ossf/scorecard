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
	"strings"
	"time"

	pubsub "cloud.google.com/go/pubsub/apiv1"
	pubsubpb "google.golang.org/genproto/googleapis/pubsub/v1"

	"github.com/ossf/scorecard/v4/cron/internal/data"
)

const (
	maxMessagesToPull         = 1
	ackDeadlineExtensionInSec = 600
	gracePeriodInSec          = 60
	gcpPubsubPrefix           = "gcppubsub://"
)

type gcsSubscriber struct {
	ctx             context.Context
	done            chan bool
	client          *pubsub.SubscriberClient
	subscriptionURL string
	recvdAckID      string
}

func createGCSSubscriber(ctx context.Context, subscriptionURL string) (*gcsSubscriber, error) {
	client, err := pubsub.NewSubscriberClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("error during NewSubscriberClient: %w", err)
	}

	return &gcsSubscriber{
		ctx:             ctx,
		client:          client,
		subscriptionURL: strings.TrimPrefix(subscriptionURL, gcpPubsubPrefix),
		done:            make(chan bool),
	}, nil
}

func (subscriber *gcsSubscriber) extendAckDeadline() {
	delay := 0 * time.Second
	for {
		select {
		case <-subscriber.ctx.Done():
			return
		case <-subscriber.done:
			return
		case <-time.After(delay):
			ackDeadline := ackDeadlineExtensionInSec * time.Second
			err := subscriber.client.ModifyAckDeadline(subscriber.ctx, &pubsubpb.ModifyAckDeadlineRequest{
				Subscription:       subscriber.subscriptionURL,
				AckIds:             []string{subscriber.recvdAckID},
				AckDeadlineSeconds: int32(ackDeadline.Seconds()),
			})
			if err != nil {
				log.Fatal(err)
			}
			delay = ackDeadline - gracePeriodInSec*time.Second
		}
	}
}

func (subscriber *gcsSubscriber) SynchronousPull() (*data.ScorecardBatchRequest, error) {
	numReceivedMessages := 0
	var msgToProcess *pubsubpb.ReceivedMessage
	subscriber.done = make(chan bool)
	// Block indefinitely until a message is received.
	for msgToProcess == nil {
		result, err := subscriber.client.Pull(subscriber.ctx, &pubsubpb.PullRequest{
			Subscription: subscriber.subscriptionURL,
			MaxMessages:  maxMessagesToPull,
		})
		if err != nil {
			log.Printf("error during Recieive: %v", err)
			return nil, nil
		}
		numReceivedMessages = len(result.ReceivedMessages)
		if numReceivedMessages > 0 {
			msgToProcess = result.GetReceivedMessages()[0]
		} else {
			//nolint:gomnd
			time.Sleep(30 * time.Second)
		}
	}

	// Sanity check.
	if numReceivedMessages > maxMessagesToPull {
		log.Fatalf("expected to receive max %d messages, got %d", maxMessagesToPull, numReceivedMessages)
	}

	subscriber.recvdAckID = msgToProcess.AckId
	// Continuously notify the server that processing is still happening on this message.
	go subscriber.extendAckDeadline()

	return parseJSONToRequest(msgToProcess.GetMessage().GetData())
}

func (subscriber *gcsSubscriber) Ack() {
	err := subscriber.client.Acknowledge(subscriber.ctx, &pubsubpb.AcknowledgeRequest{
		Subscription: subscriber.subscriptionURL,
		AckIds:       []string{subscriber.recvdAckID},
	})
	close(subscriber.done)
	if err != nil {
		log.Fatal(err)
	}
}

func (subscriber *gcsSubscriber) Nack() {
	// Stop extending Ack deadline.
	close(subscriber.done)
}

func (subscriber *gcsSubscriber) Close() error {
	close(subscriber.done)
	err := subscriber.client.Close()
	if err != nil {
		return fmt.Errorf("error during Close: %w", err)
	}
	return nil
}
