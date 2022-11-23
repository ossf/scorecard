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
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks"
	"github.com/ossf/scorecard/v4/checks/raw"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo"
	"github.com/ossf/scorecard/v4/clients/gitlabrepo"
	scut "github.com/ossf/scorecard/v4/utests"
)

// TODO: use dedicated repo that don't change.
var _ = Describe("E2E TEST:"+checks.CheckCodeReview, func() {
	Context("E2E TEST:Validating use of code reviews", func() {
		It("Should return use of code reviews", func() {
			dl := scut.TestDetailLogger{}
			repo, err := githubrepo.MakeGithubRepo("ossf-tests/airflow")
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
				Score:         checker.MinResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			}
			result := checks.CodeReview(&req)
			Expect(scut.ValidateTestReturn(nil, "use code reviews", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
		})
		It("Should return use of code reviews at commit", func() {
			dl := scut.TestDetailLogger{}
			repo, err := githubrepo.MakeGithubRepo("ossf-tests/airflow")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, "0a6850647e531b08f68118ff8ca20577a5b4062c", 0)
			Expect(err).Should(BeNil())

			req := checker.CheckRequest{
				Ctx:        context.Background(),
				RepoClient: repoClient,
				Repo:       repo,
				Dlogger:    &dl,
			}
			expected := scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			}
			result := checks.CodeReview(&req)
			Expect(scut.ValidateTestReturn(nil, "use code reviews", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
		})
		It("Should return use of implicit code reviews at commit", func() {
			repo, err := githubrepo.MakeGithubRepo("spring-projects/spring-framework")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, "ca5e453f87f7e84033bb90a2fb54ee9f7fc94d61", 0)
			Expect(err).Should(BeNil())

			reviewData, err := raw.CodeReview(repoClient)
			Expect(err).Should(BeNil())
			Expect(reviewData.DefaultBranchChangesets).ShouldNot(BeEmpty())

			gh := 0
			for _, cs := range reviewData.DefaultBranchChangesets {
				if cs.ReviewPlatform == checker.ReviewPlatformGitHub {
					fmt.Printf("found github revision %s in spring-framework", cs.RevisionID)
					gh += 1
				}
			}
			Expect(gh).Should(BeNumerically("==", 2))
			Expect(repoClient.Close()).Should(BeNil())
		})
		// GitLab doesn't seem to preserve merge requests (pull requests in github) and some users had data lost in
		// the transfer from github so this returns a different value than the above GitHub test.
		It("Should return use of code reviews - GitLab", func() {
			dl := scut.TestDetailLogger{}
			// Project url is gitlab.com/N8BWert/airflow.
			repo, err := gitlabrepo.MakeGitlabRepo("gitlab.com/N8BWert/39537795")
			Expect(err).Should(BeNil())
			repoClient, err := gitlabrepo.CreateGitlabClientWithToken(context.Background(), os.Getenv("GITLAB_AUTH_TOKEN"), repo)
			Expect(err).Should(BeNil())
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
				Score:         checker.MinResultScore,
				NumberOfWarn:  20,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			}
			result := checks.CodeReview(&req)
			Expect(scut.ValidateTestReturn(nil, "use code reviews", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
		})
		// GitLab doesn't seem to preserve merge requests (pull requests in github) and some users had data lost in
		// the transfer from github so this returns a different value than the above GitHub test.
		It("Should return use of code reviews at commit - GitLab", func() {
			dl := scut.TestDetailLogger{}
			// project url is gitlab.com/N8BWert/airflow.
			repo, err := gitlabrepo.MakeGitlabRepo("gitlab.com/N8BWert/39537795")
			Expect(err).Should(BeNil())
			repoClient, err := gitlabrepo.CreateGitlabClientWithToken(context.Background(), os.Getenv("GITLAB_AUTH_TOKEN"), repo)
			Expect(err).Should(BeNil())
			err = repoClient.InitRepo(repo, "0a6850647e531b08f68118ff8ca20577a5b4062c", 0)
			Expect(err).Should(BeNil())

			req := checker.CheckRequest{
				Ctx:        context.Background(),
				RepoClient: repoClient,
				Repo:       repo,
				Dlogger:    &dl,
			}
			expected := scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  20,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			}
			result := checks.CodeReview(&req)
			Expect(scut.ValidateTestReturn(nil, "use code reviews", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
		})
		It("Should return use of implicit code reviews at commit", func() {
			repo, err := githubrepo.MakeGithubRepo("spring-projects/spring-framework")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, "ca5e453f87f7e84033bb90a2fb54ee9f7fc94d61", 0)
			Expect(err).Should(BeNil())

			reviewData, err := raw.CodeReview(repoClient)
			Expect(err).Should(BeNil())
			Expect(reviewData.DefaultBranchChangesets).ShouldNot(BeEmpty())

			gh := 0
			for _, cs := range reviewData.DefaultBranchChangesets {
				if cs.ReviewPlatform == checker.ReviewPlatformGitHub {
					fmt.Printf("found github revision %s in spring-framework", cs.RevisionID)
					gh += 1
				}
			}
			Expect(gh).Should(BeNumerically("==", 2))
			Expect(repoClient.Close()).Should(BeNil())
		})
	})
})
