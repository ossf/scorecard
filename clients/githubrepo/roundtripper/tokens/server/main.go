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

// Package main implements the GitHub token server.
package main

import (
	"net"
	"net/http"
	"net/rpc"

	"github.com/ossf/scorecard/v4/clients/githubrepo/roundtripper/tokens"
)

func main() {
	// Sanity check
	tokenAccessor := tokens.MakeTokenAccessor()
	if tokenAccessor == nil {
		panic("")
	}

	rpcAccessor := tokens.NewTokenOverRPC(tokenAccessor)
	if err := rpc.Register(rpcAccessor); err != nil {
		panic(err)
	}
	rpc.HandleHTTP()

	//nolint:gosec // using `localhost:8080` for gosec102 causes connection refused errors.
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	//nolint:gosec // internal server.
	if err := http.Serve(l, nil); err != nil {
		panic(err)
	}
}
