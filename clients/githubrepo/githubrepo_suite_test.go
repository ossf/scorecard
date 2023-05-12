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

package githubrepo

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/google/go-github/v38/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/shurcooL/githubv4"

	"github.com/ossf/scorecard/v4/clients/githubrepo/roundtripper"
	"github.com/ossf/scorecard/v4/log"
)

func TestGithubrepo(t *testing.T) {
	if val, exists := os.LookupEnv("SKIP_GINKGO"); exists && val == "1" {
		t.Skip()
	}
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Githubrepo Suite")
}

var (
	graphClient *githubv4.Client
	ghClient    *github.Client
)

type tokenType int

const (
	patTokenType tokenType = iota
	githubWorkflowDefaultTokenType
)

var tokType tokenType

func skipIfTokenIsNot(t tokenType, msg string) {
	if tokType != t {
		Skip(msg)
	}
}

func getGithubToken() string {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GITHUB_AUTH_TOKEN")
	}
	return token
}

var _ = BeforeSuite(func() {
	ctx := context.Background()
	logger := log.NewLogger(log.DebugLevel)
	rt := roundtripper.NewTransport(ctx, logger)
	httpClient := &http.Client{
		Transport: rt,
	}
	graphClient = githubv4.NewClient(httpClient)
	ghClient = github.NewClient(httpClient)

	tt := os.Getenv("TOKEN_TYPE")
	switch tt {
	case "PAT":
		tokType = patTokenType
	case "GITHUB_TOKEN":
		tokType = githubWorkflowDefaultTokenType
	default:
		panic(fmt.Sprintf("invald TOKEN_TYPE: %s", tt))
	}
})
