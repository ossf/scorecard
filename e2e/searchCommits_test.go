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

package e2e

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo"
)

var _ = Describe("E2E TEST:SearchCommits", func() {
	Context("E2E TEST:SearchCommits", func() {
		It("Should return commits by dependabot", func() {
			repo, err := githubrepo.MakeGithubRepo("ossf/scorecard")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, clients.HeadSHA, 0)
			Expect(err).Should(BeNil())
			commits, err := repoClient.SearchCommits(clients.SearchCommitsOptions{Author: "dependabot[bot]"})
			Expect(err).Should(BeNil())
			Expect(len(commits)).Should(BeNumerically(">", 0))
		})
		It("Should return error as it is not using HEAD", func() {
			repo, err := githubrepo.MakeGithubRepo("ossf/scorecard")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, "123456789", 0)
			Expect(err).ShouldNot(Not(BeNil()))
		})
		It("Should return error as the user does not exists", func() {
			repo, err := githubrepo.MakeGithubRepo("ossf/scorecard")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, clients.HeadSHA, 0)
			Expect(err).Should(BeNil())
			_, err = repoClient.SearchCommits(clients.SearchCommitsOptions{Author: "thisuserdoesnotexists"})
			Expect(err).ShouldNot(BeNil())
		})
	})
})
