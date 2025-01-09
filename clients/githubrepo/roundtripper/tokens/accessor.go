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

// Package tokens defines interface to access GitHub PATs.
package tokens

import (
	"log"
	"os"
	"strings"
)

// githubAuthServer is the RPC URL for the token server.
const githubAuthServer = "GITHUB_AUTH_SERVER"

// env variables from which GitHub auth tokens are read, in order of precedence.
var githubAuthTokenEnvVars = []string{"GITHUB_AUTH_TOKEN", "GITHUB_TOKEN", "GH_TOKEN", "GH_AUTH_TOKEN"}

// TokenAccessor interface defines a `retrieve-once` data structure.
// Implementations of this interface must be thread-safe.
type TokenAccessor interface {
	Next() (uint64, string)
	Release(uint64)
}

var logDuplicateTokenWarning = func(firstName string, clashingName string) {
	var stringBuilder strings.Builder
	stringBuilder.WriteString("Warning: PATs stored in env variables ")
	stringBuilder.WriteString(firstName)
	stringBuilder.WriteString(" and ")
	stringBuilder.WriteString(clashingName)
	stringBuilder.WriteString(" differ. Scorecard will use the former.")
	log.Println(stringBuilder.String())
}

func readGitHubTokens() (string, bool) {
	var firstName, firstToken string
	for _, name := range githubAuthTokenEnvVars {
		if token, exists := os.LookupEnv(name); exists && token != "" {
			if firstName == "" {
				firstName = name
				firstToken = token
			} else if token != firstToken {
				logDuplicateTokenWarning(firstName, name)
			}
		}
	}
	if firstName == "" {
		return "", false
	} else {
		return firstToken, true
	}
}

// MakeTokenAccessor is a factory function of TokenAccessor.
func MakeTokenAccessor() TokenAccessor {
	if value, exists := readGitHubTokens(); exists {
		return makeRoundRobinAccessor(strings.Split(value, ","))
	}
	if value, exists := os.LookupEnv(githubAuthServer); exists && value != "" {
		return makeRPCAccessor(value)
	}
	return nil
}
