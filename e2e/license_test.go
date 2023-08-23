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

var _ = Describe("E2E TEST:"+checks.CheckLicense, func() {
	Context("E2E TEST:Validating license file check", func() {
		It("Should return license check works", func() {
			dl := scut.TestDetailLogger{}
			repo, err := githubrepo.MakeGithubRepo("ossf-tests/scorecard-check-license-e2e")
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
				NumberOfInfo:  2,
				NumberOfDebug: 0,
			}
			result := checks.License(&req)

			Expect(scut.ValidateTestReturn(nil, "license found", &expected, &result,
				&dl)).Should(BeTrue())
		})
		It("Should return license check works at commitSHA", func() {
			dl := scut.TestDetailLogger{}
			repo, err := githubrepo.MakeGithubRepo("ossf-tests/scorecard-check-license-e2e")
			Expect(err).Should(BeNil())
			repoClient := githubrepo.CreateGithubRepoClient(context.Background(), logger)
			err = repoClient.InitRepo(repo, "c3a8778e73ea95f937c228a34ee57d5e006f7304", 0)
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
				NumberOfInfo:  2,
				NumberOfDebug: 0,
			}
			result := checks.License(&req)

			Expect(scut.ValidateTestReturn(nil, "license found", &expected, &result,
				&dl)).Should(BeTrue())
		})
		It("Should return license check works for the local repoclient", func() {
			dl := scut.TestDetailLogger{}

			tmpDir, err := os.MkdirTemp("", "")
			Expect(err).Should(BeNil())
			defer os.RemoveAll(tmpDir)

			_, e := git.PlainClone(tmpDir, false, &git.CloneOptions{
				URL: "http://github.com/ossf-tests/scorecard-check-license-e2e",
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
				Score:         checker.MaxResultScore - 1,
				NumberOfWarn:  1,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			}
			result := checks.License(&req)

			Expect(scut.ValidateTestReturn(nil, "license found", &expected, &result,
				&dl)).Should(BeTrue())
		})
		It("Should return license check works - GitLab", func() {
			skipIfTokenIsNot(gitlabPATTokenType, "GitLab only")

			dl := scut.TestDetailLogger{}
			repo, err := gitlabrepo.MakeGitlabRepo("gitlab.com/ossf-test/scorecard-check-license-e2e")
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
			expected := scut.TestReturn{
				Error:        nil,
				Score:        10,
				NumberOfInfo: 2,
			}
			result := checks.License(&req)

			Expect(scut.ValidateTestReturn(nil, "license found", &expected, &result,
				&dl)).Should(BeTrue())
		})
		It("Should return license check works for unrecognized license type - GitLab", func() {
			skipIfTokenIsNot(gitlabPATTokenType, "GitLab only")

			dl := scut.TestDetailLogger{}
			repo, err := gitlabrepo.MakeGitlabRepo("gitlab.com/ossf-test/scorecard-check-license-e2e-unrecognized-license-type")
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
			expected := scut.TestReturn{
				Error:         nil,
				Score:         9,
				NumberOfWarn:  1,
				NumberOfInfo:  1,
				NumberOfDebug: 0,
			}
			result := checks.License(&req)

			Expect(scut.ValidateTestReturn(nil, "license found", &expected, &result,
				&dl)).Should(BeTrue())
		})
		It("Should return license check works at commitSHA - GitLab", func() {
			skipIfTokenIsNot(gitlabPATTokenType, "GitLab only")

			dl := scut.TestDetailLogger{}
			repo, err := gitlabrepo.MakeGitlabRepo("gitlab.com/ossf-test/scorecard-check-license-e2e")
			Expect(err).Should(BeNil())
			repoClient, err := gitlabrepo.CreateGitlabClient(context.Background(), repo.Host())
			Expect(err).Should(BeNil())
			err = repoClient.InitRepo(repo, "c3a8778e73ea95f937c228a34ee57d5e006f7304", 0)
			Expect(err).Should(BeNil())
			req := checker.CheckRequest{
				Ctx:        context.Background(),
				RepoClient: repoClient,
				Repo:       repo,
				Dlogger:    &dl,
			}
			expected := scut.TestReturn{
				Error:        nil,
				Score:        10,
				NumberOfInfo: 2,
			}
			result := checks.License(&req)

			Expect(scut.ValidateTestReturn(nil, "license found", &expected, &result,
				&dl)).Should(BeTrue())
		})
	})
})
