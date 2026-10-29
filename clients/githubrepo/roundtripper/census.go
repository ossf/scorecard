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
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/stats"
)

const fromCacheHeader = "X-From-Cache"

// MakeOTelTransport wraps input RoundTripper with OpenTelemetry monitoring logic.
func MakeOTelTransport(innerTransport http.RoundTripper) http.RoundTripper {
	return otelhttp.NewTransport(&otelTransport{
		innerTransport: innerTransport,
	})
}

// MakeCensusTransport wraps input RoundTripper with monitoring logic.
//
// Deprecated: Use MakeOTelTransport instead.
func MakeCensusTransport(innerTransport http.RoundTripper) http.RoundTripper {
	return MakeOTelTransport(innerTransport)
}

// otelTransport is a monitoring-aware http.Transport.
type otelTransport struct {
	innerTransport http.RoundTripper
}

// RoundTrip handles context update and measurement recording.
func (ot *otelTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	tagVal := "requested"
	resp, err := ot.innerTransport.RoundTrip(r)
	if err != nil {
		return nil, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("innerTransport.RoundTrip: %v", err))
	}
	if resp.Header.Get(fromCacheHeader) != "" {
		tagVal = fromCacheHeader
	}
	stats.RecordHTTPRequests(r.Context(), 1, stats.RequestTag.String(tagVal))
	return resp, nil
}
