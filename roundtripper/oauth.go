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
	"context"
	"fmt"
	"net/http"
	"sync/atomic"

	"golang.org/x/oauth2"
)

func MakeOAuthTransport(ctx context.Context, accessTokens []string) http.RoundTripper {
	tokenSource := &tokenPool{
		tokens: makeTokenAccessor(accessTokens),
	}
	return oauth2.NewClient(ctx, tokenSource).Transport
}

type tokenAccessor interface {
	next() (*oauth2.Token, error)
}

type tokenPool struct {
	tokens tokenAccessor
}

func (pool *tokenPool) Token() (*oauth2.Token, error) {
	token, err := pool.tokens.next()
	if err != nil {
		return token, fmt.Errorf("error during Next(): %w", err)
	}
	return token, nil
}

func makeTokenAccessor(accessTokens []string) tokenAccessor {
	return &roundRobinAccessor{
		accessTokens: accessTokens,
	}
}

type roundRobinAccessor struct {
	accessTokens []string
	counter      int64
}

func (roundRobin *roundRobinAccessor) next() (*oauth2.Token, error) {
	c := atomic.AddInt64(&roundRobin.counter, 1)
	// not locking it because it is never modified
	l := len(roundRobin.accessTokens)
	index := c % int64(l)
	return &oauth2.Token{
		AccessToken: roundRobin.accessTokens[index],
	}, nil
}
