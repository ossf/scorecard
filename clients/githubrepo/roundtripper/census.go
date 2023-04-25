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

	"go.opencensus.io/plugin/ochttp"
	opencensusstats "go.opencensus.io/stats"
	"go.opencensus.io/tag"

	sce "github.com/ossf/scorecard/v4/errors"
	"github.com/ossf/scorecard/v4/stats"
)

const fromCacheHeader = "X-From-Cache"

// MakeCensusTransport wraps input Roundtripper with monitoring logic.
func MakeCensusTransport(innerTransport http.RoundTripper) http.RoundTripper {
	return &ochttp.Transport{
		Base: &censusTransport{
			innerTransport: innerTransport,
		},
	}
}

// censusTransport is a monitoring aware http.Transport.
type censusTransport struct {
	innerTransport http.RoundTripper
}

// Roundtrip handles context update and measurement recording.
func (ct *censusTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	ctx, err := tag.New(r.Context(), tag.Upsert(stats.RequestTag, "requested"))
	if err != nil {
		return nil, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("tag.New: %v", err))
	}

	r = r.WithContext(ctx)
	resp, err := ct.innerTransport.RoundTrip(r)
	if err != nil {
		return nil, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("innerTransport.RoundTrip: %v", err))
	}
	if resp.Header.Get(fromCacheHeader) != "" {
		ctx, err = tag.New(ctx, tag.Upsert(stats.RequestTag, fromCacheHeader))
		if err != nil {
			return nil, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("tag.New: %v", err))
		}
	}
	opencensusstats.Record(ctx, stats.HTTPRequests.M(1))
	return resp, nil
}
