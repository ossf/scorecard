// Copyright 2020 OpenSSF Scorecard Authors
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
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/log"
)

// MakeRateLimitedTransport returns a RoundTripper which rate limits GitHub requests.
func MakeRateLimitedTransport(innerTransport http.RoundTripper, logger *log.Logger) http.RoundTripper {
	return &rateLimitTransport{
		logger:         logger,
		innerTransport: innerTransport,
	}
}

// rateLimitTransport is a rate-limit aware http.Transport for Github.
type rateLimitTransport struct {
	logger         *log.Logger
	innerTransport http.RoundTripper
}

// Roundtrip handles caching and ratelimiting of responses from GitHub.
func (gh *rateLimitTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, err := gh.innerTransport.RoundTrip(r)
	if err != nil {
		return nil, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("innerTransport.RoundTrip: %v", err))
	}

	if resp.StatusCode == http.StatusForbidden {
		data, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewBuffer(data))
		if err != nil || data == nil {
			return resp, nil
		}
		if bytes.Contains(data, []byte("You have exceeded a secondary rate limit")) {
			retryValue := resp.Header.Get("Retry-After")
			retryAfter, err := strconv.Atoi(retryValue)
			if err != nil {
				retryAfter = 60
			}
			gh.logger.Info(fmt.Sprintf("Secondary rate limit exceeded. Waiting %d to retry...", retryAfter))
			time.Sleep(time.Duration(retryAfter) * time.Second)
			gh.logger.Info("Secondary rate limit exceeded. Retrying...")
			return gh.RoundTrip(r)
		}
		gh.logger.Info("Non secondary rate limit 403")
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
		// TODO(log): Previously Warn. Consider logging an error here.
		gh.logger.Info(fmt.Sprintf("Rate limit exceeded. Waiting %s to retry...", duration))

		// Retry
		time.Sleep(duration)
		// TODO(log): Previously Warn. Consider logging an error here.
		gh.logger.Info("Rate limit exceeded. Retrying...")
		return gh.RoundTrip(r)
	}

	return resp, nil
}
