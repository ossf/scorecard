// Copyright 2021 OpenSSF Scorecard Authors
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
	"github.com/ossf/scorecard/v4/clients/gitlabrepo"
	"github.com/ossf/scorecard/v4/clients/localdir"
	scut "github.com/ossf/scorecard/v4/utests"
)

var _ = Describe("E2E TEST:"+checks.CheckSecurityPolicy, func() {
	Context("E2E TEST:Validating security policy", func() {
		It("Should return valid security policy", func() {
			dl := scut.TestDetailLogger{}
			repo, err := githubrepo.MakeGithubRepo("ossf-tests/scorecard-check-security-policy-e2e")
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
				NumberOfInfo:  4,
				NumberOfDebug: 0,
			}
			result := checks.SecurityPolicy(&req)
			// New version.
			Expect(scut.ValidateTestReturn(nil, "policy found", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
		})
		It("Should return valid security policy at commitSHA", func() {
			dl := scut.TestDetailLogger{}
			repo, err := githubrepo.MakeGithubRepo("ossf-tests/scorecard-check-security-policy-e2e")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, "46e9bc6538b2f788b6e3d18f8c8c174146565e93", 0)
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
				NumberOfInfo:  4,
				NumberOfDebug: 0,
			}
			result := checks.SecurityPolicy(&req)
			// New version.
			Expect(scut.ValidateTestReturn(nil, "policy found", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
		})
		It("Should return valid security policy for local repoClient at head", func() {
			dl := scut.TestDetailLogger{}

			tmpDir, err := os.MkdirTemp("", "")
			Expect(err).Should(BeNil())
			defer os.RemoveAll(tmpDir)

			_, e := git.PlainClone(tmpDir, false, &git.CloneOptions{
				URL: "http://github.com/ossf-tests/scorecard-check-security-policy-e2e",
			})
			Expect(e).Should(BeNil())

			repo, err := localdir.MakeLocalDirRepo(tmpDir)
			Expect(err).Should(BeNil())

			x := localdir.CreateLocalDirClient(context.Background(), logger)
			err = x.InitRepo(repo, clients.HeadSHA, 0)
			Expect(err).Should(BeNil())

			req := checker.CheckRequest{
				Ctx:        context.Background(),
				RepoClient: x,
				Repo:       repo,
				Dlogger:    &dl,
			}
			expected := scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  4,
				NumberOfDebug: 0,
			}
			result := checks.SecurityPolicy(&req)
			// New version.
			Expect(scut.ValidateTestReturn(nil, "policy found", &expected, &result, &dl)).Should(BeTrue())
			Expect(x.Close()).Should(BeNil())
		})
		It("Should return valid security policy - GitLab", func() {
			skipIfTokenIsNot(gitlabPATTokenType, "GitLab only")

			dl := scut.TestDetailLogger{}
			// project url is gitlab.com/bramw/baserow.
			repo, err := gitlabrepo.MakeGitlabRepo("gitlab.com/ossf-test/baserow")
			Expect(err).Should(BeNil())
			repoClient, err := gitlabrepo.CreateGitlabClient(context.Background(), repo.Host())
			Expect(err).Should(BeNil())
			err = repoClient.InitRepo(repo, clients.HeadSHA, 0)
			Expect(err).Should(BeNil())

			req := checker.CheckRequest{
				Ctx:        context.Background(),
				RepoClient: repoClient,
				Repo:       repo,
				Dlogger:    &dl,
			}
			// TODO: update expected based on what is returned from gitlab project.
			expected := scut.TestReturn{
				Error:         nil,
				Score:         9,
				NumberOfWarn:  1,
				NumberOfInfo:  3,
				NumberOfDebug: 0,
			}
			result := checks.SecurityPolicy(&req)
			// New version.
			Expect(scut.ValidateTestReturn(nil, "policy found", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
		})
		It("Should return valid security policy at commitSHA - GitLab", func() {
			skipIfTokenIsNot(gitlabPATTokenType, "GitLab only")

			dl := scut.TestDetailLogger{}
			// project url is gitlab.com/bramw/baserow.
			repo, err := gitlabrepo.MakeGitlabRepo("gitlab.com/ossf-test/baserow")
			Expect(err).Should(BeNil())
			repoClient, err := gitlabrepo.CreateGitlabClient(context.Background(), repo.Host())
			Expect(err).Should(BeNil())
			// url to commit is https://gitlab.com/bramw/baserow/-/commit/28e6224b7d86f7b30bad6adb6b42f26a814c2f58
			err = repoClient.InitRepo(repo, "28e6224b7d86f7b30bad6adb6b42f26a814c2f58", 0)
			Expect(err).Should(BeNil())

			req := checker.CheckRequest{
				Ctx:        context.Background(),
				RepoClient: repoClient,
				Repo:       repo,
				Dlogger:    &dl,
			}

			expected := scut.TestReturn{
				Error:         nil,
				Score:         9,
				NumberOfWarn:  1,
				NumberOfInfo:  3,
				NumberOfDebug: 0,
			}
			result := checks.SecurityPolicy(&req)
			// New version.
			Expect(scut.ValidateTestReturn(nil, "policy found", &expected, &result, &dl)).Should(BeTrue())
			Expect(repoClient.Close()).Should(BeNil())
		})
	})
})
