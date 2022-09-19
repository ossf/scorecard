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

package e2e

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/v4/log"
)

func TestE2e(t *testing.T) {
	if val, exists := os.LookupEnv("SKIP_GINKGO"); exists && val == "1" {
		t.Skip()
	}
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2e Suite")
}

var logger *log.Logger

type tokenType int

const (
	patTokenType tokenType = iota
	githubWorkflowDefaultTokenType
	gitlabPatToken tokenType = iota
)

var tokType tokenType

func skipIfTokenIsNot(t tokenType, msg string) {
	if tokType != t {
		Skip(msg)
	}
}

var _ = BeforeSuite(func() {
	logger = log.NewLogger(log.DebugLevel)
	tt := os.Getenv("TOKEN_TYPE")
	switch tt {
	case "PAT":
		tokType = patTokenType
	case "GITHUB_TOKEN":
		tokType = githubWorkflowDefaultTokenType
	case "GITLAB_AUTH_TOKEN":
		tokType = gitlabPatToken
	default:
		panic(fmt.Sprintf("invald TOKEN_TYPE: %s", tt))
	}
})
