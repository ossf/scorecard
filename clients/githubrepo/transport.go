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

package githubrepo

import (
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

func MakeGitHubTransport(innerTransport http.RoundTripper, accessTokens []string) http.RoundTripper {
	return &githubTransport{
		innerTransport: innerTransport,
		tokens:         makeTokenAccessor(accessTokens),
	}
}

// githubTransport handles authorization using GitHub personal access tokens (PATs) during HTTP requests.
type githubTransport struct {
	innerTransport http.RoundTripper
	tokens         tokenAccessor
}

type tokenAccessor interface {
	next(r *http.Request) (string, error)
}

func (gt *githubTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	token, err := gt.tokens.next(r)
	if err != nil {
		return nil, fmt.Errorf("error getting Github token: %w", err)
	}
	r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := gt.innerTransport.RoundTrip(r)
	if err != nil {
		return nil, fmt.Errorf("error in HTTP: %w", err)
	}
	ctx, err := tag.New(r.Context(), tag.Upsert(ResourceType, resp.Header.Get("X-RateLimit-Resource")))
	if err != nil {
		return nil, fmt.Errorf("error updating context: %w", err)
	}
	remaining, err := strconv.Atoi(resp.Header.Get("X-RateLimit-Remaining"))
	if err == nil {
		stats.Record(ctx, RemainingTokens.M(int64(remaining)))
	}
	return resp, nil
}

func makeTokenAccessor(accessTokens []string) tokenAccessor {
	return &roundRobinAccessor{
		accessTokens: accessTokens,
	}
}

type roundRobinAccessor struct {
	accessTokens []string
	counter      uint64
}

func (roundRobin *roundRobinAccessor) next(r *http.Request) (string, error) {
	c := atomic.AddUint64(&roundRobin.counter, 1)
	l := len(roundRobin.accessTokens)
	index := c % uint64(l)

	ctx, err := tag.New(r.Context(), tag.Upsert(TokenIndex, fmt.Sprint(index)))
	if err != nil {
		return "", fmt.Errorf("error updating context: %w", err)
	}
	*r = *r.WithContext(ctx)

	return roundRobin.accessTokens[index], nil
}
