// Copyright 2023 OpenSSF Scorecard Authors
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

package gitlabrepo

import (
	"context"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

var _ = BeforeSuite(func() {
	tt := os.Getenv("TOKEN_TYPE")
	switch tt {
	case "PAT":
		tokType = patTokenType
	case "GITHUB_TOKEN":
		tokType = githubWorkflowDefaultTokenType
	default:
		panic(fmt.Sprintf("invalid TOKEN_TYPE: %s", tt))
	}
})

var _ = Describe("E2E TEST: gitlabrepo.commitsHandler", func() {
	Context("Test whether commits are listed - GitLab", func() {
		skipIfTokenIsNot(patTokenType, "PAT only")
		repo, err := MakeGitlabRepo("https://gitlab.com/fdroid/fdroidclient")
		Expect(err).ShouldNot(BeNil())

		client, err := CreateGitlabClient(context.Background(), repo.Host())
		Expect(err).ShouldNot(BeNil())

		err = client.InitRepo(repo, "a4bbef5c70fd2ac7c15437a22ef0f9ef0b447d08", 20)
		Expect(err).ShouldNot(BeNil())

		c, err := client.ListCommits()
		Expect(err).ShouldNot(BeNil())

		Expect(len(c)).Should(BeNumerically(">", 0))
	})
})
