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

// Package pubsub handles interactions with PubSub framework.
package pubsub

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"gocloud.dev/pubsub"
	// Needed to link in GCP drivers.
	_ "gocloud.dev/pubsub/gcppubsub"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/ossf/scorecard/v4/cron/data"
)

type publishError struct {
	count uint64
}

func (pe publishError) Error() string {
	return fmt.Sprintf("total errors when publishing: %d", pe.count)
}

// Publisher interface is used to publish cron job requests to PubSub.
type Publisher interface {
	Publish(request *data.ScorecardBatchRequest) error
	Close() error
}

// CreatePublisher returns an implementation of the Publisher interface.
func CreatePublisher(ctx context.Context, topicURL string) (Publisher, error) {
	ret := publisherImpl{}
	topic, err := pubsub.OpenTopic(ctx, topicURL)
	if err != nil {
		return &ret, fmt.Errorf("error from pubsub.OpenTopic: %w", err)
	}
	return &publisherImpl{
		ctx:   ctx,
		topic: topic,
	}, nil
}

type sender interface {
	Send(ctx context.Context, msg *pubsub.Message) error
}

type publisherImpl struct {
	ctx         context.Context
	topic       sender
	wg          sync.WaitGroup
	totalErrors uint64
}

func (publisher *publisherImpl) Publish(request *data.ScorecardBatchRequest) error {
	msg, err := protojson.Marshal(request)
	if err != nil {
		return fmt.Errorf("error from protojson.Marshal: %w", err)
	}

	publisher.wg.Add(1)
	go func() {
		defer publisher.wg.Done()
		err := publisher.topic.Send(publisher.ctx, &pubsub.Message{
			Body: msg,
		})
		if err != nil {
			log.Printf("Error when publishing message %s: %v", msg, err)
			atomic.AddUint64(&publisher.totalErrors, 1)
			return
		}
		log.Print("Successfully published message")
	}()
	return nil
}

func (publisher *publisherImpl) Close() error {
	publisher.wg.Wait()
	if publisher.totalErrors > 0 {
		return publishError{count: publisher.totalErrors}
	}
	return nil
}
