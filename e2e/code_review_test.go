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

package e2e

import (
	"context"
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
				NumberOfDebug: 1,
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
				NumberOfDebug: 1,
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
					gh += 1
				}
			}
			Expect(gh).Should(BeNumerically("==", 2))

			Expect(repoClient.Close()).Should(BeNil())
		})
		It("Should return inconclusive results for a single-maintainer project with only self- or bot changesets", func() {
			dl := scut.TestDetailLogger{}
			repo, err := githubrepo.MakeGithubRepo("Kromey/fast_poisson")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, "bb7b9606690c2b386dc9e2cbe0216d389ed1f078", 0)
			Expect(err).Should(BeNil())

			req := checker.CheckRequest{
				Ctx:        context.Background(),
				RepoClient: repoClient,
				Repo:       repo,
				Dlogger:    &dl,
			}
			expected := scut.TestReturn{
				Score:         checker.InconclusiveResultScore,
				NumberOfDebug: 1,
			}
			result := checks.CodeReview(&req)
			Expect(scut.ValidateTestReturn(nil, "use code reviews", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
		})
		It("Should return minimum score for a single-maintainer project with some unreviewed human changesets", func() {
			dl := scut.TestDetailLogger{}
			repo, err := githubrepo.MakeGithubRepo("Kromey/fast_poisson")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, "10aefa7c9a6669ef34e209c3c4b6ad48dd9844e3", 0)
			Expect(err).Should(BeNil())

			req := checker.CheckRequest{
				Ctx:        context.Background(),
				RepoClient: repoClient,
				Repo:       repo,
				Dlogger:    &dl,
			}
			expected := scut.TestReturn{
				Score:         checker.MinResultScore,
				NumberOfDebug: 1,
			}
			result := checks.CodeReview(&req)
			Expect(scut.ValidateTestReturn(nil, "use code reviews", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
		})
	})
	// GitLab doesn't seem to preserve merge requests (pull requests in github) and some users had data lost in
	// the transfer from github so this returns a different value than the above GitHub test.
	It("Should return use of code reviews at commit - GitLab", func() {
		skipIfTokenIsNot(gitlabPATTokenType, "GitLab only")

		dl := scut.TestDetailLogger{}
		repo, err := gitlabrepo.MakeGitlabRepo("gitlab.com/fdroid/fdroidclient")
		Expect(err).Should(BeNil())
		repoClient, err := gitlabrepo.CreateGitlabClientWithToken(context.Background(), os.Getenv("GITLAB_AUTH_TOKEN"), repo)
		Expect(err).Should(BeNil())
		err = repoClient.InitRepo(repo, "1f7ed43c120047102862d9d1d644f5b2de7a47f2", 0)
		Expect(err).Should(BeNil())

		req := checker.CheckRequest{
			Ctx:        context.Background(),
			RepoClient: repoClient,
			Repo:       repo,
			Dlogger:    &dl,
		}
		expected := scut.TestReturn{
			Error:         nil,
			Score:         3,
			NumberOfDebug: 1,
		}
		result := checks.CodeReview(&req)
		Expect(scut.ValidateTestReturn(nil, "use code reviews", &expected, &result, &dl)).Should(BeTrue())
		Expect(repoClient.Close()).Should(BeNil())
	})
	// GitLab doesn't seem to preserve merge requests (pull requests in github) and some users had data lost in
	// the transfer from github so this returns a different value than the above GitHub test.
	It("Should return use of code reviews at HEAD - GitLab", func() {
		skipIfTokenIsNot(gitlabPATTokenType, "GitLab only")

		dl := scut.TestDetailLogger{}
		repo, err := gitlabrepo.MakeGitlabRepo("gitlab.com/gitlab-org/gitlab")
		Expect(err).Should(BeNil())
		repoClient, err := gitlabrepo.CreateGitlabClientWithToken(context.Background(), os.Getenv("GITLAB_AUTH_TOKEN"), repo)
		Expect(err).Should(BeNil())
		err = repoClient.InitRepo(repo, clients.HeadSHA, 0)
		// err = repoClient.InitRepo(repo, "0b5ba5049f3e5b8e945305acfa45c44d63df21b1", 0)
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
			NumberOfDebug: 1,
		}
		result := checks.CodeReview(&req)
		Expect(scut.ValidateTestReturn(nil, "use code reviews", &expected, &result, &dl)).Should(BeTrue())
		Expect(repoClient.Close()).Should(BeNil())
	})
})
