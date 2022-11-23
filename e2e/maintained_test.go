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
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo"
	scut "github.com/ossf/scorecard/v4/utests"
)

var _ = Describe("E2E TEST:"+checks.CheckMaintained, func() {
	Context("E2E TEST:Validating maintained status", func() {
		It("Should return valid maintained status", func() {
			dl := scut.TestDetailLogger{}
			repo, err := githubrepo.MakeGithubRepo("apache/airflow")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, clients.HeadSHA, 0)
			Expect(err).Should(BeNil())
			req := checker.CheckRequest{
				Ctx:        context.Background(),
				RepoClient: repoClient,
				Repo:       repo,
				Dlogger:    &dl,
			}
			expected := scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			}
			result := checks.Maintained(&req)
			// New version.
			Expect(scut.ValidateTestReturn(nil, "active repo", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
		})
		// To check the below maintanace status, the person running the test must own the repository.
		// Therefore, the below test works it just requires a maintained repository to exist under
		// someone's personal account, which I do not have.

		// It("Should return valid maintained status - GitLab", func() {
		// 	dl := scut.TestDetailLogger{}
		// 	// project url is gitlab.com/gitlab-org/gitlab
		// 	repo, err := gitlabrepo.MakeGitlabRepo("gitlab.com/gitlab-org/278964")
		// 	Expect(err).Should(BeNil())
		// 	repoClient, err := gitlabrepo.CreateGitlabClientWithToken(context.Background(),
		//		os.Getenv("GITLAB_AUTH_TOKEN"), repo)
		// 	Expect(err).Should(BeNil())
		// 	err = repoClient.InitRepo(repo, clients.HeadSHA, 0)
		// 	Expect(err).Should(BeNil())
		// 	req := checker.CheckRequest{
		// 		Ctx:        context.Background(),
		// 		RepoClient: repoClient,
		// 		Repo:       repo,
		// 		Dlogger:    &dl,
		// 	}
		// 	// TODO: update expected as needed
		// 	expected := scut.TestReturn{
		// 		Error:         nil,
		// 		Score:         checker.MaxResultScore,
		// 		NumberOfWarn:  0,
		// 		NumberOfInfo:  0,
		// 		NumberOfDebug: 0,
		// 	}
		// 	result := checks.Maintained(&req)
		// 	// New version.
		// 	Expect(scut.ValidateTestReturn(nil, "active repo", &expected, &result, &dl)).Should(BeTrue())
		// 	Expect(repoClient.Close()).Should(BeNil())
		// })
	})
})
