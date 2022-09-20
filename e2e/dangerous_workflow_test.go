// Copyright 2021 Security Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/go-git/go-git/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/clients/githubrepo"
	"github.com/ossf/scorecard/v4/clients/localdir"
	scut "github.com/ossf/scorecard/v4/utests"
)

var _ = Describe("E2E TEST:"+checks.CheckTokenPermissions, func() {
	Context("E2E TEST:Validating dangerous workflow check", func() {
		It("Should return dangerous workflow works", func() {
			dl := scut.TestDetailLogger{}
			repo, err := githubrepo.MakeGithubRepo("ossf-tests/scorecard-check-dangerous-workflow-e2e")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, clients.HeadSHA)
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
				NumberOfWarn:  1,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			}
			result := checks.DangerousWorkflow(&req)
			// New version.
			Expect(scut.ValidateTestReturn(nil, "dangerous workflow", &expected, &result, &dl)).Should(BeTrue())
		})
		It("Should return dangerous workflow at commit", func() {
			dl := scut.TestDetailLogger{}
			repo, err := githubrepo.MakeGithubRepo("ossf-tests/scorecard-check-dangerous-workflow-e2e")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, "8db326e9ba20517feeefd157524a89184ed41f7f")
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
				NumberOfWarn:  1,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			}
			result := checks.DangerousWorkflow(&req)
			// New version.
			Expect(scut.ValidateTestReturn(nil, "dangerous workflow", &expected, &result, &dl)).Should(BeTrue())
		})
		It("Should return dangerous workflow for local repoClient", func() {
			dl := scut.TestDetailLogger{}

			tmpDir, err := os.MkdirTemp("", "")
			Expect(err).Should(BeNil())
			defer os.RemoveAll(tmpDir)

			_, e := git.PlainClone(tmpDir, false, &git.CloneOptions{
				URL: "http://github.com/ossf-tests/scorecard-check-dangerous-workflow-e2e",
			})
			Expect(e).Should(BeNil())

			repo, err := localdir.MakeLocalDirRepo(tmpDir)
			Expect(err).Should(BeNil())

			x := localdir.CreateLocalDirClient(context.Background(), logger)
			err = x.InitRepo(repo, clients.HeadSHA)
			Expect(err).Should(BeNil())

			req := checker.CheckRequest{
				Ctx:        context.Background(),
				RepoClient: x,
				Repo:       repo,
				Dlogger:    &dl,
			}
			expected := scut.TestReturn{
				Error:         nil,
				Score:         checker.MinResultScore,
				NumberOfWarn:  1,
				NumberOfInfo:  0,
				NumberOfDebug: 0,
			}
			result := checks.DangerousWorkflow(&req)
			// New version.
			Expect(scut.ValidateTestReturn(nil, "dangerous workflow", &expected, &result, &dl)).Should(BeTrue())
		})
		// TODO: rewrite github actions as gitlab ci
		// It("Should return dangerous workflow works - GitLab", func() {
		// 	dl := scut.TestDetailLogger{}
		// 	// TODO: Write a test repository with a dangerous workflow.
		// 	repo, err := gitlabrepo.MakeGitlabRepo("gitlab.ossf.com/ossf-tests/scorecard-check-dangerous-workflow-e2e")
		// 	Expect(err).Should(BeNil())
		// 	repoClient, err := gitlabrepo.CreateGitlabClientWithToken(context.Background(),
		//		os.Getenv("GITLAB_AUTH_TOKNE"), repo)
		// 	Expect(err).Should(BeNil())
		// 	err = repoClient.InitRepo(repo, clients.HeadSHA)
		// 	Expect(err).Should(BeNil())
		// 	req := checker.CheckRequest{
		// 		Ctx:        context.Background(),
		// 		RepoClient: repoClient,
		// 		Repo:       repo,
		// 		Dlogger:    &dl,
		// 	}
		// 	expected := scut.TestReturn{
		// 		Error:         nil,
		// 		Score:         checker.MinResultScore,
		// 		NumberOfWarn:  1,
		// 		NumberOfInfo:  0,
		// 		NumberOfDebug: 0,
		// 	}
		// 	result := checks.DangerousWorkflow(&req)
		// 	// New version.
		// 	Expect(scut.ValidateTestReturn(nil, "dangerous workflow", &expected, &result, &dl)).Should(BeTrue())
		// })
		// It("Should return dangerous workflow at commit - GitLab", func() {
		// 	dl := scut.TestDetailLogger{}
		// 	// TODO: Write a test repository with a dangerous workflow.
		// 	repo, err := gitlabrepo.MakeGitlabRepo("gitlab.ossf.com/ossf-tests/scorecard-check-dangerous-workflow-e2e")
		// 	Expect(err).Should(BeNil())
		// 	repoClient, err := gitlabrepo.CreateGitlabClientWithToken(context.Background(),
		//		os.Getenv("GITLAB_AUTH_TOKEN"), repo)
		// 	Expect(err).Should(BeNil())
		// 	err = repoClient.InitRepo(repo, "8db326e9ba20517feeefd157524a89184ed41f7f")
		// 	Expect(err).Should(BeNil())
		// 	req := checker.CheckRequest{
		// 		Ctx:        context.Background(),
		// 		RepoClient: repoClient,
		// 		Repo:       repo,
		// 		Dlogger:    &dl,
		// 	}
		// 	expected := scut.TestReturn{
		// 		Error:         nil,
		// 		Score:         checker.MinResultScore,
		// 		NumberOfWarn:  1,
		// 		NumberOfInfo:  0,
		// 		NumberOfDebug: 0,
		// 	}
		// 	result := checks.DangerousWorkflow(&req)
		// 	// New version.
		// 	Expect(scut.ValidateTestReturn(nil, "dangerous workflow", &expected, &result, &dl)).Should(BeTrue())
		// })
	})
})
