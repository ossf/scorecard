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
	"errors"
	"testing"

	"gocloud.dev/pubsub"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/ossf/scorecard/v4/cron/internal/data"
)

var repo1 = "repo1"

type mockReceiver struct {
	msg           *pubsub.Message
	errOnReceive  error
	errOnShutdown error
}

func (mock *mockReceiver) Receive(ctx context.Context) (*pubsub.Message, error) {
	return mock.msg, mock.errOnReceive
}

func (mock *mockReceiver) Shutdown(ctx context.Context) error {
	return mock.errOnShutdown
}

func TestSubscriber(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		msg              *data.ScorecardBatchRequest
		errOnReceive     error
		errOnShutdown    error
		name             string
		hasErrOnReceive  bool
		hasErrOnShutdown bool
	}{
		{
			name: "Basic",
			msg: &data.ScorecardBatchRequest{
				Repos: []*data.Repo{
					{
						Url: &repo1,
					},
				},
			},
		},
		{
			name:            "ReceiveFails",
			hasErrOnReceive: true,
			//nolint: goerr113
			errOnReceive: errors.New("mock Receive failure"),
		},
		{
			name: "ShutdownFails",
			msg: &data.ScorecardBatchRequest{
				Repos: []*data.Repo{
					{
						Url: &repo1,
					},
				},
			},
			hasErrOnShutdown: true,
			//nolint: goerr113
			errOnShutdown: errors.New("mock Shutdown close"),
		},
	}
	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			msgBody, err := protojson.Marshal(testcase.msg)
			if err != nil {
				t.Errorf("testcase parsing failed during protojson.Marshal: %v", err)
			}
			subscriber := gocloudSubscriber{
				ctx: ctx,
				subscription: &mockReceiver{
					msg:           &pubsub.Message{Body: msgBody},
					errOnReceive:  testcase.errOnReceive,
					errOnShutdown: testcase.errOnShutdown,
				},
			}
			receivedMsg, err := subscriber.SynchronousPull()
			if testcase.hasErrOnReceive && receivedMsg != nil && err != nil {
				t.Errorf("test failed: expected - (nil, nil), got - (%q, %v)", receivedMsg, err)
			}
			if !proto.Equal(receivedMsg, testcase.msg) {
				t.Errorf("test failed: expected - %v, got %v", testcase.msg, receivedMsg)
			}
			err = subscriber.Close()
			if testcase.hasErrOnShutdown && !errors.Is(err, testcase.errOnShutdown) {
				t.Errorf("test failed: expected - %v, got - %v", testcase.errOnShutdown, err)
			}
		})
	}
}
