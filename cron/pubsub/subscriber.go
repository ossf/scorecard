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
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/ossf/scorecard/v2/cron/data"
)

// ErrorInParse indicates there was an error while unmarshalling the protocol buffer message.
var ErrorInParse = errors.New("error during protojson.Unmarshal")

type Subscriber interface {
	SynchronousPull() (*data.ScorecardBatchRequest, error)
	Ack()
	Nack()
	Close() error
}

func CreateSubscriber(ctx context.Context, subscriptionURL string) (Subscriber, error) {
	return createGCSSubscriber(ctx, subscriptionURL)
}

func parseJSONToRequest(jsonData []byte) (*data.ScorecardBatchRequest, error) {
	ret := &data.ScorecardBatchRequest{}
	if err := protojson.Unmarshal(jsonData, ret); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrorInParse, err)
	}
	return ret, nil
}
