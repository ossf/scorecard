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

// Token is used for GitHub token server RPC request/response.
type Token struct {
	Value string
	ID    uint64
}

// TokenOverRPC is an RPC handler implementation for Golang RPCs.
type TokenOverRPC struct {
	client TokenAccessor
}

// Next requests for the next available GitHub token.
// Server blocks the call until a token becomes available.
func (accessor *TokenOverRPC) Next(args struct{}, token *Token) error {
	id, val := accessor.client.Next()
	*token = Token{
		ID:    id,
		Value: val,
	}
	return nil
}

// Release inserts the token at `index` back into the token pool to be used by another client.
func (accessor *TokenOverRPC) Release(id uint64, reply *struct{}) error {
	accessor.client.Release(id)
	return nil
}

// NewTokenOverRPC creates a new instance of TokenOverRPC.
func NewTokenOverRPC(client TokenAccessor) *TokenOverRPC {
	return &TokenOverRPC{
		client: client,
	}
}
