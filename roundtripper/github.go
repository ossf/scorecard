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
	"sync"
	"sync/atomic"
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
	next() string
}

func (gt *githubTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", gt.tokens.next()))
	resp, err := gt.innerTransport.RoundTrip(r)
	if err != nil {
		return nil, fmt.Errorf("error in HTTP: %w", err)
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
	counter      int64
	mu           sync.Mutex
}

func (roundRobin *roundRobinAccessor) next() string {
	roundRobin.mu.Lock()
	defer roundRobin.mu.Unlock()

	c := atomic.AddInt64(&roundRobin.counter, 1)
	l := len(roundRobin.accessTokens)
	index := c % int64(l)
	return roundRobin.accessTokens[index]
}
