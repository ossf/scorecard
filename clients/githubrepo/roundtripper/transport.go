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

package roundtripper

import (
	"fmt"
	"net/http"
	"strconv"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"

	"github.com/ossf/scorecard/v4/clients/githubrepo/roundtripper/tokens"
	githubstats "github.com/ossf/scorecard/v4/clients/githubrepo/stats"
)

// makeGitHubTransport wraps input RoundTripper with GitHub authorization logic.
func makeGitHubTransport(innerTransport http.RoundTripper, accessor tokens.TokenAccessor) http.RoundTripper {
	return &githubTransport{
		innerTransport: innerTransport,
		tokens:         accessor,
	}
}

// githubTransport handles authorization using GitHub personal access tokens (PATs) during HTTP requests.
type githubTransport struct {
	innerTransport http.RoundTripper
	tokens         tokens.TokenAccessor
}

func (gt *githubTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	id, token := gt.tokens.Next()
	defer gt.tokens.Release(id)

	ctx, err := tag.New(r.Context(), tag.Upsert(githubstats.TokenIndex, fmt.Sprint(id)))
	if err != nil {
		return nil, fmt.Errorf("error updating context: %w", err)
	}
	*r = *r.WithContext(ctx)

	r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := gt.innerTransport.RoundTrip(r)
	if err != nil {
		return nil, fmt.Errorf("error in HTTP: %w", err)
	}

	ctx, err = tag.New(r.Context(), tag.Upsert(githubstats.ResourceType, resp.Header.Get("X-RateLimit-Resource")))
	if err != nil {
		return nil, fmt.Errorf("error updating context: %w", err)
	}
	remaining, err := strconv.Atoi(resp.Header.Get("X-RateLimit-Remaining"))
	if err == nil {
		stats.Record(ctx, githubstats.RemainingTokens.M(int64(remaining)))
	}

	return resp, nil
}
