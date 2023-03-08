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

package tokens

import (
	"sync/atomic"
	"time"
)

const expiryTimeInSec = 30

// roundRobinAccessor implements TokenAccessor.
type roundRobinAccessor struct {
	accessTokens []string
	accessState  []int64
	counter      uint64
}

// Next implements TokenAccessor.Next.
func (tokens *roundRobinAccessor) Next() (uint64, string) {
	c := atomic.AddUint64(&tokens.counter, 1)
	l := len(tokens.accessTokens)
	index := c % uint64(l)

	// If selected accessToken is unavailable, wait.
	for !atomic.CompareAndSwapInt64(&tokens.accessState[index], 0, time.Now().Unix()) {
		currVal := atomic.LoadInt64(&tokens.accessState[index])
		expired := time.Now().After(time.Unix(currVal, 0).Add(expiryTimeInSec * time.Second))
		if !expired {
			continue
		}
		if atomic.CompareAndSwapInt64(&tokens.accessState[index], currVal, time.Now().Unix()) {
			break
		}
	}
	return index, tokens.accessTokens[index]
}

// Release implements TokenAccessor.Release.
func (tokens *roundRobinAccessor) Release(id uint64) {
	atomic.SwapInt64(&tokens.accessState[id], 0)
}

func makeRoundRobinAccessor(accessTokens []string) TokenAccessor {
	return &roundRobinAccessor{
		accessTokens: accessTokens,
		accessState:  make([]int64, len(accessTokens)),
	}
}
