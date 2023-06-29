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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/xanzy/go-gitlab"
)

var _ = Describe("E2E TEST: gitlabrepo.commitsHandler", func() {
	Context("Test whether commits are listed - GitLab", func() {
		It("returns that the repo is inaccessible", func() {
			repo, err := MakeGitlabRepo("https://gitlab.com/fdroid/fdroidclient")
			Expect(err).Should(BeNil())

			client, err := CreateGitlabClient(context.Background(), repo.Host())
			Expect(err).Should(BeNil())

			err = client.InitRepo(repo, "a4bbef5c70fd2ac7c15437a22ef0f9ef0b447d08", 20)
			Expect(err).Should(BeNil())

			c, err := client.ListCommits()
			Expect(err).Should(BeNil())

			Expect(len(c)).Should(BeNumerically(">", 0))
		})
	})
})

var _ = Describe("E2E TEST: gitlabrepo.client", func() {
	Context("checkRepoInaccessible", func() {
		It("returns that the repo is inaccessible - GitLab", func() {
			skipIfTokenIsNot(gitlabPATTokenType, "GitLab only")
			repo, err := MakeGitlabRepo("https://gitlab.com/ossf-test/private-project")
			Expect(err).Should(BeNil())

			client, err := CreateGitlabClient(context.Background(), repo.Host())
			Expect(err).Should(BeNil())

			glRepo, ok := repo.(*repoURL)
			Expect(ok).Should(BeTrue())

			// Sanity check.
			glcl, ok := client.(*Client)
			Expect(ok).Should(BeTrue())

			path := fmt.Sprintf("%s/%s", glRepo.owner, glRepo.project)
			proj, _, err := glcl.glClient.Projects.GetProject(path, &gitlab.GetProjectOptions{})
			Expect(err).Should(BeNil())

			err = checkRepoInaccessible(proj)
			Expect(err).ShouldNot(BeNil())
		})

		It("should initialize repos without error - GitLab", func() {
			repo, err := MakeGitlabRepo("https://gitlab.com/fdroid/fdroidclient")
			Expect(err).Should(BeNil())

			client, err := CreateGitlabClient(context.Background(), repo.Host())
			Expect(err).Should(BeNil())

			err = client.InitRepo(repo, "a4bbef5c70fd2ac7c15437a22ef0f9ef0b447d08", 20)
			Expect(err).Should(BeNil())
		})
	})
})
