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
	"log"
	"net/rpc"
)

// rpcAccessor implements TokenAccessor.
type rpcAccessor struct {
	client *rpc.Client
}

// Next implements TokenAccessor.Next.
func (accessor *rpcAccessor) Next() (uint64, string) {
	var token Token
	if err := accessor.client.Call("TokenOverRPC.Next", struct{}{}, &token); err != nil {
		log.Printf("error during RPC call Next: %v", err)
		return 0, ""
	}
	return token.ID, token.Value
}

// Release implements TokenAccessor.Release.
func (accessor *rpcAccessor) Release(id uint64) {
	if err := accessor.client.Call("TokenOverRPC.Release", id, &struct{}{}); err != nil {
		log.Printf("error during RPC call Release: %v", err)
	}
}

func makeRPCAccessor(serverURL string) TokenAccessor {
	client, err := rpc.DialHTTP("tcp", serverURL)
	if err != nil {
		panic(err)
	}
	return &rpcAccessor{
		client: client,
	}
}
