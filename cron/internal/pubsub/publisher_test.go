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

package pubsub

import (
	"context"
	"fmt"
	"testing"

	"gocloud.dev/pubsub"

	"github.com/ossf/scorecard/v4/cron/data"
)

type mockSucceedTopic struct{}

func (topic *mockSucceedTopic) Send(ctx context.Context, msg *pubsub.Message) error {
	return nil
}

type mockFailTopic struct{}

func (topic *mockFailTopic) Send(ctx context.Context, msg *pubsub.Message) error {
	//nolint:goerr113
	return fmt.Errorf("mockFailTopic failed to send")
}

func TestPublish(t *testing.T) {
	t.Parallel()
	//nolint:govet
	testcases := []struct {
		numErrors uint64
		name      string
		errorMsg  string
		hasError  bool
		topic     sender
	}{
		{
			name:      "SendFails",
			topic:     &mockFailTopic{},
			hasError:  true,
			numErrors: 1,
			errorMsg:  "",
		},
		{
			name:     "SendSucceeds",
			topic:    &mockSucceedTopic{},
			hasError: false,
		},
	}
	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			publisher := publisherImpl{
				ctx:   ctx,
				topic: testcase.topic,
			}
			request := data.ScorecardBatchRequest{
				Repos: []*data.Repo{
					{
						Url: &repo1,
					},
				},
			}
			if err := publisher.Publish(&request); err != nil {
				t.Errorf("Failed to parse message: %v", err)
			}
			err := publisher.Close()
			if (err == nil) == testcase.hasError {
				t.Errorf("Test failed. Expected: %t got: %v", testcase.hasError, err)
			}
			if testcase.hasError && testcase.numErrors != publisher.totalErrors {
				t.Errorf("Test failed. Expected numErrors: %d, got: %d", testcase.numErrors, publisher.totalErrors)
			}
		})
	}
}
