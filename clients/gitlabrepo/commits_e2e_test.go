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
	gitlabPATTokenType
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
	case "GITLAB_PAT":
		tokType = gitlabPATTokenType
	default:
		panic(fmt.Sprintf("invalid TOKEN_TYPE: %s", tt))
	}
})

var _ = Describe("E2E TEST: gitlabrepo.commitsHandler", func() {
	Context("ListCommits", func() {
		It("Checks whether commits are listed - GitLab", func() {
			skipIfTokenIsNot(patTokenType, "PAT only")
			repo, err := MakeGitlabRepo("https://gitlab.com/baserow/baserow")
			Expect(err).Should(BeNil())

			client, err := CreateGitlabClient(context.Background(), repo.Host())
			Expect(err).Should(BeNil())

			err = client.InitRepo(repo, "8a38c9f724c19b5422e27864a108318d1f769b8a", 20)
			Expect(err).Should(BeNil())

			c, err := client.ListCommits()
			Expect(err).Should(BeNil())

			Expect(len(c)).Should(BeNumerically(">", 0))
		})
	})
})
