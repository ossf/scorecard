// Copyright 2020 Security Scorecard Authors
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

package roundtripper

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	sce "github.com/ossf/scorecard/v2/errors"
)

// MakeRateLimitedTransport returns a RoundTripper which rate limits GitHub requests.
func MakeRateLimitedTransport(innerTransport http.RoundTripper, logger *zap.SugaredLogger) http.RoundTripper {
	return &rateLimitTransport{
		logger:         logger,
		innerTransport: innerTransport,
	}
}

// rateLimitTransport is a rate-limit aware http.Transport for Github.
type rateLimitTransport struct {
	logger         *zap.SugaredLogger
	innerTransport http.RoundTripper
}

// Roundtrip handles caching and ratelimiting of responses from GitHub.
func (gh *rateLimitTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, err := gh.innerTransport.RoundTrip(r)
	if err != nil {
		//nolint:wrapcheck
		return nil, sce.Create(sce.ErrScorecardInternal, fmt.Sprintf("innerTransport.RoundTrip: %v", err))
	}
	rateLimit := resp.Header.Get("X-RateLimit-Remaining")
	remaining, err := strconv.Atoi(rateLimit)
	if err != nil {
		return resp, nil
	}

	if remaining <= 0 {
		reset, err := strconv.Atoi(resp.Header.Get("X-RateLimit-Reset"))
		if err != nil {
			return resp, nil
		}

		duration := time.Until(time.Unix(int64(reset), 0))
		gh.logger.Warnf("Rate limit exceeded. Waiting %s to retry...", duration)

		// Retry
		time.Sleep(duration)
		gh.logger.Warnf("Rate limit exceeded. Retrying...")
		return gh.RoundTrip(r)
	}

	return resp, nil
}
