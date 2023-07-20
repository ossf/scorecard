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
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/stats"
)

const fromCacheHeader = "X-From-Cache"

// MakeCensusTransport wraps input Roundtripper with monitoring logic.
func MakeCensusTransport(innerTransport http.RoundTripper) http.RoundTripper {
	return otelhttp.NewTransport(
		&censusTransport{
			innerTransport: innerTransport,
		},
	)
}

// censusTransport is a monitoring aware http.Transport.
type censusTransport struct {
	innerTransport http.RoundTripper
}

// Roundtrip handles context update and measurement recording.
func (ct *censusTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	*r = *r.WithContext(context.WithValue(r.Context(), stats.RequestTag, "requested"))

	resp, err := ct.innerTransport.RoundTrip(r)
	if err != nil {
		return nil, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("innerTransport.RoundTrip: %v", err))
	}
	if resp.Header.Get(fromCacheHeader) != "" {
		*r = *r.WithContext(context.WithValue(r.Context(), stats.RequestTag, fromCacheHeader))
		if err != nil {
			return nil, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("tag.New: %v", err))
		}
	}
	stats.Metrics.HTTPRequests.Record(r.Context(), 1)
	return resp, nil
}
