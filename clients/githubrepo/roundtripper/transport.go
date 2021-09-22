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

package roundtripper

import (
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"

	githubstats "github.com/ossf/scorecard/v2/clients/githubrepo/stats"
)

const expiryTimeInSec = 30

// MakeGitHubTransport wraps input RoundTripper with GitHub authorization logic.
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
	next() (uint64, string)
	release(uint64)
}

func (gt *githubTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	index, token := gt.tokens.next()
	defer gt.tokens.release(index)

	ctx, err := tag.New(r.Context(), tag.Upsert(githubstats.TokenIndex, fmt.Sprint(index)))
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

func makeTokenAccessor(accessTokens []string) tokenAccessor {
	return &roundRobinAccessor{
		accessTokens: accessTokens,
		accessState:  make([]int64, len(accessTokens)),
	}
}

type roundRobinAccessor struct {
	accessTokens []string
	accessState  []int64
	counter      uint64
}

func (roundRobin *roundRobinAccessor) next() (uint64, string) {
	c := atomic.AddUint64(&roundRobin.counter, 1)
	l := len(roundRobin.accessTokens)
	index := c % uint64(l)

	// If selected accessToken is unavailable, wait.
	for !atomic.CompareAndSwapInt64(&roundRobin.accessState[index], 0, time.Now().Unix()) {
		currVal := roundRobin.accessState[index]
		expired := time.Now().After(time.Unix(currVal, 0).Add(expiryTimeInSec * time.Second))
		if !expired {
			continue
		}
		if atomic.CompareAndSwapInt64(&roundRobin.accessState[index], currVal, time.Now().Unix()) {
			break
		}
	}
	return index, roundRobin.accessTokens[index]
}

func (roundRobin *roundRobinAccessor) release(index uint64) {
	atomic.SwapInt64(&roundRobin.accessState[index], 0)
}
