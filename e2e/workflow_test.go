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

var _ = Describe("E2E TEST:WorkflowRun", func() {
	Context("E2E TEST:WorkflowRun", func() {
		It("Should return scorecard analysis workflow run", func() {
			// using the scorecard repo as an example. The tests repo workflow won't have any runs in the future and
			// that is why we are using the scorecard repo.
			repo, err := githubrepo.MakeGithubRepo("ossf/scorecard")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, clients.HeadSHA, 0)
			Expect(err).Should(BeNil())
			runs, err := repoClient.ListSuccessfulWorkflowRuns("scorecard-analysis.yml")
			Expect(err).Should(BeNil())
			Expect(len(runs)).Should(BeNumerically(">", 0))
		})
		It("Should should fail because only head queries are supported", func() {
			// using the scorecard repo as an example. The tests repo workflow won't have any runs in the future and
			// that is why we are using the scorecard repo.
			repo, err := githubrepo.MakeGithubRepo("ossf/scorecard")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, "123456789", 0)
			Expect(err).Should(BeNil())
			runs, err := repoClient.ListSuccessfulWorkflowRuns("scorecard-analysis.yml")
			Expect(err).ShouldNot(BeNil())
			Expect(len(runs)).Should(BeNumerically("==", 0))
		})
		It("Should should fail the workflow file doesn't exist", func() {
			// using the scorecard repo as an example. The tests repo workflow won't have any runs in the future and
			// that is why we are using the scorecard repo.
			repo, err := githubrepo.MakeGithubRepo("ossf/scorecard")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, clients.HeadSHA, 0)
			Expect(err).Should(BeNil())
			runs, err := repoClient.ListSuccessfulWorkflowRuns("non-existing.yml")
			Expect(err).ShouldNot(BeNil())
			Expect(len(runs)).Should(BeNumerically("==", 0))
		})
	})
})
